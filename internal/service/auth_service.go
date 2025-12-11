package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zuquanzhi/Chirp/backend/internal/domain"
	"github.com/zuquanzhi/Chirp/backend/pkg/limiter"
	"github.com/zuquanzhi/Chirp/backend/pkg/sms"
	"github.com/zuquanzhi/Chirp/backend/pkg/util"
)

type AuthService struct {
	userRepo    domain.UserRepository
	codeRepo    domain.VerificationCodeRepository
	smsSender   sms.Sender
	rateLimiter limiter.RateLimiter
	jwtSecret   string
}

func NewAuthService(userRepo domain.UserRepository, codeRepo domain.VerificationCodeRepository, smsSender sms.Sender, rateLimiter limiter.RateLimiter, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		codeRepo:    codeRepo,
		smsSender:   smsSender,
		rateLimiter: rateLimiter,
		jwtSecret:   jwtSecret,
	}
}

func (s *AuthService) SendCode(ctx context.Context, phone, purpose string) error {
	// Rate Limit Check
	if s.rateLimiter != nil && !s.rateLimiter.Allow(phone) {
		return errors.New("too many requests, please try again later")
	}

	// Generate 6 digit secure random code
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return err
	}
	code := fmt.Sprintf("%06d", n.Int64())

	// Save to DB
	if err := s.codeRepo.Save(ctx, phone, code, purpose, 5*time.Minute); err != nil {
		return err
	}

	// Send SMS
	if err := s.smsSender.Send(ctx, phone, code, purpose); err != nil {
		return err
	}

	return nil
}

func (s *AuthService) SignupWithPhone(ctx context.Context, name, phone, code, password string) (*domain.User, error) {
	// Verify Code
	storedCode, err := s.codeRepo.Get(ctx, phone, "signup")
	if err != nil {
		return nil, err
	}
	if storedCode == "" || storedCode != code {
		return nil, errors.New("invalid or expired verification code")
	}

	// Check existing
	existing, err := s.userRepo.GetByPhoneNumber(ctx, phone)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("phone number already used")
	}

	hash, err := util.HashPassword(password)
	if err != nil {
		return nil, err
	}

	u := &domain.User{
		Name:        name,
		PhoneNumber: phone,
		Password:    hash,
		// Email is optional or can be generated/placeholder
		Email: phone + "@phone.chirp",
	}

	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}

	// Cleanup code
	s.codeRepo.Delete(ctx, phone, "signup")

	return u, nil
}

func (s *AuthService) LoginWithPhone(ctx context.Context, phone, code string) (string, error) {
	// Verify Code
	storedCode, err := s.codeRepo.Get(ctx, phone, "login")
	if err != nil {
		return "", err
	}
	if storedCode == "" || storedCode != code {
		return "", errors.New("invalid or expired verification code")
	}

	u, err := s.userRepo.GetByPhoneNumber(ctx, phone)
	if err != nil {
		return "", err
	}
	if u == nil {
		return "", errors.New("user not found")
	}

	// Generate Token
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   u.ID,
		"email": u.Email,
		"phone": u.PhoneNumber,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	})

	// Cleanup code
	s.codeRepo.Delete(ctx, phone, "login")

	return t.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) Signup(ctx context.Context, name, email, password string) (*domain.User, error) {
	existing, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("email already used")
	}

	hash, err := util.HashPassword(password)
	if err != nil {
		return nil, err
	}

	u := &domain.User{
		Name:     name,
		Email:    email,
		Password: hash,
	}

	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", err
	}
	if u == nil {
		return "", errors.New("invalid credentials")
	}

	if err := util.CheckPassword(u.Password, password); err != nil {
		return "", errors.New("invalid credentials")
	}

	// Generate Token
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   u.ID,
		"email": u.Email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	})

	return t.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *AuthService) UpdateProfile(ctx context.Context, u *domain.User) (*domain.User, error) {
	// Ensure user exists
	existing, err := s.userRepo.GetByID(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("user not found")
	}

	// Apply updates (allow empty string to clear)
	existing.Name = u.Name
	existing.School = u.School
	existing.StudentID = u.StudentID
	existing.Birthdate = u.Birthdate
	existing.Address = u.Address
	existing.Gender = u.Gender

	if err := s.userRepo.UpdateProfile(ctx, existing); err != nil {
		return nil, err
	}

	return existing, nil
}
