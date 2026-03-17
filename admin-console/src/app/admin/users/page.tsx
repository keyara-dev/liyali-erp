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
import { Input } from "@/components/ui/input";
import {
  Users,
  Filter,
  UserCheck,
  UserX,
  UserPlus,
  Eye,
  Edit,
  MoreHorizontal,
  Building2,
  Mail,
  Phone,
  Calendar,
  Activity,
  Shield,
  AlertTriangle,
} from "lucide-react";
import { notify } from "@/lib/utils";
import {
  type PlatformUser,
  type UserFilters,
} from "@/app/_actions/users";
import {
  useUsers,
  useUserStats,
  useUpdateUserStatus,
} from "@/hooks/use-users";
import { UserDetailsDialog } from "./components/user-details-dialog";
import { UserActionsDropdown } from "./components/user-actions-dropdown";
import { UserBulkActions } from "./components/user-bulk-actions";
import { UserAdvancedFilters } from "./components/user-advanced-filters";
import { UserCreateDialog } from "./components/user-create-dialog";
import { Checkbox } from "@/components/ui/checkbox";

export default function UsersPage() {
  const [selectedUser, setSelectedUser] = useState<PlatformUser | null>(null);
  const [showUserDetails, setShowUserDetails] = useState(false);
  const [selectedUsers, setSelectedUsers] = useState<string[]>([]);
  const [showCreateUser, setShowCreateUser] = useState(false);

  // Filters and pagination
  const [filters, setFilters] = useState<UserFilters>({
    search: "",
    status: "all",
    page: 1,
    limit: 20,
    sort_by: "created_at",
    sort_order: "desc",
  });

  // TanStack Query hooks
  const { data: userData, isLoading, error: userError } = useUsers(filters);
  const { data: statsData } = useUserStats();
  const updateStatusMutation = useUpdateUserStatus();

  const users = userData?.users ?? [];
  const pagination = {
    total: userData?.total ?? 0,
    page: userData?.page ?? 1,
    limit: userData?.limit ?? 20,
    totalPages: userData?.totalPages ?? 0,
  };

  const stats = statsData ?? {
    total_users: 0,
    active_users: 0,
    suspended_users: 0,
    pending_users: 0,
    users_created_this_month: 0,
    users_logged_in_today: 0,
  };

  useEffect(() => {
    if (userError) notify({ title: "Failed to load users", type: "error" });
  }, [userError]);

  const handleStatusChange = async (
    userId: string,
    status: "active" | "suspended" | "inactive",
  ) => {
    updateStatusMutation.mutate(
      { id: userId, status },
      {
        onSuccess: (result) => {
          if (result.success) {
            notify({ title: `User ${status} successfully`, type: "success" });
          } else {
            notify({ title: result.message || "Failed to update user status", type: "error" });
          }
        },
        onError: () => notify({ title: "Failed to update user status", type: "error" }),
      },
    );
  };

  const handlePageChange = (page: number) => {
    setFilters((prev) => ({ ...prev, page }));
  };

  const handleUserSelection = (userId: string, checked: boolean) => {
    if (checked) {
      setSelectedUsers((prev) => [...prev, userId]);
    } else {
      setSelectedUsers((prev) => prev.filter((id) => id !== userId));
    }
  };

  const handleFiltersChange = (newFilters: UserFilters) => {
    setFilters(newFilters);
  };

  const handleFiltersReset = () => {
    setFilters({
      search: "",
      status: "all",
      page: 1,
      limit: 20,
      sort_by: "created_at",
      sort_order: "desc",
    });
  };

  const getStatusBadge = (status: string, emailVerified: boolean) => {
    if (status === "suspended") {
      return <Badge variant="destructive">Suspended</Badge>;
    }
    if (status === "pending" || !emailVerified) {
      return <Badge variant="warning">Pending</Badge>;
    }
    if (status === "inactive") {
      return <Badge variant="secondary">Inactive</Badge>;
    }
    return <Badge variant="success">Active</Badge>;
  };

  const getRoleBadge = (role: string) => {
    const roleColors = {
      admin: "bg-red-100 text-red-800",
      manager: "bg-blue-100 text-blue-800",
      user: "bg-gray-100 text-gray-800",
      viewer: "bg-green-100 text-green-800",
    };

    return (
      <Badge
        variant="outline"
        className={
          roleColors[role as keyof typeof roleColors] ||
          "bg-gray-100 text-gray-800"
        }
      >
        {role}
      </Badge>
    );
  };

  if (isLoading && users.length === 0) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
            <Users className="h-8 w-8" />
            Platform Users
          </h1>
          <p className="text-muted-foreground">Manage platform user accounts</p>
        </div>
        <div className="grid gap-4 md:grid-cols-4">
          {[...Array(4)].map((_, i) => (
            <Card key={i}>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <div className="h-4 w-24 bg-muted animate-pulse rounded" />
                <div className="h-4 w-4 bg-muted animate-pulse rounded" />
              </CardHeader>
              <CardContent>
                <div className="h-8 w-16 bg-muted animate-pulse rounded mb-2" />
                <div className="h-3 w-20 bg-muted animate-pulse rounded" />
              </CardContent>
            </Card>
          ))}
        </div>
        <Card>
          <CardContent className="p-6">
            <div className="space-y-3">
              {[...Array(5)].map((_, i) => (
                <div key={i} className="flex items-center gap-4">
                  <div className="h-10 w-10 bg-muted animate-pulse rounded-full" />
                  <div className="flex-1 space-y-2">
                    <div className="h-4 w-48 bg-muted animate-pulse rounded" />
                    <div className="h-3 w-32 bg-muted animate-pulse rounded" />
                  </div>
                  <div className="h-6 w-16 bg-muted animate-pulse rounded-full" />
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">User Management</h1>
          <p className="text-muted-foreground">
            Manage all platform users and their organization memberships
          </p>
        </div>
        <Button onClick={() => setShowCreateUser(true)}>
          <UserPlus className="mr-2 h-4 w-4" />
          Create User
        </Button>
      </div>

      {/* Statistics Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Users</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.total_users}</div>
            <p className="text-xs text-muted-foreground">
              +{stats.users_created_this_month} this month
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Active Users</CardTitle>
            <UserCheck className="h-4 w-4 text-green-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.active_users}</div>
            <p className="text-xs text-muted-foreground">
              {stats.users_logged_in_today} logged in today
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Suspended Users
            </CardTitle>
            <UserX className="h-4 w-4 text-red-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.suspended_users}</div>
            <p className="text-xs text-muted-foreground">Require attention</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Pending Users</CardTitle>
            <AlertTriangle className="h-4 w-4 text-yellow-600" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.pending_users}</div>
            <p className="text-xs text-muted-foreground">
              Awaiting verification
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Filters and Search */}
      <Card>
        <CardHeader>
          <CardTitle>All Users</CardTitle>
          <CardDescription>
            Manage and support all platform users
          </CardDescription>
        </CardHeader>
        <CardContent>
          {/* Advanced Filters */}
          <UserAdvancedFilters
            filters={filters}
            onFiltersChange={handleFiltersChange}
            onReset={handleFiltersReset}
          />

          {/* Bulk Actions */}
          <UserBulkActions
            users={users}
            selectedUsers={selectedUsers}
            onSelectionChange={setSelectedUsers}
            onUsersUpdated={() => {}}
          />

          {/* Users List */}
          <div className="space-y-4">
            {users.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                No users found matching your criteria
              </div>
            ) : (
              users.map((user) => (
                <div
                  key={user.id}
                  className="flex items-center justify-between rounded-lg border p-4 hover:bg-muted/50"
                >
                  <div className="flex items-center space-x-4">
                    <Checkbox
                      checked={selectedUsers.includes(user.id)}
                      onCheckedChange={(checked) =>
                        handleUserSelection(user.id, checked as boolean)
                      }
                      aria-label={`Select ${user.name}`}
                    />
                    <div className="flex h-12 w-12 items-center justify-center rounded-full bg-primary/10">
                      <Users className="h-6 w-6 text-primary" />
                    </div>
                    <div>
                      <div className="flex items-center gap-2">
                        <h3 className="font-semibold">{user.name}</h3>
                        {getStatusBadge(user.status, user.email_verified)}
                        {getRoleBadge(user.role)}
                      </div>
                      <div className="flex items-center space-x-4 text-sm text-muted-foreground">
                        <span className="flex items-center">
                          <Mail className="mr-1 h-3 w-3" />
                          {user.email}
                        </span>
                        {user.phone && (
                          <span className="flex items-center">
                            <Phone className="mr-1 h-3 w-3" />
                            {user.phone}
                          </span>
                        )}
                        <span className="flex items-center">
                          <Building2 className="mr-1 h-3 w-3" />
                          {user.organizations.length} org
                          {user.organizations.length !== 1 ? "s" : ""}
                        </span>
                        <span className="flex items-center">
                          <Calendar className="mr-1 h-3 w-3" />
                          {new Date(user.created_at).toLocaleDateString()}
                        </span>
                      </div>
                      {user.organizations.length > 0 && (
                        <div className="flex items-center gap-2 mt-2">
                          <span className="text-xs text-muted-foreground">
                            Organizations:
                          </span>
                          {user.organizations.slice(0, 3).map((org) => (
                            <Badge
                              key={org.organization_id}
                              variant="outline"
                              className="text-xs"
                            >
                              {org.organization_name}
                              {org.is_primary && (
                                <span className="ml-1 text-primary">★</span>
                              )}
                            </Badge>
                          ))}
                          {user.organizations.length > 3 && (
                            <Badge variant="outline" className="text-xs">
                              +{user.organizations.length - 3} more
                            </Badge>
                          )}
                        </div>
                      )}
                    </div>
                  </div>

                  <div className="flex items-center space-x-2">
                    <div className="text-right text-sm">
                      <div className="font-medium">
                        {user.last_login
                          ? `Last login: ${new Date(user.last_login).toLocaleDateString()}`
                          : "Never logged in"}
                      </div>
                      <div className="text-muted-foreground">
                        {user.login_count} total logins
                      </div>
                    </div>

                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => {
                        setSelectedUser(user);
                        setShowUserDetails(true);
                      }}
                    >
                      <Eye className="mr-2 h-4 w-4" />
                      View Details
                    </Button>

                    <UserActionsDropdown
                      user={user}
                      onStatusChange={handleStatusChange}
                      onUserUpdated={() => {}}
                    />
                  </div>
                </div>
              ))
            )}
          </div>

          {/* Pagination */}
          {pagination.totalPages > 1 && (
            <div className="flex items-center justify-between mt-6">
              <div className="text-sm text-muted-foreground">
                Showing {(pagination.page - 1) * pagination.limit + 1} to{" "}
                {Math.min(pagination.page * pagination.limit, pagination.total)}{" "}
                of {pagination.total} users
              </div>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handlePageChange(pagination.page - 1)}
                  disabled={pagination.page <= 1}
                >
                  Previous
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handlePageChange(pagination.page + 1)}
                  disabled={pagination.page >= pagination.totalPages}
                >
                  Next
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* User Details Dialog */}
      {selectedUser && (
        <UserDetailsDialog
          user={selectedUser}
          open={showUserDetails}
          onOpenChange={setShowUserDetails}
          onUserUpdated={() => {}}
        />
      )}

      {/* User Create Dialog */}
      <UserCreateDialog
        open={showCreateUser}
        onOpenChange={setShowCreateUser}
        onUserCreated={() => {}}
      />
    </div>
  );
}
