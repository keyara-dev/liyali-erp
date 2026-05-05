"use client";

import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Mail, Phone, MapPin, User } from "lucide-react";
import type { Vendor } from "@/types/core";

interface VendorProfileCardProps {
  vendor: Vendor;
}

function Field({
  icon: Icon,
  label,
  value,
}: {
  icon: React.ElementType;
  label: string;
  value: React.ReactNode;
}) {
  return (
    <div className="flex items-start gap-3">
      <Icon
        className="h-4 w-4 text-muted-foreground shrink-0 mt-0.5"
        aria-hidden="true"
      />
      <div className="min-w-0 flex-1">
        <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
          {label}
        </p>
        <p className="text-sm break-words">{value || "—"}</p>
      </div>
    </div>
  );
}

export function VendorProfileCard({ vendor }: VendorProfileCardProps) {
  const fullAddress = [vendor.physicalAddress, vendor.city, vendor.country]
    .filter(Boolean)
    .join(", ");

  return (
    <Card className="border-border/60">
      <CardHeader>
        <CardTitle className="text-base">Profile</CardTitle>
      </CardHeader>
      <CardContent className="grid gap-4 sm:grid-cols-2">
        <Field icon={User} label="Contact person" value={vendor.contactPerson} />
        <Field icon={Mail} label="Email" value={vendor.email} />
        <Field icon={Phone} label="Phone" value={vendor.phone} />
        <Field
          icon={MapPin}
          label="Address"
          value={fullAddress || vendor.physicalAddress}
        />
      </CardContent>
    </Card>
  );
}
