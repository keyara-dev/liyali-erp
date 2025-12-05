/**
 * Workflow initialization logic
 * This file is separated from server actions to avoid "use server" constraints
 */

import {
  WorkflowDocument,
  WorkflowDocumentType,
  DocumentStatus,
} from '@/types/workflow';
import {
  createMockPurchaseOrder,
  createMockPaymentVoucher,
  createMockRequisitionForm,
  MOCK_USERS,
} from '@/lib/mock-data';
import {
  documentStore,
  isInitialized as checkInitialized,
} from '@/lib/workflow-stores';

let isInitialized = checkInitialized;

export function initializeSampleData() {
  if (isInitialized) return;

  const statuses: DocumentStatus[] = ['DRAFT', 'SUBMITTED', 'IN_REVIEW', 'APPROVED', 'REJECTED', 'REVERSED'];
  const documentTypes: WorkflowDocumentType[] = ['REQUISITION', 'PURCHASE_ORDER', 'PAYMENT_VOUCHER'];

  // Create 25 sample documents with varied data
  for (let i = 0; i < 25; i++) {
    const status = statuses[i % statuses.length];
    const type = documentTypes[i % documentTypes.length];
    const daysAgo = Math.floor(i / 2);
    const createdDate = new Date(Date.now() - daysAgo * 24 * 60 * 60 * 1000);

    let doc: WorkflowDocument;

    switch (type) {
      case 'PURCHASE_ORDER':
        doc = createMockPurchaseOrder({
          status,
          currentStage: status === 'DRAFT' ? 0 : Math.min(i % 4, 3),
          createdAt: createdDate,
          updatedAt: createdDate,
          createdBy: MOCK_USERS.REQUESTER[i % MOCK_USERS.REQUESTER.length].id,
        });
        break;
      case 'PAYMENT_VOUCHER':
        doc = createMockPaymentVoucher({
          status,
          currentStage: status === 'DRAFT' ? 0 : Math.min(i % 4, 3),
          createdAt: createdDate,
          updatedAt: createdDate,
          createdBy: MOCK_USERS.REQUESTER[i % MOCK_USERS.REQUESTER.length].id,
        });
        break;
      default:
        doc = createMockRequisitionForm({
          status,
          currentStage: status === 'DRAFT' ? 0 : Math.min(i % 4, 3),
          createdAt: createdDate,
          updatedAt: createdDate,
          createdBy: MOCK_USERS.REQUESTER[i % MOCK_USERS.REQUESTER.length].id,
        });
    }

    documentStore.set(doc.id, doc);
  }

  isInitialized = true;
}

// Initialize on module load
initializeSampleData();
