import React from "react";
import { View, Text, Image } from "@react-pdf/renderer";
import { ApprovalRecord } from "@/types/core";

interface ApprovalSignaturesSectionProps {
  approvalHistory: ApprovalRecord[];
  documentType?: string;
}

/**
 * Approval Signatures Section Component
 * Displays a professional approval signatures section at the bottom of PDF documents
 * Shows approver name, position/role, signature, and approval timestamp
 */
export const ApprovalSignaturesSection: React.FC<
  ApprovalSignaturesSectionProps
> = ({ approvalHistory, documentType }) => {
  // Filter only approved records
  const approvedRecords = approvalHistory.filter(
    (record) => record.status === "approved" || record.status === "APPROVED",
  );

  // Don't render if no approvals
  if (!approvedRecords || approvedRecords.length === 0) {
    return null;
  }

  // Format date helper
  const formatDate = (date: Date | string | undefined): string => {
    if (!date) return "N/A";
    try {
      const d = typeof date === "string" ? new Date(date) : date;
      return d.toLocaleString("en-US", {
        month: "short",
        day: "2-digit",
        year: "numeric",
        hour: "2-digit",
        minute: "2-digit",
        hour12: true,
      });
    } catch {
      return "N/A";
    }
  };

  // Get approver name
  const getApproverName = (approval: ApprovalRecord): string => {
    return (
      approval.approverName || approval.actionTakenBy || "Unknown Approver"
    );
  };

  // Get approver role/position
  const getApproverRole = (approval: ApprovalRecord): string | null => {
    return approval.assignedRole || approval.actionTakenByRole || null;
  };

  // Get approval timestamp
  const getApprovalDate = (
    approval: ApprovalRecord,
  ): Date | string | undefined => {
    return approval.approvedAt || approval.actionTakenAt;
  };

  return (
    <View
      style={{
        marginTop: 30,
        paddingTop: 20,
        borderTop: "1.5px solid #ddd",
      }}
    >
      {/* Section Header */}
      <Text
        style={{
          fontSize: 12,
          fontWeight: "bold",
          marginBottom: 15,
          color: "#333",
          textAlign: "center",
          letterSpacing: 0.5,
        }}
      >
        APPROVAL SIGNATURES
      </Text>

      {/* Approval Grid */}
      <View
        style={{
          display: "flex",
          flexDirection: "row",
          flexWrap: "wrap",
          gap: 12,
          justifyContent:
            approvedRecords.length === 1 ? "center" : "flex-start",
        }}
      >
        {approvedRecords.map((approval, index) => (
          <View
            key={index}
            style={{
              width: approvedRecords.length === 1 ? "60%" : "48%",
              border: "1px solid #e0e0e0",
              borderRadius: 4,
              padding: 12,
              backgroundColor: "#fafafa",
            }}
          >
            {/* Stage Information */}
            {approval.stageName && (
              <Text
                style={{
                  fontSize: 8,
                  fontWeight: "bold",
                  color: "#0066cc",
                  marginBottom: 8,
                  textTransform: "uppercase",
                  letterSpacing: 0.3,
                }}
              >
                {approval.stageName}
                {approval.stageNumber !== undefined &&
                  approval.stageNumber !== null &&
                  ` (Stage ${approval.stageNumber})`}
              </Text>
            )}

            {/* Approver Name */}
            <Text
              style={{
                fontSize: 10,
                fontWeight: "bold",
                marginBottom: 4,
                color: "#333",
              }}
            >
              {getApproverName(approval)}
            </Text>

            {/* Position/Role */}
            {getApproverRole(approval) && (
              <Text
                style={{
                  fontSize: 8,
                  color: "#666",
                  marginBottom: 8,
                  fontStyle: "italic",
                }}
              >
                {getApproverRole(approval)}
              </Text>
            )}

            {/* Signature */}
            <View
              style={{
                marginVertical: 8,
                minHeight: 40,
                borderBottom: "1px solid #ddd",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
              }}
            >
              {approval.signature ? (
                <Image
                  src={approval.signature}
                  style={{
                    maxHeight: 35,
                    maxWidth: "100%",
                    objectFit: "contain",
                  }}
                />
              ) : (
                <Text
                  style={{
                    fontSize: 7,
                    color: "#999",
                    fontStyle: "italic",
                    textAlign: "center",
                  }}
                >
                  [Electronically Approved]
                </Text>
              )}
            </View>

            {/* Approval Date */}
            <Text
              style={{
                fontSize: 7,
                color: "#666",
                marginTop: 4,
              }}
            >
              <Text style={{ fontWeight: "bold" }}>Approved on: </Text>
              {formatDate(getApprovalDate(approval))}
            </Text>

            {/* Comments (if any) */}
            {approval.comments && (
              <View
                style={{
                  marginTop: 6,
                  paddingTop: 6,
                  borderTop: "0.5px solid #e0e0e0",
                }}
              >
                <Text
                  style={{
                    fontSize: 7,
                    color: "#666",
                    fontStyle: "italic",
                  }}
                >
                  <Text style={{ fontWeight: "bold" }}>Note: </Text>
                  {approval.comments}
                </Text>
              </View>
            )}
          </View>
        ))}
      </View>

      {/* Footer Note */}
      <Text
        style={{
          fontSize: 7,
          color: "#999",
          marginTop: 15,
          textAlign: "center",
          fontStyle: "italic",
        }}
      >
        This document has been electronically approved by the above signatories
      </Text>
    </View>
  );
};
