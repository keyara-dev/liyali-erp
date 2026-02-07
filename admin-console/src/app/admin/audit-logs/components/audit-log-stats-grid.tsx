"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Shield,
  Activity,
  AlertTriangle,
  Users,
  Clock,
  TrendingUp,
  Eye,
  Lock,
} from "lucide-react";
import { type AuditLogStats } from "@/app/_actions/audit-logs";

interface AuditLogStatsGridProps {
  stats: AuditLogStats | null;
  isLoading?: boolean;
}

export function AuditLogStatsGrid({
  stats,
  isLoading,
}: AuditLogStatsGridProps) {
  const formatNumber = (num: number) => {
    if (num >= 1000000) {
      return (num / 1000000).toFixed(1) + "M";
    }
    if (num >= 1000) {
      return (num / 1000).toFixed(1) + "K";
    }
    return num.toString();
  };

  const getSecurityScore = () => {
    if (!stats) return 0;
    const totalEvents =
      stats.security_events.failed_logins +
      stats.security_events.suspicious_activities +
      stats.security_events.policy_violations +
      stats.security_events.unauthorized_access_attempts;

    if (totalEvents === 0) return 100;
    if (totalEvents < 10) return 85;
    if (totalEvents < 50) return 70;
    if (totalEvents < 100) return 50;
    return 25;
  };

  const getSecurityScoreColor = (score: number) => {
    if (score >= 80) return "text-green-600";
    if (score >= 60) return "text-yellow-600";
    return "text-red-600";
  };

  if (isLoading) {
    return (
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {[...Array(8)].map((_, i) => (
          <Card key={i}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                <div className="h-4 w-24 bg-muted animate-pulse rounded" />
              </CardTitle>
              <div className="h-4 w-4 bg-muted animate-pulse rounded" />
            </CardHeader>
            <CardContent>
              <div className="h-8 w-16 bg-muted animate-pulse rounded mb-2" />
              <div className="h-3 w-20 bg-muted animate-pulse rounded" />
            </CardContent>
          </Card>
        ))}
      </div>
    );
  }

  if (!stats) {
    return (
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardContent className="flex items-center justify-center h-32">
            <p className="text-muted-foreground">No statistics available</p>
          </CardContent>
        </Card>
      </div>
    );
  }

  const securityScore = getSecurityScore();

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      {/* Total Logs */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Total Logs</CardTitle>
          <Activity className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">
            {formatNumber(stats.total_logs)}
          </div>
          <div className="flex items-center text-xs text-muted-foreground">
            <TrendingUp className="h-3 w-3 mr-1" />
            <span>{formatNumber(stats.logs_today)} today</span>
          </div>
        </CardContent>
      </Card>

      {/* Failed Actions */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Failed Actions</CardTitle>
          <AlertTriangle className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold text-red-600">
            {formatNumber(stats.failed_actions)}
          </div>
          <div className="flex items-center text-xs text-muted-foreground">
            <span>
              {((stats.failed_actions / stats.total_logs) * 100).toFixed(1)}%
              failure rate
            </span>
          </div>
        </CardContent>
      </Card>

      {/* Critical Events */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Critical Events</CardTitle>
          <Shield className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold text-orange-600">
            {formatNumber(stats.critical_events)}
          </div>
          <div className="flex items-center text-xs text-muted-foreground">
            <Badge
              variant={stats.critical_events > 10 ? "destructive" : "default"}
              className="text-xs"
            >
              {stats.critical_events > 10 ? "High Alert" : "Normal"}
            </Badge>
          </div>
        </CardContent>
      </Card>

      {/* Unique Users */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Active Users</CardTitle>
          <Users className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold">
            {formatNumber(stats.unique_users)}
          </div>
          <div className="flex items-center text-xs text-muted-foreground">
            <Eye className="h-3 w-3 mr-1" />
            <span>Unique users active</span>
          </div>
        </CardContent>
      </Card>

      {/* Security Score */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Security Score</CardTitle>
          <Lock className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div
            className={`text-2xl font-bold ${getSecurityScoreColor(securityScore)}`}
          >
            {securityScore}%
          </div>
          <div className="flex items-center text-xs text-muted-foreground">
            <Badge
              variant={
                securityScore >= 80
                  ? "default"
                  : securityScore >= 60
                    ? "secondary"
                    : "destructive"
              }
              className="text-xs"
            >
              {securityScore >= 80
                ? "Excellent"
                : securityScore >= 60
                  ? "Good"
                  : "Needs Attention"}
            </Badge>
          </div>
        </CardContent>
      </Card>

      {/* Failed Logins */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">Failed Logins</CardTitle>
          <AlertTriangle className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold text-red-600">
            {formatNumber(stats.security_events.failed_logins)}
          </div>
          <div className="flex items-center text-xs text-muted-foreground">
            <span>Authentication failures</span>
          </div>
        </CardContent>
      </Card>

      {/* Suspicious Activities */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">
            Suspicious Activities
          </CardTitle>
          <Shield className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold text-orange-600">
            {formatNumber(stats.security_events.suspicious_activities)}
          </div>
          <div className="flex items-center text-xs text-muted-foreground">
            <span>Anomalous behavior detected</span>
          </div>
        </CardContent>
      </Card>

      {/* Policy Violations */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle className="text-sm font-medium">
            Policy Violations
          </CardTitle>
          <AlertTriangle className="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div className="text-2xl font-bold text-yellow-600">
            {formatNumber(stats.security_events.policy_violations)}
          </div>
          <div className="flex items-center text-xs text-muted-foreground">
            <span>Compliance issues</span>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
