package models

import (
	"time"

	"github.com/liyali/liyali-gateway/types"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// User represents a system user
type User struct {
	ID        string     `gorm:"primaryKey" json:"id"`
	Email     string     `gorm:"uniqueIndex" json:"email"`
	Name      string     `json:"name"`
	Password  string     `json:"-"` // Hidden from JSON responses
	Role      string     `json:"role"` // admin, approver, requester, finance, viewer
	Active    bool       `json:"active"`
	LastLogin *time.Time `json:"lastLogin,omitempty"`

	// Multi-tenancy fields
	CurrentOrganizationID *string        `json:"currentOrganizationId,omitempty"`
	CurrentOrganization   *Organization `gorm:"foreignKey:CurrentOrganizationID" json:"currentOrganization,omitempty"`
	IsSuperAdmin          bool           `gorm:"default:false" json:"isSuperAdmin"`
	Preferences           datatypes.JSON `gorm:"type:jsonb" json:"preferences,omitempty"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt" gorm:"index"` // Soft delete
}

// Requisition workflow document
type Requisition struct {
	ID                string          `gorm:"primaryKey" json:"id"`
	OrganizationID    string          `gorm:"index;not null" json:"organizationId"`
	Organization      *Organization   `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	REQNumber         string          `gorm:"uniqueIndex" json:"reqNumber"`
	RequesterId       string          `json:"requesterId"`
	Requester         *User           `gorm:"foreignKey:RequesterId" json:"requester,omitempty"`
	RequesterName     string          `gorm:"column:created_by_name" json:"requesterName"`
	Title             string          `json:"title"`
	Description       string          `json:"description"`
	Department        string          `json:"department"`
	DepartmentId      string          `json:"departmentId"`
	Status            string          `json:"status"` // draft, pending, approved, rejected, completed, cancelled
	Priority          string          `json:"priority"` // low, medium, high, urgent
	Items             datatypes.JSONType[[]types.RequisitionItem] `gorm:"type:jsonb" json:"items"`
	TotalAmount       float64         `json:"totalAmount"`
	Currency          string          `json:"currency"`
	ApprovalStage     int             `json:"approvalStage"`
	ApprovalHistory   datatypes.JSONType[[]types.ApprovalRecord] `gorm:"type:jsonb" json:"approvalHistory"`
	CategoryID        *string         `json:"categoryId,omitempty"`
	Category          *Category       `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	CategoryName      string          `gorm:"-" json:"categoryName"`
	PreferredVendorID *string         `json:"preferredVendorId,omitempty"`
	PreferredVendor   *Vendor         `gorm:"foreignKey:PreferredVendorID" json:"preferredVendor,omitempty"`
	PreferredVendorName string        `gorm:"-" json:"preferredVendorName"`
	IsEstimate        bool            `json:"isEstimate"`

	// Business requirement fields
	RequisitionNumber   string                                    `gorm:"-" json:"requisitionNumber"`   // Computed from REQNumber
	BudgetCode          string                                    `gorm:"-" json:"budgetCode"`          // Stored in metadata
	RequestedByName     string                                    `gorm:"-" json:"requestedByName"`     // Computed from RequesterName
	RequestedByRole     string                                    `gorm:"-" json:"requestedByRole"`     // Computed from Requester.Role
	RequestedBy         string                                    `gorm:"-" json:"requestedBy"`         // Computed from RequesterId
	TotalApprovalStages int                                       `gorm:"-" json:"totalApprovalStages"` // Computed
	RequestedDate       time.Time                                 `gorm:"-" json:"requestedDate"`       // Computed from CreatedAt
	RequiredByDate      time.Time                                 `json:"requiredByDate"`
	CostCenter          string                                    `gorm:"-" json:"costCenter"`          // Stored in metadata
	ProjectCode         string                                    `gorm:"-" json:"projectCode"`         // Stored in metadata
	CreatedBy           string                                    `gorm:"-" json:"createdBy"`           // Computed from RequesterId
	CreatedByName       string                                    `gorm:"-" json:"createdByName"`       // Computed from RequesterName
	CreatedByRole       string                                    `gorm:"-" json:"createdByRole"`       // Computed from Requester.Role
	
	// Automation fields
	AutomationUsed      bool                                      `json:"automationUsed,omitempty"`     // Whether automation was used
	AutoCreatedPO       datatypes.JSON                           `gorm:"type:jsonb" json:"autoCreatedPO,omitempty"` // Auto-created Purchase Order
	
	ActionHistory       datatypes.JSONType[[]types.ActionHistoryEntry] `gorm:"-" json:"actionHistory"`
	Metadata            datatypes.JSON                           `gorm:"type:jsonb" json:"metadata"`

	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
}

// Budget workflow document
type Budget struct {
	ID              string          `gorm:"primaryKey" json:"id"`
	OrganizationID  string          `gorm:"index;not null" json:"organizationId"`
	Organization    *Organization   `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	OwnerID         string          `json:"ownerId"`
	Owner           *User           `json:"owner,omitempty"`
	BudgetCode      string          `gorm:"index" json:"budgetCode"`
	Department      string          `json:"department"`
	Status          string          `json:"status"` // draft, pending, approved, rejected, completed, cancelled
	FiscalYear      string          `json:"fiscalYear"`
	TotalBudget     float64         `json:"totalBudget"`
	AllocatedAmount float64         `json:"allocatedAmount"`
	RemainingAmount float64         `json:"remainingAmount"`
	ApprovalStage   int             `json:"approvalStage"`
	ApprovalHistory datatypes.JSONType[[]types.ApprovalRecord] `gorm:"type:jsonb" json:"approvalHistory"`

	// Extended fields for UI compatibility and business requirements
	Name            string                                    `json:"name,omitempty"`        // Budget name/title
	Description     string                                    `json:"description,omitempty"` // Budget description
	DepartmentID    string                                    `json:"departmentId,omitempty"` // Department ID
	Currency        string                                    `json:"currency,omitempty"`    // Currency
	OwnerName       string                                    `gorm:"-" json:"ownerName,omitempty"`   // Computed from Owner.Name
	CreatedBy       string                                    `json:"createdBy,omitempty"`   // Creator user ID
	Items           datatypes.JSON                           `gorm:"type:jsonb" json:"items,omitempty"` // Budget items breakdown
	ActionHistory   datatypes.JSONType[[]types.ActionHistoryEntry] `gorm:"type:jsonb" json:"actionHistory,omitempty"` // Action history for UI
	Metadata        datatypes.JSON                           `gorm:"type:jsonb" json:"metadata,omitempty"` // Generic metadata

	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
}

// PurchaseOrder workflow document
type PurchaseOrder struct {
	ID                string          `gorm:"primaryKey" json:"id"`
	OrganizationID    string          `gorm:"index;not null" json:"organizationId"`
	Organization      *Organization   `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	PONumber          string          `gorm:"uniqueIndex" json:"poNumber"`
	VendorID          string          `json:"vendorId"`
	Vendor            *Vendor         `json:"vendor,omitempty"`
	Status            string          `json:"status"` // draft, pending, approved, rejected, fulfilled, completed, cancelled
	Items             datatypes.JSONType[[]types.POItem] `gorm:"type:jsonb" json:"items"`
	TotalAmount       float64         `json:"totalAmount"`
	Currency          string          `json:"currency"`
	DeliveryDate      time.Time       `json:"deliveryDate"`
	ApprovalStage     int             `json:"approvalStage"`
	ApprovalHistory   datatypes.JSONType[[]types.ApprovalRecord] `gorm:"type:jsonb" json:"approvalHistory"`
	LinkedRequisition string          `json:"linkedRequisition"`

	// Frontend compatibility fields - CRITICAL: These must match frontend exactly
	VendorName    string     `gorm:"-" json:"vendorName,omitempty"`    // Computed from Vendor.Name
	Department    string     `json:"department,omitempty"`    // Department
	DepartmentID  string     `json:"departmentId,omitempty"`  // Department ID
	GLCode        string     `json:"glCode,omitempty"`        // GL Code - ADDED
	Title         string     `json:"title,omitempty"`         // PO title
	Description   string     `json:"description,omitempty"`   // PO description
	Priority      string     `json:"priority,omitempty"`      // Priority level
	Subtotal      *float64   `json:"subtotal,omitempty"`      // Subtotal amount
	Tax           *float64   `json:"tax,omitempty"`           // Tax amount
	Total         *float64   `json:"total,omitempty"`         // Total amount
	BudgetCode    string     `json:"budgetCode,omitempty"`    // Budget code - ADDED
	CostCenter    string     `json:"costCenter,omitempty"`    // Cost center - ADDED
	ProjectCode   string     `json:"projectCode,omitempty"`   // Project code - ADDED
	
	// Automation fields
	AutomationUsed    bool           `json:"automationUsed,omitempty"`    // Whether automation was used
	AutoCreatedGRN    datatypes.JSON `gorm:"type:jsonb" json:"autoCreatedGRN,omitempty"` // Auto-created GRN
	
	ActionHistory datatypes.JSONType[[]types.ActionHistoryEntry] `gorm:"type:jsonb" json:"actionHistory,omitempty"` // Action history for UI
	
	// Legacy aliases for backward compatibility
	RequiredByDate          *time.Time `json:"requiredByDate,omitempty"`          // Required delivery date
	SourceRequisitionNumber string     `json:"sourceRequisitionNumber,omitempty"` // Source requisition number
	SourceRequisitionId     *string    `gorm:"column:source_requisition_id" json:"sourceRequisitionId,omitempty"`     // Source requisition ID

	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
}

// PaymentVoucher workflow document
type PaymentVoucher struct {
	ID              string          `gorm:"primaryKey" json:"id"`
	OrganizationID  string          `gorm:"index;not null" json:"organizationId"`
	Organization    *Organization   `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	VoucherNumber   string          `gorm:"uniqueIndex" json:"voucherNumber"`
	VendorID        string          `json:"vendorId"`
	Vendor          *Vendor         `json:"vendor,omitempty"`
	InvoiceNumber   string          `json:"invoiceNumber"`
	Status          string          `json:"status"` // draft, pending, approved, rejected, paid, completed, cancelled
	Amount          float64         `json:"amount"`
	Currency        string          `json:"currency"`
	PaymentMethod   string          `json:"paymentMethod"` // bank_transfer, cash
	GLCode          string          `json:"glCode"`
	Description     string          `json:"description"`
	ApprovalStage   int             `json:"approvalStage"`
	ApprovalHistory datatypes.JSONType[[]types.ApprovalRecord] `gorm:"type:jsonb" json:"approvalHistory"`
	LinkedPO        string          `json:"linkedPO"`

	// Frontend compatibility fields - CRITICAL: These must match frontend exactly
	VendorName              string                                    `json:"vendorName,omitempty"`              // Computed from Vendor.Name
	Title                   string                                    `json:"title,omitempty"`                   // Payment voucher title
	Department              string                                    `json:"department,omitempty"`              // Department
	DepartmentID            string                                    `json:"departmentId,omitempty"`            // Department ID
	Priority                string                                    `json:"priority,omitempty"`                // Priority level
	RequestedByName         string                                    `json:"requestedByName,omitempty"`         // Name of requester
	RequestedDate           *time.Time                                `json:"requestedDate,omitempty"`           // When payment was requested
	SubmittedAt             *time.Time                                `json:"submittedAt,omitempty"`             // Submission date
	ApprovedAt              *time.Time                                `json:"approvedAt,omitempty"`              // Approval date
	PaidDate                *time.Time                                `json:"paidDate,omitempty"`                // Payment date
	PaymentDueDate          *time.Time                                `json:"paymentDueDate,omitempty"`          // Payment due date - ADDED
	BudgetCode              string                                    `json:"budgetCode,omitempty"`              // Budget code
	CostCenter              string                                    `json:"costCenter,omitempty"`              // Cost center
	ProjectCode             string                                    `json:"projectCode,omitempty"`             // Project code
	TaxAmount               *float64                                  `json:"taxAmount,omitempty"`               // Tax amount
	WithholdingTaxAmount    *float64                                  `json:"withholdingTaxAmount,omitempty"`    // Withholding tax
	PaidAmount              *float64                                  `json:"paidAmount,omitempty"`              // Amount actually paid
	SourcePurchaseOrderNumber string                                  `json:"sourcePurchaseOrderNumber,omitempty"` // Source PO number
	SourceRequisitionNumber string                                    `json:"sourceRequisitionNumber,omitempty"` // Source requisition number
	BankDetails             datatypes.JSON                           `gorm:"type:jsonb" json:"bankDetails,omitempty"` // Bank details for payment
	Items                   datatypes.JSONType[[]types.PaymentItem]  `gorm:"type:jsonb" json:"items,omitempty"`       // Payment items breakdown
	ActionHistory           datatypes.JSONType[[]types.ActionHistoryEntry] `gorm:"type:jsonb" json:"actionHistory,omitempty"` // Action history for UI
	
	// Legacy aliases for backward compatibility
	PVNumber                string                                    `json:"pvNumber,omitempty"`                // Alias for VoucherNumber
	TotalAmount             float64                                   `json:"totalAmount,omitempty"`             // Alias for Amount
	CurrentApprovalStage    int                                       `json:"currentApprovalStage,omitempty"`    // Alias for ApprovalStage

	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
}

// GoodsReceivedNote workflow document
type GoodsReceivedNote struct {
	ID                string          `gorm:"primaryKey" json:"id"`
	OrganizationID    string          `gorm:"index;not null" json:"organizationId"`
	Organization      *Organization   `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	GRNNumber         string          `gorm:"uniqueIndex" json:"grnNumber"`
	PONumber          string          `json:"poNumber"`
	PurchaseOrder     *PurchaseOrder  `gorm:"foreignKey:PONumber;references:PONumber" json:"purchaseOrder,omitempty"`
	Status            string          `json:"status"` // draft, pending, approved, rejected, paid, completed, cancelled
	ReceivedDate      time.Time       `json:"receivedDate"`
	ReceivedBy        string          `json:"receivedBy"`
	Items             datatypes.JSONType[[]types.GRNItem] `gorm:"type:jsonb" json:"items"`
	QualityIssues     datatypes.JSONType[[]types.QualityIssue] `gorm:"type:jsonb" json:"qualityIssues"`
	ApprovalStage     int             `json:"approvalStage"`
	ApprovalHistory   datatypes.JSONType[[]types.ApprovalRecord] `gorm:"type:jsonb" json:"approvalHistory"`

	// Extended fields for UI compatibility and business requirements
	CreatedBy         string                                    `json:"createdBy,omitempty"`         // Creator user ID
	OwnerID           string                                    `json:"ownerId,omitempty"`           // Owner user ID (maps to createdBy)
	WarehouseLocation string                                    `json:"warehouseLocation,omitempty"` // Warehouse location
	Notes             string                                    `json:"notes,omitempty"`             // Additional notes
	CurrentStage      int                                       `json:"currentStage,omitempty"`      // Maps to ApprovalStage
	StageName         string                                    `json:"stageName,omitempty"`         // Current stage name
	ApprovedBy        string                                    `json:"approvedBy,omitempty"`        // Who approved the GRN
	AutomationUsed    bool                                      `json:"automationUsed,omitempty"`    // Whether automation was used
	AutoCreatedPV     datatypes.JSON                           `gorm:"type:jsonb" json:"autoCreatedPV,omitempty"` // Auto-created Payment Voucher
	ActionHistory     datatypes.JSONType[[]types.ActionHistoryEntry] `gorm:"type:jsonb" json:"actionHistory,omitempty"` // Action history for UI
	Metadata          datatypes.JSON                           `gorm:"type:jsonb" json:"metadata,omitempty"` // Generic metadata

	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
}

// Category master data for requisition categorization
type Category struct {
	ID             string        `gorm:"primaryKey" json:"id"`
	OrganizationID string        `gorm:"index;not null" json:"organizationId"`
	Organization   *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	Name           string        `gorm:"uniqueIndex:idx_org_category_name;index:idx_org_category_name" json:"name"`
	Description    string        `json:"description"`
	Active         bool          `json:"active"`
	CreatedAt      time.Time     `json:"createdAt"`
	UpdatedAt      time.Time     `json:"updatedAt"`
}

// CategoryBudgetCode links categories to budget codes (one-to-many relationship)
type CategoryBudgetCode struct {
	ID         string    `gorm:"primaryKey" json:"id"`
	CategoryID string    `gorm:"index" json:"categoryId"`
	Category   *Category `json:"category,omitempty"`
	BudgetCode string    `gorm:"index" json:"budgetCode"`
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// Vendor master data - Organization-scoped vendors
type Vendor struct {
	ID             string        `gorm:"primaryKey" json:"id"`
	OrganizationID string        `gorm:"index;not null" json:"organizationId"`
	Organization   *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	VendorCode     string        `gorm:"uniqueIndex:idx_org_vendor_code" json:"vendorCode"`
	Name           string        `json:"name"`
	Email          string        `gorm:"index" json:"email"`
	Phone          string        `json:"phone"`
	Country        string        `json:"country"`
	City           string        `json:"city"`
	BankAccount    string        `json:"bankAccount"`
	TaxID          string        `json:"taxId"`
	Active         bool          `json:"active"`
	CreatedBy      string        `json:"createdBy"` // User who created the vendor
	CreatedAt      time.Time     `json:"createdAt"`
	UpdatedAt      time.Time     `json:"updatedAt"`
}

// ApprovalTask represents a pending approval action
type ApprovalTask struct {
	ID               string        `gorm:"primaryKey" json:"id"`
	OrganizationID   string        `gorm:"index;not null" json:"organizationId"`
	Organization     *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	DocumentID       string        `gorm:"index" json:"documentId"`
	DocumentType     string        `json:"documentType"` // requisition, budget, po, pv, grn
	ApproverID       string        `json:"approverId"`
	Approver         *User         `json:"approver,omitempty"`
	AssignedTo       string        `json:"assignedTo"`       // Current assignee
	Status           string        `json:"status"`           // pending, approved, rejected
	Stage            int           `json:"stage"`
	Comments         *string       `json:"comments"`
	Signature        *string       `json:"signature"`
	ApprovedBy       *string       `json:"approvedBy"`
	ApprovedAt       *time.Time    `json:"approvedAt"`
	RejectedBy       *string       `json:"rejectedBy"`
	RejectedAt       *time.Time    `json:"rejectedAt"`
	RejectionReason  *string       `json:"rejectionReason"`

	// Frontend compatibility fields
	DocumentNumber   string     `gorm:"-" json:"documentNumber,omitempty"`   // Computed from document reference
	ApproverName     string     `gorm:"-" json:"approverName,omitempty"`     // Computed from Approver.Name
	Priority         string     `gorm:"-" json:"priority,omitempty"`         // Computed from document priority
	DueAt            *time.Time `json:"dueAt,omitempty"`            // Due date for the approval
	TaskType         string     `gorm:"-" json:"taskType,omitempty"`         // Computed task type for UI display
	Title            string     `gorm:"-" json:"title,omitempty"`            // Computed human-readable task title
	WorkflowID       string     `gorm:"-" json:"workflowId,omitempty"`       // Computed workflow ID for the task
	WorkflowName     string     `gorm:"-" json:"workflowName,omitempty"`     // Computed workflow name for the task
	StageName        string     `gorm:"-" json:"stageName,omitempty"`        // Computed human-readable stage name
	Importance       string     `gorm:"-" json:"importance,omitempty"`       // Computed from priority

	CreatedAt        time.Time     `json:"createdAt"`
	UpdatedAt        time.Time     `json:"updatedAt"`
}

// AuditLog tracks all document changes
type AuditLog struct {
	ID            string         `gorm:"primaryKey" json:"id"`
	DocumentID    string         `gorm:"index" json:"documentId"`
	DocumentType  string         `json:"documentType"`
	UserID        string         `json:"userId"`
	Action        string         `json:"action"` // create, update, approve, reject
	Changes       datatypes.JSONType[map[string]interface{}] `gorm:"type:jsonb" json:"changes"`
	CreatedAt     time.Time      `json:"createdAt"`
}

// Notification for email/SMS delivery
type Notification struct {
	ID             string        `gorm:"primaryKey" json:"id"`
	OrganizationID string        `gorm:"index;not null" json:"organizationId"`
	Organization   *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	RecipientID    string        `json:"recipientId"`
	Recipient      *User         `json:"recipient,omitempty"`
	Type           string        `json:"type"` // approval_required, approved, rejected, assigned
	DocumentID     string        `json:"documentId"`
	DocumentType   string        `json:"documentType"`
	Subject        string        `json:"subject"`
	Body           string        `json:"body"`
	Sent           bool          `json:"sent"`
	SentAt         *time.Time    `json:"sentAt,omitempty"`

	// Frontend compatibility fields
	EntityID           string                 `json:"entityId,omitempty"`           // Maps to documentId for backward compatibility
	EntityType         string                 `json:"entityType,omitempty"`         // Maps to documentType for backward compatibility
	EntityNumber       string                 `json:"entityNumber,omitempty"`       // Document reference number
	RelatedUserID      string                 `json:"relatedUserId,omitempty"`      // User who triggered the notification
	RelatedUserName    string                 `json:"relatedUserName,omitempty"`    // Name of the user who triggered the notification
	IsRead             bool                   `json:"isRead,omitempty"`             // Read status
	ReadAt             *time.Time             `json:"readAt,omitempty"`             // When notification was read
	ActionTaken        bool                   `json:"actionTaken,omitempty"`        // Whether action was taken
	ActionTakenAt      *time.Time             `json:"actionTakenAt,omitempty"`      // When action was taken
	Importance         string                 `json:"importance,omitempty"`         // Notification importance (HIGH, MEDIUM, LOW)
	QuickAction        datatypes.JSON         `gorm:"type:jsonb" json:"quickAction,omitempty"` // Quick action configuration
	ReassignmentReason string                 `json:"reassignmentReason,omitempty"` // Reason for reassignment (if applicable)
	Message            string                 `json:"message,omitempty"`            // Message content for filtering

	CreatedAt      time.Time     `json:"createdAt"`
	UpdatedAt      time.Time     `json:"updatedAt"`
}

// TableName specifies table names for GORM
func (User) TableName() string { return "users" }
func (Requisition) TableName() string { return "requisitions" }
func (Budget) TableName() string { return "budgets" }
func (PurchaseOrder) TableName() string { return "purchase_orders" }
func (PaymentVoucher) TableName() string { return "payment_vouchers" }
func (GoodsReceivedNote) TableName() string { return "goods_received_notes" }
func (Category) TableName() string { return "categories" }
func (CategoryBudgetCode) TableName() string { return "category_budget_codes" }
func (Vendor) TableName() string { return "vendors" }
func (ApprovalTask) TableName() string { return "approval_tasks" }
func (AuditLog) TableName() string { return "audit_logs" }
func (Notification) TableName() string { return "notifications" }
