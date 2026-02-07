# Quick Test Guide - API Endpoint Testing

**100% Coverage Achieved** ✅  
**194 Endpoints** | **9 Test Modules** | **Ready to Execute**

---

## Quick Start

### 1. Start Backend

```bash
cd backend
go run main.go
# Backend runs on http://localhost:8081
```

### 2. Run All Tests

```bash
cd backend/scripts
./run_tests.sh
```

---

## Test Modules

| Module          | Command                      | Endpoints | Description               |
| --------------- | ---------------------------- | --------- | ------------------------- |
| **All**         | `./run_tests.sh`             | 194       | Complete test suite       |
| **Auth**        | `./run_tests.sh auth`        | 11        | Authentication & sessions |
| **RBAC**        | `./run_tests.sh rbac`        | 32        | Roles & permissions       |
| **Documents**   | `./run_tests.sh documents`   | 46        | Document management       |
| **Workflows**   | `./run_tests.sh workflows`   | 37        | Workflow & approvals      |
| **Departments** | `./run_tests.sh departments` | 15        | Department management     |
| **Analytics**   | `./run_tests.sh analytics`   | 19        | Analytics & notifications |
| **Admin**       | `./run_tests.sh admin`       | 44        | Admin endpoints (NEW)     |
| **Errors**      | `./run_tests.sh errors`      | 10        | Error handling            |

---

## Admin Tests (NEW)

### Run Admin Tests Only

```bash
cd backend/scripts
./admin_tests.sh
```

### What It Tests

- ✅ Admin dashboard & analytics (7 endpoints)
- ✅ System health & monitoring (6 endpoints)
- ✅ Subscription management (13 endpoints)
- ✅ Settings management (8 endpoints)
- ✅ Feature flags management (10 endpoints)

### Admin Credentials

- **Email**: admin@liyali.com
- **Password**: password

---

## Common Test Combinations

```bash
# Security testing
./run_tests.sh auth rbac admin errors

# Admin console testing
./run_tests.sh admin analytics

# Document workflow testing
./run_tests.sh documents workflows

# Full regression
./run_tests.sh
```

---

## Test Output

### Success

```
✓ GET /admin/dashboard (HTTP 200)
✓ GET /admin/analytics (HTTP 200)
...
Success Rate: 100.0%
All tests passed!
```

### Failure

```
✗ GET /admin/dashboard (Expected HTTP 200, got 500)
...
Success Rate: 95.5%
Some tests failed.
```

---

## Troubleshooting

### Backend Not Running

```
Error: Failed to connect to http://localhost:8081
Solution: Start backend with 'go run main.go'
```

### Authentication Failed

```
Error: Admin login failed (HTTP 401)
Solution: Check admin credentials in .env file
```

### Database Not Seeded

```
Error: No data found (HTTP 404)
Solution: Run migrations and seed data
```

---

## Test Results Location

- **Console Output**: Real-time test results
- **Exit Code**: 0 = success, 1 = failure
- **Summary**: Displayed at end of test run

---

## CI/CD Integration

```yaml
# Example GitHub Actions
- name: Run API Tests
  run: |
    cd backend/scripts
    ./run_tests.sh
```

---

## Documentation

- **Full Guide**: `backend/scripts/README_TESTS.md`
- **Coverage Report**: `backend/scripts/API_ENDPOINT_COVERAGE_REPORT.md`
- **Implementation**: `API_TEST_COVERAGE_IMPLEMENTATION_SUMMARY.md`

---

**Ready to test!** 🚀
