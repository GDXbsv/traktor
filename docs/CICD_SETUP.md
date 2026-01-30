# CI/CD Quick Setup Guide

This guide will help you set up the complete CI/CD pipeline for the Traktor operator in under 10 minutes.

## Prerequisites

- GitHub repository with admin access
- Docker Hub account
- Git installed locally

## Step 1: Configure GitHub Secrets

1. Go to your GitHub repository
2. Navigate to: **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret**

Add the following secrets:

### Required Secrets

| Secret Name | Value | How to Get |
|-------------|-------|------------|
| `DOCKER_USERNAME` | Your Docker Hub username | Your Docker Hub login username |
| `DOCKER_PASSWORD` | Docker Hub access token | Create at hub.docker.com → Account Settings → Security → New Access Token |

### Creating Docker Hub Access Token

1. Go to https://hub.docker.com
2. Log in to your account
3. Click your username → **Account Settings**
4. Go to **Security** tab
5. Click **New Access Token**
6. Give it a description (e.g., "GitHub Actions")
7. Set permissions: **Read & Write**
8. Copy the token (you won't see it again!)
9. Add to GitHub secrets as `DOCKER_PASSWORD`

## Step 2: Verify Workflows Exist

Check that these files exist in `.github/workflows/`:

```bash
ls -la .github/workflows/
```

You should see:
- ✅ `test.yml` - Unit tests
- ✅ `lint.yml` - Code linting
- ✅ `test-e2e.yml` - E2E tests
- ✅ `build.yml` - Build and push images
- ✅ `release.yml` - Automated releases

## Step 3: Test the Pipeline

### 3.1 Test Unit Tests

Push any change to trigger tests:

```bash
git add .
git commit -m "test: trigger CI pipeline"
git push origin main
```

Go to **Actions** tab in GitHub and verify:
- ✅ Tests workflow runs
- ✅ Lint workflow runs
- ✅ All checks pass

### 3.2 Test Docker Build (Optional for PRs)

Create a pull request or push to `main` or `develop`:

```bash
git checkout -b feature/test-ci
git push origin feature/test-ci
# Create PR in GitHub UI
```

This will trigger:
- ✅ Test workflow
- ✅ Lint workflow
- ✅ Build workflow (images NOT pushed for PRs)

### 3.3 Test Image Push

Push to `main` branch:

```bash
git checkout main
git merge feature/test-ci
git push origin main
```

This will:
- ✅ Run all tests
- ✅ Build Docker images
- ✅ Push to Docker Hub
- ✅ Generate manifests

Check Docker Hub: https://hub.docker.com/r/gdxbsv/traktor

You should see new tags:
- `latest`
- `main`
- `main-<commit-sha>`

## Step 4: Create Your First Release

### 4.1 Verify Everything Works

```bash
# Run tests locally
make test

# Build locally
make docker-build
```

### 4.2 Create Release Tag

```bash
# Make sure you're on main branch
git checkout main
git pull origin main

# Create and push tag
git tag -a v0.0.1 -m "Release v0.0.1 - Initial release"
git push origin v0.0.1
```

### 4.3 Watch the Release Process

1. Go to **Actions** tab
2. Watch the **Release** workflow run
3. It will:
   - ✅ Run all tests
   - ✅ Build multi-arch images (amd64, arm64)
   - ✅ Generate Kubernetes manifests
   - ✅ Scan for vulnerabilities
   - ✅ Create GitHub Release

### 4.4 Verify Release

1. Go to **Releases** tab
2. You should see "Release v0.0.1"
3. It includes:
   - 📄 `install.yaml` - Installation manifest
   - 📦 `traktor-v0.0.1-manifests.tar.gz` - All manifests
   - 🔒 `sbom-v0.0.1.spdx.json` - Software Bill of Materials
   - 📝 Changelog
   - 🐳 Docker image info

4. Check Docker Hub tags:
   - `v0.0.1`
   - `v0.0`
   - `v0`
   - `latest`

## Step 5: Test Installation from Release

```bash
# Install from release
kubectl apply -f https://github.com/GDXbsv/traktor/releases/download/v0.0.1/install.yaml

# Or using Docker image
kubectl set image deployment/traktor-controller-manager \
  manager=docker.io/gdxbsv/traktor:v0.0.1 \
  -n traktor-system
```

## Workflow Triggers Summary

| Workflow | Trigger | What Happens |
|----------|---------|--------------|
| **Test** | Push to main/develop, PRs | Run unit tests, upload coverage |
| **Lint** | Any push, PRs | Code quality checks |
| **E2E** | Any push, PRs | End-to-end tests with Kind |
| **Build** | Push to main/develop, PRs, tags | Build images, push if not PR |
| **Release** | Tag push (v*.*.*) | Full release process |

## Customization

### Change Docker Registry

Edit workflows and `Makefile`:

```yaml
# In .github/workflows/*.yml
env:
  REGISTRY: ghcr.io  # or quay.io, gcr.io, etc.
  IMAGE_NAME: your-org/traktor
```

```makefile
# In Makefile
IMG ?= ghcr.io/your-org/traktor:latest
```

### Add Additional Checks

Create new workflow file:

```yaml
# .github/workflows/security.yml
name: Security Scan
on: [push, pull_request]
jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run security scan
        run: make security-scan
```

### Change Test Timeout

Edit `.github/workflows/test.yml`:

```yaml
- name: Run unit tests
  run: make test
  timeout-minutes: 10  # Add timeout
```

## Troubleshooting

### Tests Fail: "namespace is being terminated"

**Fixed!** Tests now use unique namespace names.

If you still see this, run:
```bash
make test
```

### Docker Push Fails: "unauthorized"

1. Check secrets are set correctly
2. Verify Docker Hub token hasn't expired
3. Ensure token has Read & Write permissions
4. Try creating a new token

### Release Not Created

1. Check tag format: `v1.2.3` (starts with v)
2. Ensure all tests passed
3. Check workflow logs for errors
4. Verify GitHub token permissions

### Image Tag Not Showing on Docker Hub

Wait a few minutes - it can take time to process multi-arch builds.

Check workflow logs:
```
Actions → Build and Push → build-docker job → View logs
```

## Monitoring

### GitHub Actions Dashboard

```
Repository → Actions tab
```

- View all workflow runs
- Filter by status, workflow, branch
- Download artifacts
- Re-run failed jobs

### Docker Hub

```
https://hub.docker.com/r/gdxbsv/traktor
```

- View all tags
- Check vulnerability scan
- See pull statistics
- Manage tags

### Coverage Reports

```
https://codecov.io/gh/GDXbsv/traktor
```

- View coverage trends
- Compare branches
- See coverage diff on PRs

## Next Steps

1. ✅ **Add Status Badges** to README.md
   ```markdown
   [![Tests](https://github.com/GDXbsv/traktor/actions/workflows/test.yml/badge.svg)](https://github.com/GDXbsv/traktor/actions/workflows/test.yml)
   ```

2. ✅ **Set up Branch Protection**
   - Settings → Branches → Add rule
   - Require status checks to pass
   - Require tests + lint before merging

3. ✅ **Enable Dependabot**
   - Settings → Security → Dependabot
   - Enable version updates
   - Auto-update dependencies

4. ✅ **Configure Codecov**
   - Sign up at codecov.io
   - Add repository
   - Get token and add to secrets (optional)

## Complete Setup Checklist

- [ ] GitHub secrets configured (DOCKER_USERNAME, DOCKER_PASSWORD)
- [ ] Pushed code to trigger tests
- [ ] Tests workflow passed
- [ ] Build workflow pushed image to Docker Hub
- [ ] Created first release tag (v0.0.1)
- [ ] Release workflow completed
- [ ] GitHub Release created with artifacts
- [ ] Docker images available with multiple tags
- [ ] Tested installation from release
- [ ] Added status badges to README
- [ ] Set up branch protection rules

## Support

If you encounter issues:

1. Check workflow logs in Actions tab
2. Review `.github/workflows/README.md` for detailed docs
3. Open an issue with:
   - Workflow run link
   - Error message
   - Steps to reproduce

## Resources

- [Detailed CI/CD Documentation](.github/workflows/README.md)
- [GitHub Actions Docs](https://docs.github.com/en/actions)
- [Docker Hub](https://hub.docker.com)
- [Semantic Versioning](https://semver.org/)

---

**Congratulations! Your CI/CD pipeline is now fully configured! 🎉**

Every push will be tested, linted, and built automatically. Creating a new release is as simple as pushing a tag.