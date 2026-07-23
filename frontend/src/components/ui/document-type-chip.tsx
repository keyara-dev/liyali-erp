import * as React from "react";
import { Badge } from "@/components/ui/badge";
import {
  FileText,
  ShoppingCart,
  Receipt,
  PackageCheck,
  Wallet,
  type LucideIcon,
} from "lucide-react";
import { cn } from "@/lib/utils";

export type DocumentType =
  | "requisition"
  | "purchase_order"
  | "payment_voucher"
  | "grn"
  | "goods_received_note"
  | "budget"
  | string;

interface TypeMeta {
  label: string;
  icon: LucideIcon;
}

const META: Record<string, TypeMeta> = {
  requisition: { label: "Requisition", icon: FileText },
  purchase_order: { label: "Purchase Order", icon: ShoppingCart },
  payment_voucher: { label: "Payment Voucher", icon: Receipt },
  grn: { label: "GRN", icon: PackageCheck },
  goods_received_note: { label: "GRN", icon: PackageCheck },
  budget: { label: "Budget", icon: Wallet },
};

function titleCase(s: string) {
  return s
    .replace(/_/g, " ")
    .toLowerCase()
    .replace(/\b\w/g, (c) => c.toUpperCase());
}

export interface DocumentTypeChipProps {
  type: DocumentType;
  /** Show icon next to label. Default true. */
  showIcon?: boolean;
  className?: string;
}

export function DocumentTypeChip({ type, showIcon = true, className }: DocumentTypeChipProps) {
  const key = type.toLowerCase();
  const meta = META[key];
  const label = meta?.label || titleCase(type);
  const Icon = meta?.icon || FileText;
  return (
    <Badge variant="outline" className={cn("gap-1 px-2 py-0.5 font-medium", className)}>
      {showIcon && <Icon className="h-3 w-3" aria-hidden="true" />}
      {label}
    </Badge>
  );
}
