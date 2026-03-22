"use client";

import { useState, useEffect } from "react";
import { Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
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
  active: true,
};

export function VendorFormDialog({
  open,
  onOpenChange,
  vendor,
}: VendorFormDialogProps) {
  const isEdit = !!vendor;
  const [form, setForm] = useState(EMPTY_FORM);
  const [errors, setErrors] = useState<Partial<typeof EMPTY_FORM>>({});

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
    if (errors[field as keyof typeof errors]) {
      setErrors((prev) => ({ ...prev, [field]: undefined }));
    }
  }

  function validate() {
    const next: Partial<typeof EMPTY_FORM> = {};
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
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{isEdit ? "Edit Vendor" : "Add Vendor"}</DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4 py-2">
          {/* Name */}
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

          {/* Email + Phone */}
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

          {/* Country + City */}
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

          {/* Bank Account + Tax ID */}
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-1.5">
              <Label htmlFor="bankAccount">Bank Account</Label>
              <Input
                id="bankAccount"
                value={form.bankAccount}
                onChange={(e) => set("bankAccount", e.target.value)}
                placeholder="Account number"
                disabled={isPending}
              />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="taxId">Tax ID</Label>
              <Input
                id="taxId"
                value={form.taxId}
                onChange={(e) => set("taxId", e.target.value)}
                placeholder="Tax / TPIN number"
                disabled={isPending}
              />
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

          <DialogFooter>
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
