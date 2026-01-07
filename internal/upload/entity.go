package upload

import (
	"io"
	"time"
)

type File struct {
	Name        string            `json:"name"`
	URL         string            `json:"url"`
	Content     io.ReadSeekCloser `json:"-"`
	Size        int64             `json:"size"`
	ContentType string            `json:"content_type"`
}

type FileSummary struct {
	Key               string    `json:"key"`
	URL               string    `json:"url"`
	Size              int64     `json:"size_bytes"`
	HumanReadableSize string    `json:"size_formatted"`
	Extension         string    `json:"extension"`
	StorageClass      string    `json:"storage_class"`
	LastModified      time.Time `json:"last_modified"`
}

type BucketStats struct {
	BucketName         string `json:"bucket_name"`
	TotalFiles         int    `json:"total_files"`
	TotalSizeBytes     int64  `json:"total_size_bytes"`
	TotalSizeFormatted string `json:"total_size_formatted"`
}

type BucketSummary struct {
	Name         string    `json:"name"`
	CreationDate time.Time `json:"creation_date"`
}

type PaginatedFiles struct {
	Files     []FileSummary `json:"files"`
	NextToken string        `json:"next_token,omitempty"`
}
