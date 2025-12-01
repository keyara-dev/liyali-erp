"use client";

import { useState, useRef } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Loader2, AlertCircle } from "lucide-react";
import { Notification } from "@/types";

interface SignatureCanvasProps {
  onSignatureChange: (signature: string) => void;
  isRequired?: boolean;
}

const SignatureCanvas = ({
  onSignatureChange,
  isRequired = true,
}: SignatureCanvasProps) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [isDrawing, setIsDrawing] = useState(false);
  const [hasSignature, setHasSignature] = useState(false);

  const startDrawing = (e: React.MouseEvent<HTMLCanvasElement>) => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const rect = canvas.getBoundingClientRect();
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    ctx.beginPath();
    ctx.moveTo(e.clientX - rect.left, e.clientY - rect.top);
    setIsDrawing(true);
  };

  const draw = (e: React.MouseEvent<HTMLCanvasElement>) => {
    if (!isDrawing) return;

    const canvas = canvasRef.current;
    if (!canvas) return;

    const rect = canvas.getBoundingClientRect();
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    ctx.lineTo(e.clientX - rect.left, e.clientY - rect.top);
    ctx.stroke();
  };

  const stopDrawing = () => {
    if (!isDrawing) return;

    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext("2d");
    if (ctx) {
      ctx.closePath();
    }

    setIsDrawing(false);
    setHasSignature(true);

    // Convert canvas to base64
    const signature = canvas.toDataURL("image/png");
    onSignatureChange(signature);
  };

  const clearSignature = () => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext("2d");
    if (ctx) {
      ctx.clearRect(0, 0, canvas.width, canvas.height);
    }

    setHasSignature(false);
    onSignatureChange("");
  };

  return (
    <div className="space-y-3">
      <Label>
        Digital Signature {isRequired && <span className="text-destructive">*</span>}
      </Label>
      <div className="border rounded-lg bg-white dark:bg-slate-900 overflow-hidden">
        <canvas
          ref={canvasRef}
          width={400}
          height={150}
          onMouseDown={startDrawing}
          onMouseMove={draw}
          onMouseUp={stopDrawing}
          onMouseLeave={stopDrawing}
          className="w-full cursor-crosshair bg-white dark:bg-slate-900"
        />
      </div>
      <div className="flex gap-2">
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={clearSignature}
        >
          Clear
        </Button>
        <span className="text-xs text-muted-foreground self-center">
          {hasSignature ? "✓ Signature captured" : "Draw your signature above"}
        </span>
      </div>
    </div>
  );
};

interface NotificationActionModalProps {
  notification: Notification | null;
  isOpen: boolean;
  onOpenChange: (open: boolean) => void;
  onApprove?: (
    signature: string,
    remarks: string
  ) => Promise<void>;
  onReject?: (remarks: string) => Promise<void>;
  actionType?: "approve" | "reject" | "both";
}

export function NotificationActionModal({
  notification,
  isOpen,
  onOpenChange,
  onApprove,
  onReject,
  actionType = "both",
}: NotificationActionModalProps) {
  const [remarks, setRemarks] = useState("");
  const [signature, setSignature] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [action, setAction] = useState<"approve" | "reject" | null>(null);

  if (!notification) return null;

  const handleApprove = async () => {
    setError(null);

    if (!signature) {
      setError("Signature is required");
      return;
    }

    if (!onApprove) {
      setError("Approve action not available");
      return;
    }

    setIsSubmitting(true);
    try {
      await onApprove(signature, remarks);
      setRemarks("");
      setSignature("");
      setAction(null);
      onOpenChange(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to approve");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleReject = async () => {
    setError(null);

    if (!remarks.trim()) {
      setError("Rejection reason is required");
      return;
    }

    if (!onReject) {
      setError("Reject action not available");
      return;
    }

    setIsSubmitting(true);
    try {
      await onReject(remarks);
      setRemarks("");
      setSignature("");
      setAction(null);
      onOpenChange(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to reject");
    } finally {
      setIsSubmitting(false);
    }
  };

  const getTitle = () => {
    if (action === "approve") return "Approve Submission";
    if (action === "reject") return "Reject Submission";
    return `Review ${notification.entityType} #${notification.entityNumber}`;
  };

  return (
    <Dialog open={isOpen} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>{getTitle()}</DialogTitle>
          <DialogDescription>
            {notification.message}
          </DialogDescription>
        </DialogHeader>

        {!action ? (
          // Preview Mode
          <div className="space-y-4 py-4">
            <div className="rounded-lg border bg-muted/50 p-4">
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <div className="font-semibold text-muted-foreground">Type</div>
                  <div>{notification.entityType}</div>
                </div>
                <div>
                  <div className="font-semibold text-muted-foreground">
                    Number
                  </div>
                  <div className="font-mono">{notification.entityNumber}</div>
                </div>
                <div>
                  <div className="font-semibold text-muted-foreground">
                    Created
                  </div>
                  <div>
                    {new Date(notification.createdAt).toLocaleDateString()}
                  </div>
                </div>
                <div>
                  <div className="font-semibold text-muted-foreground">
                    Status
                  </div>
                  <div>{notification.isRead ? "Read" : "Unread"}</div>
                </div>
              </div>
            </div>

            <div className="space-y-2">
              <p className="text-sm text-muted-foreground">
                Choose your action:
              </p>
              <div className="flex gap-2">
                {(actionType === "approve" || actionType === "both") && (
                  <Button
                    onClick={() => setAction("approve")}
                    className="flex-1"
                  >
                    ✓ Approve
                  </Button>
                )}
                {(actionType === "reject" || actionType === "both") && (
                  <Button
                    variant="destructive"
                    onClick={() => setAction("reject")}
                    className="flex-1"
                  >
                    ✕ Reject
                  </Button>
                )}
              </div>
            </div>
          </div>
        ) : (
          // Action Mode
          <div className="space-y-4 py-4">
            {error && (
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}

            {action === "approve" && (
              <>
                <SignatureCanvas
                  onSignatureChange={setSignature}
                  isRequired={true}
                />

                <div className="space-y-2">
                  <Label htmlFor="remarks">
                    Remarks (optional)
                  </Label>
                  <Textarea
                    id="remarks"
                    placeholder="Add any comments about this approval..."
                    value={remarks}
                    onChange={(e) => setRemarks(e.target.value)}
                    className="min-h-20"
                  />
                </div>
              </>
            )}

            {action === "reject" && (
              <div className="space-y-2">
                <Label htmlFor="rejection-reason">
                  Rejection Reason <span className="text-destructive">*</span>
                </Label>
                <Textarea
                  id="rejection-reason"
                  placeholder="Explain why you are rejecting this submission..."
                  value={remarks}
                  onChange={(e) => setRemarks(e.target.value)}
                  className="min-h-24"
                />
              </div>
            )}

            <div className="flex gap-2 pt-4">
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  setAction(null);
                  setError(null);
                }}
                disabled={isSubmitting}
              >
                Back
              </Button>
              {action === "approve" ? (
                <Button
                  onClick={handleApprove}
                  disabled={isSubmitting || !signature}
                  className="flex-1"
                >
                  {isSubmitting && (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  )}
                  Confirm Approval
                </Button>
              ) : (
                <Button
                  variant="destructive"
                  onClick={handleReject}
                  disabled={isSubmitting || !remarks.trim()}
                  className="flex-1"
                >
                  {isSubmitting && (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  )}
                  Confirm Rejection
                </Button>
              )}
            </div>
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
