import React from "react";
import { Document, Page, Text, View, Image } from "@react-pdf/renderer";
import {
  GoodsReceivedNote,
  GRNItem,
  QualityIssue,
} from "@/types/goods-received-note";
import { pdfStyles } from "../pdf/pdf-styles";
import { generateDocumentQRData } from "../pdf/qr-utils";
import { PDFHeader, PDFFooter, DocumentHeader } from "../pdf/requisition-pdf";
import { ApprovalSignaturesSection } from "../pdf/components/approval-signatures-section";
import { capitalize } from "../utils";

interface GRNPDFProps {
  grn: GoodsReceivedNote;
  qrCodeUrl?: string;
  organizationLogoUrl?: string;
  documentHeader?: DocumentHeader;
}

const getStatusColor = (status: string) => {
  switch (status?.toUpperCase()) {
    case "DRAFT":
      return pdfStyles.statusDraft;
    case "SUBMITTED":
      return pdfStyles.statusSubmitted;
    case "IN_REVIEW":
      return pdfStyles.statusInReview;
    case "APPROVED":
      return pdfStyles.statusApproved;
    case "REJECTED":
      return pdfStyles.statusRejected;
    case "COMPLETED":
      return {
        ...pdfStyles.statusApproved,
        backgroundColor: "#d1fae5",
        color: "#065f46",
      };
    default:
      return pdfStyles.statusDraft;
  }
};

const getConditionColor = (condition: string) => {
  switch (condition?.toLowerCase()) {
    case "good":
      return { backgroundColor: "#dcfce7", color: "#166534" };
    case "damaged":
      return { backgroundColor: "#fee2e2", color: "#991b1b" };
    case "missing":
      return { backgroundColor: "#fef3c7", color: "#92400e" };
    default:
      return { backgroundColor: "#f3f4f6", color: "#374151" };
  }
};

const getSeverityColor = (severity: string) => {
  switch (severity?.toLowerCase()) {
    case "low":
      return { backgroundColor: "#dbeafe", color: "#1e40af" };
    case "medium":
      return { backgroundColor: "#fef3c7", color: "#92400e" };
    case "high":
      return { backgroundColor: "#fee2e2", color: "#991b1b" };
    default:
      return { backgroundColor: "#f3f4f6", color: "#374151" };
  }
};

const GoodsReceivedNotePDF: React.FC<GRNPDFProps> = ({
  grn,
  qrCodeUrl,
  organizationLogoUrl,
  documentHeader,
}) => {
  const documentNumber = grn.documentNumber;
  const qrData = generateDocumentQRData(
    "GRN",
    documentNumber,
    grn.id,
    new Date(grn.createdAt),
  );

  return (
    <Document>
      <Page size="A4" style={pdfStyles.page}>
        {/* Header with Republic of Zambia and Logo */}
        <PDFHeader
          title="GOODS RECEIVED NOTE"
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
              Date: {new Date(grn.receivedDate).toLocaleDateString()}
            </Text>
          </View>

          {/* STATUS BADGE */}
          <View style={{ textAlign: "right" }}>
            <Text style={{ fontSize: 7, fontWeight: "bold", marginBottom: 1 }}>
              STATUS
            </Text>
            <View style={{ marginBottom: 0, flexDirection: "row", gap: 4 }}>
              <View style={[pdfStyles.statusBadge, getStatusColor(grn.status)]}>
                <Text style={{ fontSize: 9 }}>{capitalize(grn.status)}</Text>
              </View>
            </View>
          </View>
        </View>

        {/* SECTION 1: GRN DETAILS */}
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
            SECTION 1: GOODS RECEIVED DETAILS
          </Text>

          <View
            style={{
              marginBottom: 12,
              display: "flex",
              flexDirection: "row",
              gap: 20,
            }}
          >
            <View style={{ flex: 1 }}>
              <Text
                style={{
                  fontSize: 8,
                  fontWeight: "bold",
                  marginBottom: 2,
                  color: "#666",
                }}
              >
                SOURCE PO
              </Text>
              <Text style={{ fontSize: 10 }}>
                {grn.poDocumentNumber || "—"}
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
                RECEIVED DATE
              </Text>
              <Text style={{ fontSize: 10 }}>
                {new Date(grn.receivedDate).toLocaleDateString()}
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
                WAREHOUSE LOCATION
              </Text>
              <Text style={{ fontSize: 10 }}>
                {grn.warehouseLocation || "—"}
              </Text>
            </View>
          </View>

          {/* Notes */}
          {grn.notes && (
            <View style={{ marginBottom: 12 }}>
              <Text
                style={{
                  fontSize: 8,
                  fontWeight: "bold",
                  marginBottom: 2,
                  color: "#666",
                }}
              >
                NOTES / REMARKS
              </Text>
              <Text style={{ fontSize: 9 }}>{grn.notes}</Text>
            </View>
          )}

          {/* Received By Info */}
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
                RECEIVED BY
              </Text>
              <Text style={{ fontSize: 10 }}>{grn.receivedBy || "—"}</Text>
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
                CREATED BY
              </Text>
              <Text style={{ fontSize: 10 }}>{grn.createdBy || "—"}</Text>
            </View>
          </View>
        </View>

        {/* Items Received Table */}
        {grn.items && grn.items.length > 0 && (
          <View style={{ marginBottom: 20 }}>
            <Text style={{ fontSize: 10, fontWeight: "bold", marginBottom: 8 }}>
              ITEMS RECEIVED:
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
                  #
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
                    flex: 0.8,
                    padding: 5,
                    fontSize: 8,
                    fontWeight: "bold",
                    color: "#1e40af",
                    textAlign: "center",
                    borderLeftWidth: 1,
                    borderLeftColor: "#1e40af",
                  }}
                >
                  Ordered
                </Text>
                <Text
                  style={{
                    flex: 0.8,
                    padding: 5,
                    fontSize: 8,
                    fontWeight: "bold",
                    color: "#1e40af",
                    textAlign: "center",
                    borderLeftWidth: 1,
                    borderLeftColor: "#1e40af",
                  }}
                >
                  Received
                </Text>
                <Text
                  style={{
                    flex: 0.8,
                    padding: 5,
                    fontSize: 8,
                    fontWeight: "bold",
                    color: "#1e40af",
                    textAlign: "center",
                    borderLeftWidth: 1,
                    borderLeftColor: "#1e40af",
                  }}
                >
                  Variance
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
                  Condition
                </Text>
              </View>

              {/* Table Rows */}
              {grn.items.map((item: GRNItem, index: number) => (
                <View
                  key={index}
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
                    {item.description}
                  </Text>
                  <Text
                    style={{
                      flex: 0.8,
                      padding: 5,
                      fontSize: 8,
                      color: "#1f2937",
                      textAlign: "center",
                      borderLeftWidth: 1,
                      borderLeftColor: "#e5e7eb",
                    }}
                  >
                    {item.quantityOrdered}
                  </Text>
                  <Text
                    style={{
                      flex: 0.8,
                      padding: 5,
                      fontSize: 8,
                      color: "#1f2937",
                      textAlign: "center",
                      borderLeftWidth: 1,
                      borderLeftColor: "#e5e7eb",
                    }}
                  >
                    {item.quantityReceived}
                  </Text>
                  <Text
                    style={{
                      flex: 0.8,
                      padding: 5,
                      fontSize: 8,
                      color: item.variance !== 0 ? "#991b1b" : "#166534",
                      textAlign: "center",
                      borderLeftWidth: 1,
                      borderLeftColor: "#e5e7eb",
                    }}
                  >
                    {item.variance > 0 ? `+${item.variance}` : item.variance}
                  </Text>
                  <View
                    style={{
                      flex: 1,
                      padding: 5,
                      borderLeftWidth: 1,
                      borderLeftColor: "#e5e7eb",
                      alignItems: "center",
                    }}
                  >
                    <View
                      style={[
                        {
                          paddingHorizontal: 6,
                          paddingVertical: 2,
                          borderRadius: 4,
                        },
                        getConditionColor(item.condition),
                      ]}
                    >
                      <Text style={{ fontSize: 7 }}>
                        {capitalize(item.condition)}
                      </Text>
                    </View>
                  </View>
                </View>
              ))}
            </View>
          </View>
        )}

        {/* Quality Issues Section */}
        {grn.qualityIssues && grn.qualityIssues.length > 0 && (
          <View
            style={{
              marginBottom: 20,
              borderWidth: 1,
              borderColor: "#dc2626",
              padding: 10,
            }}
          >
            <Text
              style={{
                fontSize: 11,
                fontWeight: "bold",
                backgroundColor: "#fee2e2",
                padding: 5,
                marginBottom: 10,
                color: "#991b1b",
              }}
            >
              QUALITY ISSUES REPORTED
            </Text>

            {grn.qualityIssues.map((issue: QualityIssue, index: number) => (
              <View
                key={index}
                style={{
                  marginBottom: 8,
                  padding: 8,
                  borderWidth: 1,
                  borderColor: "#fecaca",
                  backgroundColor: "#fef2f2",
                }}
              >
                <View
                  style={{
                    display: "flex",
                    flexDirection: "row",
                    justifyContent: "space-between",
                    marginBottom: 4,
                  }}
                >
                  <Text style={{ fontSize: 9, fontWeight: "bold" }}>
                    {issue.itemDescription}
                  </Text>
                  <View
                    style={[
                      {
                        paddingHorizontal: 6,
                        paddingVertical: 2,
                        borderRadius: 4,
                      },
                      getSeverityColor(issue.severity),
                    ]}
                  >
                    <Text style={{ fontSize: 7 }}>
                      {capitalize(issue.severity)}
                    </Text>
                  </View>
                </View>
                <Text style={{ fontSize: 8, color: "#666", marginBottom: 2 }}>
                  Issue Type: {capitalize(issue.issueType?.replace(/_/g, " "))}
                </Text>
                <Text style={{ fontSize: 8 }}>{issue.description}</Text>
              </View>
            ))}
          </View>
        )}

        {/* APPROVAL CHAIN */}
        {grn.approvalHistory && grn.approvalHistory.length > 0 && (
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
              {grn.approvalHistory.map((stage: any, index: number) => (
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
              ))}
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
              Document ID: {grn.id}
            </Text>
            <Text style={{ fontSize: 7, marginBottom: 2 }}>
              Status: {capitalize(grn.status)}
            </Text>
            <Text style={{ fontSize: 7, marginBottom: 2 }}>
              Created: {new Date(grn.createdAt).toLocaleDateString()}{" "}
              {new Date(grn.createdAt).toLocaleTimeString()}
            </Text>
            <Text style={{ fontSize: 7 }}>
              Generated: {new Date().toLocaleDateString()}{" "}
              {new Date().toLocaleTimeString()}
            </Text>
          </View>
        </View>

        {/* Approval Signatures Section */}
        {grn.approvalHistory && grn.approvalHistory.length > 0 && (
          <ApprovalSignaturesSection
            approvalHistory={grn.approvalHistory}
            documentType="Goods Received Note"
          />
        )}

        {/* Footer */}
        <PDFFooter />
      </Page>
    </Document>
  );
};

export { GoodsReceivedNotePDF };
export default GoodsReceivedNotePDF;
