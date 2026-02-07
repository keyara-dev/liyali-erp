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
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Shield,
  RefreshCw,
  Eye,
  AlertTriangle,
  Clock,
  User,
  Building2,
  Activity,
  CheckCircle,
  XCircle,
  ChevronLeft,
  ChevronRight,
} from "lucide-react";
import { format } from "date-fns";
import { toast } from "sonner";
import {
  getAuditLogs,
  getAuditLogStats,
  getAuditLogAnalytics,
  exportAuditLogs,
  type AuditLog,
  type AuditLogFilters,
  type AuditLogStats,
  type AuditLogAnalytics,
} from "@/app/_actions/audit-logs";
import { AuditLogFiltersComponent } from "./components/audit-log-filters";
import { AuditLogStatsGrid } from "./components/audit-log-stats-grid";
import { AuditLogDetailsSheet } from "./components/audit-log-details-sheet";
import { SecurityEventsPanel } from "./components/security-events-panel";
import { AuditLogAnalyticsCharts } from "./components/audit-log-analytics-charts";

export default function AuditLogsPage() {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [stats, setStats] = useState<AuditLogStats | null>(null);
  const [analytics, setAnalytics] = useState<AuditLogAnalytics | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [selectedLog, setSelectedLog] = useState<AuditLog | null>(null);
  const [showDetailsSheet, setShowDetailsSheet] = useState(false);
  const [activeTab, setActiveTab] = useState("logs");

  // Pagination
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalLogs, setTotalLogs] = useState(0);
  const pageSize = 50;

  // Filters
  const [filters, setFilters] = useState<AuditLogFilters>({
    date_range: "24h",
  });

  // Search
  const [searchTerm, setSearchTerm] = useState("");

  useEffect(() => {
    loadAuditData();
  }, [filters, currentPage]);

  useEffect(() => {
    const delayedSearch = setTimeout(() => {
      if (searchTerm !== (filters.search || "")) {
        setFilters((prev) => ({ ...prev, search: searchTerm || undefined }));
        setCurrentPage(1);
      }
    }, 500);

    return () => clearTimeout(delayedSearch);
  }, [searchTerm]);

  const loadAuditData = async (isRefresh = false) => {
    if (isRefresh) {
      setIsRefreshing(true);
    } else {
      setIsLoading(true);
    }

    try {
      // Load all data in parallel
      const [logsResult, statsResult, analyticsResult] = await Promise.all([
        getAuditLogs(filters, currentPage, pageSize),
        getAuditLogStats(filters),
        getAuditLogAnalytics(filters),
      ]);

      // Handle logs result
      if (logsResult.success && logsResult.data) {
        setLogs(logsResult.data.logs);
        setTotalLogs(logsResult.data.total);
        setTotalPages(logsResult.data.totalPages);
      } else {
        toast.error("Failed to load audit logs");
      }

      // Handle stats result
      if (statsResult.success) {
        setStats(statsResult.data || null);
      } else {
        toast.error("Failed to load audit statistics");
      }

      // Handle analytics result
      if (analyticsResult.success) {
        setAnalytics(analyticsResult.data || null);
      } else {
        toast.error("Failed to load audit analytics");
      }
    } catch (error) {
      console.error("Error loading audit data:", error);
      toast.error("Failed to load audit data");
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  };

  const handleRefresh = () => {
    loadAuditData(true);
  };

  const handleFiltersChange = (newFilters: AuditLogFilters) => {
    setFilters(newFilters);
    setCurrentPage(1);
  };

  const handleResetFilters = () => {
    setFilters({ date_range: "24h" });
    setSearchTerm("");
    setCurrentPage(1);
  };

  const handleExport = async (format: "csv" | "json" | "pdf") => {
    try {
      const result = await exportAuditLogs(format, filters);
      if (result.success) {
        toast.success(
          `Audit logs export initiated. Download will be available shortly.`,
        );
        if (result.data?.download_url) {
          window.open(result.data.download_url, "_blank");
        }
      } else {
        toast.error("Failed to export audit logs");
      }
    } catch (error) {
      console.error("Error exporting audit logs:", error);
      toast.error("Failed to export audit logs");
    }
  };

  const handleLogClick = (log: AuditLog) => {
    setSelectedLog(log);
    setShowDetailsSheet(true);
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case "critical":
        return "text-red-600";
      case "high":
        return "text-orange-600";
      case "medium":
        return "text-yellow-600";
      case "low":
        return "text-green-600";
      default:
        return "text-gray-600";
    }
  };

  const getSeverityBadge = (severity: string) => {
    switch (severity) {
      case "critical":
        return (
          <Badge variant="destructive" className="text-xs">
            Critical
          </Badge>
        );
      case "high":
        return (
          <Badge variant="destructive" className="text-xs">
            High
          </Badge>
        );
      case "medium":
        return (
          <Badge variant="secondary" className="text-xs">
            Medium
          </Badge>
        );
      case "low":
        return (
          <Badge variant="outline" className="text-xs">
            Low
          </Badge>
        );
      default:
        return (
          <Badge variant="outline" className="text-xs">
            Unknown
          </Badge>
        );
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "success":
        return <CheckCircle className="h-4 w-4 text-green-600" />;
      case "failure":
        return <XCircle className="h-4 w-4 text-red-600" />;
      case "warning":
        return <AlertTriangle className="h-4 w-4 text-yellow-600" />;
      default:
        return <Activity className="h-4 w-4 text-gray-600" />;
    }
  };

  return (
    <div className="flex-1 space-y-4 p-4 md:p-8 pt-6">
      <div className="flex items-center justify-between space-y-2">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Audit Logs</h2>
          <p className="text-muted-foreground">
            Monitor and analyze system activities and security events
          </p>
        </div>
        <div className="flex items-center space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={handleRefresh}
            disabled={isRefreshing}
          >
            <RefreshCw
              className={`mr-2 h-4 w-4 ${isRefreshing ? "animate-spin" : ""}`}
            />
            Refresh
          </Button>
        </div>
      </div>

      {/* Filters */}
      <AuditLogFiltersComponent
        filters={filters}
        onFiltersChange={handleFiltersChange}
        onReset={handleResetFilters}
        onExport={handleExport}
        searchTerm={searchTerm}
        onSearchChange={setSearchTerm}
      />

      {/* Stats Grid */}
      <AuditLogStatsGrid stats={stats} isLoading={isLoading} />

      {/* Main Content Tabs */}
      <Tabs
        value={activeTab}
        onValueChange={setActiveTab}
        className="space-y-4"
      >
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="logs" className="flex items-center gap-2">
            <Activity className="h-4 w-4" />
            Audit Logs
          </TabsTrigger>
          <TabsTrigger value="security" className="flex items-center gap-2">
            <Shield className="h-4 w-4" />
            Security Events
          </TabsTrigger>
          <TabsTrigger value="analytics" className="flex items-center gap-2">
            <Eye className="h-4 w-4" />
            Analytics
          </TabsTrigger>
          <TabsTrigger value="compliance" className="flex items-center gap-2">
            <AlertTriangle className="h-4 w-4" />
            Compliance
          </TabsTrigger>
        </TabsList>

        <TabsContent value="logs" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Audit Log Entries</CardTitle>
              <CardDescription>
                {totalLogs.toLocaleString()} total entries • Page {currentPage}{" "}
                of {totalPages}
              </CardDescription>
            </CardHeader>
            <CardContent>
              {isLoading ? (
                <div className="space-y-4">
                  {[...Array(10)].map((_, i) => (
                    <div
                      key={i}
                      className="flex items-center space-x-4 p-4 border rounded-lg"
                    >
                      <div className="h-4 w-4 bg-muted animate-pulse rounded" />
                      <div className="flex-1 space-y-2">
                        <div className="h-4 bg-muted animate-pulse rounded w-1/4" />
                        <div className="h-3 bg-muted animate-pulse rounded w-1/2" />
                      </div>
                      <div className="h-6 w-16 bg-muted animate-pulse rounded" />
                    </div>
                  ))}
                </div>
              ) : logs.length === 0 ? (
                <div className="text-center py-8">
                  <Activity className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                  <p className="text-muted-foreground">No audit logs found</p>
                </div>
              ) : (
                <>
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Status</TableHead>
                        <TableHead>Action</TableHead>
                        <TableHead>User</TableHead>
                        <TableHead>Resource</TableHead>
                        <TableHead>Severity</TableHead>
                        <TableHead>Timestamp</TableHead>
                        <TableHead>Actions</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {logs.map((log) => (
                        <TableRow
                          key={log.id}
                          className="cursor-pointer hover:bg-muted/50"
                        >
                          <TableCell>
                            <div className="flex items-center gap-2">
                              {getStatusIcon(log.status)}
                            </div>
                          </TableCell>
                          <TableCell>
                            <div>
                              <div className="font-medium">{log.action}</div>
                              <div className="text-sm text-muted-foreground capitalize">
                                {log.action_type}
                              </div>
                            </div>
                          </TableCell>
                          <TableCell>
                            <div className="flex items-center gap-2">
                              <User className="h-4 w-4 text-muted-foreground" />
                              <div>
                                <div className="font-medium">
                                  {log.user_name}
                                </div>
                                <div className="text-sm text-muted-foreground">
                                  {log.user_email}
                                </div>
                              </div>
                            </div>
                          </TableCell>
                          <TableCell>
                            <div className="flex items-center gap-2">
                              {log.organization_name ? (
                                <Building2 className="h-4 w-4 text-muted-foreground" />
                              ) : (
                                <Activity className="h-4 w-4 text-muted-foreground" />
                              )}
                              <div>
                                <div className="font-medium capitalize">
                                  {log.resource_type}
                                </div>
                                {log.organization_name && (
                                  <div className="text-sm text-muted-foreground">
                                    {log.organization_name}
                                  </div>
                                )}
                              </div>
                            </div>
                          </TableCell>
                          <TableCell>
                            {getSeverityBadge(log.severity)}
                          </TableCell>
                          <TableCell>
                            <div className="flex items-center gap-2">
                              <Clock className="h-4 w-4 text-muted-foreground" />
                              <div>
                                <div className="text-sm">
                                  {format(
                                    new Date(log.timestamp),
                                    "MMM dd, HH:mm",
                                  )}
                                </div>
                                <div className="text-xs text-muted-foreground">
                                  {format(new Date(log.timestamp), "yyyy")}
                                </div>
                              </div>
                            </div>
                          </TableCell>
                          <TableCell>
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => handleLogClick(log)}
                            >
                              <Eye className="h-4 w-4" />
                            </Button>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>

                  {/* Pagination */}
                  {totalPages > 1 && (
                    <div className="flex items-center justify-between mt-4">
                      <div className="text-sm text-muted-foreground">
                        Showing {(currentPage - 1) * pageSize + 1} to{" "}
                        {Math.min(currentPage * pageSize, totalLogs)} of{" "}
                        {totalLogs} entries
                      </div>
                      <div className="flex items-center gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() =>
                            setCurrentPage((prev) => Math.max(1, prev - 1))
                          }
                          disabled={currentPage === 1}
                        >
                          <ChevronLeft className="h-4 w-4" />
                          Previous
                        </Button>
                        <span className="text-sm">
                          Page {currentPage} of {totalPages}
                        </span>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() =>
                            setCurrentPage((prev) =>
                              Math.min(totalPages, prev + 1),
                            )
                          }
                          disabled={currentPage === totalPages}
                        >
                          Next
                          <ChevronRight className="h-4 w-4" />
                        </Button>
                      </div>
                    </div>
                  )}
                </>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="security" className="space-y-4">
          <SecurityEventsPanel stats={stats} isLoading={isLoading} />
        </TabsContent>

        <TabsContent value="analytics" className="space-y-4">
          <AuditLogAnalyticsCharts
            analytics={analytics}
            stats={stats}
            isLoading={isLoading}
          />
        </TabsContent>

        <TabsContent value="compliance" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Compliance Dashboard</CardTitle>
              <CardDescription>
                Compliance monitoring and reporting features
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="text-center py-8">
                <AlertTriangle className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                <p className="text-muted-foreground">
                  Compliance features coming soon
                </p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* Audit Log Details Sheet */}
      <AuditLogDetailsSheet
        log={selectedLog}
        open={showDetailsSheet}
        onOpenChange={setShowDetailsSheet}
      />
    </div>
  );
}
