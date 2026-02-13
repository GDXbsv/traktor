# GitHub Personal Access Token Setup

This guide explains how to set up a Personal Access Token (PAT) for the release workflow if branch protection rules prevent the default `GITHUB_TOKEN` from creating pull requests.

## When Do You Need a PAT?

You need a PAT if:
- Your `main` branch has branch protection rules that block GitHub Actions bot
- The `update-docs` job in the release workflow fails with permission errors
- You want pull requests created by the workflow to trigger other workflows (the default `GITHUB_TOKEN` doesn't trigger workflows for security reasons)

## Current Configuration

The workflow currently uses `${{ secrets.GITHUB_TOKEN }}` with these permissions:
```yaml
permissions:
  contents: write
  pull-requests: write
```

This works for most cases. **Only follow this guide if you're experiencing permission issues.**

## Creating a Personal Access Token

### Option 1: Fine-grained Personal Access Token (Recommended)

1. Go to GitHub Settings → **Developer settings** → **Personal access tokens** → **Fine-grained tokens**
2. Click **"Generate new token"**
3. Configure the token:
   - **Name**: `Traktor Release Automation`
   - **Expiration**: Choose appropriate expiration (90 days, 1 year, or custom)
   - **Repository access**: Select "Only select repositories" → Choose `GDXbsv/traktor`
   - **Permissions**:
     - Repository permissions:
       - **Contents**: Read and write
       - **Pull requests**: Read and write
       - **Metadata**: Read-only (automatically selected)

4. Click **"Generate token"**
5. **Copy the token immediately** (you won't see it again!)

### Option 2: Classic Personal Access Token

1. Go to GitHub Settings → **Developer settings** → **Personal access tokens** → **Tokens (classic)**
2. Click **"Generate new token (classic)"**
3. Configure the token:
   - **Note**: `Traktor Release Automation`
   - **Expiration**: Choose appropriate expiration
   - **Scopes**: Select:
     - `repo` (Full control of private repositories)
     - `workflow` (Update GitHub Action workflows)

4. Click **"Generate token"**
5. **Copy the token immediately**

## Adding the Token to GitHub Secrets

1. Go to your repository: `https://github.com/GDXbsv/traktor`
2. Navigate to **Settings** → **Secrets and variables** → **Actions**
3. Click **"New repository secret"**
4. Add the secret:
   - **Name**: `PAT_TOKEN` (or `GH_PAT`)
   - **Secret**: Paste your token
5. Click **"Add secret"**

## Updating the Workflow

Once you have added the PAT secret, update `.github/workflows/release.yml`:

### Current (using GITHUB_TOKEN):
```yaml
- name: Create PR for version update
  uses: peter-evans/create-pull-request@v6
  with:
    token: ${{ secrets.GITHUB_TOKEN }}
    # ... rest of configuration
```

### Updated (using PAT):
```yaml
- name: Create PR for version update
  uses: peter-evans/create-pull-request@v6
  with:
    token: ${{ secrets.PAT_TOKEN }}  # Changed this line
    # ... rest of configuration
```

## Security Considerations

### Token Security
- ✅ Store tokens only in GitHub Secrets (never commit them)
- ✅ Use fine-grained tokens with minimal permissions
- ✅ Set reasonable expiration dates
- ✅ Rotate tokens regularly
- ✅ Revoke tokens immediately if compromised

### Token Scope
The PAT only needs:
- **Contents write**: To push to branches
- **Pull requests write**: To create and update PRs

**Never grant more permissions than necessary!**

## Testing the Configuration

After setting up the PAT:

1. Create a new release tag:
   ```bash
   git tag v0.0.6
   git push origin v0.0.6
   ```

2. Monitor the workflow:
   - Go to **Actions** tab
   - Watch the "Release" workflow
   - Check that the `update-docs` job completes successfully

3. Verify the PR was created:
   - Go to **Pull requests** tab
   - Look for a PR titled "docs: update documentation for release v0.0.6"

## Troubleshooting

### Error: "Resource not accessible by integration"
**Solution**: The token doesn't have sufficient permissions. Re-create the token with correct scopes.

### Error: "Bad credentials"
**Solution**: The token is invalid or expired. Generate a new token.

### Error: "Reference does not exist"
**Solution**: The base branch (`main`) doesn't exist or the workflow doesn't have access. Check repository settings.

### Pull request created but workflows don't trigger
**Solution**: This is expected with `GITHUB_TOKEN`. Use a PAT if you need workflows to trigger on the PR.

### Token expired
**Solution**: 
1. Generate a new token
2. Update the secret in repository settings
3. No workflow changes needed

## Alternative: GitHub App

For organizations, consider using a GitHub App instead of PATs:

**Advantages**:
- Better security model
- Fine-grained permissions
- Better audit logs
- Tokens don't expire with user account

**Setup** (advanced):
1. Create a GitHub App
2. Install it on your repository
3. Use `actions/create-github-app-token@v1` in workflow
4. Pass the app token to `create-pull-request`

Example:
```yaml
- name: Generate token
  id: generate_token
  uses: actions/create-github-app-token@v1
  with:
    app-id: ${{ secrets.APP_ID }}
    private-key: ${{ secrets.APP_PRIVATE_KEY }}

- name: Create PR for version update
  uses: peter-evans/create-pull-request@v6
  with:
    token: ${{ steps.generate_token.outputs.token }}
```

## Additional Resources

- [GitHub PAT Documentation](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token)
- [peter-evans/create-pull-request Action](https://github.com/peter-evans/create-pull-request)
- [GitHub Actions Permissions](https://docs.github.com/en/actions/security-guides/automatic-token-authentication#permissions-for-the-github_token)
- [Using secrets in GitHub Actions](https://docs.github.com/en/actions/security-guides/encrypted-secrets)

## Summary

| Scenario | Recommended Solution |
|----------|---------------------|
| No branch protection | Use `GITHUB_TOKEN` (current setup) |
| Branch protection enabled | Use fine-grained PAT |
| Need to trigger workflows on PR | Use PAT or GitHub App |
| Organization with multiple repos | Use GitHub App |

The current workflow configuration should work in most cases. Only set up a PAT if you encounter permission errors or need additional functionality.