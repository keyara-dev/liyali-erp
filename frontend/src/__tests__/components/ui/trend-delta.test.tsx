import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { TrendDelta } from "@/components/ui/trend-delta";

describe("TrendDelta", () => {
  it("shows positive percentage with up arrow when value > 0", () => {
    render(<TrendDelta value={12.34} />);
    expect(screen.getByText(/12\.3%/)).toBeInTheDocument();
    expect(screen.getByTestId("trend-arrow")).toHaveAttribute("data-direction", "up");
  });

  it("shows negative percentage with down arrow when value < 0", () => {
    render(<TrendDelta value={-5.6} />);
    expect(screen.getByText(/5\.6%/)).toBeInTheDocument();
    expect(screen.getByTestId("trend-arrow")).toHaveAttribute("data-direction", "down");
  });

  it("renders a flat indicator when value === 0", () => {
    render(<TrendDelta value={0} />);
    expect(screen.getByText(/0\.0%/)).toBeInTheDocument();
    expect(screen.getByTestId("trend-arrow")).toHaveAttribute("data-direction", "flat");
  });

  it("renders the comparison label when provided", () => {
    render(<TrendDelta value={3} label="vs last week" />);
    expect(screen.getByText(/vs last week/)).toBeInTheDocument();
  });

  it("inverts up/down semantics when invert is set (e.g. lower is better)", () => {
    render(<TrendDelta value={5} invert />);
    const root = screen.getByTestId("trend-delta");
    expect(root).toHaveAttribute("data-tone", "negative");
  });
});
