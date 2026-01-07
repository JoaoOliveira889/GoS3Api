package upload

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Repository struct {
	client *s3.Client
	region string
}

func NewS3Repository(client *s3.Client, region string) Repository {
	return &S3Repository{
		client: client,
		region: region,
	}
}

func (r *S3Repository) Upload(ctx context.Context, bucket string, file *File) (string, error) {
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(file.Name),
		Body:   file.Content,
	}

	_, err := r.client.PutObject(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to upload: %w", err)
	}

	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, r.region, file.Name), nil
}

func (r *S3Repository) List(ctx context.Context, bucket, prefix, token string, limit int32) (*PaginatedFiles, error) {
	input := &s3.ListObjectsV2Input{
		Bucket:            aws.String(bucket),
		Prefix:            aws.String(prefix),
		ContinuationToken: aws.String(token),
		MaxKeys:           aws.Int32(limit),
	}

	if token == "" {
		input.ContinuationToken = nil
	}

	output, err := r.client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	var files []FileSummary
	for _, obj := range output.Contents {
		key := aws.ToString(obj.Key)
		size := aws.ToInt64(obj.Size)
		files = append(files, FileSummary{
			Key:               key,
			Size:              size,
			HumanReadableSize: formatBytes(size),
			StorageClass:      string(obj.StorageClass),
			LastModified:      aws.ToTime(obj.LastModified),
			Extension:         strings.ToLower(filepath.Ext(key)),
			URL:               fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, r.region, key),
		})
	}

	next := ""
	if output.NextContinuationToken != nil {
		next = *output.NextContinuationToken
	}

	return &PaginatedFiles{Files: files, NextToken: next}, nil
}

func (r *S3Repository) Delete(ctx context.Context, bucket, key string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

func (r *S3Repository) Download(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	output, err := r.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return output.Body, nil
}

func (r *S3Repository) GetPresignURL(ctx context.Context, bucket, key string, exp time.Duration) (string, error) {
	pc := s3.NewPresignClient(r.client)
	req, err := pc.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(exp))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

func (r *S3Repository) CheckBucketExists(ctx context.Context, bucket string) (bool, error) {
	_, err := r.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(bucket)})
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (r *S3Repository) CreateBucket(ctx context.Context, bucket string) error {
	_, err := r.client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)})
	return err
}

func (r *S3Repository) ListBuckets(ctx context.Context) ([]BucketSummary, error) {
	out, err := r.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}
	var res []BucketSummary
	for _, b := range out.Buckets {
		res = append(res, BucketSummary{Name: aws.ToString(b.Name), CreationDate: aws.ToTime(b.CreationDate)})
	}
	return res, nil
}

func (r *S3Repository) DeleteBucket(ctx context.Context, bucket string) error {
	_, err := r.client.DeleteBucket(ctx, &s3.DeleteBucketInput{Bucket: aws.String(bucket)})
	return err
}

func (r *S3Repository) DeleteAll(ctx context.Context, bucket string) error {
	out, err := r.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: aws.String(bucket)})
	if err != nil || len(out.Contents) == 0 {
		return err
	}
	var objects []types.ObjectIdentifier
	for _, obj := range out.Contents {
		objects = append(objects, types.ObjectIdentifier{Key: obj.Key})
	}
	_, err = r.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(bucket),
		Delete: &types.Delete{Objects: objects},
	})
	return err
}

func (r *S3Repository) GetStats(ctx context.Context, bucket string) (*BucketStats, error) {
	out, err := r.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: aws.String(bucket)})
	if err != nil {
		return nil, err
	}
	var totalSize int64
	for _, obj := range out.Contents {
		totalSize += aws.ToInt64(obj.Size)
	}
	return &BucketStats{
		BucketName:         bucket,
		TotalFiles:         int(len(out.Contents)),
		TotalSizeBytes:     totalSize,
		TotalSizeFormatted: formatBytes(totalSize),
	}, nil
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
