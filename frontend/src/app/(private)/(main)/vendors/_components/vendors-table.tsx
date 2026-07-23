"use client";

import { useCallback, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { Pencil, Plus, PowerOff, Power } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { DataList, DataListColumn } from "@/components/ui/data-list";
import { FilterBar } from "@/components/ui/filter-bar";
import { Vendor } from "@/types/vendor";
import { useVendors, useToggleVendorStatus } from "@/hooks/use-vendor-queries";
import { usePermissions } from "@/hooks/use-permissions";
import { useDebounce } from "@/hooks/use-debounce";
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

export function VendorsTable({ userRole }: VendorsTableProps) {
  const router = useRouter();
  const { data: vendors = [], isLoading } = useVendors();
  const { rawPermissions } = usePermissions();
  const toggleStatus = useToggleVendorStatus();

  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingVendor, setEditingVendor] = useState<Vendor | null>(null);

  // Filter state
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const debouncedSearch = useDebounce(searchQuery, 500);

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

  function handleRowClick(vendor: Vendor) {
    router.push(`/vendors/${vendor.id}`);
  }

  // Derived filter values
  const hasActiveFilters = Boolean(searchQuery) || statusFilter !== "all";
  const clearFilters = () => {
    setSearchQuery("");
    setStatusFilter("all");
  };

  const filteredVendors = useMemo(() => {
    let filtered = vendors;
    if (statusFilter !== "all") {
      const isActive = statusFilter === "active";
      filtered = filtered.filter((v) => v.active === isActive);
    }
    if (debouncedSearch) {
      const s = debouncedSearch.toLowerCase();
      filtered = filtered.filter(
        (v) =>
          v.name?.toLowerCase().includes(s) ||
          v.vendorCode?.toLowerCase().includes(s) ||
          v.email?.toLowerCase().includes(s) ||
          v.city?.toLowerCase().includes(s) ||
          v.country?.toLowerCase().includes(s),
      );
    }
    return filtered;
  }, [vendors, statusFilter, debouncedSearch]);

  const getActionButtons = useCallback(
    (vendor: Vendor) => {
      if (!canEdit) return null;
      return (
        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7"
            onClick={() => openEdit(vendor)}
            title="Edit vendor"
          >
            <Pencil className="h-3.5 w-3.5" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="h-7 w-7"
            onClick={() =>
              toggleStatus.mutate({ id: vendor.id, active: !vendor.active })
            }
            title={vendor.active ? "Deactivate vendor" : "Activate vendor"}
          >
            {vendor.active ? (
              <PowerOff className="h-3.5 w-3.5" />
            ) : (
              <Power className="h-3.5 w-3.5" />
            )}
          </Button>
        </div>
      );
    },
    [canEdit, toggleStatus],
  );

  const columns: DataListColumn<Vendor>[] = useMemo(
    () => [
      {
        id: "vendorCode",
        header: "Code",
        priority: "always",
        cell: (row) => (
          <span className="font-mono text-xs text-muted-foreground">
            {row.vendorCode}
          </span>
        ),
      },
      {
        id: "name",
        header: "Name",
        priority: "always",
        cell: (row) => <span className="font-medium">{row.name}</span>,
      },
      {
        id: "email",
        header: "Email",
        priority: "lg",
        cell: (row) => (
          <span className="text-sm text-muted-foreground">
            {row.email || "—"}
          </span>
        ),
      },
      {
        id: "phone",
        header: "Phone",
        priority: "lg",
        cell: (row) => (
          <span className="text-sm">{row.phone || "—"}</span>
        ),
      },
      {
        id: "location",
        header: "Location",
        priority: "md",
        cell: (row) => {
          const { country, city } = row;
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
        id: "createdAt",
        header: "Created",
        priority: "lg",
        cell: (row) => (
          <span className="text-sm text-muted-foreground">
            {formatDate(row.createdAt)}
          </span>
        ),
      },
      {
        id: "status",
        header: "Status",
        priority: "always",
        cell: (row) =>
          row.active ? (
            <Badge variant="default" className="bg-green-600 hover:bg-green-700">
              Active
            </Badge>
          ) : (
            <Badge variant="secondary">Inactive</Badge>
          ),
      },
      {
        id: "actions",
        header: <span className="sr-only">Actions</span>,
        priority: "always",
        align: "right",
        cell: (row) => getActionButtons(row),
      },
    ],
    [getActionButtons],
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

      <FilterBar
        search={
          <Input
            placeholder="Search vendors..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="h-8 text-sm"
          />
        }
        filters={
          <Select value={statusFilter} onValueChange={setStatusFilter}>
            <SelectTrigger className="h-8 w-32.5 text-xs">
              <SelectValue placeholder="Status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All statuses</SelectItem>
              <SelectItem value="active">Active</SelectItem>
              <SelectItem value="inactive">Inactive</SelectItem>
            </SelectContent>
          </Select>
        }
        hasActiveFilters={hasActiveFilters}
        onReset={clearFilters}
        meta={
          hasActiveFilters
            ? `Showing ${filteredVendors.length} of ${vendors.length}`
            : `${vendors.length} vendor${vendors.length === 1 ? "" : "s"}`
        }
      />

      <DataList
        rows={filteredVendors}
        columns={columns}
        getRowId={(row) => row.id}
        isLoading={isLoading}
        onRowClick={handleRowClick}
        emptyMessage={
          hasActiveFilters
            ? "No vendors match the current filters."
            : "No vendors yet. Add your first vendor to get started."
        }
        mobileCard={(row) => (
          <div className="flex flex-col gap-2">
            <div className="flex items-start justify-between gap-2">
              <div className="min-w-0">
                <div className="font-medium line-clamp-1">{row.name}</div>
                <div className="text-xs text-muted-foreground font-mono">
                  {row.vendorCode}
                </div>
              </div>
              {row.active ? (
                <Badge
                  variant="default"
                  className="bg-green-600 hover:bg-green-700 shrink-0"
                >
                  Active
                </Badge>
              ) : (
                <Badge variant="secondary" className="shrink-0">
                  Inactive
                </Badge>
              )}
            </div>
            <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
              {(row.city || row.country) && (
                <span className="truncate">
                  {[row.city, row.country].filter(Boolean).join(", ")}
                </span>
              )}
              {row.email && <span className="truncate">{row.email}</span>}
              {row.phone && <span>{row.phone}</span>}
            </div>
            <div className="pt-1">{getActionButtons(row)}</div>
          </div>
        )}
      />

      <VendorFormDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        vendor={editingVendor}
      />
    </div>
  );
}
