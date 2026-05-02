import { renderHook, act } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { useDateRangeUrlState } from "@/hooks/use-date-range-url-state";

const replace = vi.fn();
const params = new URLSearchParams();
const mockRouter = { replace };

vi.mock("next/navigation", () => ({
  useRouter: () => mockRouter,
  usePathname: () => "/admin/reports",
  useSearchParams: () => params,
}));

beforeEach(() => {
  replace.mockClear();
  params.delete("from");
  params.delete("to");
});

describe("useDateRangeUrlState", () => {
  it("returns the default range when no URL params present", () => {
    const { result } = renderHook(() =>
      useDateRangeUrlState({ defaultFrom: "2026-01-01", defaultTo: "2026-01-31" })
    );
    expect(result.current.from).toBe("2026-01-01");
    expect(result.current.to).toBe("2026-01-31");
  });

  it("reads URL params when present", () => {
    params.set("from", "2026-03-01");
    params.set("to", "2026-03-31");
    const { result } = renderHook(() =>
      useDateRangeUrlState({ defaultFrom: "2026-01-01", defaultTo: "2026-01-31" })
    );
    expect(result.current.from).toBe("2026-03-01");
    expect(result.current.to).toBe("2026-03-31");
  });

  it("setRange triggers router.replace with new params", () => {
    const { result } = renderHook(() =>
      useDateRangeUrlState({ defaultFrom: "2026-01-01", defaultTo: "2026-01-31" })
    );
    act(() => result.current.setRange("2026-04-01", "2026-04-30"));
    expect(replace).toHaveBeenCalledTimes(1);
    expect(replace.mock.calls[0][0]).toContain("from=2026-04-01");
    expect(replace.mock.calls[0][0]).toContain("to=2026-04-30");
  });

  it("setRange does NOT call router.replace when from/to are unchanged", () => {
    const { result } = renderHook(() =>
      useDateRangeUrlState({ defaultFrom: "2026-01-01", defaultTo: "2026-01-31" })
    );
    act(() => result.current.setRange("2026-01-01", "2026-01-31"));
    expect(replace).not.toHaveBeenCalled();
  });

  it("setRange identity is stable across renders when params are unchanged", () => {
    const { result, rerender } = renderHook(() =>
      useDateRangeUrlState({ defaultFrom: "2026-01-01", defaultTo: "2026-01-31" })
    );
    const first = result.current.setRange;
    rerender();
    expect(result.current.setRange).toBe(first);
  });
});
