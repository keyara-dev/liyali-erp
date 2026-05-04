import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AdminReportsClient } from "@/app/(private)/admin/_components/admin-reports-client";

const replace = vi.fn();
const params = new URLSearchParams();

vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace }),
  usePathname: () => "/admin/reports",
  useSearchParams: () => params,
}));

vi.mock("@/hooks/use-reports-queries", () => ({
  useSystemStats: () => ({ data: undefined, isLoading: true, error: null }),
  useApprovalMetrics: () => ({ data: undefined, isLoading: true, error: null }),
  useUserActivity: () => ({ data: undefined, isLoading: true, error: null }),
  useAnalyticsDashboard: () => ({ data: undefined, isLoading: true, error: null }),
}));

vi.mock("@/components/workflows/analytics-dashboard", () => ({
  AnalyticsDashboard: ({ dateRange }: { dateRange?: { startDate?: string; endDate?: string } }) => (
    <div data-testid="analytics-dashboard">
      analytics:{dateRange?.startDate}–{dateRange?.endDate}
    </div>
  ),
}));

vi.mock("@/app/(private)/admin/_components/system-statistics", () => ({
  SystemStatistics: () => <div data-testid="system-statistics">System Statistics</div>,
}));

vi.mock("@/app/(private)/admin/_components/approval-reports", () => ({
  ApprovalReports: () => <div data-testid="approval-reports">Approval Reports</div>,
}));

vi.mock("@/app/(private)/admin/_components/user-activity-reports", () => ({
  UserActivityReports: () => <div data-testid="user-activity-reports">User Activity Reports</div>,
}));

vi.mock("@/lib/utils", async () => {
  const actual = await vi.importActual<typeof import("@/lib/utils")>("@/lib/utils");
  return { ...actual, notify: vi.fn() };
});

beforeEach(() => {
  replace.mockClear();
  params.delete("from");
  params.delete("to");
});

function renderWithClient(ui: React.ReactNode) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(<QueryClientProvider client={client}>{ui}</QueryClientProvider>);
}

describe("AdminReportsClient (integration)", () => {
  it("renders all 4 tabs with the correct titles", () => {
    renderWithClient(<AdminReportsClient userId="u1" userRole="admin" />);
    expect(screen.getByRole("tab", { name: /overview/i })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: /analytics/i })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: /approvals/i })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: /activity/i })).toBeInTheDocument();
  });

  it("switches active tab to Analytics on click", async () => {
    const user = userEvent.setup();
    renderWithClient(<AdminReportsClient userId="u1" userRole="admin" />);
    await user.click(screen.getByRole("tab", { name: /analytics/i }));
    expect(screen.getByTestId("analytics-dashboard")).toBeInTheDocument();
  });

  it("renders the page title and subtitle", () => {
    renderWithClient(<AdminReportsClient userId="u1" userRole="admin" />);
    expect(screen.getByText("Admin Reports")).toBeInTheDocument();
    expect(
      screen.getByText("Workflow approvals, user activity, system metrics")
    ).toBeInTheDocument();
  });
});
