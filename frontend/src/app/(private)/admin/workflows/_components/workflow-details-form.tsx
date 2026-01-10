"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { SelectField } from "@/components/ui/select-field";
import { Checkbox } from "@/components/ui/checkbox";
import type { WorkflowFormData } from "@/types/workflow-config";

interface WorkflowDetailsFormProps {
  data: WorkflowFormData;
  onChange: (key: keyof WorkflowFormData, value: any) => void;
  errors: Record<string, string>;
}

const DOCUMENT_TYPES = [
  { id: "REQUISITION", name: "Requisition" },
  { id: "PURCHASE_ORDER", name: "Purchase Order" },
  { id: "PAYMENT_VOUCHER", name: "Payment Voucher" },
  { id: "GOODS_RECEIVED_NOTE", name: "Goods Received Note" },
  { id: "BUDGET", name: "Budget" },
];

export function WorkflowDetailsForm({
  data,
  onChange,
  errors,
}: WorkflowDetailsFormProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Workflow Details</CardTitle>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Name */}
        <div className="flex flex-col md:flex-row gap-4 items-center">
          <Input
            label="Workflow Name"
            required
            placeholder="e.g., Standard Requisition Approval"
            value={data.name}
            onChange={(e) => onChange("name", e.target.value)}
            className={errors.name ? "border-destructive" : ""}
            isInvalid={!!errors.name}
            errorText={errors.name}
          />

          {/* Document Type */}
          <SelectField
            label="Workflow Applies To"
            required
            placeholder="Select document type"
            value={data.entityType || data.documentType}
            onValueChange={(value) => {
              onChange("entityType", value);
              onChange("documentType", value); // Keep both for compatibility
            }}
            options={DOCUMENT_TYPES}
            isInvalid={!!errors.entityType || !!errors.documentType}
            errorText={errors.entityType || errors.documentType}
          />
        </div>

        {/* Description */}
        <Textarea
          label="Description"
          placeholder="Describe the purpose and use case for this workflow..."
          value={data.description}
          onChange={(e) => onChange("description", e.target.value)}
          rows={3}
        />

        {/* Set as Default */}
        <div className="flex items-center gap-3 p-4 border rounded-lg bg-muted/30">
          <Checkbox
            id="isDefault"
            checked={data.isDefault}
            onCheckedChange={(checked) => onChange("isDefault", checked)}
          />
          <label
            htmlFor="isDefault"
            className="text-sm font-medium cursor-pointer"
          >
            Set as default workflow for{" "}
            {data.entityType || data.documentType
              ? DOCUMENT_TYPES.find(
                  (t) => t.id === (data.entityType || data.documentType)
                )?.name
              : "selected document type"}
          </label>
        </div>
      </CardContent>
    </Card>
  );
}
