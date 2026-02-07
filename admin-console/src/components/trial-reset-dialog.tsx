"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Clock, RefreshCw, AlertTriangle } from "lucide-react";
import { resetOrganizationTrial } from "@/app/_actions/organizations";
import { toast } from "sonner";

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
  const [isLoading, setIsLoading] = useState(false);
  const [trialDays, setTrialDays] = useState(30);
  const [reason, setReason] = useState("");
  const [error, setError] = useState("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!reason.trim() || reason.length < 5) {
      setError("Reason must be at least 5 characters long");
      return;
    }

    if (trialDays < 1 || trialDays > 90) {
      setError("Trial days must be between 1 and 90");
      return;
    }

    setIsLoading(true);

    try {
      const result = await resetOrganizationTrial(organization.id, {
        trialDays,
        reason: reason.trim(),
      });

      if (result.success) {
        toast.success("Trial reset successfully");
        setIsOpen(false);
        setReason("");
        setTrialDays(30);
        onSuccess();
      } else {
        setError(result.message || "Failed to reset trial");
      }
    } catch (error) {
      setError(
        error instanceof Error ? error.message : "Failed to reset trial",
      );
    } finally {
      setIsLoading(false);
    }
  };

  if (!isOpen) {
    return (
      <Button variant="outline" size="sm" onClick={() => setIsOpen(true)}>
        <Clock className="mr-2 h-4 w-4" />
        Reset Trial
      </Button>
    );
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <Card className="w-full max-w-md mx-4">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <RefreshCw className="h-5 w-5" />
            Reset Trial Period
          </CardTitle>
          <CardDescription>
            Reset the trial period for {organization.name}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
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

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label htmlFor="trialDays" className="text-sm font-medium">
                New Trial Duration (days)
              </label>
              <Input
                id="trialDays"
                type="number"
                min="1"
                max="90"
                value={trialDays}
                onChange={(e) => setTrialDays(parseInt(e.target.value) || 30)}
                className="mt-1"
              />
              <p className="text-xs text-muted-foreground mt-1">
                Between 1 and 90 days
              </p>
            </div>

            <div>
              <label htmlFor="reason" className="text-sm font-medium">
                Reason for Reset *
              </label>
              <Textarea
                id="reason"
                value={reason}
                onChange={(e) => setReason(e.target.value)}
                placeholder="Explain why you're resetting this trial period..."
                className="mt-1"
                rows={3}
              />
              <p className="text-xs text-muted-foreground mt-1">
                Minimum 5 characters required
              </p>
            </div>

            {error && (
              <div className="flex items-center gap-2 text-sm text-red-600 bg-red-50 p-2 rounded">
                <AlertTriangle className="h-4 w-4" />
                {error}
              </div>
            )}

            <div className="flex gap-2 pt-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  setIsOpen(false);
                  setError("");
                  setReason("");
                  setTrialDays(30);
                }}
                disabled={isLoading}
                className="flex-1"
              >
                Cancel
              </Button>
              <Button
                type="submit"
                disabled={isLoading || !reason.trim()}
                className="flex-1"
              >
                {isLoading ? (
                  <>
                    <RefreshCw className="mr-2 h-4 w-4 animate-spin" />
                    Resetting...
                  </>
                ) : (
                  <>
                    <Clock className="mr-2 h-4 w-4" />
                    Reset Trial
                  </>
                )}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
