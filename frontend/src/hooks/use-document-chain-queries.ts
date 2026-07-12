import { useQuery } from "@tanstack/react-query";
import {
  getChainAttachments,
  getDocumentChain,
  type ChainDocumentType,
  type ChainAttachmentsResponse,
} from "@/app/_actions/document-chain";

/**
 * Fetch every supporting-document attachment aggregated across the full
 * procurement chain anchored on one document (REQ -> PO -> GRN(s) -> PV(s)).
 * Backs the SupportingDocuments component's aggregated-files zone.
 *
 * Pattern: usePurchaseOrderChain, use-purchase-order-queries.ts:305-319.
 *
 * @example
 * const { data } = useChainAttachments(poId, "purchase-order")
 */
export const useChainAttachments = (
  docId: string,
  docType: ChainDocumentType,
) =>
  useQuery({
    queryKey: ["document-chain", docId, "attachments"],
    queryFn: async () => {
      const response = await getChainAttachments(docId, docType);
      if (!response.success) throw new Error(response.message);
      return (
        (response.data as ChainAttachmentsResponse) ?? {
          attachments: [],
          documents: [],
        }
      );
    },
    enabled: !!docId,
    staleTime: 30 * 1000, // 30 seconds
  });

/**
 * Generic document chain (parent/child documents) for any document type —
 * same shape/endpoint as usePurchaseOrderChain / usePaymentVoucherChain.
 * Used by the requisition detail page in place of the dead
 * /requisitions/:id/chain endpoint. See getDocumentChain.
 */
export const useDocumentChain = (
  docId: string,
  docType: ChainDocumentType,
  initialData?: any,
) =>
  useQuery({
    queryKey: ["document-chain", docId, docType],
    queryFn: async () => {
      const response = await getDocumentChain(docId, docType);
      if (!response.success) throw new Error(response.message);
      return response.data;
    },
    initialData,
    enabled: !!docId,
    staleTime: 30 * 1000, // 30 seconds
  });
