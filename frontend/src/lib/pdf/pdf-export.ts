import React from "react";
import { pdf } from "@react-pdf/renderer";
import RequisitionPDF from "./requisition-pdf";
import PurchaseOrderPDF from "./purchase-order-pdf";
import PaymentVoucherPDF from "./payment-voucher-pdf";
import { GoodsReceivedNotePDF } from "@/lib/pdf-generators/grn-pdf";
import { Requisition } from "@/types/requisition";
import { PurchaseOrder } from "@/types/purchase-order";
import { PaymentVoucher } from "@/types/payment-voucher";
import { GoodsReceivedNote } from "@/types/goods-received-note";
import { getDocumentQRCodeUrl } from "./qr-utils";
import { getOrganizationById } from "@/app/_actions/organizations";

/**
 * Get organization logo URL
 * @param organizationId The organization ID
 * @returns Promise with logo URL or undefined
 */
async function getOrganizationLogoUrl(
  organizationId: string,
): Promise<string | undefined> {
  try {
    const response = await getOrganizationById(organizationId);
    if (response.success && response.data?.logoUrl) {
      return response.data.logoUrl;
    }
  } catch (error) {
    console.error("Failed to fetch organization logo:", error);
  }
  return undefined;
}

/**
 * Export a Requisition as PDF
 * @param requisition The requisition to export
 * @returns Promise with blob
 */
export async function exportRequisitionPDF(
  requisition: Requisition,
): Promise<Blob> {
  const fileName = `${requisition.documentNumber}.pdf`;
  const qrCodeUrl = getDocumentQRCodeUrl(
    requisition.documentNumber,
    200,
    requisition.organizationId,
  );
  const organizationLogoUrl = await getOrganizationLogoUrl(
    requisition.organizationId,
  );
  const doc = React.createElement(RequisitionPDF, {
    requisition,
    qrCodeUrl,
    organizationLogoUrl,
  });
  const blob = await pdf(doc as any).toBlob();

  // Trigger download
  downloadBlob(blob, fileName);

  return blob;
}

/**
 * Export a Purchase Order as PDF
 * @param purchaseOrder The purchase order to export
 * @returns Promise with blob
 */
export async function exportPurchaseOrderPDF(
  purchaseOrder: PurchaseOrder,
): Promise<Blob> {
  const fileName = `${purchaseOrder.documentNumber}.pdf`;
  const qrCodeUrl = getDocumentQRCodeUrl(
    purchaseOrder.documentNumber,
    200,
    purchaseOrder.organizationId,
  );
  const organizationLogoUrl = await getOrganizationLogoUrl(
    purchaseOrder.organizationId,
  );
  const doc = React.createElement(PurchaseOrderPDF, {
    purchaseOrder,
    qrCodeUrl,
    organizationLogoUrl,
  });
  const blob = await pdf(doc as any).toBlob();

  // Trigger download
  downloadBlob(blob, fileName);

  return blob;
}

/**
 * Export a Payment Voucher as PDF
 * @param paymentVoucher The payment voucher to export
 * @returns Promise with blob
 */
export async function exportPaymentVoucherPDF(
  paymentVoucher: PaymentVoucher,
): Promise<Blob> {
  const fileName = `${paymentVoucher.documentNumber}.pdf`;
  const qrCodeUrl = getDocumentQRCodeUrl(
    paymentVoucher.documentNumber,
    200,
    paymentVoucher.organizationId,
  );
  const organizationLogoUrl = await getOrganizationLogoUrl(
    paymentVoucher.organizationId,
  );
  const doc = React.createElement(PaymentVoucherPDF, {
    paymentVoucher,
    qrCodeUrl,
    organizationLogoUrl,
  });
  const blob = await pdf(doc as any).toBlob();

  // Trigger download
  downloadBlob(blob, fileName);

  return blob;
}

/**
 * Download a blob as a file
 * @param blob The blob to download
 * @param fileName The file name
 */
export function downloadBlob(blob: Blob, fileName: string): void {
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = fileName;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

/**
 * Get PDF as blob without downloading
 * @param requisition The requisition to export
 * @returns Promise with blob
 */
export async function getRequisitionPDFBlob(
  requisition: Requisition,
): Promise<Blob> {
  const qrCodeUrl = getDocumentQRCodeUrl(
    requisition.documentNumber,
    200,
    requisition.organizationId,
  );
  const organizationLogoUrl = await getOrganizationLogoUrl(
    requisition.organizationId,
  );
  const doc = React.createElement(RequisitionPDF, {
    requisition,
    qrCodeUrl,
    organizationLogoUrl,
  });
  return pdf(doc as any).toBlob();
}

/**
 * Get PDF as blob without downloading
 * @param purchaseOrder The purchase order to export
 * @returns Promise with blob
 */
export async function getPurchaseOrderPDFBlob(
  purchaseOrder: PurchaseOrder,
): Promise<Blob> {
  const qrCodeUrl = getDocumentQRCodeUrl(
    purchaseOrder.documentNumber,
    200,
    purchaseOrder.organizationId,
  );
  const organizationLogoUrl = await getOrganizationLogoUrl(
    purchaseOrder.organizationId,
  );
  const doc = React.createElement(PurchaseOrderPDF, {
    purchaseOrder,
    qrCodeUrl,
    organizationLogoUrl,
  });
  return pdf(doc as any).toBlob();
}

/**
 * Get PDF as blob without downloading
 * @param paymentVoucher The payment voucher to export
 * @returns Promise with blob
 */
export async function getPaymentVoucherPDFBlob(
  paymentVoucher: PaymentVoucher,
): Promise<Blob> {
  const qrCodeUrl = getDocumentQRCodeUrl(
    paymentVoucher.documentNumber,
    200,
    paymentVoucher.organizationId,
  );
  const organizationLogoUrl = await getOrganizationLogoUrl(
    paymentVoucher.organizationId,
  );
  const doc = React.createElement(PaymentVoucherPDF, {
    paymentVoucher,
    qrCodeUrl,
    organizationLogoUrl,
  });
  return pdf(doc as any).toBlob();
}

/**
 * Get PDF as data URL for preview
 * @param requisition The requisition to export
 * @returns Promise with data URL
 */
export async function getRequisitionPDFUrl(
  requisition: Requisition,
): Promise<string> {
  const blob = await getRequisitionPDFBlob(requisition);
  return URL.createObjectURL(blob);
}

/**
 * Get PDF as data URL for preview
 * @param purchaseOrder The purchase order to export
 * @returns Promise with data URL
 */
export async function getPurchaseOrderPDFUrl(
  purchaseOrder: PurchaseOrder,
): Promise<string> {
  const blob = await getPurchaseOrderPDFBlob(purchaseOrder);
  return URL.createObjectURL(blob);
}

/**
 * Get PDF as data URL for preview
 * @param paymentVoucher The payment voucher to export
 * @returns Promise with data URL
 */
export async function getPaymentVoucherPDFUrl(
  paymentVoucher: PaymentVoucher,
): Promise<string> {
  const blob = await getPaymentVoucherPDFBlob(paymentVoucher);
  return URL.createObjectURL(blob);
}

/**
 * Export a GRN as PDF
 * @param grn The goods received note to export
 * @returns Promise with blob
 */
export async function exportGrnPDF(grn: GoodsReceivedNote): Promise<Blob> {
  const fileName = `${grn.documentNumber}.pdf`;
  const qrCodeUrl = getDocumentQRCodeUrl(
    grn.documentNumber,
    200,
    grn.organizationId,
  );
  const organizationLogoUrl = await getOrganizationLogoUrl(grn.organizationId);
  const doc = React.createElement(GoodsReceivedNotePDF, {
    grn,
    qrCodeUrl,
    organizationLogoUrl,
  });
  const blob = await pdf(doc as any).toBlob();

  // Trigger download
  downloadBlob(blob, fileName);

  return blob;
}

/**
 * Get GRN PDF as blob without downloading
 * @param grn The goods received note to export
 * @returns Promise with blob
 */
export async function getGrnPDFBlob(grn: GoodsReceivedNote): Promise<Blob> {
  const qrCodeUrl = getDocumentQRCodeUrl(
    grn.documentNumber,
    200,
    grn.organizationId,
  );
  const organizationLogoUrl = await getOrganizationLogoUrl(grn.organizationId);
  const doc = React.createElement(GoodsReceivedNotePDF, {
    grn,
    qrCodeUrl,
    organizationLogoUrl,
  });
  return pdf(doc as any).toBlob();
}

/**
 * Get GRN PDF as data URL for preview
 * @param grn The goods received note to export
 * @returns Promise with data URL
 */
export async function getGrnPDFUrl(grn: GoodsReceivedNote): Promise<string> {
  const blob = await getGrnPDFBlob(grn);
  return URL.createObjectURL(blob);
}
