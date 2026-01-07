# 🔐 Authentication Users Guide

## Overview

This document provides login credentials and access information for all seeded users in the Liyali Gateway system. These users are created during the database seeding process and provide comprehensive role coverage for testing and development.

---

## 🏢 Organizations

### Default Organization
- **ID**: `org-default-001`
- **Name**: Default Organization
- **Slug**: `default-org`
- **Description**: Default organization for initial setup
- **Tier**: Free

### Demo Corporation
- **ID**: `org-demo-001`
- **Name**: Demo Corporation
- **Slug**: `demo-corp`
- **Description**: Demo organization for testing and development
- **Tier**: Premium

---

## 👥 User Accounts

### 1. System Administrator
- **Email**: `admin@liyali.com`
- **Password**: `admin123`
- **Name**: System Administrator
- **Role**: `admin`
- **Organization**: Default Organization
- **Super Admin**: Yes ✅

**Access & Permissions**:
- ✅ **Full System Access**: Complete administrative control
- ✅ **User Management**: Create, edit, delete users
- ✅ **Organization Management**: Manage all organizations
- ✅ **System Configuration**: Modify system settings
- ✅ **Workflow Management**: Create and modify workflows
- ✅ **Master Data**: Manage vendors, categories, budgets
- ✅ **Reports & Analytics**: Access all reports and dashboards
- ✅ **Audit Logs**: View all system audit trails

**Typical Use Cases**:
- System setup and configuration
- User account management
- Troubleshooting and support
- System monitoring and maintenance

---

### 2. John Requester (Operations Specialist)
- **Email**: `requester@demo.com`
- **Password**: `admin123`
- **Name**: John Requester
- **Role**: `requester`
- **Organization**: Demo Corporation
- **Department**: Operations

**Access & Permissions**:
- ✅ **Create Requisitions**: Submit purchase requests
- ✅ **View Own Documents**: Access personal requisitions and related documents
- ✅ **Edit Draft Documents**: Modify documents in draft status
- ✅ **Track Approvals**: Monitor approval status and history
- ✅ **Basic Reporting**: View personal activity reports
- ❌ **Approve Documents**: Cannot approve requests
- ❌ **Manage Users**: No user management access
- ❌ **System Configuration**: No admin access

**Typical Use Cases**:
- Creating purchase requisitions
- Tracking request status
- Updating draft documents
- Viewing personal activity

---

### 3. Jane Approver (Department Head)
- **Email**: `approver@demo.com`
- **Password**: `admin123`
- **Name**: Jane Approver
- **Role**: `approver`
- **Organization**: Demo Corporation
- **Department**: Management

**Access & Permissions**:
- ✅ **Approve/Reject Documents**: Final approval authority
- ✅ **View All Pending Approvals**: Access approval queue
- ✅ **Reassign Tasks**: Delegate approval tasks
- ✅ **Add Comments**: Provide approval feedback
- ✅ **Digital Signatures**: Sign approved documents
- ✅ **Approval Reports**: View approval analytics
- ✅ **Workflow Monitoring**: Track workflow progress
- ❌ **Create Workflows**: Cannot modify workflow definitions
- ❌ **User Management**: Limited user access

**Typical Use Cases**:
- Reviewing and approving requisitions
- Managing approval workflows
- Providing approval feedback
- Monitoring team requests

---

### 4. Bob Finance (Finance Officer)
- **Email**: `finance@demo.com`
- **Password**: `admin123`
- **Name**: Bob Finance
- **Role**: `finance`
- **Organization**: Demo Corporation
- **Department**: Finance

**Access & Permissions**:
- ✅ **Financial Approvals**: Approve payment vouchers and budgets
- ✅ **Budget Management**: Create and manage budgets
- ✅ **Payment Processing**: Process payment vouchers
- ✅ **Financial Reports**: Access financial analytics
- ✅ **Vendor Management**: Manage vendor information
- ✅ **Cost Center Tracking**: Monitor cost centers and projects
- ✅ **GL Code Management**: Manage general ledger codes
- ❌ **Final Approvals**: May require additional approval for high amounts
- ❌ **System Administration**: No admin access

**Typical Use Cases**:
- Reviewing financial documents
- Managing budgets and allocations
- Processing payments
- Financial reporting and analysis

---

### 5. Alice Manager (Operations Manager)
- **Email**: `manager@demo.com`
- **Password**: `admin123`
- **Name**: Alice Manager
- **Role**: `department_manager`
- **Organization**: Demo Corporation
- **Department**: Operations

**Access & Permissions**:
- ✅ **Department Oversight**: Manage department operations
- ✅ **First-Level Approvals**: Initial approval in workflow
- ✅ **Team Management**: Oversee team members
- ✅ **Budget Oversight**: Monitor department budgets
- ✅ **Requisition Review**: Review team requisitions
- ✅ **Department Reports**: Access department analytics
- ✅ **Workflow Participation**: Participate in approval workflows
- ❌ **Final Approvals**: Usually requires higher-level approval
- ❌ **Cross-Department Access**: Limited to own department

**Typical Use Cases**:
- Managing department operations
- First-level approval of requests
- Budget monitoring and planning
- Team oversight and coordination

---

## 🔄 Workflow Roles & Responsibilities

### Standard Requisition Approval Workflow
1. **Requester** (John) → Creates requisition
2. **Department Manager** (Alice) → First review and approval
3. **Finance** (Bob) → Financial review and validation
4. **Final Approver** (Jane) → Final approval and authorization

### Express Requisition Approval Workflow (Low Value)
1. **Requester** (John) → Creates requisition
2. **Department Manager** (Alice) → Review and approval
3. **Final Approver** (Jane) → Final authorization

### Payment Voucher Workflow
1. **Finance** (Bob) → Financial review and validation
2. **Final Approver** (Jane) → Final approval and payment authorization

---

## 🏗️ Sample Data Available

### Organizations
- **2 Organizations**: Default + Demo Corporation with complete setup

### Master Data
- **5 Vendors**: Office supplies, tech solutions, facility services, catering, equipment rental
- **6 Categories**: Office supplies, IT equipment, facility maintenance, professional services, travel, marketing
- **4 Budgets**: Approved budgets for different departments (Office: $50K, IT: $100K, Facility: $75K, Marketing: $80K)

### Workflows
- **6 Complete Workflows**: Covering all document types with proper approval chains

### Sample Documents
- **3 Sample Requisitions**: Different statuses for testing (draft, pending, approved)

---

## 🔧 Testing Scenarios

### Scenario 1: Complete Requisition Flow
1. **Login as John** (`requester@demo.com`) → Create new requisition
2. **Login as Alice** (`manager@demo.com`) → Review and approve
3. **Login as Bob** (`finance@demo.com`) → Financial review
4. **Login as Jane** (`approver@demo.com`) → Final approval

### Scenario 2: Budget Management
1. **Login as Bob** (`finance@demo.com`) → Create/modify budgets
2. **Login as Alice** (`manager@demo.com`) → Review department budgets
3. **Login as Jane** (`approver@demo.com`) → Approve budget changes

### Scenario 3: System Administration
1. **Login as Admin** (`admin@liyali.com`) → Full system access
2. **User Management**: Create/modify user accounts
3. **Workflow Configuration**: Set up approval workflows
4. **System Monitoring**: View audit logs and reports

---

## 🚀 Quick Start Guide

### For Developers
1. **Use Admin Account** for initial setup and testing
2. **Test Workflows** with different user roles
3. **Verify Permissions** across different access levels
4. **Check Integration** with frontend components

### For Business Users
1. **Start with Requester Role** to understand basic functionality
2. **Progress to Manager Role** for approval workflows
3. **Use Finance Role** for financial operations
4. **Admin Role** for system configuration

### For QA Testing
1. **Test All User Roles** to verify access controls
2. **Validate Workflows** end-to-end
3. **Check Permission Boundaries** (what users can/cannot do)
4. **Verify Data Integrity** across role interactions

---

## 🔒 Security Notes

### Password Policy
- **Current Password**: `admin123` for all accounts (development only)
- **Production**: Change all passwords before production deployment
- **Recommendation**: Implement strong password requirements

### Access Control
- **Role-Based Access**: Each role has specific permissions
- **Organization Isolation**: Users can only access their organization data
- **Audit Trail**: All actions are logged for security monitoring

### Best Practices
- **Regular Password Updates**: Change passwords regularly
- **Principle of Least Privilege**: Users have minimum required access
- **Session Management**: Secure session handling with refresh tokens
- **Multi-Factor Authentication**: Consider implementing MFA for production

---

## 📞 Support & Troubleshooting

### Common Issues
1. **Login Failed**: Verify email and password (case-sensitive)
2. **Access Denied**: Check user role and organization membership
3. **Missing Data**: Ensure database seeding completed successfully
4. **Workflow Issues**: Verify workflow configuration and user assignments

### Database Verification
```sql
-- Check users
SELECT email, name, role, active FROM users;

-- Check organizations
SELECT name, slug, active FROM organizations;

-- Check organization members
SELECT u.email, u.name, om.role, o.name as organization 
FROM organization_members om
JOIN users u ON om.user_id = u.id
JOIN organizations o ON om.organization_id = o.id;
```

### Reset Instructions
```bash
# Reset database with fresh data
cd backend/database
./migrate.sh reset

# Seed data only (if schema exists)
./migrate.sh seed
```

---

## 📊 System Statistics

After successful seeding, your database contains:
- ✅ **30+ Tables** with complete schema
- ✅ **161 Indexes** for optimal performance
- ✅ **2 Organizations** with full configuration
- ✅ **5 Users** covering all role types
- ✅ **5 Vendors** with complete information
- ✅ **6 Categories** with budget code mappings
- ✅ **4 Budgets** with realistic allocations
- ✅ **6 Workflows** with complete approval chains
- ✅ **3 Sample Requisitions** for immediate testing

---

**Status**: ✅ **Ready for Development and Testing**
**Last Updated**: January 7, 2025
**Database Version**: Consolidated Schema v1.0