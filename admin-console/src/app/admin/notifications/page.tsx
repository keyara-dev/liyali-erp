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
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Bell,
  Plus,
  Search,
  RefreshCw,
  Trash2,
  Mail,
  MailOpen,
  Send,
  CheckCircle,
  AlertTriangle,
  Megaphone,
} from "lucide-react";
import { toast } from "sonner";
import {
  useAdminNotifications,
  useAdminNotificationStats,
  useCreateAdminNotification,
  useDeleteAdminNotification,
  useMarkAdminNotificationRead,
  useBulkDeleteAdminNotifications,
} from "@/hooks/use-notifications";
import type {
  NotificationFilters,
  CreateNotificationRequest,
} from "@/app/_actions/notifications";

export default function NotificationsPage() {
  const [filters, setFilters] = useState<NotificationFilters>({
    page: 1,
    limit: 20,
  });
  const [searchTerm, setSearchTerm] = useState("");
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [selectedIds, setSelectedIds] = useState<string[]>([]);

  // Debounced search
  useEffect(() => {
    const timeout = setTimeout(() => {
      if (searchTerm !== (filters.search || "")) {
        setFilters((prev) => ({ ...prev, search: searchTerm || undefined }));
      }
    }, 500);
    return () => clearTimeout(timeout);
  }, [searchTerm]);

  const {
    data: notifications = [],
    isLoading,
    error,
    refetch,
    isRefetching,
  } = useAdminNotifications(filters);
  const { data: stats } = useAdminNotificationStats();
  const createMutation = useCreateAdminNotification();
  const deleteMutation = useDeleteAdminNotification();
  const markReadMutation = useMarkAdminNotificationRead();
  const bulkDeleteMutation = useBulkDeleteAdminNotifications();

  useEffect(() => {
    if (error) toast.error("Failed to load notifications");
  }, [error]);

  const handleCreate = (req: CreateNotificationRequest) => {
    createMutation.mutate(req, {
      onSuccess: () => {
        toast.success("Notification created successfully");
        setShowCreateDialog(false);
      },
      onError: () => toast.error("Failed to create notification"),
    });
  };

  const handleDelete = (id: string) => {
    if (confirm("Are you sure you want to delete this notification?")) {
      deleteMutation.mutate(id, {
        onSuccess: () => toast.success("Notification deleted"),
        onError: () => toast.error("Failed to delete notification"),
      });
    }
  };

  const handleMarkRead = (id: string) => {
    markReadMutation.mutate(id, {
      onSuccess: () => toast.success("Marked as read"),
      onError: () => toast.error("Failed to mark as read"),
    });
  };

  const handleBulkDelete = () => {
    if (selectedIds.length === 0) return;
    if (
      confirm(
        `Are you sure you want to delete ${selectedIds.length} notifications?`,
      )
    ) {
      bulkDeleteMutation.mutate(selectedIds, {
        onSuccess: () => {
          toast.success(`${selectedIds.length} notifications deleted`);
          setSelectedIds([]);
        },
        onError: () => toast.error("Failed to delete notifications"),
      });
    }
  };

  const handleSelectToggle = (id: string) => {
    setSelectedIds((prev) =>
      prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id],
    );
  };

  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      setSelectedIds(
        (notifications as Array<{ id: string }>).map((n) => n.id),
      );
    } else {
      setSelectedIds([]);
    }
  };

  const getImportanceBadge = (importance: string) => {
    switch (importance?.toUpperCase()) {
      case "HIGH":
        return <Badge variant="destructive">High</Badge>;
      case "MEDIUM":
        return <Badge variant="default">Medium</Badge>;
      case "LOW":
        return <Badge variant="secondary">Low</Badge>;
      default:
        return <Badge variant="outline">{importance || "Unknown"}</Badge>;
    }
  };

  const getTypeBadge = (type: string) => {
    switch (type) {
      case "approval_required":
        return <Badge variant="default">Approval</Badge>;
      case "document_approved":
        return (
          <Badge className="bg-green-100 text-green-800">Approved</Badge>
        );
      case "document_rejected":
        return <Badge variant="destructive">Rejected</Badge>;
      case "admin_announcement":
        return (
          <Badge className="bg-blue-100 text-blue-800">Announcement</Badge>
        );
      default:
        return <Badge variant="outline">{type}</Badge>;
    }
  };

  return (
    <div className="flex-1 space-y-4 p-4 md:p-8 pt-6">
      <div className="flex items-center justify-between space-y-2">
        <div>
          <h2 className="text-3xl font-bold tracking-tight">Notifications</h2>
          <p className="text-muted-foreground">
            Manage platform notifications and send announcements
          </p>
        </div>
        <div className="flex items-center space-x-2">
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
          <Button onClick={() => setShowCreateDialog(true)}>
            <Plus className="mr-2 h-4 w-4" />
            Send Notification
          </Button>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total</CardTitle>
            <Bell className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.total ?? 0}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Unread</CardTitle>
            <Mail className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.unread ?? 0}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Read</CardTitle>
            <MailOpen className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.read ?? 0}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Today</CardTitle>
            <Send className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats?.today ?? 0}</div>
          </CardContent>
        </Card>
      </div>

      {/* Filters */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex items-center gap-4">
            <div className="flex-1">
              <div className="relative">
                <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Search notifications..."
                  className="pl-8"
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                />
              </div>
            </div>
            <Select
              value={filters.status || "all"}
              onValueChange={(v) =>
                setFilters((prev) => ({
                  ...prev,
                  status: v === "all" ? undefined : v,
                }))
              }
            >
              <SelectTrigger className="w-[150px]">
                <SelectValue placeholder="Status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All</SelectItem>
                <SelectItem value="unread">Unread</SelectItem>
                <SelectItem value="read">Read</SelectItem>
              </SelectContent>
            </Select>
            <Select
              value={filters.type || "all"}
              onValueChange={(v) =>
                setFilters((prev) => ({
                  ...prev,
                  type: v === "all" ? undefined : v,
                }))
              }
            >
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Type" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All Types</SelectItem>
                <SelectItem value="approval_required">Approval</SelectItem>
                <SelectItem value="document_approved">Approved</SelectItem>
                <SelectItem value="document_rejected">Rejected</SelectItem>
                <SelectItem value="admin_announcement">Announcement</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      {/* Bulk Actions */}
      {selectedIds.length > 0 && (
        <Card>
          <CardContent className="pt-4 pb-4">
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">
                {selectedIds.length} notification(s) selected
              </span>
              <Button
                variant="destructive"
                size="sm"
                onClick={handleBulkDelete}
                disabled={bulkDeleteMutation.isPending}
              >
                <Trash2 className="mr-2 h-4 w-4" />
                Delete Selected
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Notifications Table */}
      <Card>
        <CardHeader>
          <CardTitle>All Notifications</CardTitle>
          <CardDescription>
            Platform-wide notification history and management
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
                  <Skeleton className="h-4 w-4" />
                  <div className="flex-1 space-y-2">
                    <Skeleton className="h-4 w-1/4" />
                    <Skeleton className="h-3 w-1/2" />
                  </div>
                  <Skeleton className="h-6 w-16" />
                </div>
              ))}
            </div>
          ) : (notifications as Array<any>).length === 0 ? (
            <div className="text-center py-8">
              <Bell className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <p className="text-muted-foreground">No notifications found</p>
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-12">
                    <input
                      type="checkbox"
                      checked={
                        selectedIds.length ===
                        (notifications as Array<any>).length
                      }
                      onChange={(e) => handleSelectAll(e.target.checked)}
                      className="rounded border-gray-300"
                    />
                  </TableHead>
                  <TableHead>Subject</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Importance</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Date</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {(notifications as Array<any>).map((notification: any) => (
                  <TableRow key={notification.id}>
                    <TableCell>
                      <input
                        type="checkbox"
                        checked={selectedIds.includes(notification.id)}
                        onChange={() => handleSelectToggle(notification.id)}
                        className="rounded border-gray-300"
                      />
                    </TableCell>
                    <TableCell>
                      <div>
                        <div
                          className={`font-medium ${!notification.is_read ? "font-bold" : ""}`}
                        >
                          {notification.subject || "No subject"}
                        </div>
                        <div className="text-sm text-muted-foreground truncate max-w-[300px]">
                          {notification.body}
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>{getTypeBadge(notification.type)}</TableCell>
                    <TableCell>
                      {getImportanceBadge(notification.importance)}
                    </TableCell>
                    <TableCell>
                      {notification.is_read ? (
                        <Badge variant="secondary">
                          <MailOpen className="mr-1 h-3 w-3" />
                          Read
                        </Badge>
                      ) : (
                        <Badge variant="default">
                          <Mail className="mr-1 h-3 w-3" />
                          Unread
                        </Badge>
                      )}
                    </TableCell>
                    <TableCell>
                      <span className="text-sm text-muted-foreground">
                        {new Date(notification.created_at).toLocaleDateString()}
                      </span>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-1">
                        {!notification.is_read && (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleMarkRead(notification.id)}
                            title="Mark as read"
                          >
                            <CheckCircle className="h-4 w-4" />
                          </Button>
                        )}
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handleDelete(notification.id)}
                          title="Delete"
                          className="text-red-600 hover:text-red-700"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      {/* Create Notification Dialog */}
      <CreateNotificationDialog
        open={showCreateDialog}
        onOpenChange={setShowCreateDialog}
        onSubmit={handleCreate}
        isPending={createMutation.isPending}
      />
    </div>
  );
}

function CreateNotificationDialog({
  open,
  onOpenChange,
  onSubmit,
  isPending,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (req: CreateNotificationRequest) => void;
  isPending: boolean;
}) {
  const [subject, setSubject] = useState("");
  const [body, setBody] = useState("");
  const [type, setType] = useState("admin_announcement");
  const [importance, setImportance] = useState("MEDIUM");
  const [targetType, setTargetType] = useState("broadcast");

  const handleSubmit = () => {
    if (!subject.trim() || !body.trim()) {
      toast.error("Subject and body are required");
      return;
    }

    onSubmit({
      subject: subject.trim(),
      body: body.trim(),
      type,
      importance,
      broadcast: targetType === "broadcast",
    });

    // Reset form
    setSubject("");
    setBody("");
    setType("admin_announcement");
    setImportance("MEDIUM");
    setTargetType("broadcast");
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[525px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Megaphone className="h-5 w-5" />
            Send Notification
          </DialogTitle>
          <DialogDescription>
            Send a notification to platform users
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="subject">Subject</Label>
            <Input
              id="subject"
              value={subject}
              onChange={(e) => setSubject(e.target.value)}
              placeholder="Notification subject..."
            />
          </div>
          <div className="grid gap-2">
            <Label htmlFor="body">Message</Label>
            <Textarea
              id="body"
              value={body}
              onChange={(e) => setBody(e.target.value)}
              placeholder="Notification message..."
              rows={4}
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="grid gap-2">
              <Label>Type</Label>
              <Select value={type} onValueChange={setType}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="admin_announcement">
                    Announcement
                  </SelectItem>
                  <SelectItem value="status_change">Status Change</SelectItem>
                  <SelectItem value="approval_required">
                    Approval Required
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="grid gap-2">
              <Label>Importance</Label>
              <Select value={importance} onValueChange={setImportance}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="LOW">Low</SelectItem>
                  <SelectItem value="MEDIUM">Medium</SelectItem>
                  <SelectItem value="HIGH">High</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className="grid gap-2">
            <Label>Recipients</Label>
            <Select value={targetType} onValueChange={setTargetType}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="broadcast">
                  All Users (Broadcast)
                </SelectItem>
              </SelectContent>
            </Select>
            <p className="text-xs text-muted-foreground">
              Organization-specific and individual targeting will be available in
              a future update.
            </p>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={isPending}>
            {isPending ? "Sending..." : "Send Notification"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
