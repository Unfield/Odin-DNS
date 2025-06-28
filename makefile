.PHONY: swagger-gen swagger-serve swagger-validate clean build run dev test deps

SWAGGER_SOURCES := $(shell find . -name "*.go" -not -path "./docs/*" -not -path "./vendor/*")
SWAGGER_OUTPUT := docs/docs.go

$(SWAGGER_OUTPUT): $(SWAGGER_SOURCES)
	@echo "Regenerating Swagger documentation..."
	@swag init -g cmd/odin-dns/main.go -o docs --parseInternal --parseDependency

swagger-gen: $(SWAGGER_OUTPUT)

dev: $(SWAGGER_OUTPUT)
	@echo "Starting development server..."
	@go run cmd/odin-dns/main.go

swagger-force:
	@echo "Force regenerating Swagger documentation..."
	@swag init -g cmd/odin-dns/main.go -o docs --parseInternal --parseDependency

build: $(SWAGGER_OUTPUT)
	@echo "Building application..."
	@go build -o bin/odin-dns cmd/odin-dns/main.go

run: build
	@echo "Running application..."
	@./bin/odin-dns

deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

test:
	@go test ./...

clean:
	@echo "Cleaning up..."
	@rm -rf docs/ bin/

install-swagger:
	@go install github.com/swaggo/swag/cmd/swag@latest

help:
	@echo "Available commands:"
	@echo "  dev             - Start development server (smart swagger regeneration)"
	@echo "  swagger-gen     - Generate Swagger docs (only if needed)"
	@echo "  swagger-force   - Force regenerate Swagger docs"
	@echo "  build           - Build application"
	@echo "  run             - Build and run application"
	@echo "  test            - Run tests"
	@echo "  deps            - Install dependencies"
	@echo "  clean           - Clean generated files"
	@echo "  install-swagger - Install swag CLI"
	@echo "  help            - Show this help"
