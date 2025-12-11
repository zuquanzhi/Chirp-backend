package service

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type AliyunOSSStorage struct {
	bucket *oss.Bucket
	domain string
}

func NewAliyunOSSStorage(endpoint, accessKeyID, accessKeySecret, bucketName string) (*AliyunOSSStorage, error) {
	client, err := oss.New(endpoint, accessKeyID, accessKeySecret)
	if err != nil {
		return nil, err
	}
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return nil, err
	}

	// Construct domain
	// Remove protocol from endpoint to ensure correct domain construction
	cleanEndpoint := strings.TrimPrefix(endpoint, "http://")
	cleanEndpoint = strings.TrimPrefix(cleanEndpoint, "https://")
	
	// Public URL: https://bucket-name.oss-cn-hangzhou.aliyuncs.com
	domain := fmt.Sprintf("https://%s.%s", bucketName, cleanEndpoint)

	return &AliyunOSSStorage{
		bucket: bucket,
		domain: domain,
	}, nil
}

func (s *AliyunOSSStorage) Save(ctx context.Context, file io.Reader, filename string) (string, int64, error) {
	// Calculate size if possible
	var size int64
	if seeker, ok := file.(io.Seeker); ok {
		// Seek to end to get size
		size, _ = seeker.Seek(0, io.SeekEnd)
		// Reset to beginning
		seeker.Seek(0, io.SeekStart)
	}

	// Upload
	err := s.bucket.PutObject(filename, file)
	if err != nil {
		return "", 0, fmt.Errorf("oss put object: %w", err)
	}

	return filename, size, nil
}

func (s *AliyunOSSStorage) Get(ctx context.Context, path string) (io.ReadCloser, error) {
	return s.bucket.GetObject(path)
}

func (s *AliyunOSSStorage) GetPublicURL(path string) string {
	return fmt.Sprintf("%s/%s", s.domain, path)
}
