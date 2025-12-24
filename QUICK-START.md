# Quick Start Guide - Phase 2 Features

**TL;DR:** Build → Test → Deploy

---

## 🚀 In 5 Minutes

### 1. Build Backend
```bash
cd backend
go build -o liyali-gateway
```

### 2. Run Backend
```bash
./liyali-gateway
```
Server starts on `http://localhost:8080`

### 3. Test in Postman
1. Import `postman-collection.json`
2. Run Login request → copy token
3. Run Category tests
4. Run Requisition tests
5. Run Analytics tests

---

## 🧪 Test Checklist

### Phase 1: Categories
```bash
# Create
curl -X POST http://localhost:8080/api/v1/categories \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Office Supplies","description":"Supplies","budgetCodes":["BDG-001"]}'

# Expected: Status 201, returns categoryId
```

### Phase 2: Requisitions
```bash
# Create with category
curl -X POST http://localhost:8080/api/v1/requisitions \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title":"Purchase",
    "description":"Buy stuff",
    "department":"Finance",
    "priority":"high",
    "items":[{"description":"Paper","quantity":10,"unitPrice":5,"amount":50}],
    "totalAmount":50,
    "currency":"USD",
    "categoryId":"CATEGORY_ID",
    "preferredVendorId":"VENDOR_ID",
    "isEstimate":true
  }'

# Expected: Status 201, includes categoryName, preferredVendorName, isEstimate
```

### Phase 3: Last Login
```bash
# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@liyali.com","password":"admin123"}'

# Expected: Response includes user.lastLogin with timestamp
```

### Phase 4: Analytics
```bash
# Get metrics
curl -X GET http://localhost:8080/api/v1/analytics/requisitions/metrics \
  -H "Authorization: Bearer TOKEN"

# Expected: statusCounts, rejectionRate, rejectionsOverTime, rejectionReasons, topRejectingApprovers
```

---

## 📊 Verify Database

```bash
# Connect to PostgreSQL
psql -U postgres -d liyali_gateway

# Check new tables
SELECT * FROM categories;
SELECT * FROM category_budget_codes;

# Check new columns
SELECT id, category_id, preferred_vendor_id, is_estimate FROM requisitions LIMIT 5;
SELECT id, email, last_login FROM users WHERE last_login IS NOT NULL LIMIT 5;
```

---

## 🎯 Success Indicators

✅ Backend builds without errors
✅ Server starts with "Database migrations completed"
✅ Login request returns `lastLogin` field
✅ Create category returns 201 with categoryId
✅ Create requisition with categoryId works
✅ Analytics endpoint returns metrics
✅ All Postman tests pass

---

## 📚 Documentation Files

| File | Purpose |
|------|---------|
| `PHASE-2-IMPLEMENTATION-SUMMARY.md` | Complete feature overview |
| `IMPLEMENTATION-CHECKLIST.md` | Detailed next steps |
| `TESTING-GUIDE.md` | Comprehensive testing guide |
| `postman-collection.json` | 25 pre-configured API tests |
| `QUICK-START.md` | This file |

---

## 🔧 Common Commands

```bash
# Build
cd backend && go build -o liyali-gateway

# Run
./liyali-gateway

# Run tests
go test ./... -v

# Run specific tests
go test -v ./handlers -run Category
go test -v ./services -run Analytics

# Run with coverage
go test ./... -v -cover
```

---

## 🚨 If Something Breaks

### Error: "command not found: go"
→ Go not installed. Install Go 1.21+

### Error: "database connection failed"
→ PostgreSQL not running. Start with: `psql` or Docker

### Error: "migration failed"
→ Drop database: `DROP DATABASE liyali_gateway;` then restart

### Error: "404 not found"
→ Ensure routes.go has category routes (check line 87-96)

### Error: "unauthorized"
→ Forgot to add Authorization header. Get token from login first.

---

## 📞 Quick Reference

### Endpoints
```
Categories:
  POST   /api/v1/categories
  GET    /api/v1/categories
  GET    /api/v1/categories/{id}
  PUT    /api/v1/categories/{id}
  DELETE /api/v1/categories/{id}
  GET    /api/v1/categories/{id}/budget-codes
  POST   /api/v1/categories/{id}/budget-codes
  DELETE /api/v1/categories/{id}/budget-codes/{code}

Requisitions (Enhanced):
  POST   /api/v1/requisitions        (now accepts categoryId, preferredVendorId, isEstimate)
  GET    /api/v1/requisitions
  GET    /api/v1/requisitions/{id}
  PUT    /api/v1/requisitions/{id}

Analytics:
  GET /api/v1/analytics/requisitions/metrics
  GET /api/v1/analytics/approvals/metrics
  GET /api/v1/analytics/dashboard
```

### Query Parameters
```
Categories:
  ?page=1&limit=10&active=true

Requisitions:
  ?status=draft&department=Finance&priority=high

Analytics:
  ?start_date=2025-12-01&end_date=2025-12-31&period=daily&department=Finance
```

---

## ✨ New Fields

### Requisition
```json
{
  "categoryId": "uuid",
  "categoryName": "Office Supplies",
  "preferredVendorId": "uuid",
  "preferredVendorName": "Vendor Name",
  "isEstimate": true
}
```

### User (Login Response)
```json
{
  "id": "uuid",
  "email": "user@example.com",
  "lastLogin": "2025-12-24T10:30:00Z"
}
```

### Analytics Response
```json
{
  "statusCounts": { "draft": 5, "pending": 3, "approved": 10, "rejected": 2 },
  "rejectionRate": 11.76,
  "rejectionsOverTime": [...],
  "rejectionReasons": [...],
  "topRejectingApprovers": [...],
  "totalRequisitions": 20
}
```

---

## 🎓 What Was Built

| Feature | Lines | Tests | Status |
|---------|-------|-------|--------|
| Categories | 550 | 7 | ✅ |
| Requisitions | 100 | ✓ | ✅ |
| Last Login | 50 | ✓ | ✅ |
| Analytics | 320 | 6 | ✅ |
| **Total** | **1,020+** | **13** | **✅** |

---

## 🎉 You're Ready!

1. **Run the backend** - `./liyali-gateway`
2. **Test the APIs** - Use Postman collection
3. **Verify database** - Check schema with SQL
4. **Deploy** - When ready

See detailed docs for more information.

**Happy testing! 🚀**
