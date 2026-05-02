import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import EmptyState from "@/components/base/empty-state";

describe("EmptyState", () => {
  it("renders title and description", () => {
    render(<EmptyState title="Nothing here" description="Try a different filter" />);
    expect(screen.getByText("Nothing here")).toBeInTheDocument();
    expect(screen.getByText("Try a different filter")).toBeInTheDocument();
  });

  it("renders an action when provided", () => {
    render(
      <EmptyState
        title="No tasks"
        description="You're all caught up"
        action={<button>Refresh</button>}
      />
    );
    expect(screen.getByRole("button", { name: /refresh/i })).toBeInTheDocument();
  });

  it("uses muted-foreground color for description (no typo classes)", () => {
    const { container } = render(
      <EmptyState title="t" description="d" />
    );
    expect(container.innerHTML).not.toContain("text-gary-400");
  });
});
