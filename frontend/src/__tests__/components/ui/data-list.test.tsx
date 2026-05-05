import { render, screen, within, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { DataList } from "@/components/ui/data-list";

interface Row {
  id: string;
  name: string;
}

describe("DataList", () => {
  const rows: Row[] = [
    { id: "1", name: "alpha" },
    { id: "2", name: "beta" },
  ];

  const baseColumns = [
    { id: "name", header: "Name", cell: (r: Row) => r.name },
  ];

  it("renders rows in table mode", () => {
    render(
      <DataList<Row>
        rows={rows}
        getRowId={(r) => r.id}
        columns={baseColumns}
        mobileCard={(r) => <div>{r.name}</div>}
      />
    );
    // Both table (desktop) and card (mobile) branches render in jsdom.
    // "alpha" appears once in the <td> and once in the mobile card div → exactly 2.
    expect(screen.getAllByText("alpha").length).toBe(2);
    expect(screen.getAllByText("beta").length).toBe(2);

    // Table branch assertions using scoped query
    const table = screen.getByRole("table");
    expect(within(table).getByText("alpha")).toBeInTheDocument();
    expect(within(table).getByText("beta")).toBeInTheDocument();
  });

  it("renders empty state", () => {
    render(
      <DataList<Row>
        rows={[]}
        getRowId={(r) => r.id}
        columns={baseColumns}
        mobileCard={(r) => <div>{r.name}</div>}
        emptyMessage="No rows."
      />
    );
    expect(screen.getByText("No rows.")).toBeInTheDocument();
  });

  it("does not invoke onRowClick when click originates from a button child", () => {
    const onRowClick = vi.fn();
    const columns = [
      {
        id: "action",
        header: "Action",
        cell: (_r: Row) => <button type="button">Action</button>,
      },
    ];

    const { container } = render(
      <DataList<Row>
        rows={rows}
        getRowId={(r) => r.id}
        columns={columns}
        mobileCard={(r) => <div>{r.name}</div>}
        onRowClick={onRowClick}
      />
    );

    // Use querySelector to get a native <button> element inside a table cell,
    // avoiding the <TR role="button"> rows which getAllByRole also matches.
    const actionButton = container.querySelector("td button") as HTMLElement;
    expect(actionButton).toBeTruthy();
    fireEvent.click(actionButton);

    expect(onRowClick).not.toHaveBeenCalled();
  });

  it("invokes onRowClick on Enter keydown when set", () => {
    const onRowClick = vi.fn();

    const { container } = render(
      <DataList<Row>
        rows={rows}
        getRowId={(r) => r.id}
        columns={baseColumns}
        mobileCard={(r) => <div>{r.name}</div>}
        onRowClick={onRowClick}
      />
    );

    // Target a mobile card div (role="button") directly — simpler and reliable in jsdom.
    // The mobile stack renders divs with role="button" and tabIndex=0.
    const mobileCards = container.querySelectorAll(
      "div.md\\:hidden [role='button']"
    );
    expect(mobileCards.length).toBeGreaterThan(0);
    fireEvent.keyDown(mobileCards[0], { key: "Enter" });

    expect(onRowClick).toHaveBeenCalledTimes(1);
    expect(onRowClick).toHaveBeenCalledWith(rows[0]);
  });

  it("renders loading skeleton when isLoading is true", () => {
    const { container } = render(
      <DataList<Row>
        rows={[]}
        getRowId={(r) => r.id}
        columns={baseColumns}
        mobileCard={(r) => <div>{r.name}</div>}
        isLoading={true}
      />
    );

    // Skeleton uses animate-pulse
    const skeletonEls = container.querySelectorAll('[class*="animate-pulse"]');
    expect(skeletonEls.length).toBeGreaterThan(0);
  });

  it("applies align class to header and cell when align is set", () => {
    const { container } = render(
      <DataList<Row>
        rows={[{ id: "1", name: "alpha" }]}
        getRowId={(r) => r.id}
        columns={[{ id: "name", header: "Name", cell: (r) => r.name, align: "right" }]}
        mobileCard={(r) => <div>{r.name}</div>}
      />
    );
    const ths = container.querySelectorAll("th");
    const tds = container.querySelectorAll("td");
    expect(Array.from(ths).some((el) => el.className.includes("text-right"))).toBe(true);
    expect(Array.from(tds).some((el) => el.className.includes("text-right"))).toBe(true);
  });
});
