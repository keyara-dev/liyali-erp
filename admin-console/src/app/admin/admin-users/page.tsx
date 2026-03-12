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
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  UserCog,
  Plus,
  RefreshCw,
  Shield,
  Lock,
  Unlock,
  CheckCircle,
  XCircle,
  Clock,
  Monitor,
} from "lucide-react";
import { toast } from "sonner";
import {
  exportAdminUsers,
  type AdminUser,
  type AdminUserStats,
  type AdminUserFilters,
  type AdminRole,
} from "@/app/_actions/admin-users";
import {
  useAdminUsers,
  useAdminUserStats,
  useAdminRoles,
} from "@/hooks/use-admin-users";
import { AdminUserFiltersComponent } from "./components/admin-user-filters";
import { AdminUserStatsGrid } from "./components/admin-user-stats-grid";
import { AdminUserCreateDialog } from "./components/admin-user-create-dialog";
import { AdminUserEditDialog } from "./components/admin-user-edit-dialog";
import { AdminUserDetailsDialog } from "./components/admin-user-details-dialog";
import { AdminUserActionsDropdown } from "./components/admin-user-actions-dropdown";
import { AdminUserBulkActions } from "./components/admin-user-bulk-actions";

export default function AdminUsersPage() {
  const [selectedUsers, setSelectedUsers] = useState<string[]>([]);

  // Dialog states
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [showEditDialog, setShowEditDialog] = useState(false);
  const [showDetailsDialog, setShowDetailsDialog] = useState(false);
  const [selectedUser, setSelectedUser] = useState<AdminUser | null>(null);

  // Filters
  const [filters, setFilters] = useState<AdminUserFilters>({});
  const [searchTerm, setSearchTerm] = useState("");

  // Debounced search
  useEffect(() => {
    const delayedSearch = setTimeout(() => {
      if (searchTerm !== (filters.search || "")) {
        setFilters((prev) => ({ ...prev, search: searchTerm || undefined }));
      }
    }, 500);

    return () => clearTimeout(delayedSearch);
  }, [searchTerm]);

  // TanStack Query hooks
  const {
    data: users = [],
    isLoading,
    error: usersError,
    refetch: refetchUsers,
    isRefetching,
  } = useAdminUsers(filters);
  const { data: stats = null } = useAdminUserStats();
  const { data: roles = [] } = useAdminRoles();

  useEffect(() => {
    if (usersError) toast.error("Failed to load admin users");
  }, [usersError]);

  const handleRefresh = () => {
    refetchUsers();
  };

  const handleFiltersChange = (newFilters: AdminUserFilters) => {
    setFilters(newFilters);
  };

  const handleResetFilters = () => {
    setFilters({});
    setSearchTerm("");
  };

  const handleExport = async (format: "csv" | "json" | "excel") => {
    try {
      const result = await exportAdminUsers(format, filters);
      if (result.success && result.data) {
        const blob = new Blob([JSON.stringify(result.data, null, 2)], {
          type: "application/json",
        });
        const url = URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = `admin-users-export-${new Date().toISOString().split("T")[0]}.json`;
        a.click();
        URL.revokeObjectURL(url);
        toast.success("Admin users exported successfully");
      } else {
        toast.error(result.message || "Failed to export admin users");
      }
    } catch (error) {
      console.error("Error exporting admin users:", error);
      toast.error("Failed to export admin users");
    }
  };

  const handleUserSelect = (userId: string, checked: boolean) => {
    if (checked) {
      setSelectedUsers((prev) => [...prev, userId]);
    } else {
      setSelectedUsers((prev) => prev.filter((id) => id !== userId));
    }
  };

  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      setSelectedUsers(users.map((user) => user.id));
    } else {
      setSelectedUsers([]);
    }
  };

  const handleUserAction = (action: string, user: AdminUser) => {
    setSelectedUser(user);

    switch (action) {
      case "edit":
        setShowEditDialog(true);
        break;
      case "view":
        setShowDetailsDialog(true);
        break;
    }
  };

  const handleUserUpdated = () => {
    setShowCreateDialog(false);
    setShowEditDialog(false);
    setSelectedUser(null);
  };

  const getUserInitials = (user: AdminUser) => {
    return `${user.first_name.charAt(0)}${user.last_name.charAt(0)}`.toUpperCase();
  };

  const getStatusBadge = (user: AdminUser) => {
    if (user.is_locked) {
      return (
        <Badge variant="destructive" className="flex items-center gap-1">
          <Lock className="h-3 w-3" />
          Locked
        </Badge>
      );
    }
    if (!user.is_active) {
      return (
        <Badge variant="secondary" className="flex items-center gap-1">
          <XCircle className="h-3 w-3" />
          Inactive
        </Badge>
      );
    }
    return (
      <Badge variant="default" className="flex items-center gap-1">
        <CheckCircle className="h-3 w-3" />
        Active
      </Badge>
    );
  };

  return (
    <div className="flex-1 space-y-4 p-4 md:p-8 pt-6">
      <div className="flex items-center justify-between space-y-2">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Admin Users</h2>
          <p className="text-muted-foreground">
            Manage admin users and their access to the system
          </p>
        </div>
        <div className="flex items-center space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={handleRefresh}
            disabled={isRefetching}
          >
            <RefreshCw
              className={`mr-2 h-4 w-4 ${isRefetching ? "animate-spin" : ""}`}
            />
            Refresh
          </Button>
          <Button onClick={() => setShowCreateDialog(true)}>
            <Plus className="mr-2 h-4 w-4" />
            Create Admin User
          </Button>
        </div>
      </div>

      {/* Filters */}
      <AdminUserFiltersComponent
        filters={filters}
        onFiltersChange={handleFiltersChange}
        onReset={handleResetFilters}
        onExport={handleExport}
        searchTerm={searchTerm}
        onSearchChange={setSearchTerm}
        roles={roles}
      />

      {/* Stats Grid */}
      <AdminUserStatsGrid stats={stats} isLoading={isLoading} />

      {/* Bulk Actions */}
      {selectedUsers.length > 0 && (
        <AdminUserBulkActions
          selectedUsers={selectedUsers}
          onActionComplete={() => {
            setSelectedUsers([]);
          }}
          roles={roles}
        />
      )}

      {/* Users Table */}
      <Card>
        <CardHeader>
          <CardTitle>Admin Users</CardTitle>
          <CardDescription>
            {users.length} total admin users • {selectedUsers.length} selected
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-4">
              {[...Array(5)].map((_, i) => (
                <div
                  key={i}
                  className="flex items-center space-x-4 p-4 border rounded-lg"
                >
                  <div className="h-4 w-4 bg-muted animate-pulse rounded" />
                  <div className="h-10 w-10 bg-muted animate-pulse rounded-full" />
                  <div className="flex-1 space-y-2">
                    <div className="h-4 bg-muted animate-pulse rounded w-1/4" />
                    <div className="h-3 bg-muted animate-pulse rounded w-1/2" />
                  </div>
                  <div className="h-6 w-16 bg-muted animate-pulse rounded" />
                </div>
              ))}
            </div>
          ) : users.length === 0 ? (
            <div className="text-center py-8">
              <UserCog className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <p className="text-muted-foreground">No admin users found</p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-12">
                    <input
                      type="checkbox"
                      checked={selectedUsers.length === users.length}
                      onChange={(e) => handleSelectAll(e.target.checked)}
                      className="rounded border-gray-300"
                    />
                  </TableHead>
                  <TableHead>User</TableHead>
                  <TableHead>Roles</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Last Login</TableHead>
                  <TableHead>Sessions</TableHead>
                  <TableHead>2FA</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {users.map((user) => (
                  <TableRow key={user.id}>
                    <TableCell>
                      <input
                        type="checkbox"
                        checked={selectedUsers.includes(user.id)}
                        onChange={(e) =>
                          handleUserSelect(user.id, e.target.checked)
                        }
                        className="rounded border-gray-300"
                      />
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center space-x-3">
                        <Avatar className="h-8 w-8">
                          <AvatarImage src={user.avatar_url} />
                          <AvatarFallback className="text-xs">
                            {getUserInitials(user)}
                          </AvatarFallback>
                        </Avatar>
                        <div>
                          <div className="font-medium flex items-center gap-2">
                            {user.full_name}
                            {user.is_super_admin && (
                              <Shield className="h-4 w-4 text-red-600" />
                            )}
                          </div>
                          <div className="text-sm text-muted-foreground">
                            {user.email}
                          </div>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-1">
                        {user.roles.slice(0, 2).map((role) => (
                          <Badge
                            key={role.id}
                            variant="outline"
                            className="text-xs"
                          >
                            {role.display_name}
                          </Badge>
                        ))}
                        {user.roles.length > 2 && (
                          <Badge variant="outline" className="text-xs">
                            +{user.roles.length - 2} more
                          </Badge>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>{getStatusBadge(user)}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Clock className="h-4 w-4 text-muted-foreground" />
                        <span className="text-sm">
                          {user.last_login_at
                            ? new Date(user.last_login_at).toLocaleDateString()
                            : "Never"}
                        </span>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Monitor className="h-4 w-4 text-muted-foreground" />
                        <Badge variant="outline">{user.session_count}</Badge>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant={
                          user.two_factor_enabled ? "default" : "secondary"
                        }
                      >
                        {user.two_factor_enabled ? "Enabled" : "Disabled"}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <AdminUserActionsDropdown
                        user={user}
                        onAction={handleUserAction}
                        onUserUpdated={handleUserUpdated}
                        currentUserId="current-user-id"
                      />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* Dialogs */}
      <AdminUserCreateDialog
        open={showCreateDialog}
        onOpenChange={setShowCreateDialog}
        onUserCreated={handleUserUpdated}
      />

      <AdminUserEditDialog
        open={showEditDialog}
        onOpenChange={setShowEditDialog}
        user={selectedUser}
        roles={roles}
        onUserUpdated={handleUserUpdated}
      />

      <AdminUserDetailsDialog
        open={showDetailsDialog}
        onOpenChange={setShowDetailsDialog}
        user={selectedUser}
      />
    </div>
  );
}
