package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/zuquanzhi/Chirp/backend/internal/domain"
)

type codeRepository struct {
	db *sql.DB
}

func NewCodeRepository(db *sql.DB) domain.VerificationCodeRepository {
	return &codeRepository{db: db}
}

func (r *codeRepository) Save(ctx context.Context, phone, code, purpose string, duration time.Duration) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM verification_codes WHERE phone_number = ? AND purpose = ?`, phone, purpose); err != nil {
		return err
	}
	expiresAt := time.Now().Add(duration)
	_, err := r.db.ExecContext(ctx, `INSERT INTO verification_codes (phone_number, code, purpose, expires_at) VALUES (?, ?, ?, ?)`, phone, code, purpose, expiresAt)
	return err
}

func (r *codeRepository) Get(ctx context.Context, phone, purpose string) (string, error) {
	var (
		code      string
		expiresAt time.Time
	)
	err := r.db.QueryRowContext(ctx, `SELECT code, expires_at FROM verification_codes WHERE phone_number = ? AND purpose = ? ORDER BY created_at DESC LIMIT 1`, phone, purpose).Scan(&code, &expiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	if time.Now().After(expiresAt) {
		return "", nil
	}
	return code, nil
}

func (r *codeRepository) Delete(ctx context.Context, phone, purpose string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM verification_codes WHERE phone_number = ? AND purpose = ?`, phone, purpose)
	return err
}
