"use client";

import { useCallback, useState } from "react";
import { ColumnDef } from "@tanstack/react-table";
import { ArrowUpDown, Pencil, Plus, PowerOff, Power } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { DataTable } from "@/components/ui/data-table";
import { Vendor } from "@/types/vendor";
import { useVendors, useToggleVendorStatus } from "@/hooks/use-vendor-queries";
import { usePermissions } from "@/hooks/use-permissions";
import type { ActionButton } from "@/components/ui/action-buttons";
import { VendorFormDialog } from "./vendor-form-sheet";

interface VendorsTableProps {
  userId: string;
  userRole: string;
}

const VENDOR_EDIT_ROLES = ["admin", "approver"];

function formatDate(date: Date | string) {
  return new Date(date).toLocaleDateString("en-GB", {
    day: "2-digit",
    month: "short",
    year: "numeric",
  });
}

const columns: ColumnDef<Vendor>[] = [
  {
    id: "vendorCode",
    accessorKey: "vendorCode",
    header: ({ column }) => (
      <Button
        variant="ghost"
        size="sm"
        className="-ml-3 h-8"
        onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
      >
        Code
        <ArrowUpDown className="ml-2 h-3.5 w-3.5 opacity-60" />
      </Button>
    ),
    cell: ({ row }) => (
      <span className="font-mono text-xs text-muted-foreground">
        {row.original.vendorCode}
      </span>
    ),
  },
  {
    id: "name",
    accessorKey: "name",
    header: ({ column }) => (
      <Button
        variant="ghost"
        size="sm"
        className="-ml-3 h-8"
        onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
      >
        Name
        <ArrowUpDown className="ml-2 h-3.5 w-3.5 opacity-60" />
      </Button>
    ),
    cell: ({ row }) => <span className="font-medium">{row.original.name}</span>,
  },
  {
    id: "email",
    accessorKey: "email",
    header: "Email",
    cell: ({ row }) => (
      <span className="text-sm text-muted-foreground">
        {row.original.email || "—"}
      </span>
    ),
  },
  {
    id: "phone",
    accessorKey: "phone",
    header: "Phone",
    cell: ({ row }) => (
      <span className="text-sm">{row.original.phone || "—"}</span>
    ),
  },
  {
    id: "country",
    accessorKey: "country",
    header: "Location",
    cell: ({ row }) => {
      const { country, city } = row.original;
      if (!country && !city)
        return <span className="text-muted-foreground">—</span>;
      return (
        <span className="text-sm">
          {[city, country].filter(Boolean).join(", ")}
        </span>
      );
    },
  },
  {
    id: "active",
    accessorKey: "active",
    header: "Status",
    cell: ({ row }) =>
      row.original.active ? (
        <Badge variant="default" className="bg-green-600 hover:bg-green-700">
          Active
        </Badge>
      ) : (
        <Badge variant="secondary">Inactive</Badge>
      ),
  },
  {
    id: "createdAt",
    accessorKey: "createdAt",
    header: ({ column }) => (
      <Button
        variant="ghost"
        size="sm"
        className="-ml-3 h-8"
        onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
      >
        Created
        <ArrowUpDown className="ml-2 h-3.5 w-3.5 opacity-60" />
      </Button>
    ),
    cell: ({ row }) => (
      <span className="text-sm text-muted-foreground">
        {formatDate(row.original.createdAt)}
      </span>
    ),
  },
];

export function VendorsTable({ userRole }: VendorsTableProps) {
  const { data: vendors = [], isLoading } = useVendors();
  const { rawPermissions } = usePermissions();
  const toggleStatus = useToggleVendorStatus();

  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingVendor, setEditingVendor] = useState<Vendor | null>(null);

  const canEdit =
    VENDOR_EDIT_ROLES.includes(userRole) ||
    rawPermissions.includes("vendor.edit");

  const canCreate =
    VENDOR_EDIT_ROLES.includes(userRole) ||
    rawPermissions.includes("vendor.create");

  function openCreate() {
    setEditingVendor(null);
    setDialogOpen(true);
  }

  function openEdit(vendor: Vendor) {
    setEditingVendor(vendor);
    setDialogOpen(true);
  }

  const getActions = useCallback(
    (vendor: Vendor): ActionButton[] => {
      if (!canEdit) return [];
      return [
        {
          icon: <Pencil className="h-3.5 w-3.5" />,
          label: "Edit",
          tooltip: "Edit vendor",
          onClick: () => openEdit(vendor),
        },
        {
          icon: vendor.active ? (
            <PowerOff className="h-3.5 w-3.5" />
          ) : (
            <Power className="h-3.5 w-3.5" />
          ),
          label: vendor.active ? "Deactivate" : "Activate",
          tooltip: vendor.active ? "Deactivate vendor" : "Activate vendor",
          onClick: () =>
            toggleStatus.mutate({ id: vendor.id, active: !vendor.active }),
        },
      ];
    },
    [canEdit, toggleStatus],
  );

  return (
    <div className="space-y-3">
      {canCreate && (
        <div className="flex justify-end">
          <Button size="sm" onClick={openCreate}>
            <Plus className="mr-2 h-4 w-4" />
            Add Vendor
          </Button>
        </div>
      )}

      <DataTable
        columns={columns}
        data={vendors}
        isLoading={isLoading}
        searchKey="name"
        searchPlaceholder="Search vendors..."
        actions={getActions}
        emptyState={{
          title: "No vendors yet",
          description: "Add your first vendor to get started.",
        }}
      />

      <VendorFormDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        vendor={editingVendor}
      />
    </div>
  );
}
