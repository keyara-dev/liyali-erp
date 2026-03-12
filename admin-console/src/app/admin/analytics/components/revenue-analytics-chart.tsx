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
import { DollarSign, TrendingUp, Target, CreditCard } from "lucide-react";
import { type RevenueAnalytics } from "@/app/_actions/analytics";

interface RevenueAnalyticsChartProps {
  analytics: RevenueAnalytics | null;
  isLoading?: boolean;
}

const COLORS = ["#0088FE", "#00C49F", "#FFBB28", "#FF8042", "#8884D8"];

export function RevenueAnalyticsChart({
  analytics,
  isLoading,
}: RevenueAnalyticsChartProps) {
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
            No revenue analytics data available
          </p>
        </CardContent>
      </Card>
    );
  }

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: "USD",
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(amount);
  };

  const formatCurrencyDetailed = (amount: number) => {
    return new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: "USD",
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(amount);
  };

  const revenueChartData = (analytics.revenue_trend ?? []).map((point) => ({
    date: new Date(point.date).toLocaleDateString([], {
      month: "short",
      day: "numeric",
    }),
    revenue: point.revenue ?? 0,
    mrr: point.mrr ?? 0,
    newRevenue: point.new_revenue ?? 0,
    churnRevenue: point.churn_revenue ?? 0,
    netRevenue: (point.new_revenue ?? 0) - (point.churn_revenue ?? 0),
  }));

  const CustomTooltip = ({ active, payload, label }: any) => {
    if (active && payload && payload.length) {
      return (
        <div className="bg-background border rounded-lg p-3 shadow-lg">
          <p className="font-medium">{`Date: ${label}`}</p>
          {payload.map((entry: any, index: number) => (
            <p key={index} style={{ color: entry.color }}>
              {`${entry.name}: ${formatCurrency(entry.value)}`}
            </p>
          ))}
        </div>
      );
    }
    return null;
  };

  return (
    <div className="grid gap-4 md:grid-cols-2">
      {/* Revenue Trend */}
      <Card className="md:col-span-2">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <TrendingUp className="h-4 w-4" />
            Revenue Trend
          </CardTitle>
          <div className="flex items-center gap-2">
            <Badge variant="outline" className="text-xs">
              Total: {formatCurrency(analytics.total_revenue ?? 0)}
            </Badge>
            <Badge variant="outline" className="text-xs">
              MRR: {formatCurrency(analytics.monthly_recurring_revenue ?? 0)}
            </Badge>
            <Badge variant="outline" className="text-xs">
              ARR: {formatCurrency(analytics.annual_recurring_revenue ?? 0)}
            </Badge>
          </div>
        </CardHeader>
        <CardContent>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={revenueChartData}>
                <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                <XAxis
                  dataKey="date"
                  className="text-xs"
                  tick={{ fontSize: 12 }}
                />
                <YAxis
                  className="text-xs"
                  tick={{ fontSize: 12 }}
                  tickFormatter={(value) => formatCurrency(value)}
                />
                <Tooltip content={<CustomTooltip />} />
                <Legend />
                <Area
                  type="monotone"
                  dataKey="revenue"
                  stackId="1"
                  stroke="#8884d8"
                  fill="#8884d8"
                  fillOpacity={0.6}
                  name="Total Revenue"
                />
                <Area
                  type="monotone"
                  dataKey="mrr"
                  stackId="2"
                  stroke="#82ca9d"
                  fill="#82ca9d"
                  fillOpacity={0.6}
                  name="Monthly Recurring Revenue"
                />
                <Line
                  type="monotone"
                  dataKey="netRevenue"
                  stroke="#ffc658"
                  strokeWidth={2}
                  dot={{ r: 4 }}
                  name="Net New Revenue"
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </CardContent>
      </Card>

      {/* Revenue by Tier */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <DollarSign className="h-4 w-4" />
            Revenue by Tier
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={analytics.revenue_by_tier ?? []}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={({ tier, percentage }) => `${tier}: ${percentage ?? 0}%`}
                  outerRadius={80}
                  fill="#8884d8"
                  dataKey="revenue"
                >
                  {(analytics.revenue_by_tier ?? []).map((entry, index) => (
                    <Cell
                      key={`cell-${index}`}
                      fill={COLORS[index % COLORS.length]}
                    />
                  ))}
                </Pie>
                <Tooltip formatter={(value: any) => formatCurrency(value)} />
              </PieChart>
            </ResponsiveContainer>
          </div>
          <div className="space-y-2 mt-4">
            {(analytics.revenue_by_tier ?? []).map((tier, index) => (
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
                <div className="text-right">
                  <div className="font-medium">
                    {formatCurrency(tier.revenue ?? 0)}
                  </div>
                  <div className="text-xs text-muted-foreground">
                    {tier.subscriber_count ?? 0} subscribers
                  </div>
                </div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      {/* Financial Metrics */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Target className="h-4 w-4" />
            Financial Metrics
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">ARPU</span>
              <span className="font-medium">
                {formatCurrencyDetailed(
                  analytics.financial_metrics?.average_revenue_per_user ?? 0,
                )}
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">
                Customer LTV
              </span>
              <span className="font-medium">
                {formatCurrency(
                  analytics.financial_metrics?.customer_lifetime_value ?? 0,
                )}
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Churn Rate</span>
              <div className="text-right">
                <span
                  className={`font-medium ${(analytics.financial_metrics?.churn_rate ?? 0) <= 5 ? "text-green-600" : "text-red-600"}`}
                >
                  {(analytics.financial_metrics?.churn_rate ?? 0).toFixed(1)}%
                </span>
                <div className="text-xs text-muted-foreground">
                  <Badge
                    variant={
                      (analytics.financial_metrics?.churn_rate ?? 0) <= 5
                        ? "default"
                        : "destructive"
                    }
                    className="text-xs"
                  >
                    {(analytics.financial_metrics?.churn_rate ?? 0) <= 5
                      ? "Healthy"
                      : "High"}
                  </Badge>
                </div>
              </div>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">
                Net Revenue Retention
              </span>
              <div className="text-right">
                <span
                  className={`font-medium ${(analytics.financial_metrics?.net_revenue_retention ?? 0) >= 100 ? "text-green-600" : "text-red-600"}`}
                >
                  {(analytics.financial_metrics?.net_revenue_retention ?? 0).toFixed(1)}
                  %
                </span>
                <div className="text-xs text-muted-foreground">
                  <Badge
                    variant={
                      (analytics.financial_metrics?.net_revenue_retention ?? 0) >= 100
                        ? "default"
                        : "destructive"
                    }
                    className="text-xs"
                  >
                    {(analytics.financial_metrics?.net_revenue_retention ?? 0) >= 100
                      ? "Growing"
                      : "Declining"}
                  </Badge>
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Revenue Growth Analysis */}
      <Card className="md:col-span-2">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <CreditCard className="h-4 w-4" />
            Revenue Growth Analysis
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
            <div className="text-center">
              <div className="text-2xl font-bold text-blue-600">
                {formatCurrency(analytics.total_revenue ?? 0)}
              </div>
              <div className="text-sm text-muted-foreground">Total Revenue</div>
            </div>

            <div className="text-center">
              <div className="text-2xl font-bold text-green-600">
                {formatCurrency(analytics.monthly_recurring_revenue ?? 0)}
              </div>
              <div className="text-sm text-muted-foreground">
                Monthly Recurring Revenue
              </div>
            </div>

            <div className="text-center">
              <div className="text-2xl font-bold text-purple-600">
                {formatCurrency(analytics.annual_recurring_revenue ?? 0)}
              </div>
              <div className="text-sm text-muted-foreground">
                Annual Recurring Revenue
              </div>
            </div>

            <div className="text-center">
              <div
                className={`text-2xl font-bold ${(analytics.revenue_growth_rate ?? 0) >= 0 ? "text-green-600" : "text-red-600"}`}
              >
                {(analytics.revenue_growth_rate ?? 0) >= 0 ? "+" : ""}
                {(analytics.revenue_growth_rate ?? 0).toFixed(1)}%
              </div>
              <div className="text-sm text-muted-foreground">Growth Rate</div>
            </div>
          </div>

          {/* Revenue Breakdown Chart */}
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={revenueChartData.slice(-12)}>
                <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                <XAxis
                  dataKey="date"
                  className="text-xs"
                  tick={{ fontSize: 12 }}
                />
                <YAxis
                  className="text-xs"
                  tick={{ fontSize: 12 }}
                  tickFormatter={(value) => formatCurrency(value)}
                />
                <Tooltip content={<CustomTooltip />} />
                <Legend />
                <Bar dataKey="newRevenue" fill="#82ca9d" name="New Revenue" />
                <Bar
                  dataKey="churnRevenue"
                  fill="#ff7c7c"
                  name="Churned Revenue"
                />
              </BarChart>
            </ResponsiveContainer>
          </div>

          {/* Revenue Health Indicators */}
          <div className="grid grid-cols-3 gap-4 mt-6 pt-4 border-t">
            <div className="text-center">
              <div className="text-sm text-muted-foreground mb-2">
                Revenue Health
              </div>
              <Progress
                value={Math.min((analytics.revenue_growth_rate ?? 0) + 50, 100)}
                className="w-full h-2 mb-2"
              />
              <Badge
                variant={
                  (analytics.revenue_growth_rate ?? 0) > 10
                    ? "default"
                    : (analytics.revenue_growth_rate ?? 0) > 0
                      ? "secondary"
                      : "destructive"
                }
              >
                {(analytics.revenue_growth_rate ?? 0) > 10
                  ? "Excellent"
                  : (analytics.revenue_growth_rate ?? 0) > 0
                    ? "Good"
                    : "Needs Attention"}
              </Badge>
            </div>

            <div className="text-center">
              <div className="text-sm text-muted-foreground mb-2">
                Retention Health
              </div>
              <Progress
                value={analytics.financial_metrics?.net_revenue_retention ?? 0}
                className="w-full h-2 mb-2"
              />
              <Badge
                variant={
                  (analytics.financial_metrics?.net_revenue_retention ?? 0) >= 110
                    ? "default"
                    : (analytics.financial_metrics?.net_revenue_retention ?? 0) >= 100
                      ? "secondary"
                      : "destructive"
                }
              >
                {(analytics.financial_metrics?.net_revenue_retention ?? 0) >= 110
                  ? "Excellent"
                  : (analytics.financial_metrics?.net_revenue_retention ?? 0) >= 100
                    ? "Good"
                    : "At Risk"}
              </Badge>
            </div>

            <div className="text-center">
              <div className="text-sm text-muted-foreground mb-2">
                Churn Health
              </div>
              <Progress
                value={Math.max(
                  100 - (analytics.financial_metrics?.churn_rate ?? 0) * 10,
                  0,
                )}
                className="w-full h-2 mb-2"
              />
              <Badge
                variant={
                  (analytics.financial_metrics?.churn_rate ?? 0) <= 3
                    ? "default"
                    : (analytics.financial_metrics?.churn_rate ?? 0) <= 7
                      ? "secondary"
                      : "destructive"
                }
              >
                {(analytics.financial_metrics?.churn_rate ?? 0) <= 3
                  ? "Excellent"
                  : (analytics.financial_metrics?.churn_rate ?? 0) <= 7
                    ? "Good"
                    : "High"}
              </Badge>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
