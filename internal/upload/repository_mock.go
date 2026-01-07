package upload

import (
	"context"
	"io"
	"time"

	"github.com/stretchr/testify/mock"
)

type RepositoryMock struct {
	mock.Mock
}

func (m *RepositoryMock) Delete(ctx context.Context, bucket string, key string) error {
	panic("unimplemented")
}

func (m *RepositoryMock) DeleteAll(ctx context.Context, bucket string) error {
	panic("unimplemented")
}

func (m *RepositoryMock) DeleteBucket(ctx context.Context, bucket string) error {
	panic("unimplemented")
}

func (m *RepositoryMock) GetStats(ctx context.Context, bucket string) (*BucketStats, error) {
	panic("unimplemented")
}

func (m *RepositoryMock) ListBuckets(ctx context.Context) ([]BucketSummary, error) {
	panic("unimplemented")
}

func (m *RepositoryMock) Upload(ctx context.Context, bucket string, file *File) (string, error) {
	args := m.Called(ctx, bucket, file)
	return args.String(0), args.Error(1)
}

func (m *RepositoryMock) Download(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	args := m.Called(ctx, bucket, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *RepositoryMock) GetPresignURL(ctx context.Context, bucket, key string, expiration time.Duration) (string, error) {
	args := m.Called(ctx, bucket, key, expiration)
	return args.String(0), args.Error(1)
}

func (m *RepositoryMock) List(ctx context.Context, bucket, prefix, token string, limit int32) (*PaginatedFiles, error) {
	args := m.Called(ctx, bucket, prefix, token, limit)
	return args.Get(0).(*PaginatedFiles), args.Error(1)
}

func (m *RepositoryMock) CheckBucketExists(ctx context.Context, bucket string) (bool, error) {
	args := m.Called(ctx, bucket)
	return args.Bool(0), args.Error(1)
}

func (m *RepositoryMock) CreateBucket(ctx context.Context, bucket string) error {
	args := m.Called(ctx, bucket)
	return args.Error(0)
}
