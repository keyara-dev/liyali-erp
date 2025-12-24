# CI/CD Pipeline Guide - Liyali Gateway

**Status:** Complete
**Version:** 1.0
**Date:** December 23, 2025

## Overview

The Liyali Gateway project implements a comprehensive CI/CD pipeline using GitHub Actions. The pipeline automates testing, building, quality checks, Docker image creation, and deployment.

## Pipeline Architecture

### Workflows Overview

```
┌─────────────┐
│  Push Code  │
└──────┬──────┘
       │
       ├─────────────────┬──────────────┬──────────────┬──────────────┐
       │                 │              │              │              │
       v                 v              v              v              v
    ┌─────┐          ┌──────┐      ┌────────┐     ┌────────┐     ┌──────────┐
    │ CI  │          │Docker│      │Quality │     │Release │     │Scheduled │
    │Test │          │Build │      │Check   │     │Build   │     │Tasks     │
    └─────┘          └──────┘      └────────┘     └────────┘     └──────────┘
       │                 │              │              │              │
       └─────────────────┴──────────────┴──────────────┴──────────────┘
       │
       v
    ┌────────────┐
    │ Deploy     │
    │(Production)│
    └────────────┘
```

## Workflows

### 1. CI - Build & Test (`.github/workflows/ci.yml`)

**Triggers:**
- Push to `main`, `develop`, or `feat/**` branches
- Pull requests to `main` or `develop`

**Jobs:**

#### Test Job
```yaml
- Downloads Go modules with caching
- Runs full test suite
- Generates coverage reports
- Uploads to CodeCov
- Runs performance benchmarks
```

**Environment Variables:**
```
DB_HOST: localhost
DB_PORT: 5432
DB_USER: postgres
DB_PASSWORD: postgres
DB_NAME: liyali-test-db
REDIS_HOST: localhost
REDIS_PORT: 6379
JWT_SECRET: test-secret-key
ENVIRONMENT: test
```

**Services:**
- PostgreSQL 15 (port 5432)
- Redis 7 (port 6379)

**Outputs:**
- Test results
- Coverage reports (CodeCov integration)
- Benchmark results

#### Lint Job
```yaml
- golangci-lint analysis
- Go fmt check
- Go vet analysis
```

#### Security Job
```yaml
- Gosec security scanning
- SARIF report generation
```

#### Build Job (depends on Test & Lint)
```yaml
- Builds application binary
- Verifies binary generation
- Tests binary version output
```

### 2. Docker - Build & Push (`.github/workflows/docker.yml`)

**Triggers:**
- Push to `main` branch
- Tag push (`v*`)
- Pull requests to `main`

**Jobs:**

#### Docker Build & Push
```yaml
- Login to container registry (GHCR)
- Extract metadata from tags/branches
- Build multi-platform Docker image
- Push to GHCR if not PR
- Test Docker image build
- Run container health checks
```

**Image Naming:**
```
ghcr.io/owner/repo:branch-name
ghcr.io/owner/repo:sha-commit-hash
```

**Tags:**
```
main        → latest build from main branch
v1.0.0      → semantic version tag
sha-abc123  → commit hash
```

### 3. Deploy - Production (`.github/workflows/deploy.yml`)

**Triggers:**
- Push to `v*` tags (automatic)
- Manual workflow dispatch

**Jobs:**

#### Deploy Job
```yaml
- Validates deployment environment
- Sets up SSH connection
- Pulls latest code
- Builds Docker image
- Stops old containers
- Starts new containers
- Runs database migrations
- Health check verification
- Creates deployment records
```

**Deployment Flow:**
1. Determine environment (staging/production)
2. Setup SSH authentication
3. Connect to deployment host
4. Pull latest code from repository
5. Build Docker image
6. Stop running services
7. Start new services
8. Run migrations
9. Verify health checks
10. Record deployment in GitHub

**Required Secrets:**
- `DEPLOY_KEY`: SSH private key
- `DEPLOY_HOST`: Target server hostname

### 4. Code Quality (`.github/workflows/quality.yml`)

**Triggers:**
- Push to `main` or `develop`
- Pull requests to `main` or `develop`

**Jobs:**

#### Quality Analysis
```yaml
- Go Report Card analysis
- TODO/FIXME comment detection
- Code duplication checks (dupl)
- Forbidden imports detection
- Cyclomatic complexity check (gocyclo)
- SonarQube scanning
- Coverage badge generation
- PR coverage comments
```

**Checks:**

| Check | Tool | Threshold |
|-------|------|-----------|
| Duplication | dupl | 100 lines |
| Complexity | gocyclo | 15 |
| Format | gofmt | strict |
| Vet | go vet | strict |
| Lint | golangci-lint | default |

### 5. Release (`.github/workflows/release.yml`)

**Triggers:**
- Push to version tags (`v*`)

**Jobs:**

#### Build Release
```yaml
- Build for Linux (amd64, arm64)
- Build for Darwin/macOS (amd64, arm64)
- Build for Windows (amd64)
- Generate SHA256 checksums
- Upload artifacts
```

**Build Matrix:**
```
OS       | Architectures
---------|------------------
Linux    | amd64, arm64
Darwin   | amd64, arm64
Windows  | amd64
```

#### Create Release
```yaml
- Download all build artifacts
- Generate release notes from commits
- Create GitHub Release
- Publish Docker image to GHCR
- Update CHANGELOG.md
- Commit changelog update
```

**Release Artifacts:**
- Linux binary (amd64, arm64)
- macOS binary (amd64, arm64)
- Windows binary (amd64)
- SHA256 checksums for each

### 6. Scheduled Tasks (`.github/workflows/scheduled.yml`)

**Triggers:**
- Daily at 2 AM UTC (dependency check)
- Weekly Monday at 2 AM UTC (security scan)

**Jobs:**

#### Dependency Check (Daily)
```yaml
- Check for Go updates
- Run go mod tidy
- Find outdated dependencies
- Run vulnerability scan (govulncheck)
- Create GitHub issue if updates needed
```

#### Test Coverage (Daily)
```yaml
- Run full test coverage
- Calculate coverage percentage
- Post coverage comment to repo
```

#### Service Health Check (Daily)
```yaml
- Validate docker-compose configuration
- Validate OpenAPI specification
- Check documentation links
```

#### Performance Baseline (Daily)
```yaml
- Run benchmarks
- Compare with previous baseline
- Upload benchmark results
```

## Usage Guide

### Running Workflows Locally

#### Test Locally
```bash
cd backend
make test
make test-coverage
make bench
```

#### Build Locally
```bash
cd backend
make build
```

#### Docker Build Locally
```bash
docker build -t liyali-gateway:local .
docker-compose up -d
```

### Manual Workflow Dispatch

Trigger deployment manually:
```bash
gh workflow run deploy.yml \
  -f environment=staging \
  --repo owner/repo
```

View workflow runs:
```bash
gh workflow list
gh run list --workflow ci.yml
```

### Setting Up Secrets

Required GitHub Secrets:

```bash
# Deployment
gh secret set DEPLOY_KEY --body "$(cat ~/.ssh/deploy_key)"
gh secret set DEPLOY_HOST --body "production.example.com"

# Code Quality
gh secret set SONAR_HOST_URL --body "https://sonarqube.example.com"
gh secret set SONAR_LOGIN --body "your-sonar-token"
```

### Monitoring Pipeline Status

#### GitHub UI
- Actions tab: View all workflow runs
- Branch protection: Require status checks to pass
- Deployments: Track deployment history

#### Command Line
```bash
# Watch workflow runs
gh run list --workflow ci.yml --watch

# View logs for specific run
gh run view <run-id> --log

# Cancel a workflow
gh run cancel <run-id>
```

#### Status Badge
Add to README.md:
```markdown
[![CI](https://github.com/owner/repo/actions/workflows/ci.yml/badge.svg)](https://github.com/owner/repo/actions)
[![Docker](https://github.com/owner/repo/actions/workflows/docker.yml/badge.svg)](https://github.com/owner/repo/actions)
[![Quality](https://github.com/owner/repo/actions/workflows/quality.yml/badge.svg)](https://github.com/owner/repo/actions)
```

## Branching Strategy

### Branch Rules

| Branch | Purpose | Workflows |
|--------|---------|-----------|
| main | Production | All |
| develop | Development | CI, Docker, Quality |
| feat/* | Features | CI, Quality |
| hotfix/* | Hotfixes | CI, Quality |
| release/* | Releases | CI, Docker, Quality |

### Protected Branches

Main branch protection rules:
- Require PR reviews (1+ approval)
- Require status checks to pass
  - ci/test
  - ci/lint
  - ci/security
  - ci/build
  - code-quality
- Require branches to be up to date

## Release Process

### Semantic Versioning

Version format: `v<MAJOR>.<MINOR>.<PATCH>[-<PRERELEASE>]`

Examples:
```
v1.0.0      # Major release
v1.1.0      # Minor release
v1.0.1      # Patch release
v1.0.0-rc1  # Release candidate
v1.0.0-beta # Beta version
```

### Release Steps

1. **Create Release Branch**
   ```bash
   git checkout -b release/v1.1.0 develop
   ```

2. **Update Version Numbers**
   ```bash
   # Update version in code/config files
   # Update CHANGELOG.md
   ```

3. **Create Release Commit**
   ```bash
   git commit -m "chore: Bump version to v1.1.0"
   git push origin release/v1.1.0
   ```

4. **Create Pull Request**
   ```bash
   gh pr create \
     --base main \
     --title "Release v1.1.0" \
     --body "Release notes here"
   ```

5. **Merge to Main**
   ```bash
   gh pr merge --squash --delete-branch
   ```

6. **Create Tag**
   ```bash
   git tag -a v1.1.0 -m "Release version 1.1.0"
   git push origin v1.1.0
   ```

7. **Merge Back to Develop**
   ```bash
   git checkout develop
   git merge main
   git push origin develop
   ```

### Automated on Tag

Once tag is pushed:
- Build for multiple platforms
- Create GitHub Release
- Push Docker image to GHCR
- Create deployment in GitHub
- Trigger deployment workflow

## Troubleshooting

### Common Issues

#### Test Failures

**Check logs:**
```bash
gh run view <run-id> --log
```

**Run locally:**
```bash
cd backend
make test
```

**Common causes:**
- Database not running
- Redis not running
- Missing environment variables

#### Build Failures

**Check Docker build:**
```bash
docker build -t test .
```

**Common causes:**
- Missing dependencies
- Invalid Go code
- Missing build files

#### Deployment Failures

**Check secrets:**
```bash
gh secret list
```

**Verify SSH access:**
```bash
ssh -i ~/.ssh/deploy_key deploy@host
```

**Common causes:**
- Invalid deploy key
- Incorrect deploy host
- Insufficient permissions

### Debugging Workflows

Enable debug logging:
```bash
gh secret set ACTIONS_STEP_DEBUG --body "true"
```

### Viewing Artifacts

Download workflow artifacts:
```bash
gh run download <run-id>
```

## Best Practices

### Commit Messages
Follow conventional commits:
```
feat: Add new feature
fix: Fix bug
docs: Update documentation
test: Add tests
chore: Maintenance
perf: Performance improvements
ci: CI/CD updates
```

### Pull Requests
- Write clear PR descriptions
- Include testing steps
- Reference related issues
- Request reviews from maintainers

### Deployment
- Always test in staging first
- Use semantic versioning for releases
- Maintain deployment history
- Document breaking changes

## Configuration Files

### GitHub Actions Permissions

Add to repository settings:
```yaml
permissions:
  contents: read
  packages: write
  deployments: write
  statuses: write
```

### Default Environment Variables

Set at repository level:
```bash
LOG_LEVEL=info
ENVIRONMENT=staging
API_VERSION=v1
```

## Future Enhancements

1. **Automated Dependency Updates**
   - Dependabot integration
   - Auto-merge for patch updates

2. **Performance Tracking**
   - Historical benchmark data
   - Performance regression detection

3. **Deployment Tracking**
   - Slack notifications
   - Dashboard integration

4. **Security Scanning**
   - OWASP scanning
   - Dependency vulnerability tracking
   - Container scanning

## Support & References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Go Testing Guide](https://golang.org/doc/effective_go#testing)
- [Docker Documentation](https://docs.docker.com/)
- [Semantic Versioning](https://semver.org/)

---

**Last Updated:** December 23, 2025
**Status:** Complete
**Branch:** feat/go-fiber
