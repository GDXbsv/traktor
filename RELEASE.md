# Release Guide

Quick reference for creating releases of Traktor Operator.

## TL;DR - Quick Release

```bash
# 1. Create and push tag (with or without 'v' prefix)
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3

# 2. Wait for GitHub Actions to complete (~10-15 minutes)
# 3. Check https://github.com/GDXbsv/traktor/releases
```

That's it! Everything else is automated.

---

## Supported Tag Formats

Both formats are supported:

✅ **With 'v' prefix** (recommended):
```bash
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3
```

✅ **Without 'v' prefix**:
```bash
git tag -a 1.2.3 -m "Release 1.2.3"
git push origin 1.2.3
```

Both will trigger the same automated release process.

---

## What Happens Automatically

When you push a tag, the release workflow automatically:

1. ✅ **Validates** tag format
2. ✅ **Runs tests** (unit tests + linting)
3. ✅ **Builds** multi-architecture Docker images (amd64, arm64)
4. ✅ **Scans** for security vulnerabilities
5. ✅ **Generates** Kubernetes manifests
6. ✅ **Packages** Helm chart
7. ✅ **Creates** GitHub Release with:
   - Changelog from git commits
   - Installation instructions
   - Docker image tags
   - Kubernetes manifests
   - SBOM (Software Bill of Materials)
   - Helm chart package
8. ✅ **Publishes** Helm chart to GitHub Pages
9. ✅ **Updates** documentation (for stable releases)

---

## Release Types

### Stable Release

```bash
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3
```

Creates a stable release:
- Marked as latest release
- Updates `latest` Docker tag
- Updates documentation
- Published to Helm repository

### Prerelease (Alpha/Beta/RC)

```bash
# Alpha
git tag -a v1.2.3-alpha.1 -m "Alpha release"
git push origin v1.2.3-alpha.1

# Beta
git tag -a v1.2.3-beta.1 -m "Beta release"
git push origin v1.2.3-beta.1

# Release Candidate
git tag -a v1.2.3-rc.1 -m "Release candidate"
git push origin v1.2.3-rc.1
```

Prereleases:
- Marked as "Pre-release" in GitHub
- Does NOT update `latest` tag
- Does NOT update documentation automatically

---

## Semantic Versioning

We follow [Semantic Versioning](https://semver.org/):

**Format**: `MAJOR.MINOR.PATCH[-PRERELEASE]`

- **MAJOR**: Breaking changes / incompatible API changes
- **MINOR**: New features (backwards compatible)
- **PATCH**: Bug fixes (backwards compatible)

### Examples

| Version | Type | When to Use |
|---------|------|-------------|
| `v1.0.0` | Major | First stable release |
| `v1.1.0` | Minor | New feature added |
| `v1.1.1` | Patch | Bug fix |
| `v2.0.0` | Major | Breaking change |
| `v1.2.0-alpha.1` | Prerelease | Early testing version |
| `v1.2.0-beta.1` | Prerelease | Feature complete, testing |
| `v1.2.0-rc.1` | Prerelease | Release candidate |

---

## Step-by-Step Release Process

### 1. Prepare the Release

```bash
# Ensure you're on main branch
git checkout main
git pull origin main

# Verify everything is working
make test
make lint
make build

# Optional: Update CHANGELOG.md manually if needed
# (Automated changelog will be generated from git commits)
```

### 2. Create the Tag

Choose your version number based on semantic versioning:

```bash
# For a minor release (new features)
git tag -a v1.2.0 -m "Release v1.2.0

- Add feature X
- Add feature Y
- Improve performance of Z"

# For a patch release (bug fixes)
git tag -a v1.1.1 -m "Release v1.1.1

- Fix bug in controller reconciliation
- Fix memory leak in watcher"

# For a major release (breaking changes)
git tag -a v2.0.0 -m "Release v2.0.0

BREAKING CHANGES:
- API version upgraded to v1beta1
- Changed CRD structure"
```

### 3. Push the Tag

```bash
git push origin v1.2.0
```

### 4. Monitor the Release

1. Go to GitHub Actions: https://github.com/GDXbsv/traktor/actions
2. Watch the "Release" workflow run (~10-15 minutes)
3. Check for any failures

### 5. Verify the Release

Once complete, verify:

✅ **GitHub Release**: https://github.com/GDXbsv/traktor/releases
- Release created with correct version
- Changelog looks good
- All artifacts attached

✅ **Docker Hub**: https://hub.docker.com/r/gdxbsv/traktor
- Image tagged with version
- Multi-arch images present

✅ **Helm Repository**: https://gdxbsv.github.io/traktor
- Helm chart published
- Index updated

✅ **Installation Works**:
```bash
# Test kubectl installation
kubectl apply -f https://github.com/GDXbsv/traktor/releases/download/v1.2.0/install.yaml

# Test Helm installation
helm repo add traktor https://gdxbsv.github.io/traktor
helm repo update
helm install traktor traktor/traktor --version 1.2.0
```

---

## Hotfix Release

For urgent bug fixes on a released version:

```bash
# 1. Create hotfix branch from tag
git checkout -b hotfix/v1.2.1 v1.2.0

# 2. Make the fix
git commit -m "fix: critical bug in controller"

# 3. Create tag
git tag -a v1.2.1 -m "Hotfix v1.2.1

- Fix critical bug in controller"

# 4. Push tag (this triggers release)
git push origin v1.2.1

# 5. Merge back to main
git checkout main
git merge hotfix/v1.2.1
git push origin main
```

---

## Rollback a Release

If something goes wrong:

### Option 1: Delete the Tag (Before release completes)

```bash
# Delete local tag
git tag -d v1.2.3

# Delete remote tag
git push origin :refs/tags/v1.2.3
```

This will stop the release workflow if it hasn't completed yet.

### Option 2: Create a New Release (After release completes)

```bash
# Revert the changes
git revert <commit-sha>

# Create a new patch release
git tag -a v1.2.4 -m "Revert changes from v1.2.3"
git push origin v1.2.4
```

### Option 3: Mark Release as Draft

1. Go to GitHub Releases
2. Edit the release
3. Check "Set as a pre-release" or delete it

---

## Docker Image Tags

Each release creates these Docker image tags:

| Tag | Example | Description |
|-----|---------|-------------|
| Exact version | `v1.2.3` | Specific version |
| Minor version | `v1.2` | Latest patch in minor |
| Major version | `v1` | Latest minor in major (stable only) |
| Latest | `latest` | Latest stable release (stable only) |

**Note**: Prerelease versions don't update `latest` or major/minor tags.

---

## Helm Chart Publication

Helm charts are automatically published to GitHub Pages:

**Repository URL**: https://gdxbsv.github.io/traktor

**Usage**:
```bash
helm repo add traktor https://gdxbsv.github.io/traktor
helm repo update
helm search repo traktor
helm install traktor traktor/traktor --version 1.2.3
```

Charts are also attached to GitHub Releases.

---

## Troubleshooting

### Release Workflow Failed

**Check the logs**:
1. Go to Actions tab
2. Click on the failed workflow
3. Check which job failed

**Common issues**:

| Error | Solution |
|-------|----------|
| Tests failed | Fix tests, create new tag |
| Docker push failed | Check Docker Hub credentials in secrets |
| Critical vulnerability | Fix vulnerability, create new tag |
| Tag format invalid | Use correct format: `v1.2.3` or `1.2.3` |

### Tag Already Exists

```bash
# Delete and recreate
git tag -d v1.2.3
git push origin :refs/tags/v1.2.3
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3
```

### Docker Image Not Appearing

Wait a few minutes - multi-arch builds take time. Check:
1. Workflow completed successfully
2. Docker Hub rate limits not exceeded
3. Credentials are correct

### Helm Chart Not Published

Check:
1. `publish-helm-repo` job completed
2. GitHub Pages is enabled for `gh-pages` branch
3. Check gh-pages branch for chart files

---

## Best Practices

### Before Releasing

- ✅ Run tests locally: `make test`
- ✅ Run linter: `make lint`
- ✅ Update documentation if needed
- ✅ Review commits since last release
- ✅ Check for open critical issues

### Tag Message

Write clear, descriptive tag messages:

```bash
# Good ✅
git tag -a v1.2.0 -m "Release v1.2.0

Features:
- Add namespace filtering
- Improve reconciliation performance

Bug Fixes:
- Fix memory leak in watcher
- Fix nil pointer in controller

Documentation:
- Update README with new examples"

# Bad ❌
git tag -a v1.2.0 -m "new version"
```

### Versioning Guidelines

- **Patch** (v1.2.X): Bug fixes, small improvements
- **Minor** (v1.X.0): New features, non-breaking changes
- **Major** (vX.0.0): Breaking changes, major rewrites

---

## Manual Steps (If Automation Fails)

If the automated release fails, you can do it manually:

### 1. Build Docker Image

```bash
export IMG=docker.io/gdxbsv/traktor:v1.2.3
make docker-build
make docker-push
```

### 2. Generate Manifests

```bash
make build-installer IMG=$IMG
```

### 3. Create GitHub Release

1. Go to Releases → Draft a new release
2. Choose tag: v1.2.3
3. Title: Release v1.2.3
4. Add description
5. Attach `dist/install.yaml`
6. Publish release

### 4. Publish Helm Chart

```bash
# Package chart
helm package charts/traktor -d release-artifacts

# Update gh-pages
git checkout gh-pages
cp release-artifacts/*.tgz .
helm repo index . --url https://gdxbsv.github.io/traktor
git add .
git commit -m "Release Helm chart v1.2.3"
git push origin gh-pages
```

---

## CI/CD Pipeline Details

For more details about the release pipeline, see [.github/workflows/README.md](.github/workflows/README.md)

---

## Questions?

- 📖 Read the [full CI/CD documentation](.github/workflows/README.md)
- 🐛 Found a bug? [Open an issue](https://github.com/GDXbsv/traktor/issues)
- 💬 Need help? [Start a discussion](https://github.com/GDXbsv/traktor/discussions)

---

## Quick Reference

```bash
# Create stable release
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3

# Create prerelease
git tag -a v1.2.3-alpha.1 -m "Alpha release"
git push origin v1.2.3-alpha.1

# Delete tag
git tag -d v1.2.3
git push origin :refs/tags/v1.2.3

# View all tags
git tag -l

# View tag details
git show v1.2.3
```
