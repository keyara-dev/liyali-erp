"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { SelectField } from "@/components/ui/select-field";
import { Badge } from "@/components";
import { FileText, Plus, ExternalLink, X } from "lucide-react";
import FileUpload from "@/components/base/file-upload";
import { Quotation } from "@/types/core";
import { Vendor } from "@/types/vendor";
import { uploadToImageKit } from "@/lib/imagekit";
import { toast } from "sonner";
import { cn, formatCurrency } from "@/lib/utils";

interface QuotationCollectionSectionProps {
  quotations: Quotation[];
  requisitionId: string;
  currency: string;
  vendors: Vendor[];
  canEdit: boolean;
  onSave: (quotations: Quotation[]) => Promise<void>;
  /** Optional: Current selected vendor ID for the PO */
  selectedVendorId?: string;
  /** Optional: Amount of the selected quotation — disambiguates duplicate vendors */
  selectedVendorAmount?: number;
  /** Optional: fileUrl of the selected quotation — uniquely identifies which row is active */
  selectedQuotationFileId?: string;
  /** Optional: Callback when a vendor is selected from quotations */
  onSelectVendor?: (
    vendorId: string,
    vendorName: string,
    amount: number,
    fileUrl: string,
  ) => Promise<void>;
  /** Optional: Show vendor selection UI (for PO pages) */
  showVendorSelection?: boolean;
}

export function QuotationCollectionSection({
  quotations,
  currency,
  vendors,
  canEdit,
  onSave,
  selectedVendorId,
  selectedVendorAmount,
  selectedQuotationFileId,
  onSelectVendor,
  showVendorSelection = false,
}: QuotationCollectionSectionProps) {
  const [showForm, setShowForm] = useState(false);
  const [saving, setSaving] = useState(false);
  const [vendorId, setVendorId] = useState("");
  const [vendorName, setVendorName] = useState("");
  const [amount, setAmount] = useState("");
  const [file, setFile] = useState<File | null>(null);
  const [fileKey, setFileKey] = useState(0);
  const [selectingVendor, setSelectingVendor] = useState(false);

  const count = quotations.length;
  const hasEnough = count >= 3;

  function reset() {
    setVendorId("");
    setVendorName("");
    setAmount("");
    setFile(null);
    setFileKey((k) => k + 1);
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

  async function handleSelectQuotationVendor(
    quotationVendorId: string,
    quotationVendorName: string,
    quotationAmount: number,
    quotationFileUrl: string,
  ) {
    if (!onSelectVendor) return;
    setSelectingVendor(true);
    try {
      await onSelectVendor(
        quotationVendorId,
        quotationVendorName,
        quotationAmount,
        quotationFileUrl,
      );
      toast.success(`Selected ${quotationVendorName} as vendor`);
    } catch (error) {
      toast.error("Failed to select vendor");
    } finally {
      setSelectingVendor(false);
    }
  }

  return (
    <div className="mt-8 pt-6 border-t space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <h3 className="text-base font-semibold">Quotations</h3>
          <Badge
            className={`text-xs px-2 py-0.5 ${
              hasEnough
                ? "bg-green-100 text-green-800 border-green-200 dark:bg-green-900 dark:text-green-100 dark:border-green-800"
                : "bg-amber-100 text-amber-800 border-amber-200 dark:bg-amber-900 dark:text-amber-100 dark:border-amber-800"
            }`}
          >
            {count}/3
          </Badge>
          {!hasEnough && (
            <span className="text-xs text-amber-600 dark:text-amber-400">
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

          <div
            className={cn(
              "grid grid-cols-1 place-items-end sm:grid-cols-3 gap-3",
              {
                "sm:grid-cols-2": vendorId,
              },
            )}
          >
            {vendors.length > 0 ? (
              <>
                <SelectField
                  label="Vendor"
                  value={vendorId || "__none__"}
                  onValueChange={(v) =>
                    handleVendorSelect(v === "__none__" ? "" : v)
                  }
                  placeholder="Select vendor..."
                  options={[
                    { value: "__none__", label: "None (enter manually)" },
                    ...vendors.map((v) => ({ value: v.id, label: v.name })),
                  ]}
                />
                {!vendorId && (
                  <Input
                    placeholder="Or type vendor name"
                    value={vendorName}
                    onChange={(e) => setVendorName(e.target.value)}
                    className="mt-1.5 text-sm"
                  />
                )}
              </>
            ) : (
              <Input
                label="Vendor Name"
                placeholder="Vendor name"
                value={vendorName}
                onChange={(e) => setVendorName(e.target.value)}
              />
            )}

            <Input
              label={`Quoted Amount (${currency})`}
              type="number"
              placeholder="0.00"
              value={amount}
              onChange={(e) => setAmount(e.target.value)}
              min="0"
              step="0.01"
            />
          </div>

          <FileUpload
            key={fileKey}
            id="quotation-file"
            label="Quote Document (optional)"
            accept=".pdf,.docx,.jpg,.jpeg,.png,.gif,.webp,.bmp"
            maxFileSize={10}
            compact
            onFileChange={setFile}
          />

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

      {/* Quotation table */}
      {quotations.length > 0 && (
        <div className="rounded-lg border overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-muted/40">
              <tr>
                <th className="text-left p-3 font-medium">Vendor</th>
                <th className="text-right p-3 font-medium">Amount</th>
                <th className="text-left p-3 font-medium hidden sm:table-cell">
                  Date
                </th>
                <th className="p-3 font-medium">Quote</th>
                {showVendorSelection && onSelectVendor && (
                  <th className="p-3 font-medium">Action</th>
                )}
              </tr>
            </thead>
            <tbody>
              {quotations.map((q, i) => {
                // Primary: match by fileUrl (unique identifier)
                // Fallback: match by vendorId when no fileUrl is stored yet
                const isSelected =
                  showVendorSelection &&
                  (selectedQuotationFileId
                    ? q.fileUrl === selectedQuotationFileId
                    : !!selectedVendorId && q.vendorId === selectedVendorId);
                return (
                  <tr
                    key={`${q.vendorId}-${i}`}
                    className={cn("border-t", {
                      "bg-green-50 dark:bg-green-950/30": isSelected,
                    })}
                  >
                    <td className="p-3 font-medium">
                      <div className="flex items-center gap-2">
                        {q.vendorName}
                        {isSelected && (
                          <Badge className="bg-green-100 text-green-800 border-green-200 dark:bg-green-900 dark:text-green-100 dark:border-green-800 text-xs">
                            Selected
                          </Badge>
                        )}
                      </div>
                    </td>
                    <td className="p-3 text-right font-mono">
                      {/* {q.currency || currency}{" "}
                      {q.amount.toLocaleString("en-ZM", {
                        minimumFractionDigits: 2,
                        maximumFractionDigits: 2,
                      })} */}
                      {formatCurrency(q.amount, q.currency)}
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
                    {showVendorSelection && onSelectVendor && (
                      <td className="p-3 text-center">
                        {!isSelected ? (
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={() =>
                              handleSelectQuotationVendor(
                                q.vendorId,
                                q.vendorName,
                                q.amount,
                                q.fileUrl,
                              )
                            }
                            disabled={selectingVendor}
                            className="text-xs h-7"
                          >
                            Select
                          </Button>
                        ) : (
                          <span className="text-xs text-green-600 dark:text-green-400 font-medium">
                            ✓ Active
                          </span>
                        )}
                      </td>
                    )}
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      {quotations.length === 0 && !showForm && (
        <p className="text-sm text-muted-foreground">
          No quotations added yet.
          {canEdit &&
            ' Click "Add Quotation" to begin collecting vendor quotes.'}
        </p>
      )}
    </div>
  );
}
