"use client";

import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  User,
  Building2,
  Clock,
  MapPin,
  Monitor,
  Shield,
  Activity,
  AlertTriangle,
  CheckCircle,
  XCircle,
  Copy,
} from "lucide-react";
import { format } from "date-fns";
import { toast } from "sonner";
import { type AuditLog } from "@/app/_actions/audit-logs";

interface AuditLogDetailsSheetProps {
  log: AuditLog | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function AuditLogDetailsSheet({
  log,
  open,
  onOpenChange,
}: AuditLogDetailsSheetProps) {
  if (!log) return null;

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case "critical":
        return "destructive";
      case "high":
        return "destructive";
      case "medium":
        return "secondary";
      case "low":
        return "outline";
      default:
        return "outline";
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "success":
        return <CheckCircle className="h-4 w-4 text-green-600" />;
      case "failure":
        return <XCircle className="h-4 w-4 text-red-600" />;
      case "warning":
        return <AlertTriangle className="h-4 w-4 text-yellow-600" />;
      default:
        return <Activity className="h-4 w-4 text-gray-600" />;
    }
  };

  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text);
    toast.success(`${label} copied to clipboard`);
  };

  const formatDuration = (ms?: number) => {
    if (!ms) return "N/A";
    if (ms < 1000) return `${ms}ms`;
    return `${(ms / 1000).toFixed(2)}s`;
  };

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="w-[600px] sm:max-w-[600px] overflow-y-auto">
        <SheetHeader>
          <SheetTitle className="flex items-center gap-2">
            {getStatusIcon(log.status)}
            Audit Log Details
          </SheetTitle>
          <SheetDescription>
            Detailed information about this audit log entry
          </SheetDescription>
        </SheetHeader>

        <div className="space-y-6 mt-6">
          {/* Basic Information */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Basic Information</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-sm font-medium text-muted-foreground">
                    Log ID
                  </label>
                  <div className="flex items-center gap-2 mt-1">
                    <code className="text-sm bg-muted px-2 py-1 rounded">
                      {log.id}
                    </code>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => copyToClipboard(log.id, "Log ID")}
                    >
                      <Copy className="h-3 w-3" />
                    </Button>
                  </div>
                </div>

                <div>
                  <label className="text-sm font-medium text-muted-foreground">
                    Timestamp
                  </label>
                  <div className="flex items-center gap-2 mt-1">
                    <Clock className="h-4 w-4 text-muted-foreground" />
                    <span className="text-sm">
                      {format(new Date(log.timestamp), "PPpp")}
                    </span>
                  </div>
                </div>

                <div>
                  <label className="text-sm font-medium text-muted-foreground">
                    Action
                  </label>
                  <div className="mt-1">
                    <Badge variant="outline" className="capitalize">
                      {log.action}
                    </Badge>
                  </div>
                </div>

                <div>
                  <label className="text-sm font-medium text-muted-foreground">
                    Action Type
                  </label>
                  <div className="mt-1">
                    <Badge variant="secondary" className="capitalize">
                      {log.action_type}
                    </Badge>
                  </div>
                </div>

                <div>
                  <label className="text-sm font-medium text-muted-foreground">
                    Status
                  </label>
                  <div className="flex items-center gap-2 mt-1">
                    {getStatusIcon(log.status)}
                    <Badge
                      variant={
                        log.status === "success"
                          ? "default"
                          : log.status === "failure"
                            ? "destructive"
                            : "secondary"
                      }
                      className="capitalize"
                    >
                      {log.status}
                    </Badge>
                  </div>
                </div>

                <div>
                  <label className="text-sm font-medium text-muted-foreground">
                    Severity
                  </label>
                  <div className="mt-1">
                    <Badge
                      variant={getSeverityColor(log.severity)}
                      className="capitalize"
                    >
                      {log.severity}
                    </Badge>
                  </div>
                </div>

                {log.duration_ms && (
                  <div>
                    <label className="text-sm font-medium text-muted-foreground">
                      Duration
                    </label>
                    <div className="mt-1">
                      <span className="text-sm">
                        {formatDuration(log.duration_ms)}
                      </span>
                    </div>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>

          {/* User Information */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <User className="h-5 w-5" />
                User Information
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-1 gap-4">
                <div>
                  <label className="text-sm font-medium text-muted-foreground">
                    User ID
                  </label>
                  <div className="flex items-center gap-2 mt-1">
                    <code className="text-sm bg-muted px-2 py-1 rounded">
                      {log.user_id}
                    </code>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => copyToClipboard(log.user_id, "User ID")}
                    >
                      <Copy className="h-3 w-3" />
                    </Button>
                  </div>
                </div>

                <div>
                  <label className="text-sm font-medium text-muted-foreground">
                    User Name
                  </label>
                  <div className="mt-1">
                    <span className="text-sm font-medium">{log.user_name}</span>
                  </div>
                </div>

                <div>
                  <label className="text-sm font-medium text-muted-foreground">
                    User Email
                  </label>
                  <div className="mt-1">
                    <span className="text-sm">{log.user_email}</span>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Organization Information */}
          {log.organization_id && (
            <Card>
              <CardHeader>
                <CardTitle className="text-lg flex items-center gap-2">
                  <Building2 className="h-5 w-5" />
                  Organization Information
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-1 gap-4">
                  <div>
                    <label className="text-sm font-medium text-muted-foreground">
                      Organization ID
                    </label>
                    <div className="flex items-center gap-2 mt-1">
                      <code className="text-sm bg-muted px-2 py-1 rounded">
                        {log.organization_id}
                      </code>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() =>
                          copyToClipboard(
                            log.organization_id!,
                            "Organization ID",
                          )
                        }
                      >
                        <Copy className="h-3 w-3" />
                      </Button>
                    </div>
                  </div>

                  {log.organization_name && (
                    <div>
                      <label className="text-sm font-medium text-muted-foreground">
                        Organization Name
                      </label>
                      <div className="mt-1">
                        <span className="text-sm font-medium">
                          {log.organization_name}
                        </span>
                      </div>
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
          )}

          {/* Resource Information */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <Shield className="h-5 w-5" />
                Resource Information
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-sm font-medium text-muted-foreground">
                    Resource Type
                  </label>
                  <div className="mt-1">
                    <Badge variant="outline" className="capitalize">
                      {log.resource_type}
                    </Badge>
                  </div>
                </div>

                {log.resource_id && (
                  <div>
                    <label className="text-sm font-medium text-muted-foreground">
                      Resource ID
                    </label>
                    <div className="flex items-center gap-2 mt-1">
                      <code className="text-sm bg-muted px-2 py-1 rounded">
                        {log.resource_id}
                      </code>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() =>
                          copyToClipboard(log.resource_id!, "Resource ID")
                        }
                      >
                        <Copy className="h-3 w-3" />
                      </Button>
                    </div>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>

          {/* Technical Details */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <Monitor className="h-5 w-5" />
                Technical Details
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-1 gap-4">
                {log.metadata.ip_address && (
                  <div>
                    <label className="text-sm font-medium text-muted-foreground">
                      IP Address
                    </label>
                    <div className="flex items-center gap-2 mt-1">
                      <MapPin className="h-4 w-4 text-muted-foreground" />
                      <code className="text-sm bg-muted px-2 py-1 rounded">
                        {log.metadata.ip_address}
                      </code>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() =>
                          copyToClipboard(
                            log.metadata.ip_address!,
                            "IP Address",
                          )
                        }
                      >
                        <Copy className="h-3 w-3" />
                      </Button>
                    </div>
                  </div>
                )}

                {log.metadata.user_agent && (
                  <div>
                    <label className="text-sm font-medium text-muted-foreground">
                      User Agent
                    </label>
                    <div className="mt-1">
                      <code className="text-xs bg-muted p-2 rounded block break-all">
                        {log.metadata.user_agent}
                      </code>
                    </div>
                  </div>
                )}

                {log.metadata.location && (
                  <div>
                    <label className="text-sm font-medium text-muted-foreground">
                      Location
                    </label>
                    <div className="flex items-center gap-2 mt-1">
                      <MapPin className="h-4 w-4 text-muted-foreground" />
                      <span className="text-sm">{log.metadata.location}</span>
                    </div>
                  </div>
                )}

                {log.metadata.device_type && (
                  <div>
                    <label className="text-sm font-medium text-muted-foreground">
                      Device Type
                    </label>
                    <div className="mt-1">
                      <Badge variant="outline" className="capitalize">
                        {log.metadata.device_type}
                      </Badge>
                    </div>
                  </div>
                )}

                {log.metadata.session_id && (
                  <div>
                    <label className="text-sm font-medium text-muted-foreground">
                      Session ID
                    </label>
                    <div className="flex items-center gap-2 mt-1">
                      <code className="text-sm bg-muted px-2 py-1 rounded">
                        {log.metadata.session_id}
                      </code>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() =>
                          copyToClipboard(
                            log.metadata.session_id!,
                            "Session ID",
                          )
                        }
                      >
                        <Copy className="h-3 w-3" />
                      </Button>
                    </div>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>

          {/* Details */}
          {Object.keys(log.details).length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="text-lg">Additional Details</CardTitle>
                <CardDescription>
                  Specific information about this audit event
                </CardDescription>
              </CardHeader>
              <CardContent>
                <pre className="text-xs bg-muted p-4 rounded-lg overflow-auto max-h-64">
                  {JSON.stringify(log.details, null, 2)}
                </pre>
              </CardContent>
            </Card>
          )}
        </div>
      </SheetContent>
    </Sheet>
  );
}
