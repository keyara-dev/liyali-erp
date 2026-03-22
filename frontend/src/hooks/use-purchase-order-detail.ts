import { useDocumentDetail } from "./use-document-detail";
import { usePurchaseOrderById } from "./use-purchase-order-queries";
import {
  exportPurchaseOrderPDF,
  getPurchaseOrderPDFBlob,
} from "@/lib/pdf/pdf-export";
import { PurchaseOrder } from "@/types/purchase-order";

interface UsePurchaseOrderDetailProps {
  poId: string;
  userId: string;
  userRole: string;
  initialPurchaseOrder?: PurchaseOrder;
}

export function usePurchaseOrderDetail({
  poId,
  userId,
  userRole,
  initialPurchaseOrder,
}: UsePurchaseOrderDetailProps) {
  return useDocumentDetail<PurchaseOrder>({
    documentId: poId,
    userId,
    userRole,
    initialDocument: initialPurchaseOrder,
    documentType: "purchase-order",

    // Query hooks
    useDocumentQuery: usePurchaseOrderById,

    // Mutation hooks - PO typically doesn't have submit/withdraw in the same way
    useSubmitMutation: () => ({
      mutateAsync: async () => ({}),
      isPending: false,
    }),

    // PDF export
    exportPDF: exportPurchaseOrderPDF,
    getPDFBlob: getPurchaseOrderPDFBlob,

    // Permissions
    getPermissions: (po, userId, userRole) => {
      const isCreator = po.createdBy === userId;
      const canEdit =
        po.status?.toUpperCase() === "PENDING" && (isCreator || userRole === "admin");
      const canSubmit = false; // POs are typically created from approved requisitions
      const canWithdraw = false;

      return {
        isCreator,
        canEdit,
        canSubmit,
        canWithdraw,
      };
    },
  });
}
