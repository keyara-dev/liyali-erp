package models

// DocumentStatus values written and read across the PO/GRN/PV/REQ/Budget
// lifecycle. These strings ARE the wire format — stored in the database, sent
// to the frontend, and compared by handlers after a strings.ToUpper pass.
//
// Do not rename without a migration; do not introduce lowercase/mixed-case
// variants. If you need a new state, add a constant here first so every call
// site picks it up.
const (
	StatusDraft     = "DRAFT"
	StatusPending   = "PENDING"
	StatusApproved  = "APPROVED"
	StatusRejected  = "REJECTED"
	StatusRevision  = "REVISION"
	StatusPaid      = "PAID"      // PV only
	StatusCompleted = "COMPLETED" // GRN only (post-confirm)
	StatusFulfilled = "FULFILLED" // PO only (reserved, see workflow_state_machine.go)
	StatusCancelled = "CANCELLED"
)

// PO delivery_status: tracks physical receipt independent of workflow status.
const (
	DeliveryStatusNotDelivered       = "NOT_DELIVERED"
	DeliveryStatusPartiallyDelivered = "PARTIALLY_DELIVERED"
	DeliveryStatusFullyDelivered     = "FULLY_DELIVERED"
)

// PaymentVoucher-scoped alias for clarity at call sites.
const PaymentVoucherStatusPaid = StatusPaid

// WorkflowTask.Kind values. "approval" is the default for every task created
// prior to this column's introduction and for all approval-stage tasks going
// forward. "payment_execution" is created as a post-approval side-effect task
// on PaymentVouchers so the PAID transition has an audit trail (claim +
// signature + actor) instead of being a direct endpoint flip.
const (
	TaskKindApproval         = "approval"
	TaskKindPaymentExecution = "payment_execution"
)
