package utils

import (
	"encoding/json"
	"strings"

	"github.com/liyali/liyali-gateway/models"
	"github.com/liyali/liyali-gateway/types"
	"gorm.io/gorm"
)

// DocumentScope represents what level of document access a user has.
type DocumentScope struct {
	CanViewAll        bool     // admin/super_admin/manager/finance/approver: no filtering
	IsProcurement     bool     // procurement role: sees only procurement-chain documents
	HideDirectPayment bool     // true for procurement role (and non-privileged non-finance users): hides direct_payment routing type
	UserID            string   // used for owner/involvement filter when neither flag is set
	OrgID             string
	UserRole          string
	OrgRoleIDs        []string // custom org role UUIDs for the user (for UUID-stored assigned_role matching)
}

// privilegeRoles grant unrestricted document visibility (CanViewAll = true).
var privilegeRoles = []string{
	"admin", "super_admin", "manager", "department_manager", "finance", "approver",
}

// approvalPermissions — any of these on an org-role grants CanViewAll.
var approvalPermissions = []string{
	"requisition.approve", "approval.approve", "budget.approve",
	"purchase_order.approve", "payment_voucher.approve", "grn.approve",
}

// GetDocumentScope determines the document access level for the given user.
// It mirrors the permission-check logic from approval_handler.go.
func GetDocumentScope(db *gorm.DB, userID, userRole, orgID string) DocumentScope {
	scope := DocumentScope{
		UserID:   userID,
		OrgID:    orgID,
		UserRole: userRole,
	}

	// 1. Built-in privileged role → unrestricted
	lower := strings.ToLower(userRole)
	for _, r := range privilegeRoles {
		if lower == r {
			scope.CanViewAll = true
			return scope
		}
	}

	// 2. Org-role with approval OR purchase_order.create permission
	var userOrgRoles []models.UserOrganizationRole
	if err := db.Where(
		"user_id = ? AND organization_id = ? AND active = ?",
		userID, orgID, true,
	).Find(&userOrgRoles).Error; err == nil {
		// Collect custom org role UUIDs for ApplyToQuery
		for _, uor := range userOrgRoles {
			scope.OrgRoleIDs = append(scope.OrgRoleIDs, uor.RoleID.String())
		}

		for _, uor := range userOrgRoles {
			var orgRole models.OrganizationRole
			if err := db.Where("id = ? AND active = ?", uor.RoleID, true).
				First(&orgRole).Error; err != nil {
				continue
			}
			var permissions []string
			if err := json.Unmarshal(orgRole.Permissions, &permissions); err != nil {
				continue
			}
			for _, perm := range permissions {
				pLower := strings.ToLower(perm)
				for _, ap := range approvalPermissions {
					if pLower == ap {
						scope.CanViewAll = true
						return scope
					}
				}
				if pLower == "purchase_order.create" {
					scope.IsProcurement = true
				}
			}
		}
	}

	// 3. Built-in procurement role
	if lower == "procurement" {
		scope.IsProcurement = true
	}

	// 4. HideDirectPayment: procurement-role users must never see the direct_payment
	// chain (requisitions, POs, PVs, GRNs with routing_type=direct_payment).
	// finance and admin/privileged roles (CanViewAll) can see everything — they
	// exit above via the privilegeRoles loop, so CanViewAll is already true if
	// applicable.  All remaining users who are procurement (or similar limited roles)
	// should not see direct_payment documents.
	if !scope.CanViewAll {
		scope.HideDirectPayment = true
	}

	return scope
}

// ApplyToQuery applies document scope filtering to a GORM query that already has
// the organization_id filter applied.
//
//   - ownerField:      primary ownership column (e.g. "requester_id", "created_by")
//   - entityType:      workflow entity_type string (e.g. "requisition", "purchase_order")
//   - extraOwnerField: optional second owner column (e.g. "received_by" for GRNs; pass "" to skip)
//
// When CanViewAll is true only the HideDirectPayment filter (if set) is applied.
// When IsProcurement is true only the HideDirectPayment filter is applied — the
// caller is responsible for adding any further procurement-specific subquery.
// Otherwise both the owner+workflow-involvement filter and HideDirectPayment are appended.
func (s DocumentScope) ApplyToQuery(query *gorm.DB, ownerField, entityType, extraOwnerField string) *gorm.DB {
	// Apply routing_type exclusion first — this is independent of owner/involvement.
	if s.HideDirectPayment {
		switch entityType {
		case "requisition":
			query = query.Where("routing_type != ?", "direct_payment")
		case "purchase_order":
			query = query.Where("routing_type != ?", "direct_payment")
		case "payment_voucher":
			query = query.Where("routing_type != ?", "direct_payment")
		case "grn":
			// GRN has no routing_type column; infer via the linked PO.
			query = query.Where(
				"EXISTS (SELECT 1 FROM purchase_orders po WHERE po.document_number = po_document_number AND po.routing_type != ?)",
				"direct_payment",
			)
		}
		// budget: unaffected — direct payments still consume budgets.
	}

	if s.CanViewAll || s.IsProcurement {
		return query
	}

	// Build the role-matching clause for the workflow_tasks subquery.
	// Plain role name covers current data; UUID list covers custom org roles stored as UUIDs.
	roleClause := "LOWER(assigned_role) = LOWER(?)"
	roleArgs := []interface{}{s.UserRole}
	if len(s.OrgRoleIDs) > 0 {
		roleClause += " OR assigned_role IN (?)"
		roleArgs = append(roleArgs, s.OrgRoleIDs)
	}

	involvedSQL := "id IN (SELECT entity_id FROM workflow_tasks" +
		" WHERE organization_id = ? AND entity_type = ?" +
		" AND (assigned_user_id = ? OR " + roleClause + " OR claimed_by = ?))"
	involvedArgs := append([]interface{}{s.OrgID, entityType, s.UserID}, roleArgs...)
	involvedArgs = append(involvedArgs, s.UserID)

	if extraOwnerField != "" {
		allArgs := append([]interface{}{s.UserID, s.UserID}, involvedArgs...)
		return query.Where(
			ownerField+" = ? OR "+extraOwnerField+" = ? OR "+involvedSQL,
			allArgs...,
		)
	}
	allArgs := append([]interface{}{s.UserID}, involvedArgs...)
	return query.Where(
		ownerField+" = ? OR "+involvedSQL,
		allArgs...,
	)
}

// GetDocumentApprovalHistory fetches live approval history for a document.
// Primary source: stage_approval_records joined through workflow_tasks.
// Fallback source: WorkflowAssignment.StageHistory (used when stage_approval_records is empty,
// e.g. on the public verify endpoint which has no auth context).
func GetDocumentApprovalHistory(db *gorm.DB, entityID, entityType string) []types.ApprovalRecord {
	// PRIMARY: stage_approval_records
	var records []models.StageApprovalRecord
	db.Joins("JOIN workflow_tasks ON workflow_tasks.id = stage_approval_records.workflow_task_id").
		Where("workflow_tasks.entity_id = ? AND workflow_tasks.entity_type = ?", entityID, entityType).
		Order("stage_approval_records.stage_number ASC, stage_approval_records.approved_at ASC").
		Find(&records)

	if len(records) > 0 {
		result := make([]types.ApprovalRecord, 0, len(records))
		for _, r := range records {
			stageNum := r.StageNumber
			role := r.ApproverRole
			result = append(result, types.ApprovalRecord{
				ApproverID:   r.ApproverID,
				ApproverName: r.ApproverName,
				Status:       r.Action,
				Comments:     r.Comments,
				Signature:    r.Signature,
				ApprovedAt:   r.ApprovedAt,
				ManNumber:    r.ManNumber,
				Position:     r.Position,
				StageNumber:  &stageNum,
				AssignedRole: &role,
			})
		}
		return result
	}

	// FALLBACK: WorkflowAssignment.StageHistory (most recent assignment for this document)
	var assignment models.WorkflowAssignment
	if err := db.Where("entity_id = ? AND LOWER(entity_type) = LOWER(?)", entityID, entityType).
		Order("created_at DESC").
		First(&assignment).Error; err != nil {
		return nil
	}
	stageHistory, err := assignment.GetStageHistory()
	if err != nil || len(stageHistory) == 0 {
		return nil
	}
	result := make([]types.ApprovalRecord, 0, len(stageHistory))
	for _, e := range stageHistory {
		stageNum := e.StageNumber
		stageName := e.StageName
		role := e.ApproverRole
		result = append(result, types.ApprovalRecord{
			ApproverID:   e.ApproverID,
			ApproverName: e.ApproverName,
			Status:       e.Action,
			Comments:     e.Comments,
			Signature:    e.Signature,
			ApprovedAt:   e.ExecutedAt,
			StageNumber:  &stageNum,
			StageName:    &stageName,
			AssignedRole: &role,
		})
	}
	return result
}
