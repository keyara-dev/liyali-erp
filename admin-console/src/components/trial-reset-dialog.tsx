"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Clock, RefreshCw } from "lucide-react";
import { TrialResetForm } from "@/app/admin/organizations/components/trial-reset-form";

interface TrialResetDialogProps {
  organization: {
    id: string;
    name: string;
    trial_end_date: string;
    days_remaining: number;
    status: string;
  };
  onSuccess: () => void;
}

export function TrialResetDialog({
  organization,
  onSuccess,
}: TrialResetDialogProps) {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <>
      <Button variant="outline" size="sm" onClick={() => setIsOpen(true)}>
        <Clock className="mr-2 h-4 w-4" />
        Reset Trial
      </Button>

      <Dialog open={isOpen} onOpenChange={setIsOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <RefreshCw className="h-5 w-5" />
              Reset Trial Period
            </DialogTitle>
            <DialogDescription>
              Reset the trial period for{" "}
              <strong>{organization.name}</strong>
            </DialogDescription>
          </DialogHeader>

          {/* Organization Info */}
          <div className="rounded-lg border p-3 bg-muted/50">
            <div className="flex items-center justify-between">
              <span className="font-medium">{organization.name}</span>
              <Badge
                variant={
                  organization.days_remaining > 0 ? "success" : "destructive"
                }
              >
                {organization.days_remaining > 0
                  ? `${organization.days_remaining} days left`
                  : `Expired ${Math.abs(organization.days_remaining)} days ago`}
              </Badge>
            </div>
            <p className="text-sm text-muted-foreground mt-1">
              Current trial ends: {organization.trial_end_date}
            </p>
          </div>

          <TrialResetForm
            organizationId={organization.id}
            onSuccess={() => {
              setIsOpen(false);
              onSuccess();
            }}
          />
        </DialogContent>
      </Dialog>
    </>
  );
}
