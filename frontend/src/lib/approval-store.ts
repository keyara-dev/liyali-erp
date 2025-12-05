/**
 * Mock Approval Store with localStorage Persistence
 * Simulates database layer for approval operations using localStorage
 * In production, this would be replaced with actual database calls
 *
 * Features:
 * - In-memory cache with localStorage backup
 * - Persistent approval history
 * - Task state management
 * - Mock workflow progression
 * - Data serialization for storage
 */

import { ApprovalTask, ApprovalTaskDetail } from '@/types';

// ============================================================================
// STORAGE KEYS
// ============================================================================

const STORAGE_KEYS = {
  TASKS: 'approval_tasks_v1',
  HISTORY: 'approval_history_v1',
  METADATA: 'approval_metadata_v1',
};

/**
 * Get localStorage with fallback
 * Returns null in server-side context
 */
function getLocalStorage(): Storage | null {
  if (typeof window === 'undefined') return null;
  try {
    return window.localStorage;
  } catch {
    return null;
  }
}

interface ApprovalRecord {
  taskId: string;
  action: 'APPROVED' | 'REJECTED' | 'REASSIGNED';
  actionBy: string;
  actionAt: Date;
  signature?: string;
  remarks?: string;
  newAssignee?: string;
  reassignmentReason?: string;
}

interface ApprovalTaskStore extends ApprovalTask {
  approvalHistory: ApprovalRecord[];
  workflowData?: {
    id: string;
    name: string;
    totalStages: number;
    stages: Array<{
      stageNumber: number;
      name: string;
      description?: string;
    }>;
  };
  entityData?: Record<string, any>;
}

/**
 * In-memory store with localStorage persistence
 * Simulates database for approval tasks
 * Uses localStorage for persistence across page reloads
 */
class ApprovalStore {
  private tasks: Map<string, ApprovalTaskStore> = new Map();
  private approvalHistory: ApprovalRecord[] = [];

  constructor() {
    // Try to load from localStorage first
    if (!this.loadFromStorage()) {
      // If no stored data, initialize with mock data
      this.initializeMockData();
      // Save initial data to storage
      this.saveToStorage();
    }
  }

  /**
   * Load data from localStorage
   * @returns true if data was loaded, false if not available
   */
  private loadFromStorage(): boolean {
    const storage = getLocalStorage();
    if (!storage) return false;

    try {
      const tasksData = storage.getItem(STORAGE_KEYS.TASKS);
      const historyData = storage.getItem(STORAGE_KEYS.HISTORY);

      if (tasksData) {
        const parsedTasks = JSON.parse(tasksData);
        Object.entries(parsedTasks).forEach(([key, taskData]: [string, any]) => {
          // Reconstruct Date objects
          const task = {
            ...taskData,
            createdAt: new Date(taskData.createdAt),
            dueDate: taskData.dueDate ? new Date(taskData.dueDate) : undefined,
            actionDate: taskData.actionDate ? new Date(taskData.actionDate) : undefined,
            approvalHistory: (taskData.approvalHistory || []).map((record: any) => ({
              ...record,
              actionAt: new Date(record.actionAt),
            })),
          };
          this.tasks.set(key, task);
        });

        if (historyData) {
          this.approvalHistory = JSON.parse(historyData).map((record: any) => ({
            ...record,
            actionAt: new Date(record.actionAt),
          }));
        }

        console.log('✅ Loaded approval data from localStorage');
        return true;
      }
    } catch (error) {
      console.warn('⚠️  Failed to load from localStorage:', error);
      // Fall through to initialize with mock data
    }

    return false;
  }

  /**
   * Save data to localStorage
   */
  private saveToStorage(): void {
    const storage = getLocalStorage();
    if (!storage) return;

    try {
      // Convert Map to object for JSON serialization
      const tasksObject: Record<string, any> = {};
      this.tasks.forEach((task, key) => {
        tasksObject[key] = {
          ...task,
          createdAt: task.createdAt.toISOString(),
          dueDate: task.dueDate?.toISOString(),
          actionDate: task.actionDate?.toISOString(),
          approvalHistory: task.approvalHistory.map((record) => ({
            ...record,
            actionAt: record.actionAt.toISOString(),
          })),
        };
      });

      storage.setItem(STORAGE_KEYS.TASKS, JSON.stringify(tasksObject));
      storage.setItem(
        STORAGE_KEYS.HISTORY,
        JSON.stringify(
          this.approvalHistory.map((record) => ({
            ...record,
            actionAt: record.actionAt.toISOString(),
          }))
        )
      );
    } catch (error) {
      console.warn('⚠️  Failed to save to localStorage:', error);
    }
  }

  /**
   * Initialize with mock approval tasks
   */
  private initializeMockData() {
    const now = new Date();
    const mockTasks: ApprovalTaskStore[] = [
      {
        id: 'task-req-001',
        entityId: 'req-001',
        entityType: 'REQUISITION',
        entityNumber: 'REQ-2024-001',
        status: 'pending',
        stageName: 'Manager Approval',
        stageIndex: 0,
        importance: 'HIGH',
        approverName: 'John Doe',
        approverUserId: 'user-john-001',
        createdAt: new Date(now.getTime() - 2 * 24 * 60 * 60 * 1000),
        dueDate: new Date(now.getTime() + 5 * 24 * 60 * 60 * 1000),
        workflowId: 'wf-req-standard',
        workflowName: 'Standard Requisition Approval',
        approvalHistory: [],
        workflowData: {
          id: 'wf-req-standard',
          name: 'Standard Requisition Approval',
          totalStages: 3,
          stages: [
            { stageNumber: 0, name: 'Manager Approval', description: 'Initial manager review' },
            { stageNumber: 1, name: 'Director Approval', description: 'Director verification' },
            { stageNumber: 2, name: 'Final Approval', description: 'Executive final sign-off' },
          ],
        },
        entityData: {
          id: 'req-001',
          number: 'REQ-2024-001',
          amount: 25000,
          department: 'Operations',
          description: 'Office supplies and equipment',
          items: [
            { id: 'item-1', name: 'Laptops', quantity: 5, unitPrice: 1200, total: 6000 },
            { id: 'item-2', name: 'Monitors', quantity: 5, unitPrice: 400, total: 2000 },
            { id: 'item-3', name: 'Keyboards', quantity: 10, unitPrice: 150, total: 1500 },
            { id: 'item-4', name: 'Mice', quantity: 10, unitPrice: 50, total: 500 },
            { id: 'item-5', name: 'Desks', quantity: 5, unitPrice: 800, total: 4000 },
            { id: 'item-6', name: 'Chairs', quantity: 5, unitPrice: 300, total: 1500 },
            { id: 'item-7', name: 'Shelving', quantity: 8, unitPrice: 200, total: 1600 },
            { id: 'item-8', name: 'Filing Cabinets', quantity: 4, unitPrice: 400, total: 1600 },
            { id: 'item-9', name: 'Printer', quantity: 2, unitPrice: 800, total: 1600 },
            { id: 'item-10', name: 'Cables & Adapters', quantity: 20, unitPrice: 100, total: 2000 },
          ],
          createdBy: 'user-alice-001',
          createdAt: new Date(now.getTime() - 3 * 24 * 60 * 60 * 1000),
        },
      },
      {
        id: 'task-budget-001',
        entityId: 'budget-001',
        entityType: 'BUDGET',
        entityNumber: 'BUD-2024-Q1-001',
        status: 'pending',
        stageName: 'Director Review',
        stageIndex: 1,
        importance: 'MEDIUM',
        approverName: 'Jane Smith',
        approverUserId: 'user-jane-001',
        createdAt: new Date(now.getTime() - 1 * 24 * 60 * 60 * 1000),
        dueDate: new Date(now.getTime() + 3 * 24 * 60 * 60 * 1000),
        workflowId: 'wf-budget-standard',
        workflowName: 'Budget Approval Workflow',
        approvalHistory: [],
        workflowData: {
          id: 'wf-budget-standard',
          name: 'Budget Approval Workflow',
          totalStages: 2,
          stages: [
            { stageNumber: 0, name: 'Manager Review', description: 'Manager budget review' },
            { stageNumber: 1, name: 'Director Review', description: 'Director budget approval' },
          ],
        },
        entityData: {
          id: 'budget-001',
          number: 'BUD-2024-Q1-001',
          totalAmount: 500000,
          department: 'Operations',
          fiscalYear: '2024',
          description: 'Q1 2024 operational budget',
          allocations: [
            { id: 'alloc-1', category: 'Personnel', amount: 250000, percentage: 50 },
            { id: 'alloc-2', category: 'Equipment', amount: 100000, percentage: 20 },
            { id: 'alloc-3', category: 'Operations', amount: 100000, percentage: 20 },
            { id: 'alloc-4', category: 'Contingency', amount: 50000, percentage: 10 },
          ],
          createdBy: 'user-bob-001',
          createdAt: new Date(now.getTime() - 5 * 24 * 60 * 60 * 1000),
        },
      },
      {
        id: 'task-req-002',
        entityId: 'req-002',
        entityType: 'REQUISITION',
        entityNumber: 'REQ-2024-002',
        status: 'pending',
        stageName: 'Manager Approval',
        stageIndex: 0,
        importance: 'LOW',
        approverName: 'John Doe',
        approverUserId: 'user-john-001',
        createdAt: new Date(now.getTime() - 0.5 * 24 * 60 * 60 * 1000),
        dueDate: new Date(now.getTime() + 7 * 24 * 60 * 60 * 1000),
        workflowId: 'wf-req-standard',
        workflowName: 'Standard Requisition Approval',
        approvalHistory: [],
        workflowData: {
          id: 'wf-req-standard',
          name: 'Standard Requisition Approval',
          totalStages: 3,
          stages: [
            { stageNumber: 0, name: 'Manager Approval', description: 'Initial manager review' },
            { stageNumber: 1, name: 'Director Approval', description: 'Director verification' },
            { stageNumber: 2, name: 'Final Approval', description: 'Executive final sign-off' },
          ],
        },
        entityData: {
          id: 'req-002',
          number: 'REQ-2024-002',
          amount: 5000,
          department: 'IT',
          description: 'Software licenses renewal',
          items: [
            { id: 'item-1', name: 'Microsoft 365 Enterprise', quantity: 50, unitPrice: 100, total: 5000 },
          ],
          createdBy: 'user-charlie-001',
          createdAt: new Date(now.getTime() - 1 * 24 * 60 * 60 * 1000),
        },
      },
    ];

    mockTasks.forEach((task) => {
      this.tasks.set(task.id, task);
    });
  }

  /**
   * Get all pending approval tasks
   */
  getAllTasks(status?: string): ApprovalTask[] {
    const tasks = Array.from(this.tasks.values()).map((task) => {
      const { approvalHistory, workflowData, entityData, ...rest } = task;
      return rest as ApprovalTask;
    });

    if (status) {
      return tasks.filter((t) => t.status === status);
    }
    return tasks;
  }

  /**
   * Get detailed approval task with workflow and entity data
   */
  getTaskDetail(taskId: string): ApprovalTaskDetail | null {
    const task = this.tasks.get(taskId);
    if (!task) return null;

    const { approvalHistory, workflowData, entityData, ...taskBase } = task;
    return {
      task: taskBase as ApprovalTask,
      workflow: workflowData,
      entity: entityData,
      relatedApprovals: this.getRelatedApprovals(task.entityId),
    };
  }

  /**
   * Get related approvals for an entity (same workflow)
   */
  private getRelatedApprovals(entityId: string): ApprovalTask[] {
    return Array.from(this.tasks.values())
      .filter((task) => task.entityId === entityId)
      .map((task) => {
        const { approvalHistory, workflowData, entityData, ...rest } = task;
        return rest as ApprovalTask;
      });
  }

  /**
   * Approve a task
   */
  approveTask(
    taskId: string,
    approverId: string,
    signature: string,
    remarks?: string
  ): { success: boolean; message: string; nextStage?: string } {
    const task = this.tasks.get(taskId);
    if (!task) {
      return { success: false, message: 'Task not found' };
    }

    // Record approval
    const record: ApprovalRecord = {
      taskId,
      action: 'APPROVED',
      actionBy: approverId,
      actionAt: new Date(),
      signature,
      remarks,
    };
    this.approvalHistory.push(record);
    task.approvalHistory.push(record);

    // Simulate stage progression
    const currentStageIndex = task.stageIndex;
    const totalStages = task.workflowData?.totalStages || 1;

    // Save to storage after any modification
    this.saveToStorage();

    if (currentStageIndex < totalStages - 1) {
      // Move to next stage
      task.stageIndex = currentStageIndex + 1;
      const nextStage = task.workflowData?.stages[task.stageIndex];
      if (nextStage) {
        task.stageName = nextStage.name;
      }
      this.saveToStorage();
      return {
        success: true,
        message: 'Task approved. Moving to next stage.',
        nextStage: task.stageName,
      };
    } else {
      // Final approval
      task.status = 'approved';
      task.actionDate = new Date();
      this.saveToStorage();
      return {
        success: true,
        message: 'Task fully approved and workflow completed.',
        nextStage: 'COMPLETED',
      };
    }
  }

  /**
   * Reject a task
   */
  rejectTask(
    taskId: string,
    rejectorId: string,
    signature: string,
    remarks: string
  ): { success: boolean; message: string } {
    const task = this.tasks.get(taskId);
    if (!task) {
      return { success: false, message: 'Task not found' };
    }

    // Record rejection
    const record: ApprovalRecord = {
      taskId,
      action: 'REJECTED',
      actionBy: rejectorId,
      actionAt: new Date(),
      signature,
      remarks,
    };
    this.approvalHistory.push(record);
    task.approvalHistory.push(record);

    // Update task status
    task.status = 'rejected';
    task.actionDate = new Date();

    // Save to storage
    this.saveToStorage();

    return {
      success: true,
      message: 'Task rejected and returned to originator.',
    };
  }

  /**
   * Reassign a task to a different approver
   */
  reassignTask(
    taskId: string,
    reassignedBy: string,
    newApproverId: string,
    newApproverName: string,
    reason: string
  ): { success: boolean; message: string } {
    const task = this.tasks.get(taskId);
    if (!task) {
      return { success: false, message: 'Task not found' };
    }

    // Record reassignment
    const record: ApprovalRecord = {
      taskId,
      action: 'REASSIGNED',
      actionBy: reassignedBy,
      actionAt: new Date(),
      newAssignee: newApproverId,
      reassignmentReason: reason,
    };
    this.approvalHistory.push(record);
    task.approvalHistory.push(record);

    // Update task assignment
    task.approverUserId = newApproverId;
    task.approverName = newApproverName;

    // Save to storage
    this.saveToStorage();

    return {
      success: true,
      message: `Task reassigned to ${newApproverName}.`,
    };
  }

  /**
   * Get approval history for an entity
   */
  getApprovalHistory(entityId: string): ApprovalRecord[] {
    const task = Array.from(this.tasks.values()).find((t) => t.entityId === entityId);
    if (!task) return [];
    return task.approvalHistory;
  }

  /**
   * Get approval statistics
   */
  getStatistics(): {
    totalPending: number;
    highPriority: number;
    thisMonth: number;
    overdue: number;
  } {
    const tasks = Array.from(this.tasks.values());
    const now = new Date();
    const currentMonth = now.getMonth();
    const currentYear = now.getFullYear();

    return {
      totalPending: tasks.filter((t) => t.status === 'pending').length,
      highPriority: tasks.filter((t) => t.status === 'pending' && t.importance === 'HIGH').length,
      thisMonth: tasks.filter((t) => {
        const createdDate = new Date(t.createdAt);
        return (
          t.status === 'approved' &&
          createdDate.getMonth() === currentMonth &&
          createdDate.getFullYear() === currentYear
        );
      }).length,
      overdue: tasks.filter((t) => t.dueDate && new Date(t.dueDate) < now && t.status === 'pending')
        .length,
    };
  }
}

// Export singleton instance
export const approvalStore = new ApprovalStore();
