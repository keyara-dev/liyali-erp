import React from "react";
import { Document, Page, Text, View, Image } from "@react-pdf/renderer";
import { PaymentVoucher } from "@/types/payment-voucher";
import { pdfStyles } from "./pdf-styles";
import { generateDocumentQRData } from "./qr-utils";
import { PDFHeader, PDFFooter, DocumentHeader } from "./requisition-pdf";
import { capitalize } from "../utils";
import { ApprovalSignaturesSection } from "./components/approval-signatures-section";

interface PaymentVoucherPDFProps {
  paymentVoucher: PaymentVoucher;
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
    case "PAID":
      return {
        ...pdfStyles.statusApproved,
        backgroundColor: "#d1fae5",
        color: "#065f46",
      };
    default:
      return pdfStyles.statusDraft;
  }
};

const PaymentVoucherPDF: React.FC<PaymentVoucherPDFProps> = ({
  paymentVoucher,
  qrCodeUrl,
  organizationLogoUrl,
  documentHeader,
}) => {
  const documentNumber = paymentVoucher.documentNumber;
  const qrData = generateDocumentQRData(
    "PAYMENT_VOUCHER",
    documentNumber,
    paymentVoucher.id,
    new Date(paymentVoucher.createdAt),
  );

  return (
    <Document>
      <Page size="A4" style={pdfStyles.page}>
        {/* Header with Republic of Zambia and Logo */}
        <PDFHeader
          title="PAYMENT VOUCHER"
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
              Date: {new Date(paymentVoucher.createdAt).toLocaleDateString()}
            </Text>
          </View>

          {/* STATUS AND PRIORITY BADGES */}
          <View style={{ textAlign: "right" }}>
            <Text style={{ fontSize: 7, fontWeight: "bold", marginBottom: 1 }}>
              STATUS
            </Text>
            <View style={{ marginBottom: 0, flexDirection: "row", gap: 4 }}>
              <View
                style={[
                  pdfStyles.statusBadge,
                  getStatusColor(paymentVoucher.status),
                ]}
              >
                <Text style={{ fontSize: 9 }}>
                  {capitalize(paymentVoucher.status)}
                </Text>
              </View>
            </View>
          </View>
        </View>

        {/* Payment Instruction Box */}
        <View
          style={{
            marginBottom: 15,
            padding: 8,
            backgroundColor: "#fef3c7",
            borderWidth: 1,
            borderColor: "#fcd34d",
          }}
        >
          <Text style={{ fontSize: 8, fontWeight: "bold", marginBottom: 3 }}>
            PAYMENT INSTRUCTIONS:
          </Text>
          <Text style={{ fontSize: 7, lineHeight: 1.4 }}>
            • All invoices must be attached with this voucher • Payment should
            be processed within the specified terms • Keep original copy for
            audit trail • QR code below provides digital verification
          </Text>
        </View>

        {/* SECTION 1: PAYEE & PAYMENT INFORMATION */}
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
            SECTION 1: PAYEE & PAYMENT INFORMATION
          </Text>

          <View
            style={{
              display: "flex",
              flexDirection: "row",
              gap: 20,
              marginBottom: 12,
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
                PAYEE NAME
              </Text>
              <Text style={{ fontSize: 10 }}>
                {paymentVoucher.vendorName || "—"}
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
                PAYMENT METHOD
              </Text>
              <Text style={{ fontSize: 10 }}>
                {paymentVoucher.paymentMethod || "—"}
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
                PAYMENT DUE DATE
              </Text>
              <Text style={{ fontSize: 10 }}>
                {paymentVoucher.paymentDueDate
                  ? new Date(paymentVoucher.paymentDueDate).toLocaleDateString()
                  : "—"}
              </Text>
            </View>
          </View>

          {/* Description */}
          <View style={{ marginBottom: 12 }}>
            <Text
              style={{
                fontSize: 8,
                fontWeight: "bold",
                marginBottom: 2,
                color: "#666",
              }}
            >
              DESCRIPTION OF PAYMENT
            </Text>
            <Text style={{ fontSize: 9 }}>
              {paymentVoucher.description || "—"}
            </Text>
          </View>

          {/* Source and Amount Row */}
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
                SOURCE PO
              </Text>
              <Text style={{ fontSize: 10 }}>
                {paymentVoucher.linkedPO || "—"}
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
                VENDOR ID
              </Text>
              <Text style={{ fontSize: 10 }}>
                {paymentVoucher.vendorId || "—"}
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
                REQUEST DATE
              </Text>
              <Text style={{ fontSize: 10 }}>
                {paymentVoucher.requestedDate
                  ? new Date(paymentVoucher.requestedDate).toLocaleDateString()
                  : "—"}
              </Text>
            </View>
          </View>

          {/* Bank Details (if applicable) */}
          {paymentVoucher.paymentMethod === "bank_transfer" &&
            paymentVoucher.bankDetails && (
              <View
                style={{
                  marginTop: 12,
                  paddingTop: 10,
                  borderTopWidth: 1,
                  borderTopColor: "#ddd",
                }}
              >
                <Text
                  style={{
                    fontSize: 8,
                    fontWeight: "bold",
                    marginBottom: 4,
                    color: "#1e40af",
                  }}
                >
                  BANK TRANSFER DETAILS:
                </Text>
                <View
                  style={{ display: "flex", flexDirection: "row", gap: 20 }}
                >
                  {paymentVoucher.bankDetails.bankName && (
                    <View style={{ flex: 1 }}>
                      <Text
                        style={{
                          fontSize: 8,
                          fontWeight: "bold",
                          marginBottom: 2,
                          color: "#666",
                        }}
                      >
                        BANK NAME
                      </Text>
                      <Text style={{ fontSize: 9 }}>
                        {paymentVoucher.bankDetails.bankName}
                      </Text>
                    </View>
                  )}
                  {paymentVoucher.bankDetails.accountNumber && (
                    <View style={{ flex: 1 }}>
                      <Text
                        style={{
                          fontSize: 8,
                          fontWeight: "bold",
                          marginBottom: 2,
                          color: "#666",
                        }}
                      >
                        ACCOUNT NUMBER
                      </Text>
                      <Text style={{ fontSize: 9 }}>
                        {paymentVoucher.bankDetails.accountNumber}
                      </Text>
                    </View>
                  )}
                  {paymentVoucher.bankDetails.accountName && (
                    <View style={{ flex: 1 }}>
                      <Text
                        style={{
                          fontSize: 8,
                          fontWeight: "bold",
                          marginBottom: 2,
                          color: "#666",
                        }}
                      >
                        ACCOUNT HOLDER
                      </Text>
                      <Text style={{ fontSize: 9 }}>
                        {paymentVoucher.bankDetails.accountName}
                      </Text>
                    </View>
                  )}
                </View>
              </View>
            )}
        </View>

        {/* Line Items Table (if applicable) */}
        {paymentVoucher.items && paymentVoucher.items.length > 0 && (
          <View style={{ marginBottom: 20 }}>
            <Text style={{ fontSize: 10, fontWeight: "bold", marginBottom: 8 }}>
              PAYMENT BREAKDOWN:
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
                  Amount
                </Text>
              </View>

              {/* Table Rows */}
              {paymentVoucher.items.map((item: any, index: number) => (
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
                    {item.description}
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
                    {item.quantity}
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
                    {paymentVoucher.currency}{" "}
                    {item.unitPrice?.toLocaleString() || "0"}
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
                    {paymentVoucher.currency}{" "}
                    {item.amount?.toLocaleString() || "0"}
                  </Text>
                </View>
              ))}
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
                    TOTAL AMOUNT:
                  </Text>
                  <Text
                    style={{
                      fontSize: 11,
                      fontWeight: "bold",
                      color: "#166534",
                    }}
                  >
                    {paymentVoucher.currency}{" "}
                    {paymentVoucher.totalAmount?.toLocaleString() || "0"}
                  </Text>
                </View>
              </View>
            </View>
          </View>
        )}

        {/* Budget Allocation Section */}
        {(paymentVoucher.budgetCode || paymentVoucher.costCenter) && (
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
              BUDGET ALLOCATION DETAILS
            </Text>

            <View style={{ display: "flex", flexDirection: "row", gap: 20 }}>
              {paymentVoucher.budgetCode && (
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
                    {paymentVoucher.budgetCode}
                  </Text>
                </View>
              )}
              {paymentVoucher.costCenter && (
                <View style={{ flex: 1 }}>
                  <Text
                    style={{
                      fontSize: 8,
                      fontWeight: "bold",
                      marginBottom: 2,
                      color: "#666",
                    }}
                  >
                    COST CENTER
                  </Text>
                  <Text style={{ fontSize: 9 }}>
                    {paymentVoucher.costCenter}
                  </Text>
                </View>
              )}
              {paymentVoucher.projectCode && (
                <View style={{ flex: 1 }}>
                  <Text
                    style={{
                      fontSize: 8,
                      fontWeight: "bold",
                      marginBottom: 2,
                      color: "#666",
                    }}
                  >
                    PROJECT CODE
                  </Text>
                  <Text style={{ fontSize: 9 }}>
                    {paymentVoucher.projectCode}
                  </Text>
                </View>
              )}
            </View>

            {/* Tax Information (if applicable) */}
            {(paymentVoucher.taxAmount ||
              paymentVoucher.withholdingTaxAmount) && (
              <View
                style={{
                  marginTop: 10,
                  paddingTop: 10,
                  borderTopWidth: 1,
                  borderTopColor: "#ddd",
                  display: "flex",
                  flexDirection: "row",
                  gap: 20,
                }}
              >
                {paymentVoucher.taxAmount && (
                  <View style={{ flex: 1 }}>
                    <Text
                      style={{
                        fontSize: 8,
                        fontWeight: "bold",
                        marginBottom: 2,
                        color: "#666",
                      }}
                    >
                      TAX AMOUNT
                    </Text>
                    <Text style={{ fontSize: 9 }}>
                      {paymentVoucher.currency}{" "}
                      {paymentVoucher.taxAmount.toLocaleString()}
                    </Text>
                  </View>
                )}
                {paymentVoucher.withholdingTaxAmount && (
                  <View style={{ flex: 1 }}>
                    <Text
                      style={{
                        fontSize: 8,
                        fontWeight: "bold",
                        marginBottom: 2,
                        color: "#666",
                      }}
                    >
                      WITHHOLDING TAX
                    </Text>
                    <Text style={{ fontSize: 9 }}>
                      {paymentVoucher.currency}{" "}
                      {paymentVoucher.withholdingTaxAmount.toLocaleString()}
                    </Text>
                  </View>
                )}
              </View>
            )}
          </View>
        )}

        {/* Payment Confirmation (if PAID) */}
        {paymentVoucher.status === "paid" && (
          <View
            style={{
              marginBottom: 20,
              padding: 10,
              backgroundColor: "#dcfce7",
              borderLeftWidth: 4,
              borderLeftColor: "#16a34a",
            }}
          >
            <Text
              style={{
                fontSize: 8,
                fontWeight: "bold",
                marginBottom: 4,
                color: "#166534",
              }}
            >
              PAYMENT CONFIRMATION
            </Text>
            {paymentVoucher.paidAmount && (
              <Text style={{ fontSize: 9, marginBottom: 2 }}>
                Amount Paid: {paymentVoucher.currency}{" "}
                {paymentVoucher.paidAmount.toLocaleString()}
              </Text>
            )}
            {paymentVoucher.paidDate && (
              <Text style={{ fontSize: 9 }}>
                Date Paid:{" "}
                {new Date(paymentVoucher.paidDate).toLocaleDateString()}
              </Text>
            )}
          </View>
        )}

        {/* APPROVAL CHAIN */}
        {paymentVoucher.approvalChain &&
          paymentVoucher.approvalChain.length > 0 && (
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
                {paymentVoucher.approvalChain.map(
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
              Document ID: {paymentVoucher.id}
            </Text>
            <Text style={{ fontSize: 7, marginBottom: 2 }}>
              Status: {capitalize(paymentVoucher.status)}
            </Text>
            <Text style={{ fontSize: 7, marginBottom: 2 }}>
              Created: {new Date(paymentVoucher.createdAt).toLocaleDateString()}{" "}
              {new Date(paymentVoucher.createdAt).toLocaleTimeString()}
            </Text>
            <Text style={{ fontSize: 7 }}>
              Generated: {new Date().toLocaleDateString()}{" "}
              {new Date().toLocaleTimeString()}
            </Text>
          </View>
        </View>

        {/* Approval Signatures Section */}
        {paymentVoucher.approvalHistory &&
          paymentVoucher.approvalHistory.length > 0 && (
            <ApprovalSignaturesSection
              approvalHistory={paymentVoucher.approvalHistory}
              documentType="Payment Voucher"
            />
          )}

        {/* Footer */}
        <PDFFooter />
      </Page>
    </Document>
  );
};

export default PaymentVoucherPDF;
