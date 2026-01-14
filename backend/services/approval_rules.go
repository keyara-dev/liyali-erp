package services

import (
	"encoding/json"
	"fmt"

	"github.com/liyali/liyali-gateway/logging"
	"github.com/liyali/liyali-gateway/models"
	"gorm.io/gorm"
)

// ApprovalRule defines routing rules for document approval
type ApprovalRule struct {
	ID              string `gorm:"primaryKey" json:"id"`
	DocumentType    string `json:"documentType"` // requisition, budget, po, pv, grn
	Department      string `json:"department"`   // department affected, or "*" for all
	AmountRange     string `json:"amountRange"`  // low, medium, high (thresholds)
	Priority        string `json:"priority"`     // low, medium, high, or "*" for all
	RequiredStages  int    `json:"requiredStages"`
	ApprovalChain   string `json:"approvalChain"`  // JSON array of role names
	CanSkipStages   bool   `json:"canSkipStages"` // Can approvers skip stages
	RequiresFinance bool   `json:"requiresFinance"`
	CreatedAt       string `json:"createdAt"`
}

// ApprovalRoutingService handles routing logic for documents
type ApprovalRoutingService struct {
	db *gorm.DB
}

// NewApprovalRoutingService creates a new approval routing service
func NewApprovalRoutingService(db *gorm.DB) *ApprovalRoutingService {
	return &ApprovalRoutingService{db: db}
}

// GetApproversForDocument returns the list of approvers for a document
func (ars *ApprovalRoutingService) GetApproversForDocument(docType string, department string, amount float64, priority string) ([]string, error) {
	// Determine amount range
	amountRange := ars.getAmountRange(amount)

	// Find matching approval rule
	rule, err := ars.findApprovalRule(docType, department, amountRange, priority)
	if err != nil {
		logging.WithFields(map[string]interface{}{
			"operation":    "find_approval_rule",
			"doc_type":     docType,
			"department":   department,
			"amount_range": amountRange,
			"priority":     priority,
		}).WithError(err).Error("failed_to_find_approval_rule")
		return nil, fmt.Errorf("no approval rule found for document type %s", docType)
	}

	// Parse approval chain from JSON
	var approvalChain []string
	err = json.Unmarshal([]byte(rule.ApprovalChain), &approvalChain)
	if err != nil {
		logging.WithFields(map[string]interface{}{
			"operation": "unmarshal_approval_chain",
			"rule_id":   rule.ID,
		}).WithError(err).Error("failed_to_unmarshal_approval_chain")
		return nil, fmt.Errorf("invalid approval chain configuration")
	}

	// Get actual users matching the roles
	approvers, err := ars.getUsersByRoles(approvalChain)
	if err != nil {
		logging.WithFields(map[string]interface{}{
			"operation":      "get_users_by_roles",
			"approval_chain": approvalChain,
		}).WithError(err).Error("failed_to_get_users_by_roles")
		return nil, fmt.Errorf("could not find approvers for roles")
	}

	return approvers, nil
}

// RouteDocumentToApprovers creates approval tasks for a document
// DEPRECATED: This function is part of the legacy approval system.
// Use WorkflowExecutionService.StartWorkflow() instead for new implementations.
func (ars *ApprovalRoutingService) RouteDocumentToApprovers(documentID, docType, department string, amount float64, priority string) error {
	// This function is deprecated and should not be used
	// The workflow system handles approval routing automatically
	logging.WithFields(map[string]interface{}{
		"operation":   "route_document_to_approvers",
		"document_id": documentID,
	}).Warn("deprecated_function_called_use_workflow_system_instead")
	
	return fmt.Errorf("deprecated: use WorkflowExecutionService.StartWorkflow() instead")
}

// getAmountRange categorizes amount into low, medium, high
func (ars *ApprovalRoutingService) getAmountRange(amount float64) string {
	if amount < 10000 {
		return "low"
	} else if amount < 50000 {
		return "medium"
	}
	return "high"
}

// findApprovalRule finds the matching rule for a document
func (ars *ApprovalRoutingService) findApprovalRule(docType, department, amountRange, priority string) (*ApprovalRule, error) {
	var rule ApprovalRule

	// Try to find exact match first
	err := ars.db.Where(
		"document_type = ? AND (department = ? OR department = '*') AND amount_range = ? AND (priority = ? OR priority = '*')",
		docType, department, amountRange, priority,
	).First(&rule).Error

	if err == gorm.ErrRecordNotFound {
		// Try with wildcard department
		err = ars.db.Where(
			"document_type = ? AND department = '*' AND amount_range = ? AND (priority = ? OR priority = '*')",
			docType, amountRange, priority,
		).First(&rule).Error
	}

	if err != nil {
		return nil, err
	}

	return &rule, nil
}

// getUsersByRoles fetches users with specific roles
func (ars *ApprovalRoutingService) getUsersByRoles(roles []string) ([]string, error) {
	var users []models.User
	var approverIDs []string

	// Find all users with matching roles
	if err := ars.db.Where("role IN ?", roles).Where("active = ?", true).Find(&users).Error; err != nil {
		logging.WithFields(map[string]interface{}{
			"operation": "fetch_users_by_roles",
			"roles":     roles,
		}).WithError(err).Error("failed_to_fetch_users_by_roles")
		return nil, err
	}

	for _, user := range users {
		approverIDs = append(approverIDs, user.ID)
	}

	return approverIDs, nil
}

// CreateDefaultApprovalRules creates default routing rules
func (ars *ApprovalRoutingService) CreateDefaultApprovalRules() error {
	rules := []ApprovalRule{
		{
			ID:             "rule-req-low",
			DocumentType:   "requisition",
			Department:     "*",
			AmountRange:    "low",
			Priority:       "*",
			RequiredStages: 2,
			ApprovalChain:  `["approver", "finance"]`,
			CanSkipStages:  false,
			RequiresFinance: true,
		},
		{
			ID:             "rule-req-medium",
			DocumentType:   "requisition",
			Department:     "*",
			AmountRange:    "medium",
			Priority:       "*",
			RequiredStages: 3,
			ApprovalChain:  `["approver", "finance", "admin"]`,
			CanSkipStages:  false,
			RequiresFinance: true,
		},
		{
			ID:             "rule-req-high",
			DocumentType:   "requisition",
			Department:     "*",
			AmountRange:    "high",
			Priority:       "*",
			RequiredStages: 4,
			ApprovalChain:  `["approver", "finance", "admin", "admin"]`,
			CanSkipStages:  false,
			RequiresFinance: true,
		},
		{
			ID:             "rule-po-low",
			DocumentType:   "po",
			Department:     "*",
			AmountRange:    "low",
			Priority:       "*",
			RequiredStages: 2,
			ApprovalChain:  `["finance", "approver"]`,
			CanSkipStages:  false,
			RequiresFinance: true,
		},
		{
			ID:             "rule-pv-all",
			DocumentType:   "pv",
			Department:     "*",
			AmountRange:    "*",
			Priority:       "*",
			RequiredStages: 2,
			ApprovalChain:  `["finance", "admin"]`,
			CanSkipStages:  false,
			RequiresFinance: true,
		},
		{
			ID:             "rule-grn-all",
			DocumentType:   "grn",
			Department:     "*",
			AmountRange:    "*",
			Priority:       "*",
			RequiredStages: 1,
			ApprovalChain:  `["approver"]`,
			CanSkipStages:  false,
			RequiresFinance: false,
		},
		{
			ID:             "rule-budget-all",
			DocumentType:   "budget",
			Department:     "*",
			AmountRange:    "*",
			Priority:       "*",
			RequiredStages: 2,
			ApprovalChain:  `["finance", "admin"]`,
			CanSkipStages:  false,
			RequiresFinance: true,
		},
	}

	for _, rule := range rules {
		// Check if rule already exists
		var count int64
		if err := ars.db.Model(&ApprovalRule{}).Where("id = ?", rule.ID).Count(&count).Error; err != nil {
			return err
		}

		if count == 0 {
			if err := ars.db.Create(&rule).Error; err != nil {
				logging.WithFields(map[string]interface{}{
					"operation":    "create_approval_rule",
					"doc_type":     rule.DocumentType,
					"department":   rule.Department,
				}).WithError(err).Error("failed_to_create_approval_rule")
				return err
			}
		}
	}

	return nil
}
