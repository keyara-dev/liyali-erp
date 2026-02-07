"use client";

import { useState, useEffect } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import {
  Activity,
  Server,
  Database,
  Zap,
  AlertTriangle,
  CheckCircle,
  XCircle,
  Clock,
  TrendingUp,
  TrendingDown,
  RefreshCw,
  Bell,
  Settings,
} from "lucide-react";
import { toast } from "sonner";
import {
  getSystemHealth,
  getSystemMetrics,
  getSystemAlerts,
  acknowledgeAlert,
  type SystemHealth,
  type SystemMetrics,
  type SystemAlert,
} from "@/app/_actions/system-health";
import { SystemMetricsChart } from "./components/system-metrics-chart";
import { SystemAlertsPanel } from "./components/system-alerts-panel";
import { DatabaseHealthCard } from "./components/database-health-card";
import { APIHealthCard } from "./components/api-health-card";

export default function SystemHealthPage() {
  const [systemHealth, setSystemHealth] = useState<SystemHealth | null>(null);
  const [systemMetrics, setSystemMetrics] = useState<SystemMetrics | null>(
    null,
  );
  const [systemAlerts, setSystemAlerts] = useState<SystemAlert[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [autoRefresh, setAutoRefresh] = useState(true);

  useEffect(() => {
    loadSystemData();

    // Set up auto-refresh every 30 seconds
    let interval: NodeJS.Timeout;
    if (autoRefresh) {
      interval = setInterval(() => {
        loadSystemData(true);
      }, 30000);
    }

    return () => {
      if (interval) clearInterval(interval);
    };
  }, [autoRefresh]);

  const loadSystemData = async (isAutoRefresh = false) => {
    if (!isAutoRefresh) {
      setIsLoading(true);
    } else {
      setIsRefreshing(true);
    }

    try {
      // Load system health
      const healthResult = await getSystemHealth();
      if (healthResult.success && healthResult.data) {
        setSystemHealth(healthResult.data);
      }

      // Load system metrics
      const metricsResult = await getSystemMetrics();
      if (metricsResult.success && metricsResult.data) {
        setSystemMetrics(metricsResult.data);
      }

      // Load system alerts
      const alertsResult = await getSystemAlerts();
      if (alertsResult.success && alertsResult.data) {
        setSystemAlerts(
          Array.isArray(alertsResult.data) ? alertsResult.data : [],
        );
      } else {
        setSystemAlerts([]);
      }
    } catch (error) {
      if (!isAutoRefresh) {
        toast.error("Failed to load system data");
      }
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  };

  const handleRefresh = () => {
    loadSystemData();
  };

  const handleAcknowledgeAlert = async (alertId: string) => {
    try {
      const result = await acknowledgeAlert(alertId);
      if (result.success) {
        toast.success("Alert acknowledged");
        loadSystemData(true);
      } else {
        toast.error("Failed to acknowledge alert");
      }
    } catch (error) {
      toast.error("Failed to acknowledge alert");
    }
  };

  const getHealthStatusBadge = (status: string) => {
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

  const getUptimeColor = (uptime: number) => {
    if (uptime >= 99.9) return "text-green-600";
    if (uptime >= 99.0) return "text-yellow-600";
    return "text-red-600";
  };

  if (isLoading) {
    return <div>Loading system health data...</div>;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">System Health</h1>
          <p className="text-muted-foreground">
            Monitor system performance, health, and alerts
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setAutoRefresh(!autoRefresh)}
          >
            <Bell className="mr-2 h-4 w-4" />
            Auto Refresh: {autoRefresh ? "On" : "Off"}
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={handleRefresh}
            disabled={isRefreshing}
          >
            <RefreshCw
              className={`mr-2 h-4 w-4 ${isRefreshing ? "animate-spin" : ""}`}
            />
            Refresh
          </Button>
        </div>
      </div>

      {/* System Overview Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">System Status</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="flex items-center space-x-2">
              {systemHealth &&
                getHealthStatusBadge(systemHealth.overall_status)}
            </div>
            <p className="text-xs text-muted-foreground mt-2">
              Last updated: {new Date().toLocaleTimeString()}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">System Uptime</CardTitle>
            <Clock className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div
              className={`text-2xl font-bold ${getUptimeColor(systemHealth?.uptime_percentage || 0)}`}
            >
              {systemHealth?.uptime_percentage?.toFixed(2) || 0}%
            </div>
            <p className="text-xs text-muted-foreground">
              {systemHealth?.uptime_duration || "N/A"}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Active Alerts</CardTitle>
            <AlertTriangle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {Array.isArray(systemAlerts)
                ? systemAlerts.filter((alert) => alert.status === "active")
                    .length
                : 0}
            </div>
            <p className="text-xs text-muted-foreground">
              {Array.isArray(systemAlerts)
                ? systemAlerts.filter((alert) => alert.severity === "critical")
                    .length
                : 0}{" "}
              critical
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Response Time</CardTitle>
            <Zap className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {systemMetrics?.average_response_time || 0}ms
            </div>
            <div className="flex items-center text-xs text-muted-foreground">
              {systemMetrics?.response_time_trend === "up" ? (
                <TrendingUp className="mr-1 h-3 w-3 text-red-500" />
              ) : (
                <TrendingDown className="mr-1 h-3 w-3 text-green-500" />
              )}
              vs last hour
            </div>
          </CardContent>
        </Card>
      </div>

      {/* System Components Health */}
      <div className="grid gap-4 md:grid-cols-2">
        <DatabaseHealthCard
          health={systemHealth?.database}
          metrics={systemMetrics?.database}
        />
        <APIHealthCard
          health={systemHealth?.api}
          metrics={systemMetrics?.api}
        />
      </div>

      {/* Performance Metrics */}
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Server className="h-4 w-4" />
              Server Resources
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <div className="flex items-center justify-between text-sm">
                <span>CPU Usage</span>
                <span>{systemMetrics?.server?.cpu_usage || 0}%</span>
              </div>
              <Progress value={systemMetrics?.server?.cpu_usage || 0} />
            </div>

            <div className="space-y-2">
              <div className="flex items-center justify-between text-sm">
                <span>Memory Usage</span>
                <span>{systemMetrics?.server?.memory_usage || 0}%</span>
              </div>
              <Progress value={systemMetrics?.server?.memory_usage || 0} />
            </div>

            <div className="space-y-2">
              <div className="flex items-center justify-between text-sm">
                <span>Disk Usage</span>
                <span>{systemMetrics?.server?.disk_usage || 0}%</span>
              </div>
              <Progress value={systemMetrics?.server?.disk_usage || 0} />
            </div>

            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <span className="text-muted-foreground">Load Average:</span>
                <div className="font-medium">
                  {systemMetrics?.server?.load_average || "N/A"}
                </div>
              </div>
              <div>
                <span className="text-muted-foreground">
                  Active Connections:
                </span>
                <div className="font-medium">
                  {systemMetrics?.server?.active_connections || 0}
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        <SystemMetricsChart metrics={systemMetrics} />
      </div>

      {/* System Alerts */}
      <SystemAlertsPanel
        alerts={systemAlerts}
        onAcknowledgeAlert={handleAcknowledgeAlert}
        onRefresh={() => loadSystemData(true)}
      />
    </div>
  );
}
