import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { MetricCard } from "@/components/ui/metric-card";
import { FileText } from "lucide-react";

describe("MetricCard", () => {
  it("renders title, value, and icon", () => {
    render(
      <MetricCard
        title="Total Documents"
        value={42}
        icon={<FileText data-testid="metric-icon" />}
      />
    );
    expect(screen.getByText("Total Documents")).toBeInTheDocument();
    expect(screen.getByText("42")).toBeInTheDocument();
    expect(screen.getByTestId("metric-icon")).toBeInTheDocument();
  });

  it("renders secondary text when provided", () => {
    render(
      <MetricCard
        title="Approval Rate"
        value="92.3%"
        icon={<FileText />}
        secondary="last 30 days"
      />
    );
    expect(screen.getByText("last 30 days")).toBeInTheDocument();
  });

  it("renders TrendDelta when trend is provided", () => {
    render(
      <MetricCard
        title="Approvals"
        value={120}
        icon={<FileText />}
        trend={{ value: 8.5, label: "vs last week" }}
      />
    );
    expect(screen.getByTestId("trend-delta")).toBeInTheDocument();
    expect(screen.getByText(/8\.5%/)).toBeInTheDocument();
  });

  it("renders sparkline svg when sparkline data is provided", () => {
    const data = [1, 2, 3, 5, 8, 13, 21];
    render(
      <MetricCard
        title="Throughput"
        value={21}
        icon={<FileText />}
        sparkline={data}
      />
    );
    expect(screen.getByTestId("metric-sparkline")).toBeInTheDocument();
  });
});
