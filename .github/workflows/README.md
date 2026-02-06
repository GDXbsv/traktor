# Traktor Operator - CI/CD Documentation

This document describes the GitHub Actions workflows used for continuous integration, testing, and deployment of the Traktor operator.

## Overview

The CI/CD pipeline consists of several workflows:

1. **Test** - Run unit tests on every push and PR
2. **Lint** - Code quality checks
3. **E2E Tests** - End-to-end testing with Kind cluster
4. **Build** - Build and push Docker images
5. **Release** - Automated releases with versioning

## Workflows

### 1. Test Workflow (`test.yml`)

**Triggers:**
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop`

**Steps:**
1. Checkout code
2. Setup Go environment with caching
3. Run `go mod tidy` and verify no changes
4. Run unit tests with `make test`
5. Upload coverage to Codecov
6. Generate coverage report
7. Comment coverage on PR (if applicable)

**Coverage Reporting:**
- Minimum coverage target: 43.5%
- Reports uploaded to Codecov
- PR comments show coverage percentage

**Example output:**
```bash
📊 Total coverage: 43.5%
✅ All tests passing
```

---

### 2. Lint Workflow (`lint.yml`)

**Triggers:**
- Push to any branch
- Pull requests

**Steps:**
1. Checkout code
2. Setup Go environment
3. Run `golangci-lint` with version v2.1.0

**Linter Configuration:**
- Timeout: 5 minutes
- Configuration: `.golangci.yml` (if present)
- Checks: code style, best practices, potential bugs

---

### 3. E2E Tests Workflow (`test-e2e.yml`)

**Triggers:**
- Push to any branch
- Pull requests

**Steps:**
1. Checkout code
2. Setup Go environment
3. Install Kind (Kubernetes in Docker)
4. Run `make test-e2e`

**What it tests:**
- Operator deployment in real Kubernetes cluster
- CRD installation
- RBAC permissions
- Metrics endpoint functionality
- Controller manager health

**Requirements:**
- Kind cluster creation (~1-2 minutes)
- Full operator deployment
- Integration testing

---

### 4. Build and Push Workflow (`build.yml`)

**Triggers:**
- Push to `main` or `develop` branches
- Push tags matching `v*.*.*`
- Pull requests to `main` or `develop`

**Jobs:**

#### 4.1 Build and Test
- Run unit tests
- Upload coverage
- Run linter
- Build binary with `make build`

#### 4.2 Build Docker
- Multi-architecture build (amd64, arm64)
- Push to Docker Hub (docker.io/gdxbsv/traktor)
- Generate SBOM (Software Bill of Materials)
- Cache layers for faster builds

**Image Tags:**
- `latest` - Latest build from main branch
- `main` - Latest main branch build
- `develop` - Latest develop branch build
- `pr-123` - Pull request builds
- `v1.2.3` - Semantic version tags
- `main-abc1234` - Branch name + commit SHA

#### 4.3 Generate Manifests
- Generate Kubernetes install manifest
- Upload as artifact
- Create GitHub release (on tags)

#### 4.4 Security Scan
- Trivy vulnerability scanning
- Upload results to GitHub Security
- Scan for CRITICAL and HIGH vulnerabilities
- SARIF format for GitHub integration

#### 4.5 Notify
- Report build status
- Create success/failure badges

---

### 5. Release Workflow (`release.yml`)

**Triggers:**
- Push tags matching `v*.*.*` or `*.*.*` (e.g., v1.0.0, 1.0.0, v1.2.3-alpha, 1.2.3-alpha)
- Supports tags both with and without the 'v' prefix

**Jobs:**

#### 5.1 Validate Tag
- Verify tag format (v1.2.3, 1.2.3, v1.2.3-alpha, or 1.2.3-alpha)
- Extract version number (strips 'v' prefix if present)
- Detect prerelease (alpha, beta, rc)

#### 5.2 Run Tests
- Full unit test suite
- Linting checks
- Must pass before proceeding

#### 5.3 Build Multi-Architecture
- Build for linux/amd64 and linux/arm64
- Push with multiple tags:
  - `v1.2.3` - Exact version
  - `v1.2` - Minor version
  - `v1` - Major version (stable only)
  - `latest` - Latest stable release
- Generate SBOM

#### 5.4 Generate Manifests
- Create install.yaml for the release
- Generate CRDs
- Package examples
- Create tarball with manifests

#### 5.5 Security Scan
- Trivy vulnerability scanning
- Fail on CRITICAL vulnerabilities
- Upload to GitHub Security

#### 5.6 Create GitHub Release
- Generate changelog from git commits
- Create release with artifacts:
  - `install.yaml` - Installation manifest
  - `traktor-v1.2.3-manifests.tar.gz` - All manifests
  - `sbom-v1.2.3.spdx.json` - Software Bill of Materials
- Mark as prerelease if applicable
- Auto-generate release notes

#### 5.7 Update Documentation
- Update version in README
- Create PR with documentation updates
- Only for stable releases (not prereleases)

---

## Docker Images

### Registries

**Primary:** `docker.io/gdxbsv/traktor`

### Image Tags Strategy

| Tag Pattern | Example | Description | When Created |
|-------------|---------|-------------|--------------|
| `latest` | `latest` | Latest stable release | Push to main (stable) |
| `vX.Y.Z` | `v1.2.3` | Exact version | Tag push |
| `vX.Y` | `v1.2` | Minor version | Tag push |
| `vX` | `v1` | Major version | Tag push (stable only) |
| `main` | `main` | Main branch latest | Push to main |
| `develop` | `develop` | Develop branch | Push to develop |
| `main-abc123` | `main-abc1234` | Branch + SHA | Any push |
| `pr-123` | `pr-123` | Pull request | PR builds |

### Multi-Architecture Support

All images support:
- `linux/amd64` (x86_64)
- `linux/arm64` (ARM 64-bit)

Built using Docker Buildx with QEMU emulation.

---

## Secrets Required

Configure these secrets in GitHub repository settings:

| Secret | Description | Required For |
|--------|-------------|--------------|
| `DOCKER_USERNAME` | Docker Hub username | Image push |
| `DOCKER_PASSWORD` | Docker Hub token/password | Image push |
| `GITHUB_TOKEN` | GitHub Actions token | Releases (auto-provided) |
| `CODECOV_TOKEN` | Codecov upload token | Coverage (optional) |

### Setting up Docker Hub credentials:

1. Go to Docker Hub → Account Settings → Security
2. Create a new access token
3. Add to GitHub: Settings → Secrets → Actions
   - Name: `DOCKER_USERNAME` (your Docker Hub username)
   - Name: `DOCKER_PASSWORD` (the token from step 2)

---

## Release Process

### Creating a New Release

#### 1. Prepare Release Branch (Optional)
```bash
git checkout -b release/v1.2.3
# Make any final changes
git commit -m "chore: prepare v1.2.3 release"
git push origin release/v1.2.3
```

#### 2. Create and Push Tag

Tags can be created with or without the 'v' prefix:

```bash
git checkout main
git pull origin main

# Option 1: With 'v' prefix (recommended)
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3

# Option 2: Without 'v' prefix
git tag -a 1.2.3 -m "Release 1.2.3"
git push origin 1.2.3
```

#### 3. Automated Release Process
The release workflow will automatically:
- ✅ Validate tag format
- ✅ Run all tests
- ✅ Build multi-arch images
- ✅ Generate manifests
- ✅ Scan for vulnerabilities
- ✅ Create GitHub release with:
  - Changelog
  - Installation instructions
  - Docker image info
  - Manifest files
  - SBOM
- ✅ Update documentation

#### 4. Verify Release
Check:
- GitHub Releases page
- Docker Hub for new images
- Installation works: `kubectl apply -f <release-url>/install.yaml`

### Prerelease (Alpha/Beta/RC)

Tags can use either format:

```bash
# Alpha release (with 'v' prefix)
git tag -a v1.2.3-alpha.1 -m "Alpha release v1.2.3-alpha.1"
git push origin v1.2.3-alpha.1

# Beta release (without 'v' prefix)
git tag -a 1.2.3-beta.1 -m "Beta release 1.2.3-beta.1"
git push origin 1.2.3-beta.1

# Release candidate (with 'v' prefix)
git tag -a v1.2.3-rc.1 -m "Release candidate v1.2.3-rc.1"
git push origin v1.2.3-rc.1
```

Prereleases are marked as such in GitHub and won't update `latest` tag.

---

## Version Numbering

We follow [Semantic Versioning](https://semver.org/):

**Format:** `MAJOR.MINOR.PATCH[-PRERELEASE]`

- **MAJOR**: Incompatible API changes
- **MINOR**: New functionality (backwards compatible)
- **PATCH**: Bug fixes (backwards compatible)
- **PRERELEASE**: alpha, beta, rc

**Examples:**
- `v1.0.0` - First stable release
- `v1.1.0` - New feature added
- `v1.1.1` - Bug fix
- `v2.0.0` - Breaking change
- `v1.2.0-alpha.1` - Alpha version
- `v1.2.0-beta.1` - Beta version
- `v1.2.0-rc.1` - Release candidate

---

## Security Scanning

### Trivy Vulnerability Scanner

**Runs on:**
- Every build to main/develop
- Every release
- Pull requests (optional)

**Scans for:**
- OS vulnerabilities
- Application dependencies
- Known CVEs
- Configuration issues

**Severity Levels:**
- CRITICAL - Blocks release
- HIGH - Reported but doesn't block
- MEDIUM - Informational
- LOW - Informational

**Reports:**
- GitHub Security tab (SARIF format)
- Workflow logs (table format)
- Release artifacts (JSON format)

---

## Build Artifacts

Each build produces:

### 1. Docker Images
- Multi-architecture container images
- Pushed to Docker Hub
- Tagged according to strategy

### 2. Kubernetes Manifests
- `install.yaml` - Complete installation manifest
- CRDs, RBAC, Deployment, Service, etc.
- Ready to deploy with `kubectl apply`

### 3. SBOM (Software Bill of Materials)
- SPDX format JSON
- Lists all dependencies
- Security compliance

### 4. Coverage Reports
- Unit test coverage
- Uploaded to Codecov
- Available in PR comments

---

## Workflow Status Badges

Add these to your README.md:

```markdown
[![Tests](https://github.com/GDXbsv/traktor/actions/workflows/test.yml/badge.svg)](https://github.com/GDXbsv/traktor/actions/workflows/test.yml)
[![Lint](https://github.com/GDXbsv/traktor/actions/workflows/lint.yml/badge.svg)](https://github.com/GDXbsv/traktor/actions/workflows/lint.yml)
[![E2E Tests](https://github.com/GDXbsv/traktor/actions/workflows/test-e2e.yml/badge.svg)](https://github.com/GDXbsv/traktor/actions/workflows/test-e2e.yml)
[![Build](https://github.com/GDXbsv/traktor/actions/workflows/build.yml/badge.svg)](https://github.com/GDXbsv/traktor/actions/workflows/build.yml)
[![codecov](https://codecov.io/gh/GDXbsv/traktor/branch/main/graph/badge.svg)](https://codecov.io/gh/GDXbsv/traktor)
```

---

## Troubleshooting

### Build Fails on Docker Push

**Error:** `unauthorized: incorrect username or password`

**Solution:**
1. Verify `DOCKER_USERNAME` and `DOCKER_PASSWORD` secrets
2. Ensure Docker Hub token has push permissions
3. Check token hasn't expired

### Tests Fail in CI but Pass Locally

**Common causes:**
1. Missing dependencies in CI environment
2. Timing issues (use `Eventually()` in tests)
3. Namespace cleanup issues (fixed with unique names)

**Debug:**
```bash
# Run tests in CI-like environment
make test

# Check for race conditions
go test -race ./internal/controller/...
```

### Release Not Created

**Check:**
1. Tag format is correct: `v1.2.3` or `1.2.3` (both formats supported)
2. All tests passed
3. Security scan passed (no CRITICAL vulns)
4. GitHub token has proper permissions

### Image Not Pushed

**Verify:**
1. Workflow completed successfully
2. Docker Hub credentials are correct
3. Not a pull request (PRs don't push images)
4. Check Docker Hub rate limits

---

## Best Practices

### For Contributors

1. **Run tests locally** before pushing
   ```bash
   make test
   make lint
   ```

2. **Keep PRs focused** - one feature/fix per PR

3. **Write tests** for new features

4. **Update documentation** as needed

### For Maintainers

1. **Review PR checks** before merging

2. **Use semantic versioning** correctly

3. **Write meaningful release notes**

4. **Monitor security alerts**

5. **Keep dependencies updated**

---

## Local Development

### Test Workflows Locally

Use [act](https://github.com/nektos/act) to test workflows locally:

```bash
# Install act
brew install act  # macOS
# or
curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash

# Run tests workflow
act -j test

# Run build workflow
act -j build-and-test
```

### Build Docker Image Locally

```bash
# Build for your architecture
make docker-build

# Build multi-arch (requires buildx)
make docker-buildx
```

### Generate Manifests Locally

```bash
# Set image
export IMG=docker.io/gdxbsv/traktor:dev

# Generate install.yaml
make build-installer
```

---

## Monitoring

### GitHub Actions Dashboard

View workflow runs:
- Repository → Actions tab
- Filter by workflow, branch, status
- View logs and artifacts

### Docker Hub

Monitor images:
- https://hub.docker.com/r/gdxbsv/traktor
- Check pulls, tags, vulnerabilities

### Codecov

View coverage trends:
- https://codecov.io/gh/GDXbsv/traktor
- Track coverage over time
- Compare branches

---

## Support

For issues with CI/CD:
1. Check workflow logs in GitHub Actions
2. Review this documentation
3. Open an issue with workflow run link
4. Tag with `ci/cd` label

---

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Build Push Action](https://github.com/docker/build-push-action)
- [Semantic Versioning](https://semver.org/)
- [Trivy Security Scanner](https://github.com/aquasecurity/trivy)
- [SBOM Standard](https://spdx.dev/)