# Setup Guide - Go OTEL API

Panduan lengkap untuk menjalankan service ini dengan baik, baik untuk development maupun production.

## 📋 Prerequisites

### 1. Development Environment

#### Required
- **Go 1.21+** - [Download](https://go.dev/dl/)
  ```bash
  go version  # Cek versi Go Anda
  ```

- **Docker** - [Download](https://www.docker.com/products/docker-desktop)
  ```bash
  docker --version
  docker-compose --version  # Optional tapi recommended
  ```

- **PostgreSQL 15+** - Bisa via Docker atau local installation
  ```bash
  # Via Docker (recommended untuk development)
  make dev-db-up

  # Atau install lokal
  # macOS: brew install postgresql@15
  # Linux: apt-get install postgresql-15
  # Windows: Download dari postgresql.org
  ```

- **Git** - Version control
  ```bash
  git --version
  ```

#### Optional (Recommended)
- **Make** - Build automation (biasanya sudah ada di macOS/Linux)
  ```bash
  make --version
  ```

- **golangci-lint** - Code linting
  ```bash
  # Install
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
  ```

- **Air** - Hot reload untuk development
  ```bash
  go install github.com/air-verse/air@latest
  ```

### 2. Production Environment (Kubernetes)

#### Required
- **Kubernetes Cluster** (salah satu):
  - Minikube (local testing)
  - Kind (Kubernetes in Docker)
  - Docker Desktop with Kubernetes
  - Cloud provider: GKE, EKS, AKS, etc.

  ```bash
  kubectl version --client
  ```

- **kubectl** - Kubernetes CLI
  ```bash
  kubectl version --client
  ```

- **Docker Registry** (untuk push images):
  - Docker Hub
  - GitHub Container Registry (GHCR)
  - AWS ECR / GCP GCR / Azure ACR
  - Harbor (self-hosted)

#### Optional (Recommended)
- **Helm** - Package manager untuk Kubernetes
  ```bash
  helm version
  ```

- **k9s** - Kubernetes TUI (Terminal UI)
  ```bash
  k9s version
  ```

- **Lens** - Kubernetes IDE (GUI)

### 3. Observability Stack (Optional)

- **OpenTelemetry Collector** - Untuk tracing
- **Jaeger** atau **Zipkin** - Tracing UI
- **Prometheus** - Metrics
- **Grafana** - Visualization
- **Loki** - Log aggregation

---

## 🚀 Setup Steps

### A. Development Setup (Local)

#### 1. Clone Repository

```bash
git clone <repository-url>
cd go-otel
```

#### 2. Install Dependencies

```bash
# Download Go dependencies
go mod download
go mod verify

# Atau gunakan make
make deps
```

#### 3. Setup Environment Variables

```bash
# Copy example env file
cp .env.example .env

# Edit dengan editor favorit
nano .env
# atau
code .env
```

**Minimal configuration untuk development:**

```env
# Server
APP_NAME=go-otel-api
APP_ENV=development
APP_PORT=8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=go_otel_db
DB_SSL_MODE=disable

# Connection Pool
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME=5m
DB_CONN_MAX_IDLE_TIME=10m

# OpenTelemetry (optional untuk development)
OTEL_ENABLED=false
OTEL_SERVICE_NAME=go-otel-api
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
OTEL_EXPORTER_OTLP_INSECURE=true

# Logging
LOG_LEVEL=debug
LOG_FORMAT=console  # Gunakan console untuk development
```

#### 4. Start PostgreSQL

**Option A: Via Docker (Recommended)**
```bash
make dev-db-up
```

**Option B: Via Docker Compose**
```bash
# Buat file docker-compose.yml
docker-compose up -d postgres
```

**Option C: Local Installation**
```bash
# macOS
brew services start postgresql@15

# Linux
sudo systemctl start postgresql

# Windows
# Start via Services atau pgAdmin
```

**Verify PostgreSQL is running:**
```bash
psql -h localhost -U postgres -d postgres -c "SELECT version();"
```

#### 5. Create Database

```bash
# Connect to PostgreSQL
psql -h localhost -U postgres

# Create database
CREATE DATABASE go_otel_db;

# Exit
\q
```

#### 6. Run Application

**Option A: Direct run**
```bash
go run cmd/api/main.go
```

**Option B: Via Make**
```bash
make run
```

**Option C: Build then run**
```bash
make build
./bin/go-otel-api
```

**Option D: Hot reload (dengan Air)**
```bash
# Install air jika belum
go install github.com/cosmtrek/air@latest

# Run with hot reload
air
```

#### 7. Test API

```bash
# Health check
curl http://localhost:8080/health/live

# Create user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com"}'

# List users
curl http://localhost:8080/api/v1/users
```

---

### B. Development Setup dengan OpenTelemetry (Optional)

#### 1. Start OpenTelemetry Collector

**Option A: Via Docker**
```bash
make dev-otel-up
```

**Option B: Via Docker Compose**
```bash
# Tambahkan ke docker-compose.yml
docker-compose up -d otel-collector
```

#### 2. Start Jaeger (untuk visualisasi tracing)

```bash
docker run -d --name jaeger \
  -p 16686:16686 \
  -p 4317:4317 \
  -p 4318:4318 \
  jaegertracing/all-in-one:latest
```

Access Jaeger UI: http://localhost:16686

#### 3. Update .env

```env
OTEL_ENABLED=true
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
```

#### 4. Restart Application

```bash
make run
```

---

### C. Docker Setup

#### 1. Build Docker Image

```bash
make docker-build
```

#### 2. Run via Docker

```bash
# Dengan .env file
docker run -p 8080:8080 --env-file .env \
  --network host \
  your-registry/go-otel-api:latest
```

#### 3. Run dengan Docker Compose

Buat file `docker-compose.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: go_otel_db
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  otel-collector:
    image: otel/opentelemetry-collector:latest
    command: ["--config=/etc/otel-collector-config.yaml"]
    ports:
      - "4317:4317"
      - "4318:4318"
    # volumes:
    #   - ./otel-collector-config.yaml:/etc/otel-collector-config.yaml

  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: postgres
      DB_PASSWORD: postgres
      DB_NAME: go_otel_db
      DB_SSL_MODE: disable
      OTEL_ENABLED: "true"
      OTEL_EXPORTER_OTLP_ENDPOINT: otel-collector:4317
      LOG_LEVEL: info
    depends_on:
      postgres:
        condition: service_healthy
      otel-collector:
        condition: service_started

volumes:
  postgres_data:
```

Run:
```bash
docker-compose up -d
```

---

### D. Kubernetes Setup (Production)

#### 1. Prerequisites Check

```bash
# Check kubectl
kubectl version --client

# Check cluster connection
kubectl cluster-info

# Check current context
kubectl config current-context
```

#### 2. Setup Kubernetes Cluster

**Option A: Minikube (Local Testing)**
```bash
# Install
brew install minikube  # macOS
# or download from minikube.sigs.k8s.io

# Start cluster
minikube start --cpus=4 --memory=8192

# Enable ingress
minikube addons enable ingress

# Enable metrics-server (untuk HPA)
minikube addons enable metrics-server
```

**Option B: Kind (Local Testing)**
```bash
# Install
brew install kind  # macOS

# Create cluster
kind create cluster --name go-otel-cluster
```

**Option C: Docker Desktop**
```bash
# Enable Kubernetes di Docker Desktop settings
# Preferences > Kubernetes > Enable Kubernetes
```

**Option D: Cloud Provider**
- GKE (Google): `gcloud container clusters create ...`
- EKS (AWS): `eksctl create cluster ...`
- AKS (Azure): `az aks create ...`

#### 3. Setup Container Registry

**Option A: Docker Hub**
```bash
docker login
```

**Option B: GitHub Container Registry**
```bash
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin
```

**Option C: Cloud Provider Registry**
```bash
# GCP
gcloud auth configure-docker

# AWS
aws ecr get-login-password --region region | docker login --username AWS --password-stdin aws_account_id.dkr.ecr.region.amazonaws.com

# Azure
az acr login --name myregistry
```

#### 4. Update Kubernetes Manifests

**Edit k8s/secret.yaml:**
```bash
nano k8s/secret.yaml
```

Ganti password:
```yaml
stringData:
  DB_USER: "postgres"
  DB_PASSWORD: "GANTI-DENGAN-PASSWORD-YANG-AMAN"
```

**Edit k8s/deployment.yaml:**
```bash
nano k8s/deployment.yaml
```

Update image registry:
```yaml
spec:
  template:
    spec:
      containers:
        - name: go-otel-api
          image: your-registry/go-otel-api:latest  # Ganti ini
```

**Edit k8s/ingress.yaml (jika menggunakan ingress):**
```bash
nano k8s/ingress.yaml
```

Update domain:
```yaml
spec:
  rules:
    - host: api.yourdomain.com  # Ganti dengan domain Anda
```

#### 5. Build and Push Image

```bash
# Build image
docker build -t your-registry/go-otel-api:v1.0.0 .

# Push to registry
docker push your-registry/go-otel-api:v1.0.0

# Tag as latest
docker tag your-registry/go-otel-api:v1.0.0 your-registry/go-otel-api:latest
docker push your-registry/go-otel-api:latest
```

#### 6. Deploy to Kubernetes

```bash
# Deploy semua resources
make k8s-deploy

# Atau manual:
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secret.yaml
kubectl apply -f k8s/postgres-statefulset.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/hpa.yaml
kubectl apply -f k8s/pdb.yaml
```

#### 7. Verify Deployment

```bash
# Check pods
kubectl get pods -l app=go-otel-api

# Check deployment
kubectl get deployment go-otel-api

# Check service
kubectl get svc go-otel-api-service

# Check HPA
kubectl get hpa go-otel-api-hpa

# Check logs
kubectl logs -l app=go-otel-api --tail=100 -f
```

#### 8. Access Application

**Port Forward (untuk testing):**
```bash
kubectl port-forward svc/go-otel-api-service 8080:80
```

Then access: http://localhost:8080

**Via LoadBalancer (jika ada):**
```bash
kubectl get svc go-otel-api-service
# Note external IP
```

**Via Ingress:**
```bash
# Deploy ingress
kubectl apply -f k8s/ingress.yaml

# Get ingress address
kubectl get ingress go-otel-api-ingress

# Minikube specific
minikube service go-otel-api-service --url
```

---

## 🔧 Configuration Tuning

### Database Connection Pool

Sesuaikan dengan jumlah replica dan database capacity:

**Rumus perhitungan:**
```
Total Max Connections = (Number of Pods) × (DB_MAX_OPEN_CONNS)
```

**Example:**
- 3 pods × 25 connections = 75 total connections
- 10 pods × 25 connections = 250 total connections (dengan HPA)

**PostgreSQL Configuration:**
```sql
-- Set max_connections lebih besar dari total
ALTER SYSTEM SET max_connections = 300;
SELECT pg_reload_conf();
```

**Adjust in k8s/configmap.yaml:**
```yaml
data:
  DB_MAX_OPEN_CONNS: "25"      # Per pod
  DB_MAX_IDLE_CONNS: "10"      # Per pod
  DB_CONN_MAX_LIFETIME: "5m"
  DB_CONN_MAX_IDLE_TIME: "10m"
```

### Resource Limits

Sesuaikan dengan kebutuhan aplikasi Anda di `k8s/deployment.yaml`:

```yaml
resources:
  requests:
    cpu: 100m      # Minimal CPU
    memory: 128Mi  # Minimal memory
  limits:
    cpu: 500m      # Maksimal CPU
    memory: 512Mi  # Maksimal memory
```

### HPA Configuration

Sesuaikan autoscaling di `k8s/hpa.yaml`:

```yaml
spec:
  minReplicas: 3    # Minimal pods
  maxReplicas: 10   # Maksimal pods
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70  # CPU threshold
```

---

## 🔍 Troubleshooting

### Common Issues

#### 1. Database Connection Failed
```bash
# Check database is running
kubectl get pods -l app=postgres

# Check logs
kubectl logs -l app=postgres

# Test connection
kubectl run -it --rm debug --image=postgres:15-alpine --restart=Never -- \
  psql -h postgres-service -U postgres -d go_otel_db
```

#### 2. Pods Not Starting
```bash
# Check pod status
kubectl describe pod <pod-name>

# Check events
kubectl get events --sort-by=.metadata.creationTimestamp

# Check logs
kubectl logs <pod-name>
```

#### 3. Image Pull Error
```bash
# Check image exists
docker images | grep go-otel-api

# Create image pull secret (jika private registry)
kubectl create secret docker-registry regcred \
  --docker-server=<registry-url> \
  --docker-username=<username> \
  --docker-password=<password> \
  --docker-email=<email>

# Add to deployment
# imagePullSecrets:
#   - name: regcred
```

#### 4. HPA Not Scaling
```bash
# Check metrics-server is running
kubectl get deployment metrics-server -n kube-system

# Check HPA status
kubectl describe hpa go-otel-api-hpa

# Install metrics-server jika belum ada
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
```

---

## 📊 Monitoring Setup

### 1. Prometheus + Grafana

```bash
# Install via Helm
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

helm install prometheus prometheus-community/kube-prometheus-stack
```

### 2. Jaeger (Tracing)

```bash
# Install via Helm
helm repo add jaegertracing https://jaegertracing.github.io/helm-charts
helm install jaeger jaegertracing/jaeger
```

### 3. Loki (Logging)

```bash
# Install via Helm
helm repo add grafana https://grafana.github.io/helm-charts
helm install loki grafana/loki-stack
```

---

## 🧪 Testing

### Unit Tests
```bash
make test
```

### Integration Tests
```bash
# Start dependencies
make dev-db-up

# Run integration tests
go test -v -tags=integration ./...
```

### Load Testing
```bash
# Install k6
brew install k6

# Run load test
k6 run loadtest.js
```

---

## 📚 Next Steps

1. ✅ Setup development environment
2. ✅ Run application locally
3. ✅ Setup Kubernetes cluster
4. ✅ Deploy to Kubernetes
5. ⏭️ Setup monitoring (Prometheus/Grafana)
6. ⏭️ Setup logging (Loki/ELK)
7. ⏭️ Setup CI/CD pipeline
8. ⏭️ Configure backup strategy
9. ⏭️ Setup disaster recovery
10. ⏭️ Performance tuning

---

## 🆘 Need Help?

- Check [README.md](README.md) untuk dokumentasi lengkap
- Check logs: `make k8s-logs`
- Check status: `make k8s-status`
- Open issue di repository

---

## 📖 References

- [Echo Framework](https://echo.labstack.com/)
- [GORM](https://gorm.io/)
- [OpenTelemetry Go](https://opentelemetry.io/docs/instrumentation/go/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
