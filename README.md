# Traktor Operator

[![Tests](https://github.com/GDXbsv/traktor/actions/workflows/test.yml/badge.svg)](https://github.com/GDXbsv/traktor/actions/workflows/test.yml)
[![Lint](https://github.com/GDXbsv/traktor/actions/workflows/lint.yml/badge.svg)](https://github.com/GDXbsv/traktor/actions/workflows/lint.yml)
[![E2E Tests](https://github.com/GDXbsv/traktor/actions/workflows/test-e2e.yml/badge.svg)](https://github.com/GDXbsv/traktor/actions/workflows/test-e2e.yml)
[![Build](https://github.com/GDXbsv/traktor/actions/workflows/build.yml/badge.svg)](https://github.com/GDXbsv/traktor/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/GDXbsv/traktor)](https://goreportcard.com/report/github.com/GDXbsv/traktor)
[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/traktor)](https://artifacthub.io/packages/helm/traktor/traktor)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

A Kubernetes operator that automatically restarts deployments when secrets change. No more manual rollouts after secret updates!

## 🚀 Features

- **Automatic Deployment Restart** - Deployments automatically restart when their secrets change
- **Flexible Filtering** - Watch specific namespaces and secrets using label selectors
- **Zero Downtime** - Uses rolling restart strategy (Kubernetes default)
- **Self-Protection** - Operator never restarts itself
- **Multi-Architecture** - Supports AMD64 and ARM64
- **Production Ready** - Full RBAC, security scanning, comprehensive tests

## 📖 Table of Contents

- [Quick Start](#quick-start)
- [How It Works](#how-it-works)
- [Installation](#installation)
- [Configuration](#configuration)
- [Examples](#examples)
- [Development](#development)
- [Documentation](#documentation)
- [Contributing](#contributing)
- [License](#license)

## ⚡ Quick Start

### 1. Install the Operator

```bash
kubectl apply -f https://github.com/GDXbsv/traktor/releases/latest/download/install.yaml
```

### 2. Label Your Namespace

```bash
kubectl label namespace my-app environment=production
```

### 3. Label Your Secrets

```bash
kubectl label secret my-secret -n my-app auto-refresh=enabled
```

### 4. Create SecretsRefresh Resource

```yaml
apiVersion: traktor.gdxcloud.net/v1alpha1
kind: SecretsRefresh
metadata:
  name: production-watcher
  namespace: default
spec:
  namespaceSelector:
    matchLabels:
      environment: production
  secretSelector:
    matchLabels:
      auto-refresh: enabled
```

Apply it:
```bash
kubectl apply -f secretsrefresh.yaml
```

### 5. Update a Secret - Deployments Restart Automatically! 🎉

```bash
kubectl create secret generic my-secret \
  --from-literal=password=newpassword \
  -n my-app \
  --dry-run=client -o yaml | kubectl apply -f -

# Watch deployments restart
kubectl get pods -n my-app -w
```

## 🔍 How It Works

1. **Watch Secrets** - Operator watches for changes to secrets matching your selectors
2. **Detect Changes** - When a secret is updated, the operator is notified
3. **Restart Deployments** - All deployments in the same namespace are restarted by adding an annotation:
   ```yaml
   traktor.gdxcloud.net/restartedAt: "2024-01-30T10:30:00Z"
   ```
4. **Rolling Update** - Kubernetes performs a rolling restart (zero downtime)
5. **Pods Get New Secrets** - New pods automatically mount the updated secrets

**Flow Diagram:**
```
Secret Update → Operator Detects → Adds Annotation → Rolling Restart → New Pods with Updated Secrets
```

## 📦 Installation

### Prerequisites

- Kubernetes cluster v1.11.3+
- kubectl v1.11.3+
- Cluster admin permissions (for RBAC setup)

### Option 1: Install from Release (Recommended)

**Using kubectl:**
```bash
# Install latest version
kubectl apply -f https://github.com/GDXbsv/traktor/releases/latest/download/install.yaml

# Install specific version
kubectl apply -f https://github.com/GDXbsv/traktor/releases/download/v0.0.1/install.yaml
```

**Using Helm (from Artifact Hub/GitHub Pages):**
```bash
# Add Helm repository
helm repo add traktor https://gdxbsv.github.io/traktor
helm repo update

# Install latest version
helm install traktor traktor/traktor

# Install specific version
helm install traktor traktor/traktor --version 0.0.1

# Install with custom values
helm install traktor traktor/traktor -f values.yaml
```

**Using Helm (from GitHub Release):**
```bash
# Install directly from release
helm install traktor https://github.com/GDXbsv/traktor/releases/latest/download/traktor-0.0.1.tgz
```

### Option 2: Install from Source

```bash
# Clone repository
git clone https://github.com/GDXbsv/traktor.git
cd traktor

# Install CRDs
make install

# Deploy operator
make deploy IMG=docker.io/gdxbsv/traktor:v0.0.12
```

### Option 3: Install Using Helm from Source

```bash
# Clone repository
git clone https://github.com/GDXbsv/traktor.git
cd traktor

# Install with Helm
helm install traktor ./charts/traktor

# Install in custom namespace
helm install traktor ./charts/traktor -n traktor-system --create-namespace

# Install with custom values
helm install traktor ./charts/traktor -f my-values.yaml
```

### Option 4: Using Kustomize

```bash
kubectl apply -k config/default
```

### Verify Installation

```bash
# Check operator is running
kubectl get pods -n traktor-system

# Expected output:
# NAME                                          READY   STATUS    RESTARTS   AGE
# traktor-controller-manager-xxxxxxxxxx-xxxxx   1/1     Running   0          30s

# Check CRD is installed
kubectl get crd secretsrefreshes.traktor.gdxcloud.net
```

## ⚙️ Configuration

### SecretsRefresh Custom Resource

```yaml
apiVersion: traktor.gdxcloud.net/v1alpha1
kind: SecretsRefresh
metadata:
  name: my-secrets-watcher
  namespace: default
spec:
  # Filter namespaces by labels
  namespaceSelector:
    matchLabels:
      environment: production
      team: platform
    matchExpressions:
      - key: app
        operator: In
        values: [backend, frontend]
  
  # Filter secrets by labels
  secretSelector:
    matchLabels:
      auto-refresh: enabled
    matchExpressions:
      - key: type
        operator: NotIn
        values: [system]
```

### Namespace Selector

**Match by exact labels:**
```yaml
namespaceSelector:
  matchLabels:
    environment: production
```

**Match by expressions:**
```yaml
namespaceSelector:
  matchExpressions:
    - key: team
      operator: In
      values: [backend, frontend, platform]
```

**Watch all namespaces:**
```yaml
# Omit namespaceSelector entirely
spec:
  secretSelector:
    matchLabels:
      auto-refresh: enabled
```

### Secret Selector

**Match by labels:**
```yaml
secretSelector:
  matchLabels:
    auto-refresh: enabled
    type: app-config
```

**Watch all secrets in matched namespaces:**
```yaml
spec:
  namespaceSelector:
    matchLabels:
      environment: production
  # Omit secretSelector to watch all secrets
```

## 📝 Examples

### Example 1: Production Applications

Watch all secrets in production namespaces:

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

**Setup:**
```bash
# Label namespaces
kubectl label namespace app-backend environment=production
kubectl label namespace app-frontend environment=production

# Label secrets
kubectl label secret db-password -n app-backend auto-refresh=enabled
kubectl label secret api-keys -n app-frontend auto-refresh=enabled
```

### Example 2: Database Credentials

Watch only database-related secrets:

```yaml
apiVersion: traktor.gdxcloud.net/v1alpha1
kind: SecretsRefresh
metadata:
  name: database-credentials-watcher
spec:
  namespaceSelector:
    matchLabels:
      environment: production
  secretSelector:
    matchLabels:
      type: database-credentials
```

### Example 3: Multi-Team Setup

Watch secrets across different teams:

```yaml
apiVersion: traktor.gdxcloud.net/v1alpha1
kind: SecretsRefresh
metadata:
  name: multi-team-watcher
spec:
  namespaceSelector:
    matchExpressions:
      - key: team
        operator: In
        values: [backend, frontend, platform]
  secretSelector:
    matchLabels:
      auto-refresh: enabled
```

### Example 4: Single Namespace

Watch all secrets in a specific namespace:

```yaml
apiVersion: traktor.gdxcloud.net/v1alpha1
kind: SecretsRefresh
metadata:
  name: staging-watcher
spec:
  namespaceSelector:
    matchExpressions:
      - key: kubernetes.io/metadata.name
        operator: In
        values: [staging]
```

### More Examples

See [config/samples/](config/samples/) for more complete examples:
- `quickstart.yaml` - Simple getting started example
- `example-complete.yaml` - All configuration options
- `production-example.yaml` - Real-world production setup

## 🛠️ Development

### Prerequisites

- Go v1.24.0+
- Docker v17.03+
- kubectl v1.11.3+
- Access to Kubernetes cluster

### Local Development Setup

```bash
# Clone repository
git clone https://github.com/GDXbsv/traktor.git
cd traktor

# Install dependencies
go mod download

# Install CRDs
make install

# Run locally (connects to your kubeconfig cluster)
make run
```

### Running Tests

```bash
# Run unit tests
make test

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run linter
make lint

# Run E2E tests
make test-e2e
```

### Building

```bash
# Build binary
make build

# Build Docker image
make docker-build

# Build and push (multi-arch)
export IMG=docker.io/yourusername/traktor:dev
make docker-build docker-push IMG=$IMG

# Deploy to cluster
make deploy IMG=$IMG
```

### Project Structure

```
traktor/
├── api/v1alpha1/           # API definitions (CRDs)
├── cmd/                    # Main application entry point
├── config/                 # Kubernetes manifests
│   ├── crd/               # CRD definitions
│   ├── manager/           # Operator deployment
│   ├── rbac/              # RBAC manifests
│   └── samples/           # Example configurations
├── internal/controller/    # Controller logic
├── test/e2e/              # End-to-end tests
└── docs/                  # Documentation
```

## 📚 Documentation

- **[Installation Guide](DEPLOYMENT.md)** - Detailed deployment instructions
- **[Helm Chart](charts/traktor/README.md)** - Helm installation and configuration
- **[Artifact Hub](https://artifacthub.io/packages/helm/traktor/traktor)** - Discover on Artifact Hub
- **[Artifact Hub Setup](docs/ARTIFACTHUB_SETUP.md)** - How to publish to Artifact Hub
- **[Testing Guide](TESTING.md)** - How to run and write tests
- **[CI/CD Setup](docs/CICD_SETUP.md)** - Setting up GitHub Actions
- **[Examples](config/samples/)** - Configuration examples
- **[Architecture](.github/workflows/README.md)** - How the operator works

## 🔒 Security

### Vulnerability Scanning

All releases are scanned with Trivy for vulnerabilities. See [Security tab](../../security) for reports.

### RBAC Permissions

The operator requires the following permissions:
- Read secrets in all namespaces
- Read namespaces
- Update deployments

See [config/rbac/](config/rbac/) for complete RBAC configuration.

### Reporting Security Issues

Please report security vulnerabilities to security@gdxcloud.net

## 🚢 Releases

### Creating a Release

```bash
# Tag the release
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

GitHub Actions will automatically:
- Run all tests
- Build multi-arch images
- Generate Kubernetes manifests
- Create GitHub release with artifacts
- Push Docker images with proper tags

### Release Artifacts

Each release includes:
- `install.yaml` - Complete installation manifest
- `traktor-vX.Y.Z.tgz` - Helm chart package
- `traktor-vX.Y.Z-manifests.tar.gz` - All manifests packaged
- `sbom-vX.Y.Z.spdx.json` - Software Bill of Materials
- `index.yaml` - Helm repository index
- Docker images for AMD64 and ARM64

## 🤝 Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for your changes
5. Run tests locally (`make test`)
6. Commit your changes (`git commit -m 'feat: add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Code Style

- Follow Go best practices
- Run `make lint` before committing
- Write tests for new features
- Update documentation as needed

## 🐛 Troubleshooting

### Deployments Not Restarting

**Check operator logs:**
```bash
kubectl logs -n traktor-system deployment/traktor-controller-manager -f
```

**Verify labels:**
```bash
# Check namespace labels
kubectl get namespace my-app --show-labels

# Check secret labels
kubectl get secrets -n my-app --show-labels
```

**Verify SecretsRefresh exists:**
```bash
kubectl get secretsrefresh
```

### Operator Crashes

**Check for OOM:**
```bash
kubectl describe pod -n traktor-system -l control-plane=controller-manager
```

**Increase memory limit in `config/manager/manager.yaml`:**
```yaml
resources:
  limits:
    memory: 1Gi  # Increase from 512Mi
```

### More Help

- Check [Documentation](#documentation)
- Open an [Issue](../../issues)
- Review [Closed Issues](../../issues?q=is%3Aissue+is%3Aclosed)

## 📊 Metrics

The operator exposes Prometheus metrics on port 8443:

- `controller_runtime_reconcile_total` - Total reconciliations
- `controller_runtime_reconcile_errors_total` - Reconciliation errors
- `workqueue_*` - Work queue metrics

Access metrics:
```bash
kubectl port-forward -n traktor-system svc/traktor-controller-manager-metrics-service 8443:8443
curl -k https://localhost:8443/metrics
```

## 🔗 Related Projects

- [Reloader](https://github.com/stakater/Reloader) - Similar project with different approach
- [Wave](https://github.com/wave-k8s/wave) - ConfigMap and Secret change detection

## 📄 License

Copyright 2026 GDX Cloud.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

## 🌟 Star History

If you find this project useful, please consider giving it a star! ⭐

## 📞 Contact

- **Issues**: [GitHub Issues](../../issues)
- **Discussions**: [GitHub Discussions](../../discussions)
- **Artifact Hub**: [traktor on Artifact Hub](https://artifacthub.io/packages/helm/traktor/traktor)
- **Helm Repository**: https://gdxbsv.github.io/traktor
- **Email**: support@gdxcloud.net
- **Website**: https://gdxcloud.net

---

**Built with ❤️ using [Kubebuilder](https://book.kubebuilder.io/)**