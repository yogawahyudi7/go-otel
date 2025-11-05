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

### 3. Observability Stack (Included in Docker Compose)

- **Jaeger** - Distributed tracing (dengan OTLP support built-in)
- **Loki** - Log aggregation dan storage
- **Promtail** - Log collector
- **Grafana** - Unified visualization untuk logs dan traces

**Optional untuk production:**
- **Prometheus** - Metrics collection (jika ingin menambahkan metrics)

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
APP_NAME=go-otel
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
OTEL_SERVICE_NAME=go-otel
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
./bin/go-otel
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

### B. Development Setup dengan Full Observability Stack (Recommended)

#### 1. Start Full Stack dengan Docker Compose

```bash
# Start semua services: App, PostgreSQL, Jaeger, Loki, Promtail, Grafana
docker-compose up -d
```

Ini akan menjalankan:
- **PostgreSQL** - Database (port 5432)
- **Jaeger** - Tracing UI (port 16686)
- **Loki** - Log storage (port 3100)
- **Promtail** - Log collector
- **Grafana** - Unified dashboard (port 3000)
- **App** - Go application (port 8080)

#### 2. Access Observability UIs

**Jaeger (Distributed Tracing):**
- URL: http://localhost:16686
- Purpose: View request traces, latency, dependencies

**Grafana (Logs + Traces):**
- URL: http://localhost:3000
- Username: `admin`
- Password: `admin`
- Purpose: Query logs dari Loki, view traces dari Jaeger

**Loki API (Direct):**
- URL: http://localhost:3100
- Purpose: Direct LogQL queries (biasanya via Grafana)

#### 3. Verify Observability Stack

```bash
# Check all services running
docker-compose ps

# Check logs being collected
curl http://localhost:3100/loki/api/v1/labels

# Test tracing
curl http://localhost:8080/api/v1/users
# Then check Jaeger UI for traces

# Test logging in Grafana
# Open Grafana > Explore > Select Loki datasource
# Query: {service="app"}
```

#### 4. Query Logs di Grafana

**Example LogQL Queries:**

```logql
# All logs dari aplikasi
{service="app"}

# Error logs only
{service="app"} |= "level=error"

# Logs dari specific container
{container="go-otel"}

# Rate of errors
rate({service="app"} |= "error" [5m])

# Correlation: Find logs by trace_id
{service="app"} |= "trace_id=abc123"
```

#### 5. Correlation: Logs ↔ Traces

Grafana sudah dikonfigurasi untuk **automatic correlation**:

- **Dari Trace → Logs**: Click trace di Jaeger, lihat related logs
- **Dari Logs → Trace**: Click trace_id di log, jump ke Jaeger trace

#### 6. Stop Stack

```bash
# Stop all services
docker-compose down

# Stop and remove volumes (WARNING: deletes data)
docker-compose down -v
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
  your-registry/go-otel:latest
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
        - name: go-otel
          image: your-registry/go-otel:latest  # Ganti ini
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
docker build -t your-registry/go-otel:v1.0.0 .

# Push to registry
docker push your-registry/go-otel:v1.0.0

# Tag as latest
docker tag your-registry/go-otel:v1.0.0 your-registry/go-otel:latest
docker push your-registry/go-otel:latest
```

#### 6. Deploy Observability Stack to Kubernetes

```bash
# Deploy Loki stack (logs)
kubectl apply -f k8s/loki-deployment.yaml
kubectl apply -f k8s/promtail-daemonset.yaml
kubectl apply -f k8s/grafana-deployment.yaml

# Verify observability stack
kubectl get pods -l app=loki
kubectl get pods -l app=promtail
kubectl get pods -l app=grafana
```

#### 7. Deploy Application to Kubernetes

```bash
# Deploy semua application resources
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

#### 8. Verify Deployment

```bash
# Check pods
kubectl get pods -l app=go-otel

# Check deployment
kubectl get deployment go-otel

# Check service
kubectl get svc go-otel-service

# Check HPA
kubectl get hpa go-otel-hpa

# Check logs (via Loki, recommended)
# Access Grafana UI and use LogQL queries

# Or direct kubectl logs
kubectl logs -l app=go-otel --tail=100 -f
```

#### 9. Access Observability UIs

**Access Grafana (Logs + Traces):**
```bash
# Port forward Grafana
kubectl port-forward svc/grafana 3000:3000

# Open browser: http://localhost:3000
# Username: admin
# Password: admin (or check secret)
```

**Access Loki API (Direct):**
```bash
# Port forward Loki
kubectl port-forward svc/loki 3100:3100

# Query logs via API
curl http://localhost:3100/loki/api/v1/labels
```

**Query Logs di Kubernetes:**
```bash
# Via Grafana Explore
# LogQL: {app="go-otel", namespace="default"}

# See all pods logs
# LogQL: {app="go-otel"}

# Filter by log level
# LogQL: {app="go-otel"} |= "level=error"

# Aggregate errors per pod
# LogQL: sum by (pod) (count_over_time({app="go-otel"} |= "error" [5m]))
```

#### 10. Access Application

**Port Forward (untuk testing):**
```bash
kubectl port-forward svc/go-otel-service 8080:80
```

Then access: http://localhost:8080

**Via LoadBalancer (jika ada):**
```bash
kubectl get svc go-otel-service
# Note external IP
```

**Via Ingress:**
```bash
# Deploy ingress
kubectl apply -f k8s/ingress.yaml

# Get ingress address
kubectl get ingress go-otel-ingress

# Minikube specific
minikube service go-otel-service --url
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
docker images | grep go-otel

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
kubectl describe hpa go-otel-hpa

# Install metrics-server jika belum ada
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
```

---

## 📊 Observability Architecture

Project ini sudah include **full observability stack**:

### 1. **Traces** - Jaeger
- Distributed tracing via OpenTelemetry
- OTLP gRPC receiver (port 4317)
- UI: http://localhost:16686 (Docker) atau port-forward (K8s)
- **Sudah configured** - no extra setup needed

### 2. **Logs** - Loki + Promtail
- Centralized log aggregation
- 7 days retention (configurable)
- Auto-collects dari semua containers/pods
- **Sudah configured** - no extra setup needed

### 3. **Visualization** - Grafana
- Unified dashboard untuk logs dan traces
- Pre-configured datasources (Loki + Jaeger)
- Automatic correlation: logs ↔ traces
- **Sudah configured** - login dan langsung pakai

### 4. **Metrics** - Optional (Future)
```bash
# Install Prometheus via Helm (jika ingin add metrics)
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/kube-prometheus-stack
```

### Log Retention & Storage

**Docker Compose:**
- Logs disimpan di volume `loki_data`
- Retention: 7 days (default)
- Edit `config/loki-config.yaml` untuk ubah retention

**Kubernetes:**
- Logs disimpan di PersistentVolume (10Gi)
- Retention: 7 days (default)
- Edit `k8s/loki-deployment.yaml` ConfigMap untuk ubah retention

**Customize Retention:**
```yaml
# Edit retention period
limits_config:
  retention_period: 168h  # 7 days (ganti sesuai kebutuhan)
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
