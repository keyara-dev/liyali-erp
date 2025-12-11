"use client";

import { useState } from "react";
import * as React from "react";
import { useMutation } from "@tanstack/react-query";
import { ColumnDef } from "@tanstack/react-table";
import { useRouter } from "next/navigation";
import { ArrowUpDown, Eye } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { DataTable } from "@/components/ui/data-table";
import { Skeleton } from "@/components/ui/skeleton";
import {
  WorkflowDocument,
  SearchFilters,
} from "@/types/workflow";
import { DownloadButton } from "./download-button";
import {
  getPurchaseOrders,
  getRequisitions,
  getPaymentVouchers,
  getGoodsReceivedNotes,
} from "@/lib/storage";

// Table skeleton loader
function TransactionTableSkeleton() {
  return (
    <div className="rounded-md border overflow-hidden">
      <div className="space-y-2 p-4">
        {/* Header row */}
        <div className="flex gap-4 pb-4 border-b">
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} className="h-4 flex-1" />
          ))}
        </div>
        {/* Data rows */}
        {Array.from({ length: 5 }).map((_, rowIdx) => (
          <div key={rowIdx} className="flex gap-4 py-3 border-b last:border-0">
            {Array.from({ length: 5 }).map((_, colIdx) => (
              <Skeleton key={colIdx} className="h-4 flex-1" />
            ))}
          </div>
        ))}
      </div>
    </div>
  );
}

interface TransactionResultsProps {
  filters: SearchFilters;
  refreshTrigger: number;
  userRole: string;
  onSearchComplete?: () => void;
}

const STATUS_COLORS: Record<string, string> = {
  DRAFT: "outline",
  SUBMITTED: "secondary",
  IN_REVIEW: "default",
  APPROVED: "default",
  REJECTED: "destructive",
  REVERSED: "secondary",
};

const STATUS_LABELS: Record<string, string> = {
  DRAFT: "Draft",
  SUBMITTED: "Submitted",
  IN_REVIEW: "In Approval",
  APPROVED: "Approved",
  REJECTED: "Rejected",
  REVERSED: "Reversed",
};

const DOCUMENT_TYPE_LABELS: Record<string, string> = {
  REQUISITION: "Requisition",
  PURCHASE_ORDER: "Purchase Order",
  PAYMENT_VOUCHER: "Payment Voucher",
  GOODS_RECEIVED_NOTE: "GRN",
};

// Helper to convert stored documents to WorkflowDocument
function convertToWorkflowDocument(doc: any): WorkflowDocument {
  console.log("🔄 Converting document:", {
    id: doc.id,
    type: doc.type,
    documentNumber: doc.documentNumber,
    createdAt: doc.createdAt,
    createdAtType: typeof doc.createdAt
  });
  const converted = {
    id: doc.id,
    type: doc.type,
    documentNumber: doc.documentNumber,
    status: doc.status,
    currentStage: doc.currentStage || 1,
    createdBy: doc.createdBy,
    createdByUser: doc.createdByUser,
    createdAt: new Date(doc.createdAt),
    updatedAt: new Date(doc.updatedAt),
    metadata: doc.metadata || {},
  };
  console.log("✅ Converted document createdAt:", converted.createdAt);
  return converted;
}

// Search function that queries local storage
function performSearch(
  filters: SearchFilters,
  page: number,
  limit: number
): { documents: WorkflowDocument[]; total: number; totalPages: number } {
  console.log("🔍 Search starting with filters:", filters);

  // Get all documents from unified storage
  const pos = getPurchaseOrders();
  const reqs = getRequisitions();
  const pvs = getPaymentVouchers();
  const grns = getGoodsReceivedNotes();

  console.log("📦 Storage data:", { pos: pos.length, reqs: reqs.length, pvs: pvs.length, grns: grns.length });

  const allDocs: WorkflowDocument[] = [
    ...pos.map(convertToWorkflowDocument),
    ...reqs.map(convertToWorkflowDocument),
    ...pvs.map(convertToWorkflowDocument),
    ...grns.map(convertToWorkflowDocument),
  ];

  console.log("📄 All documents:", allDocs.length, allDocs);

  // Apply filters
  let filtered = allDocs.filter((doc) => {
    console.log(`\n🔍 Evaluating ${doc.documentNumber}:`);

    // Filter by document number (case-insensitive, partial match)
    if (
      filters.documentNumber &&
      !doc.documentNumber
        .toLowerCase()
        .includes(filters.documentNumber.toLowerCase())
    ) {
      console.log(`  ❌ documentNumber filter: "${filters.documentNumber}" not in "${doc.documentNumber}"`);
      return false;
    }
    if (filters.documentNumber) {
      console.log(`  ✓ documentNumber filter: "${filters.documentNumber}" found in "${doc.documentNumber}"`);
    }

    // Filter by document type
    if (filters.documentType !== "ALL" && doc.type !== filters.documentType) {
      console.log(`  ❌ type filter: doc.type="${doc.type}" !== filters.documentType="${filters.documentType}"`);
      return false;
    }
    if (filters.documentType !== "ALL") {
      console.log(`  ✓ type filter: doc.type="${doc.type}" === filters.documentType="${filters.documentType}"`);
    }

    // Filter by status
    if (filters.status !== "ALL" && doc.status !== filters.status) {
      console.log(`  ❌ status filter: doc.status="${doc.status}" !== filters.status="${filters.status}"`);
      return false;
    }
    if (filters.status !== "ALL") {
      console.log(`  ✓ status filter: doc.status="${doc.status}" === filters.status="${filters.status}"`);
    }

    // Filter by start date
    if (filters.startDate) {
      const startDate = new Date(filters.startDate);
      if (doc.createdAt < startDate) {
        console.log(`  ❌ startDate filter: doc.createdAt="${doc.createdAt.toISOString()}" < startDate="${startDate.toISOString()}"`);
        return false;
      }
      console.log(`  ✓ startDate filter: doc.createdAt="${doc.createdAt.toISOString()}" >= startDate="${startDate.toISOString()}"`);
    }

    // Filter by end date
    if (filters.endDate) {
      const endDate = new Date(filters.endDate);
      endDate.setHours(23, 59, 59, 999); // Include the entire end date
      if (doc.createdAt > endDate) {
        console.log(`  ❌ endDate filter: doc.createdAt="${doc.createdAt.toISOString()}" > endDate="${endDate.toISOString()}"`);
        return false;
      }
      console.log(`  ✓ endDate filter: doc.createdAt="${doc.createdAt.toISOString()}" <= endDate="${endDate.toISOString()}"`);
    }

    console.log("✅ Document passed all filters:", doc.documentNumber);
    return true;
  });

  console.log("🔎 After filtering:", filtered.length, "documents from", allDocs.length);

  // Sort by created date (newest first)
  filtered.sort(
    (a, b) =>
      new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
  );

  const total = filtered.length;
  const totalPages = Math.ceil(total / limit);
  const skip = (page - 1) * limit;
  const paginatedData = filtered.slice(skip, skip + limit);

  return {
    documents: paginatedData,
    total,
    totalPages,
  };
}

export function TransactionResults({
  filters,
  refreshTrigger,
  userRole,
  onSearchComplete,
}: TransactionResultsProps) {
  const router = useRouter();
  const [documents, setDocuments] = useState<WorkflowDocument[]>([]);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalDocuments, setTotalDocuments] = useState(0);

  const pageSize = 10;

  // Mutation for search
  const searchMutation = useMutation({
    mutationFn: () => {
      console.log("🔍 Searching with filters:", filters);
      try {
        const result = performSearch(filters, currentPage, pageSize);
        console.log("✅ Search completed:", result);
        return Promise.resolve(result);
      } catch (error) {
        console.error("❌ Search error:", error);
        return Promise.reject(error);
      }
    },
    onSuccess: (result) => {
      console.log("📊 Setting search results:", result);
      setDocuments(result.documents);
      setTotalDocuments(result.total);
      setTotalPages(result.totalPages);
      onSearchComplete?.();
    },
    onError: (error) => {
      console.error("Failed to search documents:", error);
      setDocuments([]);
      setTotalDocuments(0);
      setTotalPages(1);
      onSearchComplete?.();
    },
  });

  // Trigger search when filters or pagination changes
  React.useEffect(() => {
    console.log("🚀 Starting search effect with:", { filters, page: currentPage });
    searchMutation.mutate();
  }, [filters, currentPage, refreshTrigger]);

  const columns: ColumnDef<WorkflowDocument>[] = [
    {
      accessorKey: "documentNumber",
      header: ({ column }) => (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          className="-ml-3"
        >
          Document #
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      ),
      cell: ({ row }) => (
        <span className="font-medium text-primary">
          {row.getValue("documentNumber")}
        </span>
      ),
    },
    {
      accessorKey: "type",
      header: "Type",
      cell: ({ row }) => {
        const type = row.getValue("type") as string;
        return (
          <span className="text-sm">{DOCUMENT_TYPE_LABELS[type] || type}</span>
        );
      },
    },
    {
      accessorKey: "status",
      header: "Status",
      cell: ({ row }) => {
        const status = row.getValue("status") as string;
        return (
          <Badge variant={STATUS_COLORS[status] as any}>
            {STATUS_LABELS[status] || status}
          </Badge>
        );
      },
    },
    {
      accessorKey: "createdAt",
      header: ({ column }) => (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
          className="-ml-3"
        >
          Created
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      ),
      cell: ({ row }) => {
        const date = new Date(row.getValue("createdAt"));
        return (
          <span className="text-sm text-muted-foreground">
            {date.toLocaleDateString()}
          </span>
        );
      },
    },
  ];

  if (searchMutation.isPending) {
    return (
      <Card>
        <CardContent className="pt-6">
          <div className="space-y-4">
            <div className="text-sm text-muted-foreground mb-4">
              Searching documents...
            </div>
            <TransactionTableSkeleton />
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardContent className="pt-6">
        <div className="space-y-4">
          <DataTable
            columns={columns}
            data={documents}
            hideSearchBar={true}
            renderRowActions={(doc: WorkflowDocument) => {
              // Map document type to URL slug
              const typeSlug =
                {
                  REQUISITION: "requisitions",
                  PURCHASE_ORDER: "purchase-orders",
                  PAYMENT_VOUCHER: "payment-vouchers",
                  GOODS_RECEIVED_NOTE: "grn",
                }[doc.type] || "workflows";

              return (
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => router.push(`/${typeSlug}/${doc.id}`)}
                    className="gap-1"
                  >
                    <Eye className="h-4 w-4" />
                    View
                  </Button>
                  <DownloadButton
                    documentId={doc.id}
                    documentNumber={doc.documentNumber}
                  />
                </div>
              );
            }}
          />

          {/* Custom Pagination */}
          {totalPages > 1 && (
            <div className="flex items-center justify-between">
              <div className="text-sm text-muted-foreground">
                Showing {documents.length > 0 ? (currentPage - 1) * pageSize + 1 : 0} to{" "}
                {Math.min(currentPage * pageSize, totalDocuments)} of {totalDocuments}{" "}
                documents
              </div>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setCurrentPage(Math.max(1, currentPage - 1))}
                  disabled={currentPage === 1 || searchMutation.isPending}
                >
                  Previous
                </Button>
                <span className="text-sm px-3 py-2">
                  Page {currentPage} of {totalPages}
                </span>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() =>
                    setCurrentPage(Math.min(totalPages, currentPage + 1))
                  }
                  disabled={currentPage >= totalPages || searchMutation.isPending}
                >
                  Next
                </Button>
              </div>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
