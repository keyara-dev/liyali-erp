"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from "recharts";
import { TrendingUp, Activity } from "lucide-react";
import { type SystemMetrics } from "@/app/_actions/system-health";

interface SystemMetricsChartProps {
  metrics: SystemMetrics | null;
}

export function SystemMetricsChart({ metrics }: SystemMetricsChartProps) {
  if (!metrics || !metrics.performance_history) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Activity className="h-4 w-4" />
            Performance Trends
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center h-64 text-muted-foreground">
            No performance data available
          </div>
        </CardContent>
      </Card>
    );
  }

  const chartData = metrics.performance_history.map((point) => ({
    time: new Date(point.timestamp).toLocaleTimeString([], {
      hour: "2-digit",
      minute: "2-digit",
    }),
    cpu: point.cpu_usage,
    memory: point.memory_usage,
    responseTime: point.response_time,
    rps: point.requests_per_second,
  }));

  const CustomTooltip = ({ active, payload, label }: any) => {
    if (active && payload && payload.length) {
      return (
        <div className="bg-background border rounded-lg p-3 shadow-lg">
          <p className="font-medium">{`Time: ${label}`}</p>
          {payload.map((entry: any, index: number) => (
            <p key={index} style={{ color: entry.color }}>
              {`${entry.name}: ${entry.value}${entry.name === "Response Time" ? "ms" : entry.name === "RPS" ? "/s" : "%"}`}
            </p>
          ))}
        </div>
      );
    }
    return null;
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Activity className="h-4 w-4" />
          Performance Trends
        </CardTitle>
        <div className="flex items-center gap-2">
          <Badge variant="outline" className="text-xs">
            <TrendingUp className="mr-1 h-3 w-3" />
            Last 24 hours
          </Badge>
        </div>
      </CardHeader>
      <CardContent>
        <div className="h-64">
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
              <XAxis
                dataKey="time"
                className="text-xs"
                tick={{ fontSize: 12 }}
              />
              <YAxis className="text-xs" tick={{ fontSize: 12 }} />
              <Tooltip content={<CustomTooltip />} />
              <Legend />
              <Line
                type="monotone"
                dataKey="cpu"
                stroke="#8884d8"
                strokeWidth={2}
                dot={false}
                name="CPU Usage"
              />
              <Line
                type="monotone"
                dataKey="memory"
                stroke="#82ca9d"
                strokeWidth={2}
                dot={false}
                name="Memory Usage"
              />
              <Line
                type="monotone"
                dataKey="responseTime"
                stroke="#ffc658"
                strokeWidth={2}
                dot={false}
                name="Response Time"
                yAxisId="right"
              />
              <Line
                type="monotone"
                dataKey="rps"
                stroke="#ff7300"
                strokeWidth={2}
                dot={false}
                name="RPS"
                yAxisId="right"
              />
            </LineChart>
          </ResponsiveContainer>
        </div>

        {/* Performance Summary */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mt-4 pt-4 border-t">
          <div className="text-center">
            <div className="text-sm text-muted-foreground">Avg CPU</div>
            <div className="text-lg font-semibold">
              {metrics.server?.cpu_usage || 0}%
            </div>
          </div>
          <div className="text-center">
            <div className="text-sm text-muted-foreground">Avg Memory</div>
            <div className="text-lg font-semibold">
              {metrics.server?.memory_usage || 0}%
            </div>
          </div>
          <div className="text-center">
            <div className="text-sm text-muted-foreground">Avg Response</div>
            <div className="text-lg font-semibold">
              {metrics.average_response_time || 0}ms
            </div>
          </div>
          <div className="text-center">
            <div className="text-sm text-muted-foreground">Peak Response</div>
            <div className="text-lg font-semibold">
              {metrics.api?.peak_response_time || 0}ms
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
