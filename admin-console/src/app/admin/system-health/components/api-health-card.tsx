"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { Button } from "@/components/ui/button";
import {
  Zap,
  CheckCircle,
  AlertTriangle,
  XCircle,
  Activity,
  Users,
  TrendingUp,
  TrendingDown,
  Clock,
} from "lucide-react";

interface APIHealthCardProps {
  health?: {
    status: "healthy" | "warning" | "critical";
    response_time: number;
    error_rate: number;
    requests_per_minute: number;
    active_sessions: number;
  };
  metrics?: {
    total_requests: number;
    successful_requests: number;
    failed_requests: number;
    average_response_time: number;
    peak_response_time: number;
  };
}

export function APIHealthCard({ health, metrics }: APIHealthCardProps) {
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

  const getErrorRateColor = (rate: number) => {
    if (rate >= 5) return "text-red-600";
    if (rate >= 2) return "text-yellow-600";
    return "text-green-600";
  };

  const getResponseTimeColor = (time: number) => {
    if (time >= 1000) return "text-red-600";
    if (time >= 500) return "text-yellow-600";
    return "text-green-600";
  };

  const calculateSuccessRate = () => {
    if (!metrics || !metrics.total_requests) return 0;
    return (metrics.successful_requests / metrics.total_requests) * 100;
  };

  const successRate = calculateSuccessRate();

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Zap className="h-4 w-4" />
            API Health
          </div>
          {health && getStatusBadge(health.status)}
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Response Time */}
        <div className="space-y-2">
          <div className="flex items-center justify-between text-sm">
            <span className="flex items-center gap-2">
              <Clock className="h-3 w-3" />
              Response Time
            </span>
            <span
              className={`font-medium ${getResponseTimeColor(
                health?.response_time || metrics?.average_response_time || 0,
              )}`}
            >
              {health?.response_time || metrics?.average_response_time || 0}ms
            </span>
          </div>
          {metrics?.peak_response_time && (
            <div className="text-xs text-muted-foreground">
              Peak: {metrics.peak_response_time}ms
            </div>
          )}
        </div>

        {/* Error Rate */}
        {health?.error_rate !== undefined && (
          <div className="space-y-2">
            <div className="flex items-center justify-between text-sm">
              <span className="flex items-center gap-2">
                <AlertTriangle className="h-3 w-3" />
                Error Rate
              </span>
              <span
                className={`font-medium ${getErrorRateColor(health.error_rate)}`}
              >
                {health.error_rate}%
              </span>
            </div>
            <Progress value={health.error_rate} className="h-2" />
          </div>
        )}

        {/* Success Rate */}
        {metrics && (
          <div className="space-y-2">
            <div className="flex items-center justify-between text-sm">
              <span className="flex items-center gap-2">
                <CheckCircle className="h-3 w-3" />
                Success Rate
              </span>
              <span
                className={`font-medium ${
                  successRate >= 99
                    ? "text-green-600"
                    : successRate >= 95
                      ? "text-yellow-600"
                      : "text-red-600"
                }`}
              >
                {successRate.toFixed(1)}%
              </span>
            </div>
            <Progress value={successRate} className="h-2" />
          </div>
        )}

        {/* Request Volume */}
        <div className="space-y-2">
          <div className="flex items-center justify-between text-sm">
            <span className="flex items-center gap-2">
              <Activity className="h-3 w-3" />
              Requests/Min
            </span>
            <span className="font-medium">
              {health?.requests_per_minute || 0}
            </span>
          </div>
          {metrics?.total_requests && (
            <div className="text-xs text-muted-foreground">
              Total Today: {metrics.total_requests.toLocaleString()}
            </div>
          )}
        </div>

        {/* Active Sessions */}
        {health?.active_sessions !== undefined && (
          <div className="space-y-2">
            <div className="flex items-center justify-between text-sm">
              <span className="flex items-center gap-2">
                <Users className="h-3 w-3" />
                Active Sessions
              </span>
              <span className="font-medium">{health.active_sessions}</span>
            </div>
          </div>
        )}

        {/* Request Breakdown */}
        {metrics && (
          <div className="space-y-2">
            <div className="text-sm font-medium">Request Breakdown</div>
            <div className="grid grid-cols-2 gap-2 text-xs">
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Successful:</span>
                <span className="text-green-600 font-medium">
                  {metrics.successful_requests.toLocaleString()}
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Failed:</span>
                <span className="text-red-600 font-medium">
                  {metrics.failed_requests.toLocaleString()}
                </span>
              </div>
            </div>
          </div>
        )}

        {/* Action Buttons */}
        <div className="flex gap-2 pt-2">
          <Button variant="outline" size="sm" className="flex-1">
            <Zap className="mr-2 h-4 w-4" />
            View Logs
          </Button>
          <Button variant="outline" size="sm" className="flex-1">
            <Activity className="mr-2 h-4 w-4" />
            Test Endpoints
          </Button>
        </div>

        {/* Health Indicators */}
        <div className="grid grid-cols-3 gap-4 pt-2 border-t text-xs">
          <div className="text-center">
            <div className="text-muted-foreground">Avg Response</div>
            <div
              className={`font-medium ${getResponseTimeColor(
                health?.response_time || metrics?.average_response_time || 0,
              )}`}
            >
              {health?.response_time || metrics?.average_response_time || 0}ms
            </div>
          </div>
          <div className="text-center">
            <div className="text-muted-foreground">Error Rate</div>
            <div
              className={`font-medium ${getErrorRateColor(health?.error_rate || 0)}`}
            >
              {health?.error_rate || 0}%
            </div>
          </div>
          <div className="text-center">
            <div className="text-muted-foreground">Sessions</div>
            <div className="font-medium text-blue-600">
              {health?.active_sessions || 0}
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
