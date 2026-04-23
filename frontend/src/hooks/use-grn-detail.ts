import { useDocumentDetail } from "./use-document-detail";
import { useGRNById } from "./use-grn-queries";
import { exportGrnPDF, getGrnPDFBlob } from "@/lib/pdf/pdf-export";
import { submitGRNForApproval } from "@/app/_actions/grn-actions";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { QUERY_KEYS } from "@/lib/constants";
import { useSession } from "@/hooks/use-session";
import { toast } from "sonner";
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
  const queryClient = useQueryClient();
  const { user } = useSession();
  const userName = user?.name || "User";

  return useDocumentDetail<any>({
    documentId: grnId,
    userId,
    userRole,
    initialDocument: initialGRN,
    documentType: "grn",

    // Query hooks
    useDocumentQuery: useGRNById as any,

    // Mutation hooks — use a real useMutation so isPending drives the button
    // loading state and success/error toasts fire from a single place.
    useSubmitMutation: (id: string, onSuccess: () => void) =>
      // eslint-disable-next-line react-hooks/rules-of-hooks
      useMutation({
        mutationFn: async (data: { workflowId: string; comments?: string }) => {
          const result = await submitGRNForApproval({
            grnId: id,
            workflowId: data.workflowId,
            submittedBy: userId,
            submittedByName: userName,
            submittedByRole: userRole,
            comments: data.comments,
          });
          if (!result.success) {
            throw new Error(result.message || "Failed to submit GRN");
          }
          return result;
        },
        onSuccess: () => {
          toast.success("GRN submitted for approval");
          queryClient.invalidateQueries({
            queryKey: [QUERY_KEYS.GRN.BY_ID, id],
          });
          queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.GRN.ALL] });
          onSuccess();
        },
        onError: (err: Error) => {
          toast.error(err.message || "Failed to submit GRN for approval");
        },
      }),

    // PDF export
    exportPDF: exportGrnPDF as any,
    getPDFBlob: getGrnPDFBlob as any,

    // Permissions
    getPermissions: (grn, userId, _userRole) => {
      const isCreator = grn.receivedBy === userId || grn.createdBy === userId;
      const grnStatus = grn.status?.toUpperCase();
      const canEdit = grnStatus === "DRAFT" && isCreator;
      const canSubmit = grnStatus === "DRAFT" && isCreator;
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
