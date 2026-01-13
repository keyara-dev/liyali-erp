import React from "react";
import { Document, Page, Text, View, Image } from "@react-pdf/renderer";
import { PaymentVoucher } from "@/types/payment-voucher";
import { pdfStyles } from "./pdf-styles";
import { generateDocumentQRData, generateTrackingCode } from "./qr-utils";

interface PaymentVoucherPDFProps {
  paymentVoucher: PaymentVoucher;
  qrCodeUrl?: string;
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
}) => {
  const documentNumber = paymentVoucher.documentNumber;
  const trackingCode = generateTrackingCode("PAYMENT_VOUCHER", documentNumber);
  const qrData = generateDocumentQRData(
    "PAYMENT_VOUCHER",
    documentNumber,
    paymentVoucher.id,
    new Date(paymentVoucher.createdAt)
  );

  return (
    <Document>
      <Page size="A4" style={pdfStyles.page}>
        {/* Header with Republic of Zambia and Logo */}
        <View style={{ marginBottom: 20, textAlign: "center" }}>
          <Text style={{ fontSize: 11, fontWeight: "bold", marginBottom: 5 }}>
            REPUBLIC OF ZAMBIA
          </Text>
          <Text style={{ fontSize: 14, fontWeight: "bold", marginBottom: 8 }}>
            PAYMENT VOUCHER
          </Text>
        </View>

        {/* Main Header Section */}
        <View
          style={[
            pdfStyles.header,
            {
              marginBottom: 20,
              flexDirection: "row",
              justifyContent: "space-between",
            },
          ]}
        >
          <View>
            <Text style={{ fontSize: 14, fontWeight: "bold", marginBottom: 3 }}>
              Liyali
            </Text>
            <Text style={{ fontSize: 9, color: "#666" }}>
              Finance & Procurement System
            </Text>
          </View>
          <View style={{ textAlign: "right" }}>
            <Text style={{ fontSize: 11, fontWeight: "bold", marginBottom: 2 }}>
              Document No: {documentNumber}
            </Text>
            <Text style={{ fontSize: 9, color: "#666", marginBottom: 4 }}>
              Date: {new Date(paymentVoucher.createdAt).toLocaleDateString()}
            </Text>
            <View
              style={{
                borderWidth: 1,
                borderColor: "#ddd",
                padding: 6,
                width: 80,
                textAlign: "center",
                marginLeft: "auto",
              }}
            >
              <Text
                style={{ fontSize: 8, fontWeight: "bold", marginBottom: 2 }}
              >
                TRACKING CODE
              </Text>
              <Text style={{ fontSize: 7 }}>{trackingCode}</Text>
            </View>
          </View>
        </View>

        {/* Status Badges */}
        <View style={{ marginBottom: 20, flexDirection: "row", gap: 10 }}>
          <View
            style={[
              pdfStyles.statusBadge,
              getStatusColor(paymentVoucher.status),
            ]}
          >
            <Text style={{ fontSize: 9 }}>{paymentVoucher.status}</Text>
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

        {/* Vendor and Payment Information */}
        <View
          style={{
            marginBottom: 15,
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
            PAYEE &amp; PAYMENT INFORMATION
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
              <Text style={{ fontSize: 9, color: "#999", marginTop: 8 }}>
                VENDOR ID
              </Text>
              <Text style={{ fontSize: 9 }}>
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
                PAYMENT METHOD
              </Text>
              <Text style={{ fontSize: 10 }}>
                {paymentVoucher.paymentMethod || "—"}
              </Text>
              <Text style={{ fontSize: 9, color: "#999", marginTop: 8 }}>
                PAYMENT DUE DATE
              </Text>
              <Text style={{ fontSize: 10 }}>
                {paymentVoucher.paymentDueDate
                  ? new Date(paymentVoucher.paymentDueDate).toLocaleDateString()
                  : "—"}
              </Text>
            </View>
          </View>

          {/* Bank Details (if applicable) */}
          {paymentVoucher.paymentMethod === "bank_transfer" &&
            paymentVoucher.bankDetails && (
              <View
                style={{
                  borderTopWidth: 1,
                  borderTopColor: "#ddd",
                  paddingTop: 10,
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
                      <Text style={{ fontSize: 9, fontFamily: "Courier" }}>
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

        {/* Voucher Details */}
        <View
          style={{
            marginBottom: 15,
            display: "flex",
            flexDirection: "row",
            gap: 15,
          }}
        >
          <View
            style={{ flex: 1, borderWidth: 1, borderColor: "#ddd", padding: 8 }}
          >
            <Text
              style={{
                fontSize: 8,
                fontWeight: "bold",
                marginBottom: 2,
                color: "#666",
              }}
            >
              DOCUMENT NUMBER
            </Text>
            <Text style={{ fontSize: 10 }}>{documentNumber || "—"}</Text>
          </View>
          <View
            style={{ flex: 1, borderWidth: 1, borderColor: "#ddd", padding: 8 }}
          >
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
          <View
            style={{ flex: 1, borderWidth: 1, borderColor: "#ddd", padding: 8 }}
          >
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
        </View>

        {/* Description and Amount */}
        <View
          style={{
            marginBottom: 15,
            display: "flex",
            flexDirection: "row",
            gap: 15,
          }}
        >
          <View
            style={{ flex: 2, borderWidth: 1, borderColor: "#ddd", padding: 8 }}
          >
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
          <View
            style={{
              flex: 1,
              borderWidth: 1,
              borderColor: "#1e40af",
              padding: 8,
              backgroundColor: "#f0f7ff",
            }}
          >
            <Text
              style={{
                fontSize: 8,
                fontWeight: "bold",
                marginBottom: 2,
                color: "#1e40af",
              }}
            >
              TOTAL AMOUNT
            </Text>
            <Text
              style={{ fontSize: 14, fontWeight: "bold", color: "#1e40af" }}
            >
              {paymentVoucher.currency}{" "}
              {paymentVoucher.totalAmount?.toLocaleString() || "0"}
            </Text>
          </View>
        </View>

        {/* Line Items Table (if applicable) */}
        {paymentVoucher.items && paymentVoucher.items.length > 0 && (
          <View style={{ marginBottom: 15 }}>
            <Text style={{ fontSize: 9, fontWeight: "bold", marginBottom: 6 }}>
              PAYMENT BREAKDOWN
            </Text>

            {/* Table Header */}
            <View style={[pdfStyles.tableHeaderRow, { paddingVertical: 5 }]}>
              <Text style={{ ...pdfStyles.tableHeaderCell, width: "8%" }}>
                #
              </Text>
              <Text style={{ ...pdfStyles.tableHeaderCell, width: "40%" }}>
                Description
              </Text>
              <Text
                style={{
                  ...pdfStyles.tableHeaderCell,
                  width: "15%",
                  textAlign: "right",
                }}
              >
                Quantity
              </Text>
              <Text
                style={{
                  ...pdfStyles.tableHeaderCell,
                  width: "18%",
                  textAlign: "right",
                }}
              >
                Unit Price
              </Text>
              <Text
                style={{
                  ...pdfStyles.tableHeaderCell,
                  width: "19%",
                  textAlign: "right",
                }}
              >
                Amount
              </Text>
            </View>

            {/* Table Rows */}
            {paymentVoucher.items.map((item: any, index: number) => (
              <View
                key={item.id}
                style={[pdfStyles.tableRow, { paddingVertical: 4 }]}
              >
                <Text
                  style={{ ...pdfStyles.tableCell, width: "8%", fontSize: 8 }}
                >
                  {index + 1}
                </Text>
                <Text
                  style={{ ...pdfStyles.tableCell, width: "40%", fontSize: 8 }}
                >
                  {item.description}
                </Text>
                <Text
                  style={{
                    ...pdfStyles.tableCell,
                    width: "15%",
                    textAlign: "right",
                    fontSize: 8,
                  }}
                >
                  {item.quantity}
                </Text>
                <Text
                  style={{
                    ...pdfStyles.tableCell,
                    width: "18%",
                    textAlign: "right",
                    fontSize: 8,
                  }}
                >
                  {paymentVoucher.currency}{" "}
                  {item.unitPrice?.toLocaleString() || "0"}
                </Text>
                <Text
                  style={{
                    ...pdfStyles.tableCell,
                    width: "19%",
                    textAlign: "right",
                    fontSize: 8,
                  }}
                >
                  {paymentVoucher.currency}{" "}
                  {item.amount?.toLocaleString() || "0"}
                </Text>
              </View>
            ))}
          </View>
        )}

        {/* Financial Information Section */}
        {(paymentVoucher.budgetCode || paymentVoucher.costCenter) && (
          <View
            style={{
              marginBottom: 15,
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
                  <Text style={{ fontSize: 9, fontFamily: "Courier" }}>
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
                  <Text style={{ fontSize: 9, fontFamily: "Courier" }}>
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
                  <Text style={{ fontSize: 9, fontFamily: "Courier" }}>
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
              marginBottom: 15,
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

        {/* APPROVAL SIGNATURES - Dynamic based on actual approval chain */}
        {paymentVoucher.approvalChain &&
          paymentVoucher.approvalChain.length > 0 && (
            <View
              style={{
                marginBottom: 15,
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
                APPROVAL SIGNATURES - {paymentVoucher.approvalChain.length}{" "}
                Stage(s)
              </Text>

              <View
                style={{
                  display: "flex",
                  flexDirection: "row",
                  gap: 10,
                  flexWrap: "wrap",
                }}
              >
                {paymentVoucher.approvalChain!.map(
                  (stage: any, index: number) => (
                    <View
                      key={index}
                      style={{
                        flex: 1,
                        minWidth:
                          paymentVoucher.approvalChain!.length === 3
                            ? index === paymentVoucher.approvalChain!.length - 1
                              ? "100%"
                              : "48%"
                            : paymentVoucher.approvalChain!.length === 2
                              ? "48%"
                              : paymentVoucher.approvalChain!.length === 4
                                ? "48%"
                                : "48%",
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
                          marginBottom: 2,
                          color: "#1e40af",
                        }}
                      >
                        {stage.stageName ||
                          `Approval Stage ${stage.stageNumber}`}
                      </Text>
                      <Text style={{ fontSize: 7, marginBottom: 2 }}>
                        Assigned to: {stage.assignedTo}
                      </Text>
                      <Text style={{ fontSize: 7, marginBottom: 3 }}>
                        Status: {stage.status}
                      </Text>
                      {stage.actionTakenAt && (
                        <Text style={{ fontSize: 7, color: "#666" }}>
                          Date:{" "}
                          {new Date(stage.actionTakenAt).toLocaleDateString()}
                        </Text>
                      )}
                      <View
                        style={{
                          marginTop: 6,
                          minHeight: 25,
                          borderTopWidth: 1,
                          borderTopColor: "#999",
                          paddingTop: 3,
                        }}
                      >
                        <Text style={{ fontSize: 6, color: "#999" }}>
                          Signature/Stamp
                        </Text>
                      </View>
                    </View>
                  )
                )}
              </View>
            </View>
          )}

        {/* Source Documents */}
        {paymentVoucher.linkedPO && (
          <View
            style={{
              marginBottom: 15,
              display: "flex",
              flexDirection: "row",
              gap: 10,
            }}
          >
            <View
              style={{
                flex: 1,
                padding: 10,
                backgroundColor: "#f0f7ff",
                borderLeftWidth: 4,
                borderLeftColor: "#2563eb",
              }}
            >
              <Text
                style={{ fontSize: 8, fontWeight: "bold", marginBottom: 2 }}
              >
                SOURCE PO
              </Text>
              <Text style={{ fontSize: 9 }}>{paymentVoucher.linkedPO}</Text>
            </View>
          </View>
        )}

        {/* QR Code and Tracking Information */}
        <View
          style={{
            marginTop: 15,
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
              Tracking Code: {trackingCode}
            </Text>
            <Text style={{ fontSize: 7, marginBottom: 2 }}>
              Document ID: {paymentVoucher.id}
            </Text>
            <Text style={{ fontSize: 7, marginBottom: 2 }}>
              Status: {paymentVoucher.status}
            </Text>
            <Text style={{ fontSize: 7 }}>
              Generated: {new Date().toLocaleDateString()}{" "}
              {new Date().toLocaleTimeString()}
            </Text>
          </View>
        </View>

        {/* Footer */}
        <View
          style={{
            marginTop: "auto",
            paddingTop: 10,
            borderTopWidth: 1,
            borderTopColor: "#ddd",
            textAlign: "center",
          }}
        >
          <Text style={{ fontSize: 7, color: "#999" }}>
            This is a system-generated document. Digital signatures and QR codes
            verify authenticity.
          </Text>
          <Text style={{ fontSize: 7, color: "#999", marginTop: 2 }}>
            Scan the QR code above to verify this document.
          </Text>
        </View>
      </Page>
    </Document>
  );
};

export default PaymentVoucherPDF;
