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
  Settings,
  TestTube,
  AlertTriangle,
  Clock,
  TrendingUp,
  TrendingDown,
  Globe,
  Lock,
  Zap,
} from "lucide-react";
import { toast } from "sonner";
import {
  testAPIEndpoint,
  updateEndpointConfig,
  type APIEndpoint,
  type APIMetrics,
} from "@/app/_actions/api-monitoring";

interface APIEndpointsTableProps {
  endpoints: APIEndpoint[];
  metrics: APIMetrics[];
  isLoading: boolean;
  onEndpointUpdated: () => void;
}

export function APIEndpointsTable({
  endpoints,
  metrics,
  isLoading,
  onEndpointUpdated,
}: APIEndpointsTableProps) {
  const [selectedEndpoint, setSelectedEndpoint] = useState<APIEndpoint | null>(
    null,
  );
  const [showDetailsDialog, setShowDetailsDialog] = useState(false);
  const [showConfigDialog, setShowConfigDialog] = useState(false);
  const [isTestingEndpoint, setIsTestingEndpoint] = useState<string | null>(
    null,
  );

  const getEndpointMetrics = (endpointId: string) => {
    return metrics.find((m) => m.endpoint_id === endpointId);
  };

  const handleViewDetails = (endpoint: APIEndpoint) => {
    setSelectedEndpoint(endpoint);
    setShowDetailsDialog(true);
  };

  const handleConfigureEndpoint = (endpoint: APIEndpoint) => {
    setSelectedEndpoint(endpoint);
    setShowConfigDialog(true);
  };

  const handleTestEndpoint = async (endpoint: APIEndpoint) => {
    setIsTestingEndpoint(endpoint.id);
    try {
      const result = await testAPIEndpoint(endpoint.id);
      if (result.success && result.data) {
        const { success, status_code, response_time, error_message } =
          result.data;
        if (success) {
          toast.success(
            `Endpoint test successful: ${status_code} (${response_time}ms)`,
          );
        } else {
          toast.error(
            `Endpoint test failed: ${error_message || "Unknown error"}`,
          );
        }
      } else {
        toast.error("Failed to test endpoint");
      }
    } catch (error) {
      console.error("Error testing endpoint:", error);
      toast.error("Failed to test endpoint");
    } finally {
      setIsTestingEndpoint(null);
    }
  };

  const getMethodBadgeVariant = (method: string) => {
    switch (method.toUpperCase()) {
      case "GET":
        return "default";
      case "POST":
        return "secondary";
      case "PUT":
        return "outline";
      case "DELETE":
        return "destructive";
      case "PATCH":
        return "secondary";
      default:
        return "outline";
    }
  };

  const getStatusBadge = (
    endpoint: APIEndpoint,
    endpointMetrics?: APIMetrics,
  ) => {
    if (endpoint.is_deprecated) {
      return (
        <Badge variant="destructive" className="flex items-center gap-1">
          <AlertTriangle className="h-3 w-3" />
          Deprecated
        </Badge>
      );
    }

    if (endpointMetrics) {
      const errorRate = endpointMetrics.error_rate;
      if (errorRate > 10) {
        return (
          <Badge variant="destructive" className="flex items-center gap-1">
            <TrendingDown className="h-3 w-3" />
            High Errors
          </Badge>
        );
      } else if (errorRate > 5) {
        return (
          <Badge variant="secondary" className="flex items-center gap-1">
            <AlertTriangle className="h-3 w-3" />
            Some Errors
          </Badge>
        );
      }
    }

    return (
      <Badge variant="default" className="flex items-center gap-1">
        <TrendingUp className="h-3 w-3" />
        Healthy
      </Badge>
    );
  };

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>API Endpoints</CardTitle>
          <CardDescription>Loading endpoints...</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {[...Array(5)].map((_, i) => (
              <div
                key={i}
                className="flex items-center space-x-4 p-4 border rounded-lg"
              >
                <div className="h-6 w-16 bg-muted animate-pulse rounded" />
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

  if (endpoints.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>API Endpoints</CardTitle>
          <CardDescription>No endpoints found</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="text-center py-8">
            <Zap className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
            <p className="text-muted-foreground">No API endpoints found</p>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>API Endpoints</CardTitle>
          <CardDescription>
            {endpoints.length} endpoints • Monitor performance and health
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Endpoint</TableHead>
                <TableHead>Category</TableHead>
                <TableHead>Visibility</TableHead>
                <TableHead>Requests</TableHead>
                <TableHead>Avg Response</TableHead>
                <TableHead>Error Rate</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {endpoints.map((endpoint) => {
                const endpointMetrics = getEndpointMetrics(endpoint.id);
                return (
                  <TableRow key={endpoint.id}>
                    <TableCell>
                      <div className="space-y-1">
                        <div className="flex items-center gap-2">
                          <Badge
                            variant={getMethodBadgeVariant(endpoint.method)}
                            className="text-xs"
                          >
                            {endpoint.method}
                          </Badge>
                          <span className="font-mono text-sm">
                            {endpoint.path}
                          </span>
                        </div>
                        <div className="text-xs text-muted-foreground">
                          {endpoint.name}
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline" className="text-xs">
                        {endpoint.category}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1">
                        {endpoint.is_public ? (
                          <Globe className="h-4 w-4 text-blue-600" />
                        ) : (
                          <Lock className="h-4 w-4 text-gray-600" />
                        )}
                        <span className="text-sm">
                          {endpoint.is_public ? "Public" : "Private"}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="text-sm">
                        {endpointMetrics
                          ? endpointMetrics.total_requests.toLocaleString()
                          : "N/A"}
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1">
                        <Clock className="h-3 w-3 text-muted-foreground" />
                        <span className="text-sm">
                          {endpointMetrics
                            ? `${endpointMetrics.avg_response_time.toFixed(0)}ms`
                            : "N/A"}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="text-sm">
                        {endpointMetrics ? (
                          <span
                            className={
                              endpointMetrics.error_rate > 5
                                ? "text-red-600"
                                : "text-green-600"
                            }
                          >
                            {endpointMetrics.error_rate.toFixed(1)}%
                          </span>
                        ) : (
                          "N/A"
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      {getStatusBadge(endpoint, endpointMetrics)}
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
                            onClick={() => handleViewDetails(endpoint)}
                          >
                            <Eye className="mr-2 h-4 w-4" />
                            View Details
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            onClick={() => handleTestEndpoint(endpoint)}
                            disabled={isTestingEndpoint === endpoint.id}
                          >
                            <TestTube className="mr-2 h-4 w-4" />
                            {isTestingEndpoint === endpoint.id
                              ? "Testing..."
                              : "Test Endpoint"}
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem
                            onClick={() => handleConfigureEndpoint(endpoint)}
                          >
                            <Settings className="mr-2 h-4 w-4" />
                            Configure
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* Endpoint Details Dialog */}
      <Dialog open={showDetailsDialog} onOpenChange={setShowDetailsDialog}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Zap className="h-5 w-5" />
              Endpoint Details
            </DialogTitle>
            <DialogDescription>
              Detailed information about the API endpoint
            </DialogDescription>
          </DialogHeader>

          {selectedEndpoint && (
            <div className="space-y-6">
              {/* Basic Information */}
              <div className="space-y-4">
                <h3 className="text-sm font-medium">Basic Information</h3>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-sm text-muted-foreground">
                      Method & Path
                    </p>
                    <div className="flex items-center gap-2">
                      <Badge
                        variant={getMethodBadgeVariant(selectedEndpoint.method)}
                      >
                        {selectedEndpoint.method}
                      </Badge>
                      <span className="font-mono">{selectedEndpoint.path}</span>
                    </div>
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">Name</p>
                    <p className="font-medium">{selectedEndpoint.name}</p>
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">Category</p>
                    <Badge variant="outline">{selectedEndpoint.category}</Badge>
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">Version</p>
                    <p className="font-medium">{selectedEndpoint.version}</p>
                  </div>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Description</p>
                  <p className="text-sm">{selectedEndpoint.description}</p>
                </div>
              </div>

              {/* Configuration */}
              <div className="space-y-4">
                <h3 className="text-sm font-medium">Configuration</h3>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-sm text-muted-foreground">Rate Limit</p>
                    <p className="font-medium">
                      {selectedEndpoint.rate_limit} req/min
                    </p>
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">Timeout</p>
                    <p className="font-medium">{selectedEndpoint.timeout}ms</p>
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">Visibility</p>
                    <div className="flex items-center gap-1">
                      {selectedEndpoint.is_public ? (
                        <Globe className="h-4 w-4 text-blue-600" />
                      ) : (
                        <Lock className="h-4 w-4 text-gray-600" />
                      )}
                      <span>
                        {selectedEndpoint.is_public ? "Public" : "Private"}
                      </span>
                    </div>
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">Status</p>
                    <Badge
                      variant={
                        selectedEndpoint.is_deprecated
                          ? "destructive"
                          : "default"
                      }
                    >
                      {selectedEndpoint.is_deprecated ? "Deprecated" : "Active"}
                    </Badge>
                  </div>
                </div>
              </div>

              {/* Metrics */}
              {(() => {
                const endpointMetrics = getEndpointMetrics(selectedEndpoint.id);
                return endpointMetrics ? (
                  <div className="space-y-4">
                    <h3 className="text-sm font-medium">Performance Metrics</h3>
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <p className="text-sm text-muted-foreground">
                          Total Requests
                        </p>
                        <p className="font-medium">
                          {endpointMetrics.total_requests.toLocaleString()}
                        </p>
                      </div>
                      <div>
                        <p className="text-sm text-muted-foreground">
                          Success Rate
                        </p>
                        <p className="font-medium text-green-600">
                          {endpointMetrics.success_rate.toFixed(1)}%
                        </p>
                      </div>
                      <div>
                        <p className="text-sm text-muted-foreground">
                          Avg Response Time
                        </p>
                        <p className="font-medium">
                          {endpointMetrics.avg_response_time.toFixed(0)}ms
                        </p>
                      </div>
                      <div>
                        <p className="text-sm text-muted-foreground">
                          P95 Response Time
                        </p>
                        <p className="font-medium">
                          {endpointMetrics.p95_response_time.toFixed(0)}ms
                        </p>
                      </div>
                      <div>
                        <p className="text-sm text-muted-foreground">
                          Error Rate
                        </p>
                        <p
                          className={`font-medium ${endpointMetrics.error_rate > 5 ? "text-red-600" : "text-green-600"}`}
                        >
                          {endpointMetrics.error_rate.toFixed(1)}%
                        </p>
                      </div>
                      <div>
                        <p className="text-sm text-muted-foreground">
                          Last Request
                        </p>
                        <p className="font-medium">
                          {new Date(
                            endpointMetrics.last_request_at,
                          ).toLocaleString()}
                        </p>
                      </div>
                    </div>
                  </div>
                ) : null;
              })()}

              {/* Timestamps */}
              <div className="space-y-4">
                <h3 className="text-sm font-medium">Timestamps</h3>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-sm text-muted-foreground">Created</p>
                    <p className="font-medium">
                      {new Date(selectedEndpoint.created_at).toLocaleString()}
                    </p>
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">
                      Last Updated
                    </p>
                    <p className="font-medium">
                      {new Date(selectedEndpoint.updated_at).toLocaleString()}
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
