"use client";

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { InfoHint } from "@/components/ui/info-hint";
import { Clock, Users } from "lucide-react";
import { addMinutes } from "date-fns";

interface ClaimTaskModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  isLoading: boolean;
  taskDetails: {
    entityType: string;
    entityId: string;
    stageName: string;
    assignedRole: string;
  };
}

export function ClaimTaskModal({
  isOpen,
  onClose,
  onConfirm,
  isLoading,
  taskDetails,
}: ClaimTaskModalProps) {
  const claimDuration = 30; // 30 minutes
  const expiryTime = addMinutes(new Date(), claimDuration);

  const claimNotes = (
    <div className="space-y-2">
      <div>
        <p className="font-medium text-foreground">While claimed</p>
        <ul className="mt-1 space-y-1">
          <li>• Only you can approve or reject this task</li>
          <li>• Others see that you&apos;re reviewing it</li>
          <li>• The claim expires after {claimDuration} min of inactivity</li>
          <li>• You can unclaim it anytime to step away</li>
        </ul>
      </div>
      <div>
        <p className="font-medium text-foreground">Next steps</p>
        <p className="mt-1">
          After claiming you&apos;ll see Approve and Reject buttons. Each needs
          comments and your digital signature.
        </p>
      </div>
    </div>
  );

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-md max-h-[90svh] flex flex-col p-0 overflow-hidden">
        <DialogHeader className="shrink-0 px-6 pt-6 pb-4 border-b">
          <DialogTitle className="flex items-center gap-2">
            <Users className="h-5 w-5 text-blue-600" />
            Claim Task for Review
          </DialogTitle>
          <DialogDescription asChild>
            <div className="text-left space-y-3 text-muted-foreground text-sm">
              <div className="bg-blue-50 dark:bg-blue-950/30 p-3 rounded-lg border border-blue-200 dark:border-blue-800">
                <p className="font-medium text-blue-900 dark:text-blue-100">
                  {taskDetails.entityType} #{taskDetails.entityId}
                </p>
                <p className="text-sm text-blue-700 dark:text-blue-300">
                  Stage: {taskDetails.stageName}
                </p>
                <p className="text-sm text-blue-700 dark:text-blue-300">
                  Required role: {taskDetails.assignedRole}
                </p>
              </div>

              <div className="flex items-center gap-2 rounded-lg border bg-muted/40 p-3">
                <Clock className="h-4 w-4 text-amber-600 shrink-0" />
                <span className="text-sm text-foreground">
                  <strong>{claimDuration} min</strong> to act · expires{" "}
                  {expiryTime.toLocaleTimeString()}
                </span>
                {/* Mobile-only; full notes shown inline on md+ screens */}
                <InfoHint
                  label="What claiming a task means"
                  side="top"
                  align="end"
                  triggerLabel="Details"
                  className="ml-auto md:hidden"
                >
                  <div className="text-muted-foreground">{claimNotes}</div>
                </InfoHint>
              </div>

              {/* Desktop: the same notes, always visible */}
              <div className="hidden md:block rounded-lg border bg-muted/30 p-3 text-muted-foreground">
                {claimNotes}
              </div>
            </div>
          </DialogDescription>
        </DialogHeader>

        <div className="shrink-0 border-t bg-background px-6 py-4 flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
          <Button
            variant="outline"
            onClick={onClose}
            disabled={isLoading}
            className="w-full sm:w-auto"
          >
            Cancel
          </Button>
          <Button
            onClick={onConfirm}
            disabled={isLoading}
            className="w-full sm:w-auto bg-blue-600 hover:bg-blue-700"
            isLoading={isLoading}
            loadingText="Claiming Task..."
          >
            <Users className="h-4 w-4 mr-2" />
            Claim Task ({claimDuration} min)
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
