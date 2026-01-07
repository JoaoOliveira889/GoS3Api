package upload

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type readSeekCloser struct {
	*strings.Reader
}

func (rsc readSeekCloser) Close() error { return nil }

func TestUploadFile_InvalidBucket(t *testing.T) {
	mockRepo := new(RepositoryMock)
	service := NewService(mockRepo)

	result, err := service.UploadFile(context.Background(), "", &File{})

	assert.Error(t, err)

	assert.Empty(t, result)

	assert.ErrorIs(t, err, ErrBucketNameRequired)
}

func TestUploadFile_Success(t *testing.T) {
	mockRepo := new(RepositoryMock)
	service := NewService(mockRepo)
	ctx := context.Background()

	content := strings.NewReader("\x89PNG\r\n\x1a\n" + strings.Repeat("0", 512))
	file := &File{
		Name:    "test-image.png",
		Content: readSeekCloser{content},
	}

	bucket := "my-test-bucket"
	expectedURL := "https://s3.amazonaws.com/my-test-bucket/unique-id.png"

	mockRepo.On("Upload", mock.Anything, bucket, mock.AnythingOfType("*upload.File")).Return(expectedURL, nil)

	resultURL, err := service.UploadFile(ctx, bucket, file)

	assert.NoError(t, err)
	assert.NotEmpty(t, resultURL)
	assert.Equal(t, expectedURL, resultURL)

	mockRepo.AssertExpectations(t)
}

func TestGetDownloadURL_Success(t *testing.T) {
	mockRepo := new(RepositoryMock)
	service := NewService(mockRepo)

	bucket := "my-bucket"
	key := "image.png"
	expectedPresignedURL := "https://s3.amazonaws.com/my-bucket/image.png?signed=true"

	mockRepo.On("GetPresignURL", mock.Anything, bucket, key, 15*time.Minute).
		Return(expectedPresignedURL, nil)

	url, err := service.GetDownloadURL(context.Background(), bucket, key)

	assert.NoError(t, err)
	assert.Equal(t, expectedPresignedURL, url)
	mockRepo.AssertExpectations(t)
}
