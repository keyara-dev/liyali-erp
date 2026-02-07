"use client";

import { useEffect, useState } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Users,
  Building2,
  CreditCard,
  Activity,
  TrendingUp,
  AlertTriangle,
} from "lucide-react";
import { getAdminDashboardMetrics } from "@/app/_actions/dashboard";

interface DashboardData {
  total_organizations: number;
  active_organizations: number;
  trial_organizations: number;
  expiring_trials: number;
  total_users: number;
  active_users: number;
  recent_organizations: Array<{
    id: string;
    name: string;
    created_at: string;
    status: string;
  }>;
  system_health: {
    uptime: string;
    cpu_usage: number;
    memory_usage: number;
    disk_usage: number;
  };
}

export default function AdminDashboard() {
  const [data, setData] = useState<DashboardData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function fetchDashboardData() {
      try {
        setLoading(true);
        const response = await getAdminDashboardMetrics();

        if (response.success && response.data) {
          setData(response.data);
        } else {
          setError(response.message || "Failed to load dashboard data");
        }
      } catch (err) {
        setError("An error occurred while loading dashboard data");
        console.error("Dashboard error:", err);
      } finally {
        setLoading(false);
      }
    }

    fetchDashboardData();
  }, []);

  if (loading) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Admin Dashboard</h1>
          <p className="text-muted-foreground">
            System overview and management portal
          </p>
        </div>

        {/* Loading skeleton */}
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {[...Array(4)].map((_, i) => (
            <Card key={i}>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <Skeleton className="h-4 w-32" />
                <Skeleton className="h-4 w-4" />
              </CardHeader>
              <CardContent>
                <Skeleton className="h-8 w-16 mb-2" />
                <Skeleton className="h-3 w-24" />
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Admin Dashboard</h1>
          <p className="text-muted-foreground">
            System overview and management portal
          </p>
        </div>
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center space-x-2 text-red-600">
              <AlertTriangle className="h-4 w-4" />
              <span>{error}</span>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  const formatTimeAgo = (dateString: string) => {
    const date = new Date(dateString);
    const now = new Date();
    const diffInHours = Math.floor(
      (now.getTime() - date.getTime()) / (1000 * 60 * 60),
    );

    if (diffInHours < 1) return "Less than an hour ago";
    if (diffInHours < 24)
      return `${diffInHours} hour${diffInHours > 1 ? "s" : ""} ago`;

    const diffInDays = Math.floor(diffInHours / 24);
    return `${diffInDays} day${diffInDays > 1 ? "s" : ""} ago`;
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Admin Dashboard</h1>
        <p className="text-muted-foreground">
          System overview and management portal
        </p>
      </div>

      {/* Stats Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Total Organizations
            </CardTitle>
            <Building2 className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {data?.total_organizations || 0}
            </div>
            <p className="text-xs text-muted-foreground">
              {data?.active_organizations || 0} active this month
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Users</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{data?.total_users || 0}</div>
            <p className="text-xs text-muted-foreground">
              {data?.active_users || 0} active users
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Trial Organizations
            </CardTitle>
            <CreditCard className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {data?.trial_organizations || 0}
            </div>
            <p className="text-xs text-muted-foreground">
              {data?.expiring_trials || 0} expiring soon
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">System Health</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {data?.system_health?.uptime || "N/A"}
            </div>
            <p className="text-xs text-muted-foreground">Uptime this month</p>
          </CardContent>
        </Card>
      </div>

      {/* Recent Activity */}
      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Recent Organizations</CardTitle>
            <CardDescription>Latest organization registrations</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {data?.recent_organizations?.slice(0, 5).map((org) => (
                <div key={org.id} className="flex items-center justify-between">
                  <div>
                    <p className="font-medium">{org.name}</p>
                    <p className="text-sm text-muted-foreground">
                      {formatTimeAgo(org.created_at)}
                    </p>
                  </div>
                  <Badge
                    variant={org.status === "trial" ? "secondary" : "default"}
                  >
                    {org.status}
                  </Badge>
                </div>
              )) || (
                <p className="text-sm text-muted-foreground">
                  No recent organizations
                </p>
              )}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>System Alerts</CardTitle>
            <CardDescription>Important system notifications</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {data?.expiring_trials && data.expiring_trials > 0 && (
                <div className="flex items-start space-x-3">
                  <AlertTriangle className="h-4 w-4 text-yellow-500 mt-0.5" />
                  <div className="flex-1">
                    <p className="text-sm">
                      {data.expiring_trials} trial organization
                      {data.expiring_trials > 1 ? "s" : ""} expiring in 7 days
                    </p>
                    <p className="text-xs text-muted-foreground">Just now</p>
                  </div>
                </div>
              )}

              <div className="flex items-start space-x-3">
                <Activity className="h-4 w-4 text-green-500 mt-0.5" />
                <div className="flex-1">
                  <p className="text-sm">System health check completed</p>
                  <p className="text-xs text-muted-foreground">
                    CPU: {data?.system_health?.cpu_usage || 0}%, Memory:{" "}
                    {data?.system_health?.memory_usage || 0}%
                  </p>
                </div>
              </div>

              {(!data?.recent_organizations ||
                data.recent_organizations.length === 0) && (
                <div className="flex items-start space-x-3">
                  <TrendingUp className="h-4 w-4 text-blue-500 mt-0.5" />
                  <div className="flex-1">
                    <p className="text-sm">System running smoothly</p>
                    <p className="text-xs text-muted-foreground">
                      All services operational
                    </p>
                  </div>
                </div>
              )}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
