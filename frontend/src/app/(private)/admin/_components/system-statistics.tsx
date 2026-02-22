"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { useSystemStats } from "@/hooks/use-reports-queries";
import {
  FileText,
  Clock,
  CheckCircle2,
  AlertCircle,
  TrendingUp,
} from "lucide-react";
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from "recharts";

export function SystemStatistics() {
  // Fetch live statistics from database
  const { data: stats, isLoading, error } = useSystemStats();

  if (isLoading) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        Loading system statistics...
      </div>
    );
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>
          Failed to load system statistics. Please try again.
        </AlertDescription>
      </Alert>
    );
  }

  if (!stats) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        No statistics available
      </div>
    );
  }

  // Prepare chart data for document types
  const chartData = [
    { name: "Requisitions", count: stats.documentTypeBreakdown.requisitions },
    {
      name: "Purchase Orders",
      count: stats.documentTypeBreakdown.purchaseOrders,
    },
    {
      name: "Payment Vouchers",
      count: stats.documentTypeBreakdown.paymentVouchers,
    },
    { name: "GRN", count: stats.documentTypeBreakdown.grn },
    { name: "Budgets", count: stats.documentTypeBreakdown.budgets },
  ];

  return (
    <div className="space-y-6">
      {/* Key Metrics Grid */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Total Documents
            </CardTitle>
            <FileText className="h-5 w-5 text-primary" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">{stats.totalDocuments}</div>
            <p className="text-xs text-muted-foreground mt-1">All time</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Approval Rate
            </CardTitle>
            <TrendingUp className="h-5 w-5 text-secondary" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">
              {stats.approvalRate.toFixed(1)}%
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              {stats.approvedDocuments} approved
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Avg Approval Time
            </CardTitle>
            <Clock className="h-5 w-5 text-accent" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">
              {stats.averageApprovalTime.toFixed(1)}
            </div>
            <p className="text-xs text-muted-foreground mt-1">days</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Rejection Rate
            </CardTitle>
            <AlertCircle className="h-5 w-5 text-destructive" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold">
              {stats.rejectionRate.toFixed(1)}%
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              {stats.rejectedDocuments} rejected
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Document Type Distribution Chart */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Document Type Distribution</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="h-80 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="name" />
                <YAxis />
                <Tooltip />
                <Bar dataKey="count" fill="var(--primary)" />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </CardContent>
      </Card>

      {/* Status Summary Table */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Status Summary</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {[
              {
                label: "Draft",
                value: stats.statusBreakdown.draft,
                variant: "outline" as const,
              },
              {
                label: "Submitted",
                value: stats.statusBreakdown.submitted,
                variant: "secondary" as const,
              },
              {
                label: "In Review",
                value: stats.statusBreakdown.inReview,
                variant: "default" as const,
              },
              {
                label: "Approved",
                value: stats.statusBreakdown.approved,
                variant: "default" as const,
              },
              {
                label: "Rejected",
                value: stats.statusBreakdown.rejected,
                variant: "destructive" as const,
              },
            ].map((item) => (
              <div
                key={item.label}
                className="flex items-center justify-between p-3 border rounded-lg"
              >
                <span className="font-medium">{item.label}</span>
                <Badge variant={item.variant}>{item.value}</Badge>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
