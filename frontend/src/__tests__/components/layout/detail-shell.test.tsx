import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { DetailShell } from "@/components/layout/detail-shell";

describe("DetailShell", () => {
  it("renders header, main, and sidebar slots", () => {
    render(
      <DetailShell
        header={<div>HeaderContent</div>}
        sidebar={<div>SidebarContent</div>}
      >
        <div>MainContent</div>
      </DetailShell>
    );
    expect(screen.getByText("HeaderContent")).toBeInTheDocument();
    expect(screen.getByText("MainContent")).toBeInTheDocument();
    expect(screen.getByText("SidebarContent")).toBeInTheDocument();
  });

  it("renders without sidebar when not provided", () => {
    render(
      <DetailShell header={<div>H</div>}>
        <div>OnlyMain</div>
      </DetailShell>
    );
    expect(screen.getByText("OnlyMain")).toBeInTheDocument();
  });
});
