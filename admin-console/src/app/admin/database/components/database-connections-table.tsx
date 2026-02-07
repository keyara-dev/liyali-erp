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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  MoreHorizontal,
  Eye,
  TestTube,
  Database,
  Server,
  Activity,
  Clock,
  Users,
  HardDrive,
  CheckCircle,
  XCircle,
  AlertTriangle,
  Wrench,
} from "lucide-react";
import { toast } from "sonner";
import {
  testDatabaseConnection,
  getDatabaseTables,
  type DatabaseConnection,
  type DatabaseTable,
} from "@/app/_actions/database";

interface DatabaseConnectionsTableProps {
  connections: DatabaseConnection[];
  isLoading: boolean;
  onConnectionUpdated: () => void;
}

export function DatabaseConnectionsTable({
  connections,
  isLoading,
  onConnectionUpdated,
}: DatabaseConnectionsTableProps) {
  const [selectedConnection, setSelectedConnection] =
    useState<DatabaseConnection | null>(null);
  const [showDetailsDialog, setShowDetailsDialog] = useState(false);
  const [isTestingConnection, setIsTestingConnection] = useState<string | null>(
    null,
  );
  const [connectionTables, setConnectionTables] = useState<DatabaseTable[]>([]);
  const [isLoadingTables, setIsLoadingTables] = useState(false);

  const handleViewDetails = async (connection: DatabaseConnection) => {
    setSelectedConnection(connection);
    setShowDetailsDialog(true);

    // Load tables for this connection
    setIsLoadingTables(true);
    try {
      const result = await getDatabaseTables(connection.id);
      if (result.success) {
        setConnectionTables(result.data || []);
      } else {
        toast.error("Failed to load database tables");
      }
    } catch (error) {
      console.error("Error loading tables:", error);
      toast.error("Failed to load database tables");
    } finally {
      setIsLoadingTables(false);
    }
  };

  const handleTestConnection = async (connection: DatabaseConnection) => {
    setIsTestingConnection(connection.id);
    try {
      const result = await testDatabaseConnection(connection.id);
      if (result.success && result.data) {
        const { success, response_time, error_message } = result.data;
        if (success) {
          toast.success(
            `Connection test successful: ${response_time.toFixed(0)}ms`,
          );
        } else {
          toast.error(
            `Connection test failed: ${error_message || "Unknown error"}`,
          );
        }
      } else {
        toast.error("Failed to test connection");
      }
    } catch (error) {
      console.error("Error testing connection:", error);
      toast.error("Failed to test connection");
    } finally {
      setIsTestingConnection(null);
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "connected":
        return (
          <Badge variant="default" className="flex items-center gap-1">
            <CheckCircle className="h-3 w-3" />
            Connected
          </Badge>
        );
      case "disconnected":
        return (
          <Badge variant="secondary" className="flex items-center gap-1">
            <XCircle className="h-3 w-3" />
            Disconnected
          </Badge>
        );
      case "error":
        return (
          <Badge variant="destructive" className="flex items-center gap-1">
            <AlertTriangle className="h-3 w-3" />
            Error
          </Badge>
        );
      case "maintenance":
        return (
          <Badge variant="outline" className="flex items-center gap-1">
            <Wrench className="h-3 w-3" />
            Maintenance
          </Badge>
        );
      default:
        return <Badge variant="outline">{status}</Badge>;
    }
  };

  const getTypeIcon = (type: string) => {
    switch (type) {
      case "postgresql":
        return <Database className="h-4 w-4 text-blue-600" />;
      case "mysql":
        return <Database className="h-4 w-4 text-orange-600" />;
      case "mongodb":
        return <Database className="h-4 w-4 text-green-600" />;
      case "redis":
        return <Database className="h-4 w-4 text-red-600" />;
      case "elasticsearch":
        return <Database className="h-4 w-4 text-yellow-600" />;
      default:
        return <Database className="h-4 w-4 text-gray-600" />;
    }
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Database Connections</CardTitle>
          <CardDescription>Loading connections...</CardDescription>
        </CardHeader>
        <CardContent>
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
                <div className="h-6 w-20 bg-muted animate-pulse rounded" />
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    );
  }

  if (connections.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Database Connections</CardTitle>
          <CardDescription>No connections found</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="text-center py-8">
            <Database className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
            <p className="text-muted-foreground">
              No database connections found
            </p>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>Database Connections</CardTitle>
          <CardDescription>
            {connections.length} connections • Monitor health and performance
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Connection</TableHead>
                <TableHead>Type</TableHead>
                <TableHead>Role</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Connections</TableHead>
                <TableHead>Last Check</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {connections.map((connection) => (
                <TableRow key={connection.id}>
                  <TableCell>
                    <div className="space-y-1">
                      <div className="flex items-center gap-2">
                        {getTypeIcon(connection.type)}
                        <span className="font-medium">{connection.name}</span>
                      </div>
                      <div className="text-xs text-muted-foreground">
                        {connection.host}:{connection.port}/
                        {connection.database}
                      </div>
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline" className="capitalize">
                      {connection.type}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-1">
                      {connection.is_primary ? (
                        <Server className="h-4 w-4 text-blue-600" />
                      ) : (
                        <Server className="h-4 w-4 text-gray-600" />
                      )}
                      <span className="text-sm">
                        {connection.is_primary ? "Primary" : "Replica"}
                      </span>
                    </div>
                  </TableCell>
                  <TableCell>{getStatusBadge(connection.status)}</TableCell>
                  <TableCell>
                    <div className="flex items-center gap-1">
                      <Users className="h-3 w-3 text-muted-foreground" />
                      <span className="text-sm">
                        {connection.active_connections}/
                        {connection.max_connections}
                      </span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-1">
                      <Clock className="h-3 w-3 text-muted-foreground" />
                      <span className="text-sm">
                        {new Date(
                          connection.last_health_check,
                        ).toLocaleString()}
                      </span>
                    </div>
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
                        <DropdownMenuItem
                          onClick={() => handleViewDetails(connection)}
                        >
                          <Eye className="mr-2 h-4 w-4" />
                          View Details
                        </DropdownMenuItem>
                        <DropdownMenuItem
                          onClick={() => handleTestConnection(connection)}
                          disabled={isTestingConnection === connection.id}
                        >
                          <TestTube className="mr-2 h-4 w-4" />
                          {isTestingConnection === connection.id
                            ? "Testing..."
                            : "Test Connection"}
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem>
                          <Activity className="mr-2 h-4 w-4" />
                          View Metrics
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* Connection Details Dialog */}
      <Dialog open={showDetailsDialog} onOpenChange={setShowDetailsDialog}>
        <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Database className="h-5 w-5" />
              Connection Details
            </DialogTitle>
            <DialogDescription>
              Detailed information about the database connection
            </DialogDescription>
          </DialogHeader>

          {selectedConnection && (
            <div className="space-y-6">
              {/* Basic Information */}
              <div className="space-y-4">
                <h3 className="text-sm font-medium">Basic Information</h3>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-sm text-muted-foreground">
                      Connection Name
                    </p>
                    <p className="font-medium">{selectedConnection.name}</p>
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">
                      Database Type
                    </p>
                    <div className="flex items-center gap-2">
                      {getTypeIcon(selectedConnection.type)}
                      <span className="capitalize">
                        {selectedConnection.type}
                      </span>
                    </div>
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">Host & Port</p>
                    <p className="font-mono">
                      {selectedConnection.host}:{selectedConnection.port}
                    </p>
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">Database</p>
                    <p className="font-mono">{selectedConnection.database}</p>
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">Username</p>
                    <p className="font-mono">{selectedConnection.username}</p>
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">Role</p>
                    <div className="flex items-center gap-1">
                      <Server className="h-4 w-4 text-muted-foreground" />
                      <span>
                        {selectedConnection.is_primary ? "Primary" : "Replica"}
                      </span>
                    </div>
                  </div>
                </div>
              </div>

              {/* Connection Status */}
              <div className="space-y-4">
                <h3 className="text-sm font-medium">Connection Status</h3>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-sm text-muted-foreground">
                      Current Status
                    </p>
                    {getStatusBadge(selectedConnection.status)}
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">Pool Size</p>
                    <p className="font-medium">
                      {selectedConnection.connection_pool_size}
                    </p>
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">
                      Active Connections
                    </p>
                    <p className="font-medium">
                      {selectedConnection.active_connections}/
                      {selectedConnection.max_connections}
                    </p>
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">
                      Last Health Check
                    </p>
                    <p className="font-medium">
                      {new Date(
                        selectedConnection.last_health_check,
                      ).toLocaleString()}
                    </p>
                  </div>
                </div>
              </div>

              {/* Database Tables */}
              <div className="space-y-4">
                <h3 className="text-sm font-medium">Database Tables</h3>
                {isLoadingTables ? (
                  <div className="space-y-2">
                    {[...Array(3)].map((_, i) => (
                      <div
                        key={i}
                        className="h-4 bg-muted animate-pulse rounded"
                      />
                    ))}
                  </div>
                ) : connectionTables.length === 0 ? (
                  <p className="text-sm text-muted-foreground">
                    No tables found
                  </p>
                ) : (
                  <div className="space-y-2 max-h-48 overflow-y-auto">
                    {connectionTables.slice(0, 10).map((table) => (
                      <div
                        key={table.id}
                        className="flex items-center justify-between p-2 border rounded"
                      >
                        <div>
                          <div className="flex items-center gap-2">
                            <HardDrive className="h-3 w-3 text-muted-foreground" />
                            <span className="text-sm font-mono">
                              {table.schema_name}.{table.table_name}
                            </span>
                            <Badge variant="outline" className="text-xs">
                              {table.table_type}
                            </Badge>
                          </div>
                          <p className="text-xs text-muted-foreground">
                            {table.row_count.toLocaleString()} rows •{" "}
                            {formatBytes(table.size_bytes)}
                          </p>
                        </div>
                        <div className="text-xs text-muted-foreground">
                          {table.index_count} indexes
                        </div>
                      </div>
                    ))}
                    {connectionTables.length > 10 && (
                      <p className="text-xs text-muted-foreground text-center">
                        ... and {connectionTables.length - 10} more tables
                      </p>
                    )}
                  </div>
                )}
              </div>

              {/* Timestamps */}
              <div className="space-y-4">
                <h3 className="text-sm font-medium">Timestamps</h3>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-sm text-muted-foreground">Created</p>
                    <p className="font-medium">
                      {new Date(selectedConnection.created_at).toLocaleString()}
                    </p>
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">
                      Last Updated
                    </p>
                    <p className="font-medium">
                      {new Date(selectedConnection.updated_at).toLocaleString()}
                    </p>
                  </div>
                </div>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </>
  );
}
