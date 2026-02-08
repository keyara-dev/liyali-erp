# Budget API Diagnostic

## Issue

The budgets table on the frontend is not showing any budgets even though one was created.

## Potential Causes

1. **Backend not running** - The API server needs to be running on port 8080
2. **Database not seeded** - The seed data includes 4 sample budgets
3. **Permission issue** - User might not have `budget:view` permission
4. **Response format issue** - Data might be nested incorrectly
5. **Frontend cache** - React Query might be caching an empty response

## Diagnostic Steps

### 1. Check if backend is running

```bash
# Windows CMD
curl http://localhost:8080/health

# Or check if the process is running
netstat -ano | findstr :8080
```

### 2. Check database for budgets

```sql
-- Connect to your PostgreSQL database and run:
SELECT id, budget_code, name, organization_id, status, fiscal_year, total_budget
FROM budgets
WHERE organization_id = 'org-demo-001';
```

### 3. Test the API endpoint directly

```bash
# First, login to get a token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"admin@liyali.com\",\"password\":\"password\"}"

# Copy the access_token from the response, then:
curl -X GET "http://localhost:8080/api/v1/budgets?page=1&limit=10" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "X-Organization-ID: org-demo-001"
```

### 4. Check browser console

Open the browser developer tools (F12) and look for:

- The console.log output: "Budget response:"
- Any network errors in the Network tab
- Check the `/api/v1/budgets` request and response

### 5. Check user permissions

```sql
-- Check what role the logged-in user has
SELECT om.role, om.department, u.email
FROM organization_members om
JOIN users u ON u.id = om.user_id
WHERE om.organization_id = 'org-demo-001'
AND om.user_id = 'YOUR_USER_ID';
```

## Quick Fixes

### Fix 1: Clear React Query cache

In the browser console, run:

```javascript
localStorage.clear();
sessionStorage.clear();
location.reload();
```

### Fix 2: Force refetch

Click the "Create Budget" button and cancel - this triggers a refetch.

### Fix 3: Check the response in browser DevTools

1. Open DevTools (F12)
2. Go to Network tab
3. Refresh the budgets page
4. Look for the `/api/v1/budgets` request
5. Check the Response tab to see what data is being returned

## Expected Response Format

The backend should return:

```json
{
  "success": true,
  "message": "Budgets retrieved successfully",
  "data": [
    {
      "id": "budget-it-001",
      "budgetCode": "IT-EQUIP",
      "name": "IT Equipment Budget 2026",
      "totalBudget": 50000,
      "allocatedAmount": 0,
      "remainingAmount": 50000,
      "status": "active",
      "fiscalYear": "2026",
      ...
    }
  ],
  "pagination": {
    "page": 1,
    "pageSize": 10,
    "total": 4,
    "totalPages": 1,
    "hasNext": false,
    "hasPrev": false
  }
}
```

The frontend should extract `response.data` which is the budgets array.
