# GAP CLOSURE IMPLEMENTATION PLAN

**Project**: Liyali Gateway System Completion  
**Date**: February 19, 2026  
**Version**: 1.0  
**Status**: Ready for Execution

---

## OVERVIEW

This document provides a detailed, actionable implementation plan to close all identified gaps in the Liyali Gateway system. The plan is organized into 4 phases with clear deliverables, timelines, and resource requirements.

---

## PHASE 1: CRITICAL GAPS (WEEKS 1-2)

**Goal**: Address critical gaps blocking production deployment  
**Duration**: 2 weeks  
**Team Size**: 2-3 developers  
**Estimated Effort**: 80-100 hours

### Sprint 1.1: Email & Notification Infrastructure (Week 1)

#### Task 1.1.1: Email Service Integration

**Gap**: GAP-AUTH-001, GAP-AUTH-003, GAP-NOTIF-001  
**Priority**: CRITICAL  
**Effort**: 16 hours

**Implementation Steps**:

1. Choose email provider (SendGrid recommended)
   - Create SendGrid account
   - Generate API key
   - Add to environment variables

2. Backend Implementation:

   ```go
   // backend/services/email_service.go
   - Create EmailService struct
   - Implement SendEmail() method
   - Add email templates (HTML + plain text)
   - Implement retry logic with exponential backoff
   ```

3. Email Templates to Create:
   - Welcome email
   - Email verification
   - Password reset
   - Approval notification
   - Document status change

4. Integration Points:
   - `auth_service.go`: SendVerificationEmail()
   - `auth_service.go`: SendPasswordResetEmail()
   - `notification_service.go`: ProcessPendingNotifications()

5. Testing:
   - Unit tests for email service
   - Integration tests for email delivery
   - Test email templates in different clients

**Deliverables**:

- ✅ Email service fully integrated
- ✅ All email templates created
- ✅ Email delivery working end-to-end
- ✅ Tests passing

**Acceptance Criteria**:

- Users receive verification emails within 30 seconds
- Password reset emails delivered successfully
- Approval notifications sent automatically

#### Task 1.1.2: Email Verification System
**Gap**: GAP-AUTH-001  
**Priority**: CRITICAL  
**Effort**: 12 hours

**Implementation Steps**:
1. Database Schema:
   ```sql
   ALTER TABLE users ADD COLUMN email_verified BOOLEAN DEFAULT FALSE;
   ALTER TABLE users ADD COLUMN verification_token VARCHAR(255);
   ALTER TABLE users ADD COLUMN verification_token_expires_at TIMESTAMP;
   ```

2. Backend Implementation:
   - Generate verification token on registration
   - Send verification email
   - Implement `/auth/verify` endpoint
   - Add email verification check to login

3. Frontend Implementation:
   - Create email verification paws

---

**Document Version**: 1.0  
**Last Updated**: February 19, 2026  
**Status**: Ready for Approval
ementation plan provides a clear roadmap to close all identified gaps in the Liyali Gateway system. By following this phased approach, the system will achieve:

- **100% production readiness** after Phase 1
- **Enterprise-grade features** after Phase 2
- **Optimized performance** after Phase 3
- **Market-leading capabilities** after Phase 4

**Recommended Next Steps**:
1. Review and approve this plan
2. Allocate resources for Phase 1
3. Begin Phase 1 implementation immediately
4. Schedule weekly progress revie

## DEPLOYMENT STRATEGY

### Phase 1 Deployment:
- Deploy to staging environment
- Run smoke tests
- Deploy to production during low-traffic window
- Monitor for 24 hours
- Rollback plan ready

### Phase 2 Deployment:
- Feature flags for new features
- Gradual rollout (10% → 50% → 100%)
- Monitor error rates
- Rollback individual features if needed

### Phase 3 Deployment:
- Continuous deployment
- Automated testing in CI/CD
- Blue-green deployments
- Zero-downtime deployments

---

## CONCLUSION

This implg

2. **Email Deliverability Issues**
   - Mitigation: Use reputable provider (SendGrid)
   - Fallback: Implement multiple email providers

3. **Performance Degradation**
   - Mitigation: Implement caching early
   - Fallback: Horizontal scaling

### Medium-Risk Items:
1. **OAuth Integration Complexity**
   - Mitigation: Use established libraries
   - Fallback: Delay OAuth, focus on email/password

2. **File Storage Costs**
   - Mitigation: Implement file size limits
   - Fallback: Use cheaper storage tier

--- attachments working (>99% upload success)
- ✅ E2E tests passing (>95% pass rate)
- ✅ Load testing completed (documented capacity)

### Phase 3 Success Criteria:
- ✅ Query performance improved (>50% faster)
- ✅ Cache hit rate >70%
- ✅ Mobile responsive (all pages)
- ✅ WCAG 2.1 AA compliant
- ✅ Third-party integrations working

---

## RISK MITIGATION

### High-Risk Items:
1. **Payment Integration Complexity**
   - Mitigation: Use Stripe's well-documented APIs
   - Fallback: Implement manual payment processintry): $1,200/year
- Database Backups: $600/year

**Total Infrastructure**: ~$7,100/year + transaction fees

---

## SUCCESS METRICS

### Phase 1 Success Criteria:
- ✅ Email delivery working (>95% delivery rate)
- ✅ Payment processing functional (0 failed transactions)
- ✅ Database backups automated (100% success rate)
- ✅ Security audit passed (0 critical vulnerabilities)

### Phase 2 Success Criteria:
- ✅ Multiple approvals working (100% accuracy)
- ✅ PDF generation functional (all document types)
- ✅ File

### Development Costs

**Phase 1**: $20,000 - $25,000
- 2-3 developers × 2 weeks × $5,000/week

**Phase 2**: $40,000 - $50,000
- 3-4 developers × 4 weeks × $5,000/week

**Phase 3**: $80,000 - $100,000
- 3-4 developers × 12 weeks × $5,000/week

**Total Development**: $140,000 - $175,000

### Infrastructure Costs (Annual)

- Email Service (SendGrid): $1,200/year
- Payment Processing (Stripe): 2.9% + $0.30 per transaction
- File Storage (AWS S3): $500/year
- Monitoring (Datadog): $3,600/year
- Error Tracking (Senure

**Phase 1 (Weeks 1-2)**:
- 1 Senior Backend Developer (email, payment, security)
- 1 Full-Stack Developer (frontend integration)
- 1 DevOps Engineer (backups, monitoring)

**Phase 2 (Weeks 3-6)**:
- 2 Senior Backend Developers (workflows, file storage)
- 1 Frontend Developer (PDF, UI)
- 1 QA Engineer (E2E tests, load testing)

**Phase 3 (Weeks 7-18)**:
- 2 Full-Stack Developers (features, optimization)
- 1 Frontend Developer (UX, accessibility)
- 1 Backend Developer (integrations)

---

## BUDGET ESTIMATExcel export functional
- ✅ Migration tools available

---

## PHASE 4: LOW PRIORITY GAPS (WEEKS 19+)

**Goal**: Implement nice-to-have features  
**Duration**: 6+ months  
**Team Size**: 2-3 developers  
**Estimated Effort**: 200-250 hours

### Features to Implement:
1. Multi-Factor Authentication (MFA)
2. Workflow Templates Library
3. User Impersonation for Support
4. Internationalization (i18n)
5. Advanced Notification Preferences
6. Frontend Performance Optimization

---

## RESOURCE ALLOCATION

### Team Structntegration
3. Add Microsoft Teams integration
4. Create Zapier integration

**Deliverables**:
- ✅ Webhook system working
- ✅ Slack notifications functional
- ✅ Teams integration working
- ✅ Zapier app published

#### Task 3.4.2: Import/Export Functionality
**Gap**: GAP-INT-002  
**Priority**: MEDIUM  
**Effort**: 32 hours

**Implementation Steps**:
1. Implement bulk data import
2. Add CSV/Excel export
3. Create data migration tools
4. Add bulk API operations

**Deliverables**:
- ✅ Bulk import working
- ✅ CSV/Ety**: MEDIUM  
**Effort**: 48 hours

**Implementation Steps**:
1. Conduct WCAG 2.1 AA audit
2. Fix accessibility issues
3. Add keyboard navigation
4. Implement screen reader support

**Deliverables**:
- ✅ WCAG 2.1 AA compliant
- ✅ Keyboard navigation working
- ✅ Screen reader compatible

### Sprint 3.4: Integration & Export (Weeks 16-18)

#### Task 3.4.1: Third-Party Integrations
**Gap**: GAP-INT-001  
**Priority**: MEDIUM  
**Effort**: 56 hours

**Implementation Steps**:
1. Implement webhook system
2. Add Slack i%
- ✅ Response times improved

### Sprint 3.3: User Experience Improvements (Weeks 13-15)

#### Task 3.3.1: Mobile Responsiveness
**Gap**: GAP-UX-001  
**Priority**: MEDIUM  
**Effort**: 40 hours

**Implementation Steps**:
1. Audit mobile responsiveness
2. Fix layout issues
3. Optimize touch interactions
4. Test on various devices

**Deliverables**:
- ✅ All pages mobile-responsive
- ✅ Touch interactions optimized
- ✅ Tested on iOS and Android

#### Task 3.3.2: Accessibility Compliance
**Gap**: GAP-UX-002  
**Priori
1. Analyze slow queries
2. Add missing indexes
3. Optimize N+1 queries
4. Implement query result caching

**Deliverables**:
- ✅ Query performance improved by 50%
- ✅ All slow queries optimized
- ✅ Indexes added

#### Task 3.2.2: Redis Caching Layer
**Gap**: GAP-PERF-002  
**Priority**: MEDIUM  
**Effort**: 32 hours

**Implementation Steps**:
1. Set up Redis
2. Implement caching service
3. Add cache invalidation
4. Cache frequently accessed data

**Deliverables**:
- ✅ Redis caching working
- ✅ Cache hit rate >70iority**: MEDIUM  
**Effort**: 32 hours

**Implementation Steps**:
1. Implement document version storage
2. Add version comparison
3. Create version history UI
4. Add version restore functionality

**Deliverables**:
- ✅ Document versioning working
- ✅ Version comparison available
- ✅ Restore functionality implemented

### Sprint 3.2: Performance Optimization (Weeks 10-12)

#### Task 3.2.1: Database Query Optimization
**Gap**: GAP-PERF-001  
**Priority**: MEDIUM  
**Effort**: 24 hours

**Implementation Steps**: 320-400 hours

### Sprint 3.1: Advanced Workflow Features (Weeks 7-9)

#### Task 3.1.1: Conditional Workflow Routing
**Gap**: GAP-WORKFLOW-002  
**Priority**: MEDIUM  
**Effort**: 40 hours

**Implementation Steps**:
1. Design rules engine
2. Implement condition evaluation
3. Add dynamic stage selection
4. Create UI for defining conditions

**Deliverables**:
- ✅ Conditional routing working
- ✅ Rules engine functional
- ✅ UI for condition management

#### Task 3.1.2: Document Versioning
**Gap**: GAP-DOC-002  
**Pr(50),
     provider_user_id VARCHAR(255),
     access_token TEXT,
     refresh_token TEXT,
     expires_at TIMESTAMP,
     created_at TIMESTAMP,
     updated_at TIMESTAMP
   );
   ```

**Deliverables**:
- ✅ OAuth login working for Google, Microsoft, GitHub
- ✅ Account linking functional
- ✅ OAuth token refresh working

---

## PHASE 3: MEDIUM PRIORITY GAPS (WEEKS 7-18)

**Goal**: Implement medium-priority features and optimizations  
**Duration**: 12 weeks  
**Team Size**: 3-4 developers  
**Estimated Effort**:- Google OAuth
   - Microsoft OAuth
   - GitHub OAuth

3. Backend Implementation:
   ```go
   // backend/handlers/oauth_handler.go
   - HandleOAuthLogin(provider)
   - HandleOAuthCallback(provider)
   - LinkOAuthAccount(userID, provider)
   ```

4. Frontend Implementation:
   - Add OAuth login buttons
   - Create OAuth callback page
   - Add account linking UI

5. Database Schema:
   ```sql
   CREATE TABLE oauth_accounts (
     id UUID PRIMARY KEY,
     user_id UUID REFERENCES users(id),
     provider VARCHAR>80%)
   - Database connection errors
   - Failed deployments

**Deliverables**:
- ✅ Error tracking working
- ✅ Application monitoring active
- ✅ Alerts configured
- ✅ Dashboards created

### Sprint 2.4: Authentication Enhancements (Week 6)

#### Task 2.4.1: OAuth/SSO Integration
**Gap**: GAP-AUTH-004  
**Priority**: HIGH  
**Effort**: 28 hours

**Implementation Steps**:
1. Install OAuth libraries:
   ```bash
   go get golang.org/x/oauth2
   go get golang.org/x/oauth2/google
   ```

2. Implement OAuth providers:
   ```bash
   npm install @sentry/nextjs
   npm install @sentry/node
   ```

2. Configure Sentry:
   - Add Sentry DSN to environment
   - Configure error sampling
   - Set up release tracking
   - Add user context

3. Set up application monitoring (Datadog/New Relic):
   - Install monitoring agent
   - Configure metrics collection
   - Set up custom metrics
   - Create dashboards

4. Configure alerts:
   - High error rate (>1%)
   - Slow response times (>500ms)
   - High CPU usage (>80%)
   - High memory usage (ent users
   - 100 concurrent users
   - 500 concurrent users
   - 1000 concurrent users

4. Document results:
   - Response times at each load level
   - Error rates
   - Throughput (requests/second)
   - Resource utilization

**Deliverables**:
- ✅ Load testing framework set up
- ✅ Baseline performance documented
- ✅ Bottlenecks identified

#### Task 2.3.3: Monitoring & Alerting
**Gap**: GAP-DEPLOY-003  
**Priority**: HIGH  
**Effort**: 20 hours

**Implementation Steps**:
1. Set up Sentry for error tracking:
   amework set up
- ✅ Critical flows covered
- ✅ Tests running in CI/CD

#### Task 2.3.2: Load Testing
**Gap**: GAP-TEST-002  
**Priority**: HIGH  
**Effort**: 16 hours

**Implementation Steps**:
1. Set up k6:
   ```bash
   brew install k6  # or download from k6.io
   ```

2. Create load test scenarios:
   ```javascript
   // backend/tests/load/
   - auth-load-test.js (login, token refresh)
   - api-load-test.js (CRUD operations)
   - workflow-load-test.js (approval flows)
   ```

3. Run baseline tests:
   - 10 concurr   ```

2. Create E2E test suite:
   ```typescript
   // frontend/e2e/
   - auth.spec.ts (login, register, logout)
   - requisitions.spec.ts (create, approve, reject)
   - budgets.spec.ts (create, submit, approve)
   - workflows.spec.ts (workflow execution)
   - subscriptions.spec.ts (upgrade, downgrade)
   ```

3. Add to CI/CD:
   ```yaml
   # .github/workflows/e2e-tests.yml
   - Run E2E tests on PR
   - Run E2E tests before deployment
   - Generate test reports
   ```

**Deliverables**:
- ✅ E2E testing frmAV)
   - Add file type restrictions
   - Add file size limits
   - Generate signed URLs for downloads

**Deliverables**:
- ✅ File upload working
- ✅ Files stored in S3
- ✅ File download functional
- ✅ Virus scanning implemented

### Sprint 2.3: Testing & Monitoring (Weeks 5-6)

#### Task 2.3.1: E2E Testing Framework
**Gap**: GAP-TEST-001  
**Priority**: HIGH  
**Effort**: 24 hours

**Implementation Steps**:
1. Set up Playwright:
   ```bash
   cd frontend
   npm install -D @playwright/test
   npx playwright install
UID REFERENCES users(id),
     created_at TIMESTAMP
   );
   ```

3. Backend Implementation:
   ```go
   // backend/services/file_service.go
   - UploadFile(file, documentID, documentType)
   - GetFileURL(fileID)
   - DeleteFile(fileID)
   - ListFiles(documentID)
   ```

4. Frontend Implementation:
   - Create file upload component
   - Add drag-and-drop support
   - Show upload progress
   - Display attached files
   - Add file preview
   - Implement file deletion

5. Security:
   - Implement virus scanning (Cla**Gap**: GAP-DOC-004  
**Priority**: HIGH  
**Effort**: 32 hours

**Implementation Steps**:
1. Set up file storage (AWS S3):
   - Create S3 bucket
   - Configure CORS
   - Set up IAM policies
   - Generate access keys

2. Database Schema:
   ```sql
   CREATE TABLE document_attachments (
     id UUID PRIMARY KEY,
     document_id UUID,
     document_type VARCHAR(50),
     file_name VARCHAR(255),
     file_size BIGINT,
     file_type VARCHAR(100),
     s3_key VARCHAR(500),
     s3_url TEXT,
     uploaded_by UDF
   - Requisition PDF
   - Budget PDF
   - GRN PDF

3. Add PDF generation service:
   ```typescript
   // frontend/src/lib/pdf-service.ts
   - generatePDF(document, type)
   - downloadPDF(document, type)
   - emailPDF(document, type, recipients)
   ```

4. Add PDF preview:
   - Create PDF preview modal
   - Add download button
   - Add email button

**Deliverables**:
- ✅ PDF generation working for all document types
- ✅ PDF download functional
- ✅ PDF preview available

#### Task 2.2.2: File Attachment System
omparison
   - Add rollback button
   - Show version notes

**Deliverables**:
- ✅ Workflow versioning working
- ✅ Version history accessible
- ✅ Rollback functional

### Sprint 2.2: Document Management (Weeks 3-4)

#### Task 2.2.1: PDF Generation Completion
**Gap**: GAP-DOC-001  
**Priority**: HIGH  
**Effort**: 20 hours

**Implementation Steps**:
1. Install dependencies:
   ```bash
   cd frontend
   npm install @react-pdf/renderer
   ```

2. Implement PDF templates:
   - Purchase Order PDF
   - Payment Voucher PLEAN DEFAULT TRUE;
   
   CREATE TABLE workflow_versions (
     id UUID PRIMARY KEY,
     workflow_id UUID REFERENCES workflows(id),
     version INT,
     configuration JSONB,
     created_by UUID REFERENCES users(id),
     created_at TIMESTAMP,
     notes TEXT
   );
   ```

2. Backend Implementation:
   - Create new version on workflow update
   - Store previous version
   - Implement version comparison
   - Add rollback functionality

3. Frontend Implementation:
   - Show version history
   - Display version cg:
   - Test various approval scenarios
   - Test concurrent approvals
   - Test approval threshold logic

**Deliverables**:
- ✅ Multiple approvals supported
- ✅ Approval progress tracked
- ✅ UI shows approval status
- ✅ Tests passing

#### Task 2.1.2: Workflow Version Control
**Gap**: GAP-WORKFLOW-004  
**Priority**: HIGH  
**Effort**: 24 hours

**Implementation Steps**:
1. Database Schema:
   ```sql
   ALTER TABLE workflows ADD COLUMN version INT DEFAULT 1;
   ALTER TABLE workflows ADD COLUMN is_active BOOES workflow_tasks(id),
     user_id UUID REFERENCES users(id),
     approved BOOLEAN,
     comments TEXT,
     created_at TIMESTAMP
   );
   ```

2. Backend Implementation:
   - Update workflow execution logic
   - Track individual approvals
   - Check if required approvals met
   - Update task status when threshold reached

3. Frontend Implementation:
   - Show approval progress (e.g., "2/3 approved")
   - Display list of approvers
   - Show who has approved/pending
   - Update UI for partial approvals

4. Testinffort**: 160-200 hours

### Sprint 2.1: Workflow Enhancements (Weeks 3-4)

#### Task 2.1.1: Multiple Approvals Per Stage
**Gap**: GAP-WORKFLOW-001  
**Priority**: HIGH  
**Effort**: 32 hours

**Implementation Steps**:
1. Database Schema Changes:
   ```sql
   ALTER TABLE workflow_stages 
   ADD COLUMN approvals_required INT DEFAULT 1;
   
   ALTER TABLE workflow_tasks
   ADD COLUMN approval_count INT DEFAULT 0;
   
   CREATE TABLE workflow_task_approvals (
     id UUID PRIMARY KEY,
     task_id UUID REFERENChanges
   - Allow consent withdrawal

4. Privacy Policy & Terms:
   - Create privacy policy page
   - Create terms of service page
   - Add cookie consent banner
   - Implement cookie preferences

**Deliverables**:
- ✅ Data export working
- ✅ Data deletion implemented
- ✅ Consent management functional
- ✅ Privacy policy published

---

## PHASE 2: HIGH PRIORITY GAPS (WEEKS 3-6)

**Goal**: Implement high-priority features and improvements  
**Duration**: 4 weeks  
**Team Size**: 3-4 developers  
**Estimated E  // backend/handlers/gdpr_handler.go
   - Implement /api/v1/users/me/export endpoint
   - Export all user data in JSON format
   - Include all related documents
   - Add download link
   ```

2. Data Deletion (Right to be Forgotten):
   ```go
   - Implement /api/v1/users/me/delete endpoint
   - Anonymize user data
   - Delete personal information
   - Retain audit logs (anonymized)
   ```

3. Consent Management:
   - Add privacy policy acceptance tracking
   - Create consent management UI
   - Track consent ces
   - Test CSRF protection

4. Security Hardening:
   - Add security headers
   - Implement CSP policy
   - Add rate limiting
   - Enable HTTPS only
   - Implement request signing

**Deliverables**:
- ✅ Security audit report
- ✅ All critical vulnerabilities fixed
- ✅ Security hardening implemented
- ✅ Penetration test passed

#### Task 1.3.4: GDPR Compliance Implementation
**Gap**: GAP-COMP-001  
**Priority**: CRITICAL  
**Effort**: 12 hours

**Implementation Steps**:
1. Data Export Functionality:
   ```go
 
**Implementation Steps**:
1. Automated Security Scanning:
   - Run OWASP ZAP scan
   - Run Snyk vulnerability scan
   - Run npm audit / go mod audit
   - Fix critical vulnerabilities

2. Manual Security Review:
   - Review authentication flows
   - Check authorization logic
   - Audit SQL queries for injection
   - Review file upload security
   - Check CORS configuration

3. Penetration Testing:
   - Test authentication bypass
   - Test authorization bypass
   - Test SQL injection
   - Test XSS vulnerabilitie required variables on startup
   - Check variable formats
   - Fail fast if misconfigured
   ```

4. Create secrets management guide:
   - Document how to use Fly.io secrets
   - Add rotation procedures
   - Document access controls

**Deliverables**:
- ✅ Complete environment documentation
- ✅ Environment templates created
- ✅ Validation on startup
- ✅ Secrets management documented

#### Task 1.3.3: Security Audit & Penetration Testing
**Gap**: GAP-TEST-003  
**Priority**: CRITICAL  
**Effort**: 16 hours
ation Documentation
**Gap**: GAP-DEPLOY-001  
**Priority**: CRITICAL  
**Effort**: 6 hours

**Implementation Steps**:
1. Create environment variable documentation:
   ```markdown
   # PRODUCTION_ENV_VARS.md
   - List all required variables
   - Document default values
   - Add security notes
   - Include example values
   ```

2. Create environment templates:
   - `.env.production.template`
   - `.env.staging.template`

3. Add environment validation:
   ```go
   // backend/config/env_validator.go
   - Validatackup
   - Upload to S3
   - Verify backup integrity
   - Clean old backups
   ```

3. Set up backup monitoring:
   - Alert on backup failures
   - Track backup size trends
   - Monitor backup duration

4. Document restoration process:
   - Create step-by-step restoration guide
   - Test restoration procedure
   - Document RTO and RPO

**Deliverables**:
- ✅ Daily automated backups
- ✅ 30-day retention policy
- ✅ Backup monitoring alerts
- ✅ Tested restoration procedure

#### Task 1.3.2: Environment Configur3: Production Readiness (Week 2)

#### Task 1.3.1: Database Backup Automation
**Gap**: GAP-DEPLOY-002  
**Priority**: CRITICAL  
**Effort**: 8 hours

**Implementation Steps**:
1. Set up automated backups:
   ```bash
   # For Fly.io PostgreSQL
   fly postgres backup create
   
   # For AWS RDS
   - Enable automated backups
   - Set retention period to 30 days
   - Configure backup window
   ```

2. Create backup script:
   ```bash
   #!/bin/bash
   # scripts/backup-database.sh
   - Dump database
   - Compress becurely
- Subscription status synced automatically
- Failed payments handled gracefully

#### Task 1.2.2: Subscription Billing Cycle Management
**Gap**: GAP-SUB-001  
**Priority**: CRITICAL  
**Effort**: 12 hours

**Implementation Steps**:
1. Implement billing cycle tracking
2. Add prorated billing for mid-cycle upgrades
3. Create billing history page
4. Add upcoming invoice preview

**Deliverables**:
- ✅ Billing cycles tracked accurately
- ✅ Prorated billing working
- ✅ Billing history accessible

### Sprint 1. - Create payment method form
   - Add payment confirmation page
   - Implement invoice history page

5. Webhook Handlers:
   - `checkout.session.completed`
   - `invoice.paid`
   - `invoice.payment_failed`
   - `customer.subscription.updated`
   - `customer.subscription.deleted`

**Deliverables**:
- ✅ Stripe fully integrated
- ✅ Payment processing working
- ✅ Webhooks handling subscription events
- ✅ Invoice generation working

**Acceptance Criteria**:
- Users can upgrade plans successfully
- Payments processed s(4),
     exp_month INT,
     exp_year INT,
     is_default BOOLEAN,
     created_at TIMESTAMP,
     updated_at TIMESTAMP
   );

   CREATE TABLE invoices (
     id UUID PRIMARY KEY,
     organization_id UUID REFERENCES organizations(id),
     stripe_invoice_id VARCHAR(255),
     amount_due DECIMAL(10,2),
     amount_paid DECIMAL(10,2),
     status VARCHAR(50),
     invoice_pdf_url TEXT,
     created_at TIMESTAMP,
     paid_at TIMESTAMP
   );
   ```

4. Frontend Implementation:
   - Integrate Stripe Elements
   products and prices
   - Generate API keys
   - Configure webhooks

2. Backend Implementation:
   ```go
   // backend/services/payment_service.go
   - Create PaymentService
   - Implement CreateCheckoutSession()
   - Implement HandleWebhook()
   - Add subscription status sync
   ```

3. Database Schema:
   ```sql
   CREATE TABLE payment_methods (
     id UUID PRIMARY KEY,
     organization_id UUID REFERENCES organizations(id),
     stripe_payment_method_id VARCHAR(255),
     type VARCHAR(50),
     last4 VARCHARil sending
2. Add rate limiting (3 requests per hour per email)
3. Create password reset confirmation email
4. Add password reset success page

**Deliverables**:
- ✅ Password reset emails sent automatically
- ✅ Rate limiting prevents abuse
- ✅ Confirmation emails sent

### Sprint 1.2: Payment & Subscription (Week 2)

#### Task 1.2.1: Stripe Payment Integration
**Gap**: GAP-SUB-001  
**Priority**: CRITICAL  
**Effort**: 24 hours

**Implementation Steps**:
1. Stripe Setup:
   - Create Stripe account
   - Set upge
   - Add "Resend verification email" button
   - Show verification status in user profile

4. Testing:
   - Test token generation and expiration
   - Test verification flow
   - Test resend functionality

**Deliverables**:
- ✅ Email verification fully functional
- ✅ Users must verify email before full access
- ✅ Resend verification working

#### Task 1.1.3: Password Reset Email Flow
**Gap**: GAP-AUTH-003  
**Priority**: CRITICAL  
**Effort**: 8 hours

**Implementation Steps**:
1. Complete password reset ema