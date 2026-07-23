"use server";

import { APIResponse } from "@/types";
import { handleError, successResponse } from "./api-config";
import authenticatedApiClient from "./api-config";

/**
 * Document type strings accepted by the backend document-chain endpoints
 * (backend/handlers/document_chain.go verifyDocumentOwnership /
 * resolveChainDocumentSet). The backend also accepts the underscored
 * variants ("purchase_order", "payment_voucher") but every frontend caller
 * standardizes on the hyphenated form used by LinkedDocType
 * (@/components/linked-documents) for consistency.
 */
export type ChainDocumentType =
  | "requisition"
  | "purchase-order"
  | "grn"
  | "payment-voucher";

/** One supporting-document attachment aggregated from anywhere in a
 * procurement chain (requisition, PO, GRN, or PV), including proof of
 * payment. Mirrors backend ChainAttachment (document_chain.go). */
export interface ChainAttachment {
  kind: "attachment" | "quotation" | "proof_of_payment";
  sourceDocType: string;
  sourceDocId: string;
  sourceDocNumber: string;
  fileId?: string;
  fileName: string;
  fileUrl?: string; // absent for proof_of_payment — never a downloadable URL
  fileSize?: number;
  mimeType?: string;
  uploadedAt?: string;
  uploadedBy?: string;
  fromRequisition?: boolean;
  category?: string;
  downloadRef?: string; // proof_of_payment: "/payment-vouchers/{id}"
}

/** Compact reference to one document in the resolved chain. Mirrors backend
 * chainDocRef (document_chain.go). */
export interface ChainDocumentRef {
  id: string;
  type: string;
  documentNumber: string;
}

export interface ChainAttachmentsResponse {
  attachments: ChainAttachment[];
  documents: ChainDocumentRef[];
}

/**
 * Fetch every supporting-document attachment aggregated across the full
 * procurement chain anchored on one document (REQ -> PO -> GRN(s) -> PV(s)),
 * including each PV's proof of payment.
 * Calls: GET /api/v1/document-chain/:id/attachments?documentType=...
 */
export async function getChainAttachments(
  docId: string,
  docType: ChainDocumentType,
): Promise<APIResponse<ChainAttachmentsResponse>> {
  const url = `/api/v1/document-chain/${docId}/attachments?documentType=${docType}`;

  try {
    const response = await authenticatedApiClient({
      method: "GET",
      url,
    });

    return successResponse(
      response.data?.data,
      "Chain attachments retrieved successfully",
    );
  } catch (error: any) {
    return handleError(error, "GET", url);
  }
}

/**
 * Fetch the document chain (parent/child documents) for any document type.
 * Calls: GET /api/v1/document-chain/:id?documentType=...
 *
 * This is the same generic endpoint usePurchaseOrderChain /
 * usePaymentVoucherChain already call (getPurchaseOrderChain,
 * getPaymentVoucherChain in purchase-orders.ts / payment-vouchers.ts),
 * exposed here generically so the requisition detail page can move off the
 * dead /requisitions/:id/chain (legacy DocumentLink-table) endpoint — see
 * GetDocumentChain in backend/handlers/document_chain.go.
 */
export async function getDocumentChain(
  docId: string,
  docType: ChainDocumentType,
): Promise<APIResponse<any>> {
  const url = `/api/v1/document-chain/${docId}?documentType=${docType}`;

  try {
    const response = await authenticatedApiClient({
      method: "GET",
      url,
    });

    return successResponse(
      response.data?.data,
      "Document chain retrieved successfully",
    );
  } catch (error: any) {
    return handleError(error, "GET", url);
  }
}
