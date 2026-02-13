# Artifact Hub Troubleshooting Guide

This guide helps you diagnose and fix Artifact Hub validation errors for the Traktor Helm repository.

## Table of Contents

- [Quick Diagnosis](#quick-diagnosis)
- [Common Errors](#common-errors)
- [How to Check Your Repository](#how-to-check-your-repository)
- [Fixing Invalid Metadata](#fixing-invalid-metadata)
- [Verification Steps](#verification-steps)

## Quick Diagnosis

### Error: "invalid repository id"

**Symptom**: Artifact Hub shows: `error getting repository metadata: invalid metadata: invalid repository id`

**Common Causes**:
1. Old chart versions have invalid annotations (especially `artifacthub.io/signKey`)
2. `artifacthub-repo.yml` is missing or has wrong format
3. `repositoryID` doesn't match Artifact Hub dashboard

**Quick Fix**: Run the cleanup script (see below)

## How to Check Your Repository

### 1. Check Files Directly in Browser

Open these URLs to see what Artifact Hub sees:

**Metadata File**:
```
https://gdxbsv.github.io/traktor/artifacthub-repo.yml
```

Should contain:
```yaml
repositoryID: 2a1be652-53bf-42c6-a2b5-7b87f26bc585
displayName: Traktor Operator
owners:
  - name: GDX Cloud
    email: support@gdxcloud.net
```

**Index File**:
```
https://gdxbsv.github.io/traktor/index.yaml
```

Look for invalid annotations like:
```yaml
artifacthub.io/signKey: |
  fingerprint: ...
  url: https://github.com/GDXbsv/traktor/blob/main/PUBLIC_KEY.asc
```

The `fingerprint: ...` is a placeholder and causes validation errors.

### 2. Check Artifact Hub Dashboard

1. Go to https://artifacthub.io/
2. Sign in with your GitHub account
3. Click your profile icon (top right) → **Control Panel**
4. Navigate to **Repositories** tab
5. Find your repository card (ID: `2a1be652-53bf-42c6-a2b5-7b87f26bc585`)
6. Look for error messages or warning indicators

### 3. Check Repository Page

Visit your Artifact Hub repository page:
```
https://artifacthub.io/packages/search?repo=traktor
```

or

```
https://artifacthub.io/packages/helm/traktor/traktor
```

If you see errors or the page doesn't load, there's a validation issue.

### 4. Use Command Line Tools

```bash
# Check if files are accessible
curl -I https://gdxbsv.github.io/traktor/artifacthub-repo.yml
curl -I https://gdxbsv.github.io/traktor/index.yaml

# Download and inspect
curl https://gdxbsv.github.io/traktor/index.yaml > index.yaml
grep -A 3 "signKey" index.yaml

# Check for invalid annotations
curl -s https://gdxbsv.github.io/traktor/index.yaml | grep -B 5 "fingerprint: \.\.\."
```

## Common Errors

### 1. Invalid signKey Annotation

**Problem**: Old chart versions have placeholder signKey values:
```yaml
artifacthub.io/signKey: |
  fingerprint: ...
  url: https://github.com/GDXbsv/traktor/blob/main/PUBLIC_KEY.asc
```

**Why it happens**: These annotations were in Chart.yaml but are now removed. However, old published charts still have them in the index.

**Solution**: Remove old chart versions from gh-pages branch (see below).

### 2. Missing artifacthub-repo.yml

**Problem**: File doesn't exist at repository root.

**Solution**: The file should be at the same level as `index.yaml` on the gh-pages branch.

### 3. Wrong Repository ID

**Problem**: `repositoryID` in `artifacthub-repo.yml` doesn't match the one in Artifact Hub dashboard.

**Solution**: 
1. Get the correct ID from Artifact Hub Control Panel
2. Update both:
   - `charts/artifacthub-repo.yml` (in main branch)
   - Update the workflow at `.github/workflows/release.yml` line ~470

### 4. Malformed YAML

**Problem**: YAML syntax errors in metadata file.

**Solution**: Validate YAML:
```bash
# Install yq if needed
brew install yq  # macOS
# or
sudo apt install yq  # Ubuntu

# Validate
curl -s https://gdxbsv.github.io/traktor/artifacthub-repo.yml | yq .
```

## Fixing Invalid Metadata

### Method 1: Automated Cleanup Script (Recommended)

We've created a script to clean up invalid charts:

```bash
# From repository root
chmod +x scripts/cleanup-helm-repo.sh
./scripts/cleanup-helm-repo.sh
```

The script will:
1. ✅ Switch to gh-pages branch
2. ✅ Remove old chart versions with invalid metadata
3. ✅ Regenerate index.yaml
4. ✅ Update artifacthub-repo.yml
5. ✅ Commit and push changes

### Method 2: Manual Cleanup

If you prefer to do it manually:

```bash
# Clone repository
git clone https://github.com/GDXbsv/traktor.git
cd traktor

# Switch to gh-pages branch
git checkout gh-pages
git pull origin gh-pages

# Remove old charts with invalid metadata
rm -f traktor-0.0.1.tgz
rm -f traktor-0.0.2.tgz
rm -f traktor-0.0.4.tgz
rm -f traktor-0.0.5.tgz

# Regenerate index (requires Helm installed)
helm repo index . --url https://gdxbsv.github.io/traktor

# Ensure artifacthub-repo.yml is correct
cat > artifacthub-repo.yml << 'EOF'
# Artifact Hub repository metadata file
# https://artifacthub.io/docs/topics/repositories/

repositoryID: 2a1be652-53bf-42c6-a2b5-7b87f26bc585
displayName: Traktor Operator
owners:
  - name: GDX Cloud
    email: support@gdxcloud.net
EOF

# Commit changes
git add index.yaml artifacthub-repo.yml
git add -u *.tgz
git commit -m "chore: remove charts with invalid Artifact Hub metadata"

# Push to GitHub
git push origin gh-pages

# Return to main branch
git checkout main
```

### Method 3: Force Re-publish Latest Version

If you just want to fix the current version without cleaning up:

```bash
# Trigger a new release
git tag -d v0.0.6  # Delete locally if exists
git push origin :refs/tags/v0.0.6  # Delete remotely if exists

# Create fresh release
git tag v0.0.6
git push origin v0.0.6
```

The release workflow now includes automatic cleanup of old invalid charts.

## Verification Steps

After applying fixes:

### 1. Verify Files Are Updated

```bash
# Check metadata file
curl https://gdxbsv.github.io/traktor/artifacthub-repo.yml

# Check index doesn't have invalid signKey
curl -s https://gdxbsv.github.io/traktor/index.yaml | grep "fingerprint: \.\.\."
# Should return nothing if fixed
```

### 2. Test Helm Repository

```bash
# Add repository
helm repo add traktor https://gdxbsv.github.io/traktor
helm repo update

# Search for charts
helm search repo traktor

# Install (optional)
helm install test-traktor traktor/traktor --dry-run
```

### 3. Wait for Artifact Hub Re-indexing

Artifact Hub re-indexes repositories periodically:
- **Frequency**: Every 30 minutes to 2 hours
- **Trigger**: When `index.yaml` changes (the `generated` field is ignored)

You can force a re-index:
1. Go to Artifact Hub Control Panel
2. Find your repository
3. Click the **refresh icon** or **re-index** button

### 4. Check Artifact Hub Dashboard

After 30 minutes to 2 hours:
1. Visit https://artifacthub.io/packages/helm/traktor/traktor
2. Check for:
   - ✅ No error messages
   - ✅ Chart versions display correctly
   - ✅ Metadata shows properly
   - ✅ README renders

### 5. Verify Badge in README

The badge in your README should work:
```markdown
[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/traktor)](https://artifacthub.io/packages/helm/traktor/traktor)
```

## Prevention: Avoid Future Issues

### 1. Chart.yaml Best Practices

**DO**:
```yaml
annotations:
  artifacthub.io/license: Apache-2.0
  artifacthub.io/operator: "true"
  artifacthub.io/prerelease: "false"
```

**DON'T**:
```yaml
annotations:
  artifacthub.io/signKey: |
    fingerprint: ...  # ❌ Don't use placeholders
```

### 2. Test Before Releasing

Before creating a release:

```bash
# Validate Chart.yaml
helm lint charts/traktor

# Check for invalid annotations
grep -r "fingerprint: \.\.\." charts/traktor/Chart.yaml
# Should return nothing

# Package and check
helm package charts/traktor
tar -xzf traktor-*.tgz -O traktor/Chart.yaml | grep signKey
# Should return nothing or valid key
```

### 3. CI/CD Validation

The release workflow now includes automatic cleanup:

```yaml
# In .github/workflows/release.yml
- name: Update Helm repository
  run: |
    # Remove old chart versions with invalid metadata
    rm -f gh-pages/traktor-0.0.1.tgz || true
    rm -f gh-pages/traktor-0.0.2.tgz || true
    # ... etc
```

### 4. Monitor Artifact Hub

Set up monitoring:
1. Subscribe to your repository on Artifact Hub (bell icon)
2. Check Artifact Hub Control Panel weekly
3. Monitor GitHub Pages deployments

## Troubleshooting Checklist

Use this checklist when investigating issues:

- [ ] Files are accessible at URLs above
- [ ] `artifacthub-repo.yml` has correct repositoryID
- [ ] `artifacthub-repo.yml` is valid YAML
- [ ] No invalid `signKey` annotations in index.yaml
- [ ] No placeholder values (e.g., `fingerprint: ...`)
- [ ] GitHub Pages is enabled for gh-pages branch
- [ ] Repository is public (or Artifact Hub configured for private)
- [ ] Waited at least 30 minutes after changes for re-indexing
- [ ] Repository exists in Artifact Hub Control Panel
- [ ] No error messages in Artifact Hub dashboard

## Getting Help

If issues persist:

1. **Check Artifact Hub Docs**: https://artifacthub.io/docs
2. **Artifact Hub GitHub**: https://github.com/artifacthub/hub/issues
3. **Our Issues**: https://github.com/GDXbsv/traktor/issues
4. **Helm Docs**: https://helm.sh/docs/

## Example: Complete Working Setup

### File: `artifacthub-repo.yml`
```yaml
repositoryID: 2a1be652-53bf-42c6-a2b5-7b87f26bc585
displayName: Traktor Operator
owners:
  - name: GDX Cloud
    email: support@gdxcloud.net
```

### Chart.yaml Annotations (Valid)
```yaml
annotations:
  artifacthub.io/license: Apache-2.0
  artifacthub.io/operator: "true"
  artifacthub.io/operatorCapabilities: Basic Install
  artifacthub.io/prerelease: "false"
  artifacthub.io/containsSecurityUpdates: "false"
  artifacthub.io/crds: |
    - kind: SecretsRefresh
      version: v1alpha1
      name: secretsrefreshes.traktor.gdxcloud.net
  artifacthub.io/links: |
    - name: Documentation
      url: https://github.com/GDXbsv/traktor/blob/main/README.md
```

### Expected Repository Structure (gh-pages branch)
```
.
├── artifacthub-repo.yml
├── index.yaml
└── traktor-0.0.6.tgz
```

## Summary

**Most Common Issue**: Old chart versions with invalid `signKey` annotations.

**Quickest Fix**: Run `scripts/cleanup-helm-repo.sh`

**Verification**: Check https://gdxbsv.github.io/traktor/index.yaml doesn't contain `fingerprint: ...`

**Wait Time**: 30 minutes to 2 hours for Artifact Hub to re-index

**Success Indicator**: https://artifacthub.io/packages/helm/traktor/traktor loads without errors