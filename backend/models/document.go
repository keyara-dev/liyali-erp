package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Document represents a generic document that can be any business document type
type Document struct {
	ID             uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	OrganizationID string         `json:"organizationId" gorm:"not null;index"`
	DocumentType   string         `json:"documentType" gorm:"not null;index"` // REQUISITION, BUDGET, PURCHASE_ORDER, etc.
	DocumentNumber string         `json:"documentNumber" gorm:"not null;unique_index"`
	Title          string         `json:"title" gorm:"not null"`
	Description    *string        `json:"description"`
	Status         string         `json:"status" gorm:"not null;default:'draft';index"` // draft, submitted, approved, rejected
	Amount         *float64       `json:"amount"`
	Currency       *string        `json:"currency" gorm:"default:'USD'"`
	Department     *string        `json:"department" gorm:"index"`
	CreatedBy      string         `json:"createdBy" gorm:"not null;index"`
	UpdatedBy      *string        `json:"updatedBy"`
	WorkflowID     *uuid.UUID     `json:"workflowId" gorm:"type:uuid"`
	Data           datatypes.JSON `json:"data" gorm:"type:jsonb"` // Type-specific fields as JSONB
	Metadata       datatypes.JSON `json:"metadata" gorm:"type:jsonb"` // Additional metadata
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `json:"deletedAt" gorm:"index"`

	// Relationships
	Organization *Organization `json:"organization,omitempty" gorm:"foreignKey:OrganizationID"`
	Creator      *User         `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
	Updater      *User         `json:"updater,omitempty" gorm:"foreignKey:UpdatedBy"`
	Workflow     *Workflow     `json:"workflow,omitempty" gorm:"foreignKey:WorkflowID"`
}

// TableName returns the table name for the Document model
func (Document) TableName() string {
	return "documents"
}

// BeforeCreate generates document number if not provided
func (d *Document) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	
	if d.DocumentNumber == "" {
		d.DocumentNumber = d.generateDocumentNumber()
	}
	
	return nil
}

// generateDocumentNumber generates a document number based on type and timestamp
func (d *Document) generateDocumentNumber() string {
	prefix := d.getDocumentPrefix()
	timestamp := time.Now().Format("20060102")
	return prefix + "-" + timestamp + "-" + d.ID.String()[:8]
}

// getDocumentPrefix returns the prefix for document numbers based on type
func (d *Document) getDocumentPrefix() string {
	prefixes := map[string]string{
		"REQUISITION":     "REQ",
		"BUDGET":          "BUD",
		"PURCHASE_ORDER":  "PO",
		"PAYMENT_VOUCHER": "PV",
		"GRN":             "GRN",
		"CATEGORY":        "CAT",
		"VENDOR":          "VEN",
	}
	
	if prefix, exists := prefixes[d.DocumentType]; exists {
		return prefix
	}
	
	return "DOC" // Default prefix
}

// IsEditable checks if the document can be edited based on its status
func (d *Document) IsEditable() bool {
	s := strings.ToUpper(d.Status)
	return s == "DRAFT" || s == "REJECTED"
}

// CanBeSubmitted checks if the document can be submitted for approval
func (d *Document) CanBeSubmitted() bool {
	s := strings.ToUpper(d.Status)
	return s == "DRAFT" || s == "REJECTED"
}

// CanBeApproved checks if the document can be approved
func (d *Document) CanBeApproved() bool {
	return strings.ToUpper(d.Status) == "SUBMITTED"
}

// DocumentSearchResult represents a document search result with highlighting
type DocumentSearchResult struct {
	Document
	Relevance float64 `json:"relevance"`
	Matches   []string `json:"matches,omitempty"` // Fields that matched the search
}

// DocumentFilter represents filters for document queries
type DocumentFilter struct {
	DocumentNumber string    `json:"documentNumber,omitempty"` // Exact document number match
	DocumentTypes  []string  `json:"documentTypes,omitempty"`
	Statuses       []string  `json:"statuses,omitempty"`
	Departments    []string  `json:"departments,omitempty"`
	CreatedBy      []string  `json:"createdBy,omitempty"`
	DateFrom       *time.Time `json:"dateFrom,omitempty"`
	DateTo         *time.Time `json:"dateTo,omitempty"`
	AmountMin      *float64  `json:"amountMin,omitempty"`
	AmountMax      *float64  `json:"amountMax,omitempty"`
	Search         string    `json:"search,omitempty"` // Full-text search
}

// DocumentStats represents document statistics
type DocumentStats struct {
	TotalDocuments     int64                    `json:"totalDocuments"`
	DocumentsByType    map[string]int64         `json:"documentsByType"`
	DocumentsByStatus  map[string]int64         `json:"documentsByStatus"`
	DocumentsByDept    map[string]int64         `json:"documentsByDepartment"`
	RecentDocuments    int64                    `json:"recentDocuments"` // Last 7 days
	PendingApprovals   int64                    `json:"pendingApprovals"`
	TotalValue         float64                  `json:"totalValue"`
	AverageValue       float64                  `json:"averageValue"`
}