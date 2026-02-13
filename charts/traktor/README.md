# Traktor Operator Helm Chart

Official Helm chart for deploying the Traktor operator on Kubernetes.

## Overview

Traktor is a Kubernetes operator that automatically restarts deployments when secrets change. This Helm chart provides a simple way to deploy and configure the operator in your cluster.

## Prerequisites

- Kubernetes 1.11.3+
- Helm 3.0+
- Cluster admin permissions (for RBAC setup)

## Installation

### Add Helm Repository

The Traktor Helm chart is available on [Artifact Hub](https://artifacthub.io/packages/helm/traktor/traktor).

```bash
helm repo add traktor https://gdxbsv.github.io/traktor
helm repo update
```

### Install from Artifact Hub

```bash
# Install latest version
helm install traktor traktor/traktor

# Install specific version
helm install traktor traktor/traktor --version 0.0.1
```

### Install from GitHub Release (Coming Soon)

```bash
# Install latest version
helm install traktor https://github.com/GDXbsv/traktor/releases/latest/download/traktor-0.0.1.tgz

# Install specific version
helm install traktor https://github.com/GDXbsv/traktor/releases/download/v0.0.1/traktor-0.0.1.tgz
```

### Install from Local Chart

```bash
# Clone repository
git clone https://github.com/GDXbsv/traktor.git
cd traktor

# Install chart
helm install traktor ./charts/traktor
```

### Install with Custom Values

```bash
helm install traktor ./charts/traktor -f my-values.yaml
```

## Quick Start

After installation:

1. **Check operator status:**
   ```bash
   kubectl get pods -n default -l app.kubernetes.io/name=traktor
   ```

2. **Create a SecretsRefresh resource:**
   ```yaml
   apiVersion: traktor.gdxcloud.net/v1alpha1
   kind: SecretsRefresh
   metadata:
     name: production-watcher
   spec:
     namespaceSelector:
       matchLabels:
         environment: production
     secretSelector:
       matchLabels:
         auto-refresh: enabled
   ```

3. **Label your resources:**
   ```bash
   kubectl label namespace my-app environment=production
   kubectl label secret my-secret -n my-app auto-refresh=enabled
   ```

## Configuration

### Basic Configuration

```yaml
# values.yaml
replicaCount: 1

image:
  repository: docker.io/gdxbsv/traktor
  tag: "0.0.1"
  pullPolicy: IfNotPresent

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 10m
    memory: 256Mi
```

### Values

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `1` |
| `image.repository` | Image repository | `docker.io/gdxbsv/traktor` |
| `image.tag` | Image tag (defaults to chart appVersion) | `""` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `imagePullSecrets` | Image pull secrets | `[]` |
| `nameOverride` | Override chart name | `""` |
| `fullnameOverride` | Override full name | `""` |
| `serviceAccount.create` | Create service account | `true` |
| `serviceAccount.annotations` | Service account annotations | `{}` |
| `serviceAccount.name` | Service account name | `""` |
| `podAnnotations` | Pod annotations | `{}` |
| `podSecurityContext` | Pod security context | See values.yaml |
| `securityContext` | Container security context | See values.yaml |
| `resources` | Resource limits and requests | See values.yaml |
| `nodeSelector` | Node selector | `{}` |
| `tolerations` | Tolerations | `[]` |
| `affinity` | Affinity rules | `{}` |
| `priorityClassName` | Priority class name | `""` |

### Advanced Configuration

#### Leader Election

```yaml
leaderElection:
  enabled: true
```

#### Metrics and Monitoring

```yaml
metrics:
  enabled: true
  port: 8443
  service:
    type: ClusterIP
    port: 8443
    annotations:
      prometheus.io/scrape: "true"
      prometheus.io/port: "8443"
```

#### Service Monitor (Prometheus Operator)

```yaml
serviceMonitor:
  enabled: true
  interval: 30s
  scrapeTimeout: 10s
  additionalLabels:
    prometheus: kube-prometheus
```

#### Pod Disruption Budget

```yaml
podDisruptionBudget:
  enabled: true
  minAvailable: 1
```

#### Horizontal Pod Autoscaler

```yaml
autoscaling:
  enabled: true
  minReplicas: 1
  maxReplicas: 3
  targetCPUUtilizationPercentage: 80
  targetMemoryUtilizationPercentage: 80
```

#### Network Policy

```yaml
networkPolicy:
  enabled: true
  policyTypes:
    - Ingress
    - Egress
  ingress: []
  egress:
    - to:
      - namespaceSelector: {}
```

## Examples

### Production Deployment

```yaml
# production-values.yaml
replicaCount: 2

resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 100m
    memory: 512Mi

podDisruptionBudget:
  enabled: true
  minAvailable: 1

affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        labelSelector:
          matchLabels:
            app.kubernetes.io/name: traktor
        topologyKey: kubernetes.io/hostname

serviceMonitor:
  enabled: true
  interval: 30s
  additionalLabels:
    prometheus: kube-prometheus

exampleResources:
  enabled: true
```

Install:
```bash
helm install traktor ./charts/traktor -f production-values.yaml
```

### Development Deployment

```yaml
# dev-values.yaml
replicaCount: 1

image:
  tag: "dev"
  pullPolicy: Always

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 10m
    memory: 128Mi

leaderElection:
  enabled: false

exampleResources:
  enabled: true
```

Install:
```bash
helm install traktor ./charts/traktor -n traktor-dev -f dev-values.yaml
```

### With Custom Environment Variables

```yaml
env:
  - name: LOG_LEVEL
    value: "debug"
  - name: SYNC_PERIOD
    value: "5m"
```

### With Node Selector

```yaml
nodeSelector:
  kubernetes.io/os: linux
  node-role.kubernetes.io/worker: ""
```

### With Tolerations

```yaml
tolerations:
  - key: "dedicated"
    operator: "Equal"
    value: "operators"
    effect: "NoSchedule"
```

## Upgrading

### Upgrade to New Version

```bash
# Using kubectl
helm upgrade traktor https://github.com/GDXbsv/traktor/releases/download/v0.0.2/traktor-0.0.2.tgz

# Using local chart
helm upgrade traktor ./charts/traktor
```

### Upgrade with New Values

```bash
helm upgrade traktor ./charts/traktor -f updated-values.yaml
```

### View Upgrade History

```bash
helm history traktor
```

### Rollback

```bash
# Rollback to previous version
helm rollback traktor

# Rollback to specific revision
helm rollback traktor 2
```

## Uninstallation

```bash
# Uninstall the chart
helm uninstall traktor

# Uninstall and delete CRDs
helm uninstall traktor
kubectl delete crd secretsrefreshes.traktor.gdxcloud.net
```

**Note:** By default, CRDs are kept even after uninstall to prevent data loss. Delete them manually if needed.

## Troubleshooting

### Check Helm Release Status

```bash
helm status traktor
helm get all traktor
```

### View Operator Logs

```bash
kubectl logs -l app.kubernetes.io/name=traktor -f
```

### Test Template Rendering

```bash
helm template traktor ./charts/traktor --debug
```

### Validate Values

```bash
helm lint ./charts/traktor -f my-values.yaml
```

### Common Issues

**Pods not starting:**
```bash
kubectl describe pod -l app.kubernetes.io/name=traktor
kubectl get events --sort-by='.lastTimestamp'
```

**RBAC errors:**
```bash
kubectl auth can-i --list --as=system:serviceaccount:default:traktor-controller-manager
```

**CRD not found:**
```bash
# Verify CRD installation
kubectl get crd secretsrefreshes.traktor.gdxcloud.net

# Manually install CRDs
kubectl apply -f charts/traktor/crds/
```

## Development

### Lint Chart

```bash
helm lint charts/traktor
```

### Test Installation

```bash
# Dry run
helm install traktor ./charts/traktor --dry-run --debug

# Test with different values
helm install traktor ./charts/traktor -f test-values.yaml --dry-run
```

### Package Chart

```bash
helm package charts/traktor
```

### Generate Documentation

```bash
helm-docs charts/traktor
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](../../CONTRIBUTING.md) for guidelines.

## Support

- **Documentation**: [README.md](../../README.md)
- **Issues**: [GitHub Issues](https://github.com/GDXbsv/traktor/issues)
- **Discussions**: [GitHub Discussions](https://github.com/GDXbsv/traktor/discussions)

## License

Apache License 2.0 - see [LICENSE](../../LICENSE) for details.

## Links

- [Artifact Hub](https://artifacthub.io/packages/helm/traktor/traktor)
- [GitHub Repository](https://github.com/GDXbsv/traktor)
- [Installation Guide](../../DEPLOYMENT.md)
- [Examples](../../config/samples/)
- [Docker Hub](https://hub.docker.com/r/gdxbsv/traktor)

---

**Chart Version**: 0.0.1  
**App Version**: 0.0.1  
**Maintained by**: [GDX Cloud](https://gdxcloud.net)