import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { ReportsHeader } from "@/app/(private)/admin/_components/reports-header";

vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace: vi.fn() }),
  usePathname: () => "/admin/reports",
  useSearchParams: () => new URLSearchParams(),
}));

describe("ReportsHeader", () => {
  it("renders title and subtitle", () => {
    render(
      <ReportsHeader
        title="Admin Reports"
        subtitle="Workflow approvals, user activity, system metrics"
        onRefresh={vi.fn()}
        onExport={vi.fn()}
        isRefreshing={false}
      />
    );
    expect(screen.getByText("Admin Reports")).toBeInTheDocument();
    expect(
      screen.getByText("Workflow approvals, user activity, system metrics")
    ).toBeInTheDocument();
  });

  it("calls onRefresh when refresh button is clicked", () => {
    const onRefresh = vi.fn();
    render(
      <ReportsHeader
        title="X"
        subtitle="Y"
        onRefresh={onRefresh}
        onExport={vi.fn()}
        isRefreshing={false}
      />
    );
    fireEvent.click(screen.getByRole("button", { name: /refresh/i }));
    expect(onRefresh).toHaveBeenCalledTimes(1);
  });

  it("disables refresh when isRefreshing is true", () => {
    render(
      <ReportsHeader
        title="X"
        subtitle="Y"
        onRefresh={vi.fn()}
        onExport={vi.fn()}
        isRefreshing
      />
    );
    expect(screen.getByRole("button", { name: /refresh/i })).toBeDisabled();
  });

  it("invokes onExport with chosen format from menu", async () => {
    const onExport = vi.fn();
    const { default: userEvent } = await import("@testing-library/user-event");
    const user = userEvent.setup();
    render(
      <ReportsHeader
        title="X"
        subtitle="Y"
        onRefresh={vi.fn()}
        onExport={onExport}
        isRefreshing={false}
      />
    );
    await user.click(screen.getByRole("button", { name: /export/i }));
    await user.click(await screen.findByText(/csv/i));
    expect(onExport).toHaveBeenCalledWith("csv");
  });
});
