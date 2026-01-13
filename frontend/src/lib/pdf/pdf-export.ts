import React from "react";
import { pdf } from "@react-pdf/renderer";
import RequisitionPDF from "./requisition-pdf";
import PurchaseOrderPDF from "./purchase-order-pdf";
import PaymentVoucherPDF from "./payment-voucher-pdf";
import { Requisition } from "@/types/requisition";
import { PurchaseOrder } from "@/types/purchase-order";
import { PaymentVoucher } from "@/types/payment-voucher";

/**
 * Export a Requisition as PDF
 * @param requisition The requisition to export
 * @returns Promise with blob
 */
export async function exportRequisitionPDF(
  requisition: Requisition
): Promise<Blob> {
  const fileName = `REQ-${requisition.documentNumber}-${new Date().getTime()}.pdf`;
  const doc = React.createElement(RequisitionPDF, { requisition });
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
  purchaseOrder: PurchaseOrder
): Promise<Blob> {
  const fileName = `PO-${purchaseOrder.documentNumber}-${new Date().getTime()}.pdf`;
  const doc = React.createElement(PurchaseOrderPDF, { purchaseOrder });
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
  paymentVoucher: PaymentVoucher
): Promise<Blob> {
  const fileName = `PV-${paymentVoucher.documentNumber}-${new Date().getTime()}.pdf`;
  const doc = React.createElement(PaymentVoucherPDF, { paymentVoucher });
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
  requisition: Requisition
): Promise<Blob> {
  const doc = React.createElement(RequisitionPDF, { requisition });
  return pdf(doc as any).toBlob();
}

/**
 * Get PDF as blob without downloading
 * @param purchaseOrder The purchase order to export
 * @returns Promise with blob
 */
export async function getPurchaseOrderPDFBlob(
  purchaseOrder: PurchaseOrder
): Promise<Blob> {
  const doc = React.createElement(PurchaseOrderPDF, { purchaseOrder });
  return pdf(doc as any).toBlob();
}

/**
 * Get PDF as blob without downloading
 * @param paymentVoucher The payment voucher to export
 * @returns Promise with blob
 */
export async function getPaymentVoucherPDFBlob(
  paymentVoucher: PaymentVoucher
): Promise<Blob> {
  const doc = React.createElement(PaymentVoucherPDF, { paymentVoucher });
  return pdf(doc as any).toBlob();
}

/**
 * Get PDF as data URL for preview
 * @param requisition The requisition to export
 * @returns Promise with data URL
 */
export async function getRequisitionPDFUrl(
  requisition: Requisition
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
  purchaseOrder: PurchaseOrder
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
  paymentVoucher: PaymentVoucher
): Promise<string> {
  const blob = await getPaymentVoucherPDFBlob(paymentVoucher);
  return URL.createObjectURL(blob);
}
