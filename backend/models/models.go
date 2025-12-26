package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// User represents a system user
type User struct {
	ID        string     `gorm:"primaryKey" json:"id"`
	Email     string     `gorm:"uniqueIndex" json:"email"`
	Name      string     `json:"name"`
	Role      string     `json:"role"` // admin, approver, requester, finance, viewer
	Active    bool       `json:"active"`
	LastLogin *time.Time `json:"lastLogin,omitempty"`

	// Multi-tenancy fields
	CurrentOrganizationID *string        `json:"currentOrganizationId,omitempty"`
	CurrentOrganization   *Organization `gorm:"foreignKey:CurrentOrganizationID" json:"currentOrganization,omitempty"`
	IsSuperAdmin          bool           `gorm:"default:false" json:"isSuperAdmin"`
	Preferences           datatypes.JSON `gorm:"type:jsonb" json:"preferences,omitempty"`

	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `gorm:"index" json:"deletedAt,omitempty"` // Soft delete
}

// Requisition workflow document
type Requisition struct {
	ID                string          `gorm:"primaryKey" json:"id"`
	OrganizationID    string          `gorm:"index;not null" json:"organizationId"`
	Organization      *Organization   `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	REQNumber         string          `gorm:"uniqueIndex" json:"reqNumber"`
	RequesterID       string          `json:"requesterId"`
	Requester         *User           `json:"requester,omitempty"`
	Title             string          `json:"title"`
	Description       string          `json:"description"`
	Department        string          `json:"department"`
	Status            string          `json:"status"` // draft, pending, approved, rejected
	Priority          string          `json:"priority"` // low, medium, high
	Items             datatypes.JSONType `gorm:"type:jsonb" json:"items"`
	TotalAmount       float64         `json:"totalAmount"`
	Currency          string          `json:"currency"`
	ApprovalStage     int             `json:"approvalStage"`
	ApprovalHistory   datatypes.JSONType `gorm:"type:jsonb" json:"approvalHistory"`
	CategoryID        *string         `json:"categoryId,omitempty"`
	Category          *Category       `json:"category,omitempty"`
	PreferredVendorID *string         `json:"preferredVendorId,omitempty"`
	PreferredVendor   *Vendor         `json:"preferredVendor,omitempty"`
	IsEstimate        bool            `json:"isEstimate"`
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
	Status          string          `json:"status"` // draft, pending, approved, rejected
	FiscalYear      string          `json:"fiscalYear"`
	TotalBudget     float64         `json:"totalBudget"`
	AllocatedAmount float64         `json:"allocatedAmount"`
	RemainingAmount float64         `json:"remainingAmount"`
	ApprovalStage   int             `json:"approvalStage"`
	ApprovalHistory datatypes.JSONType `gorm:"type:jsonb" json:"approvalHistory"`
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
	Status            string          `json:"status"` // draft, pending, approved, rejected, fulfilled
	Items             datatypes.JSONType `gorm:"type:jsonb" json:"items"`
	TotalAmount       float64         `json:"totalAmount"`
	Currency          string          `json:"currency"`
	DeliveryDate      time.Time       `json:"deliveryDate"`
	ApprovalStage     int             `json:"approvalStage"`
	ApprovalHistory   datatypes.JSONType `gorm:"type:jsonb" json:"approvalHistory"`
	LinkedRequisition string          `json:"linkedRequisition"`
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
	Status          string          `json:"status"` // draft, pending, approved, rejected, paid
	Amount          float64         `json:"amount"`
	Currency        string          `json:"currency"`
	PaymentMethod   string          `json:"paymentMethod"` // bank_transfer, check, cash
	GLCode          string          `json:"glCode"`
	Description     string          `json:"description"`
	ApprovalStage   int             `json:"approvalStage"`
	ApprovalHistory datatypes.JSONType `gorm:"type:jsonb" json:"approvalHistory"`
	LinkedPO        string          `json:"linkedPO"`
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
	PurchaseOrder     *PurchaseOrder  `json:"purchaseOrder,omitempty"`
	Status            string          `json:"status"` // draft, pending, approved, rejected, completed
	ReceivedDate      time.Time       `json:"receivedDate"`
	ReceivedBy        string          `json:"receivedBy"`
	Items             datatypes.JSONType `gorm:"type:jsonb" json:"items"`
	QualityIssues     datatypes.JSONType `gorm:"type:jsonb" json:"qualityIssues"`
	ApprovalStage     int             `json:"approvalStage"`
	ApprovalHistory   datatypes.JSONType `gorm:"type:jsonb" json:"approvalHistory"`
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

// Vendor master data
type Vendor struct {
	ID             string        `gorm:"primaryKey" json:"id"`
	OrganizationID string        `gorm:"index;not null" json:"organizationId"`
	Organization   *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	VendorCode     string        `gorm:"uniqueIndex:idx_org_vendor_code;index:idx_org_vendor_code" json:"vendorCode"`
	Name           string        `json:"name"`
	Email          string        `json:"email"`
	Phone          string        `json:"phone"`
	Country        string        `json:"country"`
	City           string        `json:"city"`
	BankAccount    string        `json:"bankAccount"`
	TaxID          string        `json:"taxId"`
	Active         bool          `json:"active"`
	CreatedAt      time.Time     `json:"createdAt"`
	UpdatedAt      time.Time     `json:"updatedAt"`
}

// ApprovalTask represents a pending approval action
type ApprovalTask struct {
	ID             string        `gorm:"primaryKey" json:"id"`
	OrganizationID string        `gorm:"index;not null" json:"organizationId"`
	Organization   *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	DocumentID     string        `gorm:"index" json:"documentId"`
	DocumentType   string        `json:"documentType"` // requisition, budget, po, pv, grn
	ApproverID     string        `json:"approverId"`
	Approver       *User         `json:"approver,omitempty"`
	Status         string        `json:"status"` // pending, approved, rejected
	Stage          int           `json:"stage"`
	Comments       string        `json:"comments"`
	Signature      string        `json:"signature"`
	CreatedAt      time.Time     `json:"createdAt"`
	UpdatedAt      time.Time     `json:"updatedAt"`
}

// AuditLog tracks all document changes
type AuditLog struct {
	ID            string         `gorm:"primaryKey" json:"id"`
	DocumentID    string         `gorm:"index" json:"documentId"`
	DocumentType  string         `json:"documentType"`
	UserID        string         `json:"userId"`
	Action        string         `json:"action"` // create, update, approve, reject
	Changes       datatypes.JSONType `gorm:"type:jsonb" json:"changes"`
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
