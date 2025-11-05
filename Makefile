.PHONY: help build run test clean docker-build docker-push k8s-deploy k8s-delete

# Variables
APP_NAME=go-otel
DOCKER_IMAGE=your-registry/$(APP_NAME)
DOCKER_TAG=latest
GO_VERSION=1.21

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Download dependencies
	go mod download
	go mod verify

tidy: ## Tidy dependencies
	go mod tidy

build: ## Build the application
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/$(APP_NAME) ./cmd/api

run: ## Run the application locally
	go run ./cmd/api/main.go

test: ## Run tests
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests with coverage report
	go tool cover -html=coverage.out

lint: ## Run linter
	golangci-lint run ./...

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out

# Docker targets
docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-push: docker-build ## Push Docker image to registry
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-run: ## Run Docker container locally
	docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG)

# Kubernetes targets
k8s-apply-config: ## Apply ConfigMap and Secret
	kubectl apply -f k8s/configmap.yaml
	kubectl apply -f k8s/secret.yaml

k8s-apply-postgres: ## Deploy PostgreSQL
	kubectl apply -f k8s/postgres-statefulset.yaml

k8s-apply-app: ## Deploy application
	kubectl apply -f k8s/deployment.yaml
	kubectl apply -f k8s/service.yaml
	kubectl apply -f k8s/hpa.yaml
	kubectl apply -f k8s/pdb.yaml

k8s-deploy: k8s-apply-config k8s-apply-postgres k8s-apply-app ## Deploy everything to Kubernetes

k8s-delete: ## Delete all resources from Kubernetes
	kubectl delete -f k8s/deployment.yaml --ignore-not-found=true
	kubectl delete -f k8s/service.yaml --ignore-not-found=true
	kubectl delete -f k8s/hpa.yaml --ignore-not-found=true
	kubectl delete -f k8s/pdb.yaml --ignore-not-found=true
	kubectl delete -f k8s/postgres-statefulset.yaml --ignore-not-found=true
	kubectl delete -f k8s/configmap.yaml --ignore-not-found=true
	kubectl delete -f k8s/secret.yaml --ignore-not-found=true

k8s-status: ## Check deployment status
	kubectl get pods -l app=$(APP_NAME)
	kubectl get svc $(APP_NAME)-service
	kubectl get hpa $(APP_NAME)-hpa

k8s-logs: ## Show application logs
	kubectl logs -l app=$(APP_NAME) --tail=100 -f

k8s-restart: ## Restart deployment
	kubectl rollout restart deployment/$(APP_NAME)

# Development targets
dev-setup: ## Setup development environment
	cp .env.example .env
	@echo "Please edit .env file with your configuration"

dev-db-up: ## Start local PostgreSQL with Docker
	docker run -d \
		--name go-otel-postgres \
		-e POSTGRES_DB=go_otel_db \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=postgres \
		-p 5432:5432 \
		postgres:15-alpine

dev-db-down: ## Stop local PostgreSQL
	docker stop go-otel-postgres
	docker rm go-otel-postgres

dev-otel-up: ## Start local OpenTelemetry collector
	docker run -d \
		--name otel-collector \
		-p 4317:4317 \
		-p 4318:4318 \
		otel/opentelemetry-collector:latest

dev-otel-down: ## Stop OpenTelemetry collector
	docker stop otel-collector
	docker rm otel-collector

# Docker Compose targets
compose-up: ## Start all services with docker-compose
	docker-compose up -d

compose-down: ## Stop all services
	docker-compose down

compose-logs: ## Show logs from all services
	docker-compose logs -f

compose-logs-app: ## Show application logs only
	docker-compose logs -f app

compose-ps: ## List running services
	docker-compose ps

compose-restart: ## Restart all services
	docker-compose restart

compose-build: ## Rebuild and start services
	docker-compose up -d --build

compose-clean: ## Remove all containers, volumes, and networks
	docker-compose down -v --remove-orphans
