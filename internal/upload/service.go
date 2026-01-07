package upload

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type Service interface {
	UploadFile(ctx context.Context, bucket string, file *File) (string, error)
	UploadMultipleFiles(ctx context.Context, bucket string, files []*File) ([]string, error)
	GetDownloadURL(ctx context.Context, bucket, key string) (string, error)
	DownloadFile(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	ListFiles(ctx context.Context, bucket, ext, token string, limit int) (*PaginatedFiles, error)
	DeleteFile(ctx context.Context, bucket string, key string) error
	GetBucketStats(ctx context.Context, bucket string) (*BucketStats, error)
	CreateBucket(ctx context.Context, bucket string) error
	ListAllBuckets(ctx context.Context) ([]BucketSummary, error)
	DeleteBucket(ctx context.Context, bucket string) error
	EmptyBucket(ctx context.Context, bucket string) error
}

const (
	uploadTimeout       = 60 * time.Second
	deleteTimeout       = 5 * time.Second
	maxBucketNameLength = 63
	minBucketNameLength = 3
)

var (
	bucketDNSNameRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$`)
	allowedTypes       = map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"application/pdf": true,
	}
)

type uploadService struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &uploadService{repo: repo}
}

func (s *uploadService) UploadFile(ctx context.Context, bucket string, file *File) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, uploadTimeout)
	defer cancel()

	if err := s.validateBucketName(bucket); err != nil {
		return "", err
	}

	if err := s.validateFile(file); err != nil {
		slog.Error("security validation failed", "error", err, "filename", file.Name)
		return "", err
	}

	id, err := uuid.NewV7()
	if err != nil {
		slog.Error("uuid generation failed", "error", err)
		return "", fmt.Errorf("failed to generate unique id: %w", err)
	}

	file.Name = id.String() + filepath.Ext(file.Name)

	url, err := s.repo.Upload(ctx, bucket, file)
	if err != nil {
		slog.Error("repository upload failed", "error", err, "bucket", bucket)
		return "", err
	}

	file.URL = url
	slog.Info("file uploaded successfully", "url", url)
	return url, nil
}

func (s *uploadService) UploadMultipleFiles(ctx context.Context, bucket string, files []*File) ([]string, error) {
	g, ctx := errgroup.WithContext(ctx)
	results := make([]string, len(files))

	for i, f := range files {
		i, f := i, f
		g.Go(func() error {
			url, err := s.UploadFile(ctx, bucket, f)
			if err != nil {
				return err
			}
			results[i] = url
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return results, nil
}

func (s *uploadService) GetDownloadURL(ctx context.Context, bucket, key string) (string, error) {
	if err := s.validateBucketName(bucket); err != nil {
		return "", err
	}

	return s.repo.GetPresignURL(ctx, bucket, key, 15*time.Minute)
}

func (s *uploadService) DownloadFile(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	if err := s.validateBucketName(bucket); err != nil {
		return nil, err
	}
	return s.repo.Download(ctx, bucket, key)
}

func (s *uploadService) ListFiles(ctx context.Context, bucket, ext, token string, limit int) (*PaginatedFiles, error) {
	if err := s.validateBucketName(bucket); err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 10
	}

	res, err := s.repo.List(ctx, bucket, "", token, int32(limit))
	if err != nil {
		return nil, err
	}

	if ext == "" {
		return res, nil
	}

	var filtered []FileSummary
	target := strings.ToLower(ext)

	if !strings.HasPrefix(target, ".") {
		target = "." + target
	}

	for _, f := range res.Files {
		if strings.ToLower(f.Extension) == target {
			filtered = append(filtered, f)
		}
	}

	res.Files = filtered
	return res, nil
}

func (s *uploadService) DeleteFile(ctx context.Context, bucket string, key string) error {
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	if key == "" {
		return fmt.Errorf("file key is required")
	}

	if err := s.validateBucketName(bucket); err != nil {
		return err
	}

	return s.repo.Delete(ctx, bucket, key)
}

func (s *uploadService) GetBucketStats(ctx context.Context, bucket string) (*BucketStats, error) {
	if err := s.validateBucketName(bucket); err != nil {
		return nil, err
	}
	return s.repo.GetStats(ctx, bucket)
}

func (s *uploadService) CreateBucket(ctx context.Context, bucket string) error {

	if err := s.validateBucketName(bucket); err != nil {
		return err
	}

	exists, err := s.repo.CheckBucketExists(ctx, bucket)
	if err != nil {
		return err
	}
	if exists {
		return ErrBucketAlreadyExists
	}

	return s.repo.CreateBucket(ctx, bucket)
}

func (s *uploadService) DeleteBucket(ctx context.Context, bucket string) error {
	if err := s.validateBucketName(bucket); err != nil {
		return err
	}
	return s.repo.DeleteBucket(ctx, bucket)
}

func (s *uploadService) EmptyBucket(ctx context.Context, bucket string) error {
	if err := s.validateBucketName(bucket); err != nil {
		return err
	}
	return s.repo.DeleteAll(ctx, bucket)
}

func (s *uploadService) ListAllBuckets(ctx context.Context) ([]BucketSummary, error) {
	return s.repo.ListBuckets(ctx)
}

func (s *uploadService) validateBucketName(bucket string) error {
	bucket = strings.TrimSpace(strings.ToLower(bucket))
	if bucket == "" {
		return ErrBucketNameRequired
	}

	if len(bucket) < minBucketNameLength || len(bucket) > maxBucketNameLength {
		return fmt.Errorf("bucket name length must be between %d and %d", minBucketNameLength, maxBucketNameLength)
	}

	if !bucketDNSNameRegex.MatchString(bucket) {
		return fmt.Errorf("invalid bucket name pattern")
	}

	if strings.Contains(bucket, "..") {
		return fmt.Errorf("bucket name cannot contain consecutive dots")
	}

	return nil
}

func (s *uploadService) validateFile(f *File) error {
	seeker, ok := f.Content.(io.Seeker)
	if !ok {
		return fmt.Errorf("file content must support seeking")
	}

	buffer := make([]byte, 512)
	n, err := f.Content.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read file header: %w", err)
	}

	if _, err := seeker.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to reset file pointer: %w", err)
	}

	detectedType := http.DetectContentType(buffer[:n])
	if !allowedTypes[detectedType] {
		slog.Warn("rejected file type", "type", detectedType)
		return ErrInvalidFileType
	}

	return nil
}
