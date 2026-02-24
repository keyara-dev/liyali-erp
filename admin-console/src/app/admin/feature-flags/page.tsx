"use client";

import { useState } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { notify } from "@/lib/notify";
import {
  Flag,
  BarChart3,
  FileText,
  Plus,
  Download,
} from "lucide-react";

// Components
import { FeatureFlagsFilters } from "./components/feature-flags-filters";
import { FeatureFlagsStatsGrid } from "./components/feature-flags-stats-grid";
import { FeatureFlagsTable } from "./components/feature-flags-table";
import { FeatureFlagEditDialog } from "./components/feature-flag-edit-dialog";
import { FeatureFlagAnalyticsDialog } from "./components/feature-flag-analytics-dialog";

// Hooks
import {
  useFeatureFlags,
  useFeatureFlagStats,
  useCreateFeatureFlag,
  useUpdateFeatureFlag,
  useDeleteFeatureFlag,
  useToggleFeatureFlag,
  useArchiveFeatureFlag,
} from "@/hooks/use-feature-flags";

// Types from actions
import {
  type FeatureFlag,
  type FeatureFlagFilters,
  type BulkFlagOperation,
} from "@/app/_actions/feature-flags";

export default function FeatureFlagsPage() {
  // State
  const [activeTab, setActiveTab] = useState("flags");
  const [filters, setFilters] = useState<FeatureFlagFilters>({});
  const [selectedFlags, setSelectedFlags] = useState<string[]>([]);
  const [editingFlag, setEditingFlag] = useState<FeatureFlag | null>(null);
  const [analyticsFlag, setAnalyticsFlag] = useState<FeatureFlag | null>(null);
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [showAnalyticsDialog, setShowAnalyticsDialog] = useState(false);

  // TanStack Query hooks
  const {
    data: flags = [],
    isLoading,
    refetch: refetchFlags,
    isRefetching: isRefreshing,
  } = useFeatureFlags(filters);
  const { data: stats, refetch: refetchStats } = useFeatureFlagStats();

  // Mutation hooks
  const createFlagMutation = useCreateFeatureFlag();
  const updateFlagMutation = useUpdateFeatureFlag();
  const deleteFlagMutation = useDeleteFeatureFlag();
  const toggleFlagMutation = useToggleFeatureFlag();
  const archiveFlagMutation = useArchiveFeatureFlag();

  const refreshData = async () => {
    await Promise.all([refetchFlags(), refetchStats()]);
    notify("Data refreshed successfully.", {
      title: "Success",
      variant: "success",
    });
  };

  const handleCreateFlag = async (
    flagData: Omit<
      FeatureFlag,
      | "id"
      | "created_at"
      | "updated_at"
      | "created_by"
      | "updated_by"
      | "evaluation_count"
      | "last_evaluated"
    >,
  ) => {
    try {
      await createFlagMutation.mutateAsync(flagData);
      setShowCreateDialog(false);
      notify("Feature flag created successfully.", {
        title: "Success",
        variant: "success",
      });
    } catch (error) {
      notify("Failed to create feature flag.", {
        title: "Error",
        variant: "destructive",
      });
    }
  };

  const handleUpdateFlag = async (
    flagData: Omit<
      FeatureFlag,
      | "id"
      | "created_at"
      | "updated_at"
      | "created_by"
      | "updated_by"
      | "evaluation_count"
      | "last_evaluated"
    >,
  ) => {
    if (!editingFlag) return;

    try {
      await updateFlagMutation.mutateAsync({
        id: editingFlag.id,
        updates: flagData,
      });
      setEditingFlag(null);
      notify("Feature flag updated successfully.", {
        title: "Success",
        variant: "success",
      });
    } catch (error) {
      notify("Failed to update feature flag.", {
        title: "Error",
        variant: "destructive",
      });
    }
  };

  const handleDeleteFlag = async (flagId: string) => {
    try {
      await deleteFlagMutation.mutateAsync(flagId);
      notify("Feature flag deleted successfully.", {
        title: "Success",
        variant: "success",
      });
    } catch (error) {
      notify("Failed to delete feature flag.", {
        title: "Error",
        variant: "destructive",
      });
    }
  };

  const handleToggleFlag = async (flagId: string) => {
    try {
      await toggleFlagMutation.mutateAsync(flagId);
      notify("Feature flag toggled successfully.", {
        title: "Success",
        variant: "success",
      });
    } catch (error) {
      notify("Failed to toggle feature flag.", {
        title: "Error",
        variant: "destructive",
      });
    }
  };

  const handleArchiveFlag = async (flagId: string) => {
    try {
      await archiveFlagMutation.mutateAsync(flagId);
      notify("Feature flag archived successfully.", {
        title: "Success",
        variant: "success",
      });
    } catch (error) {
      notify("Failed to archive feature flag.", {
        title: "Error",
        variant: "destructive",
      });
    }
  };

  const handleDuplicateFlag = (flag: FeatureFlag) => {
    setEditingFlag({
      ...flag,
      id: "",
      key: `${flag.key}_copy`,
      name: `${flag.name} (Copy)`,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      created_by: "current-user",
      updated_by: "current-user",
      evaluation_count: 0,
      enabled: false,
    });
  };

  const handleViewAnalytics = (flag: FeatureFlag) => {
    setAnalyticsFlag(flag);
    setShowAnalyticsDialog(true);
  };

  const handleBulkOperation = async (_operation: BulkFlagOperation) => {
    notify("Bulk operations are coming soon.", {
      title: "Coming Soon",
      variant: "default",
    });
  };

  const handleExport = async () => {
    notify("Export functionality is coming soon.", {
      title: "Coming Soon",
      variant: "default",
    });
  };

  const handleImport = () => {
    notify("Import functionality is coming soon.", {
      title: "Coming Soon",
      variant: "default",
    });
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <h1 className="text-3xl font-bold flex items-center gap-2">
            <Flag className="h-8 w-8" />
            Feature Flags
          </h1>
          <p className="text-muted-foreground">
            Manage feature toggles, experiments, and rollout controls
          </p>
        </div>
        <Button onClick={() => setShowCreateDialog(true)}>
          <Plus className="h-4 w-4 mr-2" />
          Create Flag
        </Button>
      </div>

      {/* Stats Grid */}
      {stats && <FeatureFlagsStatsGrid stats={stats} isLoading={isLoading} />}

      {/* Main Content */}
      <Tabs
        value={activeTab}
        onValueChange={setActiveTab}
        className="space-y-6"
      >
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="flags" className="flex items-center gap-2">
            <Flag className="h-4 w-4" />
            Flags
          </TabsTrigger>
          <TabsTrigger value="analytics" className="flex items-center gap-2">
            <BarChart3 className="h-4 w-4" />
            Analytics
          </TabsTrigger>
          <TabsTrigger value="templates" className="flex items-center gap-2">
            <FileText className="h-4 w-4" />
            Templates
          </TabsTrigger>
          <TabsTrigger value="audit" className="flex items-center gap-2">
            <FileText className="h-4 w-4" />
            Audit
          </TabsTrigger>
        </TabsList>

        {/* Flags Tab */}
        <TabsContent value="flags" className="space-y-6">
          <FeatureFlagsFilters
            filters={filters}
            onFiltersChange={setFilters}
            onExport={handleExport}
            onImport={handleImport}
            onRefresh={refreshData}
            isLoading={isRefreshing}
          />

          <FeatureFlagsTable
            flags={flags}
            selectedFlags={selectedFlags}
            onSelectionChange={setSelectedFlags}
            onEdit={setEditingFlag}
            onDelete={handleDeleteFlag}
            onToggle={handleToggleFlag}
            onArchive={handleArchiveFlag}
            onDuplicate={handleDuplicateFlag}
            onViewAnalytics={handleViewAnalytics}
            isLoading={isLoading}
          />

          {/* Bulk Actions */}
          {selectedFlags.length > 0 && (
            <div className="flex items-center gap-2 p-4 bg-muted rounded-lg">
              <span className="text-sm font-medium">
                {selectedFlags.length} flag{selectedFlags.length > 1 ? "s" : ""}{" "}
                selected
              </span>
              <div className="flex gap-2 ml-auto">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() =>
                    handleBulkOperation({
                      action: "enable",
                      flagIds: selectedFlags,
                    })
                  }
                >
                  Enable Selected
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() =>
                    handleBulkOperation({
                      action: "disable",
                      flagIds: selectedFlags,
                    })
                  }
                >
                  Disable Selected
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() =>
                    handleBulkOperation({
                      action: "export",
                      flagIds: selectedFlags,
                    })
                  }
                >
                  <Download className="h-4 w-4 mr-2" />
                  Export
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() =>
                    handleBulkOperation({
                      action: "archive",
                      flagIds: selectedFlags,
                    })
                  }
                >
                  Archive Selected
                </Button>
              </div>
            </div>
          )}
        </TabsContent>

        {/* Analytics Tab */}
        <TabsContent value="analytics" className="space-y-6">
          <div className="text-center py-12">
            <BarChart3 className="mx-auto h-12 w-12 text-muted-foreground/50" />
            <h3 className="mt-4 text-lg font-semibold">Global Analytics</h3>
            <p className="text-sm text-muted-foreground mt-1">Coming Soon</p>
            <p className="text-xs text-muted-foreground/70 mt-2 max-w-md mx-auto">
              Track flag evaluation rates, variant distributions, and performance metrics across all feature flags.
            </p>
          </div>
        </TabsContent>

        {/* Templates Tab */}
        <TabsContent value="templates" className="space-y-6">
          <div className="text-center py-12">
            <FileText className="mx-auto h-12 w-12 text-muted-foreground/50" />
            <h3 className="mt-4 text-lg font-semibold">Flag Templates</h3>
            <p className="text-sm text-muted-foreground mt-1">Coming Soon</p>
            <p className="text-xs text-muted-foreground/70 mt-2 max-w-md mx-auto">
              Pre-built flag configurations for common patterns like gradual rollouts, A/B tests, and kill switches.
            </p>
          </div>
        </TabsContent>

        {/* Audit Tab */}
        <TabsContent value="audit" className="space-y-6">
          <div className="text-center py-12">
            <FileText className="mx-auto h-12 w-12 text-muted-foreground/50" />
            <h3 className="mt-4 text-lg font-semibold">Audit Trail</h3>
            <p className="text-sm text-muted-foreground mt-1">Coming Soon</p>
            <p className="text-xs text-muted-foreground/70 mt-2 max-w-md mx-auto">
              Complete history of flag changes including who made changes, when, and what was modified.
            </p>
          </div>
        </TabsContent>
      </Tabs>

      {/* Create/Edit Dialog */}
      <FeatureFlagEditDialog
        flag={editingFlag}
        open={showCreateDialog || !!editingFlag}
        onOpenChange={(open) => {
          if (!open) {
            setShowCreateDialog(false);
            setEditingFlag(null);
          }
        }}
        onSave={editingFlag ? handleUpdateFlag : handleCreateFlag}
        isLoading={false}
      />

      {/* Analytics Dialog */}
      <FeatureFlagAnalyticsDialog
        flag={analyticsFlag}
        open={showAnalyticsDialog}
        onOpenChange={(open) => {
          setShowAnalyticsDialog(open);
          if (!open) {
            setAnalyticsFlag(null);
          }
        }}
      />
    </div>
  );
}
