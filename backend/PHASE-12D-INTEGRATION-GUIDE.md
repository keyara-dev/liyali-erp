# Phase 12D Integration Guide
## How to Integrate Services into Phase 12C Handlers

---

## Quick Overview

Phase 12D provides 5 business logic services that must be integrated into the Phase 12C handlers. This guide shows exactly how to integrate each service.

---

## Step 1: Initialize Services in Database Configuration

**File**: `backend/config/database.go`

Add to the `InitDB()` function after database connection:

```go
// Initialize Phase 12D Services
approvalRoutingService := services.NewApprovalRoutingService(db)
workflowStateMachine := services.NewWorkflowStateMachine(db)
budgetValidationService := services.NewBudgetValidationService(db)
documentLinkingService := services.NewDocumentLinkingService(db)
notificationService := services.NewNotificationService(db)

// Create default rules and constraints
if err := approvalRoutingService.CreateDefaultApprovalRules(); err != nil {
    log.Printf("Error creating default approval rules: %v", err)
}

if err := budgetValidationService.CreateDefaultBudgetConstraints(); err != nil {
    log.Printf("Error creating default budget constraints: %v", err)
}

log.Println("Phase 12D business logic services initialized")
```

---

## Step 2: Add Service Globals in Handlers

**File**: `backend/handlers/handlers.go`

Add package-level variables at the top:

```go
package handlers

import (
    // ... existing imports ...
    "github.com/liyali/liyali-gateway/services"
)

// Phase 12D Services (initialized in main.go)
var (
    ApprovalRoutingService  *services.ApprovalRoutingService
    WorkflowStateMachine    *services.WorkflowStateMachine
    BudgetValidationService *services.BudgetValidationService
    DocumentLinkingService  *services.DocumentLinkingService
    NotificationService     *services.NotificationService
)

// InitializeServices sets up business logic services
func InitializeServices(
    ars *services.ApprovalRoutingService,
    wsm *services.WorkflowStateMachine,
    bvs *services.BudgetValidationService,
    dls *services.DocumentLinkingService,
    ns *services.NotificationService,
) {
    ApprovalRoutingService = ars
    WorkflowStateMachine = wsm
    BudgetValidationService = bvs
    DocumentLinkingService = dls
    NotificationService = ns
}
```

**File**: `backend/main.go`

Call initialization in main():

```go
// Initialize services
handlers.InitializeServices(
    approvalRoutingService,
    workflowStateMachine,
    budgetValidationService,
    documentLinkingService,
    notificationService,
)
```

---

## Step 3: Integrate Approval Routing - Create Endpoints

### Requisition Creation with Budget Validation

**File**: `backend/handlers/requisition.go`

Modify `CreateRequisition()` function:

```go
func CreateRequisition(c fiber.Ctx) error {
    var req types.CreateRequisitionRequest
    if err := c.BodyParser(&req); err != nil {
        return responses.SendValidationError(c, "Invalid request body", "")
    }

    // Validate budget first
    valid, msg, err := BudgetValidationService.ValidateBudgetForRequisition(
        req.Department,
        "2025", // TODO: Get fiscal year from context or config
        req.TotalAmount,
    )
    if err != nil {
        return responses.SendInternalError(c, "Budget validation error", err.Error())
    }
    if !valid {
        return responses.SendUnprocessableEntityError(c, msg, "")
    }

    // Create requisition (existing code)
    requisition := models.Requisition{
        ID:              uuid.New().String(),
        RequesterID:     userID,
        Title:           req.Title,
        Description:     req.Description,
        Department:      req.Department,
        Status:          "draft",
        Priority:        req.Priority,
        Items:           itemsJSON,
        TotalAmount:     req.TotalAmount,
        Currency:        req.Currency,
        ApprovalStage:   0,
        CreatedAt:       time.Now(),
        UpdatedAt:       time.Now(),
    }

    if err := db.Create(&requisition).Error; err != nil {
        return responses.SendInternalError(c, "Failed to create requisition", err.Error())
    }

    // Link to budget if one was found
    budgets, _ := BudgetValidationService.GetBudgetsByDepartment(req.Department)
    if len(budgets) > 0 {
        _ = DocumentLinkingService.LinkRequisitionToBudget(
            requisition.ID,
            budgets[0].ID,
            req.TotalAmount,
        )
    }

    return responses.SendSuccess(c, 201, "Requisition created successfully", requisition, nil)
}
```

---

## Step 4: Integrate Workflow State Machine - Approval Endpoints

### Submit for Approval

**File**: `backend/handlers/requisition.go`

Add new endpoint for submission:

```go
func SubmitRequisitionForApproval(c fiber.Ctx) error {
    requisitionID := c.Params("id")
    userID := c.Locals("userId").(string)

    var req models.Requisition
    if err := db.First(&req, "id = ?", requisitionID).Error; err != nil {
        return responses.SendNotFoundError(c, "Requisition not found", "")
    }

    // Validate state transition
    if !WorkflowStateMachine.CanTransition(
        "requisition",
        req.Status,
        "pending",
        "", // No specific role required to submit
    ) {
        return responses.SendUnprocessableEntityError(
            c,
            "Cannot submit from current state",
            fmt.Sprintf("Current state: %s", req.Status),
        )
    }

    // Perform transition
    if err := WorkflowStateMachine.SubmitForApproval("requisition", requisitionID, userID); err != nil {
        return responses.SendInternalError(c, "Failed to submit for approval", err.Error())
    }

    // Route to approvers
    if err := ApprovalRoutingService.RouteDocumentToApprovers(
        requisitionID,
        "requisition",
        req.Department,
        req.TotalAmount,
        req.Priority,
    ); err != nil {
        log.Printf("Warning: Failed to route document: %v", err)
    }

    // Fetch updated requisition
    db.First(&req, "id = ?", requisitionID)

    return responses.SendSuccess(c, 200, "Requisition submitted for approval", req, nil)
}
```

### Approve Requisition

**File**: `backend/handlers/requisition.go`

Modify `ApproveRequisition()` function:

```go
func ApproveRequisition(c fiber.Ctx) error {
    requisitionID := c.Params("id")
    userID := c.Locals("userId").(string)
    userRole := c.Locals("userRole").(string)

    var req types.ApproveDocumentRequest
    if err := c.BodyParser(&req); err != nil {
        return responses.SendValidationError(c, "Invalid request body", "")
    }

    var requisition models.Requisition
    if err := db.First(&requisition, "id = ?", requisitionID).Error; err != nil {
        return responses.SendNotFoundError(c, "Requisition not found", "")
    }

    // Validate state transition
    if !WorkflowStateMachine.CanTransition(
        "requisition",
        requisition.Status,
        "approved",
        userRole,
    ) {
        return responses.SendForbiddenError(c, "Cannot approve with current role", "")
    }

    // Perform transition
    if err := WorkflowStateMachine.ApproveDocument(
        "requisition",
        requisitionID,
        userID,
        userRole,
        req.Comments,
    ); err != nil {
        return responses.SendInternalError(c, "Failed to approve", err.Error())
    }

    // Update approval task status
    _ = db.Model(&models.ApprovalTask{}).
        Where("document_id = ? AND approver_id = ? AND status = ?",
            requisitionID, userID, "pending").
        Update("status", "approved")

    // Trigger approval notification
    NotificationService.HandleWorkflowEvent(services.NotificationEvent{
        Type:         "document_approved",
        DocumentID:   requisitionID,
        DocumentType: "requisition",
        Action:       "approve",
        ActorID:      userID,
        Timestamp:    time.Now(),
    })

    // Fetch updated requisition
    db.First(&requisition, "id = ?", requisitionID)

    return responses.SendSuccess(c, 200, "Requisition approved successfully", requisition, nil)
}
```

### Reject Requisition

**File**: `backend/handlers/requisition.go`

Modify `RejectRequisition()` function:

```go
func RejectRequisition(c fiber.Ctx) error {
    requisitionID := c.Params("id")
    userID := c.Locals("userId").(string)
    userRole := c.Locals("userRole").(string)

    var req types.RejectDocumentRequest
    if err := c.BodyParser(&req); err != nil {
        return responses.SendValidationError(c, "Invalid request body", "")
    }

    var requisition models.Requisition
    if err := db.First(&requisition, "id = ?", requisitionID).Error; err != nil {
        return responses.SendNotFoundError(c, "Requisition not found", "")
    }

    // Validate state transition
    if !WorkflowStateMachine.CanTransition(
        "requisition",
        requisition.Status,
        "rejected",
        userRole,
    ) {
        return responses.SendForbiddenError(c, "Cannot reject with current role", "")
    }

    // Perform transition
    if err := WorkflowStateMachine.RejectDocument(
        "requisition",
        requisitionID,
        userID,
        userRole,
        req.Remarks,
    ); err != nil {
        return responses.SendInternalError(c, "Failed to reject", err.Error())
    }

    // Update approval task status
    _ = db.Model(&models.ApprovalTask{}).
        Where("document_id = ? AND status = ?", requisitionID, "pending").
        Update("status", "rejected")

    // Deallocate budget if it was allocated
    linkedDocs, _ := DocumentLinkingService.GetLinkedDocuments(requisitionID, "requisition")
    for _, link := range linkedDocs {
        if link.LinkType == "allocates_to" {
            _ = BudgetValidationService.DeallocateBudget(
                link.SourceDocID,
                link.Amount,
                requisitionID,
            )
        }
    }

    // Trigger rejection notification
    NotificationService.HandleWorkflowEvent(services.NotificationEvent{
        Type:         "document_rejected",
        DocumentID:   requisitionID,
        DocumentType: "requisition",
        Action:       "reject",
        ActorID:      userID,
        Details:      req.Remarks,
        Timestamp:    time.Now(),
    })

    // Fetch updated requisition
    db.First(&requisition, "id = ?", requisitionID)

    return responses.SendSuccess(c, 200, "Requisition rejected successfully", requisition, nil)
}
```

---

## Step 5: Integrate Document Linking - Purchase Order Creation

**File**: `backend/handlers/purchase_order.go`

Modify `CreatePurchaseOrder()` function:

```go
func CreatePurchaseOrder(c fiber.Ctx) error {
    var req types.CreatePurchaseOrderRequest
    if err := c.BodyParser(&req); err != nil {
        return responses.SendValidationError(c, "Invalid request body", "")
    }

    // Validate budget for PO
    valid, msg, err := BudgetValidationService.ValidateBudgetForPurchaseOrder(
        "IT", // TODO: Get from linked requisition or request
        "2025",
        req.TotalAmount,
        req.VendorID,
    )
    if err != nil {
        return responses.SendInternalError(c, "Budget validation error", err.Error())
    }
    if !valid {
        return responses.SendUnprocessableEntityError(c, msg, "")
    }

    // Create PO (existing code)
    po := models.PurchaseOrder{
        ID:                uuid.New().String(),
        PONumber:          "PO-" + time.Now().Format("20060102150405") + "-" + uuid.New().String()[:8],
        VendorID:          req.VendorID,
        Status:            "draft",
        Items:             itemsJSON,
        TotalAmount:       req.TotalAmount,
        Currency:          req.Currency,
        DeliveryDate:      req.DeliveryDate,
        ApprovalStage:     0,
        LinkedRequisition: req.LinkedRequisition, // If provided
        CreatedAt:         time.Now(),
        UpdatedAt:         time.Now(),
    }

    if err := db.Create(&po).Error; err != nil {
        return responses.SendInternalError(c, "Failed to create PO", err.Error())
    }

    // Link to requisition if provided
    if req.LinkedRequisition != "" {
        if err := DocumentLinkingService.LinkRequisitionToPurchaseOrder(
            req.LinkedRequisition,
            po.ID,
        ); err != nil {
            log.Printf("Warning: Failed to link requisition to PO: %v", err)
        }
    }

    // Build response with linked documents
    linkedDocs, _ := DocumentLinkingService.GetLinkedDocuments(po.ID, "po")
    responseData := map[string]interface{}{
        "po":            po,
        "linkedDocuments": linkedDocs,
    }

    return responses.SendSuccess(c, 201, "Purchase order created successfully", responseData, nil)
}
```

---

## Step 6: Integrate Document Linking - GRN Creation

**File**: `backend/handlers/grn.go`

Modify `CreateGRN()` function:

```go
func CreateGRN(c fiber.Ctx) error {
    var req types.CreateGRNRequest
    if err := c.BodyParser(&req); err != nil {
        return responses.SendValidationError(c, "Invalid request body", "")
    }

    // Create GRN (existing code)
    grn := models.GoodsReceivedNote{
        ID:              uuid.New().String(),
        GRNNumber:       "GRN-" + time.Now().Format("20060102150405") + "-" + uuid.New().String()[:8],
        PONumber:        req.PONumber,
        Status:          "draft",
        ReceivedDate:    req.ReceivedDate,
        ReceivedBy:      req.ReceivedBy,
        Items:           itemsJSON,
        QualityIssues:   qualityIssuesJSON,
        ApprovalStage:   0,
        CreatedAt:       time.Now(),
        UpdatedAt:       time.Now(),
    }

    if err := db.Create(&grn).Error; err != nil {
        return responses.SendInternalError(c, "Failed to create GRN", err.Error())
    }

    // Link to PO
    if err := DocumentLinkingService.LinkPurchaseOrderToGRN(
        req.PONumber,
        grn.ID,
    ); err != nil {
        log.Printf("Warning: Failed to link PO to GRN: %v", err)
    }

    // Get full relationship chain
    chain, _ := DocumentLinkingService.GetDocumentRelationshipChain(req.PONumber)

    // Build response
    responseData := map[string]interface{}{
        "grn":                    grn,
        "documentRelationshipChain": chain,
    }

    return responses.SendSuccess(c, 201, "GRN created successfully", responseData, nil)
}
```

---

## Step 7: Add New Endpoints for Document Relationships

**File**: `backend/routes/routes.go`

Add new routes in SetupRoutes():

```go
// Document linking and workflow endpoints
protected.Get("/documents/:id/relationships", handlers.GetDocumentRelationships)
protected.Get("/documents/:id/audit-trail", handlers.GetDocumentAuditTrail)
protected.Get("/notifications/pending", handlers.GetPendingNotifications)
protected.Put("/notifications/:id/read", handlers.MarkNotificationAsRead)
protected.Get("/budget/:id/status", handlers.GetBudgetStatus)
```

---

## Step 8: Add Handler Functions for New Endpoints

**File**: `backend/handlers/requisition.go`

Add helper handlers:

```go
func GetDocumentRelationships(c fiber.Ctx) error {
    docID := c.Params("id")
    docType := c.Query("type", "requisition")

    links, err := DocumentLinkingService.GetLinkedDocuments(docID, docType)
    if err != nil {
        return responses.SendInternalError(c, "Failed to get relationships", err.Error())
    }

    return responses.SendSuccess(c, 200, "Document relationships retrieved", links, nil)
}

func GetDocumentAuditTrail(c fiber.Ctx) error {
    docID := c.Params("id")

    logs, err := WorkflowStateMachine.GetWorkflowHistory(docID)
    if err != nil {
        return responses.SendInternalError(c, "Failed to get audit trail", err.Error())
    }

    return responses.SendSuccess(c, 200, "Audit trail retrieved", logs, nil)
}

func GetPendingNotifications(c fiber.Ctx) error {
    userID := c.Locals("userId").(string)

    notifs, err := NotificationService.GetPendingNotifications(userID)
    if err != nil {
        return responses.SendInternalError(c, "Failed to get notifications", err.Error())
    }

    return responses.SendSuccess(c, 200, "Pending notifications retrieved", notifs, nil)
}

func MarkNotificationAsRead(c fiber.Ctx) error {
    notifID := c.Params("id")

    if err := NotificationService.MarkAsRead(notifID); err != nil {
        return responses.SendInternalError(c, "Failed to mark as read", err.Error())
    }

    return responses.SendSuccess(c, 200, "Notification marked as read", nil, nil)
}

func GetBudgetStatus(c fiber.Ctx) error {
    budgetID := c.Params("id")

    status, err := BudgetValidationService.GetBudgetStatus(budgetID)
    if err != nil {
        return responses.SendNotFoundError(c, "Budget not found", "")
    }

    return responses.SendSuccess(c, 200, "Budget status retrieved", status, nil)
}
```

---

## Step 9: Update Response Models

**File**: `backend/types/documents.go`

Add new response types:

```go
type DocumentRelationshipResponse struct {
    DocumentID    string                 `json:"documentId"`
    DocumentType  string                 `json:"documentType"`
    Links         []DocumentLinkResponse `json:"links"`
}

type DocumentLinkResponse struct {
    ID            string  `json:"id"`
    SourceDocID   string  `json:"sourceDocId"`
    SourceDocType string  `json:"sourceDocType"`
    TargetDocID   string  `json:"targetDocId"`
    TargetDocType string  `json:"targetDocType"`
    LinkType      string  `json:"linkType"`
    Amount        float64 `json:"amount,omitempty"`
    Proportion    float64 `json:"proportion,omitempty"`
    Status        string  `json:"status"`
    CreatedAt     string  `json:"createdAt"`
}

type NotificationResponse struct {
    ID           string `json:"id"`
    RecipientID  string `json:"recipientId"`
    Type         string `json:"type"`
    DocumentID   string `json:"documentId"`
    DocumentType string `json:"documentType"`
    Subject      string `json:"subject"`
    Body         string `json:"body"`
    Sent         bool   `json:"sent"`
    SentAt       string `json:"sentAt,omitempty"`
    CreatedAt    string `json:"createdAt"`
}
```

---

## Testing Integration

### 1. Test Approval Routing
```bash
# Create requisition
curl -X POST http://localhost:8080/api/v1/requisitions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{...}'

# Check approval tasks created
SELECT * FROM approval_tasks WHERE document_id = 'req-123'

# Check notifications created
SELECT * FROM notifications WHERE document_id = 'req-123'
```

### 2. Test State Transitions
```bash
# Submit for approval
curl -X POST http://localhost:8080/api/v1/requisitions/req-123/submit \
  -H "Authorization: Bearer $TOKEN"

# Check audit log
SELECT * FROM audit_logs WHERE document_id = 'req-123'
```

### 3. Test Budget Validation
```bash
# Create requisition that exceeds budget (should fail)
curl -X POST http://localhost:8080/api/v1/requisitions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"department":"IT","totalAmount":600000,...}'

# Response: 422 - Amount exceeds remaining budget
```

### 4. Test Document Linking
```bash
# Create PO linked to requisition
curl -X POST http://localhost:8080/api/v1/purchase-orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"linkedRequisition":"req-123",...}'

# Get relationships
curl -X GET "http://localhost:8080/api/v1/documents/po-123/relationships?type=po" \
  -H "Authorization: Bearer $TOKEN"
```

---

## Troubleshooting

### Issue: Services not initialized
**Solution**: Ensure InitializeServices() is called in main.go before starting routes

### Issue: Budget validation not working
**Solution**: Check that default constraints are created via CreateDefaultBudgetConstraints()

### Issue: Notifications not created
**Solution**: Verify NotificationService.HandleWorkflowEvent() is called in approval handlers

### Issue: Links not appearing
**Solution**: Check document IDs match exactly and DocumentLinkingService is initialized

---

## Summary

**Integration Steps**:
1. ✅ Import services package
2. ✅ Initialize services in database.go
3. ✅ Add service globals in handlers
4. ✅ Integrate budget validation in create endpoints
5. ✅ Integrate state machine in approval endpoints
6. ✅ Integrate document linking in document creation
7. ✅ Integrate approval routing in submit endpoints
8. ✅ Integrate notifications in all workflow events
9. ✅ Add new relationship/audit endpoints
10. ✅ Test all integration points

**Total Integration Time**: ~4-6 hours
**Estimated Code Changes**: 500-800 lines in handlers

---

**Phase 12D Integration**: Ready for Phase 12E Testing
**Date**: December 22, 2025
