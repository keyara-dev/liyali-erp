# Component Design System

Comprehensive guide to the UI component architecture, design patterns, and component library used in the Liyali Gateway Frontend.

## Design System Overview

The Liyali Gateway Frontend uses a **layered component architecture** built on top of shadcn/ui and Radix UI primitives, providing a consistent, accessible, and maintainable design system.

```
┌─────────────────────────────────────────────────────────────────┐
│                    Application Components                       │
├─────────────────────────────────────────────────────────────────┤
│  Feature Components - Domain-specific components               │
├─────────────────────────────────────────────────────────────────┤
│  Composite Components - Complex UI patterns                    │
├─────────────────────────────────────────────────────────────────┤
│  shadcn/ui Components - Styled component library               │
├─────────────────────────────────────────────────────────────────┤
│  Radix UI Primitives - Accessible, unstyled components         │
├─────────────────────────────────────────────────────────────────┤
│  Tailwind CSS - Utility-first styling system                  │
└─────────────────────────────────────────────────────────────────┘
```

## Component Architecture Layers

### 1. Primitive Layer (Radix UI)
Accessible, unstyled components that provide the foundation:

```tsx
// Example: Radix Dialog primitive
import * as Dialog from '@radix-ui/react-dialog';

// Provides:
// - Keyboard navigation
// - Focus management
// - ARIA attributes
// - Screen reader support
```

### 2. Base UI Layer (shadcn/ui)
Styled components built on Radix primitives:

```tsx
// src/components/ui/button.tsx
import { Slot } from '@radix-ui/react-slot';
import { cva, type VariantProps } from 'class-variance-authority';

const buttonVariants = cva(
  "inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:opacity-50 disabled:pointer-events-none ring-offset-background",
  {
    variants: {
      variant: {
        default: "bg-primary text-primary-foreground hover:bg-primary/90",
        destructive: "bg-destructive text-destructive-foreground hover:bg-destructive/90",
        outline: "border border-input hover:bg-accent hover:text-accent-foreground",
        secondary: "bg-secondary text-secondary-foreground hover:bg-secondary/80",
        ghost: "hover:bg-accent hover:text-accent-foreground",
        link: "underline-offset-4 hover:underline text-primary",
      },
      size: {
        default: "h-10 py-2 px-4",
        sm: "h-9 px-3 rounded-md",
        lg: "h-11 px-8 rounded-md",
        icon: "h-10 w-10",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
);

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean;
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild = false, ...props }, ref) => {
    const Comp = asChild ? Slot : "button";
    return (
      <Comp
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        {...props}
      />
    );
  }
);
```

### 3. Composite Layer
Complex components built from base UI components:

```tsx
// src/components/ui/data-table.tsx
interface DataTableProps<TData, TValue> {
  columns: ColumnDef<TData, TValue>[];
  data: TData[];
  searchKey?: string;
  searchPlaceholder?: string;
}

export function DataTable<TData, TValue>({
  columns,
  data,
  searchKey,
  searchPlaceholder = "Search...",
}: DataTableProps<TData, TValue>) {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({});
  const [rowSelection, setRowSelection] = useState({});

  const table = useReactTable({
    data,
    columns,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    onRowSelectionChange: setRowSelection,
    state: {
      sorting,
      columnFilters,
      columnVisibility,
      rowSelection,
    },
  });

  return (
    <div className="space-y-4">
      {/* Search and filters */}
      <div className="flex items-center justify-between">
        {searchKey && (
          <Input
            placeholder={searchPlaceholder}
            value={(table.getColumn(searchKey)?.getFilterValue() as string) ?? ""}
            onChange={(event) =>
              table.getColumn(searchKey)?.setFilterValue(event.target.value)
            }
            className="max-w-sm"
          />
        )}
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" className="ml-auto">
              Columns <ChevronDown className="ml-2 h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            {table
              .getAllColumns()
              .filter((column) => column.getCanHide())
              .map((column) => {
                return (
                  <DropdownMenuCheckboxItem
                    key={column.id}
                    className="capitalize"
                    checked={column.getIsVisible()}
                    onCheckedChange={(value) =>
                      column.toggleVisibility(!!value)
                    }
                  >
                    {column.id}
                  </DropdownMenuCheckboxItem>
                );
              })}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      {/* Table */}
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead key={header.id}>
                    {header.isPlaceholder
                      ? null
                      : flexRender(
                          header.column.columnDef.header,
                          header.getContext()
                        )}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && "selected"}
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>
                      {flexRender(
                        cell.column.columnDef.cell,
                        cell.getContext()
                      )}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell
                  colSpan={columns.length}
                  className="h-24 text-center"
                >
                  No results.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {/* Pagination */}
      <div className="flex items-center justify-end space-x-2 py-4">
        <div className="flex-1 text-sm text-muted-foreground">
          {table.getFilteredSelectedRowModel().rows.length} of{" "}
          {table.getFilteredRowModel().rows.length} row(s) selected.
        </div>
        <div className="space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.previousPage()}
            disabled={!table.getCanPreviousPage()}
          >
            Previous
          </Button>
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.nextPage()}
            disabled={!table.getCanNextPage()}
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  );
}
```

### 4. Feature Layer
Domain-specific components for business logic:

```tsx
// src/components/workflows/approval-flow-display.tsx
interface ApprovalFlowDisplayProps {
  workflow: Workflow;
  currentStage?: string;
  onStageClick?: (stageId: string) => void;
}

export function ApprovalFlowDisplay({
  workflow,
  currentStage,
  onStageClick,
}: ApprovalFlowDisplayProps) {
  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold">{workflow.name}</h3>
        <Badge variant={workflow.isActive ? "default" : "secondary"}>
          {workflow.isActive ? "Active" : "Inactive"}
        </Badge>
      </div>

      <div className="space-y-2">
        {workflow.stages.map((stage, index) => (
          <div
            key={stage.id}
            className={cn(
              "flex items-center space-x-3 p-3 rounded-lg border transition-colors",
              currentStage === stage.id && "bg-primary/5 border-primary",
              onStageClick && "cursor-pointer hover:bg-muted/50"
            )}
            onClick={() => onStageClick?.(stage.id)}
          >
            <div className="flex-shrink-0">
              <div
                className={cn(
                  "w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium",
                  currentStage === stage.id
                    ? "bg-primary text-primary-foreground"
                    : "bg-muted text-muted-foreground"
                )}
              >
                {index + 1}
              </div>
            </div>

            <div className="flex-1 min-w-0">
              <div className="flex items-center justify-between">
                <h4 className="text-sm font-medium truncate">{stage.name}</h4>
                <div className="flex items-center space-x-2">
                  {stage.isOptional && (
                    <Badge variant="outline" className="text-xs">
                      Optional
                    </Badge>
                  )}
                  {stage.timeoutHours && (
                    <span className="text-xs text-muted-foreground">
                      {stage.timeoutHours}h timeout
                    </span>
                  )}
                </div>
              </div>
              {stage.description && (
                <p className="text-sm text-muted-foreground mt-1">
                  {stage.description}
                </p>
              )}
            </div>

            {index < workflow.stages.length - 1 && (
              <ChevronRight className="w-4 h-4 text-muted-foreground" />
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
```

## Component Patterns

### 1. Compound Components
For flexible, composable interfaces:

```tsx
// Compound component pattern
export function Card({ children, className, ...props }) {
  return (
    <div className={cn("rounded-lg border bg-card text-card-foreground shadow-sm", className)} {...props}>
      {children}
    </div>
  );
}

Card.Header = function CardHeader({ children, className, ...props }) {
  return (
    <div className={cn("flex flex-col space-y-1.5 p-6", className)} {...props}>
      {children}
    </div>
  );
};

Card.Title = function CardTitle({ children, className, ...props }) {
  return (
    <h3 className={cn("text-2xl font-semibold leading-none tracking-tight", className)} {...props}>
      {children}
    </h3>
  );
};

Card.Content = function CardContent({ children, className, ...props }) {
  return (
    <div className={cn("p-6 pt-0", className)} {...props}>
      {children}
    </div>
  );
};

// Usage
<Card>
  <Card.Header>
    <Card.Title>Requisitions</Card.Title>
  </Card.Header>
  <Card.Content>
    <RequisitionsList />
  </Card.Content>
</Card>
```

### 2. Render Props Pattern
For flexible data sharing:

```tsx
// Render props for data fetching
interface DataFetcherProps<T> {
  queryKey: string[];
  queryFn: () => Promise<T>;
  children: (props: {
    data: T | undefined;
    isLoading: boolean;
    error: Error | null;
  }) => React.ReactNode;
}

export function DataFetcher<T>({ queryKey, queryFn, children }: DataFetcherProps<T>) {
  const { data, isLoading, error } = useQuery({
    queryKey,
    queryFn,
  });

  return <>{children({ data, isLoading, error })}</>;
}

// Usage
<DataFetcher
  queryKey={['requisitions']}
  queryFn={() => apiClient.requisitions.getAll()}
>
  {({ data, isLoading, error }) => {
    if (isLoading) return <Skeleton />;
    if (error) return <ErrorDisplay error={error} />;
    return <RequisitionsList data={data} />;
  }}
</DataFetcher>
```

### 3. Higher-Order Components (HOCs)
For cross-cutting concerns:

```tsx
// HOC for permission checking
export function withPermissions<P extends object>(
  Component: React.ComponentType<P>,
  requiredPermissions: string[]
) {
  return function PermissionWrappedComponent(props: P) {
    const { hasPermissions } = usePermissions();
    
    if (!hasPermissions(requiredPermissions)) {
      return <AccessDenied />;
    }
    
    return <Component {...props} />;
  };
}

// Usage
const ProtectedRequisitionForm = withPermissions(
  RequisitionForm,
  ['requisitions.create']
);
```

### 4. Custom Hooks for Component Logic
Extracting reusable component logic:

```tsx
// Custom hook for table state
export function useDataTable<T>({
  data,
  columns,
  searchKey,
}: {
  data: T[];
  columns: ColumnDef<T>[];
  searchKey?: string;
}) {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({});
  const [rowSelection, setRowSelection] = useState({});

  const table = useReactTable({
    data,
    columns,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    onRowSelectionChange: setRowSelection,
    state: {
      sorting,
      columnFilters,
      columnVisibility,
      rowSelection,
    },
  });

  const searchValue = searchKey 
    ? (table.getColumn(searchKey)?.getFilterValue() as string) ?? ""
    : "";

  const setSearchValue = (value: string) => {
    if (searchKey) {
      table.getColumn(searchKey)?.setFilterValue(value);
    }
  };

  return {
    table,
    searchValue,
    setSearchValue,
    selectedRows: table.getFilteredSelectedRowModel().rows,
  };
}

// Usage in component
function RequisitionsTable({ data }: { data: Requisition[] }) {
  const { table, searchValue, setSearchValue, selectedRows } = useDataTable({
    data,
    columns: requisitionColumns,
    searchKey: "title",
  });

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <Input
          placeholder="Search requisitions..."
          value={searchValue}
          onChange={(e) => setSearchValue(e.target.value)}
          className="max-w-sm"
        />
        {selectedRows.length > 0 && (
          <BulkOperationsToolbar selectedItems={selectedRows} />
        )}
      </div>
      {/* Table rendering */}
    </div>
  );
}
```

## Styling System

### 1. Tailwind CSS Configuration
```js
// tailwind.config.js
module.exports = {
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        border: "hsl(var(--border))",
        input: "hsl(var(--input))",
        ring: "hsl(var(--ring))",
        background: "hsl(var(--background))",
        foreground: "hsl(var(--foreground))",
        primary: {
          DEFAULT: "hsl(var(--primary))",
          foreground: "hsl(var(--primary-foreground))",
        },
        secondary: {
          DEFAULT: "hsl(var(--secondary))",
          foreground: "hsl(var(--secondary-foreground))",
        },
        // ... more colors
      },
      borderRadius: {
        lg: "var(--radius)",
        md: "calc(var(--radius) - 2px)",
        sm: "calc(var(--radius) - 4px)",
      },
    },
  },
  plugins: [require("tailwindcss-animate")],
};
```

### 2. CSS Variables for Theming
```css
/* globals.css */
@layer base {
  :root {
    --background: 0 0% 100%;
    --foreground: 222.2 84% 4.9%;
    --card: 0 0% 100%;
    --card-foreground: 222.2 84% 4.9%;
    --popover: 0 0% 100%;
    --popover-foreground: 222.2 84% 4.9%;
    --primary: 222.2 47.4% 11.2%;
    --primary-foreground: 210 40% 98%;
    /* ... more variables */
  }

  .dark {
    --background: 222.2 84% 4.9%;
    --foreground: 210 40% 98%;
    --card: 222.2 84% 4.9%;
    --card-foreground: 210 40% 98%;
    /* ... dark theme variables */
  }
}
```

### 3. Component Variants with CVA
```tsx
// Class Variance Authority for component variants
import { cva } from "class-variance-authority";

const alertVariants = cva(
  "relative w-full rounded-lg border p-4 [&>svg~*]:pl-7 [&>svg+div]:translate-y-[-3px] [&>svg]:absolute [&>svg]:left-4 [&>svg]:top-4 [&>svg]:text-foreground",
  {
    variants: {
      variant: {
        default: "bg-background text-foreground",
        destructive:
          "border-destructive/50 text-destructive dark:border-destructive [&>svg]:text-destructive",
        warning:
          "border-warning/50 text-warning dark:border-warning [&>svg]:text-warning",
        success:
          "border-success/50 text-success dark:border-success [&>svg]:text-success",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
);

export interface AlertProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof alertVariants> {}

const Alert = React.forwardRef<HTMLDivElement, AlertProps>(
  ({ className, variant, ...props }, ref) => (
    <div
      ref={ref}
      role="alert"
      className={cn(alertVariants({ variant }), className)}
      {...props}
    />
  )
);
```

## Accessibility Features

### 1. ARIA Attributes
```tsx
// Proper ARIA labeling
export function SearchField({ onSearch, placeholder = "Search..." }) {
  const [query, setQuery] = useState("");
  const searchId = useId();

  return (
    <div className="relative">
      <label htmlFor={searchId} className="sr-only">
        {placeholder}
      </label>
      <Input
        id={searchId}
        type="search"
        placeholder={placeholder}
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === 'Enter') {
            onSearch(query);
          }
        }}
        aria-describedby={`${searchId}-description`}
      />
      <div id={`${searchId}-description`} className="sr-only">
        Press Enter to search, or type to filter results
      </div>
    </div>
  );
}
```

### 2. Keyboard Navigation
```tsx
// Keyboard navigation support
export function DropdownMenu({ items, onSelect }) {
  const [isOpen, setIsOpen] = useState(false);
  const [focusedIndex, setFocusedIndex] = useState(-1);

  const handleKeyDown = (e: KeyboardEvent) => {
    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault();
        setFocusedIndex(prev => 
          prev < items.length - 1 ? prev + 1 : 0
        );
        break;
      case 'ArrowUp':
        e.preventDefault();
        setFocusedIndex(prev => 
          prev > 0 ? prev - 1 : items.length - 1
        );
        break;
      case 'Enter':
        e.preventDefault();
        if (focusedIndex >= 0) {
          onSelect(items[focusedIndex]);
          setIsOpen(false);
        }
        break;
      case 'Escape':
        setIsOpen(false);
        break;
    }
  };

  return (
    <div className="relative" onKeyDown={handleKeyDown}>
      {/* Dropdown implementation */}
    </div>
  );
}
```

### 3. Screen Reader Support
```tsx
// Screen reader announcements
export function StatusAnnouncer({ message, priority = "polite" }) {
  return (
    <div
      role="status"
      aria-live={priority}
      aria-atomic="true"
      className="sr-only"
    >
      {message}
    </div>
  );
}

// Usage for dynamic updates
function RequisitionForm() {
  const [status, setStatus] = useState("");
  
  const handleSubmit = async (data) => {
    setStatus("Saving requisition...");
    try {
      await createRequisition(data);
      setStatus("Requisition saved successfully");
    } catch (error) {
      setStatus("Failed to save requisition");
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      {/* Form fields */}
      <StatusAnnouncer message={status} />
    </form>
  );
}
```

## Component Testing Patterns

### 1. Component Unit Tests
```tsx
// Component test example
import { render, screen, fireEvent } from '@testing-library/react';
import { Button } from './button';

describe('Button', () => {
  it('renders with correct text', () => {
    render(<Button>Click me</Button>);
    expect(screen.getByRole('button', { name: /click me/i })).toBeInTheDocument();
  });

  it('calls onClick when clicked', () => {
    const handleClick = jest.fn();
    render(<Button onClick={handleClick}>Click me</Button>);
    
    fireEvent.click(screen.getByRole('button'));
    expect(handleClick).toHaveBeenCalledTimes(1);
  });

  it('applies variant styles correctly', () => {
    render(<Button variant="destructive">Delete</Button>);
    const button = screen.getByRole('button');
    expect(button).toHaveClass('bg-destructive');
  });

  it('is disabled when disabled prop is true', () => {
    render(<Button disabled>Disabled</Button>);
    const button = screen.getByRole('button');
    expect(button).toBeDisabled();
    expect(button).toHaveClass('disabled:opacity-50');
  });
});
```

### 2. Accessibility Testing
```tsx
// Accessibility test example
import { render } from '@testing-library/react';
import { axe, toHaveNoViolations } from 'jest-axe';
import { DataTable } from './data-table';

expect.extend(toHaveNoViolations);

describe('DataTable Accessibility', () => {
  it('should not have accessibility violations', async () => {
    const { container } = render(
      <DataTable
        columns={mockColumns}
        data={mockData}
        searchKey="name"
      />
    );
    
    const results = await axe(container);
    expect(results).toHaveNoViolations();
  });

  it('has proper ARIA labels', () => {
    render(<DataTable columns={mockColumns} data={mockData} />);
    
    expect(screen.getByRole('table')).toBeInTheDocument();
    expect(screen.getByRole('columnheader', { name: /name/i })).toBeInTheDocument();
    expect(screen.getAllByRole('row')).toHaveLength(mockData.length + 1); // +1 for header
  });
});
```

## Performance Optimization

### 1. Memoization
```tsx
// Memoized components
const ExpensiveComponent = React.memo(function ExpensiveComponent({ 
  data, 
  onUpdate 
}: {
  data: ComplexData[];
  onUpdate: (id: string) => void;
}) {
  const processedData = useMemo(() => {
    return data.map(item => ({
      ...item,
      computed: expensiveCalculation(item),
    }));
  }, [data]);

  const handleUpdate = useCallback((id: string) => {
    onUpdate(id);
  }, [onUpdate]);

  return (
    <div>
      {processedData.map(item => (
        <ItemComponent
          key={item.id}
          item={item}
          onUpdate={handleUpdate}
        />
      ))}
    </div>
  );
});
```

### 2. Lazy Loading
```tsx
// Lazy loaded components
const HeavyChart = lazy(() => import('./heavy-chart'));
const PDFViewer = lazy(() => import('./pdf-viewer'));

function Dashboard() {
  const [showChart, setShowChart] = useState(false);

  return (
    <div>
      <Button onClick={() => setShowChart(true)}>
        Show Chart
      </Button>
      
      {showChart && (
        <Suspense fallback={<ChartSkeleton />}>
          <HeavyChart />
        </Suspense>
      )}
    </div>
  );
}
```

### 3. Virtual Scrolling
```tsx
// Virtual scrolling for large lists
import { FixedSizeList as List } from 'react-window';

function VirtualizedList({ items }: { items: any[] }) {
  const Row = ({ index, style }: { index: number; style: React.CSSProperties }) => (
    <div style={style}>
      <ItemComponent item={items[index]} />
    </div>
  );

  return (
    <List
      height={600}
      itemCount={items.length}
      itemSize={80}
      width="100%"
    >
      {Row}
    </List>
  );
}
```

This component system provides a solid foundation for building consistent, accessible, and maintainable user interfaces while maintaining flexibility for custom requirements.