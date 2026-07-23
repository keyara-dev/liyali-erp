"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import {
  Flag,
  ToggleLeft,
  ToggleRight,
  Archive,
  Beaker,
  Shield,
  Zap,
  AlertTriangle,
  Users,
  TrendingUp,
  Clock,
  Activity,
} from "lucide-react";
import {
  PieChart,
  Pie,
  Cell,
  ResponsiveContainer,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  LineChart,
  Line,
} from "recharts";
import type { FeatureFlagStats } from "@/app/_actions/feature-flags";

interface FeatureFlagsStatsGridProps {
  stats: FeatureFlagStats;
  isLoading?: boolean;
}

export function FeatureFlagsStatsGrid({
  stats,
  isLoading = false,
}: FeatureFlagsStatsGridProps) {
  const categoryIcons = {
    feature: Flag,
    experiment: Beaker,
    operational: Shield,
    killswitch: AlertTriangle,
    permission: Users,
  };

  const categoryColors = {
    feature: "#3b82f6",
    experiment: "#10b981",
    operational: "#f59e0b",
    killswitch: "#ef4444",
    permission: "#8b5cf6",
  };

  const categoryData = Object.entries(stats.byCategory).map(
    ([category, count]) => ({
      name: category.charAt(0).toUpperCase() + category.slice(1),
      value: count,
      color:
        categoryColors[category as keyof typeof categoryColors] || "#6b7280",
    }),
  );

  const environmentData = Object.entries(stats.byEnvironment).map(
    ([env, count]) => ({
      name:
        env === "all" ? "All Envs" : env.charAt(0).toUpperCase() + env.slice(1),
      value: count,
    }),
  );

  const typeData = Object.entries(stats.byType).map(([type, count]) => ({
    name: type.charAt(0).toUpperCase() + type.slice(1),
    value: count,
  }));

  // Mock trend data for the last 7 days
  const trendData = [
    { date: "Mon", evaluations: 12000 },
    { date: "Tue", evaluations: 13500 },
    { date: "Wed", evaluations: 11800 },
    { date: "Thu", evaluations: 15200 },
    { date: "Fri", evaluations: 16800 },
    { date: "Sat", evaluations: 14200 },
    { date: "Sun", evaluations: 15420 },
  ];

  const enabledPercentage =
    stats.total > 0 ? (stats.enabled / stats.total) * 100 : 0;
  const archivedPercentage =
    stats.total > 0 ? (stats.archived / stats.total) * 100 : 0;

  if (isLoading) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {Array.from({ length: 8 }).map((_, i) => (
          <Card key={i}>
            <CardHeader className="pb-2">
              <div className="h-4 bg-muted animate-pulse rounded" />
            </CardHeader>
            <CardContent>
              <div className="h-8 bg-muted animate-pulse rounded mb-2" />
              <div className="h-3 bg-muted animate-pulse rounded w-2/3" />
            </CardContent>
          </Card>
        ))}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Overview Stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {/* Total Flags */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Flags</CardTitle>
            <Flag className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.total}</div>
            <p className="text-xs text-muted-foreground">
              Across all environments
            </p>
          </CardContent>
        </Card>

        {/* Enabled Flags */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Enabled Flags</CardTitle>
            <ToggleRight className="h-4 w-4 text-green-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">
              {stats.enabled}
            </div>
            <p className="text-xs text-muted-foreground">
              {enabledPercentage.toFixed(1)}% of total
            </p>
            <Progress value={enabledPercentage} className="mt-2 h-2" />
          </CardContent>
        </Card>

        {/* Disabled Flags */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Disabled Flags
            </CardTitle>
            <ToggleLeft className="h-4 w-4 text-red-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-600">
              {stats.disabled}
            </div>
            <p className="text-xs text-muted-foreground">
              {((stats.disabled / stats.total) * 100).toFixed(1)}% of total
            </p>
          </CardContent>
        </Card>

        {/* Evaluations Today */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Evaluations Today
            </CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-blue-600">
              {stats.evaluationsToday.toLocaleString()}
            </div>
            <p className="text-xs text-muted-foreground">Flag evaluations</p>
          </CardContent>
        </Card>
      </div>

      {/* Secondary Stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {/* Archived Flags */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Archived Flags
            </CardTitle>
            <Archive className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-amber-600">
              {stats.archived}
            </div>
            <p className="text-xs text-muted-foreground">
              {archivedPercentage.toFixed(1)}% of total
            </p>
          </CardContent>
        </Card>

        {/* Recently Created */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Recently Created
            </CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">
              {stats.recentlyCreated}
            </div>
            <p className="text-xs text-muted-foreground">Last 7 days</p>
          </CardContent>
        </Card>

        {/* Recently Updated */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Recently Updated
            </CardTitle>
            <Clock className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-blue-600">
              {stats.recentlyUpdated}
            </div>
            <p className="text-xs text-muted-foreground">Last 7 days</p>
          </CardContent>
        </Card>

        {/* Expiring Soon */}
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Expiring Soon</CardTitle>
            <AlertTriangle className="h-4 w-4 text-amber-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-amber-600">
              {stats.expiringSoon}
            </div>
            <p className="text-xs text-muted-foreground">Next 30 days</p>
          </CardContent>
        </Card>
      </div>

      {/* Distribution Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Category Distribution */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Flags by Category</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-center space-x-4">
              <div className="flex-1">
                <ResponsiveContainer width="100%" height={200}>
                  <PieChart>
                    <Pie
                      data={categoryData}
                      cx="50%"
                      cy="50%"
                      innerRadius={40}
                      outerRadius={80}
                      paddingAngle={2}
                      dataKey="value"
                    >
                      {categoryData.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={entry.color} />
                      ))}
                    </Pie>
                    <Tooltip />
                  </PieChart>
                </ResponsiveContainer>
              </div>
              <div className="space-y-2">
                {categoryData.map((category) => {
                  const Icon =
                    categoryIcons[
                      category.name.toLowerCase() as keyof typeof categoryIcons
                    ] || Flag;
                  return (
                    <div
                      key={category.name}
                      className="flex items-center space-x-2"
                    >
                      <div
                        className="w-3 h-3 rounded-full"
                        style={{ backgroundColor: category.color }}
                      />
                      <Icon className="h-4 w-4 text-muted-foreground" />
                      <span className="text-sm">{category.name}</span>
                      <Badge variant="secondary" className="ml-auto">
                        {category.value}
                      </Badge>
                    </div>
                  );
                })}
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Environment Distribution */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Flags by Environment</CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={200}>
              <BarChart data={environmentData}>
                <XAxis dataKey="name" />
                <YAxis />
                <Tooltip />
                <Bar dataKey="value" fill="#3b82f6" radius={[4, 4, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      </div>

      {/* Type Distribution and Evaluation Trends */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Type Distribution */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Flags by Type</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {typeData.map((type) => {
                const percentage = (type.value / stats.total) * 100;
                return (
                  <div key={type.name} className="space-y-1">
                    <div className="flex justify-between text-sm">
                      <span>{type.name}</span>
                      <span className="text-muted-foreground">
                        {type.value} ({percentage.toFixed(1)}%)
                      </span>
                    </div>
                    <Progress value={percentage} className="h-2" />
                  </div>
                );
              })}
            </div>
          </CardContent>
        </Card>

        {/* Evaluation Trends */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">
              Evaluation Trends (7 days)
            </CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={200}>
              <LineChart data={trendData}>
                <XAxis dataKey="date" />
                <YAxis />
                <Tooltip
                  formatter={(value) => [value.toLocaleString(), "Evaluations"]}
                />
                <Line
                  type="monotone"
                  dataKey="evaluations"
                  stroke="#3b82f6"
                  strokeWidth={2}
                  dot={{ fill: "#3b82f6", strokeWidth: 2, r: 4 }}
                />
              </LineChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      </div>

      {/* Health Summary */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Flag Health Summary</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <div className="text-center">
              <div className="text-2xl font-bold text-green-600">
                {((stats.enabled / stats.total) * 100).toFixed(0)}%
              </div>
              <p className="text-sm text-muted-foreground">Enabled Rate</p>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-blue-600">
                {Object.values(stats.byCategory).reduce(
                  (max, count) => Math.max(max, count),
                  0,
                )}
              </div>
              <p className="text-sm text-muted-foreground">Largest Category</p>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-purple-600">
                {(stats.evaluationsToday / stats.enabled).toFixed(0)}
              </div>
              <p className="text-sm text-muted-foreground">
                Avg Evaluations/Flag
              </p>
            </div>
            <div className="text-center">
              <div className="text-2xl font-bold text-amber-600">
                {stats.expiringSoon}
              </div>
              <p className="text-sm text-muted-foreground">Need Attention</p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
