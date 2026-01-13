/**
 * Email service for sending PDFs as attachments
 * Handles document distribution via email with PDF exports
 */

import {
  exportRequisitionPDF,
  exportPurchaseOrderPDF,
  exportPaymentVoucherPDF,
  getRequisitionPDFBlob,
  getPurchaseOrderPDFBlob,
  getPaymentVoucherPDFBlob,
} from "./pdf-export";
import { Requisition } from "@/types/requisition";
import { PurchaseOrder } from "@/types/purchase-order";
import { PaymentVoucher } from "@/types/payment-voucher";

export interface EmailRecipient {
  email: string;
  name: string;
}

export interface EmailOptions {
  subject: string;
  body: string;
  recipients: EmailRecipient[];
  cc?: EmailRecipient[];
  bcc?: EmailRecipient[];
}

/**
 * Send requisition PDF via email
 */
export async function sendRequisitionPDFEmail(
  requisition: Requisition,
  options: EmailOptions
): Promise<{ success: boolean; message: string }> {
  try {
    const blob = await getRequisitionPDFBlob(requisition);
    return await sendPDFEmail(
      blob,
      `REQ-${requisition.documentNumber}.pdf`,
      options
    );
  } catch (error) {
    console.error("Error sending requisition email:", error);
    return {
      success: false,
      message: "Failed to send requisition email",
    };
  }
}

/**
 * Send purchase order PDF via email
 */
export async function sendPurchaseOrderPDFEmail(
  purchaseOrder: PurchaseOrder,
  options: EmailOptions
): Promise<{ success: boolean; message: string }> {
  try {
    const blob = await getPurchaseOrderPDFBlob(purchaseOrder);
    return await sendPDFEmail(
      blob,
      `PO-${purchaseOrder.documentNumber}.pdf`,
      options
    );
  } catch (error) {
    console.error("Error sending purchase order email:", error);
    return {
      success: false,
      message: "Failed to send purchase order email",
    };
  }
}

/**
 * Send payment voucher PDF via email
 */
export async function sendPaymentVoucherPDFEmail(
  paymentVoucher: PaymentVoucher,
  options: EmailOptions
): Promise<{ success: boolean; message: string }> {
  try {
    const blob = await getPaymentVoucherPDFBlob(paymentVoucher);
    return await sendPDFEmail(
      blob,
      `PV-${paymentVoucher.documentNumber}.pdf`,
      options
    );
  } catch (error) {
    console.error("Error sending payment voucher email:", error);
    return {
      success: false,
      message: "Failed to send payment voucher email",
    };
  }
}

/**
 * Generic PDF email sender
 * In production, this would call your email service API
 */
async function sendPDFEmail(
  pdfBlob: Blob,
  fileName: string,
  options: EmailOptions
): Promise<{ success: boolean; message: string }> {
  try {
    // Convert blob to base64 for API transmission
    const buffer = await pdfBlob.arrayBuffer();
    const base64 = Buffer.from(buffer).toString("base64");

    // Call your email service API endpoint
    const response = await fetch("/api/email/send-with-attachment", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        recipients: options.recipients,
        cc: options.cc,
        bcc: options.bcc,
        subject: options.subject,
        body: options.body,
        attachment: {
          filename: fileName,
          content: base64,
          contentType: "application/pdf",
        },
      }),
    });

    if (!response.ok) {
      throw new Error(`Email API error: ${response.statusText}`);
    }

    const data = await response.json();
    return {
      success: true,
      message: `Email sent successfully to ${options.recipients.length} recipient(s)`,
    };
  } catch (error) {
    console.error("Error sending PDF email:", error);
    return {
      success: false,
      message: error instanceof Error ? error.message : "Failed to send email",
    };
  }
}

/**
 * Format email recipients list for display
 */
export function formatRecipientsDisplay(recipients: EmailRecipient[]): string {
  return recipients.map((r) => `${r.name} <${r.email}>`).join(", ");
}

/**
 * Validate email address
 */
export function isValidEmail(email: string): boolean {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
}

/**
 * Build email body for document
 */
export function buildDocumentEmailBody(
  documentType: string,
  documentNumber: string,
  recipientName: string,
  additionalMessage?: string
): string {
  const date = new Date().toLocaleDateString();
  const timeString = new Date().toLocaleTimeString();

  let typeLabel = "";
  switch (documentType) {
    case "REQUISITION":
      typeLabel = "Purchase Requisition";
      break;
    case "PURCHASE_ORDER":
      typeLabel = "Purchase Order";
      break;
    case "PAYMENT_VOUCHER":
      typeLabel = "Payment Voucher";
      break;
    default:
      typeLabel = "Document";
  }

  return `
Dear ${recipientName},

Please find attached the ${typeLabel} (${documentNumber}) generated on ${date} at ${timeString}.

${additionalMessage || ""}

This document has been digitally signed with a QR code for verification. You can scan the QR code to verify the document's authenticity and tracking information.

If you have any questions or concerns about this document, please do not hesitate to contact us.

Best regards,
Liyali Finance & Procurement System
`.trim();
}
