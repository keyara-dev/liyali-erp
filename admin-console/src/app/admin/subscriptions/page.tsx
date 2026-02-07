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
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Skeleton } from "@/components/ui/skeleton";
import {
  CreditCard,
  Settings,
  Clock,
  BarChart3,
  Plus,
  Users,
  Building2,
  TrendingUp,
  AlertTriangle,
} from "lucide-react";
import { toast } from "sonner";
import { getSubscriptionStatistics } from "@/app/_actions/subscriptions";

// Import tab components
import { SubscriptionTiersTab } from "./components/subscription-tiers-tab";
import { FeaturesManagementTab } from "./components/features-management-tab";
import { TrialManagementTab } from "./components/trial-management-tab";
import { SubscriptionAnalyticsTab } from "./components/subscription-analytics-tab";

export default function SubscriptionsPage() {
  const [activeTab, setActiveTab] = useState("tiers");
  const [stats, setStats] = useState<any>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    loadStatistics();
  }, []);

  const loadStatistics = async () => {
    try {
      const result = await getSubscriptionStatistics();
      if (result.success && result.data) {
        setStats(result.data);
      } else {
        // Fallback to default values if API fails
        setStats({
          total_tiers: 4,
          active_subscriptions: 0,
          trial_organizations: 0,
          monthly_revenue: 0,
          revenue_growth: 0,
        });
      }
    } catch (error) {
      console.error("Failed to load subscription statistics:", error);
      // Fallback to default values
      setStats({
        total_tiers: 4,
        active_subscriptions: 0,
        trial_organizations: 0,
        monthly_revenue: 0,
        revenue_growth: 0,
      });
    } finally {
      setIsLoading(false);
    }
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">
            Subscription Management
          </h1>
          <p className="text-muted-foreground">
            Manage subscription tiers, features, trials, and analytics
          </p>
        </div>

        {/* Loading skeleton */}
        <div className="grid gap-4 md:grid-cols-4">
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

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">
          Subscription Management
        </h1>
        <p className="text-muted-foreground">
          Manage subscription tiers, features, trials, and analytics
        </p>
      </div>

      {/* Overview Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Subscription Tiers
            </CardTitle>
            <CreditCard className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.total_tiers || 0}</div>
            <p className="text-xs text-muted-foreground">
              Active pricing tiers
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Active Subscriptions
            </CardTitle>
            <Users className="h-4 w-4 text-green-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {stats?.active_subscriptions || 0}
            </div>
            <p className="text-xs text-muted-foreground">
              Paying organizations
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Trial Organizations
            </CardTitle>
            <Clock className="h-4 w-4 text-blue-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {stats?.trial_organizations || 0}
            </div>
            <p className="text-xs text-muted-foreground">Currently on trial</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Monthly Revenue
            </CardTitle>
            <TrendingUp className="h-4 w-4 text-green-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              ${(stats?.monthly_revenue || 0).toLocaleString()}
            </div>
            <p className="text-xs text-muted-foreground">
              {stats?.revenue_growth > 0 ? "+" : ""}
              {stats?.revenue_growth || 0}% from last month
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Main Tabs */}
      <Card>
        <CardHeader>
          <CardTitle>Subscription Management</CardTitle>
          <CardDescription>
            Comprehensive subscription and billing management
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Tabs
            value={activeTab}
            onValueChange={setActiveTab}
            className="space-y-4"
          >
            <TabsList className="grid w-full grid-cols-4">
              <TabsTrigger value="tiers" className="flex items-center gap-2">
                <CreditCard className="h-4 w-4" />
                Subscription Tiers
              </TabsTrigger>
              <TabsTrigger value="features" className="flex items-center gap-2">
                <Settings className="h-4 w-4" />
                Features
              </TabsTrigger>
              <TabsTrigger value="trials" className="flex items-center gap-2">
                <Clock className="h-4 w-4" />
                Trial Management
                {stats?.trial_organizations > 0 && (
                  <Badge variant="secondary" className="ml-1">
                    {stats.trial_organizations}
                  </Badge>
                )}
              </TabsTrigger>
              <TabsTrigger
                value="analytics"
                className="flex items-center gap-2"
              >
                <BarChart3 className="h-4 w-4" />
                Analytics
              </TabsTrigger>
            </TabsList>

            <TabsContent value="tiers" className="space-y-4">
              <SubscriptionTiersTab />
            </TabsContent>

            <TabsContent value="features" className="space-y-4">
              <FeaturesManagementTab />
            </TabsContent>

            <TabsContent value="trials" className="space-y-4">
              <TrialManagementTab />
            </TabsContent>

            <TabsContent value="analytics" className="space-y-4">
              <SubscriptionAnalyticsTab />
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </div>
  );
}
