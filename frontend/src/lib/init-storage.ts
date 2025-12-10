/**
 * Storage Initialization
 * Populates localStorage with seed data on app startup
 * This serves as the single source of truth until backend APIs are integrated
 *
 * When backend APIs are ready, simply:
 * 1. Remove these initialization calls
 * 2. Update server actions to fetch from APIs instead of localStorage
 * 3. Remove the localStorage hooks
 */

import { PurchaseOrder, PaymentVoucher, RequisitionForm } from '@/types/workflow';
import { MOCK_USERS } from './mock-data';
import { v4 as uuidv4 } from 'uuid';

const PO_STORAGE_KEY = 'liyali_purchase_orders';
const REQUISITION_STORAGE_KEY = 'liyali_requisitions';
const PAYMENT_VOUCHER_STORAGE_KEY = 'liyali_payment_vouchers';

/**
 * Create seed purchase orders with various statuses
 */
function createSeedPurchaseOrders(): PurchaseOrder[] {
  const requester = MOCK_USERS.REQUESTER[0];
  const now = new Date();

  return [
    // Draft PO
    {
      id: `po-${uuidv4()}`,
      type: 'PURCHASE_ORDER',
      documentNumber: 'PO-2024-001',
      status: 'DRAFT',
      currentStage: 0,
      createdBy: requester.id,
      createdByUser: requester,
      createdAt: new Date(now.getTime() - 5 * 24 * 60 * 60 * 1000), // 5 days ago
      updatedAt: new Date(now.getTime() - 5 * 24 * 60 * 60 * 1000),
      metadata: {
        vendorName: 'Mitete Supplies Ltd',
        vendorId: 'VENDOR-001',
        items: [
          {
            id: uuidv4(),
            description: 'Office Chairs - Ergonomic',
            quantity: 15,
            unitCost: 450,
            totalCost: 6750,
          },
          {
            id: uuidv4(),
            description: 'Standing Desks',
            quantity: 10,
            unitCost: 1200,
            totalCost: 12000,
          },
        ],
        totalAmount: 18750,
        currency: 'ZMW',
        deliveryDate: new Date(now.getTime() + 30 * 24 * 60 * 60 * 1000),
        specialInstructions: 'Deliver to main office. Notify 2 days before delivery.',
      },
    },

    // Submitted PO
    {
      id: `po-${uuidv4()}`,
      type: 'PURCHASE_ORDER',
      documentNumber: 'PO-2024-002',
      status: 'SUBMITTED',
      currentStage: 1,
      createdBy: requester.id,
      createdByUser: requester,
      createdAt: new Date(now.getTime() - 3 * 24 * 60 * 60 * 1000), // 3 days ago
      updatedAt: new Date(now.getTime() - 3 * 24 * 60 * 60 * 1000),
      metadata: {
        vendorName: 'TechXpress Solutions',
        vendorId: 'VENDOR-002',
        items: [
          {
            id: uuidv4(),
            description: 'Laptops - Dell XPS 13',
            quantity: 5,
            unitCost: 8500,
            totalCost: 42500,
          },
          {
            id: uuidv4(),
            description: 'USB-C Chargers',
            quantity: 5,
            unitCost: 150,
            totalCost: 750,
          },
        ],
        totalAmount: 43250,
        currency: 'ZMW',
        deliveryDate: new Date(now.getTime() + 14 * 24 * 60 * 60 * 1000),
        specialInstructions: 'Ensure all units come with warranty cards',
      },
    },

    // In Review PO
    {
      id: `po-${uuidv4()}`,
      type: 'PURCHASE_ORDER',
      documentNumber: 'PO-2024-003',
      status: 'IN_REVIEW',
      currentStage: 2,
      createdBy: requester.id,
      createdByUser: requester,
      createdAt: new Date(now.getTime() - 2 * 24 * 60 * 60 * 1000), // 2 days ago
      updatedAt: new Date(now.getTime() - 1 * 24 * 60 * 60 * 1000),
      metadata: {
        vendorName: 'Office Pro Equipment',
        vendorId: 'VENDOR-003',
        items: [
          {
            id: uuidv4(),
            description: 'Printer Paper - A4 (Box of 10 Reams)',
            quantity: 20,
            unitCost: 85,
            totalCost: 1700,
          },
          {
            id: uuidv4(),
            description: 'Ink Cartridges - Black',
            quantity: 50,
            unitCost: 45,
            totalCost: 2250,
          },
        ],
        totalAmount: 3950,
        currency: 'ZMW',
        deliveryDate: new Date(now.getTime() + 7 * 24 * 60 * 60 * 1000),
        specialInstructions: 'Standard delivery preferred',
      },
    },

    // Approved PO
    {
      id: `po-${uuidv4()}`,
      type: 'PURCHASE_ORDER',
      documentNumber: 'PO-2024-004',
      status: 'APPROVED',
      currentStage: 4,
      createdBy: requester.id,
      createdByUser: requester,
      createdAt: new Date(now.getTime() - 10 * 24 * 60 * 60 * 1000), // 10 days ago
      updatedAt: new Date(now.getTime() - 1 * 24 * 60 * 60 * 1000),
      metadata: {
        vendorName: 'Maintenance Solutions Ltd',
        vendorId: 'VENDOR-004',
        items: [
          {
            id: uuidv4(),
            description: 'Office Maintenance - Monthly',
            quantity: 1,
            unitCost: 5000,
            totalCost: 5000,
          },
        ],
        totalAmount: 5000,
        currency: 'ZMW',
        deliveryDate: new Date(now.getTime() + 5 * 24 * 60 * 60 * 1000),
        specialInstructions: 'Recurring monthly service',
      },
    },

    // Rejected PO
    {
      id: `po-${uuidv4()}`,
      type: 'PURCHASE_ORDER',
      documentNumber: 'PO-2024-005',
      status: 'REJECTED',
      currentStage: 0,
      createdBy: requester.id,
      createdByUser: requester,
      createdAt: new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000), // 7 days ago
      updatedAt: new Date(now.getTime() - 2 * 24 * 60 * 60 * 1000),
      metadata: {
        vendorName: 'Budget Supplies Co',
        vendorId: 'VENDOR-005',
        items: [
          {
            id: uuidv4(),
            description: 'Conference Table',
            quantity: 2,
            unitCost: 4500,
            totalCost: 9000,
          },
        ],
        totalAmount: 9000,
        currency: 'ZMW',
        deliveryDate: new Date(now.getTime() + 20 * 24 * 60 * 60 * 1000),
        specialInstructions: 'Rejected due to budget constraints',
      },
    },
  ];
}

/**
 * Create seed requisitions with various statuses
 */
function createSeedRequisitions(): RequisitionForm[] {
  const requester = MOCK_USERS.REQUESTER[0];
  const now = new Date();

  return [
    // Draft Requisition
    {
      id: `req-${uuidv4()}`,
      type: 'REQUISITION',
      documentNumber: 'REQ-2024-001',
      status: 'DRAFT',
      currentStage: 0,
      createdBy: requester.id,
      createdByUser: requester,
      createdAt: new Date(now.getTime() - 3 * 24 * 60 * 60 * 1000),
      updatedAt: new Date(now.getTime() - 3 * 24 * 60 * 60 * 1000),
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
        ],
        justification: 'Current office furniture is worn out. Need ergonomic chairs.',
        budgetCode: 'CAP-2024-001',
      },
    },

    // Submitted Requisition
    {
      id: `req-${uuidv4()}`,
      type: 'REQUISITION',
      documentNumber: 'REQ-2024-002',
      status: 'SUBMITTED',
      currentStage: 1,
      createdBy: requester.id,
      createdByUser: requester,
      createdAt: new Date(now.getTime() - 2 * 24 * 60 * 60 * 1000),
      updatedAt: new Date(now.getTime() - 2 * 24 * 60 * 60 * 1000),
      metadata: {
        department: 'IT',
        requestedFor: 'Sarah Banda',
        items: [
          {
            id: uuidv4(),
            itemDescription: 'Laptops - Dell XPS 13',
            quantity: 5,
            estimatedCost: 42500,
          },
        ],
        justification: 'New team members require development machines.',
        budgetCode: 'CAP-2024-002',
      },
    },

    // Approved Requisition
    {
      id: `req-${uuidv4()}`,
      type: 'REQUISITION',
      documentNumber: 'REQ-2024-003',
      status: 'APPROVED',
      currentStage: 4,
      createdBy: requester.id,
      createdByUser: requester,
      createdAt: new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000),
      updatedAt: new Date(now.getTime() - 1 * 24 * 60 * 60 * 1000),
      metadata: {
        department: 'HR',
        requestedFor: 'Maria Chiyanda',
        items: [
          {
            id: uuidv4(),
            itemDescription: 'Training Materials',
            quantity: 50,
            estimatedCost: 5000,
          },
        ],
        justification: 'Employee development and training programs.',
        budgetCode: 'EXP-2024-001',
      },
    },

    // Rejected Requisition
    {
      id: `req-${uuidv4()}`,
      type: 'REQUISITION',
      documentNumber: 'REQ-2024-004',
      status: 'REJECTED',
      currentStage: 0,
      createdBy: requester.id,
      createdByUser: requester,
      createdAt: new Date(now.getTime() - 5 * 24 * 60 * 60 * 1000),
      updatedAt: new Date(now.getTime() - 3 * 24 * 60 * 60 * 1000),
      metadata: {
        department: 'Marketing',
        requestedFor: 'Paul Nkosi',
        items: [
          {
            id: uuidv4(),
            itemDescription: 'Marketing Software Licenses',
            quantity: 10,
            estimatedCost: 15000,
          },
        ],
        justification: 'Rejected due to budget constraints.',
        budgetCode: 'EXP-2024-002',
      },
    },
  ];
}

/**
 * Create seed payment vouchers with various statuses
 */
function createSeedPaymentVouchers(): PaymentVoucher[] {
  const requester = MOCK_USERS.REQUESTER[0];
  const now = new Date();

  return [
    // Draft Payment Voucher
    {
      id: `pv-${uuidv4()}`,
      type: 'PAYMENT_VOUCHER',
      documentNumber: 'PV-2024-001',
      status: 'DRAFT',
      currentStage: 0,
      createdBy: requester.id,
      createdByUser: requester,
      createdAt: new Date(now.getTime() - 2 * 24 * 60 * 60 * 1000),
      updatedAt: new Date(now.getTime() - 2 * 24 * 60 * 60 * 1000),
      metadata: {
        payeeName: 'Mitete Supplies Ltd',
        payeeId: 'VENDOR-001',
        amount: 18750,
        currency: 'ZMW',
        reason: 'Payment for office furniture - PO-2024-001',
        accountCode: '4001-001',
        department: 'Operations',
      },
    },

    // Submitted Payment Voucher
    {
      id: `pv-${uuidv4()}`,
      type: 'PAYMENT_VOUCHER',
      documentNumber: 'PV-2024-002',
      status: 'SUBMITTED',
      currentStage: 1,
      createdBy: requester.id,
      createdByUser: requester,
      createdAt: new Date(now.getTime() - 4 * 24 * 60 * 60 * 1000),
      updatedAt: new Date(now.getTime() - 4 * 24 * 60 * 60 * 1000),
      metadata: {
        payeeName: 'TechXpress Solutions',
        payeeId: 'VENDOR-002',
        amount: 43250,
        currency: 'ZMW',
        reason: 'Payment for IT equipment and software - PO-2024-002',
        accountCode: '4001-002',
        department: 'IT',
      },
    },

    // In Review Payment Voucher
    {
      id: `pv-${uuidv4()}`,
      type: 'PAYMENT_VOUCHER',
      documentNumber: 'PV-2024-003',
      status: 'IN_REVIEW',
      currentStage: 2,
      createdBy: requester.id,
      createdByUser: requester,
      createdAt: new Date(now.getTime() - 3 * 24 * 60 * 60 * 1000),
      updatedAt: new Date(now.getTime() - 2 * 24 * 60 * 60 * 1000),
      metadata: {
        payeeName: 'Office Pro Equipment',
        payeeId: 'VENDOR-003',
        amount: 3950,
        currency: 'ZMW',
        reason: 'Office supplies and materials',
        accountCode: '4001-003',
        department: 'Operations',
      },
    },

    // Approved Payment Voucher
    {
      id: `pv-${uuidv4()}`,
      type: 'PAYMENT_VOUCHER',
      documentNumber: 'PV-2024-004',
      status: 'APPROVED',
      currentStage: 4,
      createdBy: requester.id,
      createdByUser: requester,
      createdAt: new Date(now.getTime() - 10 * 24 * 60 * 60 * 1000),
      updatedAt: new Date(now.getTime() - 2 * 24 * 60 * 60 * 1000),
      metadata: {
        payeeName: 'Maintenance Solutions Ltd',
        payeeId: 'VENDOR-004',
        amount: 5000,
        currency: 'ZMW',
        reason: 'Monthly office maintenance services',
        accountCode: '4001-004',
        department: 'Operations',
      },
    },
  ];
}

/**
 * Initialize localStorage with seed data if empty
 * This is called once on app startup
 */
export function initializeStorageWithSeedData(): void {
  if (typeof window === 'undefined') return;

  try {
    // Initialize Purchase Orders
    const existingPOs = localStorage.getItem(PO_STORAGE_KEY);
    if (!existingPOs) {
      const seedData = createSeedPurchaseOrders();
      localStorage.setItem(PO_STORAGE_KEY, JSON.stringify(seedData));
      console.log(`✓ Initialized localStorage with ${seedData.length} seed purchase orders`);
    }

    // Initialize Requisitions
    const existingReqs = localStorage.getItem(REQUISITION_STORAGE_KEY);
    if (!existingReqs) {
      const seedData = createSeedRequisitions();
      localStorage.setItem(REQUISITION_STORAGE_KEY, JSON.stringify(seedData));
      console.log(`✓ Initialized localStorage with ${seedData.length} seed requisitions`);
    }

    // Initialize Payment Vouchers
    const existingPVs = localStorage.getItem(PAYMENT_VOUCHER_STORAGE_KEY);
    if (!existingPVs) {
      const seedData = createSeedPaymentVouchers();
      localStorage.setItem(PAYMENT_VOUCHER_STORAGE_KEY, JSON.stringify(seedData));
      console.log(`✓ Initialized localStorage with ${seedData.length} seed payment vouchers`);
    }
  } catch (error) {
    console.error('Failed to initialize storage with seed data:', error);
  }
}

/**
 * Clear all data from localStorage (for testing/reset)
 */
export function clearAllStorageData(): void {
  if (typeof window === 'undefined') return;

  try {
    localStorage.removeItem(PO_STORAGE_KEY);
    localStorage.removeItem(REQUISITION_STORAGE_KEY);
    localStorage.removeItem(PAYMENT_VOUCHER_STORAGE_KEY);
    console.log('✓ Cleared all documents from localStorage');
  } catch (error) {
    console.error('Failed to clear storage:', error);
  }
}

/**
 * Reset localStorage with fresh seed data (for testing)
 */
export function resetStorageWithSeedData(): void {
  clearAllStorageData();
  initializeStorageWithSeedData();
}

/**
 * Export storage keys for use in hooks
 */
export const STORAGE_KEYS = {
  PURCHASE_ORDERS: PO_STORAGE_KEY,
  REQUISITIONS: REQUISITION_STORAGE_KEY,
  PAYMENT_VOUCHERS: PAYMENT_VOUCHER_STORAGE_KEY,
};
