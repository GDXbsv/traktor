# GitHub Actions Permissions Configuration

This guide explains how to configure GitHub repository settings to allow the release workflow to create releases and perform other actions.

## Issue: "Resource not accessible by integration"

If you see this error when creating releases:
```
⚠️ GitHub release failed with status: 403
{"message":"Resource not accessible by integration","documentation_url":"https://docs.github.com/rest/releases/releases#generate-release-notes-content-for-a-release","status":"403"}
Error: Resource not accessible by integration
```

This means GitHub Actions doesn't have the required permissions at the repository level.

## Solution: Configure Repository Permissions

### Step 1: Navigate to Repository Settings

1. Go to your repository: https://github.com/GDXbsv/traktor
2. Click **Settings** (top right)
3. In the left sidebar, scroll down to **Actions** section
4. Click **General**

### Step 2: Configure Workflow Permissions

Scroll down to the **Workflow permissions** section (near the bottom).

You'll see two options:

#### Option 1: Read and write permissions (Recommended)
Select: **"Read and write permissions"**

This allows workflows to:
- ✅ Create releases
- ✅ Push commits
- ✅ Create/update branches
- ✅ Upload release assets
- ✅ Update pull requests

**Additional setting:**
☑️ Check: **"Allow GitHub Actions to create and approve pull requests"**

This is needed if you want workflows to create PRs (though we've switched to direct commits).

#### Option 2: Read repository contents and packages permissions
This is the restrictive default. **Do NOT use this** - it prevents workflows from creating releases.

### Step 3: Save Changes

Click **Save** at the bottom of the page.

### Step 4: Re-run Failed Workflow

After changing settings:

1. Go to **Actions** tab
2. Find the failed workflow run for tag `v0.0.9`
3. Click **Re-run failed jobs** or **Re-run all jobs**

OR trigger a new release:
```bash
git tag -d v0.0.9
git push origin :refs/tags/v0.0.9
git tag v0.0.9
git push origin v0.0.9
```

## Expected Settings

### ✅ Correct Configuration

```
Workflow permissions:
  ⦿ Read and write permissions
  ☑ Allow GitHub Actions to create and approve pull requests
```

### ❌ Incorrect Configuration

```
Workflow permissions:
  ⦿ Read repository contents and packages permissions
  ☐ Allow GitHub Actions to create and approve pull requests
```

## What Each Permission Does

| Permission | What It Allows |
|------------|----------------|
| **Read and write** | Full access to repository content, releases, packages |
| **Read only** | Can only read, cannot create releases or push code |
| **Create PRs** | Workflows can create pull requests (optional) |

## Permissions in Workflow Files

Even with repository-level permissions enabled, you should still specify job-level permissions for security:

### Example: create-release job

```yaml
create-release:
  name: Create GitHub Release
  runs-on: ubuntu-latest
  permissions:
    contents: write          # Required: Create releases, upload assets
    pull-requests: read      # Optional: Generate release notes
    repository-projects: read # Optional: Link to projects
```

### Why Both Levels?

1. **Repository-level permissions**: Global setting that enables features
2. **Workflow-level permissions**: Fine-grained control for security (principle of least privilege)

Even if a workflow requests `contents: write`, it won't work if the repository setting is "Read only".

## Security Considerations

### Read and Write Permissions

**Pros:**
- ✅ Workflows can fully automate releases
- ✅ Can update documentation automatically
- ✅ Can push version updates
- ✅ Simpler workflow configuration

**Cons:**
- ⚠️ More powerful, requires trust in workflow files
- ⚠️ Malicious PRs could potentially modify workflows

**Mitigation:**
- ✅ Require code review for workflow changes
- ✅ Use branch protection on `main`
- ✅ Use `pull_request_target` carefully
- ✅ Limit permissions in individual jobs

### Read-Only Permissions

**Pros:**
- ✅ More restrictive, harder to abuse
- ✅ Workflows can't modify repository

**Cons:**
- ❌ Can't create releases automatically
- ❌ Can't update documentation
- ❌ Requires manual steps for releases

**When to use:**
- Public repositories with many contributors
- When you prefer manual release approval
- When using external release tools

## Recommended Setup for Traktor

For the Traktor repository, we recommend:

1. **Enable "Read and write permissions"** - Needed for automated releases
2. **Uncheck "Allow GitHub Actions to create PRs"** - We use direct commits instead
3. **Use branch protection** on `main` branch
4. **Require approval** for workflow file changes

## Branch Protection Rules (Recommended)

To add an extra layer of security:

1. Go to **Settings** → **Branches**
2. Click **Add branch protection rule**
3. Branch name pattern: `main`
4. Enable:
   - ☑️ Require a pull request before merging
   - ☑️ Require approvals (at least 1)
   - ☑️ Require status checks to pass
   - ☑️ Include administrators

This prevents direct pushes to `main` (except from workflows and admins).

## Workflow-Specific Permissions

### Release Workflow

```yaml
create-release:
  permissions:
    contents: write          # Create releases
    pull-requests: read      # Read for release notes
    repository-projects: read # Read project info
```

### Build Workflow

```yaml
build-docker:
  permissions:
    contents: write   # Push to gh-pages
    packages: write   # Push Docker images
```

### Update Docs Workflow

```yaml
update-docs:
  permissions:
    contents: write   # Commit to main branch
```

## Troubleshooting

### Error: "Resource not accessible by integration"

**Cause:** Repository settings have "Read only" permissions

**Fix:** Change to "Read and write permissions"

### Error: "GitHub Actions is not permitted to create pull requests"

**Cause:** "Allow GitHub Actions to create and approve pull requests" is unchecked

**Fix:** Either:
- Option A: Check the box (if you use PR creation in workflows)
- Option B: Switch to direct commits (already done in our workflow)

### Error: "403 Forbidden" when pushing to branch

**Cause:** Either:
- Repository permissions are read-only
- Branch protection blocks the push
- Token doesn't have required scope

**Fix:**
1. Check repository workflow permissions
2. Check branch protection rules
3. Add workflow to bypass list if needed

### Release created but workflow shows error

**Cause:** Job-level permissions too restrictive

**Fix:** Add required permissions to the job:
```yaml
permissions:
  contents: write
```

## Testing Permissions

After changing settings, test with a simple workflow:

```yaml
name: Test Permissions
on:
  workflow_dispatch:

jobs:
  test:
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - uses: actions/checkout@v4
      
      - name: Test write permission
        run: |
          echo "test" > test.txt
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git add test.txt
          git commit -m "Test commit [skip ci]"
          git push
          
      - name: Cleanup
        run: |
          git rm test.txt
          git commit -m "Cleanup test [skip ci]"
          git push
```

Run this manually from the Actions tab. If it succeeds, permissions are correct.

## Quick Checklist

Before running release workflows:

- [ ] Repository Settings → Actions → General → Workflow permissions = "Read and write"
- [ ] Job has `contents: write` permission in workflow file
- [ ] Branch protection doesn't block workflow pushes (or workflow is in bypass list)
- [ ] Token being used is `${{ secrets.GITHUB_TOKEN }}` (not a custom PAT)
- [ ] Repository is not archived or read-only

## Additional Resources

- [GitHub Actions Permissions](https://docs.github.com/en/actions/security-guides/automatic-token-authentication#permissions-for-the-github_token)
- [Workflow Permissions Syntax](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#permissions)
- [Managing GitHub Actions Settings](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/enabling-features-for-your-repository/managing-github-actions-settings-for-a-repository)
- [Release Creation API](https://docs.github.com/en/rest/releases/releases#create-a-release)

## Summary

**Immediate Fix:**
1. Go to https://github.com/GDXbsv/traktor/settings/actions
2. Select "Read and write permissions"
3. Click Save
4. Re-run failed workflow

This will allow the release workflow to create releases and upload assets successfully! 🚀