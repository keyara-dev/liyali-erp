import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { DocumentTypeChip } from "@/components/ui/document-type-chip";

describe("DocumentTypeChip", () => {
  it("renders Requisition for 'requisition'", () => {
    render(<DocumentTypeChip type="requisition" />);
    expect(screen.getByText("Requisition")).toBeInTheDocument();
  });

  it("normalizes case (uppercase input maps to same label)", () => {
    render(<DocumentTypeChip type="PURCHASE_ORDER" />);
    expect(screen.getByText("Purchase Order")).toBeInTheDocument();
  });

  it("falls back to Title Case for unknown types", () => {
    render(<DocumentTypeChip type="custom_doc" />);
    expect(screen.getByText("Custom Doc")).toBeInTheDocument();
  });
});
