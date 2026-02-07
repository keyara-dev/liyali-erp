"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { CheckCircle, XCircle, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { bulkUpdateRoles } from "@/app/_actions/roles";

interface RoleBulkActionsProps {
  selectedRoles: string[];
  onActionComplete: () => void;
}

export function RoleBulkActions({
  selectedRoles,
  onActionComplete,
}: RoleBulkActionsProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [action, setAction] = useState("");

  const handleBulkAction = async () => {
    if (!action || selectedRoles.length === 0) {
      toast.error("Please select an action");
      return;
    }

    setIsLoading(true);

    try {
      let updates = {};

      switch (action) {
        case "activate":
          updates = { is_active: true };
          break;
        case "deactivate":
          updates = { is_active: false };
          break;
        default:
          toast.error("Invalid action selected");
          return;
      }

      const result = await bulkUpdateRoles(selectedRoles, updates);

      if (result.success) {
        toast.success(`Successfully updated ${selectedRoles.length} roles`);
        onActionComplete();
      } else {
        toast.error("Failed to update roles");
      }
    } catch (error) {
      console.error("Error performing bulk action:", error);
      toast.error("Failed to perform bulk action");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">Bulk Actions</CardTitle>
        <CardDescription>
          Perform actions on {selectedRoles.length} selected roles
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <Badge variant="secondary">{selectedRoles.length} selected</Badge>
          </div>

          <Select value={action} onValueChange={setAction}>
            <SelectTrigger className="w-48">
              <SelectValue placeholder="Select action" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="activate">
                <div className="flex items-center gap-2">
                  <CheckCircle className="h-4 w-4 text-green-600" />
                  Activate Roles
                </div>
              </SelectItem>
              <SelectItem value="deactivate">
                <div className="flex items-center gap-2">
                  <XCircle className="h-4 w-4 text-red-600" />
                  Deactivate Roles
                </div>
              </SelectItem>
            </SelectContent>
          </Select>

          <Button onClick={handleBulkAction} disabled={!action || isLoading}>
            {isLoading ? "Processing..." : "Apply Action"}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
