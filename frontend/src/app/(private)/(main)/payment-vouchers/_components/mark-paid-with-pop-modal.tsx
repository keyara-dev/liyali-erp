"use client";

import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";
import { markPaidWithPOP } from "@/app/_actions/payment-vouchers";
import { QUERY_KEYS } from "@/lib/constants";

interface MarkPaidWithPOPModalProps {
  pvId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

const ALLOWED_MIME_TYPES = ["application/pdf", "image/jpeg", "image/png"];
const MAX_FILE_SIZE = 10 * 1024 * 1024; // 10MB

export function MarkPaidWithPOPModal({
  pvId,
  open,
  onOpenChange,
  onSuccess,
}: MarkPaidWithPOPModalProps) {
  const [file, setFile] = useState<File | null>(null);
  const [paidDate, setPaidDate] = useState<string>(
    new Date().toISOString().slice(0, 10),
  );
  const [notes, setNotes] = useState("");
  const qc = useQueryClient();

  const mutation = useMutation({
    mutationFn: async () => {
      if (!file) throw new Error("Proof of payment file is required");
      if (file.size > MAX_FILE_SIZE)
        throw new Error("File exceeds 10MB limit");
      if (!ALLOWED_MIME_TYPES.includes(file.type))
        throw new Error("File must be PDF, JPG, or PNG");
      if (!paidDate) throw new Error("Paid date is required");

      const fd = new FormData();
      fd.append("popFile", file);
      fd.append("paidDate", paidDate);
      if (notes.trim()) fd.append("notes", notes.trim());

      const result = await markPaidWithPOP(pvId, fd);
      if (!result.success) {
        throw new Error(result.message || "Failed to mark as paid");
      }
      return result;
    },
    onSuccess: () => {
      toast.success("Payment voucher marked as paid");
      qc.invalidateQueries({ queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.ALL] });
      qc.invalidateQueries({
        queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.BY_ID, pvId],
      });
      qc.invalidateQueries({ queryKey: [QUERY_KEYS.PAYMENT_VOUCHERS.STATS] });
      qc.invalidateQueries({ queryKey: [QUERY_KEYS.DASHBOARD.METRICS] });
      // Reset form
      setFile(null);
      setPaidDate(new Date().toISOString().slice(0, 10));
      setNotes("");
      onOpenChange(false);
      onSuccess?.();
    },
    onError: (e: Error) => {
      toast.error(e.message || "Failed to mark as paid");
    },
  });

  const canSubmit = !!file && !!paidDate && !mutation.isPending;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Mark as Paid</DialogTitle>
          <DialogDescription asChild>
            <div>Upload proof of payment to complete this voucher.</div>
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-2">
          <div className="space-y-2">
            <Label htmlFor="popFile">
              Proof of payment file <span className="text-destructive">*</span>
            </Label>
            <Input
              id="popFile"
              type="file"
              accept=".pdf,.jpg,.jpeg,.png"
              onChange={(e) => setFile(e.target.files?.[0] ?? null)}
            />
            <p className="text-xs text-muted-foreground">
              PDF, JPG or PNG, max 10MB.
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="paidDate">
              Paid date <span className="text-destructive">*</span>
            </Label>
            <Input
              id="paidDate"
              type="date"
              value={paidDate}
              onChange={(e) => setPaidDate(e.target.value)}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="popNotes">Notes</Label>
            <Textarea
              id="popNotes"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder="Optional additional notes about this payment"
              rows={3}
            />
          </div>
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={mutation.isPending}
          >
            Cancel
          </Button>
          <Button
            onClick={() => mutation.mutate()}
            disabled={!canSubmit}
            className="bg-emerald-600 hover:bg-emerald-700"
          >
            {mutation.isPending ? "Submitting…" : "Confirm paid"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
