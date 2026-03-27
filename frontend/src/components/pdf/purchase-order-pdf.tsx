import React from "react";
import { Document, Page, Text, View } from "@react-pdf/renderer";
import { PurchaseOrder } from "@/types/purchase-order";
import { pdfStyles } from "@/lib/pdf/pdf-styles";
import { generateDocumentQRData } from "@/lib/pdf/qr-utils";
import { PDFHeader, PDFFooter, DocumentHeader } from "./requisition-pdf";
import { capitalize } from "@/lib/utils";
import { ApprovalSignaturesSection } from "./approval-signatures-section";

interface PurchaseOrderPDFProps {
  purchaseOrder: PurchaseOrder;
  qrCodeUrl?: string;
  organizationLogoUrl?: string;
  documentHeader?: DocumentHeader;
}

const getStatusColor = (status: string) => {
  switch (status) {
    case "DRAFT":
      return pdfStyles.statusDraft;
    case "SUBMITTED":
      return pdfStyles.statusSubmitted;
    case "IN_REVIEW":
      return pdfStyles.statusInReview;
    case "APPROVED":
      return pdfStyles.statusApproved;
    case "REVISION":
      return pdfStyles.statusInReview;
    case "REJECTED":
      return pdfStyles.statusRejected;
    default:
      return pdfStyles.statusDraft;
  }
};

const PurchaseOrderPDF: React.FC<PurchaseOrderPDFProps> = ({
  purchaseOrder,
  qrCodeUrl,
  documentHeader,
}) => {
  const documentNumber = purchaseOrder.documentNumber;
  const qrData = generateDocumentQRData(
    "PURCHASE_ORDER",
    documentNumber,
    purchaseOrder.id,
    new Date(purchaseOrder.createdAt),
  );

  return (
    <Document>
      <Page size="A4" style={pdfStyles.page}>
        {/* Header with Republic of Zambia and Logo */}
        <PDFHeader
          title="PURCHASE ORDER"
          logoUrl={documentHeader?.logoUrl}
          orgName={documentHeader?.orgName}
          tagline={documentHeader?.tagline}
        />

        {/* Main Header Section */}
        <View
          style={[
            pdfStyles.header,
            {
              marginBottom: 16,
              paddingBottom: 10,
              flexDirection: "row",
              justifyContent: "space-between",
            },
          ]}
        >
          <View style={{ textAlign: "left" }}>
            <Text style={{ fontSize: 10, fontWeight: "bold", marginBottom: 2 }}>
              Document No: {documentNumber}
            </Text>
            <Text style={{ fontSize: 8, color: "#666", marginBottom: 3 }}>
              Date: {new Date(purchaseOrder.createdAt).toLocaleDateString()}
            </Text>
          </View>

          {/* STATUS AND PRIORITY BADGES */}
          <View style={{ textAlign: "right" }}>
            <Text style={{ fontSize: 7, fontWeight: "bold", marginBottom: 1 }}>
              STATUS & PRIORITY
            </Text>
            <View style={{ marginBottom: 0, flexDirection: "row", gap: 4 }}>
              <View
                style={[
                  pdfStyles.statusBadge,
                  getStatusColor(purchaseOrder.status),
                ]}
              >
                <Text style={{ fontSize: 9 }}>
                  {capitalize(purchaseOrder.status)}
                </Text>
              </View>
              {purchaseOrder.priority && (
                <View
                  style={[
                    pdfStyles.statusBadge,
                    {
                      backgroundColor:
                        purchaseOrder.priority?.toUpperCase() === "URGENT"
                          ? "#fee2e2"
                          : purchaseOrder.priority?.toUpperCase() === "HIGH"
                            ? "#fed7aa"
                            : "#dbeafe",
                      color:
                        purchaseOrder.priority?.toUpperCase() === "URGENT"
                          ? "#991b1b"
                          : purchaseOrder.priority?.toUpperCase() === "HIGH"
                            ? "#92400e"
                            : "#1e40af",
                    },
                  ]}
                >
                  <Text style={{ fontSize: 9 }}>
                    {capitalize(purchaseOrder.priority)}
                  </Text>
                </View>
              )}
            </View>
          </View>
        </View>

        {/* SECTION 1: PURCHASE ORDER DETAILS */}
        <View
          style={{
            marginBottom: 12,
            borderWidth: 1,
            borderColor: "#1e40af",
            padding: 7,
          }}
        >
          <Text
            style={{
              fontSize: 9,
              fontWeight: "bold",
              backgroundColor: "#dbeafe",
              padding: 3,
              marginBottom: 6,
            }}
          >
            SECTION 1: PURCHASE ORDER DETAILS
          </Text>

          {/* Vendor Info */}
          <View
            style={{
              marginBottom: 7,
              display: "flex",
              flexDirection: "row",
              gap: 12,
            }}
          >
            <View style={{ flex: 2 }}>
              <Text
                style={{
                  fontSize: 7,
                  fontWeight: "bold",
                  marginBottom: 1,
                  color: "#666",
                }}
              >
                VENDOR/SUPPLIER
              </Text>
              <Text style={{ fontSize: 9, fontWeight: "bold" }}>
                {purchaseOrder.vendorName || "—"}
              </Text>
            </View>
            <View style={{ flex: 1 }}>
              <Text
                style={{
                  fontSize: 7,
                  fontWeight: "bold",
                  marginBottom: 1,
                  color: "#666",
                }}
              >
                DEPARTMENT
              </Text>
              <Text style={{ fontSize: 9 }}>
                {purchaseOrder.department || "—"}
              </Text>
            </View>
            <View style={{ flex: 1 }}>
              <Text
                style={{
                  fontSize: 7,
                  fontWeight: "bold",
                  marginBottom: 1,
                  color: "#666",
                }}
              >
                BUDGET CODE
              </Text>
              <Text style={{ fontSize: 8 }}>
                {purchaseOrder.budgetCode || "—"}
              </Text>
            </View>
          </View>

          {/* Order Details Row */}
          <View style={{ display: "flex", flexDirection: "row", gap: 12 }}>
            <View style={{ flex: 1 }}>
              <Text
                style={{
                  fontSize: 7,
                  fontWeight: "bold",
                  marginBottom: 1,
                  color: "#666",
                }}
              >
                ORDER DATE
              </Text>
              <Text style={{ fontSize: 9 }}>
                {new Date(purchaseOrder.createdAt).toLocaleDateString()}
              </Text>
            </View>
            <View style={{ flex: 1 }}>
              <Text
                style={{
                  fontSize: 7,
                  fontWeight: "bold",
                  marginBottom: 1,
                  color: "#666",
                }}
              >
                REQUIRED BY DATE
              </Text>
              <Text style={{ fontSize: 9 }}>
                {purchaseOrder.requiredByDate
                  ? new Date(purchaseOrder.requiredByDate).toLocaleDateString()
                  : "—"}
              </Text>
            </View>
            {purchaseOrder.linkedRequisition && (
              <View style={{ flex: 1 }}>
                <Text
                  style={{
                    fontSize: 7,
                    fontWeight: "bold",
                    marginBottom: 1,
                    color: "#666",
                  }}
                >
                  SOURCE REQUISITION
                </Text>
                <Text style={{ fontSize: 9 }}>
                  {purchaseOrder.linkedRequisition}
                </Text>
              </View>
            )}
          </View>
        </View>

        {/* Line Items Table */}
        {purchaseOrder.items && purchaseOrder.items.length > 0 && (
          <View style={{ marginBottom: 10 }}>
            <Text style={{ fontSize: 9, fontWeight: "bold", marginBottom: 4 }}>
              ORDER ITEMS:
            </Text>

            {/* Table Header */}
            <View
              style={{
                borderWidth: 1,
                borderColor: "#1e40af",
                marginBottom: 0,
              }}
            >
              <View
                style={{
                  display: "flex",
                  flexDirection: "row",
                  backgroundColor: "#f3f4f6",
                  borderBottomWidth: 1,
                  borderBottomColor: "#1e40af",
                }}
              >
                <Text
                  style={{
                    flex: 0.5,
                    paddingVertical: 3,
                    paddingHorizontal: 4,
                    fontSize: 7.5,
                    fontWeight: "bold",
                    color: "#1e40af",
                    textAlign: "center",
                  }}
                >
                  Item
                </Text>
                <Text
                  style={{
                    flex: 2,
                    paddingVertical: 3,
                    paddingHorizontal: 4,
                    fontSize: 7.5,
                    fontWeight: "bold",
                    color: "#1e40af",
                    borderLeftWidth: 1,
                    borderLeftColor: "#1e40af",
                  }}
                >
                  Description
                </Text>
                <Text
                  style={{
                    flex: 1,
                    paddingVertical: 3,
                    paddingHorizontal: 4,
                    fontSize: 7.5,
                    fontWeight: "bold",
                    color: "#1e40af",
                    textAlign: "center",
                    borderLeftWidth: 1,
                    borderLeftColor: "#1e40af",
                  }}
                >
                  Qty
                </Text>
                <Text
                  style={{
                    flex: 1,
                    paddingVertical: 3,
                    paddingHorizontal: 4,
                    fontSize: 7.5,
                    fontWeight: "bold",
                    color: "#1e40af",
                    textAlign: "right",
                    borderLeftWidth: 1,
                    borderLeftColor: "#1e40af",
                  }}
                >
                  Unit Price
                </Text>
                <Text
                  style={{
                    flex: 1,
                    paddingVertical: 3,
                    paddingHorizontal: 4,
                    fontSize: 7.5,
                    fontWeight: "bold",
                    color: "#1e40af",
                    textAlign: "right",
                    borderLeftWidth: 1,
                    borderLeftColor: "#1e40af",
                  }}
                >
                  Total
                </Text>
              </View>

              {/* Table Rows */}
              {purchaseOrder.items.map((item: any, index: number) => {
                const itemDescription =
                  item.description || item.itemDescription || "";
                const unitPrice = item.unitPrice || item.estimatedCost || 0;
                const totalPrice =
                  item.totalPrice || item.quantity * unitPrice || 0;

                return (
                  <View
                    key={item.id}
                    style={{
                      display: "flex",
                      flexDirection: "row",
                      borderBottomWidth: 1,
                      borderBottomColor: "#e5e7eb",
                    }}
                  >
                    <Text
                      style={{
                        flex: 0.5,
                        paddingVertical: 2,
                        paddingHorizontal: 4,
                        fontSize: 7.5,
                        color: "#1f2937",
                        textAlign: "center",
                      }}
                    >
                      {index + 1}
                    </Text>
                    <Text
                      style={{
                        flex: 2,
                        paddingVertical: 2,
                        paddingHorizontal: 4,
                        fontSize: 7.5,
                        color: "#1f2937",
                        borderLeftWidth: 1,
                        borderLeftColor: "#e5e7eb",
                      }}
                    >
                      {itemDescription}
                    </Text>
                    <Text
                      style={{
                        flex: 1,
                        paddingVertical: 2,
                        paddingHorizontal: 4,
                        fontSize: 7.5,
                        color: "#1f2937",
                        textAlign: "center",
                        borderLeftWidth: 1,
                        borderLeftColor: "#e5e7eb",
                      }}
                    >
                      {item.quantity} {item.unit || ""}
                    </Text>
                    <Text
                      style={{
                        flex: 1,
                        paddingVertical: 2,
                        paddingHorizontal: 4,
                        fontSize: 7.5,
                        color: "#1f2937",
                        textAlign: "right",
                        borderLeftWidth: 1,
                        borderLeftColor: "#e5e7eb",
                      }}
                    >
                      {purchaseOrder.currency}{" "}
                      {unitPrice?.toLocaleString() || "0"}
                    </Text>
                    <Text
                      style={{
                        flex: 1,
                        paddingVertical: 2,
                        paddingHorizontal: 4,
                        fontSize: 7.5,
                        color: "#1f2937",
                        textAlign: "right",
                        borderLeftWidth: 1,
                        borderLeftColor: "#e5e7eb",
                      }}
                    >
                      {purchaseOrder.currency}{" "}
                      {totalPrice?.toLocaleString() || "0"}
                    </Text>
                  </View>
                );
              })}
            </View>

            {/* Totals */}
            <View
              style={{
                display: "flex",
                flexDirection: "row",
                justifyContent: "flex-end",
                marginTop: 6,
                paddingTop: 4,
              }}
            >
              <View style={{ width: "35%" }}>
                <View
                  style={{
                    display: "flex",
                    flexDirection: "row",
                    justifyContent: "space-between",
                    paddingBottom: 4,
                    borderBottomWidth: 2,
                    borderBottomColor: "#1e40af",
                  }}
                >
                  <Text
                    style={{
                      fontSize: 8,
                      fontWeight: "bold",
                      color: "#1f2937",
                    }}
                  >
                    TOTAL ORDER VALUE:
                  </Text>
                  <Text
                    style={{
                      fontSize: 10,
                      fontWeight: "bold",
                      color: "#166534",
                    }}
                  >
                    {purchaseOrder.currency}{" "}
                    {purchaseOrder.totalAmount?.toLocaleString() || "0"}
                  </Text>
                </View>
              </View>
            </View>
          </View>
        )}

        {/* APPROVAL CHAIN */}
        {purchaseOrder.approvalHistory &&
          purchaseOrder.approvalHistory.length > 0 && (
            <ApprovalSignaturesSection
              approvalHistory={purchaseOrder.approvalHistory}
              documentType="Purchase Order"
            />
          )}

        {/* Footer: QR Code, Tracking, and Branding */}
        <PDFFooter
          documentNumber={documentNumber}
          document={purchaseOrder}
          qrCodeUrl={qrCodeUrl ?? ""}
        />
      </Page>
    </Document>
  );
};

export default PurchaseOrderPDF;
