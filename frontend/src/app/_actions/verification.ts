"use server";

import { axios } from "./api-config";

export interface VerificationDocument {
  verified: boolean;
  documentNumber: string;
  documentType: "REQUISITION" | "PURCHASE_ORDER" | "PAYMENT_VOUCHER" | "GRN" | "BUDGET";
  title: string;
  status: string;
  department?: string;
  totalAmount?: number;
  currency?: string;
  organizationId: string;
  organization?: string;
  createdByName?: string;
  createdAt: string;
  // Legacy fields for backward compatibility
  approvalStage?: number;
  totalApprovalStages?: number;
  requesterName?: string;
}

export interface VerificationResult {
  success: boolean;
  verified: boolean;
  document: VerificationDocument | null;
  message: string;
  verifiedAt: string;
}

/**
 * Verify a document by its document number
 * This is a public endpoint that doesn't require authentication
 */
export async function verifyDocument(
  documentNumber: string
): Promise<VerificationResult> {
  const verifiedAt = new Date().toISOString();

  try {
    // Try to find the document across different entity types
    // The backend should have a public verification endpoint
    const response = await axios.get(
      `/api/v1/public/verify/${encodeURIComponent(documentNumber)}`
    );

    if (response.data?.data) {
      return {
        success: true,
        verified: true,
        document: response.data.data,
        message: "Document verified successfully",
        verifiedAt,
      };
    }

    return {
      success: true,
      verified: false,
      document: null,
      message: "Document not found in the system",
      verifiedAt,
    };
  } catch (error: any) {
    console.error("Verification error:", error);

    // Handle 404 specifically
    if (error.status === 404 || error.response?.status === 404) {
      return {
        success: true,
        verified: false,
        document: null,
        message: "Document not found in the system",
        verifiedAt,
      };
    }

    return {
      success: false,
      verified: false,
      document: null,
      message: error.message || "An error occurred during verification",
      verifiedAt,
    };
  }
}

/**
 * Response from the document for PDF endpoint
 */
export interface DocumentForPDFResponse {
  documentType: "REQUISITION" | "PURCHASE_ORDER" | "PAYMENT_VOUCHER" | "GRN";
  document: any; // The full document data varies by type
}

/**
 * Fetch full document data for PDF generation
 * This is a public endpoint that doesn't require authentication
 */
export async function getDocumentForPDF(
  documentNumber: string
): Promise<DocumentForPDFResponse | null> {
  try {
    const response = await axios.get(
      `/api/v1/public/verify/${encodeURIComponent(documentNumber)}/document`
    );

    if (response.data?.data) {
      return response.data.data as DocumentForPDFResponse;
    }

    return null;
  } catch (error: any) {
    console.error("Error fetching document for PDF:", error);
    return null;
  }
}
