# Go S3 Storage Manager API

A professional, high-performance REST API built in Go to manage AWS S3 resources. This project follows Clean Architecture principles and implements advanced cloud patterns such as Streaming, Presigned URLs, and Concurrent Uploads.

## Features

- File Management:
  - Single/Multiple Upload: Concurrent uploads using errgroup for maximum performance, returning clean URLs.
  - UUID v7 Integration: Files are renamed using time-ordered UUIDs for better indexing and collision avoidance.
  - Streaming Download: Direct binary streaming from S3 to the client to keep memory usage low.
  - Presigned URLs: Generate temporary, secure links for private file access.
  - Smart Listing: Filter by extension and retrieve formatted metadata (human-readable sizes, storage classes).
- Bucket Management:
  - Create, Delete, and List buckets.
  - Bucket Stats: Real-time calculation of total files and storage usage.
  - Empty Bucket: Efficiently remove all objects within a bucket.

- Security & Reliability:
  - MIME Type Detection: Security validation to prevent malicious file uploads.
  - Domain-Driven Errors: Clear, semantic HTTP responses (400, 404, 409, etc.).
  - Structured Logging: Uses slog for JSON-based observability.
  - Context & Timeouts: Built-in resilience to prevent hanging requests.

## Tech Stack

- Language: Go 1.23+
- Web Framework: Gin Gonic
- Cloud Provider: AWS SDK for Go v2
- Observability: log/slog (Structured Logging)
- Concurrency: golang.org/x/sync/errgroup
- Testing: testify and Mocks
- DevOps: Docker, Docker-Compose, GitHub Actions

## Configuration (.env)

For security reasons, sensitive credentials are not included in the repository. You must create a `.env` file in the root directory to run the project.

1. Create a file named .env in the root folder.
2. Add your AWS credentials and application settings as shown below:

```env
# Server Settings
PORT=8080
APP_ENV=development

# AWS Configuration
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your_access_key_here
AWS_SECRET_ACCESS_KEY=your_secret_key_here

# Timeouts
UPLOAD_TIMEOUT_SECONDS=60
```

**Note:** The .env file is ignored by Git (via .gitignore) to protect your cloud credentials.

## How to Run

### Using Docker (Recommended)

This will build the Go binary using a multi-stage build and start the container:

```bash
docker-compose up --build
```

### Locally

Ensure you have Go installed and your .env file configured:

```bash
go mod download
go run cmd/api/main.go
```

## Testing

The project includes unit tests for the service layer using mocks to simulate S3 behavior.

```bash
go test ./...
```

## API Endpoints Summary

### Files

| Method | Endpoint                 | Description                          |
|--------|--------------------------|--------------------------------------|
| POST   | /api/v1/upload           | Upload a single file (Form-data)     |
| POST   | /api/v1/upload-multiple  | Concurrent upload of several files   |
| GET    | /api/v1/list             | List files with extension filter     |
| GET    | /api/v1/download         | Stream file content directly         |
| GET    | /api/v1/presign          | Generate a temporary access URL      |
| DELETE | /api/v1/delete           | Remove a file from S3                |

### Buckets

| Method | Endpoint                | Description               |
|--------|-------------------------|---------------------------|
| GET    | /api/v1/buckets/list    | List all available buckets|
| POST   | /api/v1/buckets/create  | Create a new S3 bucket    |
| GET    | /api/v1/buckets/stats   | Get usage statistics      |
| DELETE | /api/v1/buckets/delete  | Remove a bucket           |
