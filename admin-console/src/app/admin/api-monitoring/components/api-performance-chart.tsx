"use client";

import { useState, useEffect } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  AreaChart,
  Area,
  ComposedChart,
  Bar,
} from "recharts";
import { RefreshCw, TrendingUp, TrendingDown, Activity } from "lucide-react";
import { toast } from "sonner";
import {
  getAPIPerformanceData,
  getRealTimeMetrics,
  type APIPerformanceData,
} from "@/app/_actions/api-monitoring";

interface APIPerformanceChartProps {
  timeRange?: string;
  onTimeRangeChange?: (range: string) => void;
}

interface RealTimeMetrics {
  current_rps: number;
  avg_response_time: number;
  error_rate: number;
  active_connections: number;
  queue_size: number;
  cpu_usage: number;
  memory_usage: number;
  timestamp: string;
}

export function APIPerformanceChart({
  timeRange = "24h",
  onTimeRangeChange,
}: APIPerformanceChartProps) {
  const [performanceData, setPerformanceData] = useState<APIPerformanceData[]>(
    [],
  );
  const [realTimeMetrics, setRealTimeMetrics] =
    useState<RealTimeMetrics | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [selectedMetric, setSelectedMetric] = useState("response_time");

  useEffect(() => {
    loadPerformanceData();
    loadRealTimeMetrics();

    // Set up real-time updates every 30 seconds
    const interval = setInterval(() => {
      loadRealTimeMetrics();
    }, 30000);

    return () => clearInterval(interval);
  }, [timeRange]);

  const loadPerformanceData = async (isRefresh = false) => {
    if (isRefresh) {
      setIsRefreshing(true);
    } else {
      setIsLoading(true);
    }

    try {
      const interval = getIntervalForTimeRange(timeRange);
      const result = await getAPIPerformanceData(timeRange, interval);

      if (result.success) {
        setPerformanceData(result.data || []);
      } else {
        toast.error("Failed to load performance data");
      }
    } catch (error) {
      console.error("Error loading performance data:", error);
      toast.error("Failed to load performance data");
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  };

  const loadRealTimeMetrics = async () => {
    try {
      const result = await getRealTimeMetrics();
      if (result.success) {
        setRealTimeMetrics(result.data || null);
      }
    } catch (error) {
      console.error("Error loading real-time metrics:", error);
    }
  };

  const getIntervalForTimeRange = (range: string): string => {
    switch (range) {
      case "1h":
        return "1m";
      case "6h":
        return "5m";
      case "24h":
        return "15m";
      case "7d":
        return "1h";
      case "30d":
        return "6h";
      default:
        return "15m";
    }
  };

  const handleRefresh = () => {
    loadPerformanceData(true);
    loadRealTimeMetrics();
  };

  const handleTimeRangeChange = (range: string) => {
    if (onTimeRangeChange) {
      onTimeRangeChange(range);
    }
  };

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp);
    if (timeRange === "1h" || timeRange === "6h") {
      return date.toLocaleTimeString([], {
        hour: "2-digit",
        minute: "2-digit",
      });
    } else if (timeRange === "24h") {
      return date.toLocaleTimeString([], {
        hour: "2-digit",
        minute: "2-digit",
      });
    } else {
      return date.toLocaleDateString([], { month: "short", day: "numeric" });
    }
  };

  const formatTooltipValue = (value: number, name: string) => {
    switch (name) {
      case "avg_response_time":
        return [`${value.toFixed(0)}ms`, "Avg Response Time"];
      case "requests_per_minute":
        return [`${value.toFixed(0)}`, "Requests/min"];
      case "error_rate":
        return [`${value.toFixed(2)}%`, "Error Rate"];
      case "active_connections":
        return [`${value.toFixed(0)}`, "Active Connections"];
      case "cpu_usage":
        return [`${value.toFixed(1)}%`, "CPU Usage"];
      case "memory_usage":
        return [`${value.toFixed(1)}%`, "Memory Usage"];
      default:
        return [value, name];
    }
  };

  const timeRanges = [
    { value: "1h", label: "Last Hour" },
    { value: "6h", label: "Last 6 Hours" },
    { value: "24h", label: "Last 24 Hours" },
    { value: "7d", label: "Last 7 Days" },
    { value: "30d", label: "Last 30 Days" },
  ];

  const metrics = [
    { value: "response_time", label: "Response Time", color: "#8884d8" },
    { value: "requests", label: "Requests", color: "#82ca9d" },
    { value: "errors", label: "Error Rate", color: "#ffc658" },
    { value: "system", label: "System Metrics", color: "#ff7300" },
  ];

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>API Performance</CardTitle>
          <CardDescription>Loading performance data...</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="h-96 flex items-center justify-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      {/* Real-time Metrics */}
      {realTimeMetrics && (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Current RPS</CardTitle>
              <Activity className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {realTimeMetrics.current_rps.toFixed(0)}
              </div>
              <p className="text-xs text-muted-foreground">
                Requests per second
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                Response Time
              </CardTitle>
              <TrendingUp className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {realTimeMetrics.avg_response_time.toFixed(0)}ms
              </div>
              <p className="text-xs text-muted-foreground">
                Average response time
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">Error Rate</CardTitle>
              <TrendingDown className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-red-600">
                {realTimeMetrics.error_rate.toFixed(2)}%
              </div>
              <p className="text-xs text-muted-foreground">
                Current error rate
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                Active Connections
              </CardTitle>
              <Activity className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {realTimeMetrics.active_connections}
              </div>
              <p className="text-xs text-muted-foreground">
                Current connections
              </p>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Performance Charts */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>API Performance Metrics</CardTitle>
              <CardDescription>Performance trends over time</CardDescription>
            </div>
            <div className="flex items-center gap-2">
              <Select value={selectedMetric} onValueChange={setSelectedMetric}>
                <SelectTrigger className="w-40">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {metrics.map((metric) => (
                    <SelectItem key={metric.value} value={metric.value}>
                      {metric.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              <Select value={timeRange} onValueChange={handleTimeRangeChange}>
                <SelectTrigger className="w-40">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {timeRanges.map((range) => (
                    <SelectItem key={range.value} value={range.value}>
                      {range.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              <Button
                variant="outline"
                size="sm"
                onClick={handleRefresh}
                disabled={isRefreshing}
              >
                <RefreshCw
                  className={`h-4 w-4 ${isRefreshing ? "animate-spin" : ""}`}
                />
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="h-96">
            <ResponsiveContainer width="100%" height="100%">
              {(() => {
                if (selectedMetric === "response_time") {
                  return (
                    <LineChart data={performanceData}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis
                        dataKey="timestamp"
                        tickFormatter={formatTimestamp}
                      />
                      <YAxis />
                      <Tooltip
                        labelFormatter={(label) => formatTimestamp(label)}
                        formatter={formatTooltipValue}
                      />
                      <Legend />
                      <Line
                        type="monotone"
                        dataKey="avg_response_time"
                        stroke="#8884d8"
                        strokeWidth={2}
                        dot={false}
                        name="Avg Response Time"
                      />
                    </LineChart>
                  );
                }

                if (selectedMetric === "requests") {
                  return (
                    <ComposedChart data={performanceData}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis
                        dataKey="timestamp"
                        tickFormatter={formatTimestamp}
                      />
                      <YAxis yAxisId="left" />
                      <YAxis yAxisId="right" orientation="right" />
                      <Tooltip
                        labelFormatter={(label) => formatTimestamp(label)}
                        formatter={formatTooltipValue}
                      />
                      <Legend />
                      <Bar
                        yAxisId="left"
                        dataKey="requests_per_minute"
                        fill="#82ca9d"
                        name="Requests/min"
                      />
                      <Line
                        yAxisId="right"
                        type="monotone"
                        dataKey="active_connections"
                        stroke="#8884d8"
                        strokeWidth={2}
                        dot={false}
                        name="Active Connections"
                      />
                    </ComposedChart>
                  );
                }

                if (selectedMetric === "errors") {
                  return (
                    <AreaChart data={performanceData}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis
                        dataKey="timestamp"
                        tickFormatter={formatTimestamp}
                      />
                      <YAxis />
                      <Tooltip
                        labelFormatter={(label) => formatTimestamp(label)}
                        formatter={formatTooltipValue}
                      />
                      <Legend />
                      <Area
                        type="monotone"
                        dataKey="error_rate"
                        stroke="#ffc658"
                        fill="#ffc658"
                        fillOpacity={0.3}
                        name="Error Rate"
                      />
                    </AreaChart>
                  );
                }

                if (selectedMetric === "system") {
                  return (
                    <LineChart data={performanceData}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis
                        dataKey="timestamp"
                        tickFormatter={formatTimestamp}
                      />
                      <YAxis />
                      <Tooltip
                        labelFormatter={(label) => formatTimestamp(label)}
                        formatter={formatTooltipValue}
                      />
                      <Legend />
                      <Line
                        type="monotone"
                        dataKey="cpu_usage"
                        stroke="#ff7300"
                        strokeWidth={2}
                        dot={false}
                        name="CPU Usage"
                      />
                      <Line
                        type="monotone"
                        dataKey="memory_usage"
                        stroke="#8dd1e1"
                        strokeWidth={2}
                        dot={false}
                        name="Memory Usage"
                      />
                    </LineChart>
                  );
                }

                return (
                  <LineChart data={performanceData}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis
                      dataKey="timestamp"
                      tickFormatter={formatTimestamp}
                    />
                    <YAxis />
                    <Tooltip
                      labelFormatter={(label) => formatTimestamp(label)}
                      formatter={formatTooltipValue}
                    />
                    <Legend />
                    <Line
                      type="monotone"
                      dataKey="avg_response_time"
                      stroke="#8884d8"
                      strokeWidth={2}
                      dot={false}
                      name="Avg Response Time"
                    />
                  </LineChart>
                );
              })()}
            </ResponsiveContainer>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
