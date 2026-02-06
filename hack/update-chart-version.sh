#!/usr/bin/env bash

# Copyright 2024 GDX Cloud.
# Licensed under the Apache License, Version 2.0.

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
CHART_FILE="${REPO_ROOT}/charts/traktor/Chart.yaml"
VALUES_FILE="${REPO_ROOT}/charts/traktor/values.yaml"

# Function to print colored messages
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to validate version format
validate_version() {
    local version=$1
    if [[ ! $version =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)?$ ]]; then
        print_error "Invalid version format: $version"
        print_error "Expected format: X.Y.Z or X.Y.Z-suffix (e.g., 1.0.0, 1.0.0-alpha)"
        return 1
    fi
    return 0
}

# Function to update Chart.yaml
update_chart_yaml() {
    local version=$1
    
    print_info "Updating Chart.yaml..."
    
    # Update version
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        sed -i '' "s/^version:.*/version: ${version}/" "${CHART_FILE}"
        sed -i '' "s/^appVersion:.*/appVersion: \"${version}\"/" "${CHART_FILE}"
        sed -i '' "s|url: https://github.com/GDXbsv/traktor/releases/tag/v[0-9]*\.[0-9]*\.[0-9]*|url: https://github.com/GDXbsv/traktor/releases/tag/v${version}|g" "${CHART_FILE}"
    else
        # Linux
        sed -i "s/^version:.*/version: ${version}/" "${CHART_FILE}"
        sed -i "s/^appVersion:.*/appVersion: \"${version}\"/" "${CHART_FILE}"
        sed -i "s|url: https://github.com/GDXbsv/traktor/releases/tag/v[0-9]*\.[0-9]*\.[0-9]*|url: https://github.com/GDXbsv/traktor/releases/tag/v${version}|g" "${CHART_FILE}"
    fi
    
    print_info "✓ Chart.yaml updated to version ${version}"
}

# Function to update values.yaml image tag
update_values_yaml() {
    local version=$1
    
    print_info "Updating values.yaml image tag..."
    
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        sed -i '' "s/tag: \".*\"/tag: \"${version}\"/" "${VALUES_FILE}"
    else
        # Linux
        sed -i "s/tag: \".*\"/tag: \"${version}\"/" "${VALUES_FILE}"
    fi
    
    print_info "✓ values.yaml image tag updated to ${version}"
}

# Function to display current version
show_current_version() {
    if [[ ! -f "${CHART_FILE}" ]]; then
        print_error "Chart.yaml not found at ${CHART_FILE}"
        exit 1
    fi
    
    local current_version=$(grep "^version:" "${CHART_FILE}" | awk '{print $2}')
    local current_app_version=$(grep "^appVersion:" "${CHART_FILE}" | awk '{print $2}' | tr -d '"')
    
    echo ""
    echo "Current Chart Version: ${current_version}"
    echo "Current App Version:   ${current_app_version}"
    echo ""
}

# Function to show usage
usage() {
    cat << EOF
Usage: $0 [OPTIONS] VERSION

Update Helm chart version in Chart.yaml and values.yaml

OPTIONS:
    -h, --help              Show this help message
    -c, --current           Show current version and exit
    -d, --dry-run           Show what would be changed without applying
    -s, --skip-values       Skip updating values.yaml image tag

ARGUMENTS:
    VERSION                 Version number in format X.Y.Z or X.Y.Z-suffix

EXAMPLES:
    # Update to version 1.0.0
    $0 1.0.0

    # Update to pre-release version
    $0 1.0.0-alpha

    # Show current version
    $0 --current

    # Dry run to see what would change
    $0 --dry-run 1.0.0

    # Update only Chart.yaml, skip values.yaml
    $0 --skip-values 1.0.0

EOF
}

# Main script
main() {
    local version=""
    local dry_run=false
    local skip_values=false
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                usage
                exit 0
                ;;
            -c|--current)
                show_current_version
                exit 0
                ;;
            -d|--dry-run)
                dry_run=true
                shift
                ;;
            -s|--skip-values)
                skip_values=true
                shift
                ;;
            -*)
                print_error "Unknown option: $1"
                usage
                exit 1
                ;;
            *)
                version=$1
                shift
                ;;
        esac
    done
    
    # Check if version is provided
    if [[ -z "$version" ]]; then
        print_error "Version number is required"
        echo ""
        usage
        exit 1
    fi
    
    # Validate version format
    if ! validate_version "$version"; then
        exit 1
    fi
    
    # Check if files exist
    if [[ ! -f "${CHART_FILE}" ]]; then
        print_error "Chart.yaml not found at ${CHART_FILE}"
        exit 1
    fi
    
    if [[ ! -f "${VALUES_FILE}" ]] && [[ "$skip_values" == false ]]; then
        print_warn "values.yaml not found at ${VALUES_FILE}"
        print_warn "Will skip updating values.yaml"
        skip_values=true
    fi
    
    # Show current version
    show_current_version
    
    if [[ "$dry_run" == true ]]; then
        print_info "DRY RUN MODE - No changes will be applied"
        print_info "Would update Chart.yaml version to: ${version}"
        print_info "Would update Chart.yaml appVersion to: ${version}"
        print_info "Would update changelog URL to: https://github.com/GDXbsv/traktor/releases/tag/v${version}"
        if [[ "$skip_values" == false ]]; then
            print_info "Would update values.yaml image tag to: ${version}"
        fi
        exit 0
    fi
    
    # Confirm before proceeding
    echo -e "${YELLOW}About to update chart version to: ${version}${NC}"
    read -p "Continue? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Aborted by user"
        exit 0
    fi
    
    # Update files
    update_chart_yaml "$version"
    
    if [[ "$skip_values" == false ]]; then
        update_values_yaml "$version"
    fi
    
    echo ""
    print_info "✅ Version update complete!"
    print_info ""
    print_info "Next steps:"
    print_info "1. Review the changes: git diff"
    print_info "2. Commit the changes: git add charts/ && git commit -m 'chore: bump chart version to ${version}'"
    print_info "3. Create a git tag: git tag v${version}"
    print_info "4. Push changes: git push origin main --tags"
    print_info ""
    print_info "The CI/CD pipeline will automatically:"
    print_info "  - Build and push Docker images"
    print_info "  - Package and publish Helm chart"
    print_info "  - Create GitHub release"
    echo ""
}

main "$@"