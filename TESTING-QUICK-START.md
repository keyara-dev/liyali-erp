# Testing Quick Start Guide - Liyali Gateway MVP

**Last Updated**: 2025-12-26
**Status**: ✅ **Ready to Execute**

---

## 🚀 Quick Start (5 Minutes)

### Option 1: Docker Setup (Recommended)
```bash
# 1. Start all services
docker-compose up -d

# 2. Verify services are running
docker-compose ps

# 3. Check backend health
curl http://localhost:8080/health

# 4. View logs (if needed)
docker-compose logs -f backend
```

### Option 2: Manual API Testing
```bash
# Use REST Client extension or Postman
# 1. Open VS Code
# 2. Install REST Client extension
# 3. Open backend/API.http
# 4. Click "Send Request" on any endpoint
```

### Option 3: Run Go Tests
```bash
# 1. Install Go dependencies
cd backend
go mod tidy

# 2. Run all tests
go test -v ./...

# 3. Or run specific tests
go test -v ./handlers    # Handler tests
go test -v ./services    # Service tests
```

---

## 📋 What to Test (10 Critical Tests)

### Test 1: User Registration
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "TestPass123",
    "name": "Test User",
    "role": "requester"
  }'
```
**Expected**: Account created, personal organization auto-created

### Test 2: User Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "TestPass123"
  }'
```
**Expected**: JWT token returned

### Test 3: Create Requisition
```bash
TOKEN="<JWT_TOKEN_FROM_LOGIN>"
curl -X POST http://localhost:8080/api/v1/requisitions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Office Supplies",
    "description": "Monthly office supplies",
    "department": "Admin",
    "priority": "medium",
    "items": [
      {"description": "Paper", "quantity": 10, "unitPrice": 5.00}
    ],
    "totalAmount": 50.00,
    "currency": "USD"
  }'
```
**Expected**: Requisition created in draft state

### Tests 4-10
See [E2E-TEST-PLAN.md](E2E-TEST-PLAN.md) for complete test cases TC-3.2 through TC-8.2

---

## 📖 Documentation Map

| Document | Purpose | Read Time |
|----------|---------|-----------|
| **This File** | Quick start guide | 5 min |
| [E2E-TEST-PLAN.md](E2E-TEST-PLAN.md) | 50+ detailed test cases | 20 min |
| [E2E-TEST-EXECUTION-GUIDE.md](E2E-TEST-EXECUTION-GUIDE.md) | Step-by-step execution | 15 min |
| [BACKEND-TEST-REPORT-2025-12-26.md](BACKEND-TEST-REPORT-2025-12-26.md) | API & infrastructure | 25 min |
| [TEST-EXECUTION-SUMMARY-2025-12-26.md](TEST-EXECUTION-SUMMARY-2025-12-26.md) | Complete overview | 10 min |

---

## ✅ Pre-Test Checklist

- [ ] Docker installed (`docker --version`)
- [ ] Git repository cloned
- [ ] All files up to date (`git status`)
- [ ] 3+ hours available for testing
- [ ] Modern browser available (Chrome/Firefox)
- [ ] Terminal/PowerShell open

---

## 🔍 Verify Each Component

### Backend Running?
```bash
curl http://localhost:8080/health
# Should return: {"status":"ok"}
```

### Database Connected?
```bash
curl http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer <TOKEN>"
# Should return user profile
```

### Frontend Running?
```bash
curl http://localhost:3000
# Should return HTML page
```

---

## 📊 Test Results Checklist

### Critical Tests (Must All Pass)
- [ ] TC-1.1: User Registration
- [ ] TC-1.2: User Login
- [ ] TC-3.1: Create Requisition
- [ ] TC-3.2: Submit for Approval
- [ ] TC-3.3: Approve Requisition
- [ ] TC-2.1: Personal Org Auto-Creation
- [ ] TC-2.3: Data Isolation
- [ ] TC-7.2: Data Persistence
- [ ] TC-1.3: RBAC
- [ ] TC-8.2: Permission Enforcement

### Success Criteria
- ✅ 10/10 critical tests pass
- ✅ 24/26 total tests pass (92%+)
- ✅ 0 critical defects
- ✅ No 500 errors
- ✅ Data persists correctly

---

## 🛠️ Troubleshooting

### Docker won't start?
```bash
# Check Docker is running
docker ps

# If not, start Docker Desktop (Windows/Mac) or daemon (Linux)
```

### Port 8080 in use?
```bash
# Kill the process using port 8080
# Windows:
netstat -ano | findstr :8080
taskkill /PID <PID> /F

# Linux/Mac:
lsof -i :8080
kill -9 <PID>
```

### Tests fail?
1. Check logs: `docker-compose logs backend`
2. Verify database: `docker-compose logs postgres`
3. See [E2E-TEST-EXECUTION-GUIDE.md](E2E-TEST-EXECUTION-GUIDE.md) troubleshooting section

---

## 📞 Quick Links

- **API Examples**: [backend/API.http](backend/API.http)
- **Full Test Plan**: [E2E-TEST-PLAN.md](E2E-TEST-PLAN.md)
- **Execution Guide**: [E2E-TEST-EXECUTION-GUIDE.md](E2E-TEST-EXECUTION-GUIDE.md)
- **Backend Report**: [BACKEND-TEST-REPORT-2025-12-26.md](BACKEND-TEST-REPORT-2025-12-26.md)
- **Test Summary**: [TEST-EXECUTION-SUMMARY-2025-12-26.md](TEST-EXECUTION-SUMMARY-2025-12-26.md)

---

## ⏱️ Timeline

| Phase | Duration | Action |
|-------|----------|--------|
| **Setup** | 10 min | Docker Compose up |
| **Health** | 5 min | Verify endpoints |
| **Auth** | 30 min | Register & login |
| **Workflows** | 60 min | Create → Approve |
| **Data** | 30 min | Isolation & persistence |
| **Integration** | 45 min | Complex flows |
| **TOTAL** | 180 min | 3 hours |

---

## 🎯 Success Criteria

**MVP is ready when:**
- ✅ All 10 critical tests pass
- ✅ 24+ of 26 total tests pass
- ✅ No critical defects
- ✅ Team sign-off obtained
- ✅ Ready to deploy

---

## 🚀 Next Actions

1. **Now**: Read this file (5 min)
2. **Next**: Run `docker-compose up -d` (2 min)
3. **Then**: Follow [E2E-TEST-EXECUTION-GUIDE.md](E2E-TEST-EXECUTION-GUIDE.md)
4. **Finally**: Document results and sign off

---

## 📝 Notes

- All endpoint examples in [backend/API.http](backend/API.http)
- Detailed steps in [E2E-TEST-PLAN.md](E2E-TEST-PLAN.md)
- Infrastructure details in [BACKEND-TEST-REPORT-2025-12-26.md](BACKEND-TEST-REPORT-2025-12-26.md)
- Complete documentation map in [TEST-EXECUTION-SUMMARY-2025-12-26.md](TEST-EXECUTION-SUMMARY-2025-12-26.md)

---

**Status**: ✅ **Ready to test**
**Time to MVP**: ~3 hours

**Start Now**: `docker-compose up -d`

🚀 **Good luck!**
