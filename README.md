# Go OTEL API

API Golang yang maintainable dan readable dengan penerapan Echo, GORM, PostgreSQL, Structured Logging, dan OpenTelemetry untuk observability.

## Features

- **Clean Architecture**: Pemisahan domain, repository, service, dan handler
- **Echo Framework**: HTTP router dan middleware yang cepat
- **GORM**: ORM yang powerful untuk PostgreSQL
- **Structured Logging**: Menggunakan Zap untuk logging terstruktur
- **OpenTelemetry**: Distributed tracing untuk observability
- **Database Connection Pooling**: Konfigurasi optimal untuk production
- **Health Checks**: Liveness, readiness, dan startup probes untuk Kubernetes
- **Graceful Shutdown**: Proper cleanup saat aplikasi dihentikan
- **Kubernetes Ready**: Manifests lengkap dengan HPA dan PDB

## Project Structure

```
.
├── cmd/
│   └── api/
│       └── main.go              # Entry point aplikasi
├── config/
│   └── config.go                # Konfigurasi aplikasi
├── internal/
│   ├── domain/
│   │   └── user.go              # Domain models
│   ├── repository/
│   │   └── user_repository.go   # Data access layer
│   ├── service/
│   │   └── user_service.go      # Business logic layer
│   ├── handler/
│   │   ├── user_handler.go      # HTTP handlers
│   │   └── health_handler.go    # Health check handlers
│   └── middleware/
│       └── middleware.go        # Custom middlewares
├── pkg/
│   ├── database/
│   │   └── postgres.go          # Database connection
│   ├── logger/
│   │   └── logger.go            # Logger initialization
│   └── telemetry/
│       └── telemetry.go         # OpenTelemetry setup
├── k8s/
│   ├── configmap.yaml           # ConfigMap untuk environment variables
│   ├── secret.yaml              # Secret untuk credentials
│   ├── deployment.yaml          # Deployment dengan health checks
│   ├── service.yaml             # Service ClusterIP
│   ├── hpa.yaml                 # HorizontalPodAutoscaler
│   ├── pdb.yaml                 # PodDisruptionBudget
│   └── postgres-statefulset.yaml # PostgreSQL StatefulSet
├── Dockerfile                    # Multi-stage Docker build
├── Makefile                      # Build automation
├── go.mod                        # Go dependencies
└── .env.example                  # Example environment variables
```

## Database Connection Pooling

Aplikasi ini dikonfigurasi dengan connection pooling yang optimal untuk production:

### Konfigurasi Default

```
DB_MAX_OPEN_CONNS=25      # Maksimal koneksi terbuka per pod
DB_MAX_IDLE_CONNS=10      # Koneksi idle yang dijaga per pod
DB_CONN_MAX_LIFETIME=5m   # Lifetime maksimal koneksi
DB_CONN_MAX_IDLE_TIME=10m # Waktu maksimal koneksi idle
```

### Perhitungan Kapasitas

Jika Anda memiliki:
- 3 replicas (pods)
- Max 25 connections per pod
- Total: 3 × 25 = **75 connections maksimal**

PostgreSQL harus dikonfigurasi dengan `max_connections` yang lebih besar:
```
max_connections = 100  # Beri buffer 25-30%
```

### Scaling Considerations

Saat menggunakan HPA (Horizontal Pod Autoscaler):
- Min replicas: 3
- Max replicas: 10
- Max total connections: 10 × 25 = 250 connections

Pastikan database Anda dapat menangani load ini!

## Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Docker & Kubernetes (untuk deployment)
- OpenTelemetry Collector (optional, untuk tracing)

## Getting Started

### 1. Setup Development Environment

```bash
# Clone repository
git clone <repository-url>
cd go-otel

# Copy environment file
cp .env.example .env

# Edit .env dengan konfigurasi Anda
nano .env
```

### 2. Start Dependencies

```bash
# Start PostgreSQL
make dev-db-up

# Start OpenTelemetry Collector (optional)
make dev-otel-up
```

### 3. Run Application

```bash
# Download dependencies
make deps

# Run application
make run
```

Server akan berjalan di `http://localhost:8080`

## API Endpoints

### Health Checks

```bash
# Liveness probe
GET /health/live

# Readiness probe
GET /health/ready

# Startup probe
GET /health/startup
```

### User API

```bash
# Create user
POST /api/v1/users
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com"
}

# List users (with pagination)
GET /api/v1/users?limit=10&offset=0

# Get user by ID
GET /api/v1/users/{id}

# Update user
PUT /api/v1/users/{id}
Content-Type: application/json

{
  "name": "Jane Doe",
  "email": "jane@example.com"
}

# Delete user
DELETE /api/v1/users/{id}
```

## Docker

### Build Image

```bash
make docker-build
```

### Run Container Locally

```bash
make docker-run
```

### Push to Registry

```bash
# Update DOCKER_IMAGE in Makefile
make docker-push
```

## Kubernetes Deployment

### 1. Update Configurations

Edit file berikut sesuai kebutuhan:
- `k8s/secret.yaml` - Database credentials
- `k8s/configmap.yaml` - Application configuration
- `k8s/deployment.yaml` - Image registry dan resource limits

### 2. Deploy to Kubernetes

```bash
# Deploy semua resources
make k8s-deploy

# Atau deploy secara bertahap:
make k8s-apply-config    # ConfigMap dan Secret
make k8s-apply-postgres  # PostgreSQL
make k8s-apply-app       # Application
```

### 3. Check Status

```bash
make k8s-status
```

### 4. View Logs

```bash
make k8s-logs
```

### 5. Cleanup

```bash
make k8s-delete
```

## Kubernetes Features

### 1. Health Checks

Deployment menggunakan 3 jenis probes:
- **Liveness**: Mengecek apakah container masih hidup
- **Readiness**: Mengecek apakah pod siap menerima traffic
- **Startup**: Untuk aplikasi yang butuh waktu lama untuk start

### 2. Horizontal Pod Autoscaler (HPA)

```yaml
minReplicas: 3
maxReplicas: 10
```

Autoscaling berdasarkan:
- CPU utilization (target: 70%)
- Memory utilization (target: 80%)

### 3. Pod Disruption Budget (PDB)

```yaml
minAvailable: 2
```

Memastikan minimal 2 pod tersedia selama voluntary disruption.

### 4. Resource Limits

```yaml
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi
```

### 5. Security

- Non-root user
- Read-only root filesystem
- Drop all capabilities
- Pod anti-affinity untuk distribusi pods

## OpenTelemetry

Aplikasi ini sudah terintegrasi dengan OpenTelemetry untuk:
- **Distributed Tracing**: Melacak request flow
- **Span Attributes**: Metadata pada setiap operation
- **Context Propagation**: Across services

### Setup OpenTelemetry Collector

```bash
# Jalankan collector lokal
make dev-otel-up

# Atau deploy ke Kubernetes
kubectl apply -f https://raw.githubusercontent.com/open-telemetry/opentelemetry-collector/main/examples/k8s/otel-config.yaml
```

## Monitoring

### Metrics (Prometheus)

Application sudah dianotasi untuk Prometheus scraping:

```yaml
annotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "8080"
  prometheus.io/path: "/metrics"
```

### Logs

Structured logging dalam format JSON untuk easy parsing:

```json
{
  "level": "info",
  "ts": "2025-10-11T10:00:00.000Z",
  "msg": "Request processed",
  "method": "GET",
  "uri": "/api/v1/users",
  "status": 200,
  "latency": 0.025
}
```

## Testing

```bash
# Run tests
make test

# Run tests with coverage
make test-coverage

# Run linter
make lint
```

## Environment Variables

Lihat [.env.example](.env.example) untuk daftar lengkap environment variables.

Key configurations:
- `DB_MAX_OPEN_CONNS`: Maksimal open connections
- `DB_MAX_IDLE_CONNS`: Maksimal idle connections
- `DB_CONN_MAX_LIFETIME`: Connection lifetime
- `DB_CONN_MAX_IDLE_TIME`: Connection idle time

## Production Considerations

### 1. Database

- Gunakan managed database service (AWS RDS, GCP Cloud SQL, dll)
- Atau gunakan PostgreSQL operator (CloudNativePG, Patroni)
- Setup replikasi untuk high availability
- Regular backups

### 2. Observability

- Setup OpenTelemetry Collector
- Integrate dengan Jaeger/Zipkin untuk tracing
- Setup Prometheus + Grafana untuk metrics
- Centralized logging dengan ELK/Loki

### 3. Security

- Gunakan secrets management (Vault, AWS Secrets Manager)
- Enable network policies
- Regular security scanning
- TLS untuk database connections

### 4. Performance

- Tune connection pool sesuai load
- Setup CDN untuk static assets
- Enable HTTP/2
- Consider caching layer (Redis)

## Contributing

1. Fork repository
2. Create feature branch
3. Commit changes
4. Push to branch
5. Create Pull Request

## License

MIT License

## Support

Untuk pertanyaan atau issue, silakan buat issue di repository ini.