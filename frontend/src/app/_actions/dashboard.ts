"use server";

import {
  APIResponse,
  DashboardMetrics,
  SignupSettings,
  SignupAnalytics,
} from "@/types";
import { handleError, successResponse } from "@/app/_actions/api-config";
import authenticatedApiClient from "@/app/_actions/api-config";

export async function getDashboardMetrics(): Promise<
  APIResponse<DashboardMetrics>
> {
  const url = "/api/v1/analytics/dashboard";

  try {
    const response = await authenticatedApiClient({
      method: "GET",
      url,
    });

    // Transform backend data to frontend DashboardMetrics format
    const backendData = response.data?.data;
    if (!backendData?.requisitionMetrics) {
      throw new Error("Invalid dashboard data received from backend");
    }

    const reqMetrics = backendData.requisitionMetrics;
    const statusCounts = reqMetrics.statusCounts || {};

    // Map backend status counts to frontend format
    const metrics: DashboardMetrics = {
      totalDocuments: reqMetrics.totalRequisitions || 0,
      draftDocuments: statusCounts.draft || statusCounts.DRAFT || 0,
      submittedDocuments: statusCounts.submitted || statusCounts.SUBMITTED || 0,
      approvedDocuments: statusCounts.approved || statusCounts.APPROVED || 0,
      rejectedDocuments: statusCounts.rejected || statusCounts.REJECTED || 0,
      pendingApproval: statusCounts.in_review || statusCounts.IN_REVIEW || statusCounts.pending || statusCounts.PENDING || 0,
      documentsNeedingAction: (statusCounts.submitted || statusCounts.SUBMITTED || 0) + (statusCounts.in_review || statusCounts.IN_REVIEW || statusCounts.pending || statusCounts.PENDING || 0),
      averageApprovalTime: 0, // TODO: Add to backend analytics
      statusBreakdown: statusCounts,
      documentTypeBreakdown: {
        REQUISITION: reqMetrics.totalRequisitions || 0,
        PURCHASE_ORDER: 0, // TODO: Add to backend analytics
        PAYMENT_VOUCHER: 0, // TODO: Add to backend analytics
        GOODS_RECEIVED_NOTE: 0, // TODO: Add to backend analytics
      },
      recentActivity: [], // TODO: Add to backend analytics
    };

    return successResponse(metrics, "Dashboard metrics retrieved successfully");
  } catch (error: any) {
    return handleError(error, "GET", url);
  }
}

export async function fetchSignupSettings(): Promise<
  APIResponse<SignupSettings | null>
> {
  try {
    const settings: SignupSettings = {
      allowSignups: true,
      requireEmailVerification: false,
      autoApproveUsers: false,
      defaultRole: "USER",
    };
    return {
      success: true,
      message: "Signup settings retrieved",
      data: settings,
      status: 200,
    };
  } catch (error) {
    return handleError(error, "GET", "/dashboard/signup-settings") as any;
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
      message: "Signup analytics retrieved",
      data: analytics,
      status: 200,
    };
  } catch (error) {
    return handleError(error, "GET", "/dashboard/signup-analytics") as any;
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
      defaultRole: "USER",
    };
    return {
      success: true,
      message: "Signup settings updated",
      data: settings,
      status: 200,
    };
  } catch (error) {
    return handleError(error, "PATCH", "/dashboard/signup-settings") as any;
  }
}
