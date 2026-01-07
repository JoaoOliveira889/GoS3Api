package upload

import (
	"context"
	"io"
	"time"
)

type Repository interface {
	Upload(ctx context.Context, bucket string, file *File) (string, error)
	GetPresignURL(ctx context.Context, bucket, key string, expiration time.Duration) (string, error)
	Download(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	List(ctx context.Context, bucket, prefix, token string, limit int32) (*PaginatedFiles, error)
	Delete(ctx context.Context, bucket string, key string) error
	CheckBucketExists(ctx context.Context, bucket string) (bool, error)
	CreateBucket(ctx context.Context, bucket string) error
	ListBuckets(ctx context.Context) ([]BucketSummary, error)
	GetStats(ctx context.Context, bucket string) (*BucketStats, error)
	DeleteAll(ctx context.Context, bucket string) error
	DeleteBucket(ctx context.Context, bucket string) error
}
