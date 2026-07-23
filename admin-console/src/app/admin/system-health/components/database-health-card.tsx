"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { Button } from "@/components/ui/button";
import {
  Database,
  CheckCircle,
  AlertTriangle,
  XCircle,
  Activity,
  HardDrive,
  Clock,
  Zap,
} from "lucide-react";

interface DatabaseHealthCardProps {
  health?: {
    status: "healthy" | "warning" | "critical";
    connection_count: number;
    query_performance: number;
    storage_usage: number;
    last_backup: string;
  };
  metrics?: {
    active_connections: number;
    slow_queries: number;
    cache_hit_ratio: number;
    storage_size: string;
    backup_status: "success" | "failed" | "in_progress";
  };
}

export function DatabaseHealthCard({
  health,
  metrics,
}: DatabaseHealthCardProps) {
  const getStatusBadge = (status: string) => {
    switch (status) {
      case "healthy":
        return (
          <Badge variant="default" className="bg-green-100 text-green-800">
            <CheckCircle className="mr-1 h-3 w-3" />
            Healthy
          </Badge>
        );
      case "warning":
        return (
          <Badge variant="secondary" className="bg-yellow-100 text-yellow-800">
            <AlertTriangle className="mr-1 h-3 w-3" />
            Warning
          </Badge>
        );
      case "critical":
        return (
          <Badge variant="destructive">
            <XCircle className="mr-1 h-3 w-3" />
            Critical
          </Badge>
        );
      default:
        return <Badge variant="outline">{status}</Badge>;
    }
  };

  const getBackupStatusBadge = (status: string) => {
    switch (status) {
      case "success":
        return (
          <Badge variant="default" className="bg-green-100 text-green-800">
            <CheckCircle className="mr-1 h-3 w-3" />
            Success
          </Badge>
        );
      case "failed":
        return (
          <Badge variant="destructive">
            <XCircle className="mr-1 h-3 w-3" />
            Failed
          </Badge>
        );
      case "in_progress":
        return (
          <Badge variant="secondary">
            <Activity className="mr-1 h-3 w-3" />
            In Progress
          </Badge>
        );
      default:
        return <Badge variant="outline">{status}</Badge>;
    }
  };

  const getStorageUsageColor = (usage: number) => {
    if (usage >= 90) return "bg-red-500";
    if (usage >= 75) return "bg-yellow-500";
    return "bg-green-500";
  };

  const getCacheHitRatioColor = (ratio: number) => {
    if (ratio >= 90) return "text-green-600";
    if (ratio >= 75) return "text-yellow-600";
    return "text-red-600";
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Database className="h-4 w-4" />
            Database Health
          </div>
          {health && getStatusBadge(health.status)}
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Connection Status */}
        <div className="space-y-2">
          <div className="flex items-center justify-between text-sm">
            <span className="flex items-center gap-2">
              <Activity className="h-3 w-3" />
              Active Connections
            </span>
            <span className="font-medium">
              {metrics?.active_connections || health?.connection_count || 0}
            </span>
          </div>
        </div>

        {/* Storage Usage */}
        {health?.storage_usage !== undefined && (
          <div className="space-y-2">
            <div className="flex items-center justify-between text-sm">
              <span className="flex items-center gap-2">
                <HardDrive className="h-3 w-3" />
                Storage Usage
              </span>
              <span className="font-medium">{health.storage_usage}%</span>
            </div>
            <Progress value={health.storage_usage} className="h-2" />
            {metrics?.storage_size && (
              <div className="text-xs text-muted-foreground">
                Total Size: {metrics.storage_size}
              </div>
            )}
          </div>
        )}

        {/* Query Performance */}
        {health?.query_performance !== undefined && (
          <div className="space-y-2">
            <div className="flex items-center justify-between text-sm">
              <span className="flex items-center gap-2">
                <Zap className="h-3 w-3" />
                Avg Query Time
              </span>
              <span className="font-medium">{health.query_performance}ms</span>
            </div>
            {metrics?.slow_queries !== undefined && (
              <div className="text-xs text-muted-foreground">
                Slow Queries: {metrics.slow_queries}
              </div>
            )}
          </div>
        )}

        {/* Cache Hit Ratio */}
        {metrics?.cache_hit_ratio !== undefined && (
          <div className="space-y-2">
            <div className="flex items-center justify-between text-sm">
              <span>Cache Hit Ratio</span>
              <span
                className={`font-medium ${getCacheHitRatioColor(metrics.cache_hit_ratio)}`}
              >
                {metrics.cache_hit_ratio}%
              </span>
            </div>
            <Progress value={metrics.cache_hit_ratio} className="h-2" />
          </div>
        )}

        {/* Backup Status */}
        <div className="space-y-2">
          <div className="flex items-center justify-between text-sm">
            <span className="flex items-center gap-2">
              <Clock className="h-3 w-3" />
              Last Backup
            </span>
            <div className="text-right">
              {metrics?.backup_status &&
                getBackupStatusBadge(metrics.backup_status)}
              {health?.last_backup && (
                <div className="text-xs text-muted-foreground mt-1">
                  {new Date(health.last_backup).toLocaleString()}
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="flex gap-2 pt-2">
          <Button variant="outline" size="sm" className="flex-1">
            <Database className="mr-2 h-4 w-4" />
            View Details
          </Button>
          <Button variant="outline" size="sm" className="flex-1">
            <Activity className="mr-2 h-4 w-4" />
            Run Diagnostics
          </Button>
        </div>

        {/* Health Indicators */}
        <div className="grid grid-cols-2 gap-4 pt-2 border-t text-xs">
          <div className="text-center">
            <div className="text-muted-foreground">Connections</div>
            <div
              className={`font-medium ${
                (health?.connection_count || 0) > 100
                  ? "text-yellow-600"
                  : "text-green-600"
              }`}
            >
              {health?.connection_count || metrics?.active_connections || 0}
            </div>
          </div>
          <div className="text-center">
            <div className="text-muted-foreground">Performance</div>
            <div
              className={`font-medium ${
                (health?.query_performance || 0) > 1000
                  ? "text-red-600"
                  : (health?.query_performance || 0) > 500
                    ? "text-yellow-600"
                    : "text-green-600"
              }`}
            >
              {health?.query_performance || 0}ms
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
