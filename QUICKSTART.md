# 🚀 Quick Start Guide

## Yang Anda Butuhkan

### ✅ Minimum Requirements

#### Development Lokal
1. **Go 1.21+** → [Download](https://go.dev/dl/)
2. **Docker** → [Download](https://www.docker.com/products/docker-desktop)
3. **Git** → Biasanya sudah terinstall

#### Production (Kubernetes)
1. **Kubernetes Cluster** (pilih salah satu):
   - Minikube (testing lokal)
   - Docker Desktop + Kubernetes
   - Cloud: GKE, EKS, AKS
2. **kubectl** → [Install](https://kubernetes.io/docs/tasks/tools/)
3. **Container Registry**:
   - Docker Hub (gratis)
   - GitHub Container Registry (gratis)
   - Cloud provider registry

---

## 🏃 Running Paling Cepat

### Option 1: Docker Compose (Paling Mudah - Recommended)

```bash
# 1. Clone dan masuk ke folder
cd go-otel

# 2. Start FULL observability stack:
#    - PostgreSQL (database)
#    - App (Go application)
#    - Jaeger (distributed tracing)
#    - Loki (log storage)
#    - Promtail (log collector)
#    - Grafana (unified dashboard)
docker-compose up -d

# 3. Lihat logs
docker-compose logs -f app

# 4. Test API
curl http://localhost:8080/health/ready

# 5. Buat user (akan generate traces + logs)
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Test User","email":"test@example.com"}'

# 6. Access Observability UIs
open http://localhost:16686   # Jaeger (traces)
open http://localhost:3000    # Grafana (logs + traces)
# Grafana login: admin/admin

# Stop semua
docker-compose down
```

**✅ Selesai! Full observability stack running dengan:**
- ✅ Distributed tracing (Jaeger)
- ✅ Centralized logging (Loki + Promtail)
- ✅ Unified visualization (Grafana)
- ✅ Automatic logs ↔ traces correlation

---

### Option 2: Local Development

```bash
# 1. Install dependencies Go
go mod download

# 2. Setup environment
cp .env.example .env

# 3. Start PostgreSQL via Docker
docker run -d --name postgres \
  -e POSTGRES_DB=go_otel_db \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 \
  postgres:15-alpine

# 4. Run aplikasi
go run cmd/api/main.go

# Atau dengan hot reload (install air dulu)
go install github.com/air-verse/air@latest
air
```

---

### Option 3: Kubernetes (Production)

```bash
# 1. Pastikan kubectl terkoneksi
kubectl cluster-info

# 2. Edit secret (PENTING!)
nano k8s/secret.yaml
# Ganti password database

# 3. Update image di deployment
nano k8s/deployment.yaml
# Ganti: image: your-registry/go-otel:latest

# 4. Build dan push image
docker build -t your-registry/go-otel:latest .
docker push your-registry/go-otel:latest

# 5. Deploy ke Kubernetes
make k8s-deploy

# 6. Check status
kubectl get pods -l app=go-otel

# 7. Test via port-forward
kubectl port-forward svc/go-otel-service 8080:80

# 8. Test API
curl http://localhost:8080/health/ready
```

---

## 📊 Access Services

### Setelah docker-compose up:

| Service | URL | Credentials | Description |
|---------|-----|-------------|-------------|
| **API** | http://localhost:8080 | - | Main application |
| **Health Check** | http://localhost:8080/health/ready | - | Application health |
| **Jaeger UI** | http://localhost:16686 | - | Distributed tracing visualization |
| **Grafana** | http://localhost:3000 | admin/admin | Logs + Traces unified dashboard |
| **Loki API** | http://localhost:3100 | - | Log aggregation API |
| **PostgreSQL** | localhost:5432 | postgres/postgres | Database |

### Grafana Quick Access

**Pre-configured datasources:**
- **Loki** (default) - Query logs dengan LogQL
- **Jaeger** - View distributed traces

**Example queries di Grafana Explore:**
```logql
# All logs
{service="app"}

# Error logs only
{service="app"} |= "level=error"

# Logs by container
{container="go-otel"}

# Find logs by trace ID (correlation)
{service="app"} |= "trace_id=YOUR_TRACE_ID"
```

---

## 🧪 Test API Endpoints

```bash
# Health check
curl http://localhost:8080/health/live
curl http://localhost:8080/health/ready

# Create user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com"
  }'

# List users
curl http://localhost:8080/api/v1/users

# Get user by ID
curl http://localhost:8080/api/v1/users/{id}

# Update user
curl -X PUT http://localhost:8080/api/v1/users/{id} \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jane Doe",
    "email": "jane@example.com"
  }'

# Delete user
curl -X DELETE http://localhost:8080/api/v1/users/{id}
```

---

## 🔧 Configuration Files

### Development (.env)
```env
APP_PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=go_otel_db
LOG_LEVEL=debug
LOG_FORMAT=console
OTEL_ENABLED=false
```

### Kubernetes (k8s/configmap.yaml)
- Connection pool settings
- OpenTelemetry config
- Log level

### Kubernetes (k8s/secret.yaml)
- Database credentials (⚠️ **GANTI PASSWORD!**)

---

## ⚙️ Connection Pool Settings

Default settings (per pod):
```yaml
DB_MAX_OPEN_CONNS: 25      # Max connections per pod
DB_MAX_IDLE_CONNS: 10      # Idle connections per pod
DB_CONN_MAX_LIFETIME: 5m   # Connection lifetime
DB_CONN_MAX_IDLE_TIME: 10m # Idle timeout
```

**Perhitungan:**
- 3 pods × 25 connections = 75 total
- 10 pods × 25 connections = 250 total (with HPA max)

**PostgreSQL harus support lebih banyak:**
```sql
ALTER SYSTEM SET max_connections = 300;
```

---

## 🐛 Troubleshooting

### Database connection error
```bash
# Check PostgreSQL running
docker ps | grep postgres

# Check logs
docker logs postgres

# Test connection
psql -h localhost -U postgres -d go_otel_db
```

### Pods not starting (Kubernetes)
```bash
# Check pod status
kubectl get pods -l app=go-otel

# Describe pod
kubectl describe pod <pod-name>

# Check logs
kubectl logs <pod-name>

# Check events
kubectl get events --sort-by=.metadata.creationTimestamp
```

### Image pull error
```bash
# Login to registry
docker login

# Re-push image
docker push your-registry/go-otel:latest
```

---

## 📚 Next Steps

1. ✅ Run aplikasi (pilih salah satu option di atas)
2. ✅ Test API endpoints
3. ✅ Lihat traces di Jaeger UI (http://localhost:16686)
4. ✅ Query logs di Grafana (http://localhost:3000)
5. ✅ Test correlation: Click trace_id di logs → Jump ke Jaeger
6. ⏭️ Configure CI/CD
7. ⏭️ Production deployment ke Kubernetes
8. ⏭️ (Optional) Add Prometheus untuk metrics

---

## 📖 Documentation

- **[SETUP.md](SETUP.md)** - Panduan setup lengkap dan detail
- **[README.md](README.md)** - Dokumentasi project lengkap
- **[Makefile](Makefile)** - Available commands

---

## 🆘 Common Commands

```bash
# Development
make deps          # Download dependencies
make run           # Run locally
make test          # Run tests
make dev-db-up     # Start PostgreSQL
make dev-db-down   # Stop PostgreSQL

# Docker
make docker-build  # Build image
make docker-run    # Run container

# Kubernetes
make k8s-deploy    # Deploy all
make k8s-status    # Check status
make k8s-logs      # View logs
make k8s-delete    # Delete all resources

# Docker Compose
docker-compose up -d        # Start all services
docker-compose down         # Stop all services
docker-compose logs -f app  # View logs
docker-compose ps           # List services
```

---

## 🎯 Production Checklist

Before deploying to production:

### Security
- [ ] Ganti database password di `k8s/secret.yaml`
- [ ] Update Grafana admin password di `k8s/grafana-deployment.yaml`
- [ ] Configure ingress dengan TLS/SSL
- [ ] Security scan images

### Infrastructure
- [ ] Update image registry di `k8s/deployment.yaml`
- [ ] Review resource limits
- [ ] Setup persistent volume untuk PostgreSQL (done via StatefulSet)
- [ ] Setup persistent volume untuk Loki (10Gi)
- [ ] Setup persistent volume untuk Grafana (5Gi)

### Observability (Already Configured!)
- [x] ✅ Distributed tracing (Jaeger)
- [x] ✅ Log aggregation (Loki + Promtail)
- [x] ✅ Unified visualization (Grafana)
- [ ] (Optional) Setup Prometheus untuk metrics
- [ ] Configure log retention policy (default: 7 days)

### Operations
- [ ] Configure backup strategy untuk PostgreSQL
- [ ] Configure backup strategy untuk Loki data
- [ ] Setup CI/CD pipeline
- [ ] Load testing
- [ ] Setup alerting rules di Grafana

---

**🚀 Selamat! Service Anda siap digunakan!**

Need help? Check [SETUP.md](SETUP.md) for detailed guide or open an issue.
