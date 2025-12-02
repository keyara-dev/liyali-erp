"use client";

import { useState, useMemo } from "react";
import { ApprovalTask } from "@/types";
import { useGetUsers } from "@/hooks/use-users-query";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import {
  Loader2,
  AlertCircle,
  User,
  Mail,
  Building2,
} from "lucide-react";

export interface ReassignmentModalProps {
  task: ApprovalTask;
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  onReassign: (userId: string, reason: string) => Promise<void>;
}

export function ReassignmentModal({
  task,
  isOpen,
  onOpenChange,
  onReassign,
}: ReassignmentModalProps) {
  const [selectedUserId, setSelectedUserId] = useState<string>("");
  const [reason, setReason] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");

  const { data: usersData } = useGetUsers();

  // Filter users - exclude current approver and get available users
  const availableUsers = useMemo(() => {
    if (!usersData) return [];
    return usersData.filter(
      (u: any) => u.id !== task.approverUserId
    );
  }, [usersData, task.approverUserId]);

  // Filter by search query
  const filteredUsers = useMemo(() => {
    if (!searchQuery) return availableUsers;
    const query = searchQuery.toLowerCase();
    return availableUsers.filter(
      (u) =>
        u.name?.toLowerCase().includes(query) ||
        u.email?.toLowerCase().includes(query)
    );
  }, [availableUsers, searchQuery]);

  const selectedUser = availableUsers.find((u) => u.id === selectedUserId);

  const handleReassign = async () => {
    if (!selectedUserId) {
      setError("Please select a user to reassign to");
      return;
    }

    if (!reason.trim()) {
      setError("Please provide a reason for reassignment");
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      await onReassign(selectedUserId, reason);
      // Reset form
      setSelectedUserId("");
      setReason("");
      setSearchQuery("");
      onOpenChange(false);
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Failed to reassign task"
      );
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>Reassign Approval Task</DialogTitle>
          <DialogDescription>
            Transfer this task to another user. The current assignee will be
            notified.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          {/* Current Task Info */}
          <div className="p-3 bg-muted rounded-lg text-sm">
            <div className="grid grid-cols-2 gap-2">
              <div>
                <span className="text-muted-foreground font-medium">Entity:</span>
                <p className="font-mono">
                  {task.entityType} #{task.entityNumber}
                </p>
              </div>
              <div>
                <span className="text-muted-foreground font-medium">Current Assignee:</span>
                <p>{task.approverName || "Unknown"}</p>
              </div>
            </div>
          </div>

          {/* Error Alert */}
          {error && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {/* User Selection */}
          <div className="space-y-2">
            <Label>Select New Approver</Label>
            <Input
              placeholder="Search by name or email..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              disabled={isLoading}
            />

            <Select
              value={selectedUserId}
              onValueChange={setSelectedUserId}
              disabled={isLoading}
            >
              <SelectTrigger>
                <SelectValue placeholder="Choose approver..." />
              </SelectTrigger>
              <SelectContent>
                {filteredUsers.length === 0 ? (
                  <div className="p-4 text-center text-sm text-muted-foreground">
                    No users available
                  </div>
                ) : (
                  filteredUsers.map((user) => (
                    <SelectItem key={user.id} value={user.id}>
                      <div className="flex items-center gap-2">
                        <Avatar className="h-5 w-5">
                          <AvatarImage src={user.avatar} />
                          <AvatarFallback>
                            {user.name?.charAt(0).toUpperCase()}
                          </AvatarFallback>
                        </Avatar>
                        <span>{user.name || user.email}</span>
                      </div>
                    </SelectItem>
                  ))
                )}
              </SelectContent>
            </Select>
          </div>

          {/* Selected User Details */}
          {selectedUser && (
            <div className="p-3 bg-primary/5 border border-primary/20 rounded-lg space-y-2">
              <div className="flex items-center gap-2">
                <Avatar className="h-8 w-8">
                  <AvatarImage src={selectedUser.avatar} />
                  <AvatarFallback>
                    {selectedUser.name?.charAt(0).toUpperCase()}
                  </AvatarFallback>
                </Avatar>
                <div className="flex-1 min-w-0">
                  <h3 className="font-semibold text-sm">{selectedUser.name}</h3>
                  <p className="text-xs text-muted-foreground">{selectedUser.email}</p>
                </div>
              </div>

              {selectedUser.role && (
                <div className="flex items-center gap-2 text-xs">
                  <Badge variant="outline">{selectedUser.role}</Badge>
                </div>
              )}
            </div>
          )}

          {/* Reason */}
          <div className="space-y-2">
            <Label htmlFor="reason">Reason for Reassignment</Label>
            <Textarea
              id="reason"
              placeholder="Explain why you're reassigning this task..."
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              disabled={isLoading}
              className="min-h-24 resize-none"
            />
            <p className="text-xs text-muted-foreground">
              {reason.length}/500 characters
            </p>
          </div>

          {/* Info Alert */}
          <Alert>
            <AlertCircle className="h-4 w-4" />
            <AlertDescription className="text-xs">
              The new approver will be notified immediately. The reason will be
              visible in the approval history.
            </AlertDescription>
          </Alert>
        </div>

        <DialogFooter className="flex gap-2 sm:flex-row sm:justify-end">
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isLoading}
          >
            Cancel
          </Button>
          <Button
            onClick={handleReassign}
            disabled={isLoading || !selectedUserId || !reason.trim()}
          >
            {isLoading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Reassigning...
              </>
            ) : (
              "Confirm Reassignment"
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
