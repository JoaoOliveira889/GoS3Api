package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	// Internal packages
	appConfig "github.com/JoaoOliveira889/s3-api/internal/config"
	"github.com/JoaoOliveira889/s3-api/internal/middleware"
	"github.com/JoaoOliveira889/s3-api/internal/upload"
	"github.com/gin-gonic/gin"

	// External packages
	configAWS "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg := appConfig.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	r := gin.New()

	r.Use(middleware.RequestTimeoutMiddleware(cfg.UploadTimeout))
	r.Use(middleware.LoggingMiddleware())
	r.Use(gin.Recovery())

	ctx := context.Background()
	awsCfg, err := configAWS.LoadDefaultConfig(ctx, configAWS.WithRegion(cfg.AWSRegion))
	if err != nil {
		slog.Error("failed to load AWS SDK config", "error", err)
		os.Exit(1)
	}

	s3Client := s3.NewFromConfig(awsCfg)
	repo := upload.NewS3Repository(s3Client, cfg.AWSRegion)
	service := upload.NewService(repo)
	handler := upload.NewHandler(service)

	api := r.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":    "healthy",
				"env":       cfg.Env,
				"timestamp": time.Now().Format(time.RFC3339),
			})
		})

		api.GET("/list", handler.ListFiles)
		api.POST("/upload", handler.UploadFile)
		api.POST("/upload-multiple", handler.UploadMultiple)
		api.GET("/download", handler.DownloadFile)
		api.GET("/presign", handler.GetPresignedURL)
		api.DELETE("/delete", handler.DeleteFile)

		buckets := api.Group("/buckets")
		{
			buckets.POST("/create", handler.CreateBucket)
			buckets.DELETE("/delete", handler.DeleteBucket)
			buckets.GET("/stats", handler.GetBucketStats)
			buckets.GET("/list", handler.ListBuckets)
			buckets.DELETE("/empty", handler.EmptyBucket)
		}
	}

	slog.Info("server successfully started",
		"port", cfg.Port,
		"env", cfg.Env,
		"region", cfg.AWSRegion,
	)

	if err := r.Run(":" + cfg.Port); err != nil {
		slog.Error("server failed to shut down properly", "error", err)
		os.Exit(1)
	}
}
