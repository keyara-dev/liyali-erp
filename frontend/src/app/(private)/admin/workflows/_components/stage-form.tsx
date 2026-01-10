"use client";

import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Checkbox } from "@/components/ui/checkbox";
import type { WorkflowStage } from "@/types/workflow-config";

interface StageFormProps {
  stage?: WorkflowStage | null;
  onSave: (stage: WorkflowStage) => void;
  onCancel: () => void;
  errors: Record<string, string>;
}

const APPROVER_ROLES = [
  { id: "DEPARTMENT_MANAGER", name: "Department Manager" },
  { id: "FINANCE_OFFICER", name: "Finance Officer" },
  { id: "CFO", name: "CFO" },
  { id: "WAREHOUSE_MANAGER", name: "Warehouse Manager" },
  { id: "PROCUREMENT_OFFICER", name: "Procurement Officer" },
  { id: "ADMIN", name: "Admin" },
];

export function StageForm({ stage, onSave, onCancel, errors }: StageFormProps) {
  console.log("StageForm received stage:", stage);

  const [formData, setFormData] = useState<WorkflowStage>(() => {
    if (stage) {
      console.log("Initializing form with existing stage data:", stage);
      // Use existing stage data for editing
      const initialData = {
        id: stage.id || "",
        order: stage.order || stage.stageNumber || 1,
        name: stage.name || stage.stageName || "",
        description: stage.description || "",
        approverRole: stage.approverRole || stage.requiredRole || "",
        requiredApprovals: stage.requiredApprovals || 1,
        canReject: stage.canReject !== undefined ? stage.canReject : true,
        canReassign: stage.canReassign !== undefined ? stage.canReassign : true,
        stageNumber: stage.stageNumber || stage.order || 1,
        stageName: stage.stageName || stage.name || "",
        requiredRole: stage.requiredRole || stage.approverRole || "",
        canBeReassigned:
          stage.canBeReassigned !== undefined ? stage.canBeReassigned : true,
      };

      console.log("Initialized form data:", initialData);
      return initialData;
    } else {
      console.log("Initializing form with default data for new stage");
      // Default data for new stage
      return {
        id: "",
        order: 1,
        name: "",
        description: "",
        approverRole: "" as any,
        requiredApprovals: 1,
        canReject: true,
        canReassign: true,
        stageNumber: 1,
        stageName: "",
        requiredRole: "",
        canBeReassigned: true,
      };
    }
  });

  // Update form data when stage prop changes (important for editing)
  useEffect(() => {
    if (stage) {
      console.log("Stage prop changed, updating form data:", stage);
      const updatedData = {
        id: stage.id || "",
        order: stage.order || stage.stageNumber || 1,
        name: stage.name || stage.stageName || "",
        description: stage.description || "",
        approverRole: stage.approverRole || stage.requiredRole || "",
        requiredApprovals: stage.requiredApprovals || 1,
        canReject: stage.canReject !== undefined ? stage.canReject : true,
        canReassign: stage.canReassign !== undefined ? stage.canReassign : true,
        stageNumber: stage.stageNumber || stage.order || 1,
        stageName: stage.stageName || stage.name || "",
        requiredRole: stage.requiredRole || stage.approverRole || "",
        canBeReassigned:
          stage.canBeReassigned !== undefined ? stage.canBeReassigned : true,
      };

      console.log("Setting updated form data:", updatedData);
      setFormData(updatedData);
    }
  }, [stage]);

  const handleChange = (key: keyof WorkflowStage, value: any) => {
    setFormData((prev) => {
      const updated = {
        ...prev,
        [key]: value,
      };

      // Ensure consistency between name/stageName and approverRole/requiredRole
      if (key === "name") {
        updated.stageName = value;
      } else if (key === "stageName") {
        updated.name = value;
      } else if (key === "approverRole") {
        updated.requiredRole = value;
      } else if (key === "requiredRole") {
        updated.approverRole = value;
      }

      return updated;
    });
  };

  const handleSubmit = () => {
    onSave(formData);
  };

  return (
    <div className="space-y-6">
      {/* Stage Name */}
      <div className="space-y-2">
        <label className="text-sm font-medium">
          Stage Name <span className="text-destructive">*</span>
        </label>
        <Input
          placeholder="e.g., Department Manager Review"
          value={formData.name}
          onChange={(e) => handleChange("name", e.target.value)}
          className={errors.name ? "border-destructive" : ""}
        />
        {errors.name && (
          <p className="text-sm text-destructive">{errors.name}</p>
        )}
      </div>

      {/* Description */}
      <div className="space-y-2">
        <label className="text-sm font-medium">Description</label>
        <Textarea
          placeholder="Describe what this stage is responsible for..."
          value={formData.description}
          onChange={(e) => handleChange("description", e.target.value)}
          rows={2}
        />
      </div>

      {/* Approver Role */}
      <div className="space-y-2">
        <label className="text-sm font-medium">
          Approver Role <span className="text-destructive">*</span>
        </label>
        <Select
          value={formData.approverRole}
          onValueChange={(value) => handleChange("approverRole", value)}
        >
          <SelectTrigger
            className={errors.approverRole ? "border-destructive" : ""}
          >
            <SelectValue placeholder="Select approver role" />
          </SelectTrigger>
          <SelectContent>
            {APPROVER_ROLES.map((role) => (
              <SelectItem key={role.id} value={role.id}>
                {role.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        {errors.approverRole && (
          <p className="text-sm text-destructive">{errors.approverRole}</p>
        )}
      </div>

      {/* Required Approvals */}
      <div className="space-y-2">
        <label className="text-sm font-medium">
          Required Approvals <span className="text-destructive">*</span>
        </label>
        <Select
          value={String(formData.requiredApprovals)}
          onValueChange={(value) =>
            handleChange("requiredApprovals", parseInt(value))
          }
        >
          <SelectTrigger>
            <SelectValue placeholder="Select number of approvals" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="1">1 Approval</SelectItem>
            <SelectItem value="2">2 Approvals</SelectItem>
            <SelectItem value="3">3 Approvals</SelectItem>
            <SelectItem value="5">All Approvals</SelectItem>
          </SelectContent>
        </Select>
        {errors.requiredApprovals && (
          <p className="text-sm text-destructive">{errors.requiredApprovals}</p>
        )}
      </div>

      {/* Permissions */}
      <div className="space-y-3 border-t pt-4">
        <label className="text-sm font-medium">Stage Permissions</label>

        <div className="flex items-center gap-3">
          <Checkbox
            id="canReject"
            checked={formData.canReject}
            onCheckedChange={(checked) => handleChange("canReject", checked)}
          />
          <label htmlFor="canReject" className="text-sm cursor-pointer">
            Approvers can reject documents
          </label>
        </div>

        <div className="flex items-center gap-3">
          <Checkbox
            id="canReassign"
            checked={formData.canReassign}
            onCheckedChange={(checked) => handleChange("canReassign", checked)}
          />
          <label htmlFor="canReassign" className="text-sm cursor-pointer">
            Approvers can reassign to others
          </label>
        </div>
      </div>

      {/* Action Buttons */}
      <div className="flex gap-3 justify-end border-t pt-4">
        <Button variant="outline" onClick={onCancel}>
          Cancel
        </Button>
        <Button onClick={handleSubmit}>
          {stage ? "Update Stage" : "Add Stage"}
        </Button>
      </div>
    </div>
  );
}
