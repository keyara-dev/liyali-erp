// In-memory stores for workflow documents
// These are data stores, not server actions
import { WorkflowDocument, Approver, ApprovalLogEntry, Attachment } from '@/types/workflow';

export const documentStore = new Map<string, WorkflowDocument>();
export const approversStore = new Map<string, Approver[]>();
export const approvalLogsStore = new Map<string, ApprovalLogEntry[]>();
export const attachmentsStore = new Map<string, Attachment[]>();

export let isInitialized = false;

/**
 * Initialize workflow stores with seed data (for development)
 * This centralizes all mock/seed data in one location
 */
export function initializeWorkflowStores() {
  if (isInitialized) return;

  // Initialize with empty stores - data will be added via createWorkflowDocument
  // when requisitions, purchase orders, and payment vouchers are created
  isInitialized = true;
}
