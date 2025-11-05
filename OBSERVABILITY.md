# 📊 Observability Stack Documentation

## Overview

Project ini sudah include **full observability stack** untuk production-ready monitoring:

- **Traces** → Jaeger (OpenTelemetry)
- **Logs** → Loki + Promtail
- **Visualization** → Grafana (unified dashboard)
- **Correlation** → Automatic logs ↔ traces linking

---

## Architecture

```
┌─────────────────────────────────────────────────┐
│            Go Application                       │
│                                                 │
│  ┌──────────┐         ┌──────────┐            │
│  │  TRACES  │         │   LOGS   │            │
│  │  (OTLP)  │         │  (Zap)   │            │
│  └─────┬────┘         └─────┬────┘            │
└────────┼───────────────────┼──────────────────┘
         │                    │
         │ gRPC               │ stdout/stderr
         │ port 4317          │
         ↓                    ↓
    ┌─────────┐          ┌──────────┐
    │ Jaeger  │          │ Promtail │ ← Scrape logs
    └─────────┘          └─────┬────┘
         │                     │
         │                     │ HTTP push
         │                     ↓
         │                ┌─────────┐
         │                │  Loki   │ ← Store logs
         │                └─────────┘
         │                     │
         └──────────┬──────────┘
                    │
                    ↓
             ┌─────────────┐
             │   Grafana   │ ← Query & visualize
             └─────────────┘
```

---

## Quick Start

### Docker Compose (Development)

```bash
# Start full stack
docker-compose up -d

# Access services
open http://localhost:16686  # Jaeger
open http://localhost:3000   # Grafana (admin/admin)
open http://localhost:8080   # API

# Generate some traces and logs
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","email":"test@example.com"}'

# View in Jaeger: traces
# View in Grafana: logs with trace correlation
```

### Kubernetes (Production)

```bash
# Deploy observability stack
kubectl apply -f k8s/loki-deployment.yaml
kubectl apply -f k8s/promtail-daemonset.yaml
kubectl apply -f k8s/grafana-deployment.yaml

# Deploy application
kubectl apply -f k8s/

# Access Grafana
kubectl port-forward svc/grafana 3000:3000
open http://localhost:3000
```

---

## Components

### 1. Jaeger - Distributed Tracing

**Purpose:** Track request flow across services

**Features:**
- OTLP native support (no collector needed)
- Automatic trace collection
- Service dependency graph
- Latency analysis

**Access:**
- Docker: http://localhost:16686
- K8s: `kubectl port-forward svc/jaeger 16686:16686`

**Configuration:**
- File: `docker-compose.yml` (Docker)
- File: `k8s/deployment.yaml` env vars (K8s)

### 2. Loki - Log Aggregation

**Purpose:** Centralized log storage with retention

**Features:**
- Low resource usage (vs Elasticsearch)
- Efficient compression
- Label-based indexing
- 7-day retention (default)

**Access:**
- API: http://localhost:3100
- Via Grafana (recommended)

**Configuration:**
- Docker: `config/loki-config.yaml`
- K8s: `k8s/loki-deployment.yaml` ConfigMap

**Storage:**
- Docker: Volume `loki_data`
- K8s: PersistentVolumeClaim 10Gi

**Retention Policy:**
```yaml
limits_config:
  retention_period: 168h  # 7 days
```

### 3. Promtail - Log Collector

**Purpose:** Scrape logs from containers/pods and send to Loki

**Features:**
- Auto-discovery (Docker/Kubernetes)
- JSON log parsing
- Label extraction
- Pipeline processing

**Configuration:**
- Docker: `config/promtail-config.yaml`
- K8s: `k8s/promtail-daemonset.yaml` ConfigMap

**Deployment:**
- Docker: Single container
- K8s: DaemonSet (1 pod per node)

### 4. Grafana - Unified Visualization

**Purpose:** Query and visualize logs + traces

**Features:**
- Pre-configured datasources (Loki + Jaeger)
- Automatic logs ↔ traces correlation
- LogQL query language
- Alerting (can be configured)

**Access:**
- URL: http://localhost:3000
- Username: `admin`
- Password: `admin` (change in production!)

**Pre-configured Datasources:**
- **Loki** (default) - Logs
- **Jaeger** - Traces

---

## Usage Examples

### Query Logs in Grafana

Open Grafana → Explore → Select "Loki" datasource

**Basic queries:**
```logql
# All logs from application
{service="app"}

# Error logs only
{service="app"} |= "level=error"

# Logs from specific container
{container="go-otel"}

# Logs from specific pod (Kubernetes)
{pod="go-otel-7d8f9c6b5-abc12"}

# Logs by namespace
{namespace="production"}
```

**Advanced queries:**
```logql
# Rate of errors over time
rate({service="app"} |= "error" [5m])

# Count errors per pod
sum by (pod) (count_over_time({service="app"} |= "error" [5m]))

# Parse JSON and filter
{service="app"} | json | level="error"

# Find logs by trace ID
{service="app"} |= "trace_id=abc123def456"
```

### Correlation: Logs ↔ Traces

**Automatic correlation configured:**

**From Logs → Trace:**
1. View logs in Grafana Explore
2. Click on a log entry with `trace_id`
3. Click "Jaeger" button
4. Opens corresponding trace in Jaeger

**From Trace → Logs:**
1. View trace in Jaeger UI
2. Click "Logs" tab on a span
3. Shows related logs from Loki

### Query Traces in Jaeger

Open Jaeger UI → Select service "go-otel"

**Features:**
- View all traces
- Filter by operation (CreateUser, GetUser, etc)
- Search by trace ID
- View service dependencies
- Analyze latency distribution

---

## Configuration Files

### Docker Compose

```
config/
├── loki-config.yaml          # Loki configuration
├── promtail-config.yaml      # Promtail scrape config
└── grafana-datasources.yaml  # Grafana datasources
```

### Kubernetes

```
k8s/
├── loki-deployment.yaml      # Loki + ConfigMap + PVC + Service
├── promtail-daemonset.yaml   # Promtail DaemonSet + RBAC
└── grafana-deployment.yaml   # Grafana + ConfigMap + PVC + Service
```

---

## Customization

### Change Log Retention

**Docker Compose:**
Edit `config/loki-config.yaml`:
```yaml
limits_config:
  retention_period: 720h  # 30 days
```

**Kubernetes:**
Edit `k8s/loki-deployment.yaml` ConfigMap:
```yaml
data:
  loki-config.yaml: |
    limits_config:
      retention_period: 720h  # 30 days
```

Then restart Loki:
```bash
# Docker
docker-compose restart loki

# Kubernetes
kubectl rollout restart deployment loki
```

### Change Storage Size (Kubernetes)

Edit PVC in `k8s/loki-deployment.yaml`:
```yaml
spec:
  resources:
    requests:
      storage: 50Gi  # Increase from 10Gi
```

### Add Custom Labels to Logs

Edit `config/promtail-config.yaml` or `k8s/promtail-daemonset.yaml`:
```yaml
relabel_configs:
  - source_labels: [__meta_kubernetes_pod_label_version]
    target_label: version
```

### Configure Grafana Alerting

1. Open Grafana
2. Go to Alerting → Alert rules
3. Create new alert rule with LogQL query

**Example alert:**
```logql
# Alert if error rate > 10 per minute
rate({service="app"} |= "error" [1m]) > 10
```

---

## Troubleshooting

### Logs tidak muncul di Loki

**Check Promtail:**
```bash
# Docker
docker logs go-otel-promtail

# Kubernetes
kubectl logs -l app=promtail
```

**Check Loki:**
```bash
# Test Loki API
curl http://localhost:3100/ready

# Check labels (should return container, level, project, service, service_name)
curl http://localhost:3100/loki/api/v1/labels

# Check available services
curl http://localhost:3100/loki/api/v1/label/service/values

# Check Loki version and status
curl http://localhost:3100/loki/api/v1/status/buildinfo
```

**Common Issues:**

1. **"no such host" error in Promtail logs**
   - Loki container is not running
   - Solution: `docker-compose up -d loki`

2. **Loki config parse errors**
   - Check `config/loki-config.yaml` syntax
   - Common errors: deprecated cache config, missing delete_request_store
   - See fixed config in this repo

3. **Port 3100 not accessible**
   - From terminal: `curl http://localhost:3100/ready` should work
   - From browser: May need to wait for Loki to fully start (check logs)
   - Check: `docker-compose ps loki` - should show "Up"

### Traces tidak muncul di Jaeger

**Check application logs:**
```bash
# Docker
docker logs go-otel

# Kubernetes
kubectl logs -l app=go-otel
```

**Check Jaeger:**
```bash
# Test Jaeger
curl http://localhost:16686/api/services

# Check OTLP receiver
curl http://localhost:14269/  # Jaeger admin
```

### Grafana tidak terhubung ke datasources

**Check datasources:**
1. Grafana → Configuration → Data sources
2. Test connection untuk Loki dan Jaeger

**Re-provision datasources:**
```bash
# Docker
docker-compose restart grafana

# Kubernetes
kubectl rollout restart deployment grafana
```

### High storage usage

**Check Loki storage:**
```bash
# Docker
docker exec go-otel-loki du -sh /loki

# Kubernetes
kubectl exec -it loki-xxx -- du -sh /loki
```

**Solutions:**
1. Reduce retention period
2. Increase storage size
3. Add compression (already enabled)
4. Filter logs in Promtail

---

## Production Best Practices

### Security

- [ ] Change Grafana admin password
- [ ] Use HTTPS/TLS for Grafana
- [ ] Enable authentication for Loki
- [ ] Use secrets for sensitive configs
- [ ] Network policies for pod-to-pod communication

### Performance

- [ ] Monitor Loki resource usage
- [ ] Configure appropriate retention
- [ ] Use SSD for Loki storage
- [ ] Scale Promtail with cluster size
- [ ] Set resource limits on all pods

### Reliability

- [ ] Use persistent storage for Loki
- [ ] Backup Loki data regularly
- [ ] Monitor Promtail collection rate
- [ ] Set up Grafana high availability (multiple replicas)
- [ ] Configure alerts for observability stack health

### Storage Planning

**Estimate log volume:**
```
Daily logs per pod = 500 MB (average)
Number of pods = 10
Retention = 7 days

Total storage needed = 500 MB × 10 × 7 = 35 GB
+ 30% buffer = 45-50 GB
```

---

## Monitoring the Observability Stack

**Loki metrics:**
```bash
curl http://localhost:3100/metrics
```

**Promtail metrics:**
```bash
curl http://localhost:9080/metrics  # Promtail port
```

**Create Grafana dashboard to monitor:**
- Loki ingestion rate
- Promtail scrape errors
- Jaeger trace count
- Storage usage

---

## Migration from Other Stacks

### From ELK Stack
- Replace Elasticsearch with Loki (lighter)
- Replace Logstash/Filebeat with Promtail
- Keep Kibana or use Grafana

### From Splunk
- Export Splunk queries to LogQL
- Migrate dashboards to Grafana
- Train team on LogQL syntax

### Adding Prometheus (Future)
- Add Prometheus for metrics collection
- Connect to Grafana (already supports it)
- Complete observability: Traces + Logs + Metrics

---

## Resources

- [Loki Documentation](https://grafana.com/docs/loki/latest/)
- [Promtail Documentation](https://grafana.com/docs/loki/latest/clients/promtail/)
- [Grafana Documentation](https://grafana.com/docs/grafana/latest/)
- [Jaeger Documentation](https://www.jaegertracing.io/docs/)
- [LogQL Query Language](https://grafana.com/docs/loki/latest/logql/)
- [OpenTelemetry Go](https://opentelemetry.io/docs/instrumentation/go/)

---

**Need help?** Check [SETUP.md](SETUP.md) for detailed setup guide.
