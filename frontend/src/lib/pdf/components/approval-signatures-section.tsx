import React from "react";
import { View, Text, Image } from "@react-pdf/renderer";
import { ApprovalRecord } from "@/types/core";

interface ApprovalSignaturesSectionProps {
  approvalHistory: ApprovalRecord[];
  documentType?: string;
}

/**
 * Compact Approval Chain Section
 * Renders approved signatories before the PDF footer.
 * Adapts column count to chain length: 1→1col, 2→2col, 3-5→3col, 6+→4col
 */
export const ApprovalSignaturesSection: React.FC<
  ApprovalSignaturesSectionProps
> = ({ approvalHistory }) => {
  const approvedRecords = approvalHistory.filter(
    (r) => r.status === "approved" || r.status === "APPROVED",
  );

  if (!approvedRecords || approvedRecords.length === 0) return null;

  const count = approvedRecords.length;
  const cols = count === 1 ? 1 : count === 2 ? 2 : count <= 5 ? 3 : 4;
  // Percentage widths that fit within react-pdf's flex model
  const widthMap: Record<number, string> = {
    1: "58%",
    2: "47%",
    3: "30%",
    4: "22%",
  };
  const cardWidth = widthMap[cols];

  const formatDate = (date: Date | string | undefined): string => {
    if (!date) return "—";
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
      return "—";
    }
  };

  const getName = (r: ApprovalRecord) =>
    r.approverName || r.actionTakenBy || "Unknown";

  const getRole = (r: ApprovalRecord) =>
    r.assignedRole || r.actionTakenByRole || null;

  const getDate = (r: ApprovalRecord) => r.approvedAt || r.actionTakenAt;

  const getStageLabel = (r: ApprovalRecord, idx: number): string => {
    if (r.stageName) {
      return r.stageNumber != null
        ? `${r.stageName} (Stage ${r.stageNumber})`
        : r.stageName;
    }
    return r.stageNumber != null ? `Stage ${r.stageNumber}` : `Stage ${idx + 1}`;
  };

  return (
    <View
      style={{
        marginTop: 20,
        paddingTop: 14,
        borderTop: "1.5px solid #ddd",
      }}
    >
      {/* Section heading */}
      <Text
        style={{
          fontSize: 9,
          fontWeight: "bold",
          marginBottom: 10,
          color: "#1e3a8a",
          textAlign: "center",
          letterSpacing: 0.8,
          textTransform: "uppercase",
        }}
      >
        Approval Chain
      </Text>

      {/* Card grid */}
      <View
        style={{
          flexDirection: "row",
          flexWrap: "wrap",
          gap: 8,
          justifyContent: count === 1 ? "center" : "flex-start",
        }}
      >
        {approvedRecords.map((r, idx) => (
          <View
            key={idx}
            style={{
              width: cardWidth,
              border: "1px solid #d1d5db",
              borderRadius: 3,
              padding: 8,
              backgroundColor: "#f9fafb",
            }}
          >
            {/* Stage label */}
            <Text
              style={{
                fontSize: 6.5,
                fontWeight: "bold",
                color: "#1e40af",
                marginBottom: 4,
                textTransform: "uppercase",
                letterSpacing: 0.3,
              }}
            >
              {getStageLabel(r, idx)}
            </Text>

            {/* Approver name */}
            <Text
              style={{
                fontSize: 9,
                fontWeight: "bold",
                color: "#111827",
                marginBottom: 2,
              }}
            >
              {getName(r)}
            </Text>

            {/* Position / role */}
            {getRole(r) && (
              <Text
                style={{
                  fontSize: 7,
                  color: "#6b7280",
                  fontStyle: "italic",
                  marginBottom: 6,
                }}
              >
                {getRole(r)}
              </Text>
            )}

            {/* Signature area */}
            <View
              style={{
                minHeight: 30,
                borderBottom: "1px solid #d1d5db",
                marginBottom: 5,
                alignItems: "center",
                justifyContent: "center",
              }}
            >
              {r.signature ? (
                <Image
                  src={r.signature}
                  style={{ maxHeight: 28, maxWidth: "100%", objectFit: "contain" }}
                />
              ) : (
                <Text
                  style={{ fontSize: 6.5, color: "#9ca3af", fontStyle: "italic" }}
                >
                  [Electronically Approved]
                </Text>
              )}
            </View>

            {/* Approval date */}
            <Text style={{ fontSize: 6.5, color: "#6b7280" }}>
              <Text style={{ fontWeight: "bold" }}>Approved: </Text>
              {formatDate(getDate(r))}
            </Text>
          </View>
        ))}
      </View>

      {/* Footer note */}
      <Text
        style={{
          fontSize: 6.5,
          color: "#9ca3af",
          marginTop: 10,
          textAlign: "center",
          fontStyle: "italic",
        }}
      >
        This document has been electronically approved by the above signatories.
      </Text>
    </View>
  );
};
