import {
  User,
  WorkflowDocument,
  PurchaseOrder,
  PaymentVoucher,
  RequisitionForm,
  ApprovalLogEntry,
  Attachment,
  Approver,
  WorkflowDocumentType,
  UserRole,
  DocumentStatus,
} from '@/types/workflow';
import { v4 as uuidv4 } from 'uuid';

// Mock Users by Role
export const MOCK_USERS: Record<UserRole, User[]> = {
  REQUESTER: [
    {
      id: 'user-req-1',
      name: 'John Mwale',
      email: 'john.mwale@company.com',
      role: 'REQUESTER',
      department: 'Operations',
    },
    {
      id: 'user-req-2',
      name: 'Sarah Banda',
      email: 'sarah.banda@company.com',
      role: 'REQUESTER',
      department: 'HR',
    },
  ],
  DEPARTMENT_MANAGER: [
    {
      id: 'user-dm-1',
      name: 'James Chileshe',
      email: 'james.chileshe@company.com',
      role: 'DEPARTMENT_MANAGER',
      department: 'Operations',
    },
    {
      id: 'user-dm-2',
      name: 'Maria Chiyanda',
      email: 'maria.chiyanda@company.com',
      role: 'DEPARTMENT_MANAGER',
      department: 'HR',
    },
  ],
  FINANCE_OFFICER: [
    {
      id: 'user-fo-1',
      name: 'Paul Nkosi',
      email: 'paul.nkosi@company.com',
      role: 'FINANCE_OFFICER',
      department: 'Finance',
    },
    {
      id: 'user-fo-2',
      name: 'Grace Mvula',
      email: 'grace.mvula@company.com',
      role: 'FINANCE_OFFICER',
      department: 'Finance',
    },
  ],
  DIRECTOR: [
    {
      id: 'user-dir-1',
      name: 'David Moyo',
      email: 'david.moyo@company.com',
      role: 'DIRECTOR',
      department: 'Operations',
    },
  ],
  CFO: [
    {
      id: 'user-cfo-1',
      name: 'Catherine Phiri',
      email: 'catherine.phiri@company.com',
      role: 'CFO',
      department: 'Finance',
    },
  ],
  COMPLIANCE_OFFICER: [
    {
      id: 'user-co-1',
      name: 'Victor Zulu',
      email: 'victor.zulu@company.com',
      role: 'COMPLIANCE_OFFICER',
      department: 'Legal',
    },
  ],
  ADMIN: [
    {
      id: 'user-admin-1',
      name: 'Admin User',
      email: 'admin@company.com',
      role: 'ADMIN',
      department: 'IT',
    },
  ],
};

// Helper to generate document number
export function generateDocumentNumber(type: WorkflowDocumentType): string {
  const prefix = {
    PURCHASE_ORDER: 'PO',
    PAYMENT_VOUCHER: 'PV',
    REQUISITION: 'REQ',
    GOODS_RECEIVED_NOTE: 'GRN',
  }[type];

  const timestamp = Date.now().toString().slice(-6);
  const random = Math.floor(Math.random() * 1000)
    .toString()
    .padStart(3, '0');

  return `${prefix}-${timestamp}-${random}`;
}

// Mock Purchase Order Factory
export function createMockPurchaseOrder(
  overrides?: Partial<PurchaseOrder>
): PurchaseOrder {
  const id = uuidv4();
  const requester = MOCK_USERS.REQUESTER[0];

  return {
    id,
    type: 'PURCHASE_ORDER',
    documentNumber: generateDocumentNumber('PURCHASE_ORDER'),
    status: 'DRAFT',
    currentStage: 0,
    createdBy: requester.id,
    createdByUser: requester,
    createdAt: new Date(),
    updatedAt: new Date(),
    metadata: {
      vendorName: 'Mitete Supplies Ltd',
      vendorId: 'VENDOR-001',
      items: [
        {
          id: uuidv4(),
          description: 'Office Equipment - Chairs',
          quantity: 15,
          unitCost: 450,
          totalCost: 6750,
        },
        {
          id: uuidv4(),
          description: 'Office Equipment - Desks',
          quantity: 10,
          unitCost: 1200,
          totalCost: 12000,
        },
      ],
      totalAmount: 18750,
      currency: 'ZMW',
      deliveryDate: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000),
      specialInstructions: 'Deliver to main office. Please notify 2 days before.',
    },
    ...overrides,
  };
}

// Mock Payment Voucher Factory
export function createMockPaymentVoucher(
  overrides?: Partial<PaymentVoucher>
): PaymentVoucher {
  const id = uuidv4();
  const requester = MOCK_USERS.REQUESTER[0];

  return {
    id,
    type: 'PAYMENT_VOUCHER',
    documentNumber: generateDocumentNumber('PAYMENT_VOUCHER'),
    status: 'DRAFT',
    currentStage: 0,
    createdBy: requester.id,
    createdByUser: requester,
    createdAt: new Date(),
    updatedAt: new Date(),
    metadata: {
      payeeName: 'Mitete Supplies Ltd',
      payeeId: 'VENDOR-001',
      amount: 18750,
      currency: 'ZMW',
      reason: 'Payment for office equipment - PO-123456-789',
      accountCode: '4001-001',
      department: 'Operations',
    },
    ...overrides,
  };
}

// Mock Requisition Form Factory
export function createMockRequisitionForm(
  overrides?: Partial<RequisitionForm>
): RequisitionForm {
  const id = uuidv4();
  const requester = MOCK_USERS.REQUESTER[0];

  return {
    id,
    type: 'REQUISITION',
    documentNumber: generateDocumentNumber('REQUISITION'),
    status: 'DRAFT',
    currentStage: 0,
    createdBy: requester.id,
    createdByUser: requester,
    createdAt: new Date(),
    updatedAt: new Date(),
    metadata: {
      department: 'Operations',
      requestedFor: 'John Mwale',
      items: [
        {
          id: uuidv4(),
          itemDescription: 'Office Chairs - Ergonomic',
          quantity: 15,
          estimatedCost: 6750,
        },
        {
          id: uuidv4(),
          itemDescription: 'Standing Desks',
          quantity: 10,
          estimatedCost: 12000,
        },
      ],
      justification:
        'Current office furniture is worn out and causing ergonomic issues. New furniture will improve employee comfort and productivity.',
      budgetCode: 'CAP-2024-001',
    },
    ...overrides,
  };
}

// Mock Approver Factory
export function createMockApprover(
  documentId: string,
  stepOrder: number,
  user: User,
  overrides?: Partial<Approver>
): Approver {
  return {
    id: uuidv4(),
    documentId,
    stepOrder,
    userId: user.id,
    user,
    role: user.role as UserRole,
    assignedAt: new Date(),
    canReassign: true,
    status: 'PENDING',
    ...overrides,
  };
}

// Mock Approval Log Entry Factory
export function createMockApprovalLogEntry(
  documentId: string,
  approver: User,
  overrides?: Partial<ApprovalLogEntry>
): ApprovalLogEntry {
  return {
    id: uuidv4(),
    documentId,
    approverId: approver.id,
    approver,
    action: 'APPROVED',
    timestamp: new Date(),
    comments: 'Approved as per company policy.',
    ...overrides,
  };
}

// Mock Attachment Factory
export function createMockAttachment(
  documentId: string,
  uploadedBy: User,
  overrides?: Partial<Attachment>
): Attachment {
  return {
    id: uuidv4(),
    documentId,
    fileName: 'supporting-document.pdf',
    fileSize: 2048576,
    fileType: 'application/pdf',
    uploadedById: uploadedBy.id,
    uploadedBy,
    uploadedAt: new Date(),
    storagePath: '/documents/' + uuidv4() + '.pdf',
    visibleToRoles: ['DEPARTMENT_MANAGER', 'FINANCE_OFFICER', 'DIRECTOR', 'CFO'],
    ...overrides,
  };
}

// Get random user by role
export function getRandomUserByRole(role: UserRole): User {
  const users = MOCK_USERS[role];
  return users[Math.floor(Math.random() * users.length)];
}

// Get all users
export function getAllMockUsers(): User[] {
  return Object.values(MOCK_USERS).flat();
}

// Mock data store for server actions
export const store = {
  documents: new Map(),
  approvalStates: new Map(),
  users: new Map(),
  auditLogs: new Map(),
};
