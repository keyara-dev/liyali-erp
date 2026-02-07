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
import {
  Clock,
  AlertTriangle,
  RefreshCw,
  Calendar,
  Building2,
  Users,
  CheckCircle,
  XCircle,
  Search,
} from "lucide-react";
import { Input } from "@/components/ui/input";
import { toast } from "sonner";
import { TrialResetDialog } from "@/components/trial-reset-dialog";
import {
  getTrialOrganizations,
  type TrialResetRequest,
} from "@/app/_actions/subscriptions";

export function TrialManagementTab() {
  const [organizations, setOrganizations] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [searchTerm, setSearchTerm] = useState("");
  const [statusFilter, setStatusFilter] = useState<
    "all" | "active" | "expiring" | "expired"
  >("all");

  const loadTrialOrganizations = async () => {
    setIsLoading(true);
    try {
      const result = await getTrialOrganizations();
      if (result.success) {
        setOrganizations(result.data || []);
      } else {
        toast.error(result.message || "Failed to load trial organizations");
      }
    } catch (error) {
      toast.error("Failed to load trial organizations");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadTrialOrganizations();
  }, []);

  const handleTrialReset = () => {
    // Refresh the data after trial reset
    loadTrialOrganizations();
  };

  const getStatusBadge = (status: string, daysRemaining: number) => {
    if (status === "expired" || daysRemaining < 0) {
      return <Badge variant="destructive">Expired</Badge>;
    }
    if (daysRemaining <= 7) {
      return <Badge variant="warning">Expiring Soon</Badge>;
    }
    return <Badge variant="success">Active</Badge>;
  };

  // Filter organizations
  const filteredOrganizations = organizations.filter((org) => {
    const matchesSearch = org.name
      .toLowerCase()
      .includes(searchTerm.toLowerCase());

    if (!matchesSearch) return false;

    switch (statusFilter) {
      case "active":
        return org.status === "active" && org.days_remaining > 7;
      case "expiring":
        return (
          org.status === "active" &&
          org.days_remaining <= 7 &&
          org.days_remaining > 0
        );
      case "expired":
        return org.status === "expired" || org.days_remaining < 0;
      default:
        return true;
    }
  });

  const activeTrials = organizations.filter(
    (org) => org.status === "active" && org.days_remaining > 0,
  );
  const expiringSoon = organizations.filter(
    (org) => org.status === "active" && org.days_remaining <= 7,
  );
  const expired = organizations.filter(
    (org) => org.status === "expired" || org.days_remaining < 0,
  );

  if (isLoading) {
    return <div>Loading trial organizations...</div>;
  }

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-lg font-medium">Trial Management</h3>
        <p className="text-sm text-muted-foreground">
          Manage organization trial periods and monitor expiring trials
        </p>
      </div>

      {/* Summary Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Active Trials</CardTitle>
            <CheckCircle className="h-4 w-4 text-green-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{activeTrials.length}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Expiring Soon</CardTitle>
            <AlertTriangle className="h-4 w-4 text-yellow-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{expiringSoon.length}</div>
            <p className="text-xs text-muted-foreground">Within 7 days</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Expired</CardTitle>
            <XCircle className="h-4 w-4 text-red-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{expired.length}</div>
            <p className="text-xs text-muted-foreground">Require attention</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Users</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {organizations.reduce((sum, org) => sum + org.user_count, 0)}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search organizations..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-10"
          />
        </div>
        <div className="flex gap-2">
          <Button
            variant={statusFilter === "all" ? "default" : "outline"}
            size="sm"
            onClick={() => setStatusFilter("all")}
          >
            All ({organizations.length})
          </Button>
          <Button
            variant={statusFilter === "active" ? "default" : "outline"}
            size="sm"
            onClick={() => setStatusFilter("active")}
          >
            Active ({activeTrials.length})
          </Button>
          <Button
            variant={statusFilter === "expiring" ? "default" : "outline"}
            size="sm"
            onClick={() => setStatusFilter("expiring")}
          >
            Expiring ({expiringSoon.length})
          </Button>
          <Button
            variant={statusFilter === "expired" ? "default" : "outline"}
            size="sm"
            onClick={() => setStatusFilter("expired")}
          >
            Expired ({expired.length})
          </Button>
        </div>
      </div>

      {/* Organizations List */}
      <Card>
        <CardHeader>
          <CardTitle>Trial Organizations</CardTitle>
          <CardDescription>
            Manage trial periods for all organizations
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {filteredOrganizations.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                No organizations found matching your criteria
              </div>
            ) : (
              filteredOrganizations.map((org) => (
                <div
                  key={org.id}
                  className="flex items-center justify-between rounded-lg border p-4 hover:bg-muted/50"
                >
                  <div className="flex items-center space-x-4">
                    <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-muted">
                      <Building2 className="h-5 w-5" />
                    </div>
                    <div>
                      <h3 className="font-medium">{org.name}</h3>
                      <div className="flex items-center space-x-4 text-sm text-muted-foreground">
                        <span className="flex items-center">
                          <Users className="mr-1 h-3 w-3" />
                          {org.user_count} users
                        </span>
                        <span className="flex items-center">
                          <Calendar className="mr-1 h-3 w-3" />
                          {org.trial_start_date} - {org.trial_end_date}
                        </span>
                      </div>
                    </div>
                  </div>

                  <div className="flex items-center space-x-4">
                    <div className="text-right">
                      <div className="flex items-center space-x-2">
                        {getStatusBadge(org.status, org.days_remaining)}
                        <span className="text-sm font-medium">
                          {org.days_remaining > 0
                            ? `${org.days_remaining} days left`
                            : `Expired ${Math.abs(org.days_remaining)} days ago`}
                        </span>
                      </div>
                      <div className="text-xs text-muted-foreground capitalize">
                        {org.subscription_tier} tier
                      </div>
                    </div>

                    <TrialResetDialog
                      organization={org}
                      onSuccess={handleTrialReset}
                    />
                  </div>
                </div>
              ))
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
