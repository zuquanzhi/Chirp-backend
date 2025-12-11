package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"mime/multipart"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/zuquanzhi/Chirp/backend/internal/domain"
)

type ResourceService struct {
	repo    domain.ResourceRepository
	storage FileStorage
}

func NewResourceService(repo domain.ResourceRepository, storage FileStorage) *ResourceService {
	return &ResourceService{
		repo:    repo,
		storage: storage,
	}
}

func (s *ResourceService) Upload(ctx context.Context, ownerID *int64, title, desc, subject, resourceType string, file multipart.File, header *multipart.FileHeader) (*domain.Resource, error) {
	// Calculate Hash
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return nil, err
	}
	fileHash := hex.EncodeToString(hash.Sum(nil))

	// Reset file pointer
	file.Seek(0, 0)

	id := uuid.New().String()
	ext := filepath.Ext(header.Filename)
	storedName := id + ext
	
	// Use Storage Interface
	savedName, size, err := s.storage.Save(ctx, file, storedName)
	if err != nil {
		return nil, err
	}

	res := &domain.Resource{
		OwnerID:      ownerID,
		Title:        title,
		Description:  desc,
		Filename:     savedName, // Store the key/path returned by storage
		OriginalName: header.Filename,
		Size:         size,
		FileHash:     fileHash,
		Status:       domain.ResourceStatusPending,
		Subject:      subject,
		Type:         resourceType,
	}

	if err := s.repo.Create(ctx, res); err != nil {
		return nil, err
	}

	// Populate URL
	res.URL = s.storage.GetPublicURL(savedName)

	return res, nil
}

func (s *ResourceService) List(ctx context.Context, status domain.ResourceStatus, search string) ([]domain.Resource, error) {
	list, err := s.repo.List(ctx, status, search)
	if err != nil {
		return nil, err
	}
	// Populate URLs
	for i := range list {
		list[i].URL = s.storage.GetPublicURL(list[i].Filename)
	}
	return list, nil
}

func (s *ResourceService) GetDownloadPath(ctx context.Context, id int64) (*domain.Resource, string, error) {
	res, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, "", err
	}
	if res == nil {
		return nil, "", nil
	}
	// Note: This method signature implies returning a local path, which might not work for OSS.
	// Ideally, we should return a ReadCloser or a URL.
	// For now, let's keep it compatible with LocalStorage logic in Handler, 
	// but in a real OSS scenario, the Handler should use s.storage.Get() or s.storage.GetPublicURL().
	// We will refactor the Handler to use the Service's GetContent method instead.
	return res, res.Filename, nil
}

// New method to get file content
func (s *ResourceService) GetFileContent(ctx context.Context, id int64) (*domain.Resource, io.ReadCloser, error) {
	res, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if res == nil {
		return nil, nil, nil
	}
	
	reader, err := s.storage.Get(ctx, res.Filename)
	if err != nil {
		return nil, nil, err
	}
	return res, reader, nil
}


func (s *ResourceService) Review(ctx context.Context, id int64, status domain.ResourceStatus) error {
	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *ResourceService) CheckDuplicate(ctx context.Context, hash string) ([]domain.Resource, error) {
	return s.repo.GetByHash(ctx, hash)
}
