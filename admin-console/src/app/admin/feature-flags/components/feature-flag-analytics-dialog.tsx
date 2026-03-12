"use client";

import { useState, useEffect } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  BarChart3,
  TrendingUp,
  Users,
  Zap,
  Target,
  Clock,
  Activity,
  Percent,
} from "lucide-react";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
} from "recharts";
import { format } from "date-fns";
import type {
  FeatureFlag,
  FeatureFlagAnalytics,
} from "@/app/_actions/feature-flags";
import { getFeatureFlagAnalytics } from "@/app/_actions/feature-flags";

interface FeatureFlagAnalyticsDialogProps {
  flag: FeatureFlag | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function FeatureFlagAnalyticsDialog({
  flag,
  open,
  onOpenChange,
}: FeatureFlagAnalyticsDialogProps) {
  const [analytics, setAnalytics] = useState<FeatureFlagAnalytics | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    if (flag && open) {
      loadAnalytics();
    }
  }, [flag, open]);

  const loadAnalytics = async () => {
    if (!flag) return;

    setIsLoading(true);
    try {
      const data = await getFeatureFlagAnalytics(flag.key);
      if (data.success && data.data) setAnalytics(data.data);
    } catch (error) {
      console.error("Failed to load analytics:", error);
    } finally {
      setIsLoading(false);
    }
  };

  if (!flag) return null;

  const variationData = analytics
    ? Object.entries(analytics.evaluations.byVariation).map(
        ([variation, count]) => ({
          name: variation.charAt(0).toUpperCase() + variation.slice(1),
          value: count,
          percentage: ((count / analytics.evaluations.total) * 100).toFixed(1),
        }),
      )
    : [];

  const colors = ["#3b82f6", "#10b981", "#f59e0b", "#ef4444", "#8b5cf6"];

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-6xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <BarChart3 className="h-5 w-5" />
            Analytics: {flag.name}
          </DialogTitle>
          <DialogDescription>
            Detailed analytics and performance metrics for the feature flag
          </DialogDescription>
        </DialogHeader>

        {isLoading ? (
          <div className="space-y-6">
            {Array.from({ length: 4 }).map((_, i) => (
              <Card key={i}>
                <CardHeader>
                  <div className="h-6 bg-muted animate-pulse rounded w-48" />
                </CardHeader>
                <CardContent>
                  <div className="h-32 bg-muted animate-pulse rounded" />
                </CardContent>
              </Card>
            ))}
          </div>
        ) : analytics ? (
          <Tabs defaultValue="overview" className="space-y-6">
            <TabsList className="grid w-full grid-cols-4">
              <TabsTrigger value="overview">Overview</TabsTrigger>
              <TabsTrigger value="evaluations">Evaluations</TabsTrigger>
              <TabsTrigger value="targeting">Targeting</TabsTrigger>
              <TabsTrigger value="performance">Performance</TabsTrigger>
            </TabsList>

            {/* Overview Tab */}
            <TabsContent value="overview" className="space-y-6">
              {/* Key Metrics */}
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">
                      Total Evaluations
                    </CardTitle>
                    <Activity className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">
                      {analytics.evaluations.total.toLocaleString()}
                    </div>
                    <p className="text-xs text-muted-foreground">
                      All time evaluations
                    </p>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">
                      Unique Users
                    </CardTitle>
                    <Users className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">
                      {analytics.evaluations.byUser.length.toLocaleString()}
                    </div>
                    <p className="text-xs text-muted-foreground">
                      Users evaluated
                    </p>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">
                      Avg Response Time
                    </CardTitle>
                    <Zap className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold text-green-600">
                      {analytics.performance.avgEvaluationTime}ms
                    </div>
                    <p className="text-xs text-muted-foreground">
                      Average evaluation time
                    </p>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">
                      Cache Hit Rate
                    </CardTitle>
                    <TrendingUp className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold text-blue-600">
                      {(analytics.performance.cacheHitRate * 100).toFixed(1)}%
                    </div>
                    <p className="text-xs text-muted-foreground">
                      Cache efficiency
                    </p>
                  </CardContent>
                </Card>
              </div>

              {/* Variation Distribution */}
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <Card>
                  <CardHeader>
                    <CardTitle>Variation Distribution</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="flex items-center space-x-4">
                      <div className="flex-1">
                        <ResponsiveContainer width="100%" height={200}>
                          <PieChart>
                            <Pie
                              data={variationData}
                              cx="50%"
                              cy="50%"
                              innerRadius={40}
                              outerRadius={80}
                              paddingAngle={2}
                              dataKey="value"
                            >
                              {variationData.map((entry, index) => (
                                <Cell
                                  key={`cell-${index}`}
                                  fill={colors[index % colors.length]}
                                />
                              ))}
                            </Pie>
                            <Tooltip
                              formatter={(value) => [
                                value.toLocaleString(),
                                "Evaluations",
                              ]}
                            />
                          </PieChart>
                        </ResponsiveContainer>
                      </div>
                      <div className="space-y-2">
                        {variationData.map((variation, index) => (
                          <div
                            key={variation.name}
                            className="flex items-center space-x-2"
                          >
                            <div
                              className="w-3 h-3 rounded-full"
                              style={{
                                backgroundColor: colors[index % colors.length],
                              }}
                            />
                            <span className="text-sm">{variation.name}</span>
                            <Badge variant="secondary" className="ml-auto">
                              {variation.percentage}%
                            </Badge>
                          </div>
                        ))}
                      </div>
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader>
                    <CardTitle>Flag Information</CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="grid grid-cols-2 gap-4 text-sm">
                      <div>
                        <span className="font-medium">Flag Key:</span>
                        <p className="text-muted-foreground font-mono">
                          {flag.key}
                        </p>
                      </div>
                      <div>
                        <span className="font-medium">Type:</span>
                        <p className="text-muted-foreground capitalize">
                          {flag.type}
                        </p>
                      </div>
                      <div>
                        <span className="font-medium">Category:</span>
                        <p className="text-muted-foreground capitalize">
                          {flag.category}
                        </p>
                      </div>
                      <div>
                        <span className="font-medium">Environment:</span>
                        <p className="text-muted-foreground capitalize">
                          {flag.environment}
                        </p>
                      </div>
                      <div>
                        <span className="font-medium">Status:</span>
                        <Badge variant={flag.enabled ? "default" : "secondary"}>
                          {flag.enabled ? "Enabled" : "Disabled"}
                        </Badge>
                      </div>
                      <div>
                        <span className="font-medium">Created:</span>
                        <p className="text-muted-foreground">
                          {format(new Date(flag.created_at), "MMM dd, yyyy")}
                        </p>
                      </div>
                    </div>

                    {flag.tags.length > 0 && (
                      <div>
                        <span className="font-medium text-sm">Tags:</span>
                        <div className="flex flex-wrap gap-1 mt-1">
                          {flag.tags.map((tag, index) => (
                            <Badge
                              key={index}
                              variant="outline"
                              className="text-xs"
                            >
                              {tag}
                            </Badge>
                          ))}
                        </div>
                      </div>
                    )}
                  </CardContent>
                </Card>
              </div>
            </TabsContent>

            {/* Evaluations Tab */}
            <TabsContent value="evaluations" className="space-y-6">
              {/* Evaluation Trends */}
              <Card>
                <CardHeader>
                  <CardTitle>Evaluation Trends (7 days)</CardTitle>
                </CardHeader>
                <CardContent>
                  <ResponsiveContainer width="100%" height={300}>
                    <LineChart data={analytics.evaluations.byDay}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis
                        dataKey="date"
                        tickFormatter={(value) =>
                          format(new Date(value), "MMM dd")
                        }
                      />
                      <YAxis />
                      <Tooltip
                        labelFormatter={(value) =>
                          format(new Date(value), "MMM dd, yyyy")
                        }
                        formatter={(value) => [
                          value.toLocaleString(),
                          "Evaluations",
                        ]}
                      />
                      <Line
                        type="monotone"
                        dataKey="count"
                        stroke="#3b82f6"
                        strokeWidth={2}
                        dot={{ fill: "#3b82f6", strokeWidth: 2, r: 4 }}
                      />
                    </LineChart>
                  </ResponsiveContainer>
                </CardContent>
              </Card>

              {/* Top Users */}
              <Card>
                <CardHeader>
                  <CardTitle>Top Users by Evaluations</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    {analytics.evaluations.byUser
                      .slice(0, 10)
                      .map((user, index) => (
                        <div
                          key={user.userId}
                          className="flex items-center justify-between"
                        >
                          <div className="flex items-center space-x-3">
                            <div className="w-6 h-6 rounded-full bg-muted flex items-center justify-center text-xs font-medium">
                              {index + 1}
                            </div>
                            <span className="font-mono text-sm">
                              {user.userId}
                            </span>
                          </div>
                          <Badge variant="secondary">
                            {user.count.toLocaleString()}
                          </Badge>
                        </div>
                      ))}
                  </div>
                </CardContent>
              </Card>
            </TabsContent>

            {/* Targeting Tab */}
            <TabsContent value="targeting" className="space-y-6">
              {flag.targeting.enabled ? (
                <>
                  {/* Rollout Distribution */}
                  <Card>
                    <CardHeader>
                      <CardTitle className="flex items-center gap-2">
                        <Percent className="h-4 w-4" />
                        Rollout Distribution
                      </CardTitle>
                    </CardHeader>
                    <CardContent>
                      <ResponsiveContainer width="100%" height={200}>
                        <BarChart
                          data={Object.entries(
                            analytics.targeting.rolloutDistribution,
                          ).map(([key, value]) => ({
                            name: key.charAt(0).toUpperCase() + key.slice(1),
                            value: value,
                          }))}
                        >
                          <XAxis dataKey="name" />
                          <YAxis />
                          <Tooltip
                            formatter={(value) => [`${value}%`, "Rollout"]}
                          />
                          <Bar
                            dataKey="value"
                            fill="#3b82f6"
                            radius={[4, 4, 0, 0]}
                          />
                        </BarChart>
                      </ResponsiveContainer>
                    </CardContent>
                  </Card>

                  {/* Targeting Rules Performance */}
                  {Object.keys(analytics.targeting.rulesMatched).length > 0 && (
                    <Card>
                      <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                          <Target className="h-4 w-4" />
                          Targeting Rules Performance
                        </CardTitle>
                      </CardHeader>
                      <CardContent>
                        <div className="space-y-3">
                          {Object.entries(analytics.targeting.rulesMatched).map(
                            ([rule, count]) => (
                              <div
                                key={rule}
                                className="flex items-center justify-between"
                              >
                                <span className="font-medium">{rule}</span>
                                <Badge variant="secondary">
                                  {count.toLocaleString()} matches
                                </Badge>
                              </div>
                            ),
                          )}
                        </div>
                      </CardContent>
                    </Card>
                  )}

                  {/* User Segments */}
                  {Object.keys(analytics.targeting.segmentsMatched).length >
                    0 && (
                    <Card>
                      <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                          <Users className="h-4 w-4" />
                          User Segments Performance
                        </CardTitle>
                      </CardHeader>
                      <CardContent>
                        <div className="space-y-3">
                          {Object.entries(
                            analytics.targeting.segmentsMatched,
                          ).map(([segment, count]) => (
                            <div
                              key={segment}
                              className="flex items-center justify-between"
                            >
                              <span className="font-medium">{segment}</span>
                              <Badge variant="secondary">
                                {count.toLocaleString()} matches
                              </Badge>
                            </div>
                          ))}
                        </div>
                      </CardContent>
                    </Card>
                  )}
                </>
              ) : (
                <Card>
                  <CardContent className="text-center py-12">
                    <Target className="mx-auto h-12 w-12 text-muted-foreground" />
                    <h3 className="mt-4 text-lg font-semibold">
                      No Targeting Configured
                    </h3>
                    <p className="text-muted-foreground">
                      This flag does not have targeting rules enabled.
                    </p>
                  </CardContent>
                </Card>
              )}
            </TabsContent>

            {/* Performance Tab */}
            <TabsContent value="performance" className="space-y-6">
              {/* Performance Metrics */}
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">
                      Avg Evaluation Time
                    </CardTitle>
                    <Clock className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold text-green-600">
                      {analytics.performance.avgEvaluationTime}ms
                    </div>
                    <p className="text-xs text-muted-foreground">
                      Average response time
                    </p>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">
                      Error Rate
                    </CardTitle>
                    <Activity className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold text-red-600">
                      {(analytics.performance.errorRate * 100).toFixed(2)}%
                    </div>
                    <p className="text-xs text-muted-foreground">
                      Evaluation errors
                    </p>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">
                      Cache Hit Rate
                    </CardTitle>
                    <TrendingUp className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold text-blue-600">
                      {(analytics.performance.cacheHitRate * 100).toFixed(1)}%
                    </div>
                    <p className="text-xs text-muted-foreground">
                      Cache efficiency
                    </p>
                  </CardContent>
                </Card>
              </div>

              {/* Performance Recommendations */}
              <Card>
                <CardHeader>
                  <CardTitle>Performance Recommendations</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    {analytics.performance.avgEvaluationTime > 5 && (
                      <div className="flex items-start space-x-3 p-3 bg-yellow-50 rounded-lg">
                        <Clock className="h-5 w-5 text-yellow-600 mt-0.5" />
                        <div>
                          <h4 className="font-medium text-yellow-800">
                            Slow Evaluation Time
                          </h4>
                          <p className="text-sm text-yellow-700">
                            Consider optimizing targeting rules or caching
                            strategies to improve response time.
                          </p>
                        </div>
                      </div>
                    )}

                    {analytics.performance.errorRate > 0.01 && (
                      <div className="flex items-start space-x-3 p-3 bg-red-50 rounded-lg">
                        <Activity className="h-5 w-5 text-red-600 mt-0.5" />
                        <div>
                          <h4 className="font-medium text-red-800">
                            High Error Rate
                          </h4>
                          <p className="text-sm text-red-700">
                            Review flag configuration and targeting rules to
                            reduce evaluation errors.
                          </p>
                        </div>
                      </div>
                    )}

                    {analytics.performance.cacheHitRate < 0.8 && (
                      <div className="flex items-start space-x-3 p-3 bg-blue-50 rounded-lg">
                        <TrendingUp className="h-5 w-5 text-blue-600 mt-0.5" />
                        <div>
                          <h4 className="font-medium text-blue-800">
                            Low Cache Hit Rate
                          </h4>
                          <p className="text-sm text-blue-700">
                            Consider adjusting cache TTL or evaluation patterns
                            to improve cache efficiency.
                          </p>
                        </div>
                      </div>
                    )}

                    {analytics.performance.avgEvaluationTime <= 5 &&
                      analytics.performance.errorRate <= 0.01 &&
                      analytics.performance.cacheHitRate >= 0.8 && (
                        <div className="flex items-start space-x-3 p-3 bg-green-50 rounded-lg">
                          <TrendingUp className="h-5 w-5 text-green-600 mt-0.5" />
                          <div>
                            <h4 className="font-medium text-green-800">
                              Excellent Performance
                            </h4>
                            <p className="text-sm text-green-700">
                              This flag is performing well with fast evaluation
                              times and low error rates.
                            </p>
                          </div>
                        </div>
                      )}
                  </div>
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        ) : (
          <div className="text-center py-12">
            <BarChart3 className="mx-auto h-12 w-12 text-muted-foreground" />
            <h3 className="mt-4 text-lg font-semibold">No Analytics Data</h3>
            <p className="text-muted-foreground">
              Analytics data is not available for this flag.
            </p>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
