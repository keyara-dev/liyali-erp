/**
 * Seed Data for Development
 * Creates initial mock data for all document types
 *
 * Structure:
 * - Purchase Orders: Various statuses (DRAFT, SUBMITTED, IN_REVIEW, APPROVED, REJECTED)
 * - Requisitions: Various statuses and departments
 * - Payment Vouchers: Various stages in approval workflow
 */

import { PurchaseOrder, PaymentVoucher, RequisitionForm } from '@/types/workflow';
import { MOCK_USERS } from '../mock-data';
import { v4 as uuidv4 } from 'uuid';

/**
 * Create seed purchase orders with various statuses
 */
export function createSeedPurchaseOrders(): PurchaseOrder[] {
  const requester = MOCK_USERS.REQUESTER[0];
  const now = new Date();

  return [
    {
      id: `po-${uuidv4()}`,
      type: 'PURCHASE_ORDER',
      documentNumber: 'PO-2024-001',
      status: 'DRAFT',
      currentStage: 0,
      createdBy: requester.id,
      createdByUser: requester,
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
      createdBy: requester.id,
      createdByUser: requester,
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
      createdBy: requester.id,
      createdByUser: requester,
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
      createdBy: requester.id,
      createdByUser: requester,
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
      createdBy: requester.id,
      createdByUser: requester,
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
  ];
}

/**
 * Create seed requisitions
 */
export function createSeedRequisitions(): RequisitionForm[] {
  const requester = MOCK_USERS.REQUESTER[0];
  const now = new Date();

  return [
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
 * Create seed payment vouchers
 */
export function createSeedPaymentVouchers(): PaymentVoucher[] {
  const requester = MOCK_USERS.REQUESTER[0];
  const now = new Date();

  return [
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
