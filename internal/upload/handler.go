package upload

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) UploadFile(c *gin.Context) {
	bucket := c.PostForm("bucket")
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file field is required"})
		return
	}

	openedFile, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer openedFile.Close()

	file := &File{
		Name:        fileHeader.Filename,
		Content:     openedFile,
		Size:        fileHeader.Size,
		ContentType: fileHeader.Header.Get("Content-Type"),
	}

	url, err := h.service.UploadFile(c.Request.Context(), bucket, file)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"url": url})
}

func (h *Handler) UploadMultiple(c *gin.Context) {
	bucket := c.PostForm("bucket")
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart form"})
		return
	}

	filesHeaders := form.File["files"]
	if len(filesHeaders) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no files provided"})
		return
	}

	var filesToUpload []*File
	for _, header := range filesHeaders {
		openedFile, err := header.Open()
		if err != nil {
			continue
		}

		filesToUpload = append(filesToUpload, &File{
			Name:        header.Filename,
			Content:     openedFile,
			Size:        header.Size,
			ContentType: header.Header.Get("Content-Type"),
		})
	}

	defer func() {
		for _, f := range filesToUpload {
			f.Content.Close()
		}
	}()

	urls, err := h.service.UploadMultipleFiles(c.Request.Context(), bucket, filesToUpload)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"urls": urls})
}

func (h *Handler) GetPresignedURL(c *gin.Context) {
	bucket := c.Query("bucket")
	key := c.Query("key")

	url, err := h.service.GetDownloadURL(c.Request.Context(), bucket, key)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"presigned_url": url})
}

func (h *Handler) DownloadFile(c *gin.Context) {
	bucket := c.Query("bucket")
	key := c.Query("key")

	stream, err := h.service.DownloadFile(c.Request.Context(), bucket, key)
	if err != nil {
		h.handleError(c, err)
		return
	}
	defer stream.Close()

	c.Header("Content-Disposition", "attachment; filename="+key)
	c.Header("Content-Type", "application/octet-stream")

	_, _ = io.Copy(c.Writer, stream)
}

func (h *Handler) ListFiles(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	result, err := h.service.ListFiles(
		c.Request.Context(),
		c.Query("bucket"),
		c.Query("extension"),
		c.Query("token"),
		limit,
	)

	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) DeleteFile(c *gin.Context) {
	err := h.service.DeleteFile(c.Request.Context(), c.Query("bucket"), c.Query("key"))
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) GetBucketStats(c *gin.Context) {
	bucket := c.Query("bucket")
	if bucket == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bucket parameter is required"})
		return
	}

	stats, err := h.service.GetBucketStats(c.Request.Context(), bucket)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *Handler) CreateBucket(c *gin.Context) {
	var body struct {
		Name string `json:"bucket_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "valid bucket_name is required"})
		return
	}

	if err := h.service.CreateBucket(c.Request.Context(), body.Name); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusCreated)
}

func (h *Handler) ListBuckets(c *gin.Context) {
	buckets, err := h.service.ListAllBuckets(c.Request.Context())
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, buckets)
}

func (h *Handler) DeleteBucket(c *gin.Context) {
	if err := h.service.DeleteBucket(c.Request.Context(), c.Query("name")); err != nil {
		h.handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) EmptyBucket(c *gin.Context) {
	bucket := c.Query("bucket")
	if bucket == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bucket parameter is required"})
		return
	}

	err := h.service.EmptyBucket(c.Request.Context(), bucket)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidFileType),
		errors.Is(err, ErrBucketNameRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

	case errors.Is(err, ErrBucketAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})

	case errors.Is(err, ErrFileNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})

	case errors.Is(err, ErrOperationTimeout):
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "request timed out"})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "an unexpected error occurred"})
	}
}
