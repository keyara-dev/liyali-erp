"use client";

import { useState, useEffect } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  TrendingUp,
  TrendingDown,
  DollarSign,
  Users,
  Calendar,
  BarChart3,
  PieChart,
  Activity,
} from "lucide-react";
import { toast } from "sonner";
import { getSubscriptionAnalytics } from "@/app/_actions/subscriptions";

export function SubscriptionAnalyticsTab() {
  const [analytics, setAnalytics] = useState<any>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [timeRange, setTimeRange] = useState<"7d" | "30d" | "90d" | "1y">(
    "30d",
  );

  const loadAnalytics = async () => {
    setIsLoading(true);
    try {
      const result = await getSubscriptionAnalytics();
      if (result.success) {
        setAnalytics(result.data);
      } else {
        toast.error(result.message || "Failed to load analytics");
      }
    } catch (error) {
      toast.error("Failed to load analytics");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadAnalytics();
  }, [timeRange]);

  if (isLoading) {
    return <div>Loading analytics...</div>;
  }

  if (!analytics) {
    return <div>No analytics data available</div>;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-medium">Subscription Analytics</h3>
          <p className="text-sm text-muted-foreground">
            Revenue, conversion, and subscription metrics
          </p>
        </div>
        <div className="flex gap-2">
          {["7d", "30d", "90d", "1y"].map((range) => (
            <button
              key={range}
              onClick={() => setTimeRange(range as any)}
              className={`px-3 py-1 text-sm rounded-md ${
                timeRange === range
                  ? "bg-primary text-primary-foreground"
                  : "bg-muted text-muted-foreground hover:bg-muted/80"
              }`}
            >
              {range === "7d"
                ? "7 Days"
                : range === "30d"
                  ? "30 Days"
                  : range === "90d"
                    ? "90 Days"
                    : "1 Year"}
            </button>
          ))}
        </div>
      </div>

      {/* Key Metrics */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Monthly Revenue
            </CardTitle>
            <DollarSign className="h-4 w-4 text-green-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              ${analytics.revenue.monthly.toLocaleString()}
            </div>
            <div className="flex items-center text-xs">
              {analytics.revenue.trend === "up" ? (
                <TrendingUp className="h-3 w-3 text-green-600 mr-1" />
              ) : (
                <TrendingDown className="h-3 w-3 text-red-600 mr-1" />
              )}
              <span
                className={
                  analytics.revenue.trend === "up"
                    ? "text-green-600"
                    : "text-red-600"
                }
              >
                {analytics.revenue.growth}% from last month
              </span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Active Subscriptions
            </CardTitle>
            <Users className="h-4 w-4 text-blue-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {analytics.subscriptions.total}
            </div>
            <p className="text-xs text-muted-foreground">
              +{analytics.subscriptions.new_this_month} new this month
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Trial Conversion
            </CardTitle>
            <Activity className="h-4 w-4 text-purple-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {analytics.trials.conversion_rate}%
            </div>
            <p className="text-xs text-muted-foreground">
              {analytics.trials.converted_this_month} converted this month
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Churn Rate</CardTitle>
            <TrendingDown className="h-4 w-4 text-orange-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {analytics.metrics.churn_rate}%
            </div>
            <p className="text-xs text-muted-foreground">
              {analytics.subscriptions.churned_this_month} churned this month
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Revenue Metrics */}
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Revenue Metrics</CardTitle>
            <CardDescription>Key financial indicators</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium">
                Monthly Recurring Revenue (MRR)
              </span>
              <span className="text-lg font-bold">
                ${analytics.metrics.mrr.toLocaleString()}
              </span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium">
                Annual Recurring Revenue (ARR)
              </span>
              <span className="text-lg font-bold">
                ${analytics.metrics.arr.toLocaleString()}
              </span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium">
                Average Revenue Per User (ARPU)
              </span>
              <span className="text-lg font-bold">
                ${analytics.metrics.arpu.toFixed(2)}
              </span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium">
                Customer Lifetime Value (LTV)
              </span>
              <span className="text-lg font-bold">
                ${analytics.metrics.ltv.toFixed(2)}
              </span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Subscription Tiers Distribution</CardTitle>
            <CardDescription>
              Revenue and subscriber count by tier
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {analytics.tiers.map((tier, index) => (
                <div key={tier.name} className="space-y-2">
                  <div className="flex justify-between items-center">
                    <div className="flex items-center gap-2">
                      <div
                        className={`w-3 h-3 rounded-full ${
                          index === 0
                            ? "bg-blue-500"
                            : index === 1
                              ? "bg-green-500"
                              : "bg-purple-500"
                        }`}
                      />
                      <span className="font-medium">{tier.name}</span>
                      <Badge variant="secondary">{tier.percentage}%</Badge>
                    </div>
                    <div className="text-right">
                      <div className="font-semibold">
                        ${tier.revenue.toLocaleString()}
                      </div>
                      <div className="text-sm text-muted-foreground">
                        {tier.count} subscribers
                      </div>
                    </div>
                  </div>
                  <div className="w-full bg-muted rounded-full h-2">
                    <div
                      className={`h-2 rounded-full ${
                        index === 0
                          ? "bg-blue-500"
                          : index === 1
                            ? "bg-green-500"
                            : "bg-purple-500"
                      }`}
                      style={{ width: `${tier.percentage}%` }}
                    />
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Trial Analytics */}
      <Card>
        <CardHeader>
          <CardTitle>Trial Analytics</CardTitle>
          <CardDescription>
            Trial conversion and performance metrics
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-4">
            <div className="text-center">
              <div className="text-2xl font-bold text-blue-600">
                {analytics.trials.active}
              </div>
              <p className="text-sm text-muted-foreground">Active Trials</p>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-green-600">
                {analytics.trials.converted_this_month}
              </div>
              <p className="text-sm text-muted-foreground">
                Converted This Month
              </p>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-red-600">
                {analytics.trials.expired_this_month}
              </div>
              <p className="text-sm text-muted-foreground">
                Expired This Month
              </p>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-purple-600">
                {analytics.trials.conversion_rate}%
              </div>
              <p className="text-sm text-muted-foreground">Conversion Rate</p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Growth Trends */}
      <Card>
        <CardHeader>
          <CardTitle>Growth Trends</CardTitle>
          <CardDescription>Subscription growth over time</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="h-64 flex items-center justify-center text-muted-foreground">
            <div className="text-center">
              <BarChart3 className="h-12 w-12 mx-auto mb-2" />
              <p>Chart visualization would go here</p>
              <p className="text-sm">
                Integration with charting library needed
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
