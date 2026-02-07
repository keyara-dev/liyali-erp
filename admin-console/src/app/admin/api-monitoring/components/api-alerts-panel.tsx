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
  Bell,
  BellOff,
  Shield,
  Zap,
} from "lucide-react";
import { toast } from "sonner";
import {
  acknowledgeAPIAlert,
  resolveAPIAlert,
  type APIAlert,
} from "@/app/_actions/api-monitoring";

interface APIAlertsPanelProps {
  alerts: APIAlert[];
  isLoading: boolean;
  onAlertUpdated: () => void;
}

export function APIAlertsPanel({
  alerts,
  isLoading,
  onAlertUpdated,
}: APIAlertsPanelProps) {
  const [selectedAlert, setSelectedAlert] = useState<APIAlert | null>(null);
  const [showDetailsDialog, setShowDetailsDialog] = useState(false);
  const [showAcknowledgeDialog, setShowAcknowledgeDialog] = useState(false);
  const [showResolveDialog, setShowResolveDialog] = useState(false);
  const [notes, setNotes] = useState("");
  const [isProcessing, setIsProcessing] = useState(false);

  const handleViewDetails = (alert: APIAlert) => {
    setSelectedAlert(alert);
    setShowDetailsDialog(true);
  };

  const handleAcknowledgeAlert = (alert: APIAlert) => {
    setSelectedAlert(alert);
    setNotes("");
    setShowAcknowledgeDialog(true);
  };

  const handleResolveAlert = (alert: APIAlert) => {
    setSelectedAlert(alert);
    setNotes("");
    setShowResolveDialog(true);
  };

  const handleConfirmAcknowledge = async () => {
    if (!selectedAlert) return;

    setIsProcessing(true);
    try {
      const result = await acknowledgeAPIAlert(selectedAlert.id, notes);
      if (result.success) {
        toast.success("Alert acknowledged successfully");
        onAlertUpdated();
        setShowAcknowledgeDialog(false);
        setSelectedAlert(null);
        setNotes("");
      } else {
        toast.error("Failed to acknowledge alert");
      }
    } catch (error) {
      console.error("Error acknowledging alert:", error);
      toast.error("Failed to acknowledge alert");
    } finally {
      setIsProcessing(false);
    }
  };

  const handleConfirmResolve = async () => {
    if (!selectedAlert) return;

    setIsProcessing(true);
    try {
      const result = await resolveAPIAlert(selectedAlert.id, notes);
      if (result.success) {
        toast.success("Alert resolved successfully");
        onAlertUpdated();
        setShowResolveDialog(false);
        setSelectedAlert(null);
        setNotes("");
      } else {
        toast.error("Failed to resolve alert");
      }
    } catch (error) {
      console.error("Error resolving alert:", error);
      toast.error("Failed to resolve alert");
    } finally {
      setIsProcessing(false);
    }
  };

  const getSeverityBadge = (severity: string) => {
    switch (severity.toLowerCase()) {
      case "critical":
        return <Badge variant="destructive">Critical</Badge>;
      case "high":
        return (
          <Badge className="bg-orange-600 hover:bg-orange-700">High</Badge>
        );
      case "medium":
        return (
          <Badge className="bg-yellow-600 hover:bg-yellow-700">Medium</Badge>
        );
      case "low":
        return <Badge variant="secondary">Low</Badge>;
      default:
        return <Badge variant="outline">{severity}</Badge>;
    }
  };

  const getSeverityIcon = (severity: string) => {
    switch (severity.toLowerCase()) {
      case "critical":
        return <AlertTriangle className="h-4 w-4 text-red-600" />;
      case "high":
        return <AlertTriangle className="h-4 w-4 text-orange-600" />;
      case "medium":
        return <AlertTriangle className="h-4 w-4 text-yellow-600" />;
      case "low":
        return <AlertTriangle className="h-4 w-4 text-blue-600" />;
      default:
        return <AlertTriangle className="h-4 w-4 text-gray-600" />;
    }
  };

  const getAlertTypeIcon = (alertType: string) => {
    switch (alertType.toLowerCase()) {
      case "high_error_rate":
        return <XCircle className="h-4 w-4 text-red-600" />;
      case "slow_response":
        return <Clock className="h-4 w-4 text-orange-600" />;
      case "high_traffic":
        return <Zap className="h-4 w-4 text-blue-600" />;
      case "security":
        return <Shield className="h-4 w-4 text-red-600" />;
      default:
        return <Bell className="h-4 w-4 text-gray-600" />;
    }
  };

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>API Alerts</CardTitle>
          <CardDescription>Loading alerts...</CardDescription>
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

  if (alerts.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>API Alerts</CardTitle>
          <CardDescription>No alerts found</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="text-center py-8">
            <CheckCircle className="h-12 w-12 text-green-600 mx-auto mb-4" />
            <p className="text-muted-foreground">No API alerts found</p>
          </div>
        </CardContent>
      </Card>
    );
  }

  const activeAlerts = alerts.filter((alert) => alert.is_active);
  const acknowledgedAlerts = alerts.filter(
    (alert) => alert.acknowledged_at && alert.is_active,
  );

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>API Alerts</CardTitle>
          <CardDescription>
            {alerts.length} total alerts • {activeAlerts.length} active •{" "}
            {acknowledgedAlerts.length} acknowledged
          </CardDescription>
        </CardHeader>
        <CardContent>
          <ScrollArea className="h-96">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Alert</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Severity</TableHead>
                  <TableHead>Threshold</TableHead>
                  <TableHead>Current</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Triggered</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {alerts.map((alert) => (
                  <TableRow key={alert.id}>
                    <TableCell>
                      <div className="space-y-1">
                        <div className="flex items-center gap-2">
                          {getSeverityIcon(alert.severity)}
                          <span className="font-medium text-sm">
                            {alert.title}
                          </span>
                        </div>
                        <div className="text-xs text-muted-foreground">
                          {alert.description}
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        {getAlertTypeIcon(alert.alert_type)}
                        <span className="text-sm capitalize">
                          {alert.alert_type.replace("_", " ")}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>{getSeverityBadge(alert.severity)}</TableCell>
                    <TableCell>
                      <div className="text-sm font-mono">
                        {alert.threshold_value}
                      </div>
                    </TableCell>
                    <TableCell>
                      <div
                        className={`text-sm font-mono ${
                          alert.current_value > alert.threshold_value
                            ? "text-red-600"
                            : "text-green-600"
                        }`}
                      >
                        {alert.current_value}
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-col gap-1">
                        {alert.is_active ? (
                          <Badge variant="destructive" className="text-xs">
                            Active
                          </Badge>
                        ) : (
                          <Badge variant="secondary" className="text-xs">
                            Resolved
                          </Badge>
                        )}
                        {alert.acknowledged_at && (
                          <Badge variant="outline" className="text-xs">
                            Acknowledged
                          </Badge>
                        )}
                        {alert.notification_sent && (
                          <div className="flex items-center gap-1">
                            <Bell className="h-3 w-3 text-blue-600" />
                            <span className="text-xs text-muted-foreground">
                              Notified
                            </span>
                          </div>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="text-sm">
                        {new Date(alert.triggered_at).toLocaleString()}
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
                            onClick={() => handleViewDetails(alert)}
                          >
                            <Eye className="mr-2 h-4 w-4" />
                            View Details
                          </DropdownMenuItem>
                          {alert.is_active && !alert.acknowledged_at && (
                            <>
                              <DropdownMenuSeparator />
                              <DropdownMenuItem
                                onClick={() => handleAcknowledgeAlert(alert)}
                              >
                                <Bell className="mr-2 h-4 w-4" />
                                Acknowledge
                              </DropdownMenuItem>
                            </>
                          )}
                          {alert.is_active && (
                            <DropdownMenuItem
                              onClick={() => handleResolveAlert(alert)}
                            >
                              <Check className="mr-2 h-4 w-4" />
                              Resolve
                            </DropdownMenuItem>
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

      {/* Alert Details Dialog */}
      <Dialog open={showDetailsDialog} onOpenChange={setShowDetailsDialog}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Bell className="h-5 w-5" />
              Alert Details
            </DialogTitle>
            <DialogDescription>
              Detailed information about the API alert
            </DialogDescription>
          </DialogHeader>

          {selectedAlert && (
            <div className="space-y-6">
              {/* Alert Summary */}
              <div className="space-y-4">
                <h3 className="text-sm font-medium">Alert Summary</h3>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-sm text-muted-foreground">Title</p>
                    <p className="font-medium">{selectedAlert.title}</p>
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">Severity</p>
                    {getSeverityBadge(selectedAlert.severity)}
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">Alert Type</p>
                    <div className="flex items-center gap-2">
                      {getAlertTypeIcon(selectedAlert.alert_type)}
                      <span className="capitalize">
                        {selectedAlert.alert_type.replace("_", " ")}
                      </span>
                    </div>
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">Status</p>
                    <div className="flex items-center gap-2">
                      {selectedAlert.is_active ? (
                        <XCircle className="h-4 w-4 text-red-600" />
                      ) : (
                        <CheckCircle className="h-4 w-4 text-green-600" />
                      )}
                      <span>
                        {selectedAlert.is_active ? "Active" : "Resolved"}
                      </span>
                    </div>
                  </div>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Description</p>
                  <p className="text-sm">{selectedAlert.description}</p>
                </div>
              </div>

              {/* Threshold Information */}
              <div className="space-y-4">
                <h3 className="text-sm font-medium">Threshold Information</h3>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-sm text-muted-foreground">
                      Threshold Value
                    </p>
                    <p className="font-medium font-mono">
                      {selectedAlert.threshold_value}
                    </p>
                  </div>
                  <div>
                    <p className="text-sm text-muted-foreground">
                      Current Value
                    </p>
                    <p
                      className={`font-medium font-mono ${
                        selectedAlert.current_value >
                        selectedAlert.threshold_value
                          ? "text-red-600"
                          : "text-green-600"
                      }`}
                    >
                      {selectedAlert.current_value}
                    </p>
                  </div>
                </div>
              </div>

              {/* Notification Status */}
              <div className="space-y-4">
                <h3 className="text-sm font-medium">Notification Status</h3>
                <div className="flex items-center gap-2">
                  {selectedAlert.notification_sent ? (
                    <Bell className="h-4 w-4 text-blue-600" />
                  ) : (
                    <BellOff className="h-4 w-4 text-gray-600" />
                  )}
                  <span className="text-sm">
                    {selectedAlert.notification_sent
                      ? "Notification sent"
                      : "No notification sent"}
                  </span>
                </div>
              </div>

              {/* Acknowledgment Information */}
              {selectedAlert.acknowledged_at && (
                <div className="space-y-4">
                  <h3 className="text-sm font-medium">Acknowledgment</h3>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <p className="text-sm text-muted-foreground">
                        Acknowledged At
                      </p>
                      <p className="font-medium">
                        {new Date(
                          selectedAlert.acknowledged_at,
                        ).toLocaleString()}
                      </p>
                    </div>
                    <div>
                      <p className="text-sm text-muted-foreground">
                        Acknowledged By
                      </p>
                      <p className="font-medium">
                        {selectedAlert.acknowledged_by || "Unknown"}
                      </p>
                    </div>
                  </div>
                </div>
              )}

              {/* Timestamps */}
              <div className="space-y-4">
                <h3 className="text-sm font-medium">Timestamps</h3>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <p className="text-sm text-muted-foreground">
                      Triggered At
                    </p>
                    <p className="font-medium">
                      {new Date(selectedAlert.triggered_at).toLocaleString()}
                    </p>
                  </div>
                  {selectedAlert.resolved_at && (
                    <div>
                      <p className="text-sm text-muted-foreground">
                        Resolved At
                      </p>
                      <p className="font-medium">
                        {new Date(selectedAlert.resolved_at).toLocaleString()}
                      </p>
                    </div>
                  )}
                </div>
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>

      {/* Acknowledge Alert Dialog */}
      <Dialog
        open={showAcknowledgeDialog}
        onOpenChange={setShowAcknowledgeDialog}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Acknowledge Alert</DialogTitle>
            <DialogDescription>
              Acknowledge this alert to indicate you are aware of the issue.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="acknowledge-notes">Notes (Optional)</Label>
              <Textarea
                id="acknowledge-notes"
                placeholder="Add any notes about this acknowledgment..."
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                rows={3}
              />
            </div>
          </div>

          <div className="flex justify-end gap-2">
            <Button
              variant="outline"
              onClick={() => setShowAcknowledgeDialog(false)}
              disabled={isProcessing}
            >
              Cancel
            </Button>
            <Button onClick={handleConfirmAcknowledge} disabled={isProcessing}>
              {isProcessing ? "Acknowledging..." : "Acknowledge Alert"}
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      {/* Resolve Alert Dialog */}
      <Dialog open={showResolveDialog} onOpenChange={setShowResolveDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Resolve Alert</DialogTitle>
            <DialogDescription>
              Mark this alert as resolved and add resolution notes.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="resolve-notes">Resolution Notes (Optional)</Label>
              <Textarea
                id="resolve-notes"
                placeholder="Describe how this alert was resolved..."
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                rows={4}
              />
            </div>
          </div>

          <div className="flex justify-end gap-2">
            <Button
              variant="outline"
              onClick={() => setShowResolveDialog(false)}
              disabled={isProcessing}
            >
              Cancel
            </Button>
            <Button onClick={handleConfirmResolve} disabled={isProcessing}>
              {isProcessing ? "Resolving..." : "Resolve Alert"}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
