"use client";

import { useState, useEffect } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Checkbox } from "@/components/ui/checkbox";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Separator } from "@/components/ui/separator";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  CreditCard,
  ArrowUpDown,
  Shield,
  Clock,
  History,
  AlertTriangle,
  Loader2,
  Users,
  HardDrive,
  Calendar,
  CheckCircle,
  XCircle,
  RefreshCw,
  Plus,
} from "lucide-react";
import { toast } from "sonner";
import {
  getAllSubscriptionTiers,
  getAllSubscriptionFeatures,
  changeOrganizationTier,
  overrideOrganizationLimits,
  extendOrganizationTrial,
  getOrganizationAuditLogs,
  type SubscriptionTier,
  type SubscriptionFeature,
} from "@/app/_actions/subscriptions";
import { type Organization } from "@/app/_actions/organizations";
import { getTierBadge, getTrialStatusBadge } from "@/lib/tier-utils";
import { OverrideLimitsForm } from "./override-limits-form";
import { TrialResetForm } from "./trial-reset-form";
import { ChangeTierForm } from "./change-tier-form";

interface ManageSubscriptionDialogProps {
  organization: Organization;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onOrganizationUpdated: () => void;
}

export function ManageSubscriptionDialog({
  organization,
  open,
  onOpenChange,
  onOrganizationUpdated,
}: ManageSubscriptionDialogProps) {
  const [activeTab, setActiveTab] = useState("overview");
  const [tiers, setTiers] = useState<SubscriptionTier[]>([]);
  const [features, setFeatures] = useState<SubscriptionFeature[]>([]);
  const [auditLogs, setAuditLogs] = useState<any[]>([]);
  const [isLoadingData, setIsLoadingData] = useState(false);

  useEffect(() => {
    if (open) {
      setActiveTab("overview");
      loadData();
    }
  }, [open]);

  const loadData = async () => {
    setIsLoadingData(true);
    try {
      const [tiersResult, featuresResult, logsResult] = await Promise.all([
        getAllSubscriptionTiers(),
        getAllSubscriptionFeatures(),
        getOrganizationAuditLogs(organization.id, 1, 20),
      ]);

      if (tiersResult.success && tiersResult.data) {
        setTiers(tiersResult.data.filter((t) => t.isActive));
      }
      if (featuresResult.success && featuresResult.data) {
        setFeatures(featuresResult.data.filter((f) => f.isActive));
      }
      if (logsResult.success && logsResult.data) {
        setAuditLogs(
          Array.isArray(logsResult.data)
            ? logsResult.data
            : logsResult.data?.logs || [],
        );
      }
    } catch {
      // Individual sections handle their own errors
    } finally {
      setIsLoadingData(false);
    }
  };

  const handleSuccess = () => {
    onOrganizationUpdated();
    loadData();
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-3xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <CreditCard className="h-5 w-5" />
            Manage Subscription: {organization.name}
          </DialogTitle>
          <DialogDescription>
            View and manage subscription tier, limits, trial, and history
          </DialogDescription>
        </DialogHeader>

        <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
          <TabsList className="grid w-full grid-cols-5">
            <TabsTrigger value="overview">Overview</TabsTrigger>
            <TabsTrigger value="change-tier">Change Tier</TabsTrigger>
            <TabsTrigger value="overrides">Overrides</TabsTrigger>
            <TabsTrigger value="trial">Trial</TabsTrigger>
            <TabsTrigger value="history">History</TabsTrigger>
          </TabsList>

          {/* Overview Tab */}
          <TabsContent value="overview" className="space-y-4">
            <OverviewTab
              organization={organization}
              tiers={tiers}
            />
          </TabsContent>

          {/* Change Tier Tab */}
          <TabsContent value="change-tier" className="space-y-4">
            <ChangeTierTab
              organization={organization}
              tiers={tiers}
              isLoadingTiers={isLoadingData}
              onSuccess={handleSuccess}
            />
          </TabsContent>

          {/* Override Limits Tab */}
          <TabsContent value="overrides" className="space-y-4">
            <OverrideLimitsTab
              organization={organization}
              features={features}
              isLoadingFeatures={isLoadingData}
              onSuccess={handleSuccess}
            />
          </TabsContent>

          {/* Trial Management Tab */}
          <TabsContent value="trial" className="space-y-4">
            <TrialTab organization={organization} onSuccess={handleSuccess} />
          </TabsContent>

          {/* History Tab */}
          <TabsContent value="history" className="space-y-4">
            <HistoryTab auditLogs={auditLogs} isLoading={isLoadingData} />
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  );
}

/* ─── Overview Tab ───────────────────────────────────────────────── */

function OverviewTab({
  organization,
  tiers,
}: {
  organization: Organization;
  tiers: SubscriptionTier[];
}) {
  const currentTier = tiers.find(
    (t) => t.name === organization.subscription_tier,
  );
  const currentFeatures = organization.settings?.features_enabled || [];

  return (
    <div className="space-y-4">
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm flex items-center gap-2">
              <CreditCard className="h-4 w-4" />
              Subscription
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Tier:</span>
              {getTierBadge(organization.subscription_tier)}
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Status:</span>
              {getTrialStatusBadge(organization.trial_status)}
            </div>
            {currentTier && (
              <>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Price:</span>
                  <span className="text-sm font-medium">
                    ${currentTier.priceMonthly}/mo
                  </span>
                </div>
              </>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm flex items-center gap-2">
              <Users className="h-4 w-4" />
              Limits
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Users:</span>
              <span className="text-sm font-medium">
                {organization.user_count} /{" "}
                {organization.settings?.max_users ||
                  currentTier?.maxTeamMembers ||
                  "∞"}
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Storage:</span>
              <span className="text-sm font-medium">
                {currentTier?.maxDocuments || "N/A"} documents
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">
                Custom Branding:
              </span>
              {organization.settings?.custom_branding ? (
                <CheckCircle className="h-4 w-4 text-green-500" />
              ) : (
                <XCircle className="h-4 w-4 text-muted-foreground" />
              )}
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">API Access:</span>
              {organization.settings?.api_access ? (
                <CheckCircle className="h-4 w-4 text-green-500" />
              ) : (
                <XCircle className="h-4 w-4 text-muted-foreground" />
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Trial Info */}
      {(organization.trial_status === "trial" ||
        organization.trial_status === "expired") && (
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm flex items-center gap-2">
              <Clock className="h-4 w-4" />
              Trial Period
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-3 gap-4 text-sm">
              <div>
                <span className="text-muted-foreground">Start:</span>
                <p className="font-medium">
                  {organization.trial_start_date
                    ? new Date(
                        organization.trial_start_date,
                      ).toLocaleDateString()
                    : "N/A"}
                </p>
              </div>
              <div>
                <span className="text-muted-foreground">End:</span>
                <p className="font-medium">
                  {organization.trial_end_date
                    ? new Date(organization.trial_end_date).toLocaleString(undefined, { dateStyle: "medium", timeStyle: "short" })
                    : "N/A"}
                </p>
              </div>
              <div>
                <span className="text-muted-foreground">Remaining:</span>
                <p className="font-medium">
                  {organization.days_remaining !== undefined
                    ? `${organization.days_remaining} days`
                    : "N/A"}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Enabled Features */}
      {currentFeatures.length > 0 && (
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm">Enabled Features</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex flex-wrap gap-1">
              {currentFeatures.map((feature) => (
                <Badge key={feature} variant="outline" className="text-xs">
                  {feature}
                </Badge>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

/* ─── Change Tier Tab ────────────────────────────────────────────── */

function ChangeTierTab({
  organization,
  tiers,
  isLoadingTiers,
  onSuccess,
}: {
  organization: Organization;
  tiers: SubscriptionTier[];
  isLoadingTiers: boolean;
  onSuccess: () => void;
}) {
  return (
    <ChangeTierForm
      organizationId={organization.id}
      currentTier={organization.subscription_tier}
      tiers={tiers}
      isLoadingTiers={isLoadingTiers}
      showOverrideLimits
      onSuccess={onSuccess}
      fullWidthSubmit
    />
  );
}

/* ─── Override Limits Tab ────────────────────────────────────────── */

function OverrideLimitsTab({
  organization,
  features,
  isLoadingFeatures,
  onSuccess,
}: {
  organization: Organization;
  features: SubscriptionFeature[];
  isLoadingFeatures: boolean;
  onSuccess: () => void;
}) {
  return (
    <OverrideLimitsForm
      organizationId={organization.id}
      currentSettings={{
        max_users: organization.settings?.max_users,
        features_enabled: organization.settings?.features_enabled,
        subscription_tier: organization.subscription_tier,
      }}
      availableFeatures={features}
      isLoadingFeatures={isLoadingFeatures}
      onSuccess={onSuccess}
      fullWidthSubmit
    />
  );
}

/* ─── Trial Management Tab ───────────────────────────────────────── */

function TrialTab({
  organization,
  onSuccess,
}: {
  organization: Organization;
  onSuccess: () => void;
}) {
  const [mode, setMode] = useState<"extend" | "reset">("extend");
  const [daysToAdd, setDaysToAdd] = useState(7);
  const [reason, setReason] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");

  const handleExtend = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (reason.trim().length < 5) {
      setError("Reason must be at least 5 characters");
      return;
    }
    if (daysToAdd < 1 || daysToAdd > 30) {
      setError("Days must be between 1 and 30");
      return;
    }

    setIsLoading(true);
    try {
      const result = await extendOrganizationTrial(organization.id, {
        days_to_add: daysToAdd,
        reason: reason.trim(),
      });

      if (result.success) {
        toast.success(`Trial extended by ${daysToAdd} days`);
        setReason("");
        setDaysToAdd(7);
        onSuccess();
      } else {
        setError(result.message || "Failed to extend trial");
      }
    } catch {
      setError("Failed to extend trial");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="space-y-4">
      {/* Trial Status */}
      <div className="rounded-lg border p-3 bg-muted/50">
        <div className="grid grid-cols-3 gap-4 text-sm">
          <div>
            <span className="text-muted-foreground">Status:</span>
            <p className="font-medium capitalize">
              {organization.trial_status}
            </p>
          </div>
          <div>
            <span className="text-muted-foreground">End Date:</span>
            <p className="font-medium">
              {organization.trial_end_date
                ? new Date(organization.trial_end_date).toLocaleString(undefined, { dateStyle: "medium", timeStyle: "short" })
                : "N/A"}
            </p>
          </div>
          <div>
            <span className="text-muted-foreground">Remaining:</span>
            <p className="font-medium">
              {organization.days_remaining !== undefined
                ? organization.days_remaining > 0
                  ? `${organization.days_remaining} days`
                  : `Expired ${Math.abs(organization.days_remaining)} days ago`
                : "N/A"}
            </p>
          </div>
        </div>
      </div>

      {/* Mode Toggle */}
      <div className="flex gap-2">
        <Button
          variant={mode === "extend" ? "default" : "outline"}
          size="sm"
          onClick={() => {
            setMode("extend");
            setError("");
            setReason("");
          }}
        >
          <Plus className="mr-1 h-3 w-3" />
          Extend Trial
        </Button>
        <Button
          variant={mode === "reset" ? "default" : "outline"}
          size="sm"
          onClick={() => {
            setMode("reset");
            setError("");
            setReason("");
          }}
        >
          <RefreshCw className="mr-1 h-3 w-3" />
          Reset Trial
        </Button>
      </div>

      {mode === "extend" ? (
        <form onSubmit={handleExtend} className="space-y-4">
          <p className="text-sm text-muted-foreground">
            Add days to the current trial/grace period without resetting the
            start date.
          </p>
          <div className="space-y-2">
            <label className="text-sm font-medium">Days to Add</label>
            <Input
              type="number"
              min="1"
              max="30"
              value={daysToAdd}
              onChange={(e) => setDaysToAdd(parseInt(e.target.value) || 7)}
            />
            <p className="text-xs text-muted-foreground">Between 1 and 30</p>
          </div>
          <div className="space-y-2">
            <label className="text-sm font-medium">
              Reason <span className="text-red-500">*</span>
            </label>
            <Textarea
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              placeholder="Why extend? (min 5 characters)"
              rows={2}
            />
          </div>
          {error && (
            <div className="flex items-center gap-2 text-sm text-red-600 bg-red-50 dark:bg-red-950/20 p-2 rounded">
              <AlertTriangle className="h-4 w-4 flex-shrink-0" />
              {error}
            </div>
          )}
          <Button type="submit" isLoading={isLoading} loadingText="Extending..." className="w-full">
            <Plus className="mr-2 h-4 w-4" />
            Extend by {daysToAdd} Days
          </Button>
        </form>
      ) : (
        <div>
          <p className="text-sm text-muted-foreground mb-4">
            Reset the trial completely with a new start date and duration. This
            clears any grace period.
          </p>
          <TrialResetForm
            organizationId={organization.id}
            onSuccess={onSuccess}
            fullWidthSubmit
          />
        </div>
      )}
    </div>
  );
}

/* ─── History Tab ────────────────────────────────────────────────── */

function HistoryTab({
  auditLogs,
  isLoading,
}: {
  auditLogs: any[];
  isLoading: boolean;
}) {
  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-8 text-muted-foreground">
        <Loader2 className="h-4 w-4 animate-spin mr-2" />
        Loading history...
      </div>
    );
  }

  if (auditLogs.length === 0) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        <History className="h-8 w-8 mx-auto mb-2 opacity-50" />
        <p>No audit history found for this organization</p>
      </div>
    );
  }

  const getActionBadge = (action: string) => {
    if (
      action.includes("tier") ||
      action.includes("upgrade") ||
      action.includes("downgrade")
    ) {
      return <Badge className="bg-blue-100 text-blue-800">{action}</Badge>;
    }
    if (action.includes("override") || action.includes("limit")) {
      return <Badge className="bg-purple-100 text-purple-800">{action}</Badge>;
    }
    if (action.includes("trial")) {
      return <Badge className="bg-yellow-100 text-yellow-800">{action}</Badge>;
    }
    return <Badge variant="outline">{action}</Badge>;
  };

  return (
    <div className="space-y-3 max-h-[400px] overflow-y-auto">
      {auditLogs.map((log, index) => (
        <div
          key={log.id || index}
          className="flex items-start gap-3 rounded-lg border p-3"
        >
          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-primary/10 flex-shrink-0">
            <History className="h-4 w-4 text-primary" />
          </div>
          <div className="flex-1 min-w-0 space-y-1">
            <div className="flex items-center justify-between gap-2">
              {getActionBadge(log.action || "unknown")}
              <span className="text-xs text-muted-foreground flex-shrink-0">
                {log.createdAt || log.created_at
                  ? new Date(log.createdAt || log.created_at).toLocaleString()
                  : ""}
              </span>
            </div>
            {log.description && (
              <p className="text-sm text-muted-foreground">{log.description}</p>
            )}
            {log.changes && (
              <div className="text-xs text-muted-foreground bg-muted/50 p-2 rounded">
                {typeof log.changes === "string"
                  ? log.changes
                  : JSON.stringify(log.changes, null, 2)}
              </div>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}
