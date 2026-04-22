"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import {
  AlertTriangle,
  Shield,
  Lock,
  Eye,
  Activity,
  TrendingUp,
  TrendingDown,
  Clock,
} from "lucide-react";
import { type AuditLogStats } from "@/app/_actions/audit-logs";

interface SecurityEventsPanelProps {
  stats: AuditLogStats | null;
  isLoading?: boolean;
}

export function SecurityEventsPanel({
  stats,
  isLoading,
}: SecurityEventsPanelProps) {
  if (isLoading) {
    return (
      <div className="grid gap-4 md:grid-cols-2">
        {[...Array(4)].map((_, i) => (
          <Card key={i}>
            <CardHeader>
              <div className="h-6 w-32 bg-muted animate-pulse rounded" />
            </CardHeader>
            <CardContent>
              <div className="h-32 bg-muted animate-pulse rounded" />
            </CardContent>
          </Card>
        ))}
      </div>
    );
  }

  if (!stats) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center h-64">
          <p className="text-muted-foreground">
            No security events data available
          </p>
        </CardContent>
      </Card>
    );
  }

  const getThreatLevel = () => {
    const totalThreats =
      stats.security_events.failed_logins +
      stats.security_events.suspicious_activities +
      stats.security_events.policy_violations +
      stats.security_events.unauthorized_access_attempts;

    if (totalThreats === 0)
      return {
        level: "Low",
        color: "text-green-600",
        variant: "default" as const,
      };
    if (totalThreats < 10)
      return {
        level: "Medium",
        color: "text-yellow-600",
        variant: "secondary" as const,
      };
    if (totalThreats < 50)
      return {
        level: "High",
        color: "text-orange-600",
        variant: "destructive" as const,
      };
    return {
      level: "Critical",
      color: "text-red-600",
      variant: "destructive" as const,
    };
  };

  const threatLevel = getThreatLevel();

  const securityEvents = [
    {
      title: "Failed Logins",
      count: stats.security_events.failed_logins,
      icon: Lock,
      color: "text-red-600",
      description: "Authentication failures",
      severity:
        stats.security_events.failed_logins > 20
          ? "high"
          : stats.security_events.failed_logins > 5
            ? "medium"
            : "low",
    },
    {
      title: "Suspicious Activities",
      count: stats.security_events.suspicious_activities,
      icon: Eye,
      color: "text-orange-600",
      description: "Anomalous behavior detected",
      severity:
        stats.security_events.suspicious_activities > 10
          ? "high"
          : stats.security_events.suspicious_activities > 3
            ? "medium"
            : "low",
    },
    {
      title: "Policy Violations",
      count: stats.security_events.policy_violations,
      icon: Shield,
      color: "text-yellow-600",
      description: "Compliance policy breaches",
      severity:
        stats.security_events.policy_violations > 15
          ? "high"
          : stats.security_events.policy_violations > 5
            ? "medium"
            : "low",
    },
    {
      title: "Unauthorized Access",
      count: stats.security_events.unauthorized_access_attempts,
      icon: AlertTriangle,
      color: "text-red-600",
      description: "Access attempts without permission",
      severity:
        stats.security_events.unauthorized_access_attempts > 5
          ? "high"
          : stats.security_events.unauthorized_access_attempts > 1
            ? "medium"
            : "low",
    },
  ];

  const getSeverityBadge = (severity: string) => {
    switch (severity) {
      case "high":
        return (
          <Badge variant="destructive" className="text-xs">
            High Risk
          </Badge>
        );
      case "medium":
        return (
          <Badge variant="secondary" className="text-xs">
            Medium Risk
          </Badge>
        );
      case "low":
        return (
          <Badge variant="outline" className="text-xs">
            Low Risk
          </Badge>
        );
      default:
        return (
          <Badge variant="outline" className="text-xs">
            Normal
          </Badge>
        );
    }
  };

  return (
    <div className="space-y-4">
      {/* Security Overview */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Shield className="h-5 w-5" />
            Security Overview
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between mb-4">
            <div>
              <div className="text-sm text-muted-foreground">
                Current Threat Level
              </div>
              <div className={`text-2xl font-bold ${threatLevel.color}`}>
                {threatLevel.level}
              </div>
            </div>
            <Badge variant={threatLevel.variant} className="text-sm">
              {threatLevel.level} Risk
            </Badge>
          </div>

          <div className="space-y-3">
            <div className="flex items-center justify-between text-sm">
              <span>Security Score</span>
              <span className="font-medium">
                {Math.max(
                  100 -
                    (stats.security_events.failed_logins +
                      stats.security_events.suspicious_activities +
                      stats.security_events.policy_violations +
                      stats.security_events.unauthorized_access_attempts),
                  0,
                )}
                %
              </span>
            </div>
            <Progress
              value={Math.max(
                100 -
                  (stats.security_events.failed_logins +
                    stats.security_events.suspicious_activities +
                    stats.security_events.policy_violations +
                    stats.security_events.unauthorized_access_attempts),
                0,
              )}
              className="h-2"
            />
          </div>
        </CardContent>
      </Card>

      {/* Security Events Grid */}
      <div className="grid gap-4 md:grid-cols-2">
        {securityEvents.map((event, index) => (
          <Card key={`${event.title}-${index}`}>
            <CardHeader className="pb-3">
              <CardTitle className="text-sm flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <event.icon className={`h-4 w-4 ${event.color}`} />
                  {event.title}
                </div>
                {getSeverityBadge(event.severity)}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <span className={`text-2xl font-bold ${event.color}`}>
                    {event.count}
                  </span>
                  <div className="text-right">
                    <div className="text-xs text-muted-foreground">
                      {event.description}
                    </div>
                  </div>
                </div>

                {/* Trend indicator */}
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  {event.count > 0 ? (
                    <>
                      <TrendingUp className="h-3 w-3 text-red-500" />
                      <span>Requires attention</span>
                    </>
                  ) : (
                    <>
                      <TrendingDown className="h-3 w-3 text-green-500" />
                      <span>All clear</span>
                    </>
                  )}
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Activity Timeline */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Activity className="h-5 w-5" />
            Activity Timeline (Last 24 Hours)
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {stats.activity_by_hour.slice(-6).map((activity, index) => (
              <div
                key={activity.hour ?? `activity-hour-${index}`}
                className="flex items-center justify-between p-3 rounded-lg bg-muted/20"
              >
                <div className="flex items-center gap-3">
                  <Clock className="h-4 w-4 text-muted-foreground" />
                  <div>
                    <div className="font-medium text-sm">
                      {activity.hour}:00
                    </div>
                    <div className="text-xs text-muted-foreground">
                      {activity.count} total events
                    </div>
                  </div>
                </div>
                <div className="text-right">
                  <div className="font-medium text-sm">
                    {activity.count - activity.failed_count} success
                  </div>
                  {activity.failed_count > 0 && (
                    <div className="text-xs text-red-600">
                      {activity.failed_count} failed
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Quick Actions */}
      <Card>
        <CardHeader>
          <CardTitle>Security Actions</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-2 md:grid-cols-2">
            <Button variant="outline" size="sm" className="justify-start">
              <Shield className="mr-2 h-4 w-4" />
              Review Security Policies
            </Button>
            <Button variant="outline" size="sm" className="justify-start">
              <AlertTriangle className="mr-2 h-4 w-4" />
              Investigate Threats
            </Button>
            <Button variant="outline" size="sm" className="justify-start">
              <Lock className="mr-2 h-4 w-4" />
              Update Access Controls
            </Button>
            <Button variant="outline" size="sm" className="justify-start">
              <Eye className="mr-2 h-4 w-4" />
              Monitor Activities
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
