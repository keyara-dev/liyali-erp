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

// Actions (for operations not covered by hooks)
import {
  bulkUpdateFlags,
  exportFeatureFlags,
  importFeatureFlags,
  getFlagTemplates,
  type FeatureFlag,
  type FeatureFlagFilters,
  type FlagTemplate,
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

  const handleBulkOperation = async (operation: BulkFlagOperation) => {
    try {
      await bulkUpdateFlags(operation);
      setSelectedFlags([]);
      await refetchFlags();
      await refetchStats();

      notify(`Bulk ${operation.action} completed successfully.`, {
        title: "Success",
        variant: "success",
      });
    } catch (error) {
      notify(`Failed to perform bulk ${operation.action}.`, {
        title: "Error",
        variant: "destructive",
      });
    }
  };

  const handleExport = async () => {
    try {
      const exportData = await exportFeatureFlags(
        selectedFlags.length > 0 ? selectedFlags : undefined,
      );

      // Create and download file
      const blob = new Blob([JSON.stringify(exportData, null, 2)], {
        type: "application/json",
      });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `feature-flags-export-${new Date().toISOString().split("T")[0]}.json`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);

      notify("Feature flags exported successfully.", {
        title: "Success",
        variant: "success",
      });
    } catch (error) {
      notify("Failed to export feature flags.", {
        title: "Error",
        variant: "destructive",
      });
    }
  };

  const handleImport = () => {
    const input = document.createElement("input");
    input.type = "file";
    input.accept = ".json";
    input.onchange = async (e) => {
      const file = (e.target as HTMLInputElement).files?.[0];
      if (!file) return;

      try {
        const text = await file.text();
        const data = JSON.parse(text);

        const result = await importFeatureFlags({
          flags: data.flags || [data], // Support both single flag and export format
          overwriteExisting: false,
        });

        await refetchFlags();
        await refetchStats();

        notify(
          `Import completed. ${result.imported} flags imported, ${result.skipped} skipped.`,
          {
            title: "Success",
            variant: "success",
          },
        );
      } catch (error) {
        notify("Failed to import feature flags.", {
          title: "Error",
          variant: "destructive",
        });
      }
    };
    input.click();
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
            <BarChart3 className="mx-auto h-12 w-12 text-muted-foreground" />
            <h3 className="mt-4 text-lg font-semibold">Global Analytics</h3>
            <p className="text-muted-foreground">
              Global feature flag analytics dashboard will be implemented here.
            </p>
          </div>
        </TabsContent>

        {/* Templates Tab */}
        <TabsContent value="templates" className="space-y-6">
          <div className="text-center py-12">
            <FileText className="mx-auto h-12 w-12 text-muted-foreground" />
            <h3 className="mt-4 text-lg font-semibold">Flag Templates</h3>
            <p className="text-muted-foreground">
              Feature flag templates gallery will be implemented here.
            </p>
          </div>
        </TabsContent>

        {/* Audit Tab */}
        <TabsContent value="audit" className="space-y-6">
          <div className="text-center py-12">
            <FileText className="mx-auto h-12 w-12 text-muted-foreground" />
            <h3 className="mt-4 text-lg font-semibold">Audit Trail</h3>
            <p className="text-muted-foreground">
              Feature flag audit trail will be implemented here.
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
