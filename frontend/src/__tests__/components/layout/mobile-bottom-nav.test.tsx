import { render, screen } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { MobileBottomNav } from "@/components/layout/mobile-bottom-nav";

vi.mock("next/navigation", () => ({
  usePathname: () => "/home",
  useRouter: () => ({ push: vi.fn() }),
}));

describe("MobileBottomNav", () => {
  it("renders 4 primary tabs", () => {
    render(<MobileBottomNav />);
    expect(screen.getByRole("link", { name: /home/i })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /tasks/i })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /documents/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /more/i })).toBeInTheDocument();
  });
});
