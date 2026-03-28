"use client";

import { useState, useEffect } from "react";
import { Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Vendor } from "@/types/vendor";
import { useCreateVendor, useUpdateVendor } from "@/hooks/use-vendor-queries";

interface VendorFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  vendor?: Vendor | null;
}

const EMPTY_FORM = {
  name: "",
  email: "",
  phone: "",
  country: "",
  city: "",
  bankAccount: "",
  taxId: "",
  bankName: "",
  accountName: "",
  accountNumber: "",
  branchCode: "",
  swiftCode: "",
  contactPerson: "",
  physicalAddress: "",
  active: true,
};

export function VendorFormDialog({
  open,
  onOpenChange,
  vendor,
}: VendorFormDialogProps) {
  const isEdit = !!vendor;
  const [form, setForm] = useState(EMPTY_FORM);
  const [errors, setErrors] = useState<Partial<Record<keyof typeof EMPTY_FORM, string>>>({});

  const createMutation = useCreateVendor(() => onOpenChange(false));
  const updateMutation = useUpdateVendor(() => onOpenChange(false));

  const isPending = createMutation.isPending || updateMutation.isPending;

  useEffect(() => {
    if (open) {
      if (vendor) {
        setForm({
          name: vendor.name ?? "",
          email: vendor.email ?? "",
          phone: vendor.phone ?? "",
          country: vendor.country ?? "",
          city: vendor.city ?? "",
          bankAccount: vendor.bankAccount ?? "",
          taxId: vendor.taxId ?? "",
          bankName: vendor.bankName ?? "",
          accountName: vendor.accountName ?? "",
          accountNumber: vendor.accountNumber ?? "",
          branchCode: vendor.branchCode ?? "",
          swiftCode: vendor.swiftCode ?? "",
          contactPerson: vendor.contactPerson ?? "",
          physicalAddress: vendor.physicalAddress ?? "",
          active: vendor.active,
        });
      } else {
        setForm(EMPTY_FORM);
      }
      setErrors({});
    }
  }, [open, vendor]);

  function set(field: keyof typeof EMPTY_FORM, value: string | boolean) {
    setForm((prev) => ({ ...prev, [field]: value }));
    if (errors[field]) {
      setErrors((prev) => ({ ...prev, [field]: undefined }));
    }
  }

  function validate() {
    const next: Partial<Record<keyof typeof EMPTY_FORM, string>> = {};
    if (!form.name.trim()) next.name = "Name is required";
    if (!form.email.trim()) next.email = "Email is required";
    else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email))
      next.email = "Invalid email address";
    if (!form.phone.trim()) next.phone = "Phone is required";
    if (!form.country.trim()) next.country = "Country is required";
    if (!form.city.trim()) next.city = "City is required";
    setErrors(next);
    return Object.keys(next).length === 0;
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!validate()) return;

    if (isEdit && vendor) {
      updateMutation.mutate({ id: vendor.id, data: form });
    } else {
      const { active, ...createData } = form;
      createMutation.mutate(createData);
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-xl overflow-y-auto max-h-[90svh] p-0">
        <DialogHeader className="px-6 pt-6 pb-2">
          <DialogTitle>{isEdit ? "Edit Vendor" : "Add Vendor"}</DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="px-6 pb-2 space-y-5">
          {/* ── Basic Info ── */}
          <div className="space-y-1.5">
            <Label htmlFor="name">
              Name <span className="text-destructive">*</span>
            </Label>
            <Input
              id="name"
              value={form.name}
              onChange={(e) => set("name", e.target.value)}
              placeholder="Supplier name"
              disabled={isPending}
            />
            {errors.name && (
              <p className="text-xs text-destructive">{errors.name}</p>
            )}
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <Label htmlFor="email">
                Email <span className="text-destructive">*</span>
              </Label>
              <Input
                id="email"
                type="email"
                value={form.email}
                onChange={(e) => set("email", e.target.value)}
                placeholder="vendor@example.com"
                disabled={isPending}
              />
              {errors.email && (
                <p className="text-xs text-destructive">{errors.email}</p>
              )}
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="phone">
                Phone <span className="text-destructive">*</span>
              </Label>
              <Input
                id="phone"
                value={form.phone}
                onChange={(e) => set("phone", e.target.value)}
                placeholder="+260 97..."
                disabled={isPending}
              />
              {errors.phone && (
                <p className="text-xs text-destructive">{errors.phone}</p>
              )}
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <Label htmlFor="country">
                Country <span className="text-destructive">*</span>
              </Label>
              <Input
                id="country"
                value={form.country}
                onChange={(e) => set("country", e.target.value)}
                placeholder="Zambia"
                disabled={isPending}
              />
              {errors.country && (
                <p className="text-xs text-destructive">{errors.country}</p>
              )}
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="city">
                City <span className="text-destructive">*</span>
              </Label>
              <Input
                id="city"
                value={form.city}
                onChange={(e) => set("city", e.target.value)}
                placeholder="Lusaka"
                disabled={isPending}
              />
              {errors.city && (
                <p className="text-xs text-destructive">{errors.city}</p>
              )}
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <Label htmlFor="contactPerson">Contact Person</Label>
              <Input
                id="contactPerson"
                value={form.contactPerson}
                onChange={(e) => set("contactPerson", e.target.value)}
                placeholder="Primary contact name"
                disabled={isPending}
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="taxId">Tax ID / TPIN</Label>
              <Input
                id="taxId"
                value={form.taxId}
                onChange={(e) => set("taxId", e.target.value)}
                placeholder="Tax / TPIN number"
                disabled={isPending}
              />
            </div>
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="physicalAddress">Physical Address</Label>
            <Textarea
              id="physicalAddress"
              value={form.physicalAddress}
              onChange={(e) => set("physicalAddress", e.target.value)}
              placeholder="Street address, building, area..."
              rows={2}
              disabled={isPending}
            />
          </div>

          {/* ── Bank Details ── */}
          <div className="space-y-3 pt-2 border-t">
            <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
              Bank Details
            </p>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label htmlFor="bankName">Bank Name</Label>
                <Input
                  id="bankName"
                  value={form.bankName}
                  onChange={(e) => set("bankName", e.target.value)}
                  placeholder="e.g. Zanaco"
                  disabled={isPending}
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="accountName">Account Name</Label>
                <Input
                  id="accountName"
                  value={form.accountName}
                  onChange={(e) => set("accountName", e.target.value)}
                  placeholder="Name on account"
                  disabled={isPending}
                />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label htmlFor="accountNumber">Account Number</Label>
                <Input
                  id="accountNumber"
                  value={form.accountNumber}
                  onChange={(e) => set("accountNumber", e.target.value)}
                  placeholder="Account number"
                  disabled={isPending}
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="bankAccount">Legacy Bank Account</Label>
                <Input
                  id="bankAccount"
                  value={form.bankAccount}
                  onChange={(e) => set("bankAccount", e.target.value)}
                  placeholder="Legacy field"
                  disabled={isPending}
                />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1.5">
                <Label htmlFor="branchCode">Branch Code</Label>
                <Input
                  id="branchCode"
                  value={form.branchCode}
                  onChange={(e) => set("branchCode", e.target.value)}
                  placeholder="Sort/branch code"
                  disabled={isPending}
                />
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="swiftCode">SWIFT / BIC Code</Label>
                <Input
                  id="swiftCode"
                  value={form.swiftCode}
                  onChange={(e) => set("swiftCode", e.target.value)}
                  placeholder="SWIFT code"
                  disabled={isPending}
                />
              </div>
            </div>
          </div>

          {/* Active toggle — edit only */}
          {isEdit && (
            <div className="flex items-center justify-between rounded-lg border p-3">
              <div>
                <p className="text-sm font-medium">Active</p>
                <p className="text-xs text-muted-foreground">
                  Inactive vendors cannot be selected on new documents
                </p>
              </div>
              <Switch
                checked={form.active}
                onCheckedChange={(v) => set("active", v)}
                disabled={isPending}
              />
            </div>
          )}

          <DialogFooter className="pb-4 pt-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isPending}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isPending}>
              {isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              {isEdit ? "Save Changes" : "Add Vendor"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
