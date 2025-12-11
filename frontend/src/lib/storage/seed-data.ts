/**
 * Seed Data for Development
 * Creates initial mock data for all document types
 *
 * Structure:
 * - Purchase Orders: Various statuses (DRAFT, SUBMITTED, IN_REVIEW, APPROVED, REJECTED)
 * - Requisitions: Various statuses and departments
 * - Payment Vouchers: Various stages in approval workflow
 * - Goods Received Notes: Various statuses (DRAFT, IN_REVIEW, APPROVED)
 */

import { PurchaseOrder, PaymentVoucher, RequisitionForm } from '@/types/workflow';
import { MOCK_USERS } from '../mock-data';
import { v4 as uuidv4 } from 'uuid';

export interface GoodsReceivedNote {
  id: string;
  type: 'GOODS_RECEIVED_NOTE';
  documentNumber: string;
  status: 'DRAFT' | 'IN_REVIEW' | 'APPROVED';
  currentStage: number;
  createdBy: string;
  createdByUser?: any;
  createdAt: Date;
  updatedAt: Date;
  metadata: {
    poId: string;
    poNumber: string;
    vendorName: string;
    receivedQuantity: number;
    totalQuantity: number;
    amount: number;
    receivedDate: string;
    warehouseLocation?: string;
  };
}

/**
 * Create seed purchase orders with various statuses
 * Includes multiple users, vendors, departments, and amounts for search testing
 */
export function createSeedPurchaseOrders(): PurchaseOrder[] {
  const requesters = MOCK_USERS.REQUESTER || [];
  const requester1 = requesters[0] || { id: 'user-1', name: 'John Mwale', email: 'john@example.com' };
  const requester2 = requesters[1] || { id: 'user-2', name: 'Sarah Banda', email: 'sarah@example.com' };
  const now = new Date();

  return [
    {
      id: `po-${uuidv4()}`,
      type: 'PURCHASE_ORDER',
      documentNumber: 'PO-2024-001',
      status: 'DRAFT',
      currentStage: 0,
      createdBy: requester1.id,
      createdByUser: requester1,
      createdAt: new Date(now.getTime() - 5 * 24 * 60 * 60 * 1000),
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
    {
      id: `po-${uuidv4()}`,
      type: 'PURCHASE_ORDER',
      documentNumber: 'PO-2024-002',
      status: 'SUBMITTED',
      currentStage: 1,
      createdBy: requester2.id,
      createdByUser: requester2,
      createdAt: new Date(now.getTime() - 3 * 24 * 60 * 60 * 1000),
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
    {
      id: `po-${uuidv4()}`,
      type: 'PURCHASE_ORDER',
      documentNumber: 'PO-2024-003',
      status: 'IN_REVIEW',
      currentStage: 2,
      createdBy: requester1.id,
      createdByUser: requester1,
      createdAt: new Date(now.getTime() - 2 * 24 * 60 * 60 * 1000),
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
    {
      id: `po-${uuidv4()}`,
      type: 'PURCHASE_ORDER',
      documentNumber: 'PO-2024-004',
      status: 'APPROVED',
      currentStage: 4,
      createdBy: requester2.id,
      createdByUser: requester2,
      createdAt: new Date(now.getTime() - 10 * 24 * 60 * 60 * 1000),
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
    {
      id: `po-${uuidv4()}`,
      type: 'PURCHASE_ORDER',
      documentNumber: 'PO-2024-005',
      status: 'REJECTED',
      currentStage: 0,
      createdBy: requester1.id,
      createdByUser: requester1,
      createdAt: new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000),
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
    // Additional POs for search testing
    {
      id: `po-${uuidv4()}`,
      type: 'PURCHASE_ORDER',
      documentNumber: 'PO-2024-006',
      status: 'SUBMITTED',
      currentStage: 1,
      createdBy: requester2.id,
      createdByUser: requester2,
      createdAt: new Date(now.getTime() - 1 * 24 * 60 * 60 * 1000),
      updatedAt: new Date(now.getTime() - 1 * 24 * 60 * 60 * 1000),
      metadata: {
        vendorName: 'Premium Consulting Group',
        vendorId: 'VENDOR-006',
        items: [
          {
            id: uuidv4(),
            description: 'Management Training Course',
            quantity: 1,
            unitCost: 25000,
            totalCost: 25000,
          },
        ],
        totalAmount: 25000,
        currency: 'ZMW',
        deliveryDate: new Date(now.getTime() + 45 * 24 * 60 * 60 * 1000),
        specialInstructions: 'Executive level management training',
      },
    },
    {
      id: `po-${uuidv4()}`,
      type: 'PURCHASE_ORDER',
      documentNumber: 'PO-2024-007',
      status: 'IN_REVIEW',
      currentStage: 2,
      createdBy: requester2.id,
      createdByUser: requester2,
      createdAt: new Date(now.getTime() - 4 * 24 * 60 * 60 * 1000),
      updatedAt: new Date(now.getTime() - 1 * 24 * 60 * 60 * 1000),
      metadata: {
        vendorName: 'Eco Solutions International',
        vendorId: 'VENDOR-007',
        items: [
          {
            id: uuidv4(),
            description: 'Solar Panel Installation',
            quantity: 50,
            unitCost: 2500,
            totalCost: 125000,
          },
        ],
        totalAmount: 125000,
        currency: 'ZMW',
        deliveryDate: new Date(now.getTime() + 90 * 24 * 60 * 60 * 1000),
        specialInstructions: 'Includes installation labor and warranty',
      },
    },
    {
      id: `po-${uuidv4()}`,
      type: 'PURCHASE_ORDER',
      documentNumber: 'PO-2024-008',
      status: 'APPROVED',
      currentStage: 4,
      createdBy: requester2.id,
      createdByUser: requester2,
      createdAt: new Date(now.getTime() - 15 * 24 * 60 * 60 * 1000),
      updatedAt: new Date(now.getTime() - 3 * 24 * 60 * 60 * 1000),
      metadata: {
        vendorName: 'Global Logistics Partners',
        vendorId: 'VENDOR-008',
        items: [
          {
            id: uuidv4(),
            description: 'Shipping and Logistics Services',
            quantity: 1,
            unitCost: 15000,
            totalCost: 15000,
          },
        ],
        totalAmount: 15000,
        currency: 'ZMW',
        deliveryDate: new Date(now.getTime() + 60 * 24 * 60 * 60 * 1000),
        specialInstructions: 'Quarterly service agreement',
      },
    },
    {
      id: `po-${uuidv4()}`,
      type: 'PURCHASE_ORDER',
      documentNumber: 'PO-2024-009',
      status: 'DRAFT',
      currentStage: 0,
      createdBy: requester1.id,
      createdByUser: requester1,
      createdAt: new Date(now.getTime() - 1 * 24 * 60 * 60 * 1000),
      updatedAt: new Date(now.getTime() - 1 * 24 * 60 * 60 * 1000),
      metadata: {
        vendorName: 'DataSecure Systems',
        vendorId: 'VENDOR-009',
        items: [
          {
            id: uuidv4(),
            description: 'Backup Storage Solutions (5TB)',
            quantity: 3,
            unitCost: 8000,
            totalCost: 24000,
          },
        ],
        totalAmount: 24000,
        currency: 'ZMW',
        deliveryDate: new Date(now.getTime() + 21 * 24 * 60 * 60 * 1000),
        specialInstructions: 'Enterprise grade backup solutions with 24/7 support',
      },
    },
    {
      id: `po-${uuidv4()}`,
      type: 'PURCHASE_ORDER',
      documentNumber: 'PO-2024-010',
      status: 'SUBMITTED',
      currentStage: 1,
      createdBy: requester1.id,
      createdByUser: requester1,
      createdAt: new Date(now.getTime() - 3 * 24 * 60 * 60 * 1000),
      updatedAt: new Date(now.getTime() - 3 * 24 * 60 * 60 * 1000),
      metadata: {
        vendorName: 'Healthcare Plus Supplies',
        vendorId: 'VENDOR-010',
        items: [
          {
            id: uuidv4(),
            description: 'First Aid Kits',
            quantity: 25,
            unitCost: 500,
            totalCost: 12500,
          },
          {
            id: uuidv4(),
            description: 'Medical PPE Equipment',
            quantity: 100,
            unitCost: 250,
            totalCost: 25000,
          },
        ],
        totalAmount: 37500,
        currency: 'ZMW',
        deliveryDate: new Date(now.getTime() + 14 * 24 * 60 * 60 * 1000),
        specialInstructions: 'Regular health and safety supplies',
      },
    },
  ];
}

/**
 * Create seed requisitions
 * Multiple users, departments, and statuses for search testing
 */
export function createSeedRequisitions(): RequisitionForm[] {
  const requesters = MOCK_USERS.REQUESTER || [];
  const requester1 = requesters[0] || { id: 'user-1', name: 'John Mwale', email: 'john@example.com' };
  const requester2 = requesters[1] || { id: 'user-2', name: 'Sarah Banda', email: 'sarah@example.com' };
  const now = new Date();

  return [
    {
      id: `req-${uuidv4()}`,
      type: 'REQUISITION',
      documentNumber: 'REQ-2024-001',
      status: 'DRAFT',
      currentStage: 0,
      createdBy: requester1.id,
      createdByUser: requester1,
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
    {
      id: `req-${uuidv4()}`,
      type: 'REQUISITION',
      documentNumber: 'REQ-2024-002',
      status: 'SUBMITTED',
      currentStage: 1,
      createdBy: requester2.id,
      createdByUser: requester2,
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
    {
      id: `req-${uuidv4()}`,
      type: 'REQUISITION',
      documentNumber: 'REQ-2024-003',
      status: 'APPROVED',
      currentStage: 4,
      createdBy: requester1.id,
      createdByUser: requester1,
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
    {
      id: `req-${uuidv4()}`,
      type: 'REQUISITION',
      documentNumber: 'REQ-2024-004',
      status: 'REJECTED',
      currentStage: 0,
      createdBy: requester2.id,
      createdByUser: requester2,
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
    // Additional requisitions for search testing
    {
      id: `req-${uuidv4()}`,
      type: 'REQUISITION',
      documentNumber: 'REQ-2024-005',
      status: 'IN_REVIEW',
      currentStage: 2,
      createdBy: requester1.id,
      createdByUser: requester1,
      createdAt: new Date(now.getTime() - 2 * 24 * 60 * 60 * 1000),
      updatedAt: new Date(now.getTime() - 1 * 24 * 60 * 60 * 1000),
      metadata: {
        department: 'Finance',
        requestedFor: 'Alice Mulenga',
        items: [
          {
            id: uuidv4(),
            itemDescription: 'Accounting Software Suite',
            quantity: 3,
            estimatedCost: 18000,
          },
        ],
        justification: 'Needed for Q1 budget reconciliation and audit preparation.',
        budgetCode: 'OPS-2024-001',
      },
    },
    {
      id: `req-${uuidv4()}`,
      type: 'REQUISITION',
      documentNumber: 'REQ-2024-006',
      status: 'SUBMITTED',
      currentStage: 1,
      createdBy: requester2.id,
      createdByUser: requester2,
      createdAt: new Date(now.getTime() - 1 * 24 * 60 * 60 * 1000),
      updatedAt: new Date(now.getTime() - 1 * 24 * 60 * 60 * 1000),
      metadata: {
        department: 'Operations',
        requestedFor: 'Charles Kaunda',
        items: [
          {
            id: uuidv4(),
            itemDescription: 'Floor Waxing and Maintenance',
            quantity: 1,
            estimatedCost: 3500,
          },
          {
            id: uuidv4(),
            itemDescription: 'Cleaning Supplies (Monthly)',
            quantity: 4,
            estimatedCost: 2000,
          },
        ],
        justification: 'Regular building maintenance and cleanliness standards.',
        budgetCode: 'MAINT-2024-001',
      },
    },
    {
      id: `req-${uuidv4()}`,
      type: 'REQUISITION',
      documentNumber: 'REQ-2024-007',
      status: 'APPROVED',
      currentStage: 4,
      createdBy: requester2.id,
      createdByUser: requester2,
      createdAt: new Date(now.getTime() - 10 * 24 * 60 * 60 * 1000),
      updatedAt: new Date(now.getTime() - 2 * 24 * 60 * 60 * 1000),
      metadata: {
        department: 'IT',
        requestedFor: 'David Mutale',
        items: [
          {
            id: uuidv4(),
            itemDescription: 'Network Monitoring Tools',
            quantity: 1,
            estimatedCost: 22000,
          },
        ],
        justification: 'Critical infrastructure improvement for network stability.',
        budgetCode: 'CAP-2024-003',
      },
    },
  ];
}

/**
 * Create seed payment vouchers
 * Multiple users, vendors, departments, and amounts for search testing
 */
export function createSeedPaymentVouchers(): PaymentVoucher[] {
  const requesters = MOCK_USERS.REQUESTER || [];
  const requester1 = requesters[0] || { id: 'user-1', name: 'John Mwale', email: 'john@example.com' };
  const requester2 = requesters[1] || { id: 'user-2', name: 'Sarah Banda', email: 'sarah@example.com' };
  const now = new Date();

  return [
    {
      id: `pv-${uuidv4()}`,
      type: 'PAYMENT_VOUCHER',
      documentNumber: 'PV-2024-001',
      status: 'DRAFT',
      currentStage: 0,
      createdBy: requester1.id,
      createdByUser: requester1,
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
    {
      id: `pv-${uuidv4()}`,
      type: 'PAYMENT_VOUCHER',
      documentNumber: 'PV-2024-002',
      status: 'SUBMITTED',
      currentStage: 1,
      createdBy: requester2.id,
      createdByUser: requester2,
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
    {
      id: `pv-${uuidv4()}`,
      type: 'PAYMENT_VOUCHER',
      documentNumber: 'PV-2024-003',
      status: 'IN_REVIEW',
      currentStage: 2,
      createdBy: requester1.id,
      createdByUser: requester1,
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
    {
      id: `pv-${uuidv4()}`,
      type: 'PAYMENT_VOUCHER',
      documentNumber: 'PV-2024-004',
      status: 'APPROVED',
      currentStage: 4,
      createdBy: requester2.id,
      createdByUser: requester2,
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
    // Additional payment vouchers for search testing
    {
      id: `pv-${uuidv4()}`,
      type: 'PAYMENT_VOUCHER',
      documentNumber: 'PV-2024-005',
      status: 'DRAFT',
      currentStage: 0,
      createdBy: requester1.id,
      createdByUser: requester1,
      createdAt: new Date(now.getTime() - 1 * 24 * 60 * 60 * 1000),
      updatedAt: new Date(now.getTime() - 1 * 24 * 60 * 60 * 1000),
      metadata: {
        payeeName: 'Premium Consulting Group',
        payeeId: 'VENDOR-006',
        amount: 25000,
        currency: 'ZMW',
        reason: 'Management training course delivery',
        accountCode: '4001-005',
        department: 'HR',
      },
    },
    {
      id: `pv-${uuidv4()}`,
      type: 'PAYMENT_VOUCHER',
      documentNumber: 'PV-2024-006',
      status: 'IN_REVIEW',
      currentStage: 2,
      createdBy: requester2.id,
      createdByUser: requester2,
      createdAt: new Date(now.getTime() - 5 * 24 * 60 * 60 * 1000),
      updatedAt: new Date(now.getTime() - 2 * 24 * 60 * 60 * 1000),
      metadata: {
        payeeName: 'Eco Solutions International',
        payeeId: 'VENDOR-007',
        amount: 125000,
        currency: 'ZMW',
        reason: 'Solar panel installation and setup',
        accountCode: '4001-006',
        department: 'Operations',
      },
    },
    {
      id: `pv-${uuidv4()}`,
      type: 'PAYMENT_VOUCHER',
      documentNumber: 'PV-2024-007',
      status: 'SUBMITTED',
      currentStage: 1,
      createdBy: requester1.id,
      createdByUser: requester1,
      createdAt: new Date(now.getTime() - 2 * 24 * 60 * 60 * 1000),
      updatedAt: new Date(now.getTime() - 2 * 24 * 60 * 60 * 1000),
      metadata: {
        payeeName: 'Global Logistics Partners',
        payeeId: 'VENDOR-008',
        amount: 15000,
        currency: 'ZMW',
        reason: 'Q1 2024 shipping and logistics services',
        accountCode: '4001-007',
        department: 'Operations',
      },
    },
    {
      id: `pv-${uuidv4()}`,
      type: 'PAYMENT_VOUCHER',
      documentNumber: 'PV-2024-008',
      status: 'APPROVED',
      currentStage: 4,
      createdBy: requester2.id,
      createdByUser: requester2,
      createdAt: new Date(now.getTime() - 12 * 24 * 60 * 60 * 1000),
      updatedAt: new Date(now.getTime() - 5 * 24 * 60 * 60 * 1000),
      metadata: {
        payeeName: 'DataSecure Systems',
        payeeId: 'VENDOR-009',
        amount: 24000,
        currency: 'ZMW',
        reason: 'Data backup and storage systems',
        accountCode: '4001-008',
        department: 'IT',
      },
    },
    {
      id: `pv-${uuidv4()}`,
      type: 'PAYMENT_VOUCHER',
      documentNumber: 'PV-2024-009',
      status: 'DRAFT',
      currentStage: 0,
      createdBy: requester1.id,
      createdByUser: requester1,
      createdAt: new Date(now.getTime() - 1 * 24 * 60 * 60 * 1000),
      updatedAt: new Date(now.getTime() - 1 * 24 * 60 * 60 * 1000),
      metadata: {
        payeeName: 'Healthcare Plus Supplies',
        payeeId: 'VENDOR-010',
        amount: 37500,
        currency: 'ZMW',
        reason: 'First aid kits and medical equipment supplies',
        accountCode: '4001-009',
        department: 'Operations',
      },
    },
  ];
}

/**
 * Create seed goods received notes linked to approved purchase orders
 * Multiple users, vendors, and statuses for search testing
 */
export function createSeedGoodsReceivedNotes(): GoodsReceivedNote[] {
  const requesters = MOCK_USERS.REQUESTER || [];
  const requester1 = requesters[0] || { id: 'user-1', name: 'John Mwale', email: 'john@example.com' };
  const requester2 = requesters[1] || { id: 'user-2', name: 'Sarah Banda', email: 'sarah@example.com' };
  const now = new Date();

  return [
    {
      id: `grn-${uuidv4()}`,
      type: 'GOODS_RECEIVED_NOTE',
      documentNumber: 'GRN-2024-001',
      status: 'IN_REVIEW',
      currentStage: 1,
      createdBy: requester1.id,
      createdByUser: requester1,
      createdAt: new Date(now.getTime() - 2 * 24 * 60 * 60 * 1000),
      updatedAt: new Date(now.getTime() - 1 * 24 * 60 * 60 * 1000),
      metadata: {
        poId: 'po-1',
        poNumber: 'PO-2024-004',
        vendorName: 'Maintenance Solutions Ltd',
        receivedQuantity: 1,
        totalQuantity: 1,
        amount: 5000.0,
        receivedDate: new Date(now.getTime() - 2 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
        warehouseLocation: 'WAREHOUSE-A-001',
      },
    },
    {
      id: `grn-${uuidv4()}`,
      type: 'GOODS_RECEIVED_NOTE',
      documentNumber: 'GRN-2024-002',
      status: 'APPROVED',
      currentStage: 1,
      createdBy: requester2.id,
      createdByUser: requester2,
      createdAt: new Date(now.getTime() - 5 * 24 * 60 * 60 * 1000),
      updatedAt: new Date(now.getTime() - 3 * 24 * 60 * 60 * 1000),
      metadata: {
        poId: 'po-2',
        poNumber: 'PO-2024-001',
        vendorName: 'Mitete Supplies Ltd',
        receivedQuantity: 25,
        totalQuantity: 25,
        amount: 18750.0,
        receivedDate: new Date(now.getTime() - 5 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
        warehouseLocation: 'WAREHOUSE-B-002',
      },
    },
    {
      id: `grn-${uuidv4()}`,
      type: 'GOODS_RECEIVED_NOTE',
      documentNumber: 'GRN-2024-003',
      status: 'APPROVED',
      currentStage: 1,
      createdBy: requester1.id,
      createdByUser: requester1,
      createdAt: new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000),
      updatedAt: new Date(now.getTime() - 5 * 24 * 60 * 60 * 1000),
      metadata: {
        poId: 'po-3',
        poNumber: 'PO-2024-002',
        vendorName: 'TechXpress Solutions',
        receivedQuantity: 5,
        totalQuantity: 5,
        amount: 43250.0,
        receivedDate: new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
        warehouseLocation: 'WAREHOUSE-A-003',
      },
    },
    // Additional GRNs for search testing
    {
      id: `grn-${uuidv4()}`,
      type: 'GOODS_RECEIVED_NOTE',
      documentNumber: 'GRN-2024-004',
      status: 'DRAFT',
      currentStage: 0,
      createdBy: requester2.id,
      createdByUser: requester2,
      createdAt: new Date(now.getTime() - 1 * 24 * 60 * 60 * 1000),
      updatedAt: new Date(now.getTime() - 1 * 24 * 60 * 60 * 1000),
      metadata: {
        poId: 'po-4',
        poNumber: 'PO-2024-006',
        vendorName: 'Premium Consulting Group',
        receivedQuantity: 0,
        totalQuantity: 1,
        amount: 25000.0,
        receivedDate: new Date(now.getTime() - 1 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
        warehouseLocation: 'TRAINING-CENTER-001',
      },
    },
    {
      id: `grn-${uuidv4()}`,
      type: 'GOODS_RECEIVED_NOTE',
      documentNumber: 'GRN-2024-005',
      status: 'IN_REVIEW',
      currentStage: 1,
      createdBy: requester1.id,
      createdByUser: requester1,
      createdAt: new Date(now.getTime() - 3 * 24 * 60 * 60 * 1000),
      updatedAt: new Date(now.getTime() - 2 * 24 * 60 * 60 * 1000),
      metadata: {
        poId: 'po-5',
        poNumber: 'PO-2024-007',
        vendorName: 'Eco Solutions International',
        receivedQuantity: 50,
        totalQuantity: 50,
        amount: 125000.0,
        receivedDate: new Date(now.getTime() - 3 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
        warehouseLocation: 'WAREHOUSE-B-004',
      },
    },
    {
      id: `grn-${uuidv4()}`,
      type: 'GOODS_RECEIVED_NOTE',
      documentNumber: 'GRN-2024-006',
      status: 'APPROVED',
      currentStage: 1,
      createdBy: requester2.id,
      createdByUser: requester2,
      createdAt: new Date(now.getTime() - 8 * 24 * 60 * 60 * 1000),
      updatedAt: new Date(now.getTime() - 6 * 24 * 60 * 60 * 1000),
      metadata: {
        poId: 'po-6',
        poNumber: 'PO-2024-008',
        vendorName: 'Global Logistics Partners',
        receivedQuantity: 1,
        totalQuantity: 1,
        amount: 15000.0,
        receivedDate: new Date(now.getTime() - 8 * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
        warehouseLocation: 'LOGISTICS-CENTER-001',
      },
    },
  ];
}
