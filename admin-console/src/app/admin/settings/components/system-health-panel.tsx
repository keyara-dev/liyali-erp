"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import { Alert, AlertDescription } from "@/components/ui/alert";
import {
  CheckCircle,
  AlertTriangle,
  XCircle,
  RefreshCw,
  TrendingUp,
  Shield,
  Database,
  Zap,
  Settings,
  Info,
  Clock,
} from "lucide-react";
import { format } from "date-fns";
import { cn } from "@/lib/utils";
import type { SystemHealth } from "@/app/_actions/settings";

interface SystemHealthPanelProps {
  health: SystemHealth;
  onRefresh: () => void;
  isLoading?: boolean;
}

export function SystemHealthPanel({
  health,
  onRefresh,
  isLoading = false,
}: SystemHealthPanelProps) {
  const getStatusIcon = (status: "pass" | "fail" | "warning") => {
    switch (status) {
      case "pass":
        return <CheckCircle className="h-4 w-4 text-green-600" />;
      case "warning":
        return <AlertTriangle className="h-4 w-4 text-yellow-600" />;
      case "fail":
        return <XCircle className="h-4 w-4 text-red-600" />;
    }
  };

  const getStatusColor = (status: "pass" | "fail" | "warning") => {
    switch (status) {
      case "pass":
        return "text-green-600 bg-green-50 border-green-200";
      case "warning":
        return "text-yellow-600 bg-yellow-50 border-yellow-200";
      case "fail":
        return "text-red-600 bg-red-50 border-red-200";
    }
  };

  const getHealthColor = (status: "healthy" | "warning" | "critical") => {
    switch (status) {
      case "healthy":
        return "text-green-600";
      case "warning":
        return "text-yellow-600";
      case "critical":
        return "text-red-600";
    }
  };

  const getHealthBadgeVariant = (
    status: "healthy" | "warning" | "critical",
  ) => {
    switch (status) {
      case "healthy":
        return "default" as const;
      case "warning":
        return "secondary" as const;
      case "critical":
        return "destructive" as const;
    }
  };

  const getCheckIcon = (name: string) => {
    if (name.toLowerCase().includes("security")) return Shield;
    if (name.toLowerCase().includes("database")) return Database;
    if (name.toLowerCase().includes("performance")) return Zap;
    return Settings;
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Card>
          <CardHeader>
            <div className="h-6 bg-muted animate-pulse rounded w-48" />
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="h-4 bg-muted animate-pulse rounded w-32" />
              <div className="h-2 bg-muted animate-pulse rounded" />
              <div className="h-4 bg-muted animate-pulse rounded w-24" />
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Overall Health Status */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-lg font-medium">
            System Health Overview
          </CardTitle>
          <Button
            variant="outline"
            size="sm"
            onClick={onRefresh}
            disabled={isLoading}
          >
            <RefreshCw
              className={cn("h-4 w-4 mr-2", isLoading && "animate-spin")}
            />
            Refresh
          </Button>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-3">
                <TrendingUp
                  className={cn("h-8 w-8", getHealthColor(health.status))}
                />
                <div>
                  <div className="flex items-center space-x-2">
                    <h3 className="text-2xl font-bold">System Health</h3>
                    <Badge variant={getHealthBadgeVariant(health.status)}>
                      {health.status.charAt(0).toUpperCase() +
                        health.status.slice(1)}
                    </Badge>
                  </div>
                  <p className="text-muted-foreground">
                    Overall configuration health score
                  </p>
                </div>
              </div>
              <div className="text-right">
                <div
                  className={cn(
                    "text-3xl font-bold",
                    getHealthColor(health.status),
                  )}
                >
                  {health.score}%
                </div>
                <p className="text-sm text-muted-foreground">Health Score</p>
              </div>
            </div>

            <Progress
              value={health.score}
              className={cn(
                "h-3",
                health.score >= 90 && "bg-green-100",
                health.score >= 70 && health.score < 90 && "bg-yellow-100",
                health.score < 70 && "bg-red-100",
              )}
            />
          </div>
        </CardContent>
      </Card>

      {/* Health Checks */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg font-medium">Health Checks</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {health.checks.map((check, index) => {
              const CheckIcon = getCheckIcon(check.name);

              return (
                <div
                  key={index}
                  className={cn(
                    "flex items-start space-x-3 p-3 rounded-lg border",
                    getStatusColor(check.status),
                  )}
                >
                  <div className="flex items-center space-x-2 mt-0.5">
                    {getStatusIcon(check.status)}
                    <CheckIcon className="h-4 w-4" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between">
                      <h4 className="font-medium">{check.name}</h4>
                      <div className="flex items-center text-xs text-muted-foreground">
                        <Clock className="h-3 w-3 mr-1" />
                        {format(new Date(check.lastChecked), "MMM dd, HH:mm")}
                      </div>
                    </div>
                    <p className="text-sm mt-1">{check.message}</p>
                  </div>
                </div>
              );
            })}
          </div>
        </CardContent>
      </Card>

      {/* Recommendations */}
      {health.recommendations.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg font-medium flex items-center gap-2">
              <Info className="h-5 w-5" />
              Recommendations
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {health.recommendations.map((recommendation, index) => (
                <Alert key={index}>
                  <Info className="h-4 w-4" />
                  <AlertDescription>{recommendation}</AlertDescription>
                </Alert>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Health Summary Stats */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center space-x-2">
              <CheckCircle className="h-5 w-5 text-green-600" />
              <div>
                <p className="text-2xl font-bold text-green-600">
                  {health.checks.filter((c) => c.status === "pass").length}
                </p>
                <p className="text-xs text-muted-foreground">Passing Checks</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center space-x-2">
              <AlertTriangle className="h-5 w-5 text-yellow-600" />
              <div>
                <p className="text-2xl font-bold text-yellow-600">
                  {health.checks.filter((c) => c.status === "warning").length}
                </p>
                <p className="text-xs text-muted-foreground">Warnings</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center space-x-2">
              <XCircle className="h-5 w-5 text-red-600" />
              <div>
                <p className="text-2xl font-bold text-red-600">
                  {health.checks.filter((c) => c.status === "fail").length}
                </p>
                <p className="text-xs text-muted-foreground">Failed Checks</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
