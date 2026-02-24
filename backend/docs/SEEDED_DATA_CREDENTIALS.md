# 🔐 Seeded Data & Test Credentials

**Database:** Supabase PostgreSQL (Staging)  
**Last Updated:** February 24, 2026  
**Status:** ✅ Active

---

## 👥 Test Users

All users have the same password for testing purposes.

### Default Password

```
password
```

**Bcrypt Hash:** `$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi`

---

## 📋 User Accounts

### 1. System Administrator

- **Email:** `admin@liyali.com`
- **Password:** `password`
- **Name:** System Administrator
- **Role:** admin
- **Super Admin:** Yes
- **Organization:** Liyali Demo Organization
- **Department:** IT
- **User ID:** `user-admin-001`

**Permissions:**

- Full system access
- Can manage all organizations
- Can access admin console
- Can manage users and roles
- Can view all data

---

### 2. Requester

- **Email:** `requester@liyali.com`
- **Password:** `password`
- **Name:** John Requester
- **Role:** requester
- **Super Admin:** No
- **Organization:** Liyali Demo Organization
- **Department:** Operations
- **User ID:** `user-requester-001`

**Permissions:**

- Create requisitions
- View own requisitions
- Submit for approval
- Cannot approve

---

### 3. Approver

- **Email:** `approver@liyali.com`
- **Password:** `password`
- **Name:** Jane Approver
- **Role:** approver
- **Super Admin:** No
- **Organization:** Liyali Demo Organization
- **Department:** Finance
- **User ID:** `user-approver-001`

**Permissions:**

- Approve requisitions
- View requisitions assigned to them
- Create requisitions
- Cannot access finance functions

---

### 4. Finance Manager

- **Email:** `finance@liyali.com`
- **Password:** `password`
- **Name:** Bob Finance
- **Role:** finance
- **Super Admin:** No
- **Organization:** Liyali Demo Organization
- **Department:** Finance
- **User ID:** `user-finance-001`

**Permissions:**

- Full finance access
- Approve payments
- Manage budgets
- View all financial documents
- Create purchase orders

---

### 5. Manager

- **Email:** `manager@liyali.com`
- **Password:** `password`
- **Name:** Alice Manager
- **Role:** approver
- **Super Admin:** No
- **Organization:** Liyali Demo Organization
- **Department:** Operations
- **User ID:** `user-manager-001`

**Permissions:**

- Approve requisitions
- View team requisitions
- Create requisitions
- Manage department workflows

---

### 6. Viewer

- **Email:** `viewer@liyali.com`
- **Password:** `password`
- **Name:** Charlie Viewer
- **Role:** viewer
- **Super Admin:** No
- **Organization:** Liyali Demo Organization
- **Department:** IT
- **User ID:** `user-viewer-001`

**Permissions:**

- Read-only access
- View requisitions
- View reports
- Cannot create or approve

---

## 🏢 Organizations

### 1. Liyali Demo Organization

- **ID:** `org-demo-001`
- **Name:** Liyali Demo Organization
- **Slug:** `liyali-demo`
- **Tier:** Pro (basic)
- **Status:** Trial
- **Trial Start:** 2026-02-24
- **Trial End:** 2026-03-10 (30 days)
- **Description:** Default organization for testing and development

**Members:** 6 users (all test users above)

**Departments:**

- Information Technology (IT)
- Finance (FIN)
- Operations (OPS)
- Human Resources (HR)
- Procurement (PROC)

---

### 2. Enterprise Corp

- **ID:** `org-enterprise-001`
- **Name:** Enterprise Corp
- **Slug:** `enterprise-corp`
- **Tier:** Enterprise (basic)
- **Status:** Trial
- **Trial Start:** 2026-02-24
- **Trial End:** 2026-03-10 (30 days)
- **Description:** Large enterprise organization for testing enterprise features

**Members:** None (empty organization for testing)

---

## 💳 Subscription Tiers

### 1. Basic (Free)

- **ID:** `tier-basic`
- **Price:** $0/month, $0/year
- **Max Users:** 5
- **Storage:** 1 GB
- **Features:**
  - Document Management
  - Basic Workflows
  - Email Notifications

---

### 2. Professional

- **ID:** `tier-professional`
- **Price:** $50/month, $500/year (save $100)
- **Max Users:** 25
- **Storage:** 10 GB
- **Features:**
  - Document Management
  - Advanced Workflows
  - Email Notifications
  - Custom Roles
  - Analytics & Reporting
  - API Access

---

### 3. Enterprise

- **ID:** `tier-enterprise`
- **Price:** $150/month, $1,500/year (save $300)
- **Max Users:** 100
- **Storage:** 50 GB
- **Features:**
  - All Professional features
  - Single Sign-On (SSO)
  - Audit Logs
  - Priority Support (24/7)

---

### 4. Unlimited

- **ID:** `tier-unlimited`
- **Price:** $500/month, $5,000/year (save $1,000)
- **Max Users:** Unlimited (-1)
- **Storage:** Unlimited (-1)
- **Features:**
  - All Enterprise features
  - Custom Integrations
  - Dedicated Support
  - Custom SLA

---

## 🎯 Subscription Features

### Core Features

- **Document Management** - Create, edit, and manage documents

### Workflow Features

- **Basic Workflows** - Simple approval workflows
- **Advanced Workflows** - Complex multi-stage workflows with conditions

### Communication Features

- **Email Notifications** - Automated email notifications

### Security Features

- **Custom Roles** - Create and manage custom user roles
- **Single Sign-On (SSO)** - SAML/OAuth SSO integration
- **Audit Logs** - Comprehensive audit trail

### Analytics Features

- **Analytics & Reporting** - Detailed analytics and custom reports

### Integration Features

- **API Access** - REST API access for integrations
- **Custom Integrations** - Custom API integrations and webhooks

### Support Features

- **Priority Support** - 24/7 priority customer support
- **Dedicated Support** - Dedicated customer success manager

---

## 🧪 Testing Scenarios

### Scenario 1: Basic Workflow

1. Login as `requester@liyali.com`
2. Create a requisition
3. Submit for approval
4. Login as `approver@liyali.com`
5. Approve the requisition
6. Login as `finance@liyali.com`
7. Create purchase order

### Scenario 2: Multi-Level Approval

1. Login as `requester@liyali.com`
2. Create high-value requisition
3. Login as `manager@liyali.com`
4. First-level approval
5. Login as `finance@liyali.com`
6. Final approval and payment

### Scenario 3: Admin Functions

1. Login as `admin@liyali.com`
2. Access admin console
3. Manage users
4. View analytics
5. Configure system settings

### Scenario 4: Read-Only Access

1. Login as `viewer@liyali.com`
2. View requisitions (read-only)
3. View reports (read-only)
4. Cannot create or approve

---

## 🔑 API Authentication

### Login Request

```bash
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@liyali.com",
    "password": "password"
  }'
```

### Expected Response

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "user-admin-001",
    "email": "admin@liyali.com",
    "name": "System Administrator",
    "role": "admin",
    "organization_id": "org-demo-001"
  }
}
```

### Using Token

```bash
curl -X GET http://localhost:8081/api/v1/requisitions \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

---

## 📊 Database Queries

### Check All Users

```sql
SELECT id, email, name, role, active, is_super_admin
FROM users
ORDER BY email;
```

### Check Organizations

```sql
SELECT id, name, subscription_tier, subscription_status,
       trial_start_date, trial_end_date
FROM organizations
ORDER BY name;
```

### Check Subscription Tiers

```sql
SELECT id, name, display_name, price_monthly, price_yearly,
       max_users, storage_limit_gb
FROM subscription_tiers
ORDER BY sort_order;
```

### Check User Organization Membership

```sql
SELECT
    u.email,
    u.name,
    o.name as organization,
    om.role,
    om.department
FROM users u
JOIN organization_members om ON u.id = om.user_id
JOIN organizations o ON om.organization_id = o.id
ORDER BY u.email;
```

---

## 🔒 Security Notes

### Password Security

- All test accounts use the same password: `password`
- **DO NOT use these credentials in production**
- Change all passwords before deploying to production
- Implement password complexity requirements

### Super Admin Access

- Only `admin@liyali.com` has super admin privileges
- Super admin can access all organizations
- Super admin can manage system settings
- Limit super admin accounts in production

### Trial Period

- Both organizations are on 30-day trial
- Trial ends on 2026-03-10
- Implement trial expiration logic
- Set up subscription upgrade flows

---

## 🚀 Quick Login Commands

### Backend API

```bash
# Admin
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@liyali.com","password":"password"}'

# Requester
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"requester@liyali.com","password":"password"}'

# Approver
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"approver@liyali.com","password":"password"}'

# Finance
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"finance@liyali.com","password":"password"}'
```

---

## 📝 Notes

- All users are active and can login immediately
- All users belong to "Liyali Demo Organization"
- Trial period is 30 days from seed date
- Default currency is USD
- Fiscal year starts on January 1st
- Budget validation is enabled
- Digital signatures are required

---

**Status:** ✅ Ready for Testing  
**Environment:** Staging (Supabase)  
**Last Verified:** February 24, 2026
