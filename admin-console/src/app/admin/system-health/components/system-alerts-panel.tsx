"use client";

import { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { SelectField } from "@/components/ui/select-field";
import {
  AlertTriangle,
  CheckCircle,
  XCircle,
  Clock,
  RefreshCw,
  Filter,
  Database,
  Server,
  Zap,
  Activity,
} from "lucide-react";
import { type SystemAlert } from "@/app/_actions/system-health";

interface SystemAlertsPanelProps {
  alerts: SystemAlert[];
  onAcknowledgeAlert: (alertId: string) => void;
  onRefresh: () => void;
}

export function SystemAlertsPanel({
  alerts,
  onAcknowledgeAlert,
  onRefresh,
}: SystemAlertsPanelProps) {
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [severityFilter, setSeverityFilter] = useState<string>("all");

  const filteredAlerts = alerts.filter((alert) => {
    const statusMatch = statusFilter === "all" || alert.status === statusFilter;
    const severityMatch =
      severityFilter === "all" || alert.severity === severityFilter;
    return statusMatch && severityMatch;
  });

  const getSeverityBadge = (severity: string) => {
    switch (severity) {
      case "critical":
        return (
          <Badge variant="destructive">
            <XCircle className="mr-1 h-3 w-3" />
            Critical
          </Badge>
        );
      case "high":
        return (
          <Badge
            variant="destructive"
            className="bg-orange-100 text-orange-800"
          >
            <AlertTriangle className="mr-1 h-3 w-3" />
            High
          </Badge>
        );
      case "medium":
        return (
          <Badge variant="secondary" className="bg-yellow-100 text-yellow-800">
            <AlertTriangle className="mr-1 h-3 w-3" />
            Medium
          </Badge>
        );
      case "low":
        return (
          <Badge variant="outline">
            <AlertTriangle className="mr-1 h-3 w-3" />
            Low
          </Badge>
        );
      default:
        return <Badge variant="outline">{severity}</Badge>;
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "active":
        return (
          <Badge variant="destructive">
            <AlertTriangle className="mr-1 h-3 w-3" />
            Active
          </Badge>
        );
      case "acknowledged":
        return (
          <Badge variant="secondary">
            <Clock className="mr-1 h-3 w-3" />
            Acknowledged
          </Badge>
        );
      case "resolved":
        return (
          <Badge variant="default" className="bg-green-100 text-green-800">
            <CheckCircle className="mr-1 h-3 w-3" />
            Resolved
          </Badge>
        );
      default:
        return <Badge variant="outline">{status}</Badge>;
    }
  };

  const getComponentIcon = (component: string) => {
    switch (component) {
      case "database":
        return <Database className="h-4 w-4" />;
      case "api":
        return <Zap className="h-4 w-4" />;
      case "server":
        return <Server className="h-4 w-4" />;
      case "cache":
      case "queue":
        return <Activity className="h-4 w-4" />;
      default:
        return <AlertTriangle className="h-4 w-4" />;
    }
  };

  const activeAlertsCount = alerts.filter(
    (alert) => alert.status === "active",
  ).length;
  const criticalAlertsCount = alerts.filter(
    (alert) => alert.severity === "critical",
  ).length;

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <AlertTriangle className="h-4 w-4" />
            System Alerts ({filteredAlerts.length})
          </CardTitle>
          <div className="flex items-center gap-2">
            <div className="flex items-center gap-2">
              {activeAlertsCount > 0 && (
                <Badge variant="destructive" className="text-xs">
                  {activeAlertsCount} Active
                </Badge>
              )}
              {criticalAlertsCount > 0 && (
                <Badge variant="destructive" className="text-xs">
                  {criticalAlertsCount} Critical
                </Badge>
              )}
            </div>
            <Button variant="outline" size="sm" onClick={onRefresh}>
              <RefreshCw className="mr-2 h-4 w-4" />
              Refresh
            </Button>
          </div>
        </div>

        {/* Filters */}
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <Filter className="h-4 w-4 text-muted-foreground" />
            <SelectField
              placeholder="Status"
              options={[
                { value: "all", label: "All Status" },
                { value: "active", label: "Active" },
                { value: "acknowledged", label: "Acknowledged" },
                { value: "resolved", label: "Resolved" },
              ]}
              value={statusFilter}
              onValueChange={setStatusFilter}
              classNames={{ wrapper: "w-32" }}
            />
          </div>

          <SelectField
            placeholder="Severity"
            options={[
              { value: "all", label: "All Severity" },
              { value: "critical", label: "Critical" },
              { value: "high", label: "High" },
              { value: "medium", label: "Medium" },
              { value: "low", label: "Low" },
            ]}
            value={severityFilter}
            onValueChange={setSeverityFilter}
            classNames={{ wrapper: "w-32" }}
          />
        </div>
      </CardHeader>
      <CardContent>
        {filteredAlerts.length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            <CheckCircle className="mx-auto h-12 w-12 mb-4 text-green-500" />
            <h3 className="text-lg font-medium mb-2">No Alerts</h3>
            <p>All systems are running smoothly!</p>
          </div>
        ) : (
          <div className="space-y-4">
            {filteredAlerts.map((alert) => (
              <div
                key={alert.id}
                className={`rounded-lg border p-4 ${
                  alert.severity === "critical"
                    ? "border-red-200 bg-red-50"
                    : alert.severity === "high"
                      ? "border-orange-200 bg-orange-50"
                      : "border-border bg-background"
                }`}
              >
                <div className="flex items-start justify-between">
                  <div className="flex items-start gap-3">
                    <div className="flex h-8 w-8 items-center justify-center rounded-full bg-background border">
                      {getComponentIcon(alert.component)}
                    </div>
                    <div className="flex-1">
                      <div className="flex items-center gap-2 mb-1">
                        <h4 className="font-medium">{alert.title}</h4>
                        {getSeverityBadge(alert.severity)}
                        {getStatusBadge(alert.status)}
                      </div>
                      <p className="text-sm text-muted-foreground mb-2">
                        {alert.description}
                      </p>
                      <div className="flex items-center gap-4 text-xs text-muted-foreground">
                        <span className="capitalize">
                          Component: {alert.component}
                        </span>
                        <span>
                          Created: {new Date(alert.created_at).toLocaleString()}
                        </span>
                        {alert.acknowledged_at && (
                          <span>
                            Acknowledged:{" "}
                            {new Date(alert.acknowledged_at).toLocaleString()}
                          </span>
                        )}
                      </div>
                    </div>
                  </div>

                  <div className="flex items-center gap-2">
                    {alert.status === "active" && (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => onAcknowledgeAlert(alert.id)}
                      >
                        <CheckCircle className="mr-2 h-4 w-4" />
                        Acknowledge
                      </Button>
                    )}
                  </div>
                </div>

                {alert.metadata && Object.keys(alert.metadata).length > 0 && (
                  <div className="mt-3 pt-3 border-t">
                    <div className="text-xs text-muted-foreground">
                      <strong>Additional Details:</strong>
                      <div className="mt-1 space-y-1">
                        {Object.entries(alert.metadata).map(([key, value]) => (
                          <div key={key}>
                            <span className="capitalize">
                              {key.replace(/_/g, " ")}:
                            </span>{" "}
                            <span className="font-mono">{String(value)}</span>
                          </div>
                        ))}
                      </div>
                    </div>
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
