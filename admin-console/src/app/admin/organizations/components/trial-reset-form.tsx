"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { AlertTriangle, RefreshCw, Clock } from "lucide-react";
import { toast } from "sonner";
import { resetOrganizationTrial } from "@/app/_actions/subscriptions";

interface TrialResetFormProps {
  organizationId: string;
  onSuccess: () => void;
  /** If true, wraps submit button full-width (tab layout). */
  fullWidthSubmit?: boolean;
}

export function TrialResetForm({
  organizationId,
  onSuccess,
  fullWidthSubmit = false,
}: TrialResetFormProps) {
  const [trialDays, setTrialDays] = useState(30);
  const [reason, setReason] = useState("");
  const [isLoading, setIsLoading] = useState(false);
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
      const result = await resetOrganizationTrial(organizationId, {
        trial_days: trialDays,
        reason: reason.trim(),
      });

      if (result.success) {
        toast.success("Trial reset successfully");
        setReason("");
        setTrialDays(30);
        onSuccess();
      } else {
        setError(result.message || "Failed to reset trial");
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to reset trial");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <label htmlFor="trialDays" className="text-sm font-medium">
          New Trial Duration (days)
        </label>
        <Input
          id="trialDays"
          type="number"
          min="1"
          max="90"
          value={trialDays}
          onChange={(e) => {
            const v = parseInt(e.target.value);
            if (!isNaN(v)) setTrialDays(v);
          }}
        />
        <p className="text-xs text-muted-foreground">Between 1 and 90 days</p>
      </div>

      <div className="space-y-2">
        <label htmlFor="resetReason" className="text-sm font-medium">
          Reason for Reset <span className="text-red-500">*</span>
        </label>
        <Textarea
          id="resetReason"
          value={reason}
          onChange={(e) => setReason(e.target.value)}
          placeholder="Explain why you're resetting this trial period... (min 5 characters)"
          rows={3}
        />
      </div>

      {error && (
        <div className="flex items-center gap-2 text-sm text-red-600 bg-red-50 dark:bg-red-950/20 p-2 rounded">
          <AlertTriangle className="h-4 w-4 flex-shrink-0" />
          {error}
        </div>
      )}

      <Button
        type="submit"
        disabled={isLoading || !reason.trim()}
        isLoading={isLoading}
        loadingText="Resetting..."
        className={fullWidthSubmit ? "w-full" : undefined}
      >
        <Clock className="mr-2 h-4 w-4" />
        Reset Trial ({trialDays} days)
      </Button>
    </form>
  );
}
