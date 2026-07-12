"use client";

import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Mail, Phone, MapPin, User, CalendarDays } from "lucide-react";
import { cn } from "@/lib/utils";
import type { Vendor } from "@/types/core";

interface VendorProfileCardProps {
  vendor: Vendor;
  className?: string;
}

function initials(name: string) {
  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((w) => w[0]!.toUpperCase())
    .join("");
}

function Field({
  icon: Icon,
  label,
  value,
  href,
}: {
  icon: React.ElementType;
  label: string;
  value: React.ReactNode;
  href?: string;
}) {
  const body = (
    <p className="text-sm break-words">
      {value || <span className="text-muted-foreground">—</span>}
    </p>
  );
  return (
    <div className="flex items-start gap-3 min-w-0">
      <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-muted text-muted-foreground">
        <Icon className="h-4 w-4" aria-hidden="true" />
      </span>
      <div className="min-w-0 flex-1 space-y-0.5">
        <p className="text-[11px] font-medium text-muted-foreground uppercase tracking-wider">
          {label}
        </p>
        {href && value ? (
          <a
            href={href}
            className="block text-sm break-words text-primary hover:underline underline-offset-4"
          >
            {value}
          </a>
        ) : (
          body
        )}
      </div>
    </div>
  );
}

export function VendorProfileCard({ vendor, className }: VendorProfileCardProps) {
  const fullAddress = [vendor.physicalAddress, vendor.city, vendor.country]
    .filter(Boolean)
    .join(", ");

  return (
    <Card className={cn("border-border/60 overflow-hidden", className)}>
      {/* Identity strip */}
      <div className="flex flex-wrap items-center gap-3 sm:gap-4 border-b border-border/60 bg-muted/30 px-4 py-4 sm:px-6">
        <div className="flex h-12 w-12 sm:h-14 sm:w-14 shrink-0 items-center justify-center rounded-xl bg-linear-to-br from-primary/80 to-primary text-primary-foreground text-lg sm:text-xl font-bold shadow-md">
          {initials(vendor.name)}
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex flex-wrap items-center gap-2">
            <h2 className="text-base sm:text-lg font-semibold leading-tight truncate">
              {vendor.name}
            </h2>
            <Badge
              variant="outline"
              className={cn(
                "text-[11px]",
                vendor.active
                  ? "border-emerald-500/40 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
                  : "border-rose-500/40 bg-rose-500/10 text-rose-600 dark:text-rose-400"
              )}
            >
              {vendor.active ? "Active" : "Inactive"}
            </Badge>
          </div>
          <p className="mt-0.5 font-mono text-xs text-muted-foreground tracking-wide">
            {vendor.vendorCode}
          </p>
        </div>
        {vendor.createdAt && (
          <div className="hidden sm:flex items-center gap-1.5 text-xs text-muted-foreground">
            <CalendarDays className="h-3.5 w-3.5" aria-hidden="true" />
            Since {new Date(vendor.createdAt).toLocaleDateString()}
          </div>
        )}
      </div>

      <CardContent className="grid gap-4 sm:gap-5 pt-4 sm:grid-cols-2">
        <Field
          icon={User}
          label="Contact person"
          value={vendor.contactPerson}
        />
        <Field
          icon={Mail}
          label="Email"
          value={vendor.email}
          href={vendor.email ? `mailto:${vendor.email}` : undefined}
        />
        <Field
          icon={Phone}
          label="Phone"
          value={vendor.phone}
          href={vendor.phone ? `tel:${vendor.phone}` : undefined}
        />
        <Field
          icon={MapPin}
          label="Address"
          value={fullAddress || vendor.physicalAddress}
        />
      </CardContent>
    </Card>
  );
}
