.PHONY: help build run test clean docker-up docker-down ent-generate deps dev-all dev-frontend dev-backend

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

deps: ## Download dependencies
	go mod download
	go mod tidy

build: ## Build the server binary
	go build -o opendq-server ./cmd/server

run: ## Run the server
	go run ./cmd/server

test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	rm -f opendq-server
	rm -f coverage.out coverage.html

docker-up: ## Start Docker Compose services
	docker-compose up -d

docker-down: ## Stop Docker Compose services
	docker-compose down

docker-logs: ## View Docker Compose logs
	docker-compose logs -f

ent-generate: ## Generate Ent code
	go generate ./ent

fmt: ## Format code
	go fmt ./...

lint: ## Run linter
	golangci-lint run

vet: ## Run go vet
	go vet ./...

mod-update: ## Update dependencies
	go get -u ./...
	go mod tidy

dev: docker-up ## Start development environment (infrastructure only)
	@echo "Waiting for services to be ready..."
	@sleep 15
	@echo ""
	@echo "=========================================="
	@echo "  Development Environment Ready!"
	@echo "=========================================="
	@echo ""
	@echo "Infrastructure Services:"
	@echo "  PostgreSQL:        localhost:5432"
	@echo "  Redis:             localhost:6379"
	@echo ""
	@echo "Web UIs:"
	@echo "  OpenFGA Playground: http://localhost:3002"
	@echo "  Keycloak Admin:     http://localhost:8180 (admin/admin)"
	@echo "  Marquez Web:        http://localhost:3001"
	@echo ""
	@echo "APIs:"
	@echo "  OpenFGA API:        http://localhost:8081"
	@echo "  Marquez API:        http://localhost:5000"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Run 'make run' to start the backend"
	@echo "  2. Run 'make dev-frontend' to start the frontend"
	@echo ""

dev-backend: ## Start the backend server
	@echo "Starting OpenDQ backend on http://localhost:8080..."
	go run ./cmd/server

dev-frontend: ## Start the frontend development server
	@echo "Starting OpenDQ frontend on http://localhost:5173..."
	cd frontend && npm install && npm run dev

dev-all: docker-up ## Start all services (infrastructure + backend + frontend)
	@echo "Waiting for infrastructure services to be ready..."
	@sleep 15
	@echo ""
	@echo "=========================================="
	@echo "  Starting All OpenDQ Services"
	@echo "=========================================="
	@echo ""
	@echo "Infrastructure: Running"
	@echo ""
	@echo "Starting backend and frontend..."
	@echo "(Use Ctrl+C to stop, then 'make docker-down' to stop infrastructure)"
	@echo ""
	@echo "Web UIs available at:"
	@echo "  OpenDQ Frontend:    http://localhost:5173"
	@echo "  OpenDQ API:         http://localhost:8080"
	@echo "  OpenFGA Playground: http://localhost:3002"
	@echo "  Keycloak Admin:     http://localhost:8180 (admin/admin)"
	@echo "  Marquez Web:        http://localhost:3001"
	@echo ""
	@$(MAKE) -j2 dev-backend dev-frontend

install-tools: ## Install development tools
	go install entgo.io/ent/cmd/ent@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
