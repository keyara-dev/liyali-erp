import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi } from "vitest";
import { ReportsHeader } from "@/app/(private)/admin/_components/reports-header";

const baseProps = {
  title: "Admin Reports",
  subtitle: "Workflow approvals, user activity, system metrics",
  from: "2026-01-01",
  to: "2026-01-31",
  onRangeChange: vi.fn(),
  onRefresh: vi.fn(),
  onExport: vi.fn(),
  isRefreshing: false,
};

describe("ReportsHeader", () => {
  it("renders title and subtitle", () => {
    render(<ReportsHeader {...baseProps} />);
    expect(screen.getByText("Admin Reports")).toBeInTheDocument();
    expect(
      screen.getByText("Workflow approvals, user activity, system metrics")
    ).toBeInTheDocument();
  });

  it("calls onRefresh when refresh button is clicked", async () => {
    const onRefresh = vi.fn();
    const user = userEvent.setup();
    render(<ReportsHeader {...baseProps} onRefresh={onRefresh} />);
    await user.click(screen.getByRole("button", { name: /refresh/i }));
    expect(onRefresh).toHaveBeenCalledTimes(1);
  });

  it("disables refresh when isRefreshing is true", () => {
    render(<ReportsHeader {...baseProps} isRefreshing />);
    expect(screen.getByRole("button", { name: /refresh/i })).toBeDisabled();
  });

  it("invokes onExport with chosen format from menu", async () => {
    const onExport = vi.fn();
    const user = userEvent.setup();
    render(<ReportsHeader {...baseProps} onExport={onExport} />);
    await user.click(screen.getByRole("button", { name: /export/i }));
    await user.click(await screen.findByRole("menuitem", { name: /csv/i }));
    expect(onExport).toHaveBeenCalledWith("csv");
  });
});
