'use server'

import { auth } from '@/auth';
import { APIResponse, DashboardMetrics, SignupSettings, SignupAnalytics } from '@/types';
import { documentStore } from '@/lib/workflow-stores';
import { unauthorizedResponse, handleError } from '@/app/_actions/api-config';

export async function getDashboardMetrics(): Promise<APIResponse<DashboardMetrics>> {
  const session = await auth();

  if (!session?.user) {
    return unauthorizedResponse();
  }

  try {
    const allDocuments = Array.from(documentStore.values());

    // Calculate status breakdown
    const statusBreakdown: Record<string, number> = {
      DRAFT: 0,
      SUBMITTED: 0,
      IN_APPROVAL: 0,
      APPROVED: 0,
      REJECTED: 0,
      REVERSED: 0,
    };

    const documentTypeBreakdown: Record<string, number> = {
      REQUISITION: 0,
      PURCHASE_ORDER: 0,
      PAYMENT_VOUCHER: 0,
      GOODS_RECEIVED_NOTE: 0,
    };

    let totalDocuments = 0;
    let pendingApproval = 0;
    let documentsNeedingAction = 0;

    allDocuments.forEach((doc) => {
      totalDocuments++;
      statusBreakdown[doc.status]++;
      documentTypeBreakdown[doc.type]++;

      if (doc.status === 'IN_APPROVAL') {
        pendingApproval++;
        documentsNeedingAction++;
      } else if (doc.status === 'SUBMITTED') {
        documentsNeedingAction++;
      }
    });

    // Calculate average approval time (mock calculation)
    const approvedDocs = allDocuments.filter((d) => d.status === 'APPROVED');
    const averageApprovalTime =
      approvedDocs.length > 0
        ? approvedDocs.reduce((sum, doc) => {
            const days = Math.floor((doc.updatedAt.getTime() - doc.createdAt.getTime()) / (1000 * 60 * 60 * 24));
            return sum + days;
          }, 0) / approvedDocs.length
        : 0;

    // Get recent activity (last 5 documents by update time)
    const recentActivity = allDocuments
      .sort((a, b) => b.updatedAt.getTime() - a.updatedAt.getTime())
      .slice(0, 5)
      .map((doc) => ({
        id: doc.id,
        type: doc.type,
        documentNumber: doc.documentNumber,
        action: doc.status,
        timestamp: doc.updatedAt,
        user: doc.createdByUser?.name || 'Unknown User',
      }));

    const metrics: DashboardMetrics = {
      totalDocuments,
      draftDocuments: statusBreakdown.DRAFT,
      submittedDocuments: statusBreakdown.SUBMITTED,
      approvedDocuments: statusBreakdown.APPROVED,
      rejectedDocuments: statusBreakdown.REJECTED,
      pendingApproval,
      documentsNeedingAction,
      averageApprovalTime: Math.round(averageApprovalTime),
      statusBreakdown,
      documentTypeBreakdown,
      recentActivity,
    };

    return {
      success: true,
      message: 'Dashboard metrics retrieved',
      data: metrics,
      status: 200,
    };
  } catch (error) {
    return handleError(error, 'GET', '/dashboard/metrics') as any;
  }
}

export async function fetchSignupSettings(): Promise<APIResponse<SignupSettings | null>> {
  try {
    const settings: SignupSettings = {
      allowSignups: true,
      requireEmailVerification: false,
      autoApproveUsers: false,
      defaultRole: 'USER',
    };
    return {
      success: true,
      message: 'Signup settings retrieved',
      data: settings,
      status: 200,
    };
  } catch (error) {
    return handleError(error, 'GET', '/dashboard/signup-settings') as any;
  }
}

export async function fetchSignupAnalytics(params?: {
  start?: string | Date;
  end?: string | Date;
}): Promise<APIResponse<SignupAnalytics | null>> {
  try {
    const analytics: SignupAnalytics = {
      totalSignups: 0,
      recentSignups: 0,
      pendingApprovals: 0,
      rejectedCount: 0,
    };
    return {
      success: true,
      message: 'Signup analytics retrieved',
      data: analytics,
      status: 200,
    };
  } catch (error) {
    return handleError(error, 'GET', '/dashboard/signup-analytics') as any;
  }
}

export async function toggleSignupSettings(
  keyOrEnabled: keyof SignupSettings | boolean,
  value?: any
): Promise<APIResponse<SignupSettings | null>> {
  try {
    const settings: SignupSettings = {
      allowSignups: true,
      requireEmailVerification: false,
      autoApproveUsers: false,
      defaultRole: 'USER',
    };
    return {
      success: true,
      message: 'Signup settings updated',
      data: settings,
      status: 200,
    };
  } catch (error) {
    return handleError(error, 'PATCH', '/dashboard/signup-settings') as any;
  }
}
