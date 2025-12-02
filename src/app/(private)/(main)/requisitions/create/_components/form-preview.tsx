"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Loader2 } from "lucide-react";
import { createWorkflowDocument } from "@/app/_actions/workflow";
import { RequisitionFormData } from "./create-requisition-client";

interface FormPreviewProps {
  formData: RequisitionFormData;
  onBack: () => void;
  onSubmit: (data: RequisitionFormData) => void;
}

export function FormPreview({ formData, onBack, onSubmit }: FormPreviewProps) {
  const router = useRouter();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const totalAmount = formData.items.reduce(
    (sum, item) => sum + item.estimatedCost * item.quantity,
    0
  );

  const handleSubmit = async () => {
    setIsSubmitting(true);
    setError(null);

    try {
      const result = await createWorkflowDocument("REQUISITION", {
        department: formData.department,
        requestedFor: formData.requestedFor,
        items: formData.items,
        justification: formData.justification,
        budgetCode: formData.budgetCode,
      });

      if (result.success) {
        // Redirect to requisitions list
        router.push("//requisitions");
      } else {
        setError(result.message || "Failed to create requisition");
      }
    } catch (err) {
      console.error("Submit error:", err);
      setError("An error occurred while submitting the requisition");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="space-y-6">
      {error && (
        <div className="bg-destructive/10 border border-destructive/30 text-destructive p-4 rounded-lg">
          {error}
        </div>
      )}

      {/* Requisition Details */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Requisition Details</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <p className="text-sm text-muted-foreground">Department</p>
              <p className="text-base font-medium">{formData.department}</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Requested For</p>
              <p className="text-base font-medium">{formData.requestedFor}</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Budget Code</p>
              <p className="text-base font-medium">{formData.budgetCode}</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Total Amount</p>
              <p className="text-base font-bold text-primary">
                K{totalAmount.toFixed(2)}
              </p>
            </div>
          </div>

          {/* Justification */}
          <div className="mt-4">
            <p className="text-sm text-muted-foreground">Justification</p>
            <p className="text-sm mt-1 p-3 bg-muted rounded-md">
              {formData.justification}
            </p>
          </div>
        </CardContent>
      </Card>

      {/* Items Summary */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Items Summary</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="rounded-md border overflow-hidden">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Description</TableHead>
                  <TableHead className="text-right">Qty</TableHead>
                  <TableHead className="text-right">Unit Cost (K)</TableHead>
                  <TableHead className="text-right">Total (K)</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {formData.items.map((item, index) => (
                  <TableRow key={item.id}>
                    <TableCell className="font-medium">
                      {item.itemDescription}
                    </TableCell>
                    <TableCell className="text-right">
                      {item.quantity}
                    </TableCell>
                    <TableCell className="text-right">
                      {item.estimatedCost.toFixed(2)}
                    </TableCell>
                    <TableCell className="text-right font-semibold">
                      {(item.quantity * item.estimatedCost).toFixed(2)}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>

          {/* Total */}
          <div className="flex justify-end mt-4 pt-4 border-t">
            <div className="space-y-2">
              <div className="flex gap-8">
                <span className="font-medium">Grand Total:</span>
                <span className="font-bold text-lg text-primary">
                  K{totalAmount.toFixed(2)}
                </span>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Action Buttons */}
      <div className="flex gap-3 justify-end">
        <Button
          type="button"
          variant="outline"
          onClick={onBack}
          disabled={isSubmitting}
        >
          Back to Edit
        </Button>
        <Button
          type="button"
          onClick={handleSubmit}
          disabled={isSubmitting}
          className="gap-2"
        >
          {isSubmitting && <Loader2 className="h-4 w-4 animate-spin" />}
          {isSubmitting ? "Submitting..." : "Submit Requisition"}
        </Button>
      </div>

      {/* Info */}
      <div className="bg-primary/5 border border-primary/20 rounded-lg p-4">
        <p className="text-sm">
          <span className="font-medium">Note:</span> Once submitted, your
          requisition will enter the approval workflow. You will be notified
          when it progresses to the next stage.
        </p>
      </div>
    </div>
  );
}
