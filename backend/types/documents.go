package types

import "time"

// ================== REQUISITION TYPES ==================

// CreateRequisitionRequest represents a requisition creation request
type CreateRequisitionRequest struct {
	Title             string                 `json:"title" validate:"required,min=3"`
	Description       string                 `json:"description" validate:"required,min=10"`
	Department        string                 `json:"department" validate:"required"`
	Priority          string                 `json:"priority" validate:"required,oneof=low medium high"`
	Items             []RequisitionItem       `json:"items" validate:"required,min=1"`
	TotalAmount       float64                `json:"totalAmount" validate:"required,gt=0"`
	Currency          string                 `json:"currency" validate:"required"`
	CategoryID        *string                `json:"categoryId" validate:"omitempty,uuid"`
	PreferredVendorID *string                `json:"preferredVendorId" validate:"omitempty,uuid"`
	IsEstimate        bool                   `json:"isEstimate"`
}

// UpdateRequisitionRequest represents a requisition update request
type UpdateRequisitionRequest struct {
	Title             string           `json:"title"`
	Description       string           `json:"description"`
	Department        string           `json:"department"`
	Priority          string           `json:"priority"`
	Items             []RequisitionItem `json:"items"`
	TotalAmount       float64          `json:"totalAmount"`
	Currency          string           `json:"currency"`
	CategoryID        *string          `json:"categoryId" validate:"omitempty,uuid"`
	PreferredVendorID *string          `json:"preferredVendorId" validate:"omitempty,uuid"`
	IsEstimate        *bool            `json:"isEstimate"`
}

// RequisitionItem represents an item in a requisition
type RequisitionItem struct {
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unitPrice"`
	Amount      float64 `json:"amount"`
}

// RequisitionResponse represents a requisition in responses
type RequisitionResponse struct {
	ID                  string            `json:"id"`
	RequesterID         string            `json:"requesterId"`
	RequesterName       string            `json:"requesterName"`
	Title               string            `json:"title"`
	Description         string            `json:"description"`
	Department          string            `json:"department"`
	Status              string            `json:"status"`
	Priority            string            `json:"priority"`
	Items               []RequisitionItem  `json:"items"`
	TotalAmount         float64           `json:"totalAmount"`
	Currency            string            `json:"currency"`
	CategoryID          *string           `json:"categoryId,omitempty"`
	CategoryName        string            `json:"categoryName,omitempty"`
	PreferredVendorID   *string           `json:"preferredVendorId,omitempty"`
	PreferredVendorName string            `json:"preferredVendorName,omitempty"`
	IsEstimate          bool              `json:"isEstimate"`
	ApprovalStage       int               `json:"approvalStage"`
	ApprovalHistory     []ApprovalRecord  `json:"approvalHistory"`
	CreatedAt           time.Time         `json:"createdAt"`
	UpdatedAt           time.Time         `json:"updatedAt"`
}

// ================== BUDGET TYPES ==================

// CreateBudgetRequest represents a budget creation request
type CreateBudgetRequest struct {
	BudgetCode      string  `json:"budgetCode" validate:"required"`
	Department      string  `json:"department" validate:"required"`
	FiscalYear      string  `json:"fiscalYear" validate:"required"`
	TotalBudget     float64 `json:"totalBudget" validate:"required,gt=0"`
	AllocatedAmount float64 `json:"allocatedAmount" validate:"required,gte=0"`
}

// UpdateBudgetRequest represents a budget update request
type UpdateBudgetRequest struct {
	Department      string  `json:"department"`
	TotalBudget     float64 `json:"totalBudget"`
	AllocatedAmount float64 `json:"allocatedAmount"`
}

// BudgetResponse represents a budget in responses
type BudgetResponse struct {
	ID              string           `json:"id"`
	BudgetCode      string           `json:"budgetCode"`
	OwnerID         string           `json:"ownerId"`
	OwnerName       string           `json:"ownerName"`
	Department      string           `json:"department"`
	Status          string           `json:"status"`
	FiscalYear      string           `json:"fiscalYear"`
	TotalBudget     float64          `json:"totalBudget"`
	AllocatedAmount float64          `json:"allocatedAmount"`
	RemainingAmount float64          `json:"remainingAmount"`
	ApprovalStage   int              `json:"approvalStage"`
	ApprovalHistory []ApprovalRecord `json:"approvalHistory"`
	CreatedAt       time.Time        `json:"createdAt"`
	UpdatedAt       time.Time        `json:"updatedAt"`
}

// ================== PURCHASE ORDER TYPES ==================

// CreatePurchaseOrderRequest represents a PO creation request
type CreatePurchaseOrderRequest struct {
	VendorID          string        `json:"vendorId" validate:"required"`
	Items             []POItem      `json:"items" validate:"required,min=1"`
	TotalAmount       float64       `json:"totalAmount" validate:"required,gt=0"`
	Currency          string        `json:"currency" validate:"required"`
	DeliveryDate      time.Time     `json:"deliveryDate" validate:"required"`
	LinkedRequisition string        `json:"linkedRequisition"`
}

// UpdatePurchaseOrderRequest represents a PO update request
type UpdatePurchaseOrderRequest struct {
	VendorID     string    `json:"vendorId"`
	Items        []POItem  `json:"items"`
	TotalAmount  float64   `json:"totalAmount"`
	Currency     string    `json:"currency"`
	DeliveryDate time.Time `json:"deliveryDate"`
}

// POItem represents an item in a purchase order
type POItem struct {
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unitPrice"`
	Amount      float64 `json:"amount"`
}

// PurchaseOrderResponse represents a PO in responses
type PurchaseOrderResponse struct {
	ID              string           `json:"id"`
	PONumber        string           `json:"poNumber"`
	VendorID        string           `json:"vendorId"`
	VendorName      string           `json:"vendorName"`
	Status          string           `json:"status"`
	Items           []POItem         `json:"items"`
	TotalAmount     float64          `json:"totalAmount"`
	Currency        string           `json:"currency"`
	DeliveryDate    time.Time        `json:"deliveryDate"`
	ApprovalStage   int              `json:"approvalStage"`
	ApprovalHistory []ApprovalRecord `json:"approvalHistory"`
	LinkedRequisition string         `json:"linkedRequisition"`
	CreatedAt       time.Time        `json:"createdAt"`
	UpdatedAt       time.Time        `json:"updatedAt"`
}

// ================== PAYMENT VOUCHER TYPES ==================

// CreatePaymentVoucherRequest represents a payment voucher creation request
type CreatePaymentVoucherRequest struct {
	VendorID      string `json:"vendorId" validate:"required"`
	InvoiceNumber string `json:"invoiceNumber" validate:"required"`
	Amount        float64 `json:"amount" validate:"required,gt=0"`
	Currency      string `json:"currency" validate:"required"`
	PaymentMethod string `json:"paymentMethod" validate:"required,oneof=bank_transfer check cash"`
	GLCode        string `json:"glCode" validate:"required"`
	Description   string `json:"description" validate:"required,min=10"`
	LinkedPO      string `json:"linkedPO"`
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
	VoucherNumber   string           `json:"voucherNumber"`
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
	PONumber  string         `json:"poNumber" validate:"required"`
	Items     []GRNItem      `json:"items" validate:"required,min=1"`
	ReceivedBy string        `json:"receivedBy" validate:"required"`
}

// UpdateGRNRequest represents a GRN update request
type UpdateGRNRequest struct {
	Items      []GRNItem      `json:"items"`
	ReceivedBy string         `json:"receivedBy"`
	QualityIssues []QualityIssue `json:"qualityIssues"`
}

// GRNItem represents an item in a GRN
type GRNItem struct {
	Description    string  `json:"description"`
	QuantityOrdered int     `json:"quantityOrdered"`
	QuantityReceived int    `json:"quantityReceived"`
	Variance       int     `json:"variance"`
	Condition      string  `json:"condition"` // good, damaged, defective
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
	ID              string           `json:"id"`
	GRNNumber       string           `json:"grnNumber"`
	PONumber        string           `json:"poNumber"`
	Status          string           `json:"status"`
	ReceivedDate    time.Time        `json:"receivedDate"`
	ReceivedBy      string           `json:"receivedBy"`
	Items           []GRNItem        `json:"items"`
	QualityIssues   []QualityIssue   `json:"qualityIssues"`
	ApprovalStage   int              `json:"approvalStage"`
	ApprovalHistory []ApprovalRecord `json:"approvalHistory"`
	CreatedAt       time.Time        `json:"createdAt"`
	UpdatedAt       time.Time        `json:"updatedAt"`
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

// ApprovalRecord represents an approval in the history
type ApprovalRecord struct {
	ApproverID   string    `json:"approverId"`
	ApproverName string    `json:"approverName"`
	Status       string    `json:"status"` // approved, rejected
	Comments     string    `json:"comments"`
	Signature    string    `json:"signature"`
	ApprovedAt   time.Time `json:"approvedAt"`
}

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

// ListResponse represents a paginated list response
type ListResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Total   int64       `json:"total"`
	Page    int         `json:"page"`
	Limit   int         `json:"limit"`
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
