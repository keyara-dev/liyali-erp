import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { StatGrid } from "@/components/ui/stat-grid";
import { Clock } from "lucide-react";

describe("StatGrid", () => {
  it("renders all stat cells with labels and values", () => {
    render(
      <StatGrid
        items={[
          { label: "Pending", value: 7, icon: <Clock data-testid="icon" />, accent: "amber" },
          { label: "Done", value: 3, icon: <Clock />, accent: "emerald" },
        ]}
      />
    );
    expect(screen.getByText("Pending")).toBeInTheDocument();
    expect(screen.getByText("7")).toBeInTheDocument();
    expect(screen.getByText("Done")).toBeInTheDocument();
    expect(screen.getByText("3")).toBeInTheDocument();
  });

  it("renders secondary text when provided", () => {
    render(
      <StatGrid
        items={[
          { label: "X", value: 1, icon: <Clock />, accent: "blue", secondary: "extra" },
        ]}
      />
    );
    expect(screen.getByText("extra")).toBeInTheDocument();
  });
});
