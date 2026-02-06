# Hack Directory

This directory contains helper scripts and development utilities for the Traktor project.

## Scripts

### update-chart-version.sh

Automatically updates Helm chart version numbers in `Chart.yaml` and `values.yaml`.

**Note**: This script is rarely needed because the CI/CD release pipeline automatically updates versions when you push a git tag. Use this only for manual testing or local development.

#### Usage

```bash
# Update to version 1.2.3
./hack/update-chart-version.sh 1.2.3

# Update to prerelease version
./hack/update-chart-version.sh 1.2.3-alpha.1

# Show current version
./hack/update-chart-version.sh --current

# Dry run (preview changes without applying)
./hack/update-chart-version.sh --dry-run 1.2.3

# Update only Chart.yaml (skip values.yaml)
./hack/update-chart-version.sh --skip-values 1.2.3

# Show help
./hack/update-chart-version.sh --help
```

#### What it updates

1. **charts/traktor/Chart.yaml**:
   - `version: X.Y.Z`
   - `appVersion: "X.Y.Z"`
   - Release URL in `artifacthub.io/changes` annotation

2. **charts/traktor/values.yaml** (unless `--skip-values` is used):
   - `image.tag: "X.Y.Z"`

#### Examples

**Show current version:**
```bash
$ ./hack/update-chart-version.sh --current

Current Chart Version: 0.0.1
Current App Version:   0.0.1
```

**Preview changes without applying:**
```bash
$ ./hack/update-chart-version.sh --dry-run 1.2.3

Current Chart Version: 0.0.1
Current App Version:   0.0.1

[INFO] DRY RUN MODE - No changes will be applied
[INFO] Would update Chart.yaml version to: 1.2.3
[INFO] Would update Chart.yaml appVersion to: 1.2.3
[INFO] Would update changelog URL to: https://github.com/GDXbsv/traktor/releases/tag/v1.2.3
[INFO] Would update values.yaml image tag to: 1.2.3
```

**Update to version 1.2.3:**
```bash
$ ./hack/update-chart-version.sh 1.2.3

Current Chart Version: 0.0.1
Current App Version:   0.0.1

About to update chart version to: 1.2.3
Continue? (y/N) y
[INFO] Updating Chart.yaml...
[INFO] ✓ Chart.yaml updated to version 1.2.3
[INFO] Updating values.yaml image tag...
[INFO] ✓ values.yaml image tag updated to 1.2.3

[INFO] ✅ Version update complete!
[INFO]
[INFO] Next steps:
[INFO] 1. Review the changes: git diff
[INFO] 2. Commit the changes: git add charts/ && git commit -m 'chore: bump chart version to 1.2.3'
[INFO] 3. Create a git tag: git tag v1.2.3
[INFO] 4. Push changes: git push origin main --tags
[INFO]
[INFO] The CI/CD pipeline will automatically:
[INFO]   - Build and push Docker images
[INFO]   - Package and publish Helm chart
[INFO]   - Create GitHub release
```

#### Version Format

The script validates version format according to Semantic Versioning:
- **Valid**: `1.2.3`, `1.2.3-alpha`, `1.2.3-beta.1`, `1.2.3-rc.1`
- **Invalid**: `v1.2.3` (remove 'v' prefix), `1.2`, `1`

#### Platform Support

The script works on both Linux and macOS, automatically detecting the platform and using the appropriate `sed` syntax.

## Development Utilities

### boilerplate.go.txt

Standard copyright header for generated Go files. Used by `controller-gen` when generating DeepCopy implementations and other scaffolding.

## When to Use These Scripts

| Scenario | Use Script? | Recommended Action |
|----------|-------------|-------------------|
| Creating a release | ❌ No | Just push a git tag - CI/CD handles versions |
| Testing chart locally | ✅ Yes | Use script to test with different versions |
| Manual version bump | ✅ Yes | Use script then commit changes |
| Debugging CI/CD | ✅ Yes | Use script to reproduce CI/CD version updates |
| Regular development | ❌ No | Versions are managed automatically |

## CI/CD Integration

The `update-chart-version.sh` script logic is integrated into the GitHub Actions release workflow (`.github/workflows/release.yml`). When you push a git tag:

1. The workflow extracts the version from the tag (e.g., `v1.2.3` → `1.2.3`)
2. It runs commands equivalent to this script to update Chart.yaml and values.yaml
3. It packages the updated Helm chart
4. It publishes everything to GitHub Releases and GitHub Pages

See the `package-helm-chart` job in the release workflow for implementation details.

## Contributing

When adding new helper scripts to this directory:

1. Make scripts executable: `chmod +x hack/new-script.sh`
2. Add shebang: `#!/usr/bin/env bash`
3. Include usage documentation with `--help` flag
4. Add error handling with `set -e`
5. Use colored output for better UX
6. Document the script in this README

### Script Template

```bash
#!/usr/bin/env bash

# Copyright 2024 GDX Cloud.
# Licensed under the Apache License, Version 2.0.

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Description of the script.

OPTIONS:
    -h, --help    Show this help message

EXAMPLES:
    $0 --help
EOF
}

main() {
    # Script logic here
    print_info "Running script..."
}

main "$@"
```

## Related Documentation

- [Release Guide](../RELEASE.md) - How to create releases
- [Contributing Guide](../CONTRIBUTING.md) - Development guidelines
- [CI/CD Workflows]../.github/workflows/README.md) - Workflow documentation

## Questions?

- 📖 Read the full documentation: [RELEASE.md](../RELEASE.md)
- 🐛 Found a bug? [Open an issue](https://github.com/GDXbsv/traktor/issues)
- 💬 Need help? [Start a discussion](https://github.com/GDXbsv/traktor/discussions)