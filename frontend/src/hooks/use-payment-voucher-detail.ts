import { useDocumentDetail } from "./use-document-detail";
import { usePaymentVoucherById } from "./use-payment-voucher-queries";
import {
  exportPaymentVoucherPDF,
  getPaymentVoucherPDFBlob,
} from "@/lib/pdf/pdf-export";
import { PaymentVoucher } from "@/types/payment-voucher";

interface UsePaymentVoucherDetailProps {
  pvId: string;
  userId: string;
  userRole: string;
  initialPaymentVoucher?: PaymentVoucher;
}

export function usePaymentVoucherDetail({
  pvId,
  userId,
  userRole,
  initialPaymentVoucher,
}: UsePaymentVoucherDetailProps) {
  return useDocumentDetail<PaymentVoucher>({
    documentId: pvId,
    userId,
    userRole,
    initialDocument: initialPaymentVoucher,
    documentType: "payment-voucher",

    // Query hooks
    useDocumentQuery: usePaymentVoucherById,

    // Mutation hooks - PV typically doesn't have submit/withdraw in the same way
    useSubmitMutation: () => ({
      mutateAsync: async () => ({}),
      isPending: false,
    }),

    // PDF export
    exportPDF: exportPaymentVoucherPDF,
    getPDFBlob: getPaymentVoucherPDFBlob,

    // Permissions
    getPermissions: (pv, userId, userRole) => {
      const isCreator = pv.createdBy === userId;
      const canEdit =
        pv.status === "IN_REVIEW" && (isCreator || userRole === "admin");
      const canSubmit = false; // PVs are typically created from approved POs
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
