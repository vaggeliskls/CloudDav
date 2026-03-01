CMD := ./cmd/server
GOLANGCI_LINT := $(shell go env GOPATH)/bin/golangci-lint

.DEFAULT_GOAL := help

.PHONY: run
run: ## Run the server locally (loads .env)
	set -a && . ./.env && set +a && go run $(CMD)

.PHONY: minio-up
minio-up: ## Start a local MinIO instance (S3-compatible)
	docker run -d --name minio \
		-p 9000:9000 -p 9001:9001 \
		-e MINIO_ROOT_USER=minioadmin \
		-e MINIO_ROOT_PASSWORD=minioadmin \
		minio/minio server /data --console-address ":9001"
	@echo "MinIO API → http://localhost:9000"
	@echo "MinIO UI  → http://localhost:9001  (minioadmin / minioadmin)"

.PHONY: minio-down
minio-down: ## Stop and remove the local MinIO instance
	docker rm -f minio

.PHONY: test
test: ## Run all tests
	go test -race -count=1 ./...

.PHONY: deps
deps: ## Download Go module dependencies
	go mod download

.PHONY: deps-update
deps-update: ## Upgrade all dependencies to latest and tidy
	go get -u ./...
	go mod tidy

.PHONY: lint-install
lint-install: ## Install golangci-lint (dev only)
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: quality
quality: ## Run linter, vet, and tests (mirrors the CI quality workflow)
	go vet ./...
	$(GOLANGCI_LINT) run ./...
	go test -race -count=1 -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

.PHONY: help
help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\n"} \
		/^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
