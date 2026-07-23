"use client";

import { useState } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Label } from "@/components/ui/label";
import { SelectField } from "@/components/ui/select-field";
import { Input } from "@/components/ui/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import {
  HardDrive,
  Plus,
  Download,
  RotateCcw,
  Clock,
  CheckCircle,
  XCircle,
  Activity,
  MoreHorizontal,
  Database,
  AlertTriangle,
} from "lucide-react";
import { toast } from "sonner";
import {
  createDatabaseBackup,
  restoreDatabaseBackup,
  type DatabaseConnection,
  type DatabaseBackup,
} from "@/app/_actions/database";

interface DatabaseBackupsPanelProps {
  connections: DatabaseConnection[];
  backups: DatabaseBackup[];
  isLoading: boolean;
  onBackupUpdated: () => void;
}

export function DatabaseBackupsPanel({
  connections,
  backups,
  isLoading,
  onBackupUpdated,
}: DatabaseBackupsPanelProps) {
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [showRestoreDialog, setShowRestoreDialog] = useState(false);
  const [selectedBackup, setSelectedBackup] = useState<DatabaseBackup | null>(
    null,
  );
  const [isCreatingBackup, setIsCreatingBackup] = useState(false);
  const [isRestoringBackup, setIsRestoringBackup] = useState(false);

  // Create backup form state
  const [createForm, setCreateForm] = useState({
    connection_id: "",
    backup_type: "full" as "full" | "incremental" | "differential",
    retention_days: 30,
    description: "",
  });

  // Restore backup form state
  const [restoreForm, setRestoreForm] = useState({
    target_connection_id: "",
    restore_data: true,
    restore_schema: true,
  });

  const handleCreateBackup = async () => {
    if (!createForm.connection_id) {
      toast.error("Please select a connection");
      return;
    }

    setIsCreatingBackup(true);
    try {
      const result = await createDatabaseBackup(createForm.connection_id, {
        backup_type: createForm.backup_type,
        retention_days: createForm.retention_days,
        description: createForm.description,
      });

      if (result.success) {
        toast.success("Database backup initiated successfully");
        onBackupUpdated();
        setShowCreateDialog(false);
        setCreateForm({
          connection_id: "",
          backup_type: "full",
          retention_days: 30,
          description: "",
        });
      } else {
        toast.error("Failed to create backup");
      }
    } catch (error) {
      console.error("Error creating backup:", error);
      toast.error("Failed to create backup");
    } finally {
      setIsCreatingBackup(false);
    }
  };

  const handleRestoreBackup = async () => {
    if (!selectedBackup) return;

    setIsRestoringBackup(true);
    try {
      const result = await restoreDatabaseBackup(selectedBackup.id, {
        target_connection_id: restoreForm.target_connection_id || undefined,
        restore_data: restoreForm.restore_data,
        restore_schema: restoreForm.restore_schema,
      });

      if (result.success) {
        toast.success("Database restore initiated successfully");
        onBackupUpdated();
        setShowRestoreDialog(false);
        setSelectedBackup(null);
        setRestoreForm({
          target_connection_id: "",
          restore_data: true,
          restore_schema: true,
        });
      } else {
        toast.error("Failed to restore backup");
      }
    } catch (error) {
      console.error("Error restoring backup:", error);
      toast.error("Failed to restore backup");
    } finally {
      setIsRestoringBackup(false);
    }
  };

  const handleRestoreClick = (backup: DatabaseBackup) => {
    setSelectedBackup(backup);
    setRestoreForm({
      target_connection_id: backup.connection_id,
      restore_data: true,
      restore_schema: true,
    });
    setShowRestoreDialog(true);
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "running":
        return (
          <Badge variant="default" className="flex items-center gap-1">
            <Activity className="h-3 w-3" />
            Running
          </Badge>
        );
      case "completed":
        return (
          <Badge variant="default" className="flex items-center gap-1">
            <CheckCircle className="h-3 w-3" />
            Completed
          </Badge>
        );
      case "failed":
        return (
          <Badge variant="destructive" className="flex items-center gap-1">
            <XCircle className="h-3 w-3" />
            Failed
          </Badge>
        );
      case "cancelled":
        return (
          <Badge variant="secondary" className="flex items-center gap-1">
            <XCircle className="h-3 w-3" />
            Cancelled
          </Badge>
        );
      default:
        return <Badge variant="outline">{status}</Badge>;
    }
  };

  const getBackupTypeColor = (type: string) => {
    switch (type) {
      case "full":
        return "bg-blue-100 text-blue-800";
      case "incremental":
        return "bg-green-100 text-green-800";
      case "differential":
        return "bg-orange-100 text-orange-800";
      default:
        return "bg-gray-100 text-gray-800";
    }
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  const formatDuration = (startTime: string, endTime?: string) => {
    const start = new Date(startTime);
    const end = endTime ? new Date(endTime) : new Date();
    const duration = end.getTime() - start.getTime();
    const minutes = Math.floor(duration / 60000);
    const seconds = Math.floor((duration % 60000) / 1000);
    return `${minutes}m ${seconds}s`;
  };

  const activeConnections = connections.filter((c) => c.status === "connected");

  const connectionOptions = activeConnections.map((c) => ({
    value: c.id,
    label: `${c.name} (${c.type})`,
  }));

  return (
    <div className="space-y-6">
      {/* Backup Actions */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Database Backups</CardTitle>
              <CardDescription>
                Create and manage database backups for disaster recovery
              </CardDescription>
            </div>
            <Button onClick={() => setShowCreateDialog(true)}>
              <Plus className="mr-2 h-4 w-4" />
              Create Backup
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-4">
              {[...Array(5)].map((_, i) => (
                <div
                  key={i}
                  className="flex items-center space-x-4 p-4 border rounded-lg"
                >
                  <div className="h-4 w-4 bg-muted animate-pulse rounded" />
                  <div className="flex-1 space-y-2">
                    <div className="h-4 bg-muted animate-pulse rounded w-1/3" />
                    <div className="h-3 bg-muted animate-pulse rounded w-1/2" />
                  </div>
                  <div className="h-6 w-16 bg-muted animate-pulse rounded" />
                </div>
              ))}
            </div>
          ) : backups.length === 0 ? (
            <div className="text-center py-8">
              <HardDrive className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <p className="text-muted-foreground">No backups found</p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Backup</TableHead>
                  <TableHead>Connection</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Size</TableHead>
                  <TableHead>Duration</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {backups.map((backup, index) => {
                  const connection = connections.find(
                    (c) => c.id === backup.connection_id,
                  );
                  return (
                    <TableRow
                      key={
                        backup.id ||
                        `${backup.connection_id || "backup"}-${backup.started_at || "started"}-${index}`
                      }
                    >
                      <TableCell>
                        <div className="space-y-1">
                          <div className="flex items-center gap-2">
                            <HardDrive className="h-4 w-4 text-muted-foreground" />
                            <span className="font-medium">
                              {backup.backup_method}
                            </span>
                            {backup.is_automated && (
                              <Badge variant="outline" className="text-xs">
                                Auto
                              </Badge>
                            )}
                          </div>
                          <p className="text-xs text-muted-foreground">
                            Retention: {backup.retention_days} days
                          </p>
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <Database className="h-3 w-3 text-muted-foreground" />
                          <span className="text-sm">
                            {connection?.name || "Unknown"}
                          </span>
                        </div>
                      </TableCell>
                      <TableCell>
                        <span
                          className={`px-2 py-1 rounded text-xs font-medium ${getBackupTypeColor(backup.backup_type)}`}
                        >
                          {backup.backup_type.charAt(0).toUpperCase() +
                            backup.backup_type.slice(1)}
                        </span>
                      </TableCell>
                      <TableCell>{getStatusBadge(backup.status)}</TableCell>
                      <TableCell>
                        <span className="text-sm">
                          {formatBytes(backup.file_size)}
                        </span>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-1">
                          <Clock className="h-3 w-3 text-muted-foreground" />
                          <span className="text-sm">
                            {formatDuration(
                              backup.started_at,
                              backup.completed_at,
                            )}
                          </span>
                        </div>
                      </TableCell>
                      <TableCell>
                        <span className="text-sm">
                          {new Date(backup.started_at).toLocaleString()}
                        </span>
                      </TableCell>
                      <TableCell>
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" className="h-8 w-8 p-0">
                              <MoreHorizontal className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuLabel>Actions</DropdownMenuLabel>
                            {backup.status === "completed" && (
                              <>
                                <DropdownMenuItem
                                  onClick={() => handleRestoreClick(backup)}
                                >
                                  <RotateCcw className="mr-2 h-4 w-4" />
                                  Restore Backup
                                </DropdownMenuItem>
                                <DropdownMenuItem>
                                  <Download className="mr-2 h-4 w-4" />
                                  Download
                                </DropdownMenuItem>
                                <DropdownMenuSeparator />
                              </>
                            )}
                            {backup.error_message && (
                              <DropdownMenuItem>
                                <AlertTriangle className="mr-2 h-4 w-4" />
                                View Error
                              </DropdownMenuItem>
                            )}
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* Create Backup Dialog */}
      <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create Database Backup</DialogTitle>
            <DialogDescription>
              Create a new backup for your database connection
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <SelectField
              label="Database Connection"
              required
              placeholder="Select a connection"
              options={connectionOptions}
              value={createForm.connection_id}
              onValueChange={(value) =>
                setCreateForm((prev) => ({ ...prev, connection_id: value }))
              }
            />

            <SelectField
              label="Backup Type"
              options={[
                { value: "full", label: "Full Backup" },
                { value: "incremental", label: "Incremental Backup" },
                { value: "differential", label: "Differential Backup" },
              ]}
              value={createForm.backup_type}
              onValueChange={(value: string) =>
                setCreateForm((prev) => ({
                  ...prev,
                  backup_type: value as "full" | "incremental" | "differential",
                }))
              }
            />

            <Input
              label="Retention Days"
              type="number"
              min="1"
              max="365"
              value={createForm.retention_days}
              onChange={(e) =>
                setCreateForm((prev) => ({
                  ...prev,
                  retention_days: parseInt(e.target.value) || 30,
                }))
              }
            />

            <Input
              label="Description (Optional)"
              placeholder="Backup description..."
              value={createForm.description}
              onChange={(e) =>
                setCreateForm((prev) => ({
                  ...prev,
                  description: e.target.value,
                }))
              }
            />
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowCreateDialog(false)}
              disabled={isCreatingBackup}
            >
              Cancel
            </Button>
            <Button
              onClick={handleCreateBackup}
              isLoading={isCreatingBackup}
              loadingText="Creating..."
              disabled={!createForm.connection_id}
            >
              Create Backup
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Restore Backup Dialog */}
      <AlertDialog open={showRestoreDialog} onOpenChange={setShowRestoreDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Restore Database Backup</AlertDialogTitle>
            <AlertDialogDescription>
              This will restore the selected backup. This action cannot be
              undone. Please ensure you have a current backup before proceeding.
            </AlertDialogDescription>
          </AlertDialogHeader>

          <div className="space-y-4">
            <SelectField
              label="Target Connection"
              placeholder="Select target connection"
              options={connectionOptions}
              value={restoreForm.target_connection_id}
              onValueChange={(value) =>
                setRestoreForm((prev) => ({
                  ...prev,
                  target_connection_id: value,
                }))
              }
            />

            <div className="space-y-3">
              <Label>Restore Options</Label>
              <div className="flex items-center space-x-2">
                <input
                  type="checkbox"
                  id="restore-schema"
                  checked={restoreForm.restore_schema}
                  onChange={(e) =>
                    setRestoreForm((prev) => ({
                      ...prev,
                      restore_schema: e.target.checked,
                    }))
                  }
                  className="rounded border-gray-300"
                />
                <Label htmlFor="restore-schema" className="text-sm">
                  Restore Schema
                </Label>
              </div>
              <div className="flex items-center space-x-2">
                <input
                  type="checkbox"
                  id="restore-data"
                  checked={restoreForm.restore_data}
                  onChange={(e) =>
                    setRestoreForm((prev) => ({
                      ...prev,
                      restore_data: e.target.checked,
                    }))
                  }
                  className="rounded border-gray-300"
                />
                <Label htmlFor="restore-data" className="text-sm">
                  Restore Data
                </Label>
              </div>
            </div>
          </div>

          <AlertDialogFooter>
            <AlertDialogCancel disabled={isRestoringBackup}>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={handleRestoreBackup}
              disabled={
                isRestoringBackup ||
                !restoreForm.target_connection_id ||
                (!restoreForm.restore_schema && !restoreForm.restore_data)
              }
              className="bg-red-600 hover:bg-red-700"
            >
              {isRestoringBackup ? "Restoring..." : "Restore Backup"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
