"use client";

import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components";
import { Badge } from "@/components/ui/badge";
import { DigitalSignaturePad } from "@/components/ui/digital-signature-pad";
import {
  CheckCircle2,
  ShieldCheck,
  PenLine,
  Lock,
  Stamp,
  Upload,
  X,
} from "lucide-react";
import { useOrganizationSettingsQuery } from "@/hooks/use-organization-queries";
import { QUERY_KEYS } from "@/lib/constants";
import {
  signReceiveGRNAction,
  certifyGRNAction,
} from "@/app/_actions/grn-actions";
import type {
  GoodsReceivedNote,
  GRNSignoffStatus,
} from "@/types/goods-received-note";

const CERTIFIER_ROLES = new Set([
  "admin",
  "super_admin",
  "manager",
  "finance",
  "approver",
]);

interface GRNSignoffPanelProps {
  grn: GoodsReceivedNote;
  userId: string;
  userRole: string;
  defaultReceiverName?: string;
}

function formatStamp(value?: string | Date): string {
  if (!value) return "—";
  const d = typeof value === "string" ? new Date(value) : value;
  if (Number.isNaN(d.getTime())) return "—";
  return d.toLocaleString();
}

export function GRNSignoffPanel({
  grn,
  userId,
  userRole,
  defaultReceiverName = "",
}: GRNSignoffPanelProps) {
  const queryClient = useQueryClient();
  const role = userRole.toLowerCase();
  const signoffStatus: GRNSignoffStatus =
    grn.signoffStatus ?? "PENDING_RECEIVER";

  // ── Receiver state ─────────────────────────────────────────────────
  const [receiverName, setReceiverName] = useState(defaultReceiverName);
  const [receiverSignature, setReceiverSignature] = useState("");
  const [receiverSubmitting, setReceiverSubmitting] = useState(false);

  // ── Certifier state ────────────────────────────────────────────────
  const [certifierSignature, setCertifierSignature] = useState("");
  const [certifierComments, setCertifierComments] = useState("");
  const [certifierSubmitting, setCertifierSubmitting] = useState(false);
  const [stampDataUri, setStampDataUri] = useState<string>("");

  // Fallback rubber stamp from org settings — rendered as the default preview
  // when the certifier doesn't upload a per-GRN stamp.
  const { data: orgSettings } = useOrganizationSettingsQuery();
  const orgStampUrl =
    (orgSettings as { stampImageUrl?: string } | undefined)?.stampImageUrl ?? "";
  const effectiveStampPreview = stampDataUri || orgStampUrl;

  const handleStampUpload = (file: File | null) => {
    if (!file) {
      setStampDataUri("");
      return;
    }
    if (!file.type.startsWith("image/")) {
      toast.error("Stamp must be an image file");
      return;
    }
    if (file.size > 2 * 1024 * 1024) {
      toast.error("Stamp image must be under 2 MB");
      return;
    }
    const reader = new FileReader();
    reader.onload = () => {
      if (typeof reader.result === "string") setStampDataUri(reader.result);
    };
    reader.readAsDataURL(file);
  };

  const isReceiverDone = Boolean(grn.receivedAt && grn.receivedBySignature);
  const isCertifierDone = Boolean(grn.certifiedAt && grn.certifiedBySignature);

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.GRN.ALL] });
    queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.GRN.BY_ID, grn.id] });
  };

  const handleReceiverSubmit = async () => {
    if (!receiverName.trim() || !receiverSignature) {
      toast.error("Name and signature are required");
      return;
    }
    setReceiverSubmitting(true);
    try {
      const res = await signReceiveGRNAction({
        grnId: grn.id,
        receivedByName: receiverName.trim(),
        signature: receiverSignature,
      });
      if (res.success) {
        toast.success("Receiver sign-off recorded");
        invalidate();
      } else {
        toast.error(res.message || "Failed to record receiver sign-off");
      }
    } finally {
      setReceiverSubmitting(false);
    }
  };

  const handleCertifierSubmit = async () => {
    if (!certifierSignature) {
      toast.error("Signature is required");
      return;
    }
    if (!CERTIFIER_ROLES.has(role)) {
      toast.error("You do not have permission to certify a GRN");
      return;
    }
    if (grn.createdBy === userId || grn.receivedBy === userId) {
      toast.error("Certifier must be different from the creator and receiver");
      return;
    }
    setCertifierSubmitting(true);
    try {
      const res = await certifyGRNAction({
        grnId: grn.id,
        signature: certifierSignature,
        comments: certifierComments.trim(),
        // Send the uploaded stamp; backend persists per-GRN and PDF prefers
        // it over the org-level fallback.
        stampImageUrl: stampDataUri || undefined,
      });
      if (res.success) {
        toast.success("GRN certified");
        invalidate();
      } else {
        toast.error(res.message || "Failed to certify GRN");
      }
    } finally {
      setCertifierSubmitting(false);
    }
  };

  // ── Render ─────────────────────────────────────────────────────────
  return (
    <Card className="border-border/60">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between gap-2">
          <CardTitle className="text-base flex items-center gap-2">
            <ShieldCheck className="h-4 w-4 text-blue-600" />
            Sign-off
          </CardTitle>
          <SignoffBadge status={signoffStatus} />
        </div>
      </CardHeader>
      <CardContent className="grid gap-6 md:grid-cols-2">
        {/* Receiver block */}
        <SignoffBlock
          title="Received By"
          done={isReceiverDone}
          name={grn.receivedByName}
          signature={grn.receivedBySignature}
          stamp={grn.receivedAt}
        >
          {!isReceiverDone && signoffStatus === "PENDING_RECEIVER" && (
            <div className="space-y-3">
              <Input
                label="Receiver name"
                value={receiverName}
                onChange={(e) => setReceiverName(e.target.value)}
                placeholder="Full name as printed"
                disabled={receiverSubmitting}
              />
              <div>
                <p className="mb-1.5 text-xs font-medium text-muted-foreground uppercase tracking-wider">
                  Signature
                </p>
                <DigitalSignaturePad
                  onSignatureChange={setReceiverSignature}
                  disabled={receiverSubmitting}
                />
              </div>
              <Button
                onClick={handleReceiverSubmit}
                disabled={receiverSubmitting || !receiverSignature}
                isLoading={receiverSubmitting}
                loadingText="Saving…"
                className="w-full gap-1.5"
              >
                <PenLine className="h-4 w-4" />
                Save receiver sign-off
              </Button>
            </div>
          )}
          {!isReceiverDone && signoffStatus !== "PENDING_RECEIVER" && (
            <LockedHint text="Receiver sign-off is locked in this state." />
          )}
        </SignoffBlock>

        {/* Certifier block */}
        <SignoffBlock
          title="Certified By"
          done={isCertifierDone}
          name={grn.certifiedByName}
          signature={grn.certifiedBySignature}
          stamp={grn.certifiedAt}
        >
          {!isCertifierDone &&
            signoffStatus === "PENDING_CERTIFIER" &&
            CERTIFIER_ROLES.has(role) &&
            grn.createdBy !== userId &&
            grn.receivedBy !== userId && (
              <div className="space-y-3">
                <div>
                  <p className="mb-1.5 text-xs font-medium text-muted-foreground uppercase tracking-wider">
                    Signature
                  </p>
                  <DigitalSignaturePad
                    onSignatureChange={setCertifierSignature}
                    disabled={certifierSubmitting}
                  />
                </div>
                <Textarea
                  label="Comments"
                  rows={2}
                  value={certifierComments}
                  onChange={(e) => setCertifierComments(e.target.value)}
                  placeholder="Optional"
                  disabled={certifierSubmitting}
                />

                {/* Issuing-officer stamp (per-GRN; falls back to org stamp) */}
                <div>
                  <div className="mb-1.5 flex items-center justify-between">
                    <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider flex items-center gap-1">
                      <Stamp className="h-3 w-3" /> Issuing Officer Stamp
                    </span>
                    {stampDataUri ? (
                      <button
                        type="button"
                        onClick={() => setStampDataUri("")}
                        className="text-xs text-muted-foreground hover:text-foreground inline-flex items-center gap-1"
                      >
                        <X className="h-3 w-3" />
                        Clear
                      </button>
                    ) : null}
                  </div>
                  <div className="flex items-center gap-3">
                    <div className="h-16 w-24 shrink-0 rounded border border-dashed bg-white flex items-center justify-center overflow-hidden">
                      {effectiveStampPreview ? (
                        // eslint-disable-next-line @next/next/no-img-element
                        <img
                          src={effectiveStampPreview}
                          alt="Stamp preview"
                          className="max-h-full max-w-full object-contain"
                        />
                      ) : (
                        <span className="text-[10px] text-muted-foreground">
                          No stamp
                        </span>
                      )}
                    </div>
                    <div className="flex-1 min-w-0">
                      <label className="inline-flex items-center gap-1.5 cursor-pointer rounded border px-2 py-1 text-xs hover:bg-muted/50">
                        <Upload className="h-3 w-3" />
                        Upload stamp
                        <input
                          type="file"
                          accept="image/*"
                          className="hidden"
                          disabled={certifierSubmitting}
                          onChange={(e) =>
                            handleStampUpload(e.target.files?.[0] ?? null)
                          }
                        />
                      </label>
                      <p className="mt-1 text-[10px] text-muted-foreground">
                        {stampDataUri
                          ? "Per-GRN stamp will be embedded in the PDF."
                          : orgStampUrl
                            ? "Using organization stamp. Upload to override."
                            : "No org stamp configured. Optional upload."}
                      </p>
                    </div>
                  </div>
                </div>

                <Button
                  onClick={handleCertifierSubmit}
                  disabled={certifierSubmitting || !certifierSignature}
                  isLoading={certifierSubmitting}
                  loadingText="Certifying…"
                  className="w-full gap-1.5"
                >
                  <ShieldCheck className="h-4 w-4" />
                  Certify GRN
                </Button>
              </div>
            )}
          {!isCertifierDone && signoffStatus === "PENDING_RECEIVER" && (
            <LockedHint text="Waiting for receiver sign-off first." />
          )}
          {!isCertifierDone &&
            signoffStatus === "PENDING_CERTIFIER" &&
            !CERTIFIER_ROLES.has(role) && (
              <LockedHint text="Only admin, manager, finance or approver roles may certify a GRN." />
            )}
          {!isCertifierDone &&
            signoffStatus === "PENDING_CERTIFIER" &&
            CERTIFIER_ROLES.has(role) &&
            (grn.createdBy === userId || grn.receivedBy === userId) && (
              <LockedHint text="The certifying officer must be different from the creator and receiver." />
            )}
        </SignoffBlock>
      </CardContent>
    </Card>
  );

  function SignoffBlock({
    title,
    done,
    name,
    signature,
    stamp,
    children,
  }: {
    title: string;
    done: boolean;
    name?: string;
    signature?: string;
    stamp?: string | Date;
    children?: React.ReactNode;
  }) {
    return (
      <div className="rounded-lg border border-border/60 p-4 space-y-3">
        <div className="flex items-center justify-between">
          <span className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            {title}
          </span>
          {done && (
            <Badge
              variant="outline"
              className="gap-1 border-emerald-300 text-emerald-700 bg-emerald-50"
            >
              <CheckCircle2 className="h-3 w-3" />
              Signed
            </Badge>
          )}
        </div>
        {done ? (
          <div className="space-y-2">
            <div className="text-sm font-medium">{name || "—"}</div>
            {signature ? (
              <div className="rounded border border-dashed bg-white p-2">
                {/* signature is a data URI from the digital pad */}
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img
                  src={signature}
                  alt={`${title} signature`}
                  className="max-h-20 object-contain"
                />
              </div>
            ) : null}
            <p className="text-xs text-muted-foreground">
              {formatStamp(stamp)}
            </p>
          </div>
        ) : (
          children
        )}
      </div>
    );
  }
}

function SignoffBadge({ status }: { status: GRNSignoffStatus }) {
  const meta: Record<
    GRNSignoffStatus,
    { label: string; cls: string }
  > = {
    PENDING_RECEIVER: {
      label: "Awaiting receiver",
      cls: "border-amber-300 text-amber-700 bg-amber-50",
    },
    PENDING_CERTIFIER: {
      label: "Awaiting certifier",
      cls: "border-blue-300 text-blue-700 bg-blue-50",
    },
    READY: {
      label: "Ready",
      cls: "border-emerald-300 text-emerald-700 bg-emerald-50",
    },
    COMPLETED: {
      label: "Completed",
      cls: "border-slate-300 text-slate-700 bg-slate-50",
    },
  };
  const m = meta[status] ?? meta.PENDING_RECEIVER;
  return (
    <Badge variant="outline" className={m.cls}>
      {m.label}
    </Badge>
  );
}

function LockedHint({ text }: { text: string }) {
  return (
    <div className="flex items-start gap-2 rounded-md bg-muted/40 p-2 text-xs text-muted-foreground">
      <Lock className="mt-0.5 h-3.5 w-3.5" />
      <span>{text}</span>
    </div>
  );
}
