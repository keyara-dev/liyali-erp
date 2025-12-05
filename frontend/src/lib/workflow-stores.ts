// In-memory stores for workflow documents
// These are data stores, not server actions
import { WorkflowDocument, Approver, ApprovalLogEntry, Attachment } from '@/types/workflow';

export const documentStore = new Map<string, WorkflowDocument>();
export const approversStore = new Map<string, Approver[]>();
export const approvalLogsStore = new Map<string, ApprovalLogEntry[]>();
export const attachmentsStore = new Map<string, Attachment[]>();

export let isInitialized = false;
