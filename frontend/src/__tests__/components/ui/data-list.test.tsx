import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { DataList } from "@/components/ui/data-list";

interface Row { id: string; name: string; }

describe("DataList", () => {
  const rows: Row[] = [
    { id: "1", name: "alpha" },
    { id: "2", name: "beta" },
  ];

  it("renders rows in table mode", () => {
    render(
      <DataList<Row>
        rows={rows}
        getRowId={(r) => r.id}
        columns={[{ id: "name", header: "Name", cell: (r) => r.name }]}
        mobileCard={(r) => <div>{r.name}</div>}
      />
    );
    // Both table (desktop) and card (mobile) branches render in jsdom — use getAllByText
    expect(screen.getAllByText("alpha").length).toBeGreaterThan(0);
    expect(screen.getAllByText("beta").length).toBeGreaterThan(0);
  });

  it("renders empty state", () => {
    render(
      <DataList<Row>
        rows={[]}
        getRowId={(r) => r.id}
        columns={[{ id: "name", header: "Name", cell: (r) => r.name }]}
        mobileCard={(r) => <div>{r.name}</div>}
        emptyMessage="No rows."
      />
    );
    expect(screen.getByText("No rows.")).toBeInTheDocument();
  });
});
