# Go S3 Storage Manager API

> **Article:** [Go S3 API: Managing Cloud Storage with Clean Architecture](https://joaooliveira.net/en/blog/2026/01/go-s3-api/)

A professional, high-performance REST API built in Go to manage AWS S3 resources. This project follows Clean Architecture principles and implements advanced cloud patterns such as binary streaming, concurrent uploads, and secure access management.

---

## What This Project Covers

* File Management: Single/multiple uploads, streaming downloads, and secure deletions.
* Bucket Operations: CRUD operations for buckets, including usage statistics and efficient "empty bucket" logic.
* Performance: Concurrent processing using errgroup and memory-efficient streaming.
* Security: MIME type validation, UUID v7 file renaming, and temporary Presigned URLs.
* Observability: Structured JSON logging with slog.
* Context & Timeouts: Built-in resilience to prevent hanging requests.

## Concepts Overview

### Clean Architecture

Decouples the business logic from external dependencies (AWS SDK, Gin Gonic), making the code testable and maintainable.

### Streaming & Concurrency

* Streaming: Direct binary transfer from S3 to the client, keeping RAM usage minimal even for large files.
* Concurrency: Uses golang.org/x/sync/errgroup to handle multiple uploads in parallel with built-in error propagation.

### Data Integrity

* UUID v7: Implementation of time-ordered UUIDs for optimized database indexing and file collision avoidance.
* Domain-Driven Errors: Semantic error handling that maps internal logic issues to clear HTTP responses.

## Tech Stack

* Language: Go 1.25+
* Framework: Gin Gonic (Web)
* Cloud: AWS SDK for Go v2
* Concurrency: errgroup
* Testing: Testify & Mocks
* Containerization: Docker & Docker Compose

## Environment Setup

Create a .env file in the root directory:

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

## How to Run

Using Docker (Recommended)

```bash
docker-compose up --build
```

Locally

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

## About

This repository is part of my technical writing and learning notes.  
If you found it useful, consider starring the repo and sharing feedback.

- Author: Joao Oliveira
- Blog: https://joaooliveira.net
- Topics: .NET, Redis, backend engineering, system design

## Contributing

Issues and pull requests are welcome.  
If you plan a larger change, please open an issue first so we can align on scope.

## License

Licensed under the **MIT License**. See the `LICENSE` file for details.
