# Automated Version Synchronization

This document explains how Traktor automatically synchronizes version numbers across all components during releases.

## Overview

Traktor uses an automated approach to version management. When you push a git tag, the CI/CD pipeline automatically updates all version references throughout the project. This eliminates manual version updates and ensures consistency across all components.

## How It Works

### 1. Developer Action
```bash
# Developer only needs to create and push a tag
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3
```

### 2. Automated Pipeline
The GitHub Actions release workflow (`.github/workflows/release.yml`) automatically:

1. **Extracts version** from the git tag (`v1.2.3` → `1.2.3`)
2. **Updates Chart.yaml**:
   - `version: 1.2.3`
   - `appVersion: "1.2.3"`
   - Release URL: `https://github.com/GDXbsv/traktor/releases/tag/v1.2.3`
3. **Updates values.yaml**:
   - `image.tag: "1.2.3"`
4. **Builds and tags Docker images**:
   - `docker.io/gdxbsv/traktor:v1.2.3`
   - `docker.io/gdxbsv/traktor:v1.2` (minor version)
   - `docker.io/gdxbsv/traktor:v1` (major version)
   - `docker.io/gdxbsv/traktor:latest` (stable releases only)
5. **Packages Helm chart** with updated version
6. **Creates GitHub Release** with all artifacts

## Files That Are Auto-Updated

| File | What Gets Updated | Example |
|------|------------------|---------|
| `charts/traktor/Chart.yaml` | `version`, `appVersion`, changelog URL | `version: 1.2.3` |
| `charts/traktor/values.yaml` | Docker image tag | `tag: "1.2.3"` |
| `config/manager/kustomization.yaml` | Controller image during build | `newTag: v1.2.3` |
| Generated manifests | All image references | `image: gdxbsv/traktor:v1.2.3` |

## Implementation Details

### Release Workflow (`package-helm-chart` job)

```yaml
- name: Update Chart version
  run: |
    VERSION=${{ needs.validate-tag.outputs.version }}
    sed -i "s/^version:.*/version: ${VERSION}/" charts/traktor/Chart.yaml
    sed -i "s/^appVersion:.*/appVersion: \"${VERSION}\"/" charts/traktor/Chart.yaml
    sed -i "s|url: https://github.com/GDXbsv/traktor/releases/tag/v[0-9]*\.[0-9]*\.[0-9]*|url: https://github.com/GDXbsv/traktor/releases/tag/v${VERSION}|g" charts/traktor/Chart.yaml

- name: Update Chart image tag
  run: |
    VERSION=${{ needs.validate-tag.outputs.version }}
    sed -i "s/tag: \"\"/tag: \"${VERSION}\"/" charts/traktor/values.yaml
```

### Helper Script

For manual testing or local development, use the helper script:

```bash
./hack/update-chart-version.sh 1.2.3
```

This script performs the same updates as the CI/CD pipeline but for local use.

## Version Number Sources

```
                    ┌─────────────┐
                    │  Git Tag    │
                    │  v1.2.3     │
                    └──────┬──────┘
                           │
                           ↓
                ┌──────────────────────┐
                │  GitHub Actions      │
                │  Extract Version     │
                └──────────┬───────────┘
                           │
              ┌────────────┴─────────────┐
              ↓                          ↓
    ┌─────────────────┐        ┌─────────────────┐
    │  Chart.yaml     │        │  values.yaml    │
    │  version: 1.2.3 │        │  tag: "1.2.3"   │
    └─────────────────┘        └─────────────────┘
              │                          │
              └────────────┬─────────────┘
                           ↓
                 ┌──────────────────┐
                 │  Docker Images   │
                 │  :v1.2.3         │
                 │  :v1.2           │
                 │  :v1             │
                 │  :latest         │
                 └──────────────────┘
```

## Chart.yaml Template

The `Chart.yaml` maintains placeholder values that are replaced during release:

```yaml
# NOTE: version and appVersion are automatically updated during release by the CI/CD pipeline
version: 0.0.1
appVersion: "0.0.1"

# ...

annotations:
  # NOTE: The release URL below is automatically updated during release by the CI/CD pipeline
  artifacthub.io/changes: |
    - kind: added
      description: Initial release of Traktor operator
      links:
        - name: GitHub Release
          url: https://github.com/GDXbsv/traktor/releases/tag/v0.0.1
```

These placeholder values (`0.0.1`) are never committed with real version numbers. The CI/CD pipeline updates them dynamically during each release.

## Benefits

### ✅ Advantages

1. **No Manual Updates**: Developers never need to update version numbers manually
2. **Consistency**: All components use the same version number automatically
3. **Error Prevention**: Eliminates typos and mismatched versions
4. **Single Source of Truth**: Git tag is the only version source
5. **Audit Trail**: Git tags provide complete release history
6. **Rollback Safety**: Previous versions remain unchanged in git history

### ❌ What to Avoid

1. **Don't commit version updates**: The CI/CD handles this
2. **Don't manually edit Chart.yaml versions**: Use git tags instead
3. **Don't use different versions**: Everything syncs from the git tag
4. **Don't skip tagging**: No tag = no release

## Comparison: Before vs After

### Before (Manual Process) ❌

```bash
# 1. Update Chart.yaml manually
vim charts/traktor/Chart.yaml  # Change version to 1.2.3
vim charts/traktor/Chart.yaml  # Change appVersion to 1.2.3
vim charts/traktor/Chart.yaml  # Update changelog URL

# 2. Update values.yaml manually
vim charts/traktor/values.yaml  # Change tag to 1.2.3

# 3. Update Makefile
vim Makefile  # Change VERSION to 1.2.3

# 4. Commit changes
git add .
git commit -m "chore: bump version to 1.2.3"

# 5. Create tag
git tag -a v1.2.3 -m "Release 1.2.3"

# 6. Push everything
git push origin main --tags

# Risks: Typos, forgotten files, version mismatches
```

### After (Automated) ✅

```bash
# 1. Create and push tag - that's it!
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3

# Everything else happens automatically
# Zero risk of version mismatches
```

## Testing

### Test the Helper Script

```bash
# Check current version
./hack/update-chart-version.sh --current

# Preview changes (dry run)
./hack/update-chart-version.sh --dry-run 1.2.3

# Apply changes locally
./hack/update-chart-version.sh 1.2.3
git diff  # Review changes
git restore charts/  # Undo if needed
```

### Test the CI/CD Pipeline

Create a test tag to verify the automation:

```bash
# Create test tag
git tag -a v0.0.99-test -m "Test release automation"
git push origin v0.0.99-test

# Watch the workflow
# https://github.com/GDXbsv/traktor/actions

# Clean up test release
# Delete tag and release if needed
git tag -d v0.0.99-test
git push origin :refs/tags/v0.0.99-test
```

## Troubleshooting

### Version mismatch after release

**Problem**: Chart version doesn't match tag version.

**Solution**: The workflow might have failed. Check GitHub Actions logs.

### Chart.yaml committed with wrong version

**Problem**: Manually updated versions were committed.

**Solution**: 
```bash
# Revert to template version
git checkout origin/main -- charts/traktor/Chart.yaml
git commit -m "fix: revert Chart.yaml to template version"
git push
```

### Need to update version without releasing

**Problem**: Want to test with a different version locally.

**Solution**: Use the helper script (don't commit):
```bash
./hack/update-chart-version.sh 1.2.3-test
# Test locally
git restore charts/  # Undo changes
```

### Detached HEAD error in update-docs job

**Problem**: `peter-evans/create-pull-request` fails with:
```
Error: When the repository is checked out on a commit instead of a branch, 
the 'base' input must be supplied.
```

**Cause**: The `actions/checkout@v4` by default checks out the tag (detached HEAD state) when triggered by a tag push.

**Solution**: Already fixed in the workflow. The `update-docs` job now explicitly checks out the `main` branch:

```yaml
- name: Checkout code
  uses: actions/checkout@v4
  with:
    ref: main              # Checkout main branch instead of tag
    fetch-depth: 0         # Full history for proper git operations

- name: Create PR for version update
  uses: peter-evans/create-pull-request@v6
  with:
    base: main             # Specify base branch explicitly
    # ... other options
```

If you see this error, ensure your workflow has these two additions:
1. `ref: main` in the checkout step
2. `base: main` in the create-pull-request step

## Semantic Versioning

The automated system supports full semantic versioning:

| Tag Format | Chart Version | Docker Tags | Release Type |
|------------|---------------|-------------|--------------|
| `v1.2.3` | `1.2.3` | `v1.2.3`, `v1.2`, `v1`, `latest` | Stable |
| `v1.2.3-alpha.1` | `1.2.3-alpha.1` | `v1.2.3-alpha.1` | Prerelease |
| `v1.2.3-beta.1` | `1.2.3-beta.1` | `v1.2.3-beta.1` | Prerelease |
| `v1.2.3-rc.1` | `1.2.3-rc.1` | `v1.2.3-rc.1` | Prerelease |

Prerelease versions:
- Do NOT update `latest` tag
- Do NOT update major/minor tags (`v1`, `v1.2`)
- Are marked as "Pre-release" in GitHub

## Related Documentation

- **[RELEASE.md](../RELEASE.md)**: Complete release guide
- **[hack/README.md](../hack/README.md)**: Helper scripts documentation
- **[.github/workflows/release.yml](../.github/workflows/release.yml)**: CI/CD implementation
- **[charts/traktor/Chart.yaml](../charts/traktor/Chart.yaml)**: Helm chart metadata

## Contributing

When modifying the version sync logic:

1. Update both the CI/CD workflow AND the helper script
2. Keep them in sync (same sed commands)
3. Test with dry-run mode first
4. Document changes in this file
5. Test with a prerelease tag before stable release

## Questions?

- 📖 Read the [Release Guide](../RELEASE.md)
- 🐛 Found a bug? [Open an issue](https://github.com/GDXbsv/traktor/issues)
- 💬 Need help? [Start a discussion](https://github.com/GDXbsv/traktor/discussions)