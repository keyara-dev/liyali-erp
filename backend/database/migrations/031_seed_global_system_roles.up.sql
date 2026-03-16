-- Migration 031: Seed global system roles
-- EnsureGlobalSystemRoles() runs at app startup but after a db-reset the server
-- is not restarted, leaving organization_roles empty. Seed them here so they are
-- always present after any migration run.

INSERT INTO organization_roles (name, description, is_system_role, permissions, active)
VALUES
(
    'admin',
    'Full administrative access',
    true,
    '["requisition:view","requisition:create","requisition:edit","requisition:delete","requisition:approve","requisition:reject","budget:view","budget:create","budget:edit","budget:delete","budget:approve","budget:reject","purchase_order:view","purchase_order:create","purchase_order:edit","purchase_order:delete","purchase_order:approve","purchase_order:reject","payment_voucher:view","payment_voucher:create","payment_voucher:edit","payment_voucher:delete","payment_voucher:approve","payment_voucher:reject","grn:view","grn:create","grn:edit","grn:delete","vendor:view","vendor:create","vendor:edit","vendor:delete","category:view","category:create","category:edit","category:delete","organization:view","organization:edit","organization:manage_users","organization:manage_workflows","analytics:view","audit_log:view"]'::jsonb,
    true
),
(
    'approver',
    'Can approve documents',
    true,
    '["requisition:view","requisition:approve","requisition:reject","budget:view","budget:approve","budget:reject","purchase_order:view","purchase_order:approve","purchase_order:reject","payment_voucher:view","payment_voucher:approve","payment_voucher:reject","grn:view","vendor:view","category:view"]'::jsonb,
    true
),
(
    'requester',
    'Can create and manage own requests',
    true,
    '["requisition:view","requisition:create","requisition:edit","budget:view","budget:create","budget:edit","vendor:view","category:view"]'::jsonb,
    true
),
(
    'finance',
    'Finance team — manage and approve budgets, purchase orders, and payment vouchers',
    true,
    '["requisition:view","budget:view","budget:create","budget:edit","budget:approve","budget:reject","purchase_order:view","purchase_order:create","purchase_order:edit","purchase_order:approve","purchase_order:reject","payment_voucher:view","payment_voucher:create","payment_voucher:edit","payment_voucher:approve","payment_voucher:reject","vendor:view","category:view","analytics:view","audit_log:view"]'::jsonb,
    true
),
(
    'viewer',
    'Read-only access',
    true,
    '["requisition:view","budget:view","purchase_order:view","payment_voucher:view","grn:view","vendor:view","category:view"]'::jsonb,
    true
)
ON CONFLICT (name) WHERE organization_id IS NULL AND is_system_role = true DO UPDATE SET
    description = EXCLUDED.description,
    permissions = EXCLUDED.permissions,
    active      = EXCLUDED.active,
    updated_at  = CURRENT_TIMESTAMP;
