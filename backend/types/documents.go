package types

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ================== REQUISITION TYPES ==================

// CreateRequisitionRequest represents a requisition creation request
type CreateRequisitionRequest struct {
	Title             string            `json:"title" validate:"required,min=3"`
	Description       string            `json:"description" validate:"required,min=10"`
	Department        string            `json:"department" validate:"required"`
	DepartmentId      string            `json:"departmentId"`
	Priority          string            `json:"priority" validate:"required,oneof=low medium high urgent"`
	Items             []RequisitionItem `json:"items" validate:"required,min=1"`
	TotalAmount       float64           `json:"totalAmount" validate:"required,gt=0"`
	Currency          string            `json:"currency" validate:"required"`
	CategoryID        *string           `json:"categoryId" validate:"omitempty,uuid"`
	PreferredVendorID *string           `json:"preferredVendorId" validate:"omitempty,uuid"`
	IsEstimate        bool              `json:"isEstimate"`

	// Business requirement fields
	BudgetCode        string                 `json:"budgetCode"`
	CostCenter        string                 `json:"costCenter"`
	ProjectCode       string                 `json:"projectCode"`
	RequiredByDate    time.Time              `json:"requiredByDate"`
	RequestedFor      string                 `json:"requestedFor"`      // Who the requisition is for
	OtherCategoryText string                 `json:"otherCategoryText"` // Custom category name when "OTHER" is selected
	Metadata          map[string]interface{} `json:"metadata"`          // Additional metadata (e.g. attachments)
}

// UpdateRequisitionRequest represents a requisition update request
type UpdateRequisitionRequest struct {
	Title             string                 `json:"title"`
	Description       string                 `json:"description"`
	Department        string                 `json:"department"`
	Priority          string                 `json:"priority"`
	Items             []RequisitionItem      `json:"items"`
	TotalAmount       float64                `json:"totalAmount"`
	Currency          string                 `json:"currency"`
	CategoryID        *string                `json:"categoryId" validate:"omitempty,uuid"`
	PreferredVendorID *string                `json:"preferredVendorId" validate:"omitempty,uuid"`
	IsEstimate        *bool                  `json:"isEstimate"`
	Metadata          map[string]interface{} `json:"metadata"` // Additional metadata (e.g. attachments)
}

// RequisitionItem represents an item in a requisition
type RequisitionItem struct {
	ID              *string  `json:"id,omitempty"`
	Description     string   `json:"description"`
	ItemDescription *string  `json:"itemDescription,omitempty"` // Alias for description
	Quantity        int      `json:"quantity"`
	UnitPrice       float64  `json:"unitPrice"`
	Amount          float64  `json:"amount"`
	EstimatedCost   *float64 `json:"estimatedCost,omitempty"` // Alias for amount
	Unit            *string  `json:"unit,omitempty"`
	Category        *string  `json:"category,omitempty"`
	Notes           *string  `json:"notes,omitempty"`
}

// RequisitionResponse represents a requisition in responses
type RequisitionResponse struct {
	ID                  string            `json:"id"`
	DocumentNumber      string            `json:"documentNumber"`
	RequesterID         string            `json:"requesterId"`
	RequesterName       string            `json:"requesterName"`
	Title               string            `json:"title"`
	Description         string            `json:"description"`
	Department          string            `json:"department"`
	Status              string            `json:"status"`
	Priority            string            `json:"priority"`
	Items               []RequisitionItem `json:"items"`
	TotalAmount         float64           `json:"totalAmount"`
	Currency            string            `json:"currency"`
	CategoryID          *string           `json:"categoryId,omitempty"`
	CategoryName        string            `json:"categoryName,omitempty"`
	PreferredVendorID   *string           `json:"preferredVendorId,omitempty"`
	PreferredVendorName string            `json:"preferredVendorName,omitempty"`
	IsEstimate          bool              `json:"isEstimate"`
	ApprovalStage       int               `json:"approvalStage"`
	ApprovalHistory     []ApprovalRecord  `json:"approvalHistory"`

	// Business requirement fields
	BudgetCode        string    `json:"budgetCode,omitempty"`
	CostCenter        string    `json:"costCenter,omitempty"`
	ProjectCode       string    `json:"projectCode,omitempty"`
	RequiredByDate    time.Time `json:"requiredByDate,omitempty"`
	RequestedFor      string    `json:"requestedFor,omitempty"`      // From metadata
	OtherCategoryText string    `json:"otherCategoryText,omitempty"` // From metadata

	// Additional metadata (e.g. attachments)
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Action history for frontend
	ActionHistory []ActionHistoryEntry `json:"actionHistory,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ================== BUDGET TYPES ==================

// CreateBudgetRequest represents a budget creation request
type CreateBudgetRequest struct {
	BudgetCode      string  `json:"budgetCode,omitempty"`  // Optional - can be auto-generated
	Name            string  `json:"name,omitempty"`        // Budget name/title
	Description     string  `json:"description,omitempty"` // Budget description
	Department      string  `json:"department" validate:"required"`
	DepartmentID    string  `json:"departmentId,omitempty"` // Department ID
	FiscalYear      string  `json:"fiscalYear" validate:"required"`
	TotalBudget     float64 `json:"totalBudget" validate:"required,gt=0"`
	AllocatedAmount float64 `json:"allocatedAmount" validate:"required,gte=0"`
	Currency        string  `json:"currency,omitempty"` // Currency
}

// UpdateBudgetRequest represents a budget update request
type UpdateBudgetRequest struct {
	Department      string        `json:"department"`
	TotalBudget     float64       `json:"totalBudget"`
	AllocatedAmount float64       `json:"allocatedAmount"`
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	Currency        string        `json:"currency"`
	Items           []interface{} `json:"items"`
}

// BudgetResponse represents a budget in responses
type BudgetResponse struct {
	ID              string               `json:"id"`
	BudgetCode      string               `json:"budgetCode"`
	OwnerID         string               `json:"ownerId"`
	OwnerName       string               `json:"ownerName"`
	Department      string               `json:"department"`
	DepartmentID    string               `json:"departmentId,omitempty"`
	Status          string               `json:"status"`
	FiscalYear      string               `json:"fiscalYear"`
	TotalBudget     float64              `json:"totalBudget"`
	AllocatedAmount float64              `json:"allocatedAmount"`
	RemainingAmount float64              `json:"remainingAmount"`
	ApprovalStage   int                  `json:"approvalStage"`
	ApprovalHistory []ApprovalRecord     `json:"approvalHistory"`
	ActionHistory   []ActionHistoryEntry `json:"actionHistory,omitempty"`
	Name            string               `json:"name,omitempty"`
	Description     string               `json:"description,omitempty"`
	Currency        string               `json:"currency,omitempty"`
	CreatedBy       string               `json:"createdBy,omitempty"`
	Items           []interface{}        `json:"items"`
	CreatedAt       time.Time            `json:"createdAt"`
	UpdatedAt       time.Time            `json:"updatedAt"`
}

// ================== PURCHASE ORDER TYPES ==================

// FlexibleDate handles both RFC3339 and date-only formats
type FlexibleDate struct {
	time.Time
}

// UnmarshalJSON implements custom JSON unmarshaling for flexible date formats
func (fd *FlexibleDate) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), "\"")

	// Try RFC3339 format first (2006-01-02T15:04:05Z07:00)
	if t, err := time.Parse(time.RFC3339, str); err == nil {
		fd.Time = t
		return nil
	}

	// Try date-only format (2006-01-02)
	if t, err := time.Parse("2006-01-02", str); err == nil {
		fd.Time = t
		return nil
	}

	// Try datetime format without timezone (2006-01-02T15:04:05)
	if t, err := time.Parse("2006-01-02T15:04:05", str); err == nil {
		fd.Time = t
		return nil
	}

	return fmt.Errorf("invalid date format: %s (expected RFC3339, date-only, or datetime)", str)
}

// MarshalJSON implements custom JSON marshaling
func (fd FlexibleDate) MarshalJSON() ([]byte, error) {
	return json.Marshal(fd.Time.Format(time.RFC3339))
}

// CreatePurchaseOrderRequest represents a PO creation request
type CreatePurchaseOrderRequest struct {
	VendorID          string       `json:"vendorId" validate:"required"`
	Items             []POItem     `json:"items" validate:"required,min=1"`
	TotalAmount       float64      `json:"totalAmount" validate:"required,gt=0"`
	Currency          string       `json:"currency" validate:"required"`
	DeliveryDate      FlexibleDate `json:"deliveryDate" validate:"required"`
	LinkedRequisition string       `json:"linkedRequisition"`
}

// UpdatePurchaseOrderRequest represents a PO update request
type UpdatePurchaseOrderRequest struct {
	VendorID     string       `json:"vendorId"`
	Items        []POItem     `json:"items"`
	TotalAmount  float64      `json:"totalAmount"`
	Currency     string       `json:"currency"`
	DeliveryDate FlexibleDate `json:"deliveryDate"`
}

// POItem represents an item in a purchase order
type POItem struct {
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unitPrice"`
	Amount      float64 `json:"amount"`

	// Frontend compatibility fields - CRITICAL: These must match frontend exactly
	ID         string  `json:"id,omitempty"`         // Item identifier - ADDED
	ItemNumber string  `json:"itemNumber,omitempty"` // Item number/SKU - ADDED
	ItemCode   string  `json:"itemCode,omitempty"`   // Item code - ADDED
	Category   string  `json:"category,omitempty"`   // Item category - ADDED
	Unit       string  `json:"unit,omitempty"`       // Unit of measurement - ADDED
	TotalPrice float64 `json:"totalPrice,omitempty"` // Total price (alias for amount) - ADDED
	Notes      string  `json:"notes,omitempty"`      // Item notes - ADDED
}

// PurchaseOrderResponse represents a PO in responses
type PurchaseOrderResponse struct {
	ID                string           `json:"id"`
	DocumentNumber    string           `json:"documentNumber"`
	VendorID          string           `json:"vendorId"`
	VendorName        string           `json:"vendorName"`
	Status            string           `json:"status"`
	Items             []POItem         `json:"items"`
	TotalAmount       float64          `json:"totalAmount"`
	Currency          string           `json:"currency"`
	DeliveryDate      time.Time        `json:"deliveryDate"`
	ApprovalStage     int              `json:"approvalStage"`
	ApprovalHistory   []ApprovalRecord `json:"approvalHistory"`
	LinkedRequisition string           `json:"linkedRequisition"`
	CreatedAt         time.Time        `json:"createdAt"`
	UpdatedAt         time.Time        `json:"updatedAt"`
}

// ================== PAYMENT VOUCHER TYPES ==================

// CreatePaymentVoucherRequest represents a payment voucher creation request
type CreatePaymentVoucherRequest struct {
	VendorID      string  `json:"vendorId" validate:"required"`
	InvoiceNumber string  `json:"invoiceNumber" validate:"required"`
	Amount        float64 `json:"amount" validate:"required,gt=0"`
	Currency      string  `json:"currency" validate:"required"`
	PaymentMethod string  `json:"paymentMethod" validate:"required,oneof=bank_transfer cash"`
	GLCode        string  `json:"glCode" validate:"required"`
	Description   string  `json:"description" validate:"required,min=10"`
	LinkedPO      string  `json:"linkedPO"`
}

// UpdatePaymentVoucherRequest represents a payment voucher update request
type UpdatePaymentVoucherRequest struct {
	VendorID      string  `json:"vendorId"`
	InvoiceNumber string  `json:"invoiceNumber"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	PaymentMethod string  `json:"paymentMethod"`
	GLCode        string  `json:"glCode"`
	Description   string  `json:"description"`
}

// PaymentVoucherResponse represents a payment voucher in responses
type PaymentVoucherResponse struct {
	ID              string           `json:"id"`
	DocumentNumber  string           `json:"documentNumber"`
	VendorID        string           `json:"vendorId"`
	VendorName      string           `json:"vendorName"`
	InvoiceNumber   string           `json:"invoiceNumber"`
	Status          string           `json:"status"`
	Amount          float64          `json:"amount"`
	Currency        string           `json:"currency"`
	PaymentMethod   string           `json:"paymentMethod"`
	GLCode          string           `json:"glCode"`
	Description     string           `json:"description"`
	ApprovalStage   int              `json:"approvalStage"`
	ApprovalHistory []ApprovalRecord `json:"approvalHistory"`
	LinkedPO        string           `json:"linkedPO"`
	CreatedAt       time.Time        `json:"createdAt"`
	UpdatedAt       time.Time        `json:"updatedAt"`
}

// ================== GRN TYPES ==================

// CreateGRNRequest represents a GRN creation request
type CreateGRNRequest struct {
	PODocumentNumber string    `json:"poDocumentNumber" validate:"required"`
	Items            []GRNItem `json:"items" validate:"required,min=1"`
	ReceivedBy       string    `json:"receivedBy" validate:"required"`
}

// UpdateGRNRequest represents a GRN update request
type UpdateGRNRequest struct {
	Items         []GRNItem      `json:"items"`
	ReceivedBy    string         `json:"receivedBy"`
	QualityIssues []QualityIssue `json:"qualityIssues"`
}

// GRNItem represents an item in a GRN
type GRNItem struct {
	Description      string  `json:"description"`
	QuantityOrdered  int     `json:"quantityOrdered"`
	QuantityReceived int     `json:"quantityReceived"`
	Variance         int     `json:"variance"`
	Condition        string  `json:"condition"`       // good, damaged, defective
	Notes            *string `json:"notes,omitempty"` // Optional notes for the item
}

// QualityIssue represents a quality issue in GRN
type QualityIssue struct {
	ItemDescription string `json:"itemDescription"`
	IssueType       string `json:"issueType"`
	Description     string `json:"description"`
	Severity        string `json:"severity"` // low, medium, high
}

// GRNResponse represents a GRN in responses
type GRNResponse struct {
	ID               string           `json:"id"`
	DocumentNumber   string           `json:"documentNumber"`
	PODocumentNumber string           `json:"poDocumentNumber"`
	Status           string           `json:"status"`
	ReceivedDate     time.Time        `json:"receivedDate"`
	ReceivedBy       string           `json:"receivedBy"`
	Items            []GRNItem        `json:"items"`
	QualityIssues    []QualityIssue   `json:"qualityIssues"`
	ApprovalStage    int              `json:"approvalStage"`
	ApprovalHistory  []ApprovalRecord `json:"approvalHistory"`
	CreatedAt        time.Time        `json:"createdAt"`
	UpdatedAt        time.Time        `json:"updatedAt"`
}

// ================== VENDOR TYPES ==================

// CreateVendorRequest represents a vendor creation request
type CreateVendorRequest struct {
	Name        string `json:"name" validate:"required,min=3"`
	Email       string `json:"email" validate:"required,email"`
	Phone       string `json:"phone" validate:"required"`
	Country     string `json:"country" validate:"required"`
	City        string `json:"city" validate:"required"`
	BankAccount string `json:"bankAccount" validate:"required"`
	TaxID       string `json:"taxId" validate:"required"`
}

// UpdateVendorRequest represents a vendor update request
type UpdateVendorRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Country     string `json:"country"`
	City        string `json:"city"`
	BankAccount string `json:"bankAccount"`
	TaxID       string `json:"taxId"`
	Active      bool   `json:"active"`
}

// VendorResponse represents a vendor in responses
type VendorResponse struct {
	ID          string    `json:"id"`
	VendorCode  string    `json:"vendorCode"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	Country     string    `json:"country"`
	City        string    `json:"city"`
	BankAccount string    `json:"bankAccount"`
	TaxID       string    `json:"taxId"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ================== APPROVAL TYPES ==================

// ApproveDocumentRequest represents a document approval request
type ApproveDocumentRequest struct {
	Comments  string `json:"comments"`
	Signature string `json:"signature" validate:"required"`
}

// RejectDocumentRequest represents a document rejection request
type RejectDocumentRequest struct {
	Remarks   string `json:"remarks" validate:"required,min=10"`
	Signature string `json:"signature" validate:"required"`
}

// ReassignDocumentRequest represents a document reassignment request
type ReassignDocumentRequest struct {
	NewApproverID string `json:"newApproverId" validate:"required"`
	Reason        string `json:"reason"`
}

// ================== COMMON RESPONSE TYPES ==================

// PaginationMeta represents pagination information
type PaginationMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"totalPages"`
	HasNext    bool  `json:"hasNext"`
	HasPrev    bool  `json:"hasPrev"`
}

// ListResponse represents a paginated list response
type ListResponse struct {
	Success    bool            `json:"success"`
	Data       interface{}     `json:"data"`
	Pagination *PaginationMeta `json:"pagination,omitempty"`
}

// DetailResponse represents a single resource response
type DetailResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ================== ADDITIONAL TYPES FOR FRONTEND COMPATIBILITY ==================

// ActionHistoryEntry represents an action history entry
type ActionHistoryEntry struct {
	ID              string                 `json:"id"`
	Action          string                 `json:"action"`
	PerformedBy     string                 `json:"performedBy"`
	PerformedByName string                 `json:"performedByName"`
	PerformedByRole string                 `json:"performedByRole,omitempty"`
	Timestamp       time.Time              `json:"timestamp"`
	PerformedAt     time.Time              `json:"performedAt,omitempty"` // Alias for timestamp
	Changes         map[string]interface{} `json:"changes,omitempty"`
	Comments        string                 `json:"comments,omitempty"`
	ActionType      string                 `json:"actionType,omitempty"`
	NewStatus       string                 `json:"newStatus,omitempty"`
	PreviousStatus  string                 `json:"previousStatus,omitempty"` // Previous status before action - ADDED
	Remarks         string                 `json:"remarks,omitempty"`
	StageNumber     int                    `json:"stageNumber,omitempty"`
	StageName       string                 `json:"stageName,omitempty"`
	ChangedFields   map[string]interface{} `json:"changedFields,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// PaymentItem represents an item in a payment voucher
type PaymentItem struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	GLCode      string  `json:"glCode,omitempty"`
	TaxAmount   float64 `json:"taxAmount,omitempty"`
}

// SubmitDocumentRequest represents workflow selection for submission.
type SubmitDocumentRequest struct {
	WorkflowID string `json:"workflowId" validate:"required,uuid"`
}
