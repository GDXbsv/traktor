# Contributing to Traktor Operator

Thank you for your interest in contributing to Traktor! We welcome contributions from the community.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Coding Guidelines](#coding-guidelines)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Issue Guidelines](#issue-guidelines)

## Code of Conduct

This project adheres to a code of conduct. By participating, you are expected to uphold this code. Please be respectful and constructive in all interactions.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR-USERNAME/traktor.git
   cd traktor
   ```
3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/GDXbsv/traktor.git
   ```

## Development Setup

### Prerequisites

- Go 1.24.0+
- Docker 17.03+
- kubectl 1.11.3+
- Access to a Kubernetes cluster (minikube, kind, or real cluster)
- Make

### Install Dependencies

```bash
# Install Go dependencies
go mod download

# Install development tools
make setup-envtest
```

### Run Locally

```bash
# Install CRDs
make install

# Run operator locally (connects to your kubeconfig cluster)
make run
```

### Build and Test

```bash
# Run tests
make test

# Run linter
make lint

# Build binary
make build

# Build Docker image
make docker-build
```

## How to Contribute

### Types of Contributions

We welcome many types of contributions:

- **Bug Fixes** - Fix issues in the codebase
- **New Features** - Add new functionality
- **Documentation** - Improve or add documentation
- **Tests** - Add or improve test coverage
- **Examples** - Add usage examples
- **Performance** - Optimize performance
- **Refactoring** - Improve code quality

### Before You Start

1. **Check existing issues** - Someone might already be working on it
2. **Open an issue first** - For major changes, discuss your approach
3. **Keep changes focused** - One feature/fix per PR
4. **Read the docs** - Familiarize yourself with the codebase

## Coding Guidelines

### Go Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting (automatically done by `make fmt`)
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use meaningful variable names
- Write clear comments for exported functions

### Code Structure

```
traktor/
├── api/v1alpha1/           # API definitions (CRDs)
├── cmd/                    # Main application
├── config/                 # Kubernetes manifests
├── internal/controller/    # Controller logic
├── test/                   # Tests
└── docs/                   # Documentation
```

### Naming Conventions

- **Files**: `snake_case.go`
- **Functions**: `PascalCase` (exported), `camelCase` (private)
- **Variables**: `camelCase`
- **Constants**: `PascalCase` or `SCREAMING_SNAKE_CASE`
- **Types**: `PascalCase`

### Comments

```go
// Package controller implements the SecretsRefresh controller.
package controller

// SecretsRefreshReconciler reconciles a SecretsRefresh object.
type SecretsRefreshReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

// Reconcile handles the reconciliation loop.
// It restarts deployments when secrets change.
func (r *SecretsRefreshReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // Implementation
}
```

### Error Handling

```go
// Good: Return errors to caller
func (r *Reconciler) doSomething(ctx context.Context) error {
    if err := r.Client.Get(ctx, key, obj); err != nil {
        return fmt.Errorf("failed to get object: %w", err)
    }
    return nil
}

// Good: Log and handle errors appropriately
if err := r.update(ctx, obj); err != nil {
    logger.Error(err, "Failed to update object")
    return ctrl.Result{}, err
}
```

## Testing

### Writing Tests

- Write tests for all new features
- Maintain or improve code coverage
- Use table-driven tests where appropriate
- Test both success and failure cases

### Test Structure

```go
var _ = Describe("MyFeature", func() {
    Context("When something happens", func() {
        BeforeEach(func() {
            // Setup
        })

        AfterEach(func() {
            // Cleanup
        })

        It("should do the right thing", func() {
            // Test logic
            Expect(result).To(BeTrue())
        })
    })
})
```

### Running Tests

```bash
# Run all tests
make test

# Run specific test
go test ./internal/controller/... -v -run TestName

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run E2E tests
make test-e2e
```

### Test Guidelines

- **Unit tests** should be fast and isolated
- **Integration tests** can use envtest
- **E2E tests** should test complete workflows
- Use `Eventually()` for async operations
- Clean up resources in `AfterEach` blocks
- Use unique names to avoid conflicts

## Pull Request Process

### 1. Create a Branch

```bash
git checkout -b feature/amazing-feature
```

Use descriptive branch names:
- `feature/add-something` - New features
- `fix/issue-123` - Bug fixes
- `docs/improve-readme` - Documentation
- `refactor/cleanup-code` - Refactoring
- `test/add-coverage` - Tests

### 2. Make Your Changes

- Write clear, concise commit messages
- Keep commits focused and atomic
- Add tests for new functionality
- Update documentation as needed

### 3. Commit Your Changes

```bash
git add .
git commit -m "feat: add amazing feature"
```

**Commit Message Format:**

```
type(scope): subject

body (optional)

footer (optional)
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring
- `style`: Formatting changes
- `chore`: Maintenance tasks
- `ci`: CI/CD changes

**Examples:**
```
feat(controller): add namespace filtering
fix(reconcile): prevent operator self-restart
docs(readme): add quick start guide
test(controller): improve coverage for reconcile
```

### 4. Sync with Upstream

```bash
git fetch upstream
git rebase upstream/main
```

### 5. Run Tests

```bash
# Run tests
make test

# Run linter
make lint

# Verify build
make build
```

### 6. Push Your Branch

```bash
git push origin feature/amazing-feature
```

### 7. Open a Pull Request

1. Go to https://github.com/GDXbsv/traktor
2. Click "New Pull Request"
3. Select your branch
4. Fill in the PR template:

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] E2E tests added/updated
- [ ] Tested locally

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Comments added where needed
- [ ] Documentation updated
- [ ] Tests pass locally
```

### 8. Review Process

- Maintainers will review your PR
- Address feedback and comments
- Make requested changes
- Push updates to your branch
- PR will be merged once approved

## Issue Guidelines

### Reporting Bugs

Use the bug report template and include:

- **Description** - Clear description of the bug
- **Steps to Reproduce** - Detailed steps
- **Expected Behavior** - What should happen
- **Actual Behavior** - What actually happens
- **Environment** - OS, Kubernetes version, operator version
- **Logs** - Relevant logs or error messages

**Example:**

```markdown
**Bug Description**
Deployments not restarting when secret changes

**Steps to Reproduce**
1. Install operator
2. Create SecretsRefresh with namespace selector
3. Update labeled secret
4. Deployments don't restart

**Expected Behavior**
Deployments should restart automatically

**Environment**
- Kubernetes: v1.28.0
- Operator: v0.0.1
- Platform: AWS EKS

**Logs**
[Paste relevant logs here]
```

### Feature Requests

Include:

- **Problem** - What problem does this solve?
- **Proposed Solution** - How should it work?
- **Alternatives** - Other approaches considered
- **Additional Context** - Examples, use cases

### Questions

- Check documentation first
- Search existing issues
- Be specific and clear
- Provide context

## Development Tips

### Debugging

```bash
# Enable verbose logging
go run ./cmd/main.go --zap-devel

# Debug specific namespace
kubectl logs -n traktor-system deployment/traktor-controller-manager -f

# Check events
kubectl get events -n my-namespace --sort-by=.lastTimestamp
```

### Testing Locally

```bash
# Quick iteration loop
make install      # Install CRDs
make run          # Run locally
# Make changes
# Ctrl+C and run again
```

### Common Issues

**"Namespace is terminating"** in tests:
- Tests now use unique namespace names to avoid this

**RBAC errors:**
- Make sure CRDs are installed: `make install`
- Check your kubeconfig has cluster-admin access

**Build fails:**
- Run `go mod tidy`
- Check Go version: `go version`

## Getting Help

- **Documentation**: See [README.md](README.md) and [docs/](docs/)
- **Issues**: Open an [issue](../../issues)
- **Discussions**: Start a [discussion](../../discussions)
- **Slack**: Join our Slack channel (coming soon)

## Recognition

Contributors will be:
- Listed in release notes
- Added to [CONTRIBUTORS.md](CONTRIBUTORS.md)
- Mentioned in relevant documentation

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.

---

**Thank you for contributing to Traktor! 🎉**