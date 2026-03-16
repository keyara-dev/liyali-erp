-- ============================================================================
-- ROLLBACK: 004_seed_data
-- Delete in FK-safe order (most-dependent first)
-- ============================================================================

DELETE FROM workflow_tasks        WHERE organization_id IN ('org-demo-001','org-enterprise-001');
DELETE FROM workflow_assignments  WHERE organization_id IN ('org-demo-001','org-enterprise-001');
DELETE FROM stage_approval_records WHERE organization_id IN ('org-demo-001','org-enterprise-001');
DELETE FROM task_assignment_history WHERE organization_id IN ('org-demo-001','org-enterprise-001');
DELETE FROM goods_received_notes  WHERE organization_id IN ('org-demo-001','org-enterprise-001');
DELETE FROM payment_vouchers      WHERE organization_id IN ('org-demo-001','org-enterprise-001');
DELETE FROM purchase_orders       WHERE organization_id IN ('org-demo-001','org-enterprise-001');
DELETE FROM budgets               WHERE organization_id IN ('org-demo-001','org-enterprise-001');
DELETE FROM requisitions          WHERE organization_id IN ('org-demo-001','org-enterprise-001');
DELETE FROM vendors               WHERE organization_id IN ('org-demo-001','org-enterprise-001');
DELETE FROM categories            WHERE organization_id IN ('org-demo-001','org-enterprise-001');
DELETE FROM workflow_defaults     WHERE organization_id IN ('org-demo-001','org-enterprise-001');
DELETE FROM workflows             WHERE organization_id IN ('org-demo-001','org-enterprise-001');
DELETE FROM user_organization_roles WHERE organization_id IN ('org-demo-001','org-enterprise-001');
DELETE FROM organization_members  WHERE organization_id IN ('org-demo-001','org-enterprise-001');
DELETE FROM organization_departments WHERE organization_id IN ('org-demo-001','org-enterprise-001');
DELETE FROM organization_settings WHERE organization_id IN ('org-demo-001','org-enterprise-001');
DELETE FROM subscription_audit_logs WHERE organization_id IN ('org-demo-001','org-enterprise-001');
DELETE FROM organization_subscriptions WHERE organization_id IN ('org-demo-001','org-enterprise-001');
DELETE FROM organization_roles    WHERE organization_id IS NULL AND is_system_role = true;
DELETE FROM organizations         WHERE id IN ('org-demo-001','org-enterprise-001');
DELETE FROM users                 WHERE id IN (
    'user-super-admin-001','user-admin-001','user-requester-001',
    'user-approver-001','user-finance-001','user-manager-001','user-viewer-001'
);
