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
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
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
  AlertTriangle,
  Clock,
  CheckCircle,
  XCircle,
  Eye,
  Check,
  MoreHorizontal,
  MapPin,
  Monitor,
  User,
} from "lucide-react";
import { toast } from "sonner";
import { resolveAPIError, type APIError } from "@/app/_actions/api-monitoring";

interface APIErrorsPanelProps {
  errors: APIError[];
  isLoading: boolean;
  onErrorUpdated: () => void;
}

export function APIErrorsPanel({
  errors,
  isLoading,
  onErrorUpdated,
}: APIErrorsPanelProps) {
  const [selectedError, setSelectedError] = useState<APIError | null>(null);
  const [showDetailsDialog, setShowDetailsDialog] = useState(false);
  const [showResolveDialog, setShowResolveDialog] = useState(false);
  const [resolutionNotes, setResolutionNotes] = useState("");
  const [isResolving, setIsResolving] = useState(false);

  const handleViewDetails = (error: APIError) => {
    setSelectedError(error);
    setShowDetailsDialog(true);
  };

  const handleResolveError = (error: APIError) => {
    setSelectedError(error);
    setResolutionNotes("");
    setShowResolveDialog(true);
  };

  const handleConfirmResolve = async () => {
    if (!selectedError) return;

    setIsResolving(true);
    try {
      const result = await resolveAPIError(selectedError.id, resolutionNotes);
      if (result.success) {
        toast.success("Error resolved successfully");
        onErrorUpdated();
        setShowResolveDialog(false);
        setSelectedError(null);
        setResolutionNotes("");
      } else {
        toast.error("Failed to resolve error");
      }
    } catch (error) {
      console.error("Error resolving API error:", error);
      toast.error("Failed to resolve error");
    } finally {
      setIsResolving(false);
    }
  };

  const getStatusCodeBadge = (statusCode: number) => {
    if (statusCode >= 200 && statusCode < 300) {
      return <Badge variant="default">{statusCode}</Badge>;
    } else if (statusCode >= 400 && statusCode < 500) {
      return <Badge variant="secondary">{statusCode}</Badge>;
    } else if (statusCode >= 500) {
      return <Badge variant="destructive">{statusCode}</Badge>;
    }
    return <Badge variant="outline">{statusCode}</Badge>;
  };

  const getErrorTypeIcon = (errorType: string) => {
    switch (errorType.toLowerCase()) {
      case "timeout":
        return <Clock className="h-4 w-4 text-orange-600" />;
      case "validation":
        return <AlertTriangle className="h-4 w-4 text-yellow-600" />;
      case "authentication":
      case "authorization":
        return <XCircle className="h-4 w-4 text-red-600" />;
      default:
        return <AlertTriangle className="h-4 w-4 text-red-600" />;
    }
  };

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>API Errors</CardTitle>
          <CardDescription>Loading errors...</CardDescription>
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
                <div className="h-6 w-16 bg-muted animate-pulse rounded" />
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    );
  }

  if (errors.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>API Errors</CardTitle>
          <CardDescription>No errors found</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="text-center py-8">
            <CheckCircle className="h-12 w-12 text-green-600 mx-auto mb-4" />
            <p className="text-muted-foreground">No API errors found</p>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>API Errors</CardTitle>
          <CardDescription>
            {errors.length} errors •{" "}
            {errors.filter((e) => !e.is_resolved).length} unresolved
          </CardDescription>
        </CardHeader>
        <CardContent>
          <ScrollArea className="h-96">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Error</TableHead>
                  <TableHead>Endpoint</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Time</TableHead>
                  <TableHead>Response Time</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {errors.map((error) => (
                  <TableRow key={error.id}>
                    <TableCell>
                      <div className="space-y-1">
                        <div className="flex items-center gap-2">
                          {error.is_resolved ? (
                            <CheckCircle className="h-4 w-4 text-green-600" />
                          ) : (
                            <XCircle className="h-4 w-4 text-red-600" />
                          )}
                          <span className="font-medium text-sm">
                            {error.error_message.substring(0, 50)}...
                          </span>
                        </div>
                        <div className="text-xs text-muted-foreground">
                          Request ID: {error.request_id}
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="space-y-1">
                        <div className="flex items-center gap-2">
                          <Badge variant="outline" className="text-xs">
                            {error.method}
                          </Badge>
                          <span className="font-mono text-sm">
                            {error.endpoint_path}
                          </span>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      {getStatusCodeBadge(error.status_code)}
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        {getErrorTypeIcon(error.error_type)}
                        <span className="text-sm capitalize">
                          {error.error_type.replace("_", " ")}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="text-sm">
                        {new Date(error.occurred_at).toLocaleString()}
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="text-sm">
                        {error.response_time.toFixed(0)}ms
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
                            onClick={() => handleViewDetails(error)}
                          >
                            <Eye className="mr-2 h-4 w-4" />
                            View Details
                          </DropdownMenuItem>
                          {!error.is_resolved && (
                            <>
                              <DropdownMenuSeparator />
                              <DropdownMenuItem
                                onClick={() => handleResolveError(error)}
                              >
                                <Check className="mr-2 h-4 w-4" />
                                Mark Resolved
                              </DropdownMenuItem>
                            </>
                          )}
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </ScrollArea>
        </CardContent>
      </Card>

      {/* Error Details Dialog */}
      <Dialog open={showDetailsDialog} onOpenChange={setShowDetailsDialog}>
        <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <AlertTriangle className="h-5 w-5" />
              Error Details
            </DialogTitle>
            <DialogDescription>
              Detailed information about the API error
            </DialogDescription>
          </DialogHeader>

          {selectedError && (
            <div className="space-y-6">
              {/* Error Summary */}
              <div className="space-y-4">
                <h3 className="text-sm font-medium">Error Summary</h3>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-sm text-muted-foreground">Status</p>
                    <div className="flex items-center gap-2">
                      {selectedError.is_resolved ? (
                        <CheckCircle className="h-4 w-4 text-green-600" />
                      ) : (
                        <XCircle className="h-4 w-4 text-red-600" />
                      )}
                      <span>
                        {selectedError.is_resolved ? "Resolved" : "Unresolved"}
                      </span>
                    </div>
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">Error Type</p>
                    <div className="flex items-center gap-2">
                      {getErrorTypeIcon(selectedError.error_type)}
                      <span className="capitalize">
                        {selectedError.error_type.replace("_", " ")}
                      </span>
                    </div>
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">Status Code</p>
                    {getStatusCodeBadge(selectedError.status_code)}
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">
                      Response Time
                    </p>
                    <p className="font-medium">
                      {selectedError.response_time.toFixed(0)}ms
                    </p>
                  </div>
                </div>
              </div>

              {/* Request Information */}
              <div className="space-y-4">
                <h3 className="text-sm font-medium">Request Information</h3>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-sm text-muted-foreground">Endpoint</p>
                    <div className="flex items-center gap-2">
                      <Badge variant="outline" className="text-xs">
                        {selectedError.method}
                      </Badge>
                      <span className="font-mono">
                        {selectedError.endpoint_path}
                      </span>
                    </div>
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">Request ID</p>
                    <p className="font-mono text-sm">
                      {selectedError.request_id}
                    </p>
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">IP Address</p>
                    <div className="flex items-center gap-1">
                      <MapPin className="h-3 w-3 text-muted-foreground" />
                      <span className="font-mono text-sm">
                        {selectedError.ip_address}
                      </span>
                    </div>
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">User ID</p>
                    <div className="flex items-center gap-1">
                      <User className="h-3 w-3 text-muted-foreground" />
                      <span className="text-sm">
                        {selectedError.user_id || "Anonymous"}
                      </span>
                    </div>
                  </div>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">User Agent</p>
                  <div className="flex items-center gap-1">
                    <Monitor className="h-3 w-3 text-muted-foreground" />
                    <span className="text-sm break-all">
                      {selectedError.user_agent}
                    </span>
                  </div>
                </div>
              </div>

              {/* Error Message */}
              <div className="space-y-4">
                <h3 className="text-sm font-medium">Error Message</h3>
                <div className="p-3 bg-red-50 border border-red-200 rounded-lg">
                  <p className="text-sm text-red-800">
                    {selectedError.error_message}
                  </p>
                </div>
              </div>

              {/* Request/Response Bodies */}
              {(selectedError.request_body || selectedError.response_body) && (
                <div className="space-y-4">
                  <h3 className="text-sm font-medium">Request/Response Data</h3>
                  <div className="grid gap-4">
                    {selectedError.request_body && (
                      <div>
                        <p className="text-sm text-muted-foreground mb-2">
                          Request Body
                        </p>
                        <pre className="p-3 bg-muted rounded-lg text-xs overflow-x-auto">
                          {JSON.stringify(
                            JSON.parse(selectedError.request_body),
                            null,
                            2,
                          )}
                        </pre>
                      </div>
                    )}
                    {selectedError.response_body && (
                      <div>
                        <p className="text-sm text-muted-foreground mb-2">
                          Response Body
                        </p>
                        <pre className="p-3 bg-muted rounded-lg text-xs overflow-x-auto">
                          {JSON.stringify(
                            JSON.parse(selectedError.response_body),
                            null,
                            2,
                          )}
                        </pre>
                      </div>
                    )}
                  </div>
                </div>
              )}

              {/* Timestamps */}
              <div className="space-y-4">
                <h3 className="text-sm font-medium">Timestamps</h3>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-sm text-muted-foreground">Occurred At</p>
                    <p className="font-medium">
                      {new Date(selectedError.occurred_at).toLocaleString()}
                    </p>
                  </div>
                  {selectedError.resolved_at && (
                    <div>
                      <p className="text-sm text-muted-foreground">
                        Resolved At
                      </p>
                      <p className="font-medium">
                        {new Date(selectedError.resolved_at).toLocaleString()}
                      </p>
                    </div>
                  )}
                </div>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>

      {/* Resolve Error Dialog */}
      <Dialog open={showResolveDialog} onOpenChange={setShowResolveDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Resolve Error</DialogTitle>
            <DialogDescription>
              Mark this error as resolved and add resolution notes.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="resolution-notes">
                Resolution Notes (Optional)
              </Label>
              <Textarea
                id="resolution-notes"
                placeholder="Describe how this error was resolved..."
                value={resolutionNotes}
                onChange={(e) => setResolutionNotes(e.target.value)}
                rows={4}
              />
            </div>
          </div>

          <div className="flex justify-end gap-2">
            <Button
              variant="outline"
              onClick={() => setShowResolveDialog(false)}
              disabled={isResolving}
            >
              Cancel
            </Button>
            <Button onClick={handleConfirmResolve} disabled={isResolving} isLoading={isResolving} loadingText="Resolving...">
              Mark Resolved
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
