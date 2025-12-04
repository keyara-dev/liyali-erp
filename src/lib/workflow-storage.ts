// localStorage utility for workflow persistence

export interface StoredWorkflow {
  id: string;
  name: string;
  description: string;
  documentType: string;
  stages: number;
  status: 'ACTIVE' | 'DEPRECATED';
  createdAt: string;
  updatedAt: string;
  createdBy: string;
  fullData: {
    name: string;
    description: string;
    documentType: string;
    isDefault: boolean;
    stages: Array<{
      id: string;
      order: number;
      name: string;
      description: string;
      approverRole: string;
      requiredApprovals: number;
      canReject: boolean;
      canReassign: boolean;
    }>;
  };
}

const WORKFLOWS_KEY = 'liyali_workflows';

// Initialize with mock data if localStorage is empty
function initializeWithMockData(): StoredWorkflow[] {
  const mockWorkflows: StoredWorkflow[] = [
    {
      id: 'wf-1',
      name: 'Standard Requisition Approval',
      description: '4-stage approval process for purchase requisitions',
      documentType: 'REQUISITION',
      stages: 4,
      status: 'ACTIVE',
      createdAt: '2024-01-15T10:30:00Z',
      updatedAt: '2024-11-20T14:22:00Z',
      createdBy: 'admin@example.com',
      fullData: {
        name: 'Standard Requisition Approval',
        description: '4-stage approval process for purchase requisitions',
        documentType: 'REQUISITION',
        isDefault: true,
        stages: [
          {
            id: 'stage-1',
            order: 1,
            name: 'Department Manager Review',
            description: 'Initial review by department manager',
            approverRole: 'DEPARTMENT_MANAGER',
            requiredApprovals: 1,
            canReject: true,
            canReassign: true,
          },
          {
            id: 'stage-2',
            order: 2,
            name: 'Finance Officer Review',
            description: 'Budget and finance validation',
            approverRole: 'FINANCE_OFFICER',
            requiredApprovals: 1,
            canReject: true,
            canReassign: true,
          },
          {
            id: 'stage-3',
            order: 3,
            name: 'CFO Approval',
            description: 'Final approval by CFO',
            approverRole: 'CFO',
            requiredApprovals: 1,
            canReject: true,
            canReassign: false,
          },
          {
            id: 'stage-4',
            order: 4,
            name: 'Admin Final Review',
            description: 'Admin verification',
            approverRole: 'ADMIN',
            requiredApprovals: 1,
            canReject: false,
            canReassign: true,
          },
        ],
      },
    },
    {
      id: 'wf-2',
      name: 'Purchase Order Approval',
      description: '4-stage approval with CFO override capability',
      documentType: 'PURCHASE_ORDER',
      stages: 4,
      status: 'ACTIVE',
      createdAt: '2024-02-10T08:15:00Z',
      updatedAt: '2024-11-18T11:45:00Z',
      createdBy: 'admin@example.com',
      fullData: {
        name: 'Purchase Order Approval',
        description: '4-stage approval with CFO override capability',
        documentType: 'PURCHASE_ORDER',
        isDefault: false,
        stages: [
          {
            id: 'stage-1',
            order: 1,
            name: 'Department Manager Approval',
            description: 'Department head approval',
            approverRole: 'DEPARTMENT_MANAGER',
            requiredApprovals: 1,
            canReject: true,
            canReassign: true,
          },
          {
            id: 'stage-2',
            order: 2,
            name: 'Procurement Review',
            description: 'Procurement officer verification',
            approverRole: 'PROCUREMENT_OFFICER',
            requiredApprovals: 1,
            canReject: true,
            canReassign: true,
          },
          {
            id: 'stage-3',
            order: 3,
            name: 'Finance Validation',
            description: 'Budget check and approval',
            approverRole: 'FINANCE_OFFICER',
            requiredApprovals: 2,
            canReject: true,
            canReassign: true,
          },
          {
            id: 'stage-4',
            order: 4,
            name: 'CFO Override',
            description: 'CFO final approval or override',
            approverRole: 'CFO',
            requiredApprovals: 1,
            canReject: true,
            canReassign: false,
          },
        ],
      },
    },
    {
      id: 'wf-3',
      name: 'Payment Voucher Review',
      description: 'Finance review workflow for payment processing',
      documentType: 'PAYMENT_VOUCHER',
      stages: 3,
      status: 'ACTIVE',
      createdAt: '2024-03-05T09:20:00Z',
      updatedAt: '2024-11-19T13:10:00Z',
      createdBy: 'finance-admin@example.com',
      fullData: {
        name: 'Payment Voucher Review',
        description: 'Finance review workflow for payment processing',
        documentType: 'PAYMENT_VOUCHER',
        isDefault: false,
        stages: [
          {
            id: 'stage-1',
            order: 1,
            name: 'Finance Officer Review',
            description: 'Initial payment verification',
            approverRole: 'FINANCE_OFFICER',
            requiredApprovals: 1,
            canReject: true,
            canReassign: true,
          },
          {
            id: 'stage-2',
            order: 2,
            name: 'CFO Approval',
            description: 'CFO payment authorization',
            approverRole: 'CFO',
            requiredApprovals: 1,
            canReject: true,
            canReassign: false,
          },
          {
            id: 'stage-3',
            order: 3,
            name: 'Admin Processing',
            description: 'Admin final processing',
            approverRole: 'ADMIN',
            requiredApprovals: 1,
            canReject: false,
            canReassign: true,
          },
        ],
      },
    },
    {
      id: 'wf-4',
      name: 'GRN Confirmation Flow',
      description: 'Simple goods receipt confirmation workflow',
      documentType: 'GOODS_RECEIVED_NOTE',
      stages: 2,
      status: 'ACTIVE',
      createdAt: '2024-04-12T11:00:00Z',
      updatedAt: '2024-11-17T10:30:00Z',
      createdBy: 'warehouse@example.com',
      fullData: {
        name: 'GRN Confirmation Flow',
        description: 'Simple goods receipt confirmation workflow',
        documentType: 'GOODS_RECEIVED_NOTE',
        isDefault: false,
        stages: [
          {
            id: 'stage-1',
            order: 1,
            name: 'Warehouse Confirmation',
            description: 'Warehouse manager goods receipt verification',
            approverRole: 'WAREHOUSE_MANAGER',
            requiredApprovals: 1,
            canReject: true,
            canReassign: true,
          },
          {
            id: 'stage-2',
            order: 2,
            name: 'Finance Reconciliation',
            description: 'Finance officer reconciliation check',
            approverRole: 'FINANCE_OFFICER',
            requiredApprovals: 1,
            canReject: false,
            canReassign: true,
          },
        ],
      },
    },
  ];

  return mockWorkflows;
}

export function getAllWorkflows(): StoredWorkflow[] {
  if (typeof window === 'undefined') {
    return [];
  }

  try {
    const data = localStorage.getItem(WORKFLOWS_KEY);
    if (!data) {
      // Initialize with mock data
      const mockData = initializeWithMockData();
      localStorage.setItem(WORKFLOWS_KEY, JSON.stringify(mockData));
      return mockData;
    }
    return JSON.parse(data);
  } catch (error) {
    console.error('Error reading workflows from localStorage:', error);
    return initializeWithMockData();
  }
}

export function getWorkflowById(id: string) {
  if (typeof window === 'undefined') {
    return null;
  }

  try {
    const workflows = getAllWorkflows();
    return workflows.find((w) => w.id === id) || null;
  } catch (error) {
    console.error('Error fetching workflow:', error);
    return null;
  }
}

export function saveWorkflow(
  workflow: Omit<StoredWorkflow, 'createdAt' | 'createdBy'> & {
    createdAt?: string;
    createdBy?: string;
  }
) {
  if (typeof window === 'undefined') {
    return null;
  }

  try {
    const workflows = getAllWorkflows();
    const existingIndex = workflows.findIndex((w) => w.id === workflow.id);

    const workflowToSave: StoredWorkflow = {
      ...workflow,
      createdAt: workflow.createdAt || new Date().toISOString(),
      createdBy: workflow.createdBy || 'current-user',
    };

    if (existingIndex >= 0) {
      // Update existing
      workflows[existingIndex] = workflowToSave;
    } else {
      // Add new
      workflows.push(workflowToSave);
    }

    localStorage.setItem(WORKFLOWS_KEY, JSON.stringify(workflows));
    return workflowToSave;
  } catch (error) {
    console.error('Error saving workflow:', error);
    return null;
  }
}

export function deleteWorkflow(id: string) {
  if (typeof window === 'undefined') {
    return false;
  }

  try {
    const workflows = getAllWorkflows();
    const filtered = workflows.filter((w) => w.id !== id);
    localStorage.setItem(WORKFLOWS_KEY, JSON.stringify(filtered));
    return true;
  } catch (error) {
    console.error('Error deleting workflow:', error);
    return false;
  }
}

export function duplicateWorkflow(id: string): StoredWorkflow | null {
  if (typeof window === 'undefined') {
    return null;
  }

  try {
    const workflow = getWorkflowById(id);
    if (!workflow) return null;

    const newWorkflow: StoredWorkflow = {
      ...workflow,
      id: `wf-${Date.now()}`,
      name: `${workflow.name} (Copy)`,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      fullData: {
        ...workflow.fullData,
        name: `${workflow.fullData.name} (Copy)`,
      },
    };

    return saveWorkflow(newWorkflow);
  } catch (error) {
    console.error('Error duplicating workflow:', error);
    return null;
  }
}
