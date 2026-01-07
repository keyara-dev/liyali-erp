import {
  UserRole,
  Permission,
  WorkflowDocumentType,
  WorkflowStep,
} from '@/types/workflow';

// All Available Permissions - Consolidated Permission System
export const ALL_PERMISSIONS: Permission[] = [
  'view_draft',
  'edit_draft',
  'submit_document',
  'approve_document',
  'reject_document',
  'reassign_approver',
  'view_attachments',
  'add_attachments',
  'view_comments',
  'add_comments',
  'view_audit_log',
  'manage_approvers',
  'manage_workflows',
];

// Permission Descriptions for UI
export const PERMISSION_DESCRIPTIONS: Record<Permission, string> = {
  view_draft: 'View draft documents',
  edit_draft: 'Edit draft documents',
  submit_document: 'Submit documents for approval',
  approve_document: 'Approve documents',
  reject_document: 'Reject documents',
  reassign_approver: 'Reassign approvers',
  view_attachments: 'View attachments',
  add_attachments: 'Upload attachments',
  view_comments: 'View approval comments',
  add_comments: 'Add approval comments',
  view_audit_log: 'View audit logs',
  manage_approvers: 'Manage approvers',
  manage_workflows: 'Manage workflow configuration',
};

// Custom Role Type for user-defined roles
export type CustomRole = {
  id: string;
  name: string;
  description: string;
  permissions: Permission[];
  isBuiltIn: boolean;
  createdAt?: Date;
  updatedAt?: Date;
};

// Default Built-in Role-Permission Mapping - aligned with backend UserRole enum
export const DEFAULT_ROLE_PERMISSIONS: Record<UserRole, Permission[]> = {
  admin: [
    'view_draft',
    'edit_draft',
    'submit_document',
    'approve_document',
    'reject_document',
    'reassign_approver',
    'view_attachments',
    'add_attachments',
    'view_comments',
    'add_comments',
    'view_audit_log',
    'manage_approvers',
    'manage_workflows',
  ],
  superadmin: [
    'view_draft',
    'edit_draft',
    'submit_document',
    'approve_document',
    'reject_document',
    'reassign_approver',
    'view_attachments',
    'add_attachments',
    'view_comments',
    'add_comments',
    'view_audit_log',
    'manage_approvers',
    'manage_workflows',
  ],
  approver: [
    'view_draft',
    'approve_document',
    'reject_document',
    'reassign_approver',
    'view_attachments',
    'view_comments',
    'add_comments',
    'view_audit_log',
  ],
  finance: [
    'approve_document',
    'reject_document',
    'reassign_approver',
    'view_attachments',
    'view_comments',
    'add_comments',
    'view_audit_log',
  ],
  requester: [
    'view_draft',
    'edit_draft',
    'submit_document',
    'view_attachments',
    'add_attachments',
    'view_comments',
    'add_comments',
  ],
  viewer: [
    'view_audit_log',
    'view_attachments',
    'view_comments',
  ],
  department_manager: [
    'view_draft',
    'edit_draft',
    'submit_document',
    'approve_document',
    'reject_document',
    'reassign_approver',
    'view_attachments',
    'add_attachments',
    'view_comments',
    'add_comments',
    'view_audit_log',
  ],
  finance_manager: [
    'approve_document',
    'reject_document',
    'reassign_approver',
    'view_attachments',
    'view_comments',
    'add_comments',
    'view_audit_log',
    'manage_workflows',
  ],
  finance_officer: [
    'approve_document',
    'reject_document',
    'view_attachments',
    'view_comments',
    'add_comments',
    'view_audit_log',
  ],
  director: [
    'view_draft',
    'edit_draft',
    'submit_document',
    'approve_document',
    'reject_document',
    'reassign_approver',
    'view_attachments',
    'add_attachments',
    'view_comments',
    'add_comments',
    'view_audit_log',
    'manage_approvers',
    'manage_workflows',
  ],
  cfo: [
    'approve_document',
    'reject_document',
    'reassign_approver',
    'view_attachments',
    'view_comments',
    'add_comments',
    'view_audit_log',
    'manage_workflows',
  ],
  compliance_officer: [
    'view_draft',
    'approve_document',
    'reject_document',
    'view_attachments',
    'view_comments',
    'add_comments',
    'view_audit_log',
  ],
  ceo: [
    'view_draft',
    'edit_draft',
    'submit_document',
    'approve_document',
    'reject_document',
    'reassign_approver',
    'view_attachments',
    'add_attachments',
    'view_comments',
    'add_comments',
    'view_audit_log',
    'manage_approvers',
    'manage_workflows',
  ],
};

// In-memory store for custom roles (in production, this would be a database)
const customRolesStore = new Map<string, CustomRole>();

// Initialize with built-in roles as custom roles for reference
function initializeBuiltInRoles() {
  const builtInRoles: CustomRole[] = [
    {
      id: 'role-requester',
      name: 'requester',
      description: 'Users who create and submit documents',
      permissions: DEFAULT_ROLE_PERMISSIONS.requester,
      isBuiltIn: true,
    },
    {
      id: 'role-approver',
      name: 'approver',
      description: 'Users who can approve documents',
      permissions: DEFAULT_ROLE_PERMISSIONS.approver,
      isBuiltIn: true,
    },
    {
      id: 'role-finance',
      name: 'finance',
      description: 'Finance team members who verify documents',
      permissions: DEFAULT_ROLE_PERMISSIONS.finance,
      isBuiltIn: true,
    },
    {
      id: 'role-viewer',
      name: 'viewer',
      description: 'Read-only access to documents',
      permissions: DEFAULT_ROLE_PERMISSIONS.viewer,
      isBuiltIn: true,
    },
    {
      id: 'role-admin',
      name: 'admin',
      description: 'System administrator with full access',
      permissions: DEFAULT_ROLE_PERMISSIONS.admin,
      isBuiltIn: true,
    },
  ];

  builtInRoles.forEach((role) => {
    customRolesStore.set(role.id, role);
  });
}

// Initialize on module load
initializeBuiltInRoles();

// Keep backward compatibility with direct role mapping
export const ROLE_PERMISSIONS: Record<UserRole, Permission[]> = DEFAULT_ROLE_PERMISSIONS;

// Workflow Steps Configuration
export const WORKFLOW_STEPS: Record<WorkflowDocumentType, WorkflowStep[]> = {
  requisition: [
    {
      workflowType: 'requisition',
      stepOrder: 1,
      roleName: 'approver',
      description: 'Department Level Approval',
      isRequired: true,
    },
    {
      workflowType: 'requisition',
      stepOrder: 2,
      roleName: 'finance',
      description: 'Finance Review',
      isRequired: true,
    },
  ],
  budget: [
    {
      workflowType: 'budget',
      stepOrder: 1,
      roleName: 'approver',
      description: 'Budget Review',
      isRequired: true,
    },
    {
      workflowType: 'budget',
      stepOrder: 2,
      roleName: 'admin',
      description: 'Final Approval',
      isRequired: true,
    },
  ],
  purchase_order: [
    {
      workflowType: 'purchase_order',
      stepOrder: 1,
      roleName: 'approver',
      description: 'Department Level Approval',
      isRequired: true,
    },
    {
      workflowType: 'purchase_order',
      stepOrder: 2,
      roleName: 'finance',
      description: 'Finance Review',
      isRequired: true,
    },
  ],
  payment_voucher: [
    {
      workflowType: 'payment_voucher',
      stepOrder: 1,
      roleName: 'approver',
      description: 'Payment Authorization',
      isRequired: true,
    },
    {
      workflowType: 'payment_voucher',
      stepOrder: 2,
      roleName: 'finance',
      description: 'Finance Processing',
      isRequired: true,
    },
  ],
  goods_received_note: [
    {
      workflowType: 'goods_received_note',
      stepOrder: 1,
      roleName: 'approver',
      description: 'Goods Verification',
      isRequired: true,
    },
  ],
  // Additional document types
  pv: [
    {
      workflowType: 'pv',
      stepOrder: 1,
      roleName: 'approver',
      description: 'Payment Authorization',
      isRequired: true,
    },
  ],
  po: [
    {
      workflowType: 'po',
      stepOrder: 1,
      roleName: 'approver',
      description: 'Purchase Order Approval',
      isRequired: true,
    },
  ],
  grn: [
    {
      workflowType: 'grn',
      stepOrder: 1,
      roleName: 'approver',
      description: 'Goods Receipt Verification',
      isRequired: true,
    },
  ],
  // Legacy compatibility
  REQUISITION: [
    {
      workflowType: 'requisition',
      stepOrder: 1,
      roleName: 'approver',
      description: 'Department Level Approval',
      isRequired: true,
    },
  ],
  BUDGET: [
    {
      workflowType: 'budget',
      stepOrder: 1,
      roleName: 'approver',
      description: 'Budget Review',
      isRequired: true,
    },
  ],
  PURCHASE_ORDER: [
    {
      workflowType: 'purchase_order',
      stepOrder: 1,
      roleName: 'approver',
      description: 'Department Level Approval',
      isRequired: true,
    },
  ],
  PAYMENT_VOUCHER: [
    {
      workflowType: 'payment_voucher',
      stepOrder: 1,
      roleName: 'approver',
      description: 'Payment Authorization',
      isRequired: true,
    },
  ],
  GOODS_RECEIVED_NOTE: [
    {
      workflowType: 'goods_received_note',
      stepOrder: 1,
      roleName: 'approver',
      description: 'Goods Verification',
      isRequired: true,
    },
  ],
  PV: [
    {
      workflowType: 'pv',
      stepOrder: 1,
      roleName: 'approver',
      description: 'Payment Authorization',
      isRequired: true,
    },
  ],
  PO: [
    {
      workflowType: 'po',
      stepOrder: 1,
      roleName: 'approver',
      description: 'Purchase Order Approval',
      isRequired: true,
    },
  ],
  GRN: [
    {
      workflowType: 'grn',
      stepOrder: 1,
      roleName: 'approver',
      description: 'Goods Receipt Verification',
      isRequired: true,
    },
  ],
};

// RBAC Utility Functions
export function hasPermission(userRole: UserRole, permission: Permission): boolean {
  return ROLE_PERMISSIONS[userRole]?.includes(permission) ?? false;
}

export function canApproveAtStep(
  userRole: UserRole,
  documentType: WorkflowDocumentType,
  stepOrder: number
): boolean {
  const steps = WORKFLOW_STEPS[documentType];
  const stepRequirement = steps.find((s) => s.stepOrder === stepOrder);

  if (!stepRequirement) return false;

  return (
    stepRequirement.roleName === userRole &&
    hasPermission(userRole, 'approve_document')
  );
}

export function getWorkflowStepsForType(
  documentType: WorkflowDocumentType
): WorkflowStep[] {
  return WORKFLOW_STEPS[documentType] ?? [];
}

export function getApprovalChain(documentType: WorkflowDocumentType): UserRole[] {
  return getWorkflowStepsForType(documentType)
    .sort((a, b) => a.stepOrder - b.stepOrder)
    .map((s) => s.roleName);
}

export function getNextApproverRole(
  documentType: WorkflowDocumentType,
  currentStage: number
): UserRole | null {
  const steps = getWorkflowStepsForType(documentType);
  const nextStep = steps.find((s) => s.stepOrder === currentStage + 1);
  return nextStep?.roleName ?? null;
}

export function isLastApprovalStep(
  documentType: WorkflowDocumentType,
  currentStage: number
): boolean {
  const steps = getWorkflowStepsForType(documentType);
  const lastStep = steps[steps.length - 1];
  return currentStage >= lastStep.stepOrder;
}

// =============== CUSTOM ROLE MANAGEMENT ===============

export function createCustomRole(
  name: string,
  description: string,
  permissions: Permission[]
): CustomRole {
  const id = `role-${Date.now()}-${Math.random().toString(36).substring(2, 11)}`;
  const newRole: CustomRole = {
    id,
    name,
    description,
    permissions,
    isBuiltIn: false,
    createdAt: new Date(),
    updatedAt: new Date(),
  };

  customRolesStore.set(id, newRole);
  return newRole;
}

export function updateCustomRole(
  roleId: string,
  updates: Partial<Omit<CustomRole, 'id' | 'isBuiltIn'>>
): CustomRole | null {
  const role = customRolesStore.get(roleId);
  if (!role || role.isBuiltIn) {
    return null; // Cannot update built-in roles
  }

  const updated: CustomRole = {
    ...role,
    ...updates,
    updatedAt: new Date(),
  };

  customRolesStore.set(roleId, updated);
  return updated;
}

export function deleteCustomRole(roleId: string): boolean {
  const role = customRolesStore.get(roleId);
  if (!role || role.isBuiltIn) {
    return false; // Cannot delete built-in roles
  }

  customRolesStore.delete(roleId);
  return true;
}

export function getCustomRole(roleId: string): CustomRole | null {
  return customRolesStore.get(roleId) ?? null;
}

export function getAllCustomRoles(): CustomRole[] {
  return Array.from(customRolesStore.values());
}

export function getCustomRolesByIsBuiltIn(isBuiltIn: boolean): CustomRole[] {
  return Array.from(customRolesStore.values()).filter(
    (role) => role.isBuiltIn === isBuiltIn
  );
}

export function getPermissionsForCustomRole(roleId: string): Permission[] {
  const role = customRolesStore.get(roleId);
  return role?.permissions ?? [];
}

export function hasCustomRolePermission(
  roleId: string,
  permission: Permission
): boolean {
  const permissions = getPermissionsForCustomRole(roleId);
  return permissions.includes(permission);
}

export function updateCustomRolePermissions(
  roleId: string,
  permissions: Permission[]
): CustomRole | null {
  return updateCustomRole(roleId, { permissions });
}

export function addPermissionToCustomRole(
  roleId: string,
  permission: Permission
): CustomRole | null {
  const role = customRolesStore.get(roleId);
  if (!role) return null;

  const updatedPermissions = [...new Set([...role.permissions, permission])];
  return updateCustomRole(roleId, { permissions: updatedPermissions });
}

export function removePermissionFromCustomRole(
  roleId: string,
  permission: Permission
): CustomRole | null {
  const role = customRolesStore.get(roleId);
  if (!role) return null;

  const updatedPermissions = role.permissions.filter((p) => p !== permission);
  return updateCustomRole(roleId, { permissions: updatedPermissions });
}
