"use client";

import { useState, useEffect } from "react";
import { NotificationTypeEnum as NotificationType } from "@/types";
import {
  useGetNotificationPreferences,
  useUpdateNotificationPreferences,
} from "@/hooks/use-notifications";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { Loader2, CheckCircle2 } from "lucide-react";

const notificationTypeLabels: Record<NotificationType, string> = {
  TASK_ASSIGNED: "Task Assigned",
  TASK_REASSIGNED: "Task Reassigned",
  TASK_APPROVED: "Task Approved",
  TASK_REJECTED: "Task Rejected",
  WORKFLOW_COMPLETE: "Workflow Complete",
  APPROVAL_OVERDUE: "Approval Overdue",
  COMMENT_ADDED: "Comment Added",
};

const notificationTypeDescriptions: Record<NotificationType, string> = {
  TASK_ASSIGNED:
    "Receive notifications when a new approval task is assigned to you",
  TASK_REASSIGNED:
    "Receive notifications when an approval task is reassigned to you",
  TASK_APPROVED: "Receive notifications when your submission is approved",
  TASK_REJECTED: "Receive notifications when your submission is rejected",
  WORKFLOW_COMPLETE:
    "Receive notifications when your workflow is fully completed",
  APPROVAL_OVERDUE: "Receive notifications for overdue approvals",
  COMMENT_ADDED: "Receive notifications when comments are added to your items",
};

export interface NotificationPreferencesProps {
  userId: string;
  onSaved?: () => void;
}

export function NotificationPreferences({
  userId,
  onSaved,
}: NotificationPreferencesProps) {
  const { data: preferences, isLoading } = useGetNotificationPreferences({
    userId,
  });
  const updateMutation = useUpdateNotificationPreferences();

  const [savedMessage, setSavedMessage] = useState(false);
  const [localPreferences, setLocalPreferences] = useState<
    Record<NotificationType, boolean>
  >({
    TASK_ASSIGNED: true,
    TASK_REASSIGNED: true,
    TASK_APPROVED: true,
    TASK_REJECTED: true,
    WORKFLOW_COMPLETE: true,
    APPROVAL_OVERDUE: true,
    COMMENT_ADDED: true,
  });

  useEffect(() => {
    if (preferences?.data) {
      const prefs = preferences.data;
      setLocalPreferences({
        TASK_ASSIGNED: prefs.notifyOn?.taskAssigned ?? true,
        TASK_REASSIGNED: prefs.notifyOn?.taskReassigned ?? true,
        TASK_APPROVED: prefs.notifyOn?.taskApproved ?? true,
        TASK_REJECTED: prefs.notifyOn?.taskRejected ?? true,
        WORKFLOW_COMPLETE: prefs.notifyOn?.workflowComplete ?? true,
        APPROVAL_OVERDUE: prefs.notifyOn?.approvalOverdue ?? true,
        COMMENT_ADDED: prefs.notifyOn?.commentsAdded ?? true,
      });
    }
  }, [preferences]);

  const handleToggle = (type: NotificationType) => {
    setLocalPreferences((prev) => ({
      ...prev,
      [type]: !prev[type],
    }));
    setSavedMessage(false);
  };

  const handleSave = async () => {
    try {
      await updateMutation.mutateAsync({
        userId,
        preferences: {
          notifyOn: {
            taskAssigned: localPreferences.TASK_ASSIGNED,
            taskReassigned: localPreferences.TASK_REASSIGNED,
            taskApproved: localPreferences.TASK_APPROVED,
            taskRejected: localPreferences.TASK_REJECTED,
            workflowComplete: localPreferences.WORKFLOW_COMPLETE,
            approvalOverdue: localPreferences.APPROVAL_OVERDUE,
            commentsAdded: localPreferences.COMMENT_ADDED,
          },
        },
      });
      setSavedMessage(true);
      onSaved?.();
      setTimeout(() => setSavedMessage(false), 3000);
    } catch (error) {
      console.error("Failed to update preferences:", error);
    }
  };

  const hasChanges =
    JSON.stringify(preferences?.data?.notifyOn) !==
    JSON.stringify(localPreferences);

  if (isLoading) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-12">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Notification Preferences</CardTitle>
        <CardDescription>
          Customize which notifications you want to receive
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        <div className="space-y-4">
          {Object.entries(notificationTypeLabels).map(([type, label]) => (
            <div key={type} className="flex items-start justify-between gap-4">
              <div className="space-y-1 flex-1">
                <Label className="text-base font-medium">{label}</Label>
                <p className="text-sm text-muted-foreground">
                  {notificationTypeDescriptions[type as NotificationType]}
                </p>
              </div>
              <div className="flex items-center gap-3">
                <Switch
                  checked={localPreferences[type as NotificationType] ?? true}
                  onCheckedChange={() => handleToggle(type as NotificationType)}
                  disabled={updateMutation.isPending}
                />
              </div>
            </div>
          ))}
        </div>

        <div className="border-t pt-6 flex gap-3 items-center justify-between">
          <Button
            onClick={handleSave}
            disabled={!hasChanges || updateMutation.isPending}
            isLoading={updateMutation.isPending}
            loadingText="Saving..."
          >
            Save Preferences
          </Button>

          {savedMessage && (
            <div className="flex items-center gap-2 text-sm text-green-600 dark:text-green-400">
              <CheckCircle2 className="h-4 w-4" />
              Preferences saved successfully
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
