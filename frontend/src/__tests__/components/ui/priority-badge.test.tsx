import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { PriorityBadge } from "@/components/ui/priority-badge";

describe("PriorityBadge", () => {
  it("renders the priority text", () => {
    render(<PriorityBadge priority="HIGH" />);
    expect(screen.getByText("HIGH")).toBeInTheDocument();
  });

  it("falls back to MEDIUM when priority is undefined", () => {
    render(<PriorityBadge priority={undefined} />);
    expect(screen.getByText("MEDIUM")).toBeInTheDocument();
  });

  it("normalizes case for variant lookup", () => {
    render(<PriorityBadge priority="urgent" />);
    expect(screen.getByText("urgent")).toBeInTheDocument();
  });
});
