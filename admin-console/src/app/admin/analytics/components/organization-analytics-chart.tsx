"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import {
  LineChart,
  Line,
  AreaChart,
  Area,
  BarChart,
  Bar,
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
import { Building2, TrendingUp, Target, Clock } from "lucide-react";
import { type OrganizationAnalytics } from "@/app/_actions/analytics";

interface OrganizationAnalyticsChartProps {
  analytics: OrganizationAnalytics | null;
  isLoading?: boolean;
}

const COLORS = ["#0088FE", "#00C49F", "#FFBB28", "#FF8042", "#8884D8"];

export function OrganizationAnalyticsChart({
  analytics,
  isLoading,
}: OrganizationAnalyticsChartProps) {
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

  if (!analytics) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center h-64">
          <p className="text-muted-foreground">
            No organization analytics data available
          </p>
        </CardContent>
      </Card>
    );
  }

  const growthChartData = analytics.organization_growth_trend.map((point) => ({
    date: new Date(point.date).toLocaleDateString([], {
      month: "short",
      day: "numeric",
    }),
    totalOrganizations: point.total_organizations,
    newOrganizations: point.new_organizations,
    activeOrganizations: point.active_organizations,
  }));

  const CustomTooltip = ({ active, payload, label }: any) => {
    if (active && payload && payload.length) {
      return (
        <div className="bg-background border rounded-lg p-3 shadow-lg">
          <p className="font-medium">{`Date: ${label}`}</p>
          {payload.map((entry: any, index: number) => (
            <p key={index} style={{ color: entry.color }}>
              {`${entry.name}: ${entry.value.toLocaleString()}`}
            </p>
          ))}
        </div>
      );
    }
    return null;
  };

  return (
    <div className="grid gap-4 md:grid-cols-2">
      {/* Organization Growth Trend */}
      <Card className="md:col-span-2">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <TrendingUp className="h-4 w-4" />
            Organization Growth Trend
          </CardTitle>
          <div className="flex items-center gap-2">
            <Badge variant="outline" className="text-xs">
              Total: {analytics.total_organizations.toLocaleString()}
            </Badge>
            <Badge variant="outline" className="text-xs">
              New this period:{" "}
              {analytics.new_organizations_this_period.toLocaleString()}
            </Badge>
          </div>
        </CardHeader>
        <CardContent>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={growthChartData}>
                <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                <XAxis
                  dataKey="date"
                  className="text-xs"
                  tick={{ fontSize: 12 }}
                />
                <YAxis className="text-xs" tick={{ fontSize: 12 }} />
                <Tooltip content={<CustomTooltip />} />
                <Legend />
                <Area
                  type="monotone"
                  dataKey="totalOrganizations"
                  stackId="1"
                  stroke="#8884d8"
                  fill="#8884d8"
                  fillOpacity={0.6}
                  name="Total Organizations"
                />
                <Area
                  type="monotone"
                  dataKey="activeOrganizations"
                  stackId="2"
                  stroke="#82ca9d"
                  fill="#82ca9d"
                  fillOpacity={0.6}
                  name="Active Organizations"
                />
                <Line
                  type="monotone"
                  dataKey="newOrganizations"
                  stroke="#ffc658"
                  strokeWidth={2}
                  dot={{ r: 4 }}
                  name="New Organizations"
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </CardContent>
      </Card>

      {/* Subscription Tier Distribution */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Building2 className="h-4 w-4" />
            Subscription Tiers
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={
                    analytics.organization_distribution.by_subscription_tier
                  }
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={({ tier, percentage }) => `${tier}: ${percentage}%`}
                  outerRadius={80}
                  fill="#8884d8"
                  dataKey="count"
                >
                  {analytics.organization_distribution.by_subscription_tier.map(
                    (entry, index) => (
                      <Cell
                        key={`cell-${index}`}
                        fill={COLORS[index % COLORS.length]}
                      />
                    ),
                  )}
                </Pie>
                <Tooltip />
              </PieChart>
            </ResponsiveContainer>
          </div>
          <div className="space-y-2 mt-4">
            {analytics.organization_distribution.by_subscription_tier.map(
              (tier, index) => (
                <div
                  key={tier.tier}
                  className="flex items-center justify-between text-sm"
                >
                  <div className="flex items-center gap-2">
                    <div
                      className="w-3 h-3 rounded-full"
                      style={{ backgroundColor: COLORS[index % COLORS.length] }}
                    />
                    <span className="capitalize">{tier.tier}</span>
                  </div>
                  <span className="font-medium">
                    {tier.count} ({tier.percentage}%)
                  </span>
                </div>
              ),
            )}
          </div>
        </CardContent>
      </Card>

      {/* Organization Size Distribution */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Target className="h-4 w-4" />
            Organization Sizes
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart
                data={analytics.organization_distribution.by_user_count}
              >
                <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                <XAxis
                  dataKey="range"
                  className="text-xs"
                  tick={{ fontSize: 12 }}
                />
                <YAxis className="text-xs" tick={{ fontSize: 12 }} />
                <Tooltip />
                <Bar dataKey="count" fill="#8884d8" />
              </BarChart>
            </ResponsiveContainer>
          </div>
          <div className="grid grid-cols-1 gap-2 mt-4 text-sm">
            {analytics.organization_distribution.by_user_count.map((size) => (
              <div
                key={size.range}
                className="flex items-center justify-between"
              >
                <span>{size.range} users</span>
                <div className="flex items-center gap-2">
                  <Progress value={size.percentage} className="w-16 h-2" />
                  <span className="font-medium w-12 text-right">
                    {size.count}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Trial Metrics */}
      <Card className="md:col-span-2">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Clock className="h-4 w-4" />
            Trial Metrics
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div className="text-center">
              <div className="text-2xl font-bold text-blue-600">
                {analytics.trial_metrics.trial_organizations.toLocaleString()}
              </div>
              <div className="text-sm text-muted-foreground">
                Organizations on Trial
              </div>
              <div className="text-xs text-muted-foreground mt-1">
                {(
                  (analytics.trial_metrics.trial_organizations /
                    analytics.total_organizations) *
                  100
                ).toFixed(1)}
                % of total
              </div>
            </div>

            <div className="text-center">
              <div className="text-2xl font-bold text-green-600">
                {analytics.trial_metrics.trial_conversion_rate.toFixed(1)}%
              </div>
              <div className="text-sm text-muted-foreground">
                Trial Conversion Rate
              </div>
              <div className="text-xs text-muted-foreground mt-1">
                <Badge
                  variant={
                    analytics.trial_metrics.trial_conversion_rate > 20
                      ? "default"
                      : "secondary"
                  }
                >
                  {analytics.trial_metrics.trial_conversion_rate > 20
                    ? "Good"
                    : "Needs Improvement"}
                </Badge>
              </div>
            </div>

            <div className="text-center">
              <div className="text-2xl font-bold text-purple-600">
                {analytics.trial_metrics.average_trial_duration.toFixed(0)} days
              </div>
              <div className="text-sm text-muted-foreground">
                Avg Trial Duration
              </div>
              <div className="text-xs text-muted-foreground mt-1">
                Before conversion
              </div>
            </div>

            <div className="text-center">
              <div className="text-2xl font-bold text-orange-600">
                {analytics.trial_metrics.trials_expiring_soon.toLocaleString()}
              </div>
              <div className="text-sm text-muted-foreground">Expiring Soon</div>
              <div className="text-xs text-muted-foreground mt-1">
                <Badge
                  variant={
                    analytics.trial_metrics.trials_expiring_soon > 0
                      ? "destructive"
                      : "default"
                  }
                >
                  {analytics.trial_metrics.trials_expiring_soon > 0
                    ? "Action Needed"
                    : "All Good"}
                </Badge>
              </div>
            </div>
          </div>

          {/* Trial Conversion Funnel */}
          <div className="mt-6 pt-4 border-t">
            <h4 className="text-sm font-medium mb-4">
              Trial Conversion Funnel
            </h4>
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-sm">Trial Started</span>
                <div className="flex items-center gap-2">
                  <Progress value={100} className="w-32 h-2" />
                  <span className="text-sm font-medium w-16 text-right">
                    {analytics.trial_metrics.trial_organizations}
                  </span>
                </div>
              </div>

              <div className="flex items-center justify-between">
                <span className="text-sm">Active in Trial</span>
                <div className="flex items-center gap-2">
                  <Progress value={85} className="w-32 h-2" />
                  <span className="text-sm font-medium w-16 text-right">
                    {Math.round(
                      analytics.trial_metrics.trial_organizations * 0.85,
                    )}
                  </span>
                </div>
              </div>

              <div className="flex items-center justify-between">
                <span className="text-sm">Converted to Paid</span>
                <div className="flex items-center gap-2">
                  <Progress
                    value={analytics.trial_metrics.trial_conversion_rate}
                    className="w-32 h-2"
                  />
                  <span className="text-sm font-medium w-16 text-right">
                    {Math.round(
                      analytics.trial_metrics.trial_organizations *
                        (analytics.trial_metrics.trial_conversion_rate / 100),
                    )}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
