"use client";

import { useState } from "react";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { SearchSelectField } from "@/components/ui/search-select-field";
import { useVendors } from "@/hooks/use-vendor-queries";
import { useUsersQuery } from "@/hooks/use-users-query";
import { usePayees } from "@/hooks/use-payee-queries";
import { createPayee } from "@/app/_actions/payees";
import type { PayeeSnapshot, PayeeType, CreatePayeeInput } from "@/types/payee";
import type { Vendor } from "@/types/core";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export type PayeeSource = "new" | "vendor" | "employee" | "other";

export interface PayeeBlockValues {
  payeeId?: string;
  payeeSnapshot?: PayeeSnapshot;
  /** Only populated when source === "new", used for persistence at submit */
  newPayeeInput?: NewPayeeFields;
  /** The source mode selected by the user */
  payeeSource: PayeeSource;
}

interface NewPayeeFields {
  name: string;
  email: string;
  phone: string;
  bankName: string;
  bankAccount: string;
  taxId: string;
  payeeType: PayeeType;
}

interface PayeeBlockProps {
  value: PayeeBlockValues;
  onChange: (next: PayeeBlockValues) => void;
}

// ---------------------------------------------------------------------------
// Helper — persist a "new" payee on form submit
// ---------------------------------------------------------------------------

/**
 * If `values.payeeSource === "new"` and no payeeId is set yet, calls the
 * createPayee server action and returns the resolved { payeeId, payeeSnapshot }.
 * Otherwise returns the values unchanged.
 */
export async function persistNewPayeeIfNeeded(
  values: PayeeBlockValues,
): Promise<{ payeeId: string | undefined; payeeSnapshot: PayeeSnapshot | undefined }> {
  if (values.payeeSource !== "new" || values.payeeId) {
    return { payeeId: values.payeeId, payeeSnapshot: values.payeeSnapshot };
  }

  const fields = values.newPayeeInput;
  if (!fields?.name?.trim()) {
    throw new Error("Payee name is required for direct payment");
  }

  const input: CreatePayeeInput = {
    payeeType: fields.payeeType,
    name: fields.name.trim(),
    email: fields.email?.trim() || undefined,
    phone: fields.phone?.trim() || undefined,
    bankName: fields.bankName?.trim() || undefined,
    bankAccount: fields.bankAccount?.trim() || undefined,
    taxId: fields.taxId?.trim() || undefined,
  };

  const result = await createPayee(input);
  if (!result.success || !result.data) {
    throw new Error(result.message || "Failed to create payee");
  }

  const payee = result.data;
  return {
    payeeId: payee.id,
    payeeSnapshot: {
      name: payee.name,
      payeeType: payee.payeeType,
      email: payee.email,
      phone: payee.phone,
      bankName: payee.bankName,
      bankAccount: payee.bankAccount,
      taxId: payee.taxId,
    },
  };
}

// ---------------------------------------------------------------------------
// PayeeBlock component
// ---------------------------------------------------------------------------

export function PayeeBlock({ value, onChange }: PayeeBlockProps) {
  const { data: vendors = [], isLoading: vendorsLoading } = useVendors();
  const { data: usersData, isLoading: usersLoading } = useUsersQuery(1, 50);
  const { data: otherPayees = [], isLoading: otherLoading } = usePayees("other");

  const users = Array.isArray(usersData) ? usersData : (usersData as any)?.items ?? [];

  // Local new-payee form fields (controlled internally, lifted via onChange)
  const newFields: NewPayeeFields = value.newPayeeInput ?? {
    name: "",
    email: "",
    phone: "",
    bankName: "",
    bankAccount: "",
    taxId: "",
    payeeType: "other",
  };

  const setSource = (source: PayeeSource) => {
    onChange({
      payeeSource: source,
      payeeId: undefined,
      payeeSnapshot: undefined,
      newPayeeInput: source === "new" ? newFields : undefined,
    });
  };

  const updateNewField = (field: keyof NewPayeeFields, val: string) => {
    const updated: NewPayeeFields = { ...newFields, [field]: val };
    onChange({
      ...value,
      payeeSource: "new",
      newPayeeInput: updated,
      // Update snapshot preview so parent can show a name
      payeeSnapshot: {
        name: updated.name,
        payeeType: updated.payeeType,
        email: updated.email || undefined,
        phone: updated.phone || undefined,
        bankName: updated.bankName || undefined,
        bankAccount: updated.bankAccount || undefined,
        taxId: updated.taxId || undefined,
      },
    });
  };

  const handleVendorSelect = (vendorId: string) => {
    const vendor = vendors.find((v: Vendor) => v.id === vendorId);
    if (!vendor) return;
    onChange({
      payeeSource: "vendor",
      payeeId: undefined,
      payeeSnapshot: {
        name: vendor.name,
        payeeType: "vendor",
        email: vendor.email,
        phone: vendor.phone,
        bankName: vendor.bankName,
        bankAccount: vendor.accountNumber,
        taxId: vendor.taxId,
      },
    });
  };

  const handleEmployeeSelect = (userId: string) => {
    const user = users.find((u: any) => u.id === userId);
    if (!user) return;
    onChange({
      payeeSource: "employee",
      payeeId: undefined,
      payeeSnapshot: {
        name: user.name || `${user.first_name ?? ""} ${user.last_name ?? ""}`.trim(),
        payeeType: "employee",
        email: user.email,
      },
    });
  };

  const handleOtherPayeeSelect = (payeeId: string) => {
    const payee = otherPayees.find((p: any) => p.id === payeeId);
    if (!payee) return;
    onChange({
      payeeSource: "other",
      payeeId: payee.id,
      payeeSnapshot: {
        name: payee.name,
        payeeType: "other",
        email: payee.email,
        phone: payee.phone,
        bankName: payee.bankName,
        bankAccount: payee.bankAccount,
        taxId: payee.taxId,
      },
    });
  };

  // Derive current selected IDs for controlled selects
  const selectedVendorId =
    value.payeeSource === "vendor"
      ? (vendors.find((v: Vendor) => v.name === value.payeeSnapshot?.name)?.id ?? "")
      : "";

  const selectedUserId =
    value.payeeSource === "employee"
      ? (users.find((u: any) => u.name === value.payeeSnapshot?.name)?.id ?? "")
      : "";

  const selectedOtherPayeeId =
    value.payeeSource === "other" ? (value.payeeId ?? "") : "";

  return (
    <div className="rounded-lg border border-border p-4 space-y-4">
      <div className="flex items-center justify-between">
        <Label className="text-sm font-semibold">Payee</Label>
      </div>

      {/* Source mode tabs */}
      <RadioGroup
        value={value.payeeSource}
        onValueChange={(v) => setSource(v as PayeeSource)}
        className="grid grid-cols-2 sm:grid-cols-4 gap-2"
      >
        {(
          [
            { value: "new", label: "New" },
            { value: "vendor", label: "Vendor" },
            { value: "employee", label: "Employee" },
            { value: "other", label: "Other" },
          ] as { value: PayeeSource; label: string }[]
        ).map((opt) => (
          <label
            key={opt.value}
            className={`flex items-center gap-2 cursor-pointer rounded-md border px-3 py-2 text-sm transition-colors ${
              value.payeeSource === opt.value
                ? "border-primary bg-primary/5 text-primary font-medium"
                : "border-border text-muted-foreground hover:border-foreground/30"
            }`}
          >
            <RadioGroupItem value={opt.value} />
            {opt.label}
          </label>
        ))}
      </RadioGroup>

      {/* ── New payee form ── */}
      {value.payeeSource === "new" && (
        <div className="space-y-3">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <Input
              label="Full Name"
              required
              placeholder="e.g., Mwale Phiri"
              value={newFields.name}
              onChange={(e) => updateNewField("name", e.target.value)}
            />
            <div className="space-y-1">
              <Label className="text-sm font-medium text-slate-700 dark:text-slate-300">
                Payee Type
              </Label>
              <RadioGroup
                value={newFields.payeeType}
                onValueChange={(v) => updateNewField("payeeType", v as PayeeType)}
                className="flex gap-4"
              >
                {(["employee", "vendor", "other"] as PayeeType[]).map((t) => (
                  <label
                    key={t}
                    className="flex items-center gap-1.5 cursor-pointer text-sm capitalize"
                  >
                    <RadioGroupItem value={t} />
                    {t}
                  </label>
                ))}
              </RadioGroup>
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <Input
              label="Email"
              type="email"
              placeholder="email@example.com"
              value={newFields.email}
              onChange={(e) => updateNewField("email", e.target.value)}
            />
            <Input
              label="Phone"
              placeholder="+260 97X XXX XXX"
              value={newFields.phone}
              onChange={(e) => updateNewField("phone", e.target.value)}
            />
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <Input
              label="Bank Name"
              placeholder="e.g., Zanaco"
              value={newFields.bankName}
              onChange={(e) => updateNewField("bankName", e.target.value)}
            />
            <Input
              label="Bank Account"
              placeholder="Account number"
              value={newFields.bankAccount}
              onChange={(e) => updateNewField("bankAccount", e.target.value)}
            />
          </div>

          <Input
            label="Tax ID (TPIN)"
            placeholder="Optional"
            value={newFields.taxId}
            onChange={(e) => updateNewField("taxId", e.target.value)}
          />
        </div>
      )}

      {/* ── Vendor picker ── */}
      {value.payeeSource === "vendor" && (
        <SearchSelectField
          label="Select Vendor"
          placeholder="Search vendors…"
          isLoading={vendorsLoading}
          value={selectedVendorId}
          options={vendors.map((v: Vendor) => ({ id: v.id, name: v.name }))}
          onValueChange={handleVendorSelect}
        />
      )}

      {/* ── Employee picker ── */}
      {value.payeeSource === "employee" && (
        <SearchSelectField
          label="Select Employee"
          placeholder="Search employees…"
          isLoading={usersLoading}
          value={selectedUserId}
          options={users.map((u: any) => ({
            id: u.id,
            name: u.name || `${u.first_name ?? ""} ${u.last_name ?? ""}`.trim(),
          }))}
          onValueChange={handleEmployeeSelect}
        />
      )}

      {/* ── Other payee picker ── */}
      {value.payeeSource === "other" && (
        <SearchSelectField
          label="Select Payee"
          placeholder="Search payees…"
          isLoading={otherLoading}
          value={selectedOtherPayeeId}
          options={(otherPayees as any[]).map((p) => ({ id: p.id, name: p.name }))}
          onValueChange={handleOtherPayeeSelect}
        />
      )}

      {/* Snapshot preview */}
      {value.payeeSnapshot?.name && value.payeeSource !== "new" && (
        <div className="rounded-md bg-muted/40 px-3 py-2 text-xs text-muted-foreground space-y-0.5">
          <p className="font-medium text-foreground">{value.payeeSnapshot.name}</p>
          {value.payeeSnapshot.email && <p>{value.payeeSnapshot.email}</p>}
          {value.payeeSnapshot.bankName && (
            <p>
              {value.payeeSnapshot.bankName}
              {value.payeeSnapshot.bankAccount
                ? ` · ${value.payeeSnapshot.bankAccount}`
                : ""}
            </p>
          )}
        </div>
      )}
    </div>
  );
}
