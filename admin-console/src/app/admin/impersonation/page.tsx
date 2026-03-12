"use client";

import { useState } from "react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { SelectField } from "@/components/ui/select-field";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Eye, RefreshCw, Shield, Users, AlertTriangle, Activity } from "lucide-react";
import { toast } from "sonner";
import {
  useImpersonationLogs,
  useImpersonationStats,
  useRevokeImpersonationLog,
} from "@/hooks/use-impersonation";
import type { ImpersonationLogFilters } from "@/app/_actions/impersonation";

export default function ImpersonationPage() {
  const [filters, setFilters] = useState<ImpersonationLogFilters>({});
  const [impersonatorSearch, setImpersonatorSearch] = useState("");
  const [targetSearch, setTargetSearch] = useState("");

  const {
    data: logsData,
    isLoading,
    refetch,
    isRefetching,
  } = useImpersonationLogs(filters);
  const { data: stats } = useImpersonationStats();
  const revokeMutation = useRevokeImpersonationLog();

  const logs = (logsData as any[]) ?? [];

  const handleRevoke = (id: string) => {
    if (!confirm("Mark this impersonation log as revoked? This cannot be undone.")) return;
    revokeMutation.mutate(id, {
      onSuccess: (result) => {
        if (result.success) {
          toast.success("Impersonation log revoked");
        } else {
          toast.error(result.message || "Failed to revoke");
        }
      },
    });
  };

  const handleSearch = () => {
    setFilters((prev) => ({
      ...prev,
      impersonatorId: impersonatorSearch || undefined,
      targetId: targetSearch || undefined,
    }));
  };

  const handleReset = () => {
    setFilters({});
    setImpersonatorSearch("");
    setTargetSearch("");
  };

  return (
    <div className="flex-1 space-y-4 p-4 md:p-8 pt-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">
            Impersonation Logs
          </h2>
          <p className="text-muted-foreground">
            Audit trail of all impersonation sessions — visible to super admins only
          </p>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={() => refetch()}
          disabled={isRefetching}
        >
          <RefreshCw
            className={`mr-2 h-4 w-4 ${isRefetching ? "animate-spin" : ""}`}
          />
          Refresh
        </Button>
      </div>

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Sessions</CardTitle>
            <Eye className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.total ?? "—"}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Active Now</CardTitle>
            <Activity className="h-4 w-4 text-green-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-green-600">
              {stats?.active ?? "—"}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Platform Users</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.platform_user ?? "—"}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Admin Users</CardTitle>
            <Shield className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.admin_user ?? "—"}</div>
          </CardContent>
        </Card>
      </div>

      {/* Filters */}
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Filters</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-3">
            <Input
              placeholder="Impersonator user ID"
              value={impersonatorSearch}
              onChange={(e) => setImpersonatorSearch(e.target.value)}
              className="w-52"
            />
            <Input
              placeholder="Target user ID"
              value={targetSearch}
              onChange={(e) => setTargetSearch(e.target.value)}
              className="w-52"
            />
            <SelectField
              options={[
                { value: "all", label: "All types" },
                { value: "platform_user", label: "Platform User" },
                { value: "admin_user", label: "Admin User" },
              ]}
              value={filters.impersonationType ?? "all"}
              onValueChange={(v) =>
                setFilters((prev) => ({
                  ...prev,
                  impersonationType: v === "all" ? undefined : (v as "platform_user" | "admin_user"),
                }))
              }
              classNames={{ wrapper: "w-44" }}
            />
            <SelectField
              options={[
                { value: "all", label: "All statuses" },
                { value: "not_revoked", label: "Not revoked" },
                { value: "revoked", label: "Revoked" },
              ]}
              value={filters.revoked === undefined ? "all" : filters.revoked ? "revoked" : "not_revoked"}
              onValueChange={(v) =>
                setFilters((prev) => ({
                  ...prev,
                  revoked: v === "all" ? undefined : v === "revoked" ? true : false,
                }))
              }
              classNames={{ wrapper: "w-36" }}
            />
            <Button onClick={handleSearch} size="sm">
              Search
            </Button>
            <Button onClick={handleReset} variant="outline" size="sm">
              Reset
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Table */}
      <Card>
        <CardHeader>
          <CardTitle>Session Log</CardTitle>
          <CardDescription>{logs.length} entries</CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-3">
              {[...Array(5)].map((_, i) => (
                <div key={i} className="h-12 bg-muted animate-pulse rounded" />
              ))}
            </div>
          ) : logs.length === 0 ? (
            <div className="text-center py-8">
              <AlertTriangle className="h-10 w-10 text-muted-foreground mx-auto mb-3" />
              <p className="text-muted-foreground">No impersonation logs found</p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Impersonator</TableHead>
                  <TableHead>Target</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Date</TableHead>
                  <TableHead>Expires</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {logs.map((log: any) => {
                  const isExpired = new Date(log.expires_at) < new Date();
                  const isActive = !log.revoked && !isExpired;
                  return (
                    <TableRow key={log.id}>
                      <TableCell>
                        <div className="font-medium text-sm">
                          {log.impersonator_email}
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="font-medium text-sm">
                          {log.target_email}
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge variant="outline" className="text-xs capitalize">
                          {log.impersonation_type?.replace("_", " ")}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-sm">
                        {new Date(log.created_at).toLocaleString()}
                      </TableCell>
                      <TableCell className="text-sm">
                        {new Date(log.expires_at).toLocaleString()}
                      </TableCell>
                      <TableCell>
                        {log.revoked ? (
                          <Badge variant="destructive">Revoked</Badge>
                        ) : isExpired ? (
                          <Badge variant="secondary">Expired</Badge>
                        ) : (
                          <Badge variant="default">Active</Badge>
                        )}
                      </TableCell>
                      <TableCell>
                        {isActive && (
                          <Button
                            variant="destructive"
                            size="sm"
                            onClick={() => handleRevoke(log.id)}
                            disabled={revokeMutation.isPending}
                          >
                            Revoke
                          </Button>
                        )}
                      </TableCell>
                    </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
