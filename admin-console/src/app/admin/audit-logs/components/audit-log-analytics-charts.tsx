"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  BarChart,
  Bar,
  LineChart,
  Line,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from "recharts";
import { Users, Globe, Monitor, Activity } from "lucide-react";
import {
  type AuditLogAnalytics,
  type AuditLogStats,
} from "@/app/_actions/audit-logs";

interface AuditLogAnalyticsChartsProps {
  analytics: AuditLogAnalytics | null;
  stats: AuditLogStats | null;
  isLoading?: boolean;
}

const COLORS = ["#0088FE", "#00C49F", "#FFBB28", "#FF8042", "#8884D8"];

export function AuditLogAnalyticsCharts({
  analytics,
  stats,
  isLoading,
}: AuditLogAnalyticsChartsProps) {
  if (isLoading) {
    return (
      <div className="grid gap-4 md:grid-cols-2">
        {[...Array(4)].map((_, i) => (
          <Card key={i}>
            <CardHeader>
              <div className="h-6 w-32 bg-muted animate-pulse rounded" />
            </CardHeader>
            <CardContent>
              <div className="h-64 bg-muted animate-pulse rounded" />
            </CardContent>
          </Card>
        ))}
      </div>
    );
  }

  if (!analytics || !stats) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center h-64">
          <p className="text-muted-foreground">No analytics data available</p>
        </CardContent>
      </Card>
    );
  }

  const CustomTooltip = ({ active, payload, label }: any) => {
    if (active && payload && payload.length) {
      return (
        <div className="bg-background border rounded-lg p-3 shadow-lg">
          <p className="font-medium">{`${label}`}</p>
          {payload.map((entry: any, index: number) => (
            <p key={index} style={{ color: entry.color }}>
              {`${entry.name}: ${entry.value}`}
            </p>
          ))}
        </div>
      );
    }
    return null;
  };

  return (
    <div className="grid gap-4 md:grid-cols-2">
      {/* Top Actions Chart */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Activity className="h-4 w-4" />
            Top Actions
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={(stats.top_actions ?? []).slice(0, 8)}>
                <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                <XAxis
                  dataKey="action"
                  className="text-xs"
                  tick={{ fontSize: 10 }}
                  angle={-45}
                  textAnchor="end"
                  height={80}
                />
                <YAxis className="text-xs" tick={{ fontSize: 12 }} />
                <Tooltip content={<CustomTooltip />} />
                <Bar dataKey="count" fill="#8884d8" />
              </BarChart>
            </ResponsiveContainer>
          </div>
          <div className="mt-4 space-y-2">
            {(stats.top_actions ?? []).slice(0, 3).map((action, index) => (
              <div
                key={action.action}
                className="flex items-center justify-between text-sm"
              >
                <span className="capitalize">{action.action}</span>
                <div className="flex items-center gap-2">
                  <span className="font-medium">{action.count}</span>
                  <Badge variant="outline" className="text-xs">
                    {(action.percentage ?? 0).toFixed(1)}%
                  </Badge>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* User Activity Chart */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Users className="h-4 w-4" />
            Top Active Users
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={(analytics.user_activity ?? []).slice(0, 8)}>
                <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                <XAxis
                  dataKey="user_name"
                  className="text-xs"
                  tick={{ fontSize: 10 }}
                  angle={-45}
                  textAnchor="end"
                  height={80}
                />
                <YAxis className="text-xs" tick={{ fontSize: 12 }} />
                <Tooltip content={<CustomTooltip />} />
                <Bar dataKey="action_count" fill="#82ca9d" />
              </BarChart>
            </ResponsiveContainer>
          </div>
          <div className="mt-4 space-y-2">
            {(analytics.user_activity ?? []).slice(0, 3).map((user, index) => (
              <div
                key={user.user_id}
                className="flex items-center justify-between text-sm"
              >
                <span className="truncate">{user.user_name}</span>
                <div className="flex items-center gap-2">
                  <span className="font-medium">{user.action_count}</span>
                  <Badge
                    variant={
                      user.risk_score > 70
                        ? "destructive"
                        : user.risk_score > 40
                          ? "secondary"
                          : "outline"
                    }
                    className="text-xs"
                  >
                    Risk: {user.risk_score}
                  </Badge>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Geographic Distribution */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Globe className="h-4 w-4" />
            Geographic Distribution
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={(analytics.geographic_distribution ?? []).slice(0, 6)}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={({ country, percentage }) =>
                    `${country}: ${percentage}%`
                  }
                  outerRadius={80}
                  fill="#8884d8"
                  dataKey="count"
                >
                  {(analytics.geographic_distribution ?? [])
                    .slice(0, 6)
                    .map((entry, index) => (
                      <Cell
                        key={`cell-${index}`}
                        fill={COLORS[index % COLORS.length]}
                      />
                    ))}
                </Pie>
                <Tooltip />
              </PieChart>
            </ResponsiveContainer>
          </div>
          <div className="mt-4 space-y-2">
            {(analytics.geographic_distribution ?? [])
              .slice(0, 5)
              .map((location, index) => (
                <div
                  key={`${location.country}-${location.region}`}
                  className="flex items-center justify-between text-sm"
                >
                  <div className="flex items-center gap-2">
                    <div
                      className="w-3 h-3 rounded-full"
                      style={{ backgroundColor: COLORS[index % COLORS.length] }}
                    />
                    <span>{location.country}</span>
                    {location.region && (
                      <span className="text-muted-foreground">
                        ({location.region})
                      </span>
                    )}
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="font-medium">{location.count}</span>
                    <Badge variant="outline" className="text-xs">
                      {(location.percentage ?? 0).toFixed(1)}%
                    </Badge>
                  </div>
                </div>
              ))}
          </div>
        </CardContent>
      </Card>

      {/* Device Analytics */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Monitor className="h-4 w-4" />
            Device Types
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={analytics.device_analytics ?? []}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={({ device_type, percentage }) =>
                    `${device_type}: ${percentage}%`
                  }
                  outerRadius={80}
                  fill="#8884d8"
                  dataKey="count"
                >
                  {(analytics.device_analytics ?? []).map((entry, index) => (
                    <Cell
                      key={`cell-${index}`}
                      fill={COLORS[index % COLORS.length]}
                    />
                  ))}
                </Pie>
                <Tooltip />
              </PieChart>
            </ResponsiveContainer>
          </div>
          <div className="mt-4 space-y-2">
            {(analytics.device_analytics ?? []).map((device, index) => (
              <div
                key={device.device_type}
                className="flex items-center justify-between text-sm"
              >
                <div className="flex items-center gap-2">
                  <div
                    className="w-3 h-3 rounded-full"
                    style={{ backgroundColor: COLORS[index % COLORS.length] }}
                  />
                  <span className="capitalize">{device.device_type}</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="font-medium">{device.count}</span>
                  <Badge variant="outline" className="text-xs">
                    {(device.percentage ?? 0).toFixed(1)}%
                  </Badge>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Resource Access Patterns */}
      <Card className="md:col-span-2">
        <CardHeader>
          <CardTitle>Resource Access Patterns</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={(analytics.resource_access ?? []).slice(0, 10)}>
                <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                <XAxis
                  dataKey="resource_type"
                  className="text-xs"
                  tick={{ fontSize: 12 }}
                />
                <YAxis className="text-xs" tick={{ fontSize: 12 }} />
                <Tooltip content={<CustomTooltip />} />
                <Legend />
                <Bar
                  dataKey="access_count"
                  fill="#8884d8"
                  name="Total Access"
                />
                <Bar
                  dataKey="unique_users"
                  fill="#82ca9d"
                  name="Unique Users"
                />
              </BarChart>
            </ResponsiveContainer>
          </div>
          <div className="mt-4 grid grid-cols-2 md:grid-cols-4 gap-4">
            {(analytics.resource_access ?? []).slice(0, 4).map((resource) => (
              <div
                key={resource.resource_type}
                className="text-center p-3 rounded-lg bg-muted/20"
              >
                <div className="font-medium text-sm capitalize">
                  {resource.resource_type}
                </div>
                <div className="text-xs text-muted-foreground mt-1">
                  {resource.access_count} accesses
                </div>
                <div className="text-xs text-muted-foreground">
                  {resource.unique_users} users
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Activity Timeline */}
      <Card className="md:col-span-2">
        <CardHeader>
          <CardTitle>Activity Timeline (24 Hours)</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={stats.activity_by_hour ?? []}>
                <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                <XAxis
                  dataKey="hour"
                  className="text-xs"
                  tick={{ fontSize: 12 }}
                />
                <YAxis className="text-xs" tick={{ fontSize: 12 }} />
                <Tooltip content={<CustomTooltip />} />
                <Legend />
                <Line
                  type="monotone"
                  dataKey="count"
                  stroke="#8884d8"
                  strokeWidth={2}
                  dot={{ r: 4 }}
                  name="Total Events"
                />
                <Line
                  type="monotone"
                  dataKey="failed_count"
                  stroke="#ff7c7c"
                  strokeWidth={2}
                  dot={{ r: 4 }}
                  name="Failed Events"
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
