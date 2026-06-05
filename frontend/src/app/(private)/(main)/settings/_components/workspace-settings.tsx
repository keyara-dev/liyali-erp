"use client";

import { useState, useEffect } from "react";
import { useOrganizationContext } from "@/hooks/use-organization";
import {
  useUpdateOrganizationMutation,
  useDeleteOrganizationMutation,
  useUpdateSettingsMutation,
} from "@/hooks/use-organization-mutations";
import { useOrganizationSettingsQuery } from "@/hooks/use-organization-queries";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { OrganizationLogoUpload } from "@/components/ui/organization-logo-upload";
import { Separator } from "@/components/ui/separator";
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
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Switch } from "@/components/ui/switch";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Loader2,
  Trash2,
  Save,
  Building2,
  FileText,
  ArrowDownUp,
  Stamp,
  Upload,
  X,
  Zap,
} from "lucide-react";
import { toast } from "sonner";
import { uploadToImageKit, validateImageFile } from "@/lib/imagekit";

export function WorkspaceSettings() {
  const { currentOrganization } = useOrganizationContext();
  const { updateOrganization, isPending: isUpdating } =
    useUpdateOrganizationMutation();
  const { deleteOrganization, isPending: isDeleting } =
    useDeleteOrganizationMutation();
  const { updateSettings, isPending: isSavingSettings } =
    useUpdateSettingsMutation();
  const { data: settingsData } = useOrganizationSettingsQuery();

  const [formData, setFormData] = useState({
    name: currentOrganization?.name || "",
    description: currentOrganization?.description || "",
    logoUrl: currentOrganization?.logoUrl || "",
    tagline: currentOrganization?.tagline || "",
  });

  const [hasChanges, setHasChanges] = useState(false);

  const [procurementFlow, setProcurementFlow] = useState<
    "goods_first" | "payment_first"
  >("goods_first");

  // "Stamp of Issuing Officer" — rendered as a fallback in every GRN PDF
  // when the certifying officer does not upload a per-GRN stamp.
  const [stampImageUrl, setStampImageUrl] = useState<string>("");
  const [stampUploading, setStampUploading] = useState(false);
  const [stampHasChanges, setStampHasChanges] = useState(false);

  // Procurement automation state
  const [autoCreateGRNFromPO, setAutoCreateGRNFromPO] = useState(false);
  const [autoCreatePVFromPO, setAutoCreatePVFromPO] = useState(false);
  const [autoCreatePVFromGRN, setAutoCreatePVFromGRN] = useState(false);
  const [pvAutomationLevel, setPvAutomationLevel] = useState<
    "manual" | "auto_submit" | "auto_approve"
  >("manual");
  const [autoApproveMaxAmount, setAutoApproveMaxAmount] = useState<
    number | ""
  >("");
  const [automationHasChanges, setAutomationHasChanges] = useState(false);

  // Sync form data when currentOrganization changes
  useEffect(() => {
    if (currentOrganization) {
      setFormData({
        name: currentOrganization.name || "",
        description: currentOrganization.description || "",
        logoUrl: currentOrganization.logoUrl || "",
        tagline: currentOrganization.tagline || "",
      });
      setHasChanges(false);
    }
  }, [currentOrganization]);

  // Sync procurement flow when settings load
  useEffect(() => {
    if (settingsData?.procurementFlow) {
      setProcurementFlow(settingsData.procurementFlow);
    }
  }, [settingsData]);

  // Sync stamp image from settings on load (and on remote refresh)
  useEffect(() => {
    setStampImageUrl(settingsData?.stampImageUrl ?? "");
    setStampHasChanges(false);
  }, [settingsData?.stampImageUrl]);

  // Sync procurement automation settings on load
  useEffect(() => {
    if (!settingsData) return;
    setAutoCreateGRNFromPO(settingsData.autoCreateGRNFromPO ?? false);
    setAutoCreatePVFromPO(settingsData.autoCreatePVFromPO ?? false);
    setAutoCreatePVFromGRN(settingsData.autoCreatePVFromGRN ?? false);
    setPvAutomationLevel(settingsData.pvAutomationLevel ?? "manual");
    setAutoApproveMaxAmount(settingsData.autoApproveMaxAmount ?? "");
    setAutomationHasChanges(false);
  }, [
    settingsData?.autoCreateGRNFromPO,
    settingsData?.autoCreatePVFromPO,
    settingsData?.autoCreatePVFromGRN,
    settingsData?.pvAutomationLevel,
    settingsData?.autoApproveMaxAmount,
  ]);

  const handleInputChange = (field: string, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    setHasChanges(true);
  };

  const handleLogoChange = (url: string) => {
    setFormData((prev) => ({ ...prev, logoUrl: url }));
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
        logoUrl: formData.logoUrl,
        tagline: formData.tagline.trim(),
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

  const handleSaveProcurementFlow = async () => {
    if (!settingsData) return;
    try {
      await updateSettings({ ...settingsData, procurementFlow });
    } catch (error) {
      console.error("Failed to update procurement flow:", error);
    }
  };

  const handleSaveAutomation = async () => {
    if (!settingsData) return;
    try {
      await updateSettings({
        ...settingsData,
        autoCreateGRNFromPO,
        autoCreatePVFromPO,
        autoCreatePVFromGRN,
        pvAutomationLevel,
        autoApproveMaxAmount:
          autoApproveMaxAmount === "" ? undefined : autoApproveMaxAmount,
      });
      setAutomationHasChanges(false);
      toast.success("Automation settings saved");
    } catch (error) {
      console.error("Failed to update automation settings:", error);
    }
  };

  const handleStampFile = async (file: File | null) => {
    if (!file) return;
    const validation = validateImageFile(file);
    if (!validation.valid) {
      toast.error(validation.error ?? "Invalid image");
      return;
    }
    setStampUploading(true);
    try {
      const response = await uploadToImageKit(file, "organizations/stamps");
      setStampImageUrl(response.url);
      setStampHasChanges(true);
      toast.success("Stamp uploaded — remember to save");
    } catch (error: any) {
      console.error("Stamp upload failed:", error);
      toast.error(error.message || "Failed to upload stamp");
    } finally {
      setStampUploading(false);
    }
  };

  const handleClearStamp = () => {
    setStampImageUrl("");
    setStampHasChanges(true);
  };

  const handleSaveStamp = async () => {
    if (!settingsData) return;
    try {
      await updateSettings({ ...settingsData, stampImageUrl });
      setStampHasChanges(false);
    } catch (error) {
      console.error("Failed to update stamp:", error);
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
      {/* Workspace Settings — Logo, Document Header, Details */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Building2 className="h-5 w-5" />
            Workspace Settings
          </CardTitle>
          <CardDescription>
            Manage your workspace branding, document header, and details
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Responsive two-column layout on lg+ */}
          <div className="grid grid-cols-1 lg:grid-cols-2 lg:gap-8">
            {/* Left — Logo */}
            <div className="space-y-3">
              <div>
                <p className="text-sm font-medium">Workspace Logo</p>
                <p className="text-sm text-muted-foreground">
                  Upload a logo to represent your workspace
                </p>
              </div>
              <OrganizationLogoUpload
                currentLogoUrl={formData.logoUrl}
                organizationName={formData.name || "Workspace"}
                onLogoChange={handleLogoChange}
                disabled={isUpdating}
                size="lg"
              />{" "}
              {/* Header Preview */}
              <div className="space-y-1.5 mt-8">
                <Label className="text-sm font-medium">Preview</Label>
                <div className="border rounded-md p-4 bg-card flex flex-row items-center gap-3">
                  {formData.logoUrl ? (
                    // eslint-disable-next-line @next/next/no-img-element
                    <div className="w-14 h-14 rounded-xl overflow-clip">
                      <img
                        src={formData.logoUrl}
                        alt="Logo"
                        className="w-full h-full object-contain shrink-0"
                      />
                    </div>
                  ) : (
                    <div className="w-10 h-10 rounded-full bg-muted flex items-center justify-center shrink-0">
                      <Building2 className="h-5 w-5 text-muted-foreground" />
                    </div>
                  )}
                  <div className="text-center">
                    <p className="text-sm font-bold leading-tight">
                      {formData.name || "Organization Name"}
                    </p>
                    {
                      <p className="text-xs  text-left text-muted-foreground leading-tight mt-0.5">
                        {formData.tagline || "[Organization Tagline]"}
                      </p>
                    }
                  </div>
                </div>
              </div>
            </div>

            {/* Mobile separator between logo and fields */}
            <Separator className="lg:hidden my-6" />

            {/* Right — Document Header + Description */}
            <div className="space-y-6">
              {/* Document Header */}
              <div className="space-y-4">
                <div>
                  <p className="text-sm font-medium flex items-center gap-1.5">
                    <FileText className="h-4 w-4" />
                    Document Header
                  </p>
                  <p className="text-sm text-muted-foreground">
                    Appears on all generated PDFs (Requisition, PO, Payment
                    Voucher, GRN)
                  </p>
                </div>

                <Input
                  label="Organization Name"
                  id="doc-header-name"
                  value={formData.name}
                  onChange={(e) => handleInputChange("name", e.target.value)}
                  placeholder="Enter organization name"
                  disabled={isUpdating}
                />
                <Input
                  label="Tagline"
                  id="doc-header-tagline"
                  value={formData.tagline}
                  onChange={(e) => handleInputChange("tagline", e.target.value)}
                  placeholder="e.g. Ministry of Finance — Procurement Division"
                  disabled={isUpdating}
                />
              </div>

              {/* Description */}
              <Textarea
                label="Description"
                id="workspace-description"
                value={formData.description}
                onChange={(e) =>
                  handleInputChange("description", e.target.value)
                }
                placeholder="Enter workspace description (optional)"
                rows={3}
                disabled={isUpdating}
              />
            </div>
          </div>

          <div className="flex justify-end pt-2">
            <Button
              onClick={handleUpdateWorkspace}
              disabled={!hasChanges || isUpdating || !formData.name.trim()}
              isLoading={isUpdating}
              loadingText="Saving..."
            >
              <Save className="h-4 w-4" />
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

      {/* Procurement Flow */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <ArrowDownUp className="h-5 w-5" />
            Procurement Flow
          </CardTitle>
          <CardDescription>
            Set the default document ordering for all purchase orders in this
            workspace. Individual POs can override this setting at creation
            time.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <RadioGroup
            value={procurementFlow}
            onValueChange={(v) =>
              setProcurementFlow(v as "goods_first" | "payment_first")
            }
            className="space-y-3"
          >
            <div className="flex items-start gap-3 rounded-lg border p-4 cursor-pointer hover:bg-muted/50 transition-colors">
              <RadioGroupItem
                value="goods_first"
                id="flow-goods-first"
                className="mt-0.5"
              />
              <Label
                htmlFor="flow-goods-first"
                className="cursor-pointer space-y-1"
              >
                <span className="font-medium">
                  Goods-First (Recommended for government)
                </span>
                <p className="text-sm text-muted-foreground font-normal">
                  Goods must be received and the GRN approved before a payment
                  voucher can be created. Flow: REQ → PO → GRN → PV → Payment
                </p>
              </Label>
            </div>
            <div className="flex items-start gap-3 rounded-lg border p-4 cursor-pointer hover:bg-muted/50 transition-colors">
              <RadioGroupItem
                value="payment_first"
                id="flow-payment-first"
                className="mt-0.5"
              />
              <Label
                htmlFor="flow-payment-first"
                className="cursor-pointer space-y-1"
              >
                <span className="font-medium">
                  Payment-First (Commercial / upfront payment)
                </span>
                <p className="text-sm text-muted-foreground font-normal">
                  Payment is processed before goods are delivered. A GRN is
                  created after delivery to confirm receipt. Flow: REQ → PO → PV
                  → Payment → GRN
                </p>
              </Label>
            </div>
          </RadioGroup>
          <div className="flex justify-end pt-2">
            <Button
              onClick={handleSaveProcurementFlow}
              disabled={
                isSavingSettings ||
                !settingsData ||
                procurementFlow === settingsData.procurementFlow
              }
              isLoading={isSavingSettings}
              loadingText="Saving..."
            >
              <Save className="h-4 w-4" />
              Save Flow Setting
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Procurement Automation */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Zap className="h-5 w-5" />
            Procurement Automation
          </CardTitle>
          <CardDescription>
            Control which documents are created automatically and how far they
            progress without manual intervention.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Flow chain explainer */}
          <div className="rounded-lg border bg-muted/40 px-4 py-3 space-y-2">
            <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wide">
              Procurement chains
            </p>
            <div className="flex flex-wrap gap-3 text-xs">
              <div className="flex items-center gap-1.5">
                <span className="inline-flex items-center rounded-full border px-2 py-0.5 font-medium bg-background text-foreground">
                  Goods-first
                </span>
                <span className="text-muted-foreground font-mono">
                  PO → GRN → PV
                </span>
              </div>
              <div className="flex items-center gap-1.5">
                <span className="inline-flex items-center rounded-full border px-2 py-0.5 font-medium bg-background text-foreground">
                  Payment-first
                </span>
                <span className="text-muted-foreground font-mono">
                  PO → PV → GRN
                </span>
              </div>
            </div>
          </div>

          {/* Toggle controls */}
          <div className="space-y-4">
            {/* autoCreatePVFromPO */}
            <div className="flex items-start justify-between gap-4 rounded-lg border p-4">
              <div className="space-y-0.5">
                <p className="text-sm font-medium leading-none">
                  Auto-create PV from approved PO
                </p>
                <p className="text-xs text-muted-foreground pt-1">
                  Payment-first chain — a Payment Voucher is created
                  automatically once a PO is approved.
                </p>
              </div>
              <Switch
                checked={autoCreatePVFromPO}
                onCheckedChange={(v) => {
                  setAutoCreatePVFromPO(v);
                  setAutomationHasChanges(true);
                }}
              />
            </div>

            {/* autoCreatePVFromGRN */}
            <div className="flex items-start justify-between gap-4 rounded-lg border p-4">
              <div className="space-y-0.5">
                <p className="text-sm font-medium leading-none">
                  Auto-create PV after GRN completes
                </p>
                <p className="text-xs text-muted-foreground pt-1">
                  Goods-first chain — a Payment Voucher is created automatically
                  once a GRN is fully signed off and approved.
                </p>
              </div>
              <Switch
                checked={autoCreatePVFromGRN}
                onCheckedChange={(v) => {
                  setAutoCreatePVFromGRN(v);
                  setAutomationHasChanges(true);
                }}
              />
            </div>

            {/* autoCreateGRNFromPO */}
            <div className="flex items-start justify-between gap-4 rounded-lg border p-4">
              <div className="space-y-0.5">
                <p className="text-sm font-medium leading-none">
                  Auto-create draft GRN from approved PO
                </p>
                <p className="text-xs text-muted-foreground pt-1">
                  Goods-first chain — creates a draft GRN placeholder for the
                  receiver to sign. Goods receipt still requires two human
                  signatures before submission.
                </p>
              </div>
              <Switch
                checked={autoCreateGRNFromPO}
                onCheckedChange={(v) => {
                  setAutoCreateGRNFromPO(v);
                  setAutomationHasChanges(true);
                }}
              />
            </div>
          </div>

          {/* PV automation level */}
          <div className="space-y-2">
            <div>
              <p className="text-sm font-medium">
                Payment Voucher automation level
              </p>
              <p className="text-xs text-muted-foreground mt-0.5">
                How far auto-created Payment Vouchers progress without manual
                intervention.
              </p>
            </div>
            <Select
              value={pvAutomationLevel}
              onValueChange={(v) => {
                setPvAutomationLevel(
                  v as "manual" | "auto_submit" | "auto_approve",
                );
                setAutomationHasChanges(true);
              }}
            >
              <SelectTrigger className="w-full sm:w-72">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="manual">Manual</SelectItem>
                <SelectItem value="auto_submit">
                  Auto-submit to approval
                </SelectItem>
                <SelectItem value="auto_approve">
                  Auto-approve under cap
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Auto-approve cap — only when auto_approve is selected */}
          {pvAutomationLevel === "auto_approve" && (
            <div className="space-y-2">
              <div>
                <p className="text-sm font-medium">
                  Auto-approve maximum amount
                </p>
                <p className="text-xs text-muted-foreground mt-0.5">
                  PVs at or below this amount are auto-approved; larger ones are
                  only submitted for human approval.
                </p>
              </div>
              <Input
                id="auto-approve-max-amount"
                type="number"
                min={0}
                step={1}
                value={autoApproveMaxAmount}
                onChange={(e) => {
                  const parsed = parseFloat(e.target.value);
                  setAutoApproveMaxAmount(
                    isNaN(parsed) ? "" : parsed,
                  );
                  setAutomationHasChanges(true);
                }}
                placeholder="e.g. 50000"
                className="w-full sm:w-72"
              />
            </div>
          )}

          <div className="flex justify-end pt-2">
            <Button
              onClick={handleSaveAutomation}
              disabled={!automationHasChanges || isSavingSettings || !settingsData}
              isLoading={isSavingSettings}
              loadingText="Saving..."
            >
              <Save className="h-4 w-4" />
              Save Automation
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Stamp of Issuing Officer */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Stamp className="h-5 w-5" />
            Stamp of Issuing Officer
          </CardTitle>
          <CardDescription>
            Rubber-stamp image printed on every Goods Received Note PDF.
            Certifying officers can override this on a per-GRN basis when
            signing.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 sm:grid-cols-[140px_1fr] gap-4 items-start">
            <div className="h-32 w-32 rounded-md border border-dashed bg-white flex items-center justify-center overflow-hidden">
              {stampImageUrl ? (
                // eslint-disable-next-line @next/next/no-img-element
                <img
                  src={stampImageUrl}
                  alt="Organization stamp"
                  className="max-h-full max-w-full object-contain"
                />
              ) : (
                <div className="flex flex-col items-center text-muted-foreground">
                  <Stamp className="h-8 w-8 mb-1" />
                  <span className="text-[10px]">No stamp uploaded</span>
                </div>
              )}
            </div>
            <div className="space-y-3">
              <div className="flex flex-wrap gap-2">
                <label className="inline-flex items-center gap-1.5 cursor-pointer rounded-md border px-3 py-2 text-sm hover:bg-muted/50">
                  {stampUploading ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <Upload className="h-4 w-4" />
                  )}
                  {stampImageUrl ? "Replace stamp" : "Upload stamp"}
                  <input
                    type="file"
                    accept="image/*"
                    className="hidden"
                    disabled={stampUploading || isSavingSettings}
                    onChange={(e) =>
                      handleStampFile(e.target.files?.[0] ?? null)
                    }
                  />
                </label>
                {stampImageUrl ? (
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={handleClearStamp}
                    disabled={stampUploading || isSavingSettings}
                  >
                    <X className="h-4 w-4 mr-1" />
                    Remove
                  </Button>
                ) : null}
              </div>
              <p className="text-xs text-muted-foreground">
                Transparent PNG recommended. JPG, PNG, GIF or WebP up to 10 MB.
              </p>
            </div>
          </div>
          <div className="flex justify-end pt-2">
            <Button
              onClick={handleSaveStamp}
              disabled={
                !stampHasChanges || isSavingSettings || stampUploading
              }
              isLoading={isSavingSettings}
              loadingText="Saving..."
            >
              <Save className="h-4 w-4" />
              Save Stamp
            </Button>
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
                isLoading={isDeleting}
                loadingText="Deleting..."
              >
                <Trash2 className="h-4 w-4" />
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
