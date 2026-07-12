import { useDocumentDetail } from "./use-document-detail";
import {
  useRequisitionById,
  useSubmitRequisitionForApproval,
} from "./use-requisition-queries";
import { useDocumentChain } from "./use-document-chain-queries";
import { useWithdrawRequisition } from "./use-requisition-mutations";
import { useRequisitionStorage } from "./use-requisition-storage";
import { useApprovalPanelData } from "./use-approval-history";
import {
  exportRequisitionPDF,
  getRequisitionPDFBlob,
} from "@/lib/pdf/pdf-export";
import { Requisition, RequisitionAttachment } from "@/types/requisition";

interface UseRequisitionDetailProps {
  requisitionId: string;
  userId: string;
  userRole: string;
  initialRequisition?: Requisition;
}

export function useRequisitionDetail({
  requisitionId,
  userId,
  userRole,
  initialRequisition,
}: UseRequisitionDetailProps) {
  return useDocumentDetail<Requisition, RequisitionAttachment>({
    documentId: requisitionId,
    userId,
    userRole,
    initialDocument: initialRequisition,
    documentType: "requisition",

    // Query hooks
    useDocumentQuery: useRequisitionById,
    // Generic document-chain endpoint (documentType=requisition) — the old
    // /requisitions/:id/chain endpoint reads a legacy DocumentLink table
    // nothing in the live flow populates. See getDocumentChain,
    // use-document-chain-queries.ts.
    useChainQuery: (id: string) => useDocumentChain(id, "requisition"),
    useApprovalDataQuery: useApprovalPanelData,

    // Mutation hooks
    useSubmitMutation: useSubmitRequisitionForApproval,
    useWithdrawMutation: useWithdrawRequisition,

    // PDF export
    exportPDF: exportRequisitionPDF,
    getPDFBlob: getRequisitionPDFBlob,

    // Storage
    useStorage: useRequisitionStorage,

    // Permissions
    getPermissions: (requisition, userId, _userRole) => {
      const isCreator =
        requisition.requestedBy === userId ||
        requisition.requesterId === userId;
      const reqStatus = requisition.status?.toUpperCase();
      const canEdit =
        isCreator && (reqStatus === "DRAFT" || reqStatus === "REJECTED");
      const canSubmit = reqStatus === "DRAFT" && isCreator;
      const canWithdraw = reqStatus === "PENDING" && isCreator;

      return {
        isCreator,
        canEdit,
        canSubmit,
        canWithdraw,
      };
    },

    // Auto-routing after submit
    getAutoRouteAfterSubmit: (result) => {
      const responseData = result?.data;
      const routingData = responseData?.routing;
      const autoCreatedPO = responseData?.autoCreatedPO;

      if (routingData?.autoApproved && autoCreatedPO?.id) {
        return `/purchase-orders/${autoCreatedPO.id}`;
      }
      return null;
    },
  });
}
