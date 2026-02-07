"use client";

import { useState, useEffect } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { notify } from "@/lib/notify";
import {
  Settings,
  Activity,
  FileText,
  Plus,
  Download,
  Upload,
} from "lucide-react";

// Components
import { SettingsFilters as SettingsFiltersComponent } from "./components/settings-filters";
import { SettingsStatsGrid } from "./components/settings-stats-grid";
import { SettingsTable } from "./components/settings-table";
import { SettingEditDialog } from "./components/setting-edit-dialog";
import { SystemHealthPanel } from "./components/system-health-panel";
import { ConfigurationTemplates } from "./components/configuration-templates";

// Actions
import {
  getSystemSettings,
  getSettingsStats,
  getSystemHealth,
  getConfigurationTemplates,
  createSystemSetting,
  updateSystemSetting,
  deleteSystemSetting,
  bulkUpdateSettings,
  exportConfiguration,
  importConfiguration,
  resetToDefaults,
  type SystemSetting,
  type SettingsFilters,
  type SettingsStats,
  type SystemHealth,
  type ConfigurationTemplate,
  type BulkSettingsOperation,
} from "@/app/_actions/settings";

export default function SettingsPage() {
  // State
  const [activeTab, setActiveTab] = useState("settings");
  const [settings, setSettings] = useState<SystemSetting[]>([]);
  const [stats, setStats] = useState<SettingsStats | null>(null);
  const [health, setHealth] = useState<SystemHealth | null>(null);
  const [templates, setTemplates] = useState<ConfigurationTemplate[]>([]);
  const [filters, setFilters] = useState<SettingsFilters>({});
  const [selectedSettings, setSelectedSettings] = useState<string[]>([]);
  const [editingSetting, setEditingSetting] = useState<SystemSetting | null>(
    null,
  );
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);

  // Load initial data
  useEffect(() => {
    loadInitialData();
  }, []);

  // Load settings when filters change
  useEffect(() => {
    loadSettings();
  }, [filters]);

  const loadInitialData = async () => {
    setIsLoading(true);
    try {
      const [settingsData, statsData, healthData, templatesData] =
        await Promise.all([
          getSystemSettings(),
          getSettingsStats(),
          getSystemHealth(),
          getConfigurationTemplates(),
        ]);

      setSettings(settingsData);
      setStats(statsData);
      setHealth(healthData);
      setTemplates(templatesData);
    } catch (error) {
      notify("Failed to load system settings data.", {
        title: "Error",
        variant: "destructive",
      });
    } finally {
      setIsLoading(false);
    }
  };

  const loadSettings = async () => {
    if (isLoading) return; // Don't reload during initial load

    try {
      const settingsData = await getSystemSettings(filters);
      setSettings(settingsData);
    } catch (error) {
      notify("Failed to load settings.", {
        title: "Error",
        variant: "destructive",
      });
    }
  };

  const refreshData = async () => {
    setIsRefreshing(true);
    try {
      const [settingsData, statsData, healthData] = await Promise.all([
        getSystemSettings(filters),
        getSettingsStats(),
        getSystemHealth(),
      ]);

      setSettings(settingsData);
      setStats(statsData);
      setHealth(healthData);

      notify("Data refreshed successfully.", {
        title: "Success",
        variant: "success",
      });
    } catch (error) {
      notify("Failed to refresh data.", {
        title: "Error",
        variant: "destructive",
      });
    } finally {
      setIsRefreshing(false);
    }
  };

  const handleCreateSetting = async (
    settingData: Omit<SystemSetting, "id" | "lastModified" | "modifiedBy">,
  ) => {
    try {
      await createSystemSetting(settingData);
      setShowCreateDialog(false);
      await loadSettings();
      await loadStats();

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
    settingData: Omit<SystemSetting, "id" | "lastModified" | "modifiedBy">,
  ) => {
    if (!editingSetting) return;

    try {
      await updateSystemSetting(editingSetting.id, settingData);
      setEditingSetting(null);
      await loadSettings();
      await loadStats();

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
      await deleteSystemSetting(settingId);
      await loadSettings();
      await loadStats();

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

  const handleBulkOperation = async (operation: BulkSettingsOperation) => {
    try {
      await bulkUpdateSettings(operation);
      setSelectedSettings([]);
      await loadSettings();
      await loadStats();

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
      const exportData = await exportConfiguration(
        "all",
        selectedSettings.length > 0 ? selectedSettings : undefined,
      );

      // Create and download file
      const blob = new Blob([JSON.stringify(exportData, null, 2)], {
        type: "application/json",
      });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `settings-export-${new Date().toISOString().split("T")[0]}.json`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);

      notify("Settings exported successfully.", {
        title: "Success",
        variant: "success",
      });
    } catch (error) {
      notify("Failed to export settings.", {
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
        const result = await importConfiguration({
          file,
          environment: "all",
          overwriteExisting: false,
          validateOnly: false,
        });

        await loadSettings();
        await loadStats();

        notify(
          `Import completed. ${result.imported} settings imported, ${result.skipped} skipped.`,
          {
            title: "Success",
            variant: "success",
          },
        );
      } catch (error) {
        notify("Failed to import settings.", {
          title: "Error",
          variant: "destructive",
        });
      }
    };
    input.click();
  };

  const handleResetToDefaults = async (settingIds: string[]) => {
    try {
      await resetToDefaults(settingIds);
      await loadSettings();

      notify("Settings reset to defaults successfully.", {
        title: "Success",
        variant: "success",
      });
    } catch (error) {
      notify("Failed to reset settings.", {
        title: "Error",
        variant: "destructive",
      });
    }
  };

  const handleApplyTemplate = async (templateId: string) => {
    try {
      // This would apply the template settings
      // Implementation depends on backend API
      await loadSettings();
      await loadStats();

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

  const loadStats = async () => {
    try {
      const statsData = await getSettingsStats();
      setStats(statsData);
    } catch (error) {
      // Silently fail for stats
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
              // Implementation for toggling secret status
              notify("Secret status toggle not implemented yet.", {
                title: "Info",
              });
            }}
            onResetToDefault={(settingId) => handleResetToDefaults([settingId])}
            onDuplicate={(setting) => {
              // Create a duplicate setting
              setEditingSetting({
                ...setting,
                id: "",
                key: `${setting.key}_copy`,
                lastModified: new Date().toISOString(),
                modifiedBy: "current-user",
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
                  const healthData = await getSystemHealth();
                  setHealth(healthData);
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
            <FileText className="mx-auto h-12 w-12 text-muted-foreground" />
            <h3 className="mt-4 text-lg font-semibold">Audit Trail</h3>
            <p className="text-muted-foreground">
              Configuration audit trail will be implemented here.
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
