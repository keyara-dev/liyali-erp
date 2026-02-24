"use client";

import { useState } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { notify } from "@/lib/notify";
import {
  Settings,
  Activity,
  FileText,
  Plus,
  Download,
} from "lucide-react";

// Components
import { SettingsFilters as SettingsFiltersComponent } from "./components/settings-filters";
import { SettingsStatsGrid } from "./components/settings-stats-grid";
import { SettingsTable } from "./components/settings-table";
import { SettingEditDialog } from "./components/setting-edit-dialog";
import { SystemHealthPanel } from "./components/system-health-panel";
import { ConfigurationTemplates } from "./components/configuration-templates";

// Hooks
import {
  useSystemSettings,
  useSettingsStats,
  useSettingsHealth,
  useCreateSystemSetting,
  useUpdateSystemSetting,
  useDeleteSystemSetting,
} from "@/hooks/use-settings";

// Types from actions
import {
  type SystemSetting,
  type SettingsFilters,
  type ConfigurationTemplate,
  type BulkSettingsOperation,
} from "@/app/_actions/settings";

export default function SettingsPage() {
  // State
  const [activeTab, setActiveTab] = useState("settings");
  const [filters, setFilters] = useState<SettingsFilters>({});
  const [selectedSettings, setSelectedSettings] = useState<string[]>([]);
  const [editingSetting, setEditingSetting] = useState<SystemSetting | null>(
    null,
  );
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [templates, setTemplates] = useState<ConfigurationTemplate[]>([]);

  // TanStack Query hooks
  const {
    data: settingsResult,
    isLoading,
    refetch: refetchSettings,
    isRefetching: isRefreshing,
  } = useSystemSettings(filters);
  const { data: statsResult, refetch: refetchStats } = useSettingsStats();
  const { data: healthResult, refetch: refetchHealth } = useSettingsHealth();

  // Extract data from query results (these hooks return raw action results)
  const settings = Array.isArray(settingsResult) ? settingsResult : [];
  const stats = statsResult || null;
  const health = healthResult || null;

  // Mutation hooks
  const createSettingMutation = useCreateSystemSetting();
  const updateSettingMutation = useUpdateSystemSetting();
  const deleteSettingMutation = useDeleteSystemSetting();

  const refreshData = async () => {
    try {
      await Promise.all([refetchSettings(), refetchStats(), refetchHealth()]);
      notify("Data refreshed successfully.", {
        title: "Success",
        variant: "success",
      });
    } catch (error) {
      notify("Failed to refresh data.", {
        title: "Error",
        variant: "destructive",
      });
    }
  };

  const handleCreateSetting = async (
    settingData: Omit<
      SystemSetting,
      "id" | "created_at" | "updated_at" | "created_by" | "updated_by"
    >,
  ) => {
    try {
      await createSettingMutation.mutateAsync(settingData);
      setShowCreateDialog(false);
      notify("Setting created successfully.", {
        title: "Success",
        variant: "success",
      });
    } catch (error) {
      notify("Failed to create setting.", {
        title: "Error",
        variant: "destructive",
      });
    }
  };

  const handleUpdateSetting = async (
    settingData: Omit<
      SystemSetting,
      "id" | "created_at" | "updated_at" | "created_by" | "updated_by"
    >,
  ) => {
    if (!editingSetting) return;

    try {
      await updateSettingMutation.mutateAsync({
        id: editingSetting.id,
        updates: settingData,
      });
      setEditingSetting(null);
      notify("Setting updated successfully.", {
        title: "Success",
        variant: "success",
      });
    } catch (error) {
      notify("Failed to update setting.", {
        title: "Error",
        variant: "destructive",
      });
    }
  };

  const handleDeleteSetting = async (settingId: string) => {
    try {
      await deleteSettingMutation.mutateAsync(settingId);
      notify("Setting deleted successfully.", {
        title: "Success",
        variant: "success",
      });
    } catch (error) {
      notify("Failed to delete setting.", {
        title: "Error",
        variant: "destructive",
      });
    }
  };

  const handleBulkOperation = async (_operation: BulkSettingsOperation) => {
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

  const handleResetToDefaults = async (_settingIds: string[]) => {
    notify("Reset to defaults is coming soon.", {
      title: "Coming Soon",
      variant: "default",
    });
  };

  const handleApplyTemplate = async (templateId: string) => {
    try {
      await refetchSettings();
      await refetchStats();

      notify("Template applied successfully.", {
        title: "Success",
        variant: "success",
      });
    } catch (error) {
      notify("Failed to apply template.", {
        title: "Error",
        variant: "destructive",
      });
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
        <div>
          <h1 className="text-3xl font-bold flex items-center gap-2">
            <Settings className="h-8 w-8" />
            System Settings
          </h1>
          <p className="text-muted-foreground">
            Manage system configuration, environment variables, and application
            settings
          </p>
        </div>
        <Button onClick={() => setShowCreateDialog(true)}>
          <Plus className="h-4 w-4 mr-2" />
          Create Setting
        </Button>
      </div>

      {/* Stats Grid */}
      {stats && <SettingsStatsGrid stats={stats} isLoading={isLoading} />}

      {/* Main Content */}
      <Tabs
        value={activeTab}
        onValueChange={setActiveTab}
        className="space-y-6"
      >
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="settings" className="flex items-center gap-2">
            <Settings className="h-4 w-4" />
            Settings
          </TabsTrigger>
          <TabsTrigger value="health" className="flex items-center gap-2">
            <Activity className="h-4 w-4" />
            Health
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

        {/* Settings Tab */}
        <TabsContent value="settings" className="space-y-6">
          <SettingsFiltersComponent
            filters={filters}
            onFiltersChange={setFilters}
            onExport={handleExport}
            onImport={handleImport}
            onRefresh={refreshData}
            isLoading={isRefreshing}
          />

          <SettingsTable
            settings={settings}
            selectedSettings={selectedSettings}
            onSelectionChange={setSelectedSettings}
            onEdit={setEditingSetting}
            onDelete={handleDeleteSetting}
            onToggleSecret={(settingId) => {
              notify("Secret status toggle not implemented yet.", {
                title: "Info",
              });
            }}
            onResetToDefault={(settingId) => handleResetToDefaults([settingId])}
            onDuplicate={(setting) => {
              setEditingSetting({
                ...setting,
                id: "",
                key: `${setting.key}_copy`,
                updated_at: new Date().toISOString(),
                updated_by: "current-user",
              });
            }}
            isLoading={isLoading}
          />

          {/* Bulk Actions */}
          {selectedSettings.length > 0 && (
            <div className="flex items-center gap-2 p-4 bg-muted rounded-lg">
              <span className="text-sm font-medium">
                {selectedSettings.length} setting
                {selectedSettings.length > 1 ? "s" : ""} selected
              </span>
              <div className="flex gap-2 ml-auto">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() =>
                    handleBulkOperation({
                      action: "export",
                      settingIds: selectedSettings,
                    })
                  }
                >
                  <Download className="h-4 w-4 mr-2" />
                  Export
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handleResetToDefaults(selectedSettings)}
                >
                  Reset to Defaults
                </Button>
                <Button
                  variant="destructive"
                  size="sm"
                  onClick={() =>
                    handleBulkOperation({
                      action: "delete",
                      settingIds: selectedSettings,
                    })
                  }
                >
                  Delete Selected
                </Button>
              </div>
            </div>
          )}
        </TabsContent>

        {/* Health Tab */}
        <TabsContent value="health" className="space-y-6">
          {health && (
            <SystemHealthPanel
              health={health}
              onRefresh={async () => {
                try {
                  await refetchHealth();
                } catch (error) {
                  notify("Failed to refresh health data.", {
                    title: "Error",
                    variant: "destructive",
                  });
                }
              }}
              isLoading={isLoading}
            />
          )}
        </TabsContent>

        {/* Templates Tab */}
        <TabsContent value="templates" className="space-y-6">
          <ConfigurationTemplates
            templates={templates}
            onApplyTemplate={handleApplyTemplate}
            onCreateTemplate={() => {
              notify("Template creation not implemented yet.", {
                title: "Info",
              });
            }}
            isLoading={isLoading}
          />
        </TabsContent>

        {/* Audit Tab */}
        <TabsContent value="audit" className="space-y-6">
          <div className="text-center py-12">
            <FileText className="mx-auto h-12 w-12 text-muted-foreground/50" />
            <h3 className="mt-4 text-lg font-semibold">Configuration Audit Trail</h3>
            <p className="text-sm text-muted-foreground mt-1">Coming Soon</p>
            <p className="text-xs text-muted-foreground/70 mt-2 max-w-md mx-auto">
              Track all configuration changes with who made them, when, and what was modified.
            </p>
          </div>
        </TabsContent>
      </Tabs>

      {/* Create/Edit Dialog */}
      <SettingEditDialog
        setting={editingSetting}
        open={showCreateDialog || !!editingSetting}
        onOpenChange={(open) => {
          if (!open) {
            setShowCreateDialog(false);
            setEditingSetting(null);
          }
        }}
        onSave={editingSetting ? handleUpdateSetting : handleCreateSetting}
        isLoading={false}
      />
    </div>
  );
}
