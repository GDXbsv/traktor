# Artifact Hub Setup Guide

This guide explains how to publish the Traktor Helm chart to Artifact Hub and GitHub Pages.

## Overview

Artifact Hub is a web-based application that enables finding, installing, and publishing packages and configurations for CNCF projects. It provides a central place for discovering Helm charts.

## Prerequisites

- GitHub repository with Helm chart
- GitHub Pages enabled
- Repository admin access

## Setup Steps

### Step 1: Enable GitHub Pages

1. Go to your GitHub repository
2. Navigate to **Settings** → **Pages**
3. Under "Source", select:
   - Branch: `gh-pages`
   - Folder: `/ (root)`
4. Click **Save**
5. Note your GitHub Pages URL: `https://gdxbsv.github.io/traktor`

### Step 2: Create gh-pages Branch

```bash
# Create orphan branch for GitHub Pages
git checkout --orphan gh-pages
git rm -rf .

# Create initial index
cat > index.html << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>Traktor Helm Repository</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        h1 { color: #333; }
        code { background: #f4f4f4; padding: 2px 6px; border-radius: 3px; }
        pre { background: #f4f4f4; padding: 15px; border-radius: 5px; overflow-x: auto; }
    </style>
</head>
<body>
    <h1>Traktor Helm Repository</h1>
    <p>Official Helm repository for Traktor - Kubernetes operator that automatically restarts deployments when secrets change.</p>
    
    <h2>Add Repository</h2>
    <pre>helm repo add traktor https://gdxbsv.github.io/traktor
helm repo update</pre>
    
    <h2>Install Chart</h2>
    <pre>helm install traktor traktor/traktor</pre>
    
    <h2>Links</h2>
    <ul>
        <li><a href="https://github.com/GDXbsv/traktor">GitHub Repository</a></li>
        <li><a href="https://artifacthub.io/packages/helm/traktor/traktor">Artifact Hub</a></li>
        <li><a href="index.yaml">Helm Repository Index</a></li>
    </ul>
</body>
</html>
EOF

# Create Artifact Hub metadata
cat > artifacthub-repo.yml << 'EOF'
repositoryID: traktor
owners:
  - name: GDX Cloud
    email: support@gdxcloud.net
EOF

# Commit and push
git add .
git commit -m "Initial GitHub Pages setup"
git push origin gh-pages
```

### Step 3: Verify GitHub Pages

Wait a few minutes, then check:
```
https://gdxbsv.github.io/traktor
```

You should see the Helm repository homepage.

### Step 4: Register on Artifact Hub

1. Go to https://artifacthub.io
2. Sign in with your GitHub account
3. Click **Add repository** in the control panel
4. Fill in the form:
   - **Kind**: Helm charts
   - **Repository name**: traktor
   - **Display name**: Traktor Operator
   - **URL**: `https://gdxbsv.github.io/traktor`
   - **Official**: Yes (if you're the maintainer)
   - **Verified Publisher**: Request verification

5. Click **Add**

### Step 5: Verify on Artifact Hub

After a few minutes, your chart will appear on Artifact Hub:
```
https://artifacthub.io/packages/helm/traktor/traktor
```

## Automatic Updates

The workflow `.github/workflows/helm-repo.yml` automatically:

1. **On new version tag** (e.g., `v1.0.0`):
   - Packages the Helm chart
   - Updates the version in Chart.yaml
   - Publishes to GitHub Pages
   - Updates the Helm repository index
   - Artifact Hub automatically syncs within 30 minutes

2. **What happens**:
   ```
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   
   → Workflow runs
   → Chart packaged: traktor-1.0.0.tgz
   → Pushed to gh-pages branch
   → Available at: https://gdxbsv.github.io/traktor
   → Synced to Artifact Hub
   ```

## Artifact Hub Metadata

The chart includes Artifact Hub annotations in `Chart.yaml`:

```yaml
annotations:
  artifacthub.io/license: Apache-2.0
  artifacthub.io/operator: "true"
  artifacthub.io/operatorCapabilities: Basic Install
  artifacthub.io/prerelease: "false"
  artifacthub.io/crds: |
    - kind: SecretsRefresh
      version: v1alpha1
      name: secretsrefreshes.traktor.gdxcloud.net
  artifacthub.io/links: |
    - name: Documentation
      url: https://github.com/GDXbsv/traktor/blob/main/README.md
```

These annotations provide rich metadata on Artifact Hub.

## Using the Helm Repository

### Add Repository

```bash
helm repo add traktor https://gdxbsv.github.io/traktor
helm repo update
```

### Search for Charts

```bash
helm search repo traktor
```

Output:
```
NAME              CHART VERSION  APP VERSION  DESCRIPTION
traktor/traktor   1.0.0          1.0.0        A Kubernetes operator that automatically rest...
```

### Install Chart

```bash
# Install latest version
helm install traktor traktor/traktor

# Install specific version
helm install traktor traktor/traktor --version 1.0.0

# Install with custom values
helm install traktor traktor/traktor -f values.yaml
```

### Upgrade Chart

```bash
# Update repository
helm repo update

# Upgrade to latest
helm upgrade traktor traktor/traktor

# Upgrade to specific version
helm upgrade traktor traktor/traktor --version 1.1.0
```

## Artifact Hub Features

Once published, your chart will have:

### 1. Chart Page
- **Description** and **README**
- **Installation instructions**
- **Values.yaml** documentation
- **Version history**
- **Security report** (if vulnerabilities found)

### 2. Discovery
- **Searchable** on artifacthub.io
- **Categories** and **keywords**
- **Verified publisher badge** (after verification)

### 3. Metadata
- **CRD documentation**
- **Example configurations**
- **Links** to GitHub, docs, etc.
- **Maintainer information**

### 4. Notifications
- **Webhook** for new versions
- **Email alerts** for users tracking the chart
- **Security alerts** for vulnerabilities

## Verification Badge

To get the "Verified Publisher" badge:

1. Go to your Artifact Hub settings
2. Add your GitHub repository
3. Add `artifacthub-repo.yml` to your repository
4. Request verification in Artifact Hub
5. Artifact Hub team will verify ownership

## Custom Artifact Hub Page

Enhance your chart's Artifact Hub page:

### 1. Add Screenshots

Create `charts/traktor/screenshots/` and reference in Chart.yaml:

```yaml
annotations:
  artifacthub.io/screenshots: |
    - title: Operator Dashboard
      url: https://raw.githubusercontent.com/GDXbsv/traktor/main/docs/images/dashboard.png
```

### 2. Add Videos

```yaml
annotations:
  artifacthub.io/videos: |
    - title: Getting Started
      url: https://www.youtube.com/watch?v=xxxxx
```

### 3. Security Report

Artifact Hub automatically scans for:
- CVEs in container images
- Deprecated Kubernetes APIs
- Security best practices

## Monitoring

### Check Sync Status

1. Go to Artifact Hub control panel
2. View your repository
3. Check "Last scan" timestamp
4. View any errors or warnings

### Manual Sync

If changes don't appear:
1. Go to repository settings on Artifact Hub
2. Click "Request rescan"
3. Wait a few minutes

## Troubleshooting

### Chart Not Appearing on Artifact Hub

**Check:**
1. GitHub Pages is enabled and working
2. `artifacthub-repo.yml` exists in gh-pages branch
3. Repository is registered on Artifact Hub
4. Wait 30 minutes for automatic sync

**Verify GitHub Pages:**
```bash
curl -I https://gdxbsv.github.io/traktor/index.yaml
```

Should return `200 OK`.

### Invalid Repository

**Error:** "Repository not found"

**Solution:**
- Verify URL: `https://gdxbsv.github.io/traktor` (no trailing slash)
- Check index.yaml exists and is valid
- Ensure gh-pages branch is published

### Metadata Not Showing

**Error:** Annotations not appearing on Artifact Hub

**Solution:**
- Check Chart.yaml syntax
- Ensure annotations use correct format
- Request rescan on Artifact Hub

### Version Not Updating

**Error:** Old version still showing

**Solution:**
- Verify workflow ran successfully
- Check gh-pages branch has new package
- Request manual rescan on Artifact Hub
- Clear browser cache

## Best Practices

### 1. Semantic Versioning

Follow semantic versioning for chart versions:
```
v1.0.0 - Initial release
v1.1.0 - New features
v1.1.1 - Bug fixes
v2.0.0 - Breaking changes
```

### 2. Changelog

Update Chart.yaml annotations with each release:
```yaml
artifacthub.io/changes: |
  - kind: added
    description: Add support for network policies
  - kind: changed
    description: Update default memory limits
  - kind: fixed
    description: Fix RBAC permissions
```

### 3. Documentation

Keep chart README up to date:
- Installation instructions
- Configuration options
- Examples
- Troubleshooting

### 4. Security

- Scan images for vulnerabilities
- Use specific image tags (not `latest`)
- Document security considerations
- Keep dependencies updated

## Analytics

Artifact Hub provides analytics:
- **Downloads** per version
- **Repository views**
- **Installs** tracking
- **Popular versions**

Access in your Artifact Hub dashboard.

## Support

### Artifact Hub Support
- Documentation: https://artifacthub.io/docs
- GitHub: https://github.com/artifacthub/hub
- Slack: CNCF Slack #artifact-hub

### Traktor Support
- GitHub Issues: https://github.com/GDXbsv/traktor/issues
- Documentation: https://github.com/GDXbsv/traktor

## Quick Reference

```bash
# Add repository
helm repo add traktor https://gdxbsv.github.io/traktor

# Update repositories
helm repo update

# Search for chart
helm search repo traktor

# View chart info
helm show chart traktor/traktor
helm show values traktor/traktor
helm show readme traktor/traktor

# Install chart
helm install traktor traktor/traktor

# Upgrade chart
helm upgrade traktor traktor/traktor

# Uninstall chart
helm uninstall traktor
```

## Release Checklist

- [ ] Update Chart.yaml version
- [ ] Update Chart.yaml appVersion
- [ ] Update CHANGELOG in annotations
- [ ] Tag release: `git tag -a vX.Y.Z`
- [ ] Push tag: `git push origin vX.Y.Z`
- [ ] Workflow publishes to GitHub Pages
- [ ] Verify on: `https://gdxbsv.github.io/traktor`
- [ ] Wait for Artifact Hub sync (30 mins)
- [ ] Verify on: `https://artifacthub.io/packages/helm/traktor/traktor`
- [ ] Announce release

---

**Your Helm chart will now be discoverable on Artifact Hub! 🎉**