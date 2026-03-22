import { useDocumentDetail } from "./use-document-detail";
import { useGRNById } from "./use-grn-queries";
import { exportGrnPDF, getGrnPDFBlob } from "@/lib/pdf/pdf-export";
import type { GoodsReceivedNote } from "@/types/goods-received-note";

interface UseGRNDetailProps {
  grnId: string;
  userId: string;
  userRole: string;
  initialGRN?: any; // Use any to avoid type conflicts between action and type definitions
}

export function useGRNDetail({
  grnId,
  userId,
  userRole,
  initialGRN,
}: UseGRNDetailProps) {
  return useDocumentDetail<any>({
    documentId: grnId,
    userId,
    userRole,
    initialDocument: initialGRN,
    documentType: "grn",

    // Query hooks
    useDocumentQuery: useGRNById as any,

    // Mutation hooks - GRN typically doesn't have submit/withdraw in the same way
    useSubmitMutation: () => ({
      mutateAsync: async () => ({}),
      isPending: false,
    }),

    // PDF export
    exportPDF: exportGrnPDF as any,
    getPDFBlob: getGrnPDFBlob as any,

    // Permissions
    getPermissions: (grn, userId, userRole) => {
      const isCreator = grn.receivedBy === userId || grn.createdBy === userId;
      const canEdit = grn.status?.toUpperCase() === "DRAFT" && isCreator;
      const canSubmit = grn.status?.toUpperCase() === "SUBMITTED" && userRole === "admin";
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
