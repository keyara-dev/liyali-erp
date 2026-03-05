import React from "react";
import { Document, Page, Text, View, Image } from "@react-pdf/renderer";
import { PurchaseOrder } from "@/types/purchase-order";
import { pdfStyles } from "./pdf-styles";
import { generateDocumentQRData } from "./qr-utils";
import { PDFHeader, PDFFooter, DocumentHeader } from "./requisition-pdf";
import { capitalize } from "../utils";
import { ApprovalSignaturesSection } from "./components/approval-signatures-section";

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
  organizationLogoUrl,
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
                        purchaseOrder.priority === "urgent"
                          ? "#fee2e2"
                          : purchaseOrder.priority === "high"
                            ? "#fed7aa"
                            : "#dbeafe",
                      color:
                        purchaseOrder.priority === "urgent"
                          ? "#991b1b"
                          : purchaseOrder.priority === "high"
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
            marginBottom: 20,
            borderWidth: 1,
            borderColor: "#1e40af",
            padding: 10,
          }}
        >
          <Text
            style={{
              fontSize: 11,
              fontWeight: "bold",
              backgroundColor: "#dbeafe",
              padding: 5,
              marginBottom: 10,
            }}
          >
            SECTION 1: PURCHASE ORDER DETAILS
          </Text>

          {/* Vendor Info */}
          <View
            style={{
              marginBottom: 12,
              display: "flex",
              flexDirection: "row",
              gap: 20,
            }}
          >
            <View style={{ flex: 2 }}>
              <Text
                style={{
                  fontSize: 8,
                  fontWeight: "bold",
                  marginBottom: 2,
                  color: "#666",
                }}
              >
                VENDOR/SUPPLIER
              </Text>
              <Text style={{ fontSize: 10, fontWeight: "bold" }}>
                {purchaseOrder.vendorName || "—"}
              </Text>
            </View>
            <View style={{ flex: 1 }}>
              <Text
                style={{
                  fontSize: 8,
                  fontWeight: "bold",
                  marginBottom: 2,
                  color: "#666",
                }}
              >
                DEPARTMENT
              </Text>
              <Text style={{ fontSize: 10 }}>
                {purchaseOrder.department || "—"}
              </Text>
            </View>
            <View style={{ flex: 1 }}>
              <Text
                style={{
                  fontSize: 8,
                  fontWeight: "bold",
                  marginBottom: 2,
                  color: "#666",
                }}
              >
                BUDGET CODE
              </Text>
              <Text style={{ fontSize: 9 }}>
                {purchaseOrder.budgetCode || "—"}
              </Text>
            </View>
          </View>

          {/* Order Details Row */}
          <View style={{ display: "flex", flexDirection: "row", gap: 20 }}>
            <View style={{ flex: 1 }}>
              <Text
                style={{
                  fontSize: 8,
                  fontWeight: "bold",
                  marginBottom: 2,
                  color: "#666",
                }}
              >
                ORDER DATE
              </Text>
              <Text style={{ fontSize: 10 }}>
                {new Date(purchaseOrder.createdAt).toLocaleDateString()}
              </Text>
            </View>
            <View style={{ flex: 1 }}>
              <Text
                style={{
                  fontSize: 8,
                  fontWeight: "bold",
                  marginBottom: 2,
                  color: "#666",
                }}
              >
                REQUIRED BY DATE
              </Text>
              <Text style={{ fontSize: 10 }}>
                {purchaseOrder.requiredByDate
                  ? new Date(purchaseOrder.requiredByDate).toLocaleDateString()
                  : "—"}
              </Text>
            </View>
            {purchaseOrder.linkedRequisition && (
              <View style={{ flex: 1 }}>
                <Text
                  style={{
                    fontSize: 8,
                    fontWeight: "bold",
                    marginBottom: 2,
                    color: "#666",
                  }}
                >
                  SOURCE REQUISITION
                </Text>
                <Text style={{ fontSize: 10 }}>
                  {purchaseOrder.linkedRequisition}
                </Text>
              </View>
            )}
          </View>
        </View>

        {/* Line Items Table */}
        {purchaseOrder.items && purchaseOrder.items.length > 0 && (
          <View style={{ marginBottom: 20 }}>
            <Text style={{ fontSize: 10, fontWeight: "bold", marginBottom: 8 }}>
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
                    padding: 5,
                    fontSize: 8,
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
                    padding: 5,
                    fontSize: 8,
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
                    padding: 5,
                    fontSize: 8,
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
                    padding: 5,
                    fontSize: 8,
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
                    padding: 5,
                    fontSize: 8,
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
                        padding: 5,
                        fontSize: 8,
                        color: "#1f2937",
                        textAlign: "center",
                      }}
                    >
                      {index + 1}
                    </Text>
                    <Text
                      style={{
                        flex: 2,
                        padding: 5,
                        fontSize: 8,
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
                        padding: 5,
                        fontSize: 8,
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
                        padding: 5,
                        fontSize: 8,
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
                        padding: 5,
                        fontSize: 8,
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
                marginTop: 15,
                paddingTop: 10,
              }}
            >
              <View style={{ width: "35%" }}>
                <View
                  style={{
                    display: "flex",
                    flexDirection: "row",
                    justifyContent: "space-between",
                    paddingBottom: 5,
                    borderBottomWidth: 2,
                    borderBottomColor: "#1e40af",
                  }}
                >
                  <Text
                    style={{
                      fontSize: 9,
                      fontWeight: "bold",
                      color: "#1f2937",
                    }}
                  >
                    TOTAL ORDER VALUE:
                  </Text>
                  <Text
                    style={{
                      fontSize: 11,
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
        {purchaseOrder.approvalChain &&
          purchaseOrder.approvalChain.length > 0 && (
            <View
              style={{
                marginBottom: 20,
                borderWidth: 1,
                borderColor: "#1e40af",
                padding: 10,
              }}
            >
              <Text
                style={{
                  fontSize: 11,
                  fontWeight: "bold",
                  backgroundColor: "#dbeafe",
                  padding: 5,
                  marginBottom: 10,
                }}
              >
                APPROVAL CHAIN
              </Text>

              <View
                style={{
                  display: "flex",
                  flexDirection: "row",
                  gap: 10,
                  flexWrap: "wrap",
                }}
              >
                {purchaseOrder.approvalChain.map(
                  (stage: any, index: number) => (
                    <View
                      key={index}
                      style={{
                        flex: index % 2 === 0 ? 1 : 1,
                        minWidth: "45%",
                        borderWidth: 1,
                        borderColor: "#ddd",
                        padding: 8,
                        marginBottom: 8,
                      }}
                    >
                      <Text
                        style={{
                          fontSize: 8,
                          fontWeight: "bold",
                          marginBottom: 3,
                          color: "#1e40af",
                        }}
                      >
                        {stage.stageName || `Stage ${stage.stageNumber}`}
                      </Text>
                      <Text style={{ fontSize: 8, marginBottom: 2 }}>
                        Assigned to: {stage.assignedTo}
                      </Text>
                      <Text style={{ fontSize: 8, marginBottom: 4 }}>
                        Status: {stage.status}
                      </Text>
                      {stage.actionTakenAt && (
                        <Text style={{ fontSize: 7, color: "#666" }}>
                          Approved:{" "}
                          {new Date(stage.actionTakenAt).toLocaleDateString()}
                        </Text>
                      )}
                      {stage.signature && (
                        <Text
                          style={{
                            fontSize: 7,
                            fontStyle: "italic",
                            color: "#999",
                            marginTop: 3,
                          }}
                        >
                          Signature: {stage.signature}
                        </Text>
                      )}
                    </View>
                  ),
                )}
              </View>
            </View>
          )}

        {/* QR Code and Tracking Information */}
        <View
          style={{
            marginTop: 20,
            paddingTop: 10,
            borderTopWidth: 1,
            borderTopColor: "#ddd",
            display: "flex",
            flexDirection: "row",
            gap: 15,
            alignItems: "flex-start",
          }}
        >
          {/* QR Code Section */}
          {qrCodeUrl && (
            <View style={{ width: 80, height: 80 }}>
              <Image source={qrCodeUrl} style={{ width: 80, height: 80 }} />
            </View>
          )}

          {/* Tracking Information */}
          <View style={{ flex: 1 }}>
            <Text style={{ fontSize: 8, fontWeight: "bold", marginBottom: 4 }}>
              DOCUMENT TRACKING
            </Text>
            <Text style={{ fontSize: 7, marginBottom: 2 }}>
              Tracking Code: {documentNumber}
            </Text>
            <Text style={{ fontSize: 7, marginBottom: 2 }}>
              Document ID: {purchaseOrder.id}
            </Text>
            <Text style={{ fontSize: 7, marginBottom: 2 }}>
              Status: {capitalize(purchaseOrder.status)}
            </Text>
            <Text style={{ fontSize: 7, marginBottom: 2 }}>
              Created: {new Date(purchaseOrder.createdAt).toLocaleDateString()}{" "}
              {new Date(purchaseOrder.createdAt).toLocaleTimeString()}
            </Text>
            <Text style={{ fontSize: 7 }}>
              Generated: {new Date().toLocaleDateString()}{" "}
              {new Date().toLocaleTimeString()}
            </Text>
          </View>
        </View>

        {/* Approval Signatures Section */}
        {purchaseOrder.approvalHistory &&
          purchaseOrder.approvalHistory.length > 0 && (
            <ApprovalSignaturesSection
              approvalHistory={purchaseOrder.approvalHistory}
              documentType="Purchase Order"
            />
          )}

        {/* Footer */}
        <PDFFooter />
      </Page>
    </Document>
  );
};

export default PurchaseOrderPDF;
