# Liyali Gateway API - Postman Collection

Complete Postman collection for testing the Liyali Gateway Backend API with all endpoints organized by functionality.

## 📁 Collection Structure

The collection is organized into the following folders:

### 🔐 Authentication
- User registration and login
- Token management (refresh, verify)
- Password reset functionality
- Profile management
- Session management

### 🏢 Organizations
- Organization CRUD operations
- Multi-tenancy support
- Member management
- Organization settings
- Organization switching

### 👥 Roles & Permissions
- Role-based access control (RBAC)
- Custom role creation
- Permission management
- Role assignment

### 📋 Requisitions
- Purchase requisition lifecycle
- Approval workflows
- Item management
- Status tracking

### 💰 Budgets
- Budget allocation and tracking
- Category-based budgeting
- Fiscal year management
- Budget approval workflows

### 🛒 Purchase Orders
- PO creation from requisitions
- Vendor management integration
- Delivery tracking
- Terms and conditions

### 🔍 Document Search & Management
- Cross-document search
- Generic document operations
- Document statistics
- Approval history

### ⚡ Workflows
- Workflow template management
- Dynamic approval processes
- Stage configuration
- Workflow execution

### ✅ Approvals
- Approval task management
- Bulk operations
- Overdue task tracking
- Delegation and reassignment

### 📊 Analytics
- Dashboard metrics
- Requisition analytics
- Approval metrics
- Custom reporting

### 🔔 Notifications
- User notifications
- Read/unread management
- Notification statistics
- Bulk operations

### 🏥 Health & Monitoring
- System health checks
- API status monitoring

## 🚀 Quick Start

### 1. Import Collection

1. Download the collection file: `liyali-gateway-complete-postman-collection.json`
2. Open Postman
3. Click "Import" button
4. Select the downloaded JSON file
5. The collection will be imported with all folders and requests

### 2. Import Environment

Choose the appropriate environment:

**Development Environment:**
- File: `liyali-gateway-postman-environment.json`
- Base URL: `http://localhost:8080`

**Production Environment:**
- File: `liyali-gateway-postman-production-environment.json`
- Base URL: `https://api.liyali.com`

### 3. Authentication Setup

1. Navigate to the "🔐 Authentication" folder
2. Run the "Login" request with valid credentials
3. The auth token will be automatically saved to environment variables
4. All subsequent requests will use this token automatically

### 4. Test Workflow

Follow this recommended testing sequence:

1. **Login** - Get authentication token
2. **Create Organization** - Set up organization context
3. **Create Categories** - Set up item categories
4. **Create Vendors** - Set up vendor information
5. **Create Requisition** - Create a purchase request
6. **Test Approvals** - Test approval workflows
7. **Create Purchase Order** - Convert approved requisition
8. **Test Analytics** - View metrics and reports

## 🔧 Environment Variables

The collection uses the following environment variables:

### Authentication
- `auth_token` - JWT access token (auto-populated)
- `refresh_token` - JWT refresh token (auto-populated)
- `user_id` - Current user ID (auto-populated)
- `organization_id` - Current organization ID (auto-populated)

### Test Data IDs
- `requisition_id` - Sample requisition ID
- `budget_id` - Sample budget ID
- `purchase_order_id` - Sample purchase order ID
- `document_id` - Sample document ID
- `workflow_id` - Sample workflow ID
- `approval_id` - Sample approval ID
- `category_id` - Sample category ID
- `vendor_id` - Sample vendor ID

### Configuration
- `base_url` - API base URL
- `admin_email` - Default admin email
- `admin_password` - Default admin password

## 📝 Request Features

### Automatic Token Management
- Tokens are automatically extracted from login responses
- All protected endpoints use the stored token
- Token refresh is handled automatically

### Response Validation
- Status code validation
- Response structure validation
- Data extraction for subsequent requests

### Correlation ID Tracking
- Each request includes a unique correlation ID
- Enables request tracing across the system
- Useful for debugging and monitoring

### Dynamic Data Generation
- Uses Postman's dynamic variables for realistic test data
- Generates random emails, names, and other data
- Reduces test data maintenance

## 🧪 Testing Features

### Pre-request Scripts
- Automatic correlation ID generation
- Token validation and refresh
- Dynamic data preparation

### Test Scripts
- Response validation
- Data extraction for chaining requests
- Environment variable updates
- Error handling

### Example Test Flow

```javascript
// Login and save token
pm.test('Login successful', function () {
    const jsonData = pm.response.json();
    pm.expect(jsonData.success).to.be.true;
    pm.environment.set('auth_token', jsonData.data.tokens.accessToken);
});

// Create requisition and save ID
pm.test('Requisition created', function () {
    const jsonData = pm.response.json();
    pm.expect(jsonData.success).to.be.true;
    pm.environment.set('requisition_id', jsonData.data.id);
});
```

## 🔍 Advanced Usage

### Bulk Operations
The collection includes bulk operation examples:
- Bulk approve multiple requisitions
- Bulk reject approval tasks
- Bulk reassign approvals

### Search and Filtering
Comprehensive search examples with:
- Full-text search across documents
- Advanced filtering options
- Pagination handling
- Sort and order options

### Workflow Testing
Complete workflow testing scenarios:
- Create custom workflows
- Test approval stages
- Handle conditional approvals
- Test escalation scenarios

## 🐛 Troubleshooting

### Common Issues

**Authentication Errors:**
- Ensure you've run the Login request first
- Check that the auth_token variable is populated
- Verify the token hasn't expired

**Environment Variables:**
- Make sure the correct environment is selected
- Check that base_url points to the running server
- Verify all required variables are set

**Request Failures:**
- Check server is running on the specified port
- Verify database is connected and migrated
- Check server logs for detailed error messages

### Debug Tips

1. **Enable Console Logging:**
   ```javascript
   console.log('Auth token:', pm.environment.get('auth_token'));
   ```

2. **Check Response Data:**
   ```javascript
   console.log('Response:', pm.response.json());
   ```

3. **Validate Environment:**
   ```javascript
   console.log('Base URL:', pm.environment.get('base_url'));
   ```

## 📚 Additional Resources

- **API Documentation**: `backend/docs/13-api-reference.md`
- **Development Guide**: `backend/docs/11-development.md`
- **Testing Guide**: `backend/docs/12-testing.md`
- **Deployment Guide**: `backend/docs/14-deployment.md`

## 🤝 Contributing

When adding new endpoints to the API:

1. Add the corresponding Postman request to the appropriate folder
2. Include proper test scripts for validation
3. Update environment variables if needed
4. Add documentation for any new features
5. Test the request in both development and production environments

## 📄 License

This collection is part of the Liyali Gateway project and follows the same license terms.