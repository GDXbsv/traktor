# Repository-Level Permissions Setup Guide

This guide provides exact steps to configure GitHub repository permissions so that workflows can create releases.

## The Problem

If you see this error in your workflow:
```
⚠️ GitHub release failed with status: 403
{"message":"Resource not accessible by integration"}
Error: Resource not accessible by integration
```

This means GitHub Actions doesn't have permission to create releases at the **repository level**.

## Solution: Step-by-Step Instructions

### Step 1: Navigate to Repository Settings

1. Open your browser and go to:
   ```
   https://github.com/GDXbsv/traktor
   ```

2. Click the **"Settings"** tab (top navigation bar, far right)
   
   ```
   < > Code    Issues    Pull requests    Actions    Projects    Wiki    Security    Insights    [Settings]
                                                                                                      ↑ Click here
   ```

   **Note:** You must be a repository owner or admin to see this tab.

### Step 2: Navigate to Actions Settings

In the left sidebar, scroll down to find the **"Actions"** section:

```
Settings
├── General
├── Access
│   ├── Collaborators
│   └── Moderation options
├── Code and automation
│   ├── Branches
│   ├── Tags
│   ├── Rules
│   ├── Actions              ← You want this section
│   │   ├── General          ← Click this
│   │   └── Runners
│   ├── Webhooks
│   └── Environments
└── ...
```

Click on: **Actions** → **General**

Full URL: `https://github.com/GDXbsv/traktor/settings/actions`

### Step 3: Find Workflow Permissions Section

Scroll down on the Actions General page until you see **"Workflow permissions"**

It's usually near the bottom of the page, after:
- Actions permissions
- Artifact and log retention
- Fork pull request workflows

### Step 4: Configure Permissions

You'll see two radio button options:

```
Workflow permissions

Choose the default permissions granted to the GITHUB_TOKEN when running workflows in this repository.
You can specify more granular permissions in the workflow using the permissions key.
Learn more about managing permissions

  ○ Read repository contents and packages permissions
    Workflows have read access to the repository and packages.

  ○ Read and write permissions
    Workflows have read and write permissions in the repository for all scopes.

  ☐ Allow GitHub Actions to create and approve pull requests
    This controls whether GitHub Actions can create pull requests or submit approving pull request reviews.
```

### Step 5: Select "Read and Write Permissions"

**Select the SECOND option:**

```
  ⦿ Read and write permissions
    Workflows have read and write permissions in the repository for all scopes.
```

**Optional:** Check the box below (only if you want workflows to create PRs):

```
  ☑ Allow GitHub Actions to create and approve pull requests
```

For Traktor, you can **leave this UNCHECKED** since we use direct commits instead of PRs.

### Step 6: Save Changes

1. Scroll to the bottom of the page
2. Click the **"Save"** button

```
[Save]
```

You should see a confirmation message: "Workflow permissions updated"

## What This Changes

### Before (Default - Restrictive)
```
Read repository contents and packages permissions
```

**Workflows CAN:**
- ✅ Read code
- ✅ Read packages
- ✅ Run tests

**Workflows CANNOT:**
- ❌ Create releases
- ❌ Push commits
- ❌ Upload release assets
- ❌ Update tags
- ❌ Modify repository

### After (Recommended - Permissive)
```
Read and write permissions
```

**Workflows CAN:**
- ✅ Read code
- ✅ Read packages
- ✅ Run tests
- ✅ Create releases ← **This is what we need**
- ✅ Push commits
- ✅ Upload release assets
- ✅ Update tags
- ✅ Modify repository content

## Verify the Change

After saving:

1. Go back to the same page:
   ```
   https://github.com/GDXbsv/traktor/settings/actions
   ```

2. Scroll to "Workflow permissions"

3. Verify it shows:
   ```
   ⦿ Read and write permissions
   ```

## Re-run Failed Workflow

### Option 1: Re-run Existing Workflow

1. Go to Actions tab:
   ```
   https://github.com/GDXbsv/traktor/actions
   ```

2. Find the failed workflow run (e.g., "Release" for v0.0.11)

3. Click on the workflow run

4. Click **"Re-run failed jobs"** or **"Re-run all jobs"** button (top right)

### Option 2: Create New Release Tag

```bash
cd /home/gdx/GDX_files/progects/go/traktor

# Create and push new tag
git tag v0.0.12
git push origin v0.0.12
```

This will trigger a fresh workflow run with the new permissions.

## Visual Reference

Here's what you're looking for:

```
┌─────────────────────────────────────────────────────────────┐
│ Workflow permissions                                        │
│                                                             │
│ Choose the default permissions granted to the              │
│ GITHUB_TOKEN when running workflows in this repository.    │
│                                                             │
│ ○ Read repository contents and packages permissions        │
│   Workflows have read access to the repository and         │
│   packages.                                                 │
│                                                             │
│ ⦿ Read and write permissions                               │  ← Select this
│   Workflows have read and write permissions in the         │
│   repository for all scopes.                               │
│                                                             │
│ ☐ Allow GitHub Actions to create and approve pull requests │  ← Optional
│   This controls whether GitHub Actions can create pull     │
│   requests or submit approving pull request reviews.       │
│                                                             │
│                                                  [Save]     │  ← Click Save
└─────────────────────────────────────────────────────────────┘
```

## Troubleshooting

### I don't see the "Settings" tab

**Cause:** You don't have admin access to the repository.

**Solution:** Contact the repository owner to:
- Grant you admin access, OR
- Have them configure the permissions

### I saved but it still fails with 403

**Possible causes:**

1. **Changes not applied yet:**
   - Wait 1-2 minutes after saving
   - Hard refresh the settings page (Ctrl+F5 or Cmd+Shift+R)
   - Verify the radio button is still selected

2. **Organization-level restrictions:**
   - If the repo is in an organization, check organization settings
   - Go to: `https://github.com/organizations/GDXbsv/settings/actions`
   - Check if organization allows "Read and write permissions"

3. **Branch protection:**
   - Check if branch protection blocks workflow pushes
   - Go to: Settings → Branches
   - Review protection rules for `main` branch

### The setting keeps reverting

**Cause:** Organization policy might be overriding repository settings.

**Solution:** Check organization-level Actions settings:
```
https://github.com/organizations/GDXbsv/settings/actions
```

Under "Workflow permissions", ensure it's NOT set to:
```
Read repository contents permission (MOST RESTRICTIVE)
```

### I'm in an organization and can't change this

**Cause:** Organization admins control this setting.

**Solution:** Contact your organization admin to:
1. Enable "Read and write permissions" at org level, OR
2. Allow repositories to override the org setting

## Security Considerations

### Is "Read and write permissions" safe?

**Yes, if you follow best practices:**

✅ **Safe when:**
- You review all workflow changes before merging
- You use branch protection on `main`
- You trust repository collaborators
- Workflows only run on trusted branches/tags
- You use `pull_request_target` carefully

⚠️ **Risk factors:**
- Malicious contributors could modify workflows
- Compromised accounts could push harmful workflows
- Pull requests from forks could be risky (use `pull_request_target` with caution)

### Mitigation Strategies

1. **Require code review for `.github/workflows/` changes:**
   - Use CODEOWNERS file:
     ```
     .github/workflows/* @GDXbsv/admins
     ```

2. **Enable branch protection:**
   - Require PR reviews before merging
   - Require status checks to pass
   - Include administrators in restrictions

3. **Use job-level permissions:**
   Even with write access, limit individual jobs:
   ```yaml
   jobs:
     my-job:
       permissions:
         contents: read  # Only read for this job
   ```

4. **Monitor workflow runs:**
   - Check Actions tab regularly
   - Review unexpected workflow runs
   - Check audit log for changes

## Organization Settings (If Applicable)

If your repository is in an organization (`GDXbsv`), org admins should check:

### Navigate to Organization Settings

1. Go to: `https://github.com/GDXbsv`
2. Click **"Settings"** tab
3. Click **"Actions"** → **"General"** in left sidebar

### Configure Organization Policy

Find "Workflow permissions" section:

**Option 1: Most Flexible (Recommended)**
```
⦿ Read and write permissions
  Allow repositories to override this setting
```

**Option 2: Let Each Repo Decide**
```
⦿ Read repository contents permission
☑ Allow repositories to choose their own permission level
```

**Option 3: Most Restrictive (Blocks releases)**
```
⦿ Read repository contents permission
  Apply to all repositories (no override)
```

If your org uses Option 3, you'll need to convince admins to change to Option 1 or 2.

## Quick Checklist

Before running release workflows, verify:

- [ ] Repository Settings → Actions → General page opened
- [ ] "Workflow permissions" section found
- [ ] "Read and write permissions" radio button selected
- [ ] Changes saved (green confirmation banner appeared)
- [ ] Settings page refreshed to confirm change persisted
- [ ] If in organization: org policy allows repository-level permissions
- [ ] Workflow re-run or new tag created to test

## Expected Result

After configuration, the release workflow should:

✅ Complete without 403 errors
✅ Create GitHub release
✅ Upload all release assets
✅ Update documentation
✅ Publish Helm chart

## Still Having Issues?

If you've followed all steps and still get 403 errors:

1. **Verify your role:**
   ```
   Repository Settings → Collaborators → Your username should show "Admin"
   ```

2. **Check organization restrictions:**
   - Contact organization owner
   - Ask them to review org-level Actions policies

3. **Try using a Personal Access Token (PAT):**
   - As a last resort, use a PAT with `repo` scope
   - See: `docs/GITHUB_PAT_SETUP.md`

4. **Check GitHub status:**
   - Visit: https://www.githubstatus.com/
   - Ensure Actions service is operational

## Additional Resources

- [GitHub Actions Permissions](https://docs.github.com/en/actions/security-guides/automatic-token-authentication)
- [Managing GitHub Actions Settings](https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/enabling-features-for-your-repository/managing-github-actions-settings-for-a-repository)
- [Organization Workflow Permissions](https://docs.github.com/en/organizations/managing-organization-settings/disabling-or-limiting-github-actions-for-your-organization#setting-the-permissions-of-the-github_token-for-your-organization)

## Summary

**To fix the 403 error:**

1. Go to: `https://github.com/GDXbsv/traktor/settings/actions`
2. Scroll to: "Workflow permissions"
3. Select: "Read and write permissions"
4. Click: "Save"
5. Re-run: Failed workflow or create new tag

**This gives workflows the permission to create releases!** 🚀