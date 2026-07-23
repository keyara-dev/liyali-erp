import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { ApprovalChainStepper } from "@/components/workflows/approval-chain-stepper";

const stages = [
  { id: "s1", name: "Department Head", status: "approved" as const, actor: "Jane Doe", at: "2026-04-30T10:00:00Z" },
  { id: "s2", name: "Finance", status: "current" as const },
  { id: "s3", name: "CEO", status: "pending" as const },
];

describe("ApprovalChainStepper", () => {
  it("renders all stage names", () => {
    render(<ApprovalChainStepper stages={stages} />);
    expect(screen.getByText("Department Head")).toBeInTheDocument();
    expect(screen.getByText("Finance")).toBeInTheDocument();
    expect(screen.getByText("CEO")).toBeInTheDocument();
  });

  it("renders the actor name on completed stages", () => {
    render(<ApprovalChainStepper stages={stages} />);
    expect(screen.getByText(/Jane Doe/)).toBeInTheDocument();
  });

  it("marks the current stage with aria-current=step", () => {
    render(<ApprovalChainStepper stages={stages} />);
    const current = screen.getByText("Finance").closest("[aria-current]");
    expect(current).toHaveAttribute("aria-current", "step");
  });

  it("renders a rejected status with proper data attribute", () => {
    render(
      <ApprovalChainStepper
        stages={[{ id: "x", name: "Finance", status: "rejected", actor: "Bob", at: "2026-04-29" }]}
      />
    );
    expect(screen.getByTestId("stage-marker-x")).toHaveAttribute("data-status", "rejected");
  });
});
