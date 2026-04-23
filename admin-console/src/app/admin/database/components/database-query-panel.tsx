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
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Eye, Clock, CheckCircle, XCircle, Database, Activity } from "lucide-react";
import type { DatabaseConnection, DatabaseQuery } from "@/app/_actions/database";

interface DatabaseQueryPanelProps {
  connections: DatabaseConnection[];
  queries: DatabaseQuery[];
  isLoading: boolean;
}

export function DatabaseQueryPanel({
  connections,
  queries,
  isLoading,
}: DatabaseQueryPanelProps) {
  const [selectedQuery, setSelectedQuery] = useState<DatabaseQuery | null>(
    null,
  );
  const [showQueryDialog, setShowQueryDialog] = useState(false);

  const formatQueryText = (queryText?: string | null) => {
    const text = queryText?.trim();
    if (!text) return "No query text available";
    return text;
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
            <XCircle className="h-3 w-3" />
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

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>Query History</CardTitle>
          <CardDescription>
            Read-only view of recent database queries and execution status
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
                {queries.map((query, index) => {
                  const connection = connections.find(
                    (c) => c.id === query.connection_id,
                  );
                  return (
                    <TableRow
                      key={
                        query.id ||
                        `${query.connection_id || "query"}-${query.started_at || "started"}-${index}`
                      }
                    >
                      <TableCell>
                        <div className="max-w-xs">
                          <p className="font-mono text-sm truncate">
                            {formatQueryText(query.query_text)}
                          </p>
                          <p className="text-xs text-muted-foreground">
                            Hash: {(query.query_hash ?? "unknown").slice(0, 8)}
                            ...
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
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleViewQuery(query)}
                        >
                          <Eye className="mr-2 h-4 w-4" />
                          View Details
                        </Button>
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

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

              <div className="space-y-4">
                <h3 className="text-sm font-medium">Query Text</h3>
                <div className="p-3 bg-muted rounded-lg">
                  <pre className="text-sm font-mono whitespace-pre-wrap">
                    {formatQueryText(selectedQuery.query_text)}
                  </pre>
                </div>
              </div>

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
