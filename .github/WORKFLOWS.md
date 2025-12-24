# GitHub Actions Workflows

## Overview

This directory contains all GitHub Actions workflow configurations for the Liyali Gateway project.

## Workflow Files

### 1. ci.yml - Continuous Integration
- **File:** `.github/workflows/ci.yml`
- **Triggers:** Push to main/develop/feat/*, Pull requests
- **Jobs:**
  - test (unit & integration tests)
  - lint (code quality)
  - security (vulnerability scanning)
  - build (binary compilation)
- **Duration:** ~10-15 minutes

### 2. docker.yml - Docker Build & Push
- **File:** `.github/workflows/docker.yml`
- **Triggers:** Push to main, tag push, pull requests
- **Jobs:**
  - build (multi-platform Docker image)
- **Outputs:** Docker images pushed to GHCR
- **Duration:** ~5-10 minutes

### 3. deploy.yml - Deployment
- **File:** `.github/workflows/deploy.yml`
- **Triggers:** Version tags (v*), manual dispatch
- **Jobs:**
  - deploy (SSH-based deployment)
- **Environments:** Staging, Production
- **Duration:** ~5-10 minutes

### 4. quality.yml - Code Quality Analysis
- **File:** `.github/workflows/quality.yml`
- **Triggers:** Push to main/develop, pull requests
- **Jobs:**
  - quality (analysis & reporting)
- **Tools:** golangci-lint, dupl, gocyclo, SonarQube
- **Duration:** ~5-10 minutes

### 5. release.yml - Release Build & Publish
- **File:** `.github/workflows/release.yml`
- **Triggers:** Push to version tags (v*)
- **Jobs:**
  - build (cross-platform compilation)
  - release (GitHub release & distribution)
- **Outputs:**
  - Multi-platform binaries
  - GitHub Release with artifacts
  - Updated CHANGELOG
- **Duration:** ~10-15 minutes

### 6. scheduled.yml - Scheduled Tasks
- **File:** `.github/workflows/scheduled.yml`
- **Triggers:** Schedule (daily/weekly)
- **Jobs:**
  - dependency-check (daily at 2 AM UTC)
  - test-coverage (daily at 2 AM UTC)
  - health-check (daily at 2 AM UTC)
  - performance-baseline (daily at 2 AM UTC)
- **Duration:** ~10-20 minutes

## Quick Reference

### Workflow Status Badges

Add to README.md:

```markdown
[![CI](https://github.com/owner/repo/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/owner/repo/actions)
[![Docker](https://github.com/owner/repo/actions/workflows/docker.yml/badge.svg?branch=main)](https://github.com/owner/repo/actions)
[![Release](https://github.com/owner/repo/actions/workflows/release.yml/badge.svg)](https://github.com/owner/repo/actions)
```

### Key Directories & Files

```
.github/
├── workflows/
│   ├── ci.yml              # Main CI pipeline
│   ├── docker.yml          # Docker build & push
│   ├── deploy.yml          # Production deployment
│   ├── quality.yml         # Code quality checks
│   ├── release.yml         # Release management
│   └── scheduled.yml       # Scheduled tasks
└── WORKFLOWS.md            # This file
```

## Environment Variables

### Required GitHub Secrets

```bash
# Deployment (for deploy.yml)
DEPLOY_KEY          # SSH private key
DEPLOY_HOST         # Target server hostname

# Code Quality (for quality.yml)
SONAR_HOST_URL      # SonarQube server URL
SONAR_LOGIN         # SonarQube authentication token
```

### Auto-provided by GitHub

```
GITHUB_TOKEN        # GitHub Actions token
GITHUB_ACTOR        # Triggering user
GITHUB_REF          # Git reference (branch/tag)
GITHUB_SHA          # Commit SHA
```

## Execution Flow

### On Every Push to Feature Branch
```
1. CI workflow triggers
   ├─ Test job
   ├─ Lint job (parallel)
   └─ Security job (parallel)
```

### On Pull Request to Main
```
1. All CI checks run (test, lint, security)
2. Build job waits for all checks
3. Quality analysis runs
4. Comments added to PR with results
```

### On Merge to Main
```
1. CI pipeline (test, lint, build)
2. Docker image builds and pushes
3. Quality checks run
4. Artifacts uploaded
```

### On Version Tag Push (v1.0.0)
```
1. Release workflow triggers
   ├─ Build for Linux, macOS, Windows
   ├─ Generate SHA256 checksums
   └─ Create GitHub Release
2. Docker image builds and pushes with version tag
3. Deploy workflow can be triggered
```

### On Schedule (Daily/Weekly)
```
1. Dependency checks
2. Coverage analysis
3. Health checks
4. Performance baseline
```

## Customization

### Adding a New Workflow

1. Create file: `.github/workflows/my-workflow.yml`
2. Define triggers, jobs, and steps
3. Test locally if possible: `act`
4. Push and monitor in GitHub Actions tab

### Modifying Existing Workflow

1. Edit `.github/workflows/*.yml`
2. Create PR for review
3. Merge to activate changes
4. Changes take effect immediately

### Adding Secrets

```bash
# Using gh CLI
gh secret set SECRET_NAME --body "secret-value"

# Using GitHub UI
Settings → Secrets and variables → Actions → New repository secret
```

## Monitoring & Debugging

### View Workflow Runs

```bash
# List all recent runs
gh run list

# View specific workflow runs
gh run list --workflow ci.yml

# Watch a workflow in real-time
gh run watch <run-id>
```

### View Logs

```bash
# Download full logs
gh run download <run-id>

# View logs in terminal
gh run view <run-id> --log
```

### Enable Debug Logging

```bash
# Enable step debug logging
gh secret set ACTIONS_STEP_DEBUG --body "true"

# Disable when done
gh secret delete ACTIONS_STEP_DEBUG
```

## Performance Tips

1. **Use caching:**
   - Go modules cache
   - Docker layer cache
   - Dependency caches

2. **Parallelize jobs:**
   - Run lint and security checks in parallel
   - Run tests on multiple platforms

3. **Conditional execution:**
   - Only build Docker on main branch
   - Only deploy on version tags
   - Skip expensive jobs for docs-only changes

## Best Practices

1. **Always test locally first:**
   ```bash
   make test
   make lint
   make build
   ```

2. **Use descriptive commit messages:**
   - Helps with release notes generation
   - Follows conventional commits

3. **Keep workflows simple:**
   - Each workflow should have a clear purpose
   - Break complex pipelines into smaller workflows

4. **Monitor workflow health:**
   - Review failed runs
   - Update dependencies when needed
   - Keep tools up to date

5. **Secure sensitive data:**
   - Never commit secrets
   - Use GitHub Secrets for all sensitive data
   - Rotate SSH keys periodically

---

**Documentation:** See [CI-CD-GUIDE.md](../CI-CD-GUIDE.md) for comprehensive CI/CD documentation.
