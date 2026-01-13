"use client";

import { useState } from "react";
import { useOrganizationContext } from "@/hooks/use-organization";
import {
  useUpdateOrganizationMutation,
  useDeleteOrganizationMutation,
} from "@/hooks/use-organization-mutations";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Loader2, Trash2, Save, Building2 } from "lucide-react";
import { toast } from "sonner";

export function WorkspaceSettings() {
  const { currentOrganization } = useOrganizationContext();
  const { updateOrganization, isPending: isUpdating } =
    useUpdateOrganizationMutation();
  const { deleteOrganization, isPending: isDeleting } =
    useDeleteOrganizationMutation();

  const [formData, setFormData] = useState({
    name: currentOrganization?.name || "",
    description: currentOrganization?.description || "",
  });

  const [hasChanges, setHasChanges] = useState(false);

  const handleInputChange = (field: string, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    setHasChanges(true);
  };

  const handleUpdateWorkspace = async () => {
    if (!currentOrganization) {
      toast.error("No workspace selected");
      return;
    }

    if (!formData.name.trim()) {
      toast.error("Workspace name is required");
      return;
    }

    try {
      await updateOrganization({
        id: currentOrganization.id,
        name: formData.name.trim(),
        description: formData.description.trim(),
      });
      setHasChanges(false);
    } catch (error) {
      console.error("Failed to update workspace:", error);
    }
  };

  const handleDeleteWorkspace = async () => {
    if (!currentOrganization) {
      toast.error("No workspace selected");
      return;
    }

    try {
      await deleteOrganization(currentOrganization.id);
    } catch (error) {
      console.error("Failed to delete workspace:", error);
    }
  };

  if (!currentOrganization) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-8">
          <div className="text-center">
            <Building2 className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
            <p className="text-muted-foreground">No workspace selected</p>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      {/* Workspace Details */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Building2 className="h-5 w-5" />
            Workspace Details
          </CardTitle>
          <CardDescription>
            Update your workspace name and description
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="workspace-name">Workspace Name</Label>
            <Input
              id="workspace-name"
              value={formData.name}
              onChange={(e) => handleInputChange("name", e.target.value)}
              placeholder="Enter workspace name"
              disabled={isUpdating}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="workspace-description">Description</Label>
            <Textarea
              id="workspace-description"
              value={formData.description}
              onChange={(e) => handleInputChange("description", e.target.value)}
              placeholder="Enter workspace description (optional)"
              rows={3}
              disabled={isUpdating}
            />
          </div>

          <div className="flex justify-end">
            <Button
              onClick={handleUpdateWorkspace}
              disabled={!hasChanges || isUpdating || !formData.name.trim()}
              className="flex items-center gap-2"
            >
              {isUpdating ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <Save className="h-4 w-4" />
              )}
              Save Changes
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Workspace Information */}
      <Card>
        <CardHeader>
          <CardTitle>Workspace Information</CardTitle>
          <CardDescription>
            Read-only information about your workspace
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <Label className="text-sm font-medium text-muted-foreground">
                Workspace ID
              </Label>
              <p className="text-sm font-mono bg-muted px-2 py-1 rounded mt-1">
                {currentOrganization.id}
              </p>
            </div>
            <div>
              <Label className="text-sm font-medium text-muted-foreground">
                Slug
              </Label>
              <p className="text-sm font-mono bg-muted px-2 py-1 rounded mt-1">
                {currentOrganization.slug}
              </p>
            </div>
            <div>
              <Label className="text-sm font-medium text-muted-foreground">
                Tier
              </Label>
              <p className="text-sm capitalize bg-muted px-2 py-1 rounded mt-1">
                {currentOrganization.tier}
              </p>
            </div>
            <div>
              <Label className="text-sm font-medium text-muted-foreground">
                Created
              </Label>
              <p className="text-sm bg-muted px-2 py-1 rounded mt-1">
                {new Date(currentOrganization.createdAt).toLocaleDateString()}
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Danger Zone */}
      <Card className="border-destructive/20">
        <CardHeader>
          <CardTitle className="text-destructive">Danger Zone</CardTitle>
          <CardDescription>
            Irreversible actions that will affect your workspace
          </CardDescription>
        </CardHeader>
        <CardContent>
          <AlertDialog>
            <AlertDialogTrigger asChild>
              <Button
                variant="destructive"
                disabled={isDeleting}
                className="flex items-center gap-2"
              >
                {isDeleting ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <Trash2 className="h-4 w-4" />
                )}
                Delete Workspace
              </Button>
            </AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
                <AlertDialogDescription className="space-y-2">
                  <p>
                    This action will permanently delete the workspace{" "}
                    <strong>"{currentOrganization.name}"</strong> and all
                    associated data.
                  </p>
                  <p className="text-sm text-muted-foreground">
                    This includes all workflows, requests, users, and settings.
                    This action cannot be undone.
                  </p>
                  <p className="text-sm font-medium">
                    You will be redirected to the workspace selection screen
                    after deletion.
                  </p>
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel disabled={isDeleting}>
                  Cancel
                </AlertDialogCancel>
                <AlertDialogAction
                  onClick={handleDeleteWorkspace}
                  disabled={isDeleting}
                  className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                >
                  {isDeleting ? (
                    <>
                      <Loader2 className="h-4 w-4 animate-spin mr-2" />
                      Deleting...
                    </>
                  ) : (
                    "Delete Workspace"
                  )}
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        </CardContent>
      </Card>
    </div>
  );
}
