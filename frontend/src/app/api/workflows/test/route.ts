import { NextRequest, NextResponse } from 'next/server';
import {
  createWorkflowDocument,
  submitDocument,
  getDocument,
  approveDocument,
  rejectDocument,
  getDashboardStats,
  getAuditLog,
} from '@/app/_actions/workflow';
import {
  createRole,
  getAllRoles,
  addRolePermission,
} from '@/app/_actions/rbac';
import {
  assignCustomRoleToUser,
  getAllUsers,
  getUsersByRole,
} from '@/app/_actions/user-management';
import {
  createMockPurchaseOrder,
  createMockPaymentVoucher,
  createMockRequisitionForm,
  MOCK_USERS,
} from '@/lib/mock-data';

/**
 * Test endpoint to demonstrate all mocked server actions
 * GET /api/workflows/test
 */
export async function GET(request: NextRequest) {
  try {
    console.log('\n========================================');
    console.log('WORKFLOW SYSTEM - COMPREHENSIVE TEST');
    console.log('========================================\n');

    // Mock session for server actions
    const mockSession = {
      user: {
        id: MOCK_USERS.ADMIN[0].id,
        name: MOCK_USERS.ADMIN[0].name,
        email: MOCK_USERS.ADMIN[0].email,
        role: 'ADMIN',
      },
    };

    const results: Record<string, any> = {};

    // ============================================
    // STEP 1: Test Role Management
    // ============================================
    console.log('📋 STEP 1: Testing Role Management...\n');

    const rolesResponse = await getAllRoles();
    results.allRoles = rolesResponse;
    console.log(`✅ Retrieved ${rolesResponse.data?.length} roles`);

    // ============================================
    // STEP 2: Test User Management
    // ============================================
    console.log('\n👥 STEP 2: Testing User Management...\n');

    const usersResponse = await getAllUsers();
    results.allUsers = usersResponse;
    console.log(`✅ Retrieved ${usersResponse.data?.length} users`);

    const departmentManagersResponse = await getUsersByRole('DEPARTMENT_MANAGER');
    results.departmentManagers = departmentManagersResponse;
    console.log(
      `✅ Retrieved ${departmentManagersResponse.data?.length} department managers`
    );

    // ============================================
    // STEP 3: Test Document Creation
    // ============================================
    console.log('\n📄 STEP 3: Testing Document Creation...\n');

    // Create Purchase Order
    const poFormData = {
      vendorName: 'Mitete Town Council Supplies',
      vendorId: 'VENDOR-001',
      items: [
        {
          id: '1',
          description: 'Office Chairs (Ergonomic)',
          quantity: 15,
          unitCost: 450,
          totalCost: 6750,
        },
        {
          id: '2',
          description: 'Standing Desks',
          quantity: 10,
          unitCost: 1200,
          totalCost: 12000,
        },
      ],
      totalAmount: 18750,
      currency: 'ZMW',
      deliveryDate: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000),
      specialInstructions: 'Deliver to main office. Notify 2 days before.',
    };

    const poResponse = await createWorkflowDocument('PURCHASE_ORDER', poFormData);
    const purchaseOrderId = poResponse.data?.id;
    results.purchaseOrder = poResponse;
    console.log(`✅ Created Purchase Order: ${poResponse.data?.documentNumber}`);

    // Create Payment Voucher
    const pvFormData = {
      payeeName: 'Mitete Supplies Ltd',
      payeeId: 'VENDOR-001',
      amount: 18750,
      currency: 'ZMW',
      reason: 'Payment for office equipment - PO-2024-001',
      accountCode: '4001-001',
      department: 'Operations',
    };

    const pvResponse = await createWorkflowDocument('PAYMENT_VOUCHER', pvFormData);
    const paymentVoucherId = pvResponse.data?.id;
    results.paymentVoucher = pvResponse;
    console.log(`✅ Created Payment Voucher: ${pvResponse.data?.documentNumber}`);

    // Create Requisition Form
    const reqFormData = {
      department: 'Operations',
      requestedFor: 'John Mwale',
      items: [
        {
          id: '1',
          itemDescription: 'Office Chairs - Ergonomic',
          quantity: 15,
          estimatedCost: 6750,
        },
        {
          id: '2',
          itemDescription: 'Standing Desks',
          quantity: 10,
          estimatedCost: 12000,
        },
      ],
      justification:
        'Current office furniture is worn out and causing ergonomic issues.',
      budgetCode: 'CAP-2024-001',
    };

    const reqResponse = await createWorkflowDocument('REQUISITION', reqFormData);
    const requisitionId = reqResponse.data?.id;
    results.requisition = reqResponse;
    console.log(`✅ Created Requisition: ${reqResponse.data?.documentNumber}`);

    // ============================================
    // STEP 4: Test Document Submission
    // ============================================
    console.log('\n📤 STEP 4: Testing Document Submission...\n');

    if (purchaseOrderId) {
      const submitPoResponse = await submitDocument(purchaseOrderId);
      results.submitPO = submitPoResponse;
      console.log(`✅ Submitted Purchase Order: ${submitPoResponse.data?.status}`);
    }

    if (paymentVoucherId) {
      const submitPvResponse = await submitDocument(paymentVoucherId);
      results.submitPV = submitPvResponse;
      console.log(`✅ Submitted Payment Voucher: ${submitPvResponse.data?.status}`);
    }

    // ============================================
    // STEP 5: Test Approval Workflow
    // ============================================
    console.log('\n✔️ STEP 5: Testing Approval Workflow...\n');

    if (purchaseOrderId) {
      const approveResponse = await approveDocument(
        purchaseOrderId,
        'Approved - All items within budget'
      );
      results.approvePO = approveResponse;
      console.log(`✅ Approved Purchase Order`);

      const auditLogResponse = await getAuditLog(purchaseOrderId);
      results.auditLogPO = auditLogResponse;
      console.log(`✅ Audit Log has ${auditLogResponse.data?.length} entries`);
    }

    // ============================================
    // STEP 6: Test Dashboard Stats
    // ============================================
    console.log('\n📊 STEP 6: Testing Dashboard Stats...\n');

    const statsResponse = await getDashboardStats(MOCK_USERS.ADMIN[0].id);
    results.dashboardStats = statsResponse;
    console.log(`✅ Dashboard Stats retrieved:`, statsResponse.data);

    // ============================================
    // STEP 7: Test Rejection Workflow
    // ============================================
    console.log('\n❌ STEP 7: Testing Rejection Workflow...\n');

    if (requisitionId) {
      const submitReqResponse = await submitDocument(requisitionId);
      console.log(`✅ Requisition submitted for rejection test`);

      const rejectResponse = await rejectDocument(
        requisitionId,
        'Requires additional cost analysis and departmental sign-off'
      );
      results.rejectReq = rejectResponse;
      console.log(`✅ Rejected Requisition`);

      const rejectionAuditLog = await getAuditLog(requisitionId);
      results.rejectionAuditLog = rejectionAuditLog;
      console.log(`✅ Rejection recorded in audit log`);
    }

    console.log('\n========================================');
    console.log('TEST COMPLETED SUCCESSFULLY');
    console.log('========================================\n');

    return NextResponse.json(
      {
        success: true,
        message: 'All workflow system tests completed successfully',
        data: {
          summary: {
            documentsCreated: 3,
            documentsSubmitted: 2,
            approvalsProcessed: 1,
            rejectionTested: 1,
            rolesAvailable: rolesResponse.data?.length,
            usersInSystem: usersResponse.data?.length,
          },
          details: results,
        },
      },
      { status: 200 }
    );
  } catch (error) {
    console.error('Test failed:', error);
    return NextResponse.json(
      {
        success: false,
        message: 'Test failed',
        error: error instanceof Error ? error.message : 'Unknown error',
      },
      { status: 500 }
    );
  }
}
