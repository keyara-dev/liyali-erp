import React from "react";
import { Document, Page, Text, View, Image } from "@react-pdf/renderer";
import {
  GoodsReceivedNote,
  GRNItem,
  QualityIssue,
} from "@/types/goods-received-note";
import { pdfStyles } from "@/lib/pdf/pdf-styles";
import { generateDocumentQRData } from "@/lib/pdf/qr-utils";
import { PDFHeader, PDFFooter, DocumentHeader } from "./requisition-pdf";
import { ApprovalSignaturesSection } from "./approval-signatures-section";
import { capitalize } from "@/lib/utils";

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
    case "PENDING":
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

function formatDate(value?: string | Date): string {
  if (!value) return "—";
  const d = typeof value === "string" ? new Date(value) : value;
  if (Number.isNaN(d.getTime())) return "—";
  return d.toLocaleDateString();
}

const GoodsReceivedNotePDF: React.FC<GRNPDFProps> = ({
  grn,
  qrCodeUrl,
  documentHeader,
}) => {
  const documentNumber = grn.documentNumber;
  generateDocumentQRData(
    "GRN",
    documentNumber,
    grn.id,
    new Date(grn.createdAt),
  );

  // Workflow chain is present when at least one approval record was emitted.
  // Used to switch the signature footer between "form-only" (receiver +
  // certifier) and "form + workflow chain".
  const wentThroughWorkflow =
    Array.isArray(grn.approvalHistory) && grn.approvalHistory.length > 0;

  return (
    <Document>
      <Page size="A4" style={pdfStyles.page}>
        {/* Republic of Zambia + organization logo banner */}
        <PDFHeader
          title="GOODS RECEIVED NOTE"
          logoUrl={documentHeader?.logoUrl}
          orgName={documentHeader?.orgName}
          tagline={documentHeader?.tagline}
        />

        {/* Document number + status row */}
        <View
          style={[
            pdfStyles.header,
            {
              flexDirection: "row",
              justifyContent: "space-between",
              alignItems: "flex-start",
            },
          ]}
        >
          <View>
            <Text style={{ fontSize: 10, fontWeight: "bold", marginBottom: 2 }}>
              Document No: {documentNumber}
            </Text>
            <Text style={{ fontSize: 8, color: "#666" }}>
              Issued: {formatDate(grn.receivedDate)}
            </Text>
          </View>
          <View style={{ alignItems: "flex-end", gap: 3 }}>
            <View style={{ flexDirection: "row", gap: 4 }}>
              <View style={[pdfStyles.statusBadge, getStatusColor(grn.status)]}>
                <Text style={{ fontSize: 9 }}>{capitalize(grn.status)}</Text>
              </View>
            </View>
            <Text style={{ fontSize: 7, color: "#666" }}>
              {wentThroughWorkflow ? "Workflow approved" : "Direct sign-off"}
            </Text>
          </View>
        </View>

        {/* Supplier + Consignment Note table — mirrors the sample form header */}
        <View
          style={{
            marginBottom: 10,
            borderWidth: 1,
            borderColor: "#1e40af",
          }}
        >
          {/* Column headers */}
          <View
            style={{
              flexDirection: "row",
              backgroundColor: "#e5e7eb",
              borderBottomWidth: 1,
              borderBottomColor: "#1e40af",
            }}
          >
            <Text
              style={{
                flex: 2,
                padding: 5,
                fontSize: 8,
                fontWeight: "bold",
                color: "#1e40af",
              }}
            >
              Name and Address of Supplier
            </Text>
            <Text
              style={{
                flex: 1,
                padding: 5,
                fontSize: 8,
                fontWeight: "bold",
                color: "#1e40af",
                borderLeftWidth: 1,
                borderLeftColor: "#1e40af",
              }}
            >
              Delivery Consignment Note
            </Text>
          </View>
          <View style={{ flexDirection: "row" }}>
            <View style={{ flex: 2, padding: 6 }}>
              <Text style={{ fontSize: 10, fontWeight: "bold" }}>
                {grn.vendorName || "—"}
              </Text>
              {grn.vendorAddress ? (
                <Text style={{ fontSize: 8, color: "#444", marginTop: 2 }}>
                  {grn.vendorAddress}
                </Text>
              ) : null}
            </View>
            <View
              style={{
                flex: 1,
                padding: 6,
                borderLeftWidth: 1,
                borderLeftColor: "#1e40af",
              }}
            >
              <Text style={{ fontSize: 10 }}>
                {grn.consignmentNote || "—"}
              </Text>
              <Text style={{ fontSize: 7, color: "#666", marginTop: 3 }}>
                Date: {formatDate(grn.receivedDate)}
              </Text>
            </View>
          </View>
        </View>

        {/* Source PO + Warehouse + Linked PV (when applicable) */}
        <View
          style={{
            flexDirection: "row",
            gap: 20,
            marginBottom: 12,
            paddingHorizontal: 2,
          }}
        >
          <MetaField label="Source PO" value={grn.poDocumentNumber || "—"} />
          <MetaField
            label="Warehouse Location"
            value={grn.warehouseLocation || "—"}
          />
          {grn.linkedPV ? (
            <MetaField label="Linked PV" value={grn.linkedPV} />
          ) : null}
        </View>

        {/* Items table — Item Code | Description | Ordered | Received | Balance | Condition | Remarks */}
        {grn.items && grn.items.length > 0 && (
          <View style={{ marginBottom: 12 }}>
            <View
              style={{
                borderWidth: 1,
                borderColor: "#1e40af",
              }}
            >
              {/* Header */}
              <View
                style={{
                  flexDirection: "row",
                  backgroundColor: "#f3f4f6",
                  borderBottomWidth: 1,
                  borderBottomColor: "#1e40af",
                }}
              >
                {[
                  { label: "Item Code", flex: 0.9 },
                  { label: "Description", flex: 2 },
                  { label: "Qty Ordered", flex: 0.7, align: "center" as const },
                  { label: "Qty Received", flex: 0.7, align: "center" as const },
                  { label: "Balance", flex: 0.7, align: "center" as const },
                  { label: "Condition", flex: 0.9, align: "center" as const },
                  { label: "Remarks", flex: 1.5 },
                ].map((c, i) => (
                  <Text
                    key={c.label}
                    style={{
                      flex: c.flex,
                      padding: 5,
                      fontSize: 8,
                      fontWeight: "bold",
                      color: "#1e40af",
                      textAlign: c.align ?? "left",
                      borderLeftWidth: i === 0 ? 0 : 1,
                      borderLeftColor: "#1e40af",
                    }}
                  >
                    {c.label}
                  </Text>
                ))}
              </View>
              {/* Rows */}
              {grn.items.map((item: GRNItem, index: number) => {
                // PDF sample uses "Balance" = ordered - received (positive when
                // there's a shortfall). We previously stored signed `variance`
                // (received - ordered); recompute here so the column matches
                // the form regardless of how variance was recorded.
                const balance =
                  Number(item.quantityOrdered) - Number(item.quantityReceived);
                return (
                  <View
                    key={index}
                    style={{
                      flexDirection: "row",
                      borderBottomWidth:
                        index === grn.items.length - 1 ? 0 : 1,
                      borderBottomColor: "#e5e7eb",
                    }}
                  >
                    <Text
                      style={{
                        flex: 0.9,
                        padding: 5,
                        fontSize: 8,
                        fontFamily: "Courier",
                      }}
                    >
                      {item.itemCode || "—"}
                    </Text>
                    <Text
                      style={{
                        flex: 2,
                        padding: 5,
                        fontSize: 8,
                        borderLeftWidth: 1,
                        borderLeftColor: "#e5e7eb",
                      }}
                    >
                      {item.description}
                    </Text>
                    <Text
                      style={{
                        flex: 0.7,
                        padding: 5,
                        fontSize: 8,
                        textAlign: "center",
                        borderLeftWidth: 1,
                        borderLeftColor: "#e5e7eb",
                      }}
                    >
                      {item.quantityOrdered}
                    </Text>
                    <Text
                      style={{
                        flex: 0.7,
                        padding: 5,
                        fontSize: 8,
                        textAlign: "center",
                        borderLeftWidth: 1,
                        borderLeftColor: "#e5e7eb",
                      }}
                    >
                      {item.quantityReceived}
                    </Text>
                    <Text
                      style={{
                        flex: 0.7,
                        padding: 5,
                        fontSize: 8,
                        textAlign: "center",
                        borderLeftWidth: 1,
                        borderLeftColor: "#e5e7eb",
                        color: balance > 0 ? "#991b1b" : "#166534",
                      }}
                    >
                      {balance > 0 ? `+${balance}` : balance === 0 ? "0" : balance}
                    </Text>
                    <View
                      style={{
                        flex: 0.9,
                        padding: 5,
                        alignItems: "center",
                        borderLeftWidth: 1,
                        borderLeftColor: "#e5e7eb",
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
                    <Text
                      style={{
                        flex: 1.5,
                        padding: 5,
                        fontSize: 8,
                        borderLeftWidth: 1,
                        borderLeftColor: "#e5e7eb",
                      }}
                    >
                      {item.remarks || item.notes || ""}
                    </Text>
                  </View>
                );
              })}
            </View>
          </View>
        )}

        {/* Notes block (general remarks at GRN level) */}
        {grn.notes ? (
          <View style={{ marginBottom: 12 }}>
            <Text
              style={{
                fontSize: 8,
                fontWeight: "bold",
                color: "#666",
                marginBottom: 2,
              }}
            >
              NOTES / REMARKS
            </Text>
            <Text style={{ fontSize: 9 }}>{grn.notes}</Text>
          </View>
        ) : null}

        {/* Quality issues (when reported) */}
        {grn.qualityIssues && grn.qualityIssues.length > 0 && (
          <View
            style={{
              marginBottom: 12,
              borderWidth: 1,
              borderColor: "#dc2626",
              padding: 8,
            }}
          >
            <Text
              style={{
                fontSize: 10,
                fontWeight: "bold",
                color: "#991b1b",
                marginBottom: 6,
              }}
            >
              QUALITY ISSUES REPORTED
            </Text>
            {grn.qualityIssues.map((issue: QualityIssue, index: number) => (
              <View
                key={index}
                style={{
                  marginBottom: 6,
                  padding: 6,
                  borderWidth: 1,
                  borderColor: "#fecaca",
                  backgroundColor: "#fef2f2",
                }}
              >
                <View
                  style={{
                    flexDirection: "row",
                    justifyContent: "space-between",
                    marginBottom: 3,
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

        {/* Receiver + Certifier signature blocks — always rendered to mirror
            the printed form. */}
        <View
          style={{
            flexDirection: "row",
            gap: 10,
            marginBottom: 10,
          }}
        >
          <SignatureBox
            label="Received By"
            name={grn.receivedByName}
            signature={grn.receivedBySignature}
            date={grn.receivedAt}
          />
          <SignatureBox
            label="Certified By"
            name={grn.certifiedByName}
            signature={grn.certifiedBySignature}
            date={grn.certifiedAt}
          />
          {/* Per-GRN stamp takes precedence; falls back to the org-wide stamp. */}
          <StampBox
            stampImageUrl={
              grn.stampImageUrl || documentHeader?.stampImageUrl
            }
          />
        </View>

        {/* Workflow approval-chain signatures (only when the GRN was actually
            submitted to a workflow). Always rendered on top of the form's two
            statutory signatures above. */}
        {wentThroughWorkflow ? (
          <ApprovalSignaturesSection
            approvalHistory={grn.approvalHistory}
            documentType="Goods Received Note"
          />
        ) : null}

        <PDFFooter
          qrCodeUrl={qrCodeUrl || ""}
          documentNumber={grn.documentNumber}
          document={grn}
        />
      </Page>
    </Document>
  );
};

function MetaField({ label, value }: { label: string; value: string }) {
  return (
    <View style={{ flex: 1 }}>
      <Text
        style={{
          fontSize: 7,
          fontWeight: "bold",
          color: "#666",
          marginBottom: 2,
        }}
      >
        {label.toUpperCase()}
      </Text>
      <Text style={{ fontSize: 10 }}>{value}</Text>
    </View>
  );
}

function SignatureBox({
  label,
  name,
  signature,
  date,
}: {
  label: string;
  name?: string;
  signature?: string;
  date?: string | Date;
}) {
  return (
    <View
      style={{
        flex: 1,
        borderWidth: 1,
        borderColor: "#9ca3af",
        padding: 6,
        minHeight: 110,
      }}
    >
      <Text
        style={{
          fontSize: 8,
          fontWeight: "bold",
          color: "#1e40af",
          marginBottom: 4,
        }}
      >
        {label.toUpperCase()}
      </Text>
      <Text style={{ fontSize: 7, color: "#666" }}>NAME:</Text>
      <Text style={{ fontSize: 9, marginBottom: 4 }}>{name || ""}</Text>
      <Text style={{ fontSize: 7, color: "#666" }}>SIGNATURE:</Text>
      <View
        style={{
          height: 40,
          marginBottom: 4,
          alignItems: "center",
          justifyContent: "center",
        }}
      >
        {signature ? (
          <Image
            src={signature}
            style={{ maxHeight: 38, maxWidth: "100%" }}
          />
        ) : null}
      </View>
      <Text style={{ fontSize: 7, color: "#666" }}>DATE:</Text>
      <Text style={{ fontSize: 9 }}>{formatDate(date)}</Text>
    </View>
  );
}

function StampBox({ stampImageUrl }: { stampImageUrl?: string }) {
  return (
    <View
      style={{
        flex: 1,
        borderWidth: 1,
        borderColor: "#9ca3af",
        padding: 6,
        minHeight: 110,
        alignItems: "center",
        justifyContent: "center",
      }}
    >
      <Text
        style={{
          fontSize: 8,
          fontWeight: "bold",
          color: "#1e40af",
          marginBottom: 4,
        }}
      >
        STAMP OF ISSUING OFFICER
      </Text>
      <View
        style={{
          flex: 1,
          alignItems: "center",
          justifyContent: "center",
          width: "100%",
        }}
      >
        {stampImageUrl ? (
          <Image
            src={stampImageUrl}
            style={{ maxHeight: 70, maxWidth: "90%" }}
          />
        ) : null}
      </View>
    </View>
  );
}

export { GoodsReceivedNotePDF };
export default GoodsReceivedNotePDF;
