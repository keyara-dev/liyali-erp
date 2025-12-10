"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  ArrowLeft,
  Package,
  FileText,
  AlertTriangle,
  AlertCircle,
  Plus,
} from "lucide-react";
import { Skeleton } from "@/components/ui/skeleton";
import { PageHeader } from "@/components/base/page-header";
import { GRNItemsMatchingTable } from "./grn-items-matching-table";
import { QualityIssueReportDialog } from "./quality-issue-dialog";
import { useAddQualityIssueMutation } from "@/hooks/use-quality-issue-mutations";
import { Badge } from "@/components";

interface GRNDetailClientProps {
  grnId: string;
  userId: string;
  userRole: string;
}

interface ReceivedItem {
  id: string;
  itemNumber: number;
  description: string;
  poQuantity: number;
  receivedQuantity: number;
  unit: string;
  variance: number;
  damage: number;
  damageNotes?: string;
  condition: "GOOD" | "DAMAGED" | "PARTIAL";
}

interface GoodsReceivedNote {
  id: string;
  grnNumber: string;
  poNumber: string;
  status: "DRAFT" | "SUBMITTED" | "CONFIRMED" | "REJECTED";
  warehouseLocation: string;
  receivedDate: string;
  receivedBy: string;
  approvedBy?: string;
  items: ReceivedItem[];
  qualityIssues: Array<{
    id: string;
    itemId: string;
    description: string;
    severity: "LOW" | "MEDIUM" | "HIGH";
  }>;
  notes?: string;
  currentStage: number;
  stageName: string;
  createdAt: string;
  updatedAt: string;
}

const STAGE_NAMES: Record<number, string> = {
  1: "Warehouse Clerk Receipt",
  2: "Department Manager Confirmation",
};

// Mock data generator
function generateMockGRN(grnId: string): GoodsReceivedNote {
  const currentStage = Math.floor(Math.random() * 2) + 1;

  return {
    id: grnId,
    grnNumber: `GRN-2024-${String(Math.floor(Math.random() * 9000) + 1000).padStart(4, "0")}`,
    poNumber: `PO-2024-${String(Math.floor(Math.random() * 9000) + 1000).padStart(4, "0")}`,
    status: currentStage === 2 ? "SUBMITTED" : "SUBMITTED",
    warehouseLocation: "Warehouse A - Section 3",
    receivedDate: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(),
    receivedBy: "WAREHOUSE-USER-001",
    approvedBy: currentStage === 2 ? "DEPT-MANAGER-001" : undefined,
    items: [
      {
        id: "item-1",
        itemNumber: 1,
        description: "Office Chairs - Ergonomic",
        poQuantity: 10,
        receivedQuantity: 10,
        unit: "units",
        variance: 0,
        damage: 0,
        condition: "GOOD",
      },
      {
        id: "item-2",
        itemNumber: 2,
        description: "Standing Desks - Electric",
        poQuantity: 5,
        receivedQuantity: 4,
        unit: "units",
        variance: -1,
        damage: 1,
        damageNotes: "One unit arrived with damaged motor",
        condition: "DAMAGED",
      },
      {
        id: "item-3",
        itemNumber: 3,
        description: "Computer Monitors - 27 inch",
        poQuantity: 8,
        receivedQuantity: 8,
        unit: "units",
        variance: 0,
        damage: 0,
        condition: "GOOD",
      },
    ],
    qualityIssues: [
      {
        id: "issue-1",
        itemId: "item-2",
        description: "Standing Desk motor malfunction",
        severity: "HIGH",
      },
    ],
    notes: "General inspection completed. One standing desk has motor issues.",
    currentStage,
    stageName: STAGE_NAMES[currentStage],
    createdAt: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(),
    updatedAt: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000).toISOString(),
  };
}

export function GRNDetailClient({
  grnId,
  userId: _userId,
  userRole: _userRole,
}: GRNDetailClientProps) {
  const router = useRouter();
  const [grn, setGRN] = useState<GoodsReceivedNote | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isQualityDialogOpen, setIsQualityDialogOpen] = useState(false);

  // Mutation for adding quality issues
  const addQualityIssueMutation = useAddQualityIssueMutation(grnId);

  useEffect(() => {
    // Simulate data loading
    const timer = setTimeout(() => {
      setGRN(generateMockGRN(grnId));
      setIsLoading(false);
    }, 500);

    return () => clearTimeout(timer);
  }, [grnId]);

  const handleConfirm = () => {
    toast.success("Navigating to confirmation...");
    router.push(`/grn/${grnId}/confirmation`);
  };

  const handleBack = () => {
    router.back();
  };

  const handleAddQualityIssue = async (issue: Omit<{
    id: string;
    itemId: string;
    description: string;
    severity: "LOW" | "MEDIUM" | "HIGH";
  }, 'id'>) => {
    try {
      // Call mutation to save to localStorage via server action
      const updatedGRN = await addQualityIssueMutation.mutateAsync(issue);

      // Update local state with the response from server
      setGRN(updatedGRN);

      toast.success("Quality issue reported and saved");
    } catch (error) {
      console.error("Error saving quality issue:", error);
      toast.error("Failed to save quality issue");
    }
  };

  if (isLoading || !grn) {
    return (
      <div className="space-y-6">
        <Button variant="ghost" size="sm" onClick={handleBack}>
          <ArrowLeft className="h-4 w-4 mr-2" />
          Back
        </Button>
        <div className="space-y-4">
          <Skeleton className="h-12 w-48" />
          <Skeleton className="h-96 w-full" />
          <Skeleton className="h-96 w-full" />
        </div>
      </div>
    );
  }

  const hasQualityIssues = grn.qualityIssues.length > 0;
  const hasVariances = grn.items.some((item) => item.variance !== 0);

  return (
    <div className="space-y-6">
      {/* Header */}
      <PageHeader
        title={grn.grnNumber}
        subtitle="Goods Received Note"
        badges={[
          {
            status: grn.status,
            type: "document",
          },
        ]}
        onBackClick={handleBack}
        showBackButton={true}
      />

      {/* Status and Stage Info */}
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">Current Stage</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-lg font-semibold">{grn.stageName}</div>
            <p className="text-xs text-muted-foreground mt-1">
              Stage {grn.currentStage} of 2
            </p>
            <div className="flex gap-1 mt-3">
              {[1, 2].map((stage) => (
                <div
                  key={stage}
                  className={`h-2 flex-1 rounded-full ${
                    stage <= grn.currentStage ? "bg-blue-600" : "bg-gray-200"
                  }`}
                />
              ))}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium">Total Items</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{grn.items.length}</div>
            <p className="text-xs text-muted-foreground mt-1">
              {grn.items.filter((i) => i.condition === "GOOD").length} in good
              condition
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Alerts for Issues */}
      {(hasQualityIssues || hasVariances) && (
        <div className="space-y-2">
          {hasQualityIssues && (
            <Card className="border-yellow-200 bg-yellow-50">
              <CardContent className="pt-4 flex gap-3">
                <AlertTriangle className="h-5 w-5 text-yellow-600 shrink-0 mt-0.5" />
                <div>
                  <p className="font-semibold text-yellow-900">
                    Quality Issues Detected
                  </p>
                  <p className="text-sm text-yellow-800">
                    {grn.qualityIssues.length} issue(s) reported during
                    inspection
                  </p>
                </div>
              </CardContent>
            </Card>
          )}
          {hasVariances && (
            <Card className="border-orange-200 bg-orange-50">
              <CardContent className="pt-4 flex gap-3">
                <AlertCircle className="h-5 w-5 text-orange-600 shrink-0 mt-0.5" />
                <div>
                  <p className="font-semibold text-orange-900">
                    Quantity Variances
                  </p>
                  <p className="text-sm text-orange-800">
                    Some items received differ from PO quantities
                  </p>
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      )}

      {/* GRN Information */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5" />
            GRN Information
          </CardTitle>
        </CardHeader>
        <CardContent className="grid gap-4 md:grid-cols-3">
          <div>
            <p className="text-sm text-muted-foreground">PO Number</p>
            <p className="font-semibold">{grn.poNumber}</p>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">Warehouse Location</p>
            <p className="font-semibold">{grn.warehouseLocation}</p>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">Received Date</p>
            <p className="font-semibold">
              {new Date(grn.receivedDate).toLocaleDateString()}
            </p>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">Received By</p>
            <p className="font-semibold">{grn.receivedBy}</p>
          </div>
          {grn.approvedBy && (
            <div>
              <p className="text-sm text-muted-foreground">Approved By</p>
              <p className="font-semibold">{grn.approvedBy}</p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Items Matching */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Package className="h-5 w-5" />
            Items Received vs. Ordered
          </CardTitle>
        </CardHeader>
        <CardContent>
          <GRNItemsMatchingTable items={grn.items} />
        </CardContent>
      </Card>

      {/* Quality Issues */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="flex items-center gap-2">
              <AlertTriangle className="h-5 w-5" />
              Quality Issues Reported
            </CardTitle>
            <Button
              size="sm"
              variant="outline"
              onClick={() => setIsQualityDialogOpen(true)}
              className="gap-2"
            >
              <Plus className="h-4 w-4" />
              Report Issue
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {hasQualityIssues ? (
            <div className="space-y-3">
              {grn.qualityIssues.map((issue) => {
                const item = grn.items.find((i) => i.id === issue.itemId);
                const severityColors = {
                  LOW: "bg-yellow-100 text-yellow-800 border-yellow-200",
                  MEDIUM: "bg-orange-100 text-orange-800 border-orange-200",
                  HIGH: "bg-red-100 text-red-800 border-red-200",
                };
                return (
                  <div
                    key={issue.id}
                    className={`p-4 border rounded-lg ${severityColors[issue.severity]}`}
                  >
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <p className="font-semibold">{item?.description}</p>
                        <p className="text-sm mt-1">{issue.description}</p>
                      </div>
                      <Badge variant="outline" className="ml-2">
                        {issue.severity}
                      </Badge>
                    </div>
                  </div>
                );
              })}
            </div>
          ) : (
            <div className="text-center py-6">
              <p className="text-sm text-muted-foreground">
                No quality issues reported yet
              </p>
              <p className="text-xs text-muted-foreground mt-1">
                Click "Report Issue" to add quality concerns during inspection
              </p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Notes */}
      {grn.notes && (
        <Card>
          <CardHeader>
            <CardTitle>Notes</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm whitespace-pre-wrap">{grn.notes}</p>
          </CardContent>
        </Card>
      )}

      {/* Action Buttons */}
      <div className="flex gap-4 pt-4">
        <Button variant="outline" onClick={handleBack}>
          Cancel
        </Button>
        {grn.status === "SUBMITTED" && (
          <Button
            onClick={handleConfirm}
            className="bg-blue-600 hover:bg-blue-700"
          >
            Confirm Receipt
          </Button>
        )}
      </div>

      {/* Quality Issue Report Dialog */}
      <QualityIssueReportDialog
        open={isQualityDialogOpen}
        onOpenChange={setIsQualityDialogOpen}
        items={grn.items}
        onAddIssue={handleAddQualityIssue}
      />
    </div>
  );
}
