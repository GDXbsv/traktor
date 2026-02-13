#!/bin/bash
set -e

# Script to clean up old Helm charts with invalid Artifact Hub metadata
# This removes chart versions that have the invalid signKey annotation

echo "=========================================="
echo "Helm Repository Cleanup Script"
echo "=========================================="
echo ""

# Check if we're in the right repository
if [ ! -d ".git" ]; then
  echo "Error: Not in a git repository. Please run from repository root."
  exit 1
fi

# Get current branch
CURRENT_BRANCH=$(git branch --show-current)
echo "Current branch: $CURRENT_BRANCH"

# Fetch latest changes
echo "Fetching latest changes..."
git fetch origin

# Check if gh-pages branch exists
if ! git ls-remote --heads origin gh-pages | grep -q gh-pages; then
  echo "Error: gh-pages branch does not exist"
  exit 1
fi

# Checkout gh-pages branch
echo ""
echo "Checking out gh-pages branch..."
git checkout gh-pages
git pull origin gh-pages

# List current charts
echo ""
echo "Current charts in repository:"
ls -lh *.tgz 2>/dev/null || echo "No chart packages found"

# Remove old chart versions with invalid metadata
echo ""
echo "Removing old chart versions with invalid metadata..."
CHARTS_TO_REMOVE=(
  "traktor-0.0.1.tgz"
  "traktor-0.0.2.tgz"
  "traktor-0.0.4.tgz"
  "traktor-0.0.5.tgz"
)

REMOVED_COUNT=0
for chart in "${CHARTS_TO_REMOVE[@]}"; do
  if [ -f "$chart" ]; then
    echo "  Removing: $chart"
    rm -f "$chart"
    REMOVED_COUNT=$((REMOVED_COUNT + 1))
  else
    echo "  Skipping: $chart (not found)"
  fi
done

echo ""
echo "Removed $REMOVED_COUNT chart(s)"

# Check if helm is installed
if ! command -v helm &> /dev/null; then
  echo ""
  echo "Error: helm command not found. Please install Helm:"
  echo "  https://helm.sh/docs/intro/install/"
  exit 1
fi

# Regenerate index
echo ""
echo "Regenerating Helm repository index..."
helm repo index . --url https://gdxbsv.github.io/traktor

# Ensure artifacthub-repo.yml exists with correct content
echo ""
echo "Updating artifacthub-repo.yml..."
cat > artifacthub-repo.yml << 'EOF'
# Artifact Hub repository metadata file
# https://artifacthub.io/docs/topics/repositories/

repositoryID: 2a1be652-53bf-42c6-a2b5-7b87f26bc585
displayName: Traktor Operator
owners:
  - name: GDX Cloud
    email: support@gdxcloud.net
EOF

# Show status
echo ""
echo "Current status:"
git status

# Commit changes
echo ""
read -p "Commit and push changes? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
  echo "Committing changes..."
  git add index.yaml artifacthub-repo.yml
  git add -u *.tgz 2>/dev/null || true
  git commit -m "chore: remove old chart versions with invalid Artifact Hub metadata

- Removed charts with invalid signKey annotation
- Regenerated index.yaml
- Updated artifacthub-repo.yml
"
  
  echo ""
  echo "Pushing to origin/gh-pages..."
  git push origin gh-pages
  
  echo ""
  echo "✅ Successfully cleaned up Helm repository!"
else
  echo "Changes not committed. You can review and commit manually."
fi

# Return to original branch
echo ""
echo "Returning to $CURRENT_BRANCH branch..."
git checkout "$CURRENT_BRANCH"

echo ""
echo "=========================================="
echo "Cleanup complete!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "1. Wait for Artifact Hub to re-index (usually within 30 minutes)"
echo "2. Check your repository: https://artifacthub.io/packages/search?repo=traktor"
echo "3. Verify no errors in Artifact Hub control panel"
echo ""
echo "Repository URL: https://gdxbsv.github.io/traktor"
echo "Metadata URL: https://gdxbsv.github.io/traktor/artifacthub-repo.yml"
echo "Index URL: https://gdxbsv.github.io/traktor/index.yaml"
echo ""