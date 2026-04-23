"use client";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  Zap,
  Globe,
  Lock,
  AlertTriangle,
  Clock,
  TrendingUp,
  TrendingDown,
  Activity,
  Server,
  CheckCircle,
  XCircle,
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
  CartesianGrid,
  Tooltip,
  Legend,
  LineChart,
  Line,
} from "recharts";
import type { APIStats } from "@/app/_actions/api-monitoring";

interface APIStatsGridProps {
  stats: APIStats | null;
  isLoading: boolean;
}

const COLORS = ["#0088FE", "#00C49F", "#FFBB28", "#FF8042", "#8884D8"];

export function APIStatsGrid({ stats, isLoading }: APIStatsGridProps) {
  if (isLoading) {
    return (
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {[...Array(8)].map((_, i) => (
          <Card key={i}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <div className="h-4 w-20 bg-muted animate-pulse rounded" />
              <div className="h-4 w-4 bg-muted animate-pulse rounded" />
            </CardHeader>
            <CardContent>
              <div className="h-8 w-16 bg-muted animate-pulse rounded mb-2" />
              <div className="h-3 w-24 bg-muted animate-pulse rounded" />
            </CardContent>
          </Card>
        ))}
      </div>
    );
  }

  if (!stats) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center h-32">
          <p className="text-muted-foreground">No statistics available</p>
        </CardContent>
      </Card>
    );
  }

  const categoryData = (stats.endpoints_by_category ?? []).map((item) => ({
    name: item.category,
    value: item.count ?? 0,
    percentage: item.percentage ?? 0,
  }));

  const methodData = (stats.requests_by_method ?? []).map((item) => ({
    name: item.method,
    value: item.count ?? 0,
    percentage: item.percentage ?? 0,
  }));

  const errorData = (stats.error_distribution ?? []).map((item) => ({
    name: (item.status_code ?? 0).toString(),
    value: item.count ?? 0,
    percentage: item.percentage ?? 0,
  }));

  const topEndpointsData = (stats.top_endpoints ?? []).slice(0, 5);
  const slowestEndpointsData = (stats.slowest_endpoints ?? []).slice(0, 5);
  const getEndpointKey = (
    endpoint: { endpoint_id?: string; path: string; method: string },
    index: number,
  ) => endpoint.endpoint_id || `${endpoint.method}-${endpoint.path}-${index}`;

  return (
    <div className="space-y-6">
      {/* Main Stats Grid */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Total Endpoints
            </CardTitle>
            <Zap className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.total_endpoints}</div>
            <p className="text-xs text-muted-foreground">
              {stats.active_endpoints} active, {stats.deprecated_endpoints}{" "}
              deprecated
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Requests Today
            </CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {(stats.total_requests_today ?? 0).toLocaleString()}
            </div>
            <p className="text-xs text-muted-foreground">
              {stats.total_errors_today ?? 0} errors (
              {stats.total_requests_today
                ? (
                    ((stats.total_errors_today ?? 0) /
                      stats.total_requests_today) *
                    100
                  ).toFixed(2)
                : "0.00"}
              %)
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Avg Response Time
            </CardTitle>
            <Clock className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {(stats.avg_response_time_today ?? 0).toFixed(0)}ms
            </div>
            <p className="text-xs text-muted-foreground">Today's average</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">System Uptime</CardTitle>
            <Server className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">
              {(stats.uptime_percentage ?? 0).toFixed(2)}%
            </div>
            <p className="text-xs text-muted-foreground">Last 30 days</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Error Rate</CardTitle>
            <AlertTriangle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-600">
              {(stats.error_rate_today ?? 0).toFixed(2)}%
            </div>
            <p className="text-xs text-muted-foreground">Today's error rate</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Active Alerts</CardTitle>
            <AlertTriangle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-orange-600">
              {stats.active_alerts}
            </div>
            <p className="text-xs text-muted-foreground">
              {stats.critical_alerts} critical
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Public Endpoints
            </CardTitle>
            <Globe className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.public_endpoints}</div>
            <p className="text-xs text-muted-foreground">
              {stats.total_endpoints
                ? (
                    ((stats.public_endpoints ?? 0) /
                      stats.total_endpoints) *
                    100
                  ).toFixed(1)
                : "0.0"}
              % of total
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Private Endpoints
            </CardTitle>
            <Lock className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.private_endpoints}</div>
            <p className="text-xs text-muted-foreground">
              {stats.total_endpoints
                ? (
                    ((stats.private_endpoints ?? 0) /
                      stats.total_endpoints) *
                    100
                  ).toFixed(1)
                : "0.0"}
              % of total
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Charts Grid */}
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {/* Endpoints by Category */}
        <Card className="col-span-1">
          <CardHeader>
            <CardTitle className="text-lg">Endpoints by Category</CardTitle>
            <CardDescription>
              Distribution of API endpoints by category
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="h-64">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie
                    data={categoryData}
                    cx="50%"
                    cy="50%"
                    labelLine={false}
                    label={({ name, percentage }) =>
                      `${name} (${(percentage ?? 0).toFixed(1)}%)`
                    }
                    outerRadius={80}
                    fill="#8884d8"
                    dataKey="value"
                  >
                    {categoryData.map((entry, index) => (
                      <Cell
                        key={`${entry.name}-${index}`}
                        fill={COLORS[index % COLORS.length]}
                      />
                    ))}
                  </Pie>
                  <Tooltip />
                </PieChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>

        {/* Requests by Method */}
        <Card className="col-span-1">
          <CardHeader>
            <CardTitle className="text-lg">Requests by Method</CardTitle>
            <CardDescription>HTTP method distribution</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="h-64">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={methodData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="name" />
                  <YAxis />
                  <Tooltip />
                  <Bar dataKey="value" fill="#0088FE" />
                </BarChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>

        {/* Error Distribution */}
        <Card className="col-span-1">
          <CardHeader>
            <CardTitle className="text-lg">Error Distribution</CardTitle>
            <CardDescription>Errors by status code</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="h-64">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={errorData}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="name" />
                  <YAxis />
                  <Tooltip />
                  <Bar dataKey="value" fill="#FF8042" />
                </BarChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Top Endpoints Tables */}
      <div className="grid gap-6 md:grid-cols-2">
        {/* Top Endpoints by Requests */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Top Endpoints by Requests</CardTitle>
            <CardDescription>
              Most frequently accessed API endpoints
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {topEndpointsData.map((endpoint, index) => (
                <div
                  key={getEndpointKey(endpoint, index)}
                  className="flex items-center justify-between p-3 border rounded-lg"
                >
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <Badge variant="outline" className="text-xs">
                        {endpoint.method}
                      </Badge>
                      <span className="font-mono text-sm">{endpoint.path}</span>
                    </div>
                    <div className="flex items-center gap-4 mt-1 text-xs text-muted-foreground">
                      <span>
                        {(endpoint.request_count ?? 0).toLocaleString()} requests
                      </span>
                      <span>{(endpoint.avg_response_time ?? 0).toFixed(0)}ms avg</span>
                      <span
                        className={`${endpoint.error_rate > 5 ? "text-red-600" : "text-green-600"}`}
                      >
                        {(endpoint.error_rate ?? 0).toFixed(1)}% errors
                      </span>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="text-lg font-bold">#{index + 1}</div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>

        {/* Slowest Endpoints */}
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Slowest Endpoints</CardTitle>
            <CardDescription>
              Endpoints with highest response times
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {slowestEndpointsData.map((endpoint, index) => (
                <div
                  key={getEndpointKey(endpoint, index)}
                  className="flex items-center justify-between p-3 border rounded-lg"
                >
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <Badge variant="outline" className="text-xs">
                        {endpoint.method}
                      </Badge>
                      <span className="font-mono text-sm">{endpoint.path}</span>
                    </div>
                    <div className="flex items-center gap-4 mt-1 text-xs text-muted-foreground">
                      <span>{(endpoint.avg_response_time ?? 0).toFixed(0)}ms avg</span>
                      <span>{(endpoint.p95_response_time ?? 0).toFixed(0)}ms p95</span>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="text-lg font-bold text-orange-600">
                      {(endpoint.avg_response_time ?? 0).toFixed(0)}ms
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
