"use client";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { DollarSign, TrendingUp, TrendingDown, Users } from "lucide-react";
import { useRevenueAnalytics, useOrganizationAnalytics } from "@/hooks/use-analytics";

function formatCurrency(value: number) {
  if (value >= 1_000_000) return `$${(value / 1_000_000).toFixed(2)}M`;
  if (value >= 1_000) return `$${(value / 1_000).toFixed(1)}K`;
  return `$${value.toFixed(0)}`;
}

interface RevenueKPICardProps {
  title: string;
  value: string;
  subtitle?: string;
  trend?: number;
  isLoading?: boolean;
}

function RevenueKPICard({ title, value, subtitle, trend, isLoading }: RevenueKPICardProps) {
  if (isLoading) {
    return (
      <Card>
        <CardHeader className="pb-2">
          <Skeleton className="h-4 w-28" />
        </CardHeader>
        <CardContent>
          <Skeleton className="h-8 w-24 mb-1" />
          <Skeleton className="h-3 w-20" />
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{value}</div>
        <div className="flex items-center gap-2 mt-1">
          {subtitle && (
            <p className="text-xs text-muted-foreground">{subtitle}</p>
          )}
          {trend != null && (
            <div
              className={`flex items-center gap-0.5 text-xs ${
                trend >= 0 ? "text-green-600" : "text-red-600"
              }`}
            >
              {trend >= 0 ? (
                <TrendingUp className="h-3 w-3" />
              ) : (
                <TrendingDown className="h-3 w-3" />
              )}
              {Math.abs(trend).toFixed(1)}%
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

export function SubscriptionsTab() {
  const { data: revenueData, isLoading: revenueLoading } = useRevenueAnalytics();
  const { data: orgData, isLoading: orgLoading } = useOrganizationAnalytics();

  const tierDistribution = orgData?.organization_distribution?.by_subscription_tier ?? [];
  const revenueByTier = revenueData?.revenue_by_tier ?? [];
  const churnRate = revenueData?.financial_metrics?.churn_rate ?? 0;
  const trialConversionRate = orgData?.trial_metrics?.trial_conversion_rate ?? 0;

  return (
    <div className="space-y-6">
      {/* Revenue KPIs */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <RevenueKPICard
          title="Monthly Recurring Revenue"
          value={formatCurrency(revenueData?.monthly_recurring_revenue ?? 0)}
          trend={revenueData?.revenue_growth_rate}
          isLoading={revenueLoading}
        />
        <RevenueKPICard
          title="Annual Recurring Revenue"
          value={formatCurrency(revenueData?.annual_recurring_revenue ?? 0)}
          isLoading={revenueLoading}
        />
        <RevenueKPICard
          title="Trial Conversion Rate"
          value={`${trialConversionRate.toFixed(1)}%`}
          subtitle="trials → paid"
          isLoading={orgLoading}
        />
        <RevenueKPICard
          title="Churn Rate"
          value={`${churnRate.toFixed(1)}%`}
          subtitle="this period"
          isLoading={revenueLoading}
        />
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        {/* Revenue by tier */}
        <Card>
          <CardHeader>
            <CardTitle>Revenue by Tier</CardTitle>
            <CardDescription>Breakdown of revenue per subscription plan</CardDescription>
          </CardHeader>
          <CardContent>
            {revenueLoading ? (
              <div className="space-y-3">
                {[...Array(4)].map((_, i) => (
                  <div key={i} className="flex items-center justify-between">
                    <Skeleton className="h-4 w-24" />
                    <Skeleton className="h-4 w-16" />
                  </div>
                ))}
              </div>
            ) : revenueByTier.length > 0 ? (
              <div className="space-y-3">
                {revenueByTier.map((tier, index) => (
                  <div
                    key={tier.tier || tier.revenue || `revenue-tier-${index}`}
                    className="space-y-1"
                  >
                    <div className="flex items-center justify-between text-sm">
                      <div className="flex items-center gap-2">
                        <span className="font-medium capitalize">{tier.tier ?? "—"}</span>
                        <Badge variant="secondary" className="text-xs">
                          {tier.subscriber_count ?? 0} orgs
                        </Badge>
                      </div>
                      <div className="flex items-center gap-2">
                        <span className="text-muted-foreground text-xs">
                          {(tier.percentage ?? 0).toFixed(1)}%
                        </span>
                        <span className="font-semibold">
                          {formatCurrency(tier.revenue ?? 0)}
                        </span>
                      </div>
                    </div>
                    <div className="h-2 rounded-full bg-muted overflow-hidden">
                      <div
                        className="h-full rounded-full bg-primary"
                        style={{ width: `${tier.percentage ?? 0}%` }}
                      />
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-sm text-muted-foreground text-center py-6">
                No revenue data available
              </p>
            )}
          </CardContent>
        </Card>

        {/* Org tier distribution */}
        <Card>
          <CardHeader>
            <CardTitle>Organization Distribution</CardTitle>
            <CardDescription>Organizations by subscription tier</CardDescription>
          </CardHeader>
          <CardContent>
            {orgLoading ? (
              <div className="space-y-3">
                {[...Array(4)].map((_, i) => (
                  <div key={i} className="flex items-center justify-between">
                    <Skeleton className="h-4 w-24" />
                    <Skeleton className="h-4 w-12" />
                  </div>
                ))}
              </div>
            ) : tierDistribution.length > 0 ? (
              <div className="space-y-3">
                {tierDistribution.map((tier: any, index: number) => {
                  const total = tierDistribution.reduce(
                    (sum: number, t: any) => sum + (t.count ?? 0),
                    0,
                  );
                  const pct = total > 0 ? ((tier.count ?? 0) / total) * 100 : 0;
                  return (
                    <div
                      key={tier.tier || tier.name || `tier-${index}`}
                      className="space-y-1"
                    >
                      <div className="flex items-center justify-between text-sm">
                        <div className="flex items-center gap-2">
                          <Users className="h-3.5 w-3.5 text-muted-foreground" />
                          <span className="font-medium capitalize">
                            {tier.tier ?? tier.name}
                          </span>
                        </div>
                        <div className="flex items-center gap-2">
                          <span className="text-muted-foreground text-xs">
                            {pct.toFixed(1)}%
                          </span>
                          <span className="font-semibold">{tier.count ?? 0}</span>
                        </div>
                      </div>
                      <div className="h-2 rounded-full bg-muted overflow-hidden">
                        <div
                          className="h-full rounded-full bg-blue-500"
                          style={{ width: `${pct}%` }}
                        />
                      </div>
                    </div>
                  );
                })}
              </div>
            ) : (
              <p className="text-sm text-muted-foreground text-center py-6">
                No distribution data available
              </p>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Additional financial metrics */}
      {!revenueLoading && revenueData?.financial_metrics && (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">Avg Revenue / User</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {formatCurrency(revenueData.financial_metrics.average_revenue_per_user ?? 0)}
              </div>
              <p className="text-xs text-muted-foreground">ARPU</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">Customer Lifetime Value</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {formatCurrency(revenueData.financial_metrics.customer_lifetime_value ?? 0)}
              </div>
              <p className="text-xs text-muted-foreground">LTV</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">Net Revenue Retention</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {(revenueData.financial_metrics.net_revenue_retention ?? 0).toFixed(1)}%
              </div>
              <p className="text-xs text-muted-foreground">NRR</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium">Trials Expiring Soon</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">
                {orgData?.trial_metrics?.trials_expiring_soon ?? 0}
              </div>
              <p className="text-xs text-muted-foreground">next 7 days</p>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}
