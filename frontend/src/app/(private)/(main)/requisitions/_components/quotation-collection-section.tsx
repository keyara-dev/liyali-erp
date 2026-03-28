"use client";

import { useState, useRef } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components";
import {
  FileText,
  Plus,
  ExternalLink,
  Upload,
  X,
} from "lucide-react";
import { Quotation } from "@/types/core";
import { Vendor } from "@/types/vendor";
import { uploadToImageKit } from "@/lib/imagekit";
import { toast } from "sonner";

interface QuotationCollectionSectionProps {
  quotations: Quotation[];
  requisitionId: string;
  currency: string;
  vendors: Vendor[];
  canEdit: boolean;
  onSave: (quotations: Quotation[]) => Promise<void>;
}

export function QuotationCollectionSection({
  quotations,
  currency,
  vendors,
  canEdit,
  onSave,
}: QuotationCollectionSectionProps) {
  const [showForm, setShowForm] = useState(false);
  const [saving, setSaving] = useState(false);
  const [vendorId, setVendorId] = useState("");
  const [vendorName, setVendorName] = useState("");
  const [amount, setAmount] = useState("");
  const [file, setFile] = useState<File | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const count = quotations.length;
  const hasEnough = count >= 3;

  function reset() {
    setVendorId("");
    setVendorName("");
    setAmount("");
    setFile(null);
    if (fileInputRef.current) fileInputRef.current.value = "";
    setShowForm(false);
  }

  async function handleAdd() {
    if (!vendorName.trim()) {
      toast.error("Vendor name is required");
      return;
    }
    if (!amount || isNaN(parseFloat(amount)) || parseFloat(amount) <= 0) {
      toast.error("Enter a valid amount");
      return;
    }
    setSaving(true);
    try {
      let fileId = "";
      let fileName = "";
      let fileUrl = "";
      if (file) {
        const result = await uploadToImageKit(file, "requisitions/quotations");
        fileId = result.fileId;
        fileName = result.name;
        fileUrl = result.url;
      }
      const newQuotation: Quotation = {
        vendorId,
        vendorName: vendorName.trim(),
        amount: parseFloat(amount),
        currency,
        fileId,
        fileName,
        fileUrl,
        uploadedAt: new Date().toISOString(),
      };
      await onSave([...quotations, newQuotation]);
      reset();
      toast.success("Quotation added");
    } catch {
      toast.error("Failed to add quotation");
    } finally {
      setSaving(false);
    }
  }

  function handleVendorSelect(id: string) {
    setVendorId(id);
    const vendor = vendors.find((v) => v.id === id);
    if (vendor) setVendorName(vendor.name);
  }

  return (
    <div className="mt-8 pt-6 border-t space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <h3 className="text-base font-semibold">Quotations</h3>
          <Badge
            className={`text-xs px-2 py-0.5 ${
              hasEnough
                ? "bg-green-100 text-green-800 border-green-200"
                : "bg-amber-100 text-amber-800 border-amber-200"
            }`}
          >
            {count}/3
          </Badge>
          {!hasEnough && (
            <span className="text-xs text-amber-600">
              {3 - count} more required
            </span>
          )}
        </div>
        {canEdit && !showForm && (
          <Button
            variant="outline"
            size="sm"
            className="gap-2"
            onClick={() => setShowForm(true)}
          >
            <Plus className="h-4 w-4" />
            Add Quotation
          </Button>
        )}
      </div>

      {/* Quotation table */}
      {quotations.length > 0 && (
        <div className="rounded-lg border overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-muted/40">
              <tr>
                <th className="text-left p-3 font-medium">Vendor</th>
                <th className="text-right p-3 font-medium">Amount</th>
                <th className="text-left p-3 font-medium hidden sm:table-cell">Date</th>
                <th className="p-3 font-medium">Quote</th>
              </tr>
            </thead>
            <tbody>
              {quotations.map((q, i) => (
                <tr key={`${q.vendorId}-${i}`} className="border-t">
                  <td className="p-3 font-medium">{q.vendorName}</td>
                  <td className="p-3 text-right font-mono">
                    {q.currency || currency}{" "}
                    {q.amount.toLocaleString("en-ZM", {
                      minimumFractionDigits: 2,
                      maximumFractionDigits: 2,
                    })}
                  </td>
                  <td className="p-3 text-muted-foreground text-xs hidden sm:table-cell">
                    {new Date(q.uploadedAt).toLocaleDateString("en-ZM", {
                      year: "numeric",
                      month: "short",
                      day: "numeric",
                    })}
                  </td>
                  <td className="p-3 text-center">
                    {q.fileUrl ? (
                      <a
                        href={q.fileUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="inline-flex items-center gap-1 text-xs text-blue-600 hover:underline"
                      >
                        <FileText className="h-3.5 w-3.5" />
                        <ExternalLink className="h-3 w-3" />
                      </a>
                    ) : (
                      <span className="text-xs text-muted-foreground">—</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Add quotation form */}
      {showForm && (
        <div className="rounded-lg border p-4 space-y-3 bg-muted/20">
          <div className="flex items-center justify-between">
            <p className="text-sm font-medium">New Quotation</p>
            <button
              type="button"
              onClick={reset}
              className="text-muted-foreground hover:text-foreground"
            >
              <X className="h-4 w-4" />
            </button>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            {vendors.length > 0 ? (
              <div className="space-y-1.5">
                <Label>Vendor</Label>
                <select
                  className="w-full h-9 rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm focus:outline-none focus:ring-1 focus:ring-ring"
                  value={vendorId}
                  onChange={(e) => handleVendorSelect(e.target.value)}
                >
                  <option value="">Select vendor...</option>
                  {vendors.map((v) => (
                    <option key={v.id} value={v.id}>
                      {v.name}
                    </option>
                  ))}
                </select>
                {!vendorId && (
                  <Input
                    placeholder="Or type vendor name"
                    value={vendorName}
                    onChange={(e) => setVendorName(e.target.value)}
                    className="mt-1.5 text-sm"
                  />
                )}
              </div>
            ) : (
              <div className="space-y-1.5">
                <Label>Vendor Name</Label>
                <Input
                  placeholder="Vendor name"
                  value={vendorName}
                  onChange={(e) => setVendorName(e.target.value)}
                />
              </div>
            )}

            <div className="space-y-1.5">
              <Label>Quoted Amount ({currency})</Label>
              <Input
                type="number"
                placeholder="0.00"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                min="0"
                step="0.01"
              />
            </div>
          </div>

          <div className="space-y-1.5">
            <Label>Quote Document (optional)</Label>
            <div className="flex items-center gap-2">
              <input
                ref={fileInputRef}
                type="file"
                className="hidden"
                accept="application/pdf,image/*"
                onChange={(e) => setFile(e.target.files?.[0] ?? null)}
              />
              <Button
                type="button"
                variant="outline"
                size="sm"
                className="gap-2"
                onClick={() => fileInputRef.current?.click()}
              >
                <Upload className="h-4 w-4" />
                {file ? file.name : "Choose file"}
              </Button>
              {file && (
                <button
                  type="button"
                  className="text-muted-foreground hover:text-destructive"
                  onClick={() => {
                    setFile(null);
                    if (fileInputRef.current) fileInputRef.current.value = "";
                  }}
                >
                  <X className="h-4 w-4" />
                </button>
              )}
            </div>
          </div>

          <div className="flex justify-end gap-2 pt-1">
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={reset}
              disabled={saving}
            >
              Cancel
            </Button>
            <Button
              type="button"
              size="sm"
              disabled={saving}
              isLoading={saving}
              onClick={handleAdd}
            >
              Add Quotation
            </Button>
          </div>
        </div>
      )}

      {quotations.length === 0 && !showForm && (
        <p className="text-sm text-muted-foreground">
          No quotations added yet.
          {canEdit && ' Click "Add Quotation" to begin collecting vendor quotes.'}
        </p>
      )}
    </div>
  );
}
