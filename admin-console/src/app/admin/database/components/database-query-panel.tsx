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
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { SelectField } from "@/components/ui/select-field";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Dialog,
  DialogContent,
  DialogDescription,
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
  Play,
  Square,
  Clock,
  CheckCircle,
  XCircle,
  AlertTriangle,
  Eye,
  MoreHorizontal,
  Database,
  Activity,
} from "lucide-react";
import { notify } from "@/lib/utils";
import {
  executeDatabaseQuery,
  cancelDatabaseQuery,
  getDatabaseQueries,
  type DatabaseConnection,
  type DatabaseQuery,
} from "@/app/_actions/database";

interface DatabaseQueryPanelProps {
  connections: DatabaseConnection[];
  queries: DatabaseQuery[];
  isLoading: boolean;
  onQueryUpdated: () => void;
}

interface QueryResult {
  columns: string[];
  rows: any[][];
  row_count: number;
  execution_time: number;
  query_id: string;
}

export function DatabaseQueryPanel({
  connections,
  queries,
  isLoading,
  onQueryUpdated,
}: DatabaseQueryPanelProps) {
  const [selectedConnectionId, setSelectedConnectionId] = useState<string>("");
  const [queryText, setQueryText] = useState("");
  const [isExecuting, setIsExecuting] = useState(false);
  const [queryResult, setQueryResult] = useState<QueryResult | null>(null);
  const [queryError, setQueryError] = useState<string | null>(null);
  const [selectedQuery, setSelectedQuery] = useState<DatabaseQuery | null>(
    null,
  );
  const [showQueryDialog, setShowQueryDialog] = useState(false);

  const handleExecuteQuery = async () => {
    if (!selectedConnectionId || !queryText.trim()) {
      notify({ title: "Please select a connection and enter a query", type: "error" });
      return;
    }

    setIsExecuting(true);
    setQueryResult(null);
    setQueryError(null);

    try {
      const result = await executeDatabaseQuery(
        selectedConnectionId,
        queryText,
        {
          limit: 1000,
          timeout: 30000,
        },
      );

      if (result.success && result.data) {
        setQueryResult(result.data);
        notify({ title: `Query executed successfully in ${result.data.execution_time.toFixed(0)}ms`, type: "success" });
        onQueryUpdated();
      } else {
        setQueryError(result.message || "Query execution failed");
        notify({ title: "Query execution failed", type: "error" });
      }
    } catch (error) {
      console.error("Error executing query:", error);
      setQueryError("Failed to execute query");
      notify({ title: "Failed to execute query", type: "error" });
    } finally {
      setIsExecuting(false);
    }
  };

  const handleCancelQuery = async (queryId: string) => {
    try {
      const result = await cancelDatabaseQuery(queryId);
      if (result.success) {
        notify({ title: "Query cancelled successfully", type: "success" });
        onQueryUpdated();
      } else {
        notify({ title: "Failed to cancel query", type: "error" });
      }
    } catch (error) {
      console.error("Error cancelling query:", error);
      notify({ title: "Failed to cancel query", type: "error" });
    }
  };

  const handleViewQuery = (query: DatabaseQuery) => {
    setSelectedQuery(query);
    setShowQueryDialog(true);
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
            <Square className="h-3 w-3" />
            Cancelled
          </Badge>
        );
      default:
        return <Badge variant="outline">{status}</Badge>;
    }
  };

  const formatDuration = (startTime: string, endTime?: string) => {
    const start = new Date(startTime);
    const end = endTime ? new Date(endTime) : new Date();
    const duration = end.getTime() - start.getTime();
    return `${duration.toFixed(0)}ms`;
  };

  const activeConnections = connections.filter((c) => c.status === "connected");

  return (
    <div className="space-y-6">
      {/* Query Executor */}
      <Card>
        <CardHeader>
          <CardTitle>Query Executor</CardTitle>
          <CardDescription>
            Execute SQL queries against your database connections
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <SelectField
              label="Database Connection"
              placeholder="Select a connection"
              options={activeConnections.map((c) => ({ value: c.id, label: `${c.name} (${c.type})` }))}
              value={selectedConnectionId}
              onValueChange={setSelectedConnectionId}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="query-text">SQL Query</Label>
            <Textarea
              id="query-text"
              placeholder="Enter your SQL query here..."
              value={queryText}
              onChange={(e) => setQueryText(e.target.value)}
              rows={6}
              className="font-mono"
            />
          </div>

          <div className="flex items-center gap-2">
            <Button
              onClick={handleExecuteQuery}
              disabled={!selectedConnectionId || !queryText.trim()}
              isLoading={isExecuting}
              loadingText="Executing..."
            >
              <Play className="mr-2 h-4 w-4" />
              Execute Query
            </Button>
            {isExecuting && (
              <Button variant="outline" onClick={() => setIsExecuting(false)}>
                <Square className="mr-2 h-4 w-4" />
                Cancel
              </Button>
            )}
          </div>

          {/* Query Results */}
          {queryResult && (
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <h3 className="text-sm font-medium">Query Results</h3>
                <div className="flex items-center gap-4 text-sm text-muted-foreground">
                  <span>{queryResult.row_count} rows</span>
                  <span>{queryResult.execution_time.toFixed(0)}ms</span>
                </div>
              </div>

              <ScrollArea className="h-96 border rounded-lg">
                <Table>
                  <TableHeader>
                    <TableRow>
                      {queryResult.columns.map((column, index) => (
                        <TableHead key={index} className="font-mono">
                          {column}
                        </TableHead>
                      ))}
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {queryResult.rows.map((row, rowIndex) => (
                      <TableRow key={rowIndex}>
                        {row.map((cell, cellIndex) => (
                          <TableCell
                            key={cellIndex}
                            className="font-mono text-sm"
                          >
                            {cell === null ? (
                              <span className="text-muted-foreground italic">
                                NULL
                              </span>
                            ) : (
                              String(cell)
                            )}
                          </TableCell>
                        ))}
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </ScrollArea>
            </div>
          )}

          {/* Query Error */}
          {queryError && (
            <div className="p-3 bg-red-50 border border-red-200 rounded-lg">
              <div className="flex items-center gap-2">
                <AlertTriangle className="h-4 w-4 text-red-600" />
                <span className="text-sm font-medium text-red-800">
                  Query Error
                </span>
              </div>
              <p className="text-sm text-red-700 mt-1 font-mono">
                {queryError}
              </p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Query History */}
      <Card>
        <CardHeader>
          <CardTitle>Query History</CardTitle>
          <CardDescription>
            Recent database queries and their execution status
          </CardDescription>
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
          ) : queries.length === 0 ? (
            <div className="text-center py-8">
              <Database className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <p className="text-muted-foreground">No queries found</p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Query</TableHead>
                  <TableHead>Connection</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Duration</TableHead>
                  <TableHead>Rows</TableHead>
                  <TableHead>Started</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {queries.map((query) => {
                  const connection = connections.find(
                    (c) => c.id === query.connection_id,
                  );
                  return (
                    <TableRow key={query.id}>
                      <TableCell>
                        <div className="max-w-xs">
                          <p className="font-mono text-sm truncate">
                            {query.query_text}
                          </p>
                          <p className="text-xs text-muted-foreground">
                            Hash: {query.query_hash.substring(0, 8)}...
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
                      <TableCell>{getStatusBadge(query.status)}</TableCell>
                      <TableCell>
                        <div className="flex items-center gap-1">
                          <Clock className="h-3 w-3 text-muted-foreground" />
                          <span className="text-sm">
                            {query.execution_time
                              ? `${query.execution_time.toFixed(0)}ms`
                              : formatDuration(
                                  query.started_at,
                                  query.completed_at,
                                )}
                          </span>
                        </div>
                      </TableCell>
                      <TableCell>
                        <span className="text-sm">
                          {query.rows_affected.toLocaleString()}
                        </span>
                      </TableCell>
                      <TableCell>
                        <span className="text-sm">
                          {new Date(query.started_at).toLocaleString()}
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
                            <DropdownMenuItem
                              onClick={() => handleViewQuery(query)}
                            >
                              <Eye className="mr-2 h-4 w-4" />
                              View Details
                            </DropdownMenuItem>
                            {query.status === "running" && (
                              <>
                                <DropdownMenuSeparator />
                                <DropdownMenuItem
                                  onClick={() => handleCancelQuery(query.id)}
                                  className="text-red-600"
                                >
                                  <Square className="mr-2 h-4 w-4" />
                                  Cancel Query
                                </DropdownMenuItem>
                              </>
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

      {/* Query Details Dialog */}
      <Dialog open={showQueryDialog} onOpenChange={setShowQueryDialog}>
        <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Database className="h-5 w-5" />
              Query Details
            </DialogTitle>
            <DialogDescription>
              Detailed information about the database query
            </DialogDescription>
          </DialogHeader>

          {selectedQuery && (
            <div className="space-y-6">
              {/* Query Summary */}
              <div className="space-y-4">
                <h3 className="text-sm font-medium">Query Summary</h3>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-sm text-muted-foreground">Status</p>
                    {getStatusBadge(selectedQuery.status)}
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">Connection</p>
                    <p className="font-medium">
                      {connections.find(
                        (c) => c.id === selectedQuery.connection_id,
                      )?.name || "Unknown"}
                    </p>
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">
                      Execution Time
                    </p>
                    <p className="font-medium">
                      {selectedQuery.execution_time
                        ? `${selectedQuery.execution_time.toFixed(0)}ms`
                        : "N/A"}
                    </p>
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">
                      Rows Affected
                    </p>
                    <p className="font-medium">
                      {selectedQuery.rows_affected.toLocaleString()}
                    </p>
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">Application</p>
                    <p className="font-medium">{selectedQuery.application}</p>
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">Query Hash</p>
                    <p className="font-mono text-sm">
                      {selectedQuery.query_hash}
                    </p>
                  </div>
                </div>
              </div>

              {/* Query Text */}
              <div className="space-y-4">
                <h3 className="text-sm font-medium">Query Text</h3>
                <div className="p-3 bg-muted rounded-lg">
                  <pre className="text-sm font-mono whitespace-pre-wrap">
                    {selectedQuery.query_text}
                  </pre>
                </div>
              </div>

              {/* Error Message */}
              {selectedQuery.error_message && (
                <div className="space-y-4">
                  <h3 className="text-sm font-medium">Error Message</h3>
                  <div className="p-3 bg-red-50 border border-red-200 rounded-lg">
                    <p className="text-sm text-red-800">
                      {selectedQuery.error_message}
                    </p>
                  </div>
                </div>
              )}

              {/* Timestamps */}
              <div className="space-y-4">
                <h3 className="text-sm font-medium">Timestamps</h3>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-sm text-muted-foreground">Started At</p>
                    <p className="font-medium">
                      {new Date(selectedQuery.started_at).toLocaleString()}
                    </p>
                  </div>
                  {selectedQuery.completed_at && (
                    <div>
                      <p className="text-sm text-muted-foreground">
                        Completed At
                      </p>
                      <p className="font-medium">
                        {new Date(selectedQuery.completed_at).toLocaleString()}
                      </p>
                    </div>
                  )}
                </div>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
