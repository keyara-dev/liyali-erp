# Component Library

Comprehensive documentation of the reusable UI components in the Liyali Gateway Frontend, built on top of shadcn/ui and Radix UI primitives.

## Component Architecture

### Design System Hierarchy

```
┌─────────────────────────────────────────────────────────────┐
│                    Component Hierarchy                      │
├─────────────────────────────────────────────────────────────┤
│ Application Components (Feature-specific)                  │
│   ├── RequisitionForm, ApprovalWorkflow, etc.             │
│                                                             │
│ Composite Components (Complex UI patterns)                 │
│   ├── DataTable, FormBuilder, SearchFilter, etc.          │
│                                                             │
│ Base Components (shadcn/ui)                                │
│   ├── Button, Input, Dialog, Card, etc.                   │
│                                                             │
│ Primitive Components (Radix UI)                            │
│   ├── Accessible, unstyled components                      │
│                                                             │
│ Design Tokens (CSS Variables)                              │
│   ├── Colors, spacing, typography, etc.                   │
└─────────────────────────────────────────────────────────────┘
```

## Base UI Components

### Button Component

A versatile button component with multiple variants and states.

```typescript
// src/components/ui/button.tsx
import * as React from "react";
import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const buttonVariants = cva(
  "inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50",
  {
    variants: {
      variant: {
        default: "bg-primary text-primary-foreground hover:bg-primary/90",
        destructive: "bg-destructive text-destructive-foreground hover:bg-destructive/90",
        outline: "border border-input bg-background hover:bg-accent hover:text-accent-foreground",
        secondary: "bg-secondary text-secondary-foreground hover:bg-secondary/80",
        ghost: "hover:bg-accent hover:text-accent-foreground",
        link: "text-primary underline-offset-4 hover:underline",
      },
      size: {
        default: "h-10 px-4 py-2",
        sm: "h-9 rounded-md px-3",
        lg: "h-11 rounded-md px-8",
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
  isLoading?: boolean;
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild = false, isLoading, children, disabled, ...props }, ref) => {
    const Comp = asChild ? Slot : "button";
    
    return (
      <Comp
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        disabled={disabled || isLoading}
        {...props}
      >
        {isLoading ? (
          <>
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            Loading...
          </>
        ) : (
          children
        )}
      </Comp>
    );
  }
);

Button.displayName = "Button";

export { Button, buttonVariants };
```

#### Usage Examples

```typescript
// Basic usage
<Button>Click me</Button>

// With variants
<Button variant="outline">Outline Button</Button>
<Button variant="destructive">Delete</Button>
<Button variant="ghost">Ghost Button</Button>

// With sizes
<Button size="sm">Small</Button>
<Button size="lg">Large</Button>
<Button size="icon"><Plus className="h-4 w-4" /></Button>

// With loading state
<Button isLoading>Saving...</Button>

// As child (renders as different element)
<Button asChild>
  <Link href="/dashboard">Go to Dashboard</Link>
</Button>
```

### Input Component

Form input component with validation states and accessibility features.

```typescript
// src/components/ui/input.tsx
import * as React from "react";
import { cn } from "@/lib/utils";

export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  error?: boolean;
  helperText?: string;
}

const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, error, helperText, ...props }, ref) => {
    return (
      <div className="space-y-1">
        <input
          type={type}
          className={cn(
            "flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50",
            error && "border-destructive focus-visible:ring-destructive",
            className
          )}
          ref={ref}
          {...props}
        />
        {helperText && (
          <p className={cn(
            "text-xs",
            error ? "text-destructive" : "text-muted-foreground"
          )}>
            {helperText}
          </p>
        )}
      </div>
    );
  }
);

Input.displayName = "Input";

export { Input };
```

#### Usage Examples

```typescript
// Basic input
<Input placeholder="Enter your name" />

// With error state
<Input 
  error 
  helperText="This field is required" 
  placeholder="Email address" 
/>

// Different types
<Input type="email" placeholder="Email" />
<Input type="password" placeholder="Password" />
<Input type="number" placeholder="Amount" />
```

### Card Component

Container component for grouping related content.

```typescript
// src/components/ui/card.tsx
import * as React from "react";
import { cn } from "@/lib/utils";

const Card = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      className={cn(
        "rounded-lg border bg-card text-card-foreground shadow-sm",
        className
      )}
      {...props}
    />
  )
);

const CardHeader = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      className={cn("flex flex-col space-y-1.5 p-6", className)}
      {...props}
    />
  )
);

const CardTitle = React.forwardRef<HTMLParagraphElement, React.HTMLAttributes<HTMLHeadingElement>>(
  ({ className, ...props }, ref) => (
    <h3
      ref={ref}
      className={cn(
        "text-2xl font-semibold leading-none tracking-tight",
        className
      )}
      {...props}
    />
  )
);

const CardDescription = React.forwardRef<HTMLParagraphElement, React.HTMLAttributes<HTMLParagraphElement>>(
  ({ className, ...props }, ref) => (
    <p
      ref={ref}
      className={cn("text-sm text-muted-foreground", className)}
      {...props}
    />
  )
);

const CardContent = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div ref={ref} className={cn("p-6 pt-0", className)} {...props} />
  )
);

const CardFooter = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      className={cn("flex items-center p-6 pt-0", className)}
      {...props}
    />
  )
);

Card.displayName = "Card";
CardHeader.displayName = "CardHeader";
CardTitle.displayName = "CardTitle";
CardDescription.displayName = "CardDescription";
CardContent.displayName = "CardContent";
CardFooter.displayName = "CardFooter";

export { Card, CardHeader, CardFooter, CardTitle, CardDescription, CardContent };
```

#### Usage Examples

```typescript
// Basic card
<Card>
  <CardHeader>
    <CardTitle>Requisitions</CardTitle>
    <CardDescription>Manage your purchase requests</CardDescription>
  </CardHeader>
  <CardContent>
    <p>Card content goes here</p>
  </CardContent>
  <CardFooter>
    <Button>Action</Button>
  </CardFooter>
</Card>

// Interactive card
<Card className="hover:shadow-md transition-shadow cursor-pointer">
  <CardContent className="p-4">
    <h4 className="font-semibold">REQ-001</h4>
    <p className="text-sm text-muted-foreground">Office Supplies</p>
  </CardContent>
</Card>
```

## Composite Components

### DataTable Component

Advanced table component with sorting, filtering, and pagination.

```typescript
// src/components/ui/data-table.tsx
import * as React from "react";
import {
  ColumnDef,
  ColumnFiltersState,
  SortingState,
  VisibilityState,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
} from "@tanstack/react-table";
import { ChevronDown, Search } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

interface DataTableProps<TData, TValue> {
  columns: ColumnDef<TData, TValue>[];
  data: TData[];
  searchKey?: string;
  searchPlaceholder?: string;
  onRowClick?: (row: TData) => void;
}

export function DataTable<TData, TValue>({
  columns,
  data,
  searchKey,
  searchPlaceholder = "Search...",
  onRowClick,
}: DataTableProps<TData, TValue>) {
  const [sorting, setSorting] = React.useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>([]);
  const [columnVisibility, setColumnVisibility] = React.useState<VisibilityState>({});
  const [rowSelection, setRowSelection] = React.useState({});

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
      {/* Toolbar */}
      <div className="flex items-center justify-between">
        {searchKey && (
          <div className="flex items-center space-x-2">
            <Search className="h-4 w-4 text-muted-foreground" />
            <Input
              placeholder={searchPlaceholder}
              value={(table.getColumn(searchKey)?.getFilterValue() as string) ?? ""}
              onChange={(event) =>
                table.getColumn(searchKey)?.setFilterValue(event.target.value)
              }
              className="max-w-sm"
            />
          </div>
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
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead key={header.id}>
                      {header.isPlaceholder
                        ? null
                        : flexRender(
                            header.column.columnDef.header,
                            header.getContext()
                          )}
                    </TableHead>
                  );
                })}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  data-state={row.getIsSelected() && "selected"}
                  className={onRowClick ? "cursor-pointer hover:bg-muted/50" : ""}
                  onClick={() => onRowClick?.(row.original)}
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

#### Usage Example

```typescript
// Define columns
const columns: ColumnDef<Requisition>[] = [
  {
    id: "select",
    header: ({ table }) => (
      <Checkbox
        checked={table.getIsAllPageRowsSelected()}
        onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
      />
    ),
    cell: ({ row }) => (
      <Checkbox
        checked={row.getIsSelected()}
        onCheckedChange={(value) => row.toggleSelected(!!value)}
      />
    ),
  },
  {
    accessorKey: "documentNumber",
    header: "Number",
  },
  {
    accessorKey: "title",
    header: "Title",
  },
  {
    accessorKey: "status",
    header: "Status",
    cell: ({ row }) => (
      <Badge variant={getStatusVariant(row.getValue("status"))}>
        {row.getValue("status")}
      </Badge>
    ),
  },
  {
    accessorKey: "totalAmount",
    header: "Amount",
    cell: ({ row }) => {
      const amount = parseFloat(row.getValue("totalAmount"));
      return new Intl.NumberFormat("en-US", {
        style: "currency",
        currency: "USD",
      }).format(amount);
    },
  },
];

// Use in component
<DataTable
  columns={columns}
  data={requisitions}
  searchKey="title"
  searchPlaceholder="Search requisitions..."
  onRowClick={(requisition) => router.push(`/requisitions/${requisition.id}`)}
/>
```

### FormBuilder Component

Dynamic form builder for creating forms from schema definitions.

```typescript
// src/components/ui/form-builder.tsx
import * as React from "react";
import { useForm, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";

export interface FormField {
  name: string;
  label: string;
  type: 'text' | 'email' | 'password' | 'number' | 'textarea' | 'select' | 'checkbox';
  placeholder?: string;
  required?: boolean;
  options?: { value: string; label: string }[];
  validation?: z.ZodSchema;
}

export interface FormBuilderProps {
  fields: FormField[];
  onSubmit: (data: any) => void;
  defaultValues?: Record<string, any>;
  submitLabel?: string;
  isLoading?: boolean;
}

export function FormBuilder({
  fields,
  onSubmit,
  defaultValues = {},
  submitLabel = "Submit",
  isLoading = false,
}: FormBuilderProps) {
  // Generate Zod schema from fields
  const schema = React.useMemo(() => {
    const schemaFields: Record<string, z.ZodSchema> = {};
    
    fields.forEach(field => {
      let fieldSchema: z.ZodSchema = z.string();
      
      switch (field.type) {
        case 'number':
          fieldSchema = z.number();
          break;
        case 'email':
          fieldSchema = z.string().email();
          break;
        case 'checkbox':
          fieldSchema = z.boolean();
          break;
        default:
          fieldSchema = z.string();
      }
      
      if (field.required && field.type !== 'checkbox') {
        fieldSchema = fieldSchema.min(1, `${field.label} is required`);
      }
      
      if (field.validation) {
        fieldSchema = field.validation;
      }
      
      schemaFields[field.name] = fieldSchema;
    });
    
    return z.object(schemaFields);
  }, [fields]);

  const {
    control,
    handleSubmit,
    formState: { errors },
  } = useForm({
    resolver: zodResolver(schema),
    defaultValues,
  });

  const renderField = (field: FormField) => {
    const error = errors[field.name];
    
    switch (field.type) {
      case 'textarea':
        return (
          <Controller
            name={field.name}
            control={control}
            render={({ field: { onChange, value } }) => (
              <Textarea
                placeholder={field.placeholder}
                value={value || ''}
                onChange={onChange}
                error={!!error}
                helperText={error?.message as string}
              />
            )}
          />
        );
        
      case 'select':
        return (
          <Controller
            name={field.name}
            control={control}
            render={({ field: { onChange, value } }) => (
              <Select value={value} onValueChange={onChange}>
                <SelectTrigger className={error ? "border-destructive" : ""}>
                  <SelectValue placeholder={field.placeholder} />
                </SelectTrigger>
                <SelectContent>
                  {field.options?.map(option => (
                    <SelectItem key={option.value} value={option.value}>
                      {option.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          />
        );
        
      case 'checkbox':
        return (
          <Controller
            name={field.name}
            control={control}
            render={({ field: { onChange, value } }) => (
              <div className="flex items-center space-x-2">
                <Checkbox
                  checked={value}
                  onCheckedChange={onChange}
                />
                <Label>{field.label}</Label>
              </div>
            )}
          />
        );
        
      default:
        return (
          <Controller
            name={field.name}
            control={control}
            render={({ field: { onChange, value } }) => (
              <Input
                type={field.type}
                placeholder={field.placeholder}
                value={value || ''}
                onChange={onChange}
                error={!!error}
                helperText={error?.message as string}
              />
            )}
          />
        );
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      {fields.map(field => (
        <div key={field.name} className="space-y-2">
          {field.type !== 'checkbox' && (
            <Label htmlFor={field.name}>
              {field.label}
              {field.required && <span className="text-destructive ml-1">*</span>}
            </Label>
          )}
          {renderField(field)}
        </div>
      ))}
      
      <Button type="submit" isLoading={isLoading} className="w-full">
        {submitLabel}
      </Button>
    </form>
  );
}
```

#### Usage Example

```typescript
const userFormFields: FormField[] = [
  {
    name: 'name',
    label: 'Full Name',
    type: 'text',
    placeholder: 'Enter your full name',
    required: true,
  },
  {
    name: 'email',
    label: 'Email Address',
    type: 'email',
    placeholder: 'Enter your email',
    required: true,
  },
  {
    name: 'department',
    label: 'Department',
    type: 'select',
    required: true,
    options: [
      { value: 'it', label: 'IT' },
      { value: 'hr', label: 'HR' },
      { value: 'finance', label: 'Finance' },
    ],
  },
  {
    name: 'bio',
    label: 'Biography',
    type: 'textarea',
    placeholder: 'Tell us about yourself',
  },
  {
    name: 'newsletter',
    label: 'Subscribe to newsletter',
    type: 'checkbox',
  },
];

<FormBuilder
  fields={userFormFields}
  onSubmit={(data) => console.log(data)}
  submitLabel="Create User"
  isLoading={isCreating}
/>
```

## Application-Specific Components

### StatusBadge Component

Displays status with appropriate colors and styling.

```typescript
// src/components/ui/status-badge.tsx
import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const statusBadgeVariants = cva(
  "inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium",
  {
    variants: {
      status: {
        draft: "bg-gray-100 text-gray-800",
        pending: "bg-yellow-100 text-yellow-800",
        approved: "bg-green-100 text-green-800",
        rejected: "bg-red-100 text-red-800",
        completed: "bg-blue-100 text-blue-800",
      },
    },
    defaultVariants: {
      status: "draft",
    },
  }
);

export interface StatusBadgeProps
  extends React.HTMLAttributes<HTMLSpanElement>,
    VariantProps<typeof statusBadgeVariants> {
  status: 'draft' | 'pending' | 'approved' | 'rejected' | 'completed';
}

const StatusBadge = React.forwardRef<HTMLSpanElement, StatusBadgeProps>(
  ({ className, status, ...props }, ref) => {
    return (
      <span
        className={cn(statusBadgeVariants({ status }), className)}
        ref={ref}
        {...props}
      >
        {status.charAt(0).toUpperCase() + status.slice(1)}
      </span>
    );
  }
);

StatusBadge.displayName = "StatusBadge";

export { StatusBadge, statusBadgeVariants };
```

### EmptyState Component

Displays empty states with optional actions.

```typescript
// src/components/ui/empty-state.tsx
import * as React from "react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";

export interface EmptyStateProps extends React.HTMLAttributes<HTMLDivElement> {
  icon?: React.ReactNode;
  title: string;
  description?: string;
  action?: {
    label: string;
    onClick: () => void;
  };
}

const EmptyState = React.forwardRef<HTMLDivElement, EmptyStateProps>(
  ({ className, icon, title, description, action, ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={cn(
          "flex flex-col items-center justify-center py-12 px-4 text-center",
          className
        )}
        {...props}
      >
        {icon && (
          <div className="mb-4 text-muted-foreground">
            {icon}
          </div>
        )}
        
        <h3 className="text-lg font-semibold text-foreground mb-2">
          {title}
        </h3>
        
        {description && (
          <p className="text-sm text-muted-foreground mb-6 max-w-sm">
            {description}
          </p>
        )}
        
        {action && (
          <Button onClick={action.onClick}>
            {action.label}
          </Button>
        )}
      </div>
    );
  }
);

EmptyState.displayName = "EmptyState";

export { EmptyState };
```

#### Usage Example

```typescript
import { FileText, Plus } from "lucide-react";

<EmptyState
  icon={<FileText className="h-12 w-12" />}
  title="No requisitions found"
  description="You haven't created any requisitions yet. Create your first requisition to get started."
  action={{
    label: "Create Requisition",
    onClick: () => setShowCreateModal(true),
  }}
/>
```

## Component Documentation

### Storybook Integration

```typescript
// src/components/ui/button.stories.tsx
import type { Meta, StoryObj } from '@storybook/react';
import { Button } from './button';
import { Plus, Download } from 'lucide-react';

const meta: Meta<typeof Button> = {
  title: 'UI/Button',
  component: Button,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
  argTypes: {
    variant: {
      control: { type: 'select' },
      options: ['default', 'destructive', 'outline', 'secondary', 'ghost', 'link'],
    },
    size: {
      control: { type: 'select' },
      options: ['default', 'sm', 'lg', 'icon'],
    },
  },
};

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    children: 'Button',
  },
};

export const Variants: Story = {
  render: () => (
    <div className="flex gap-2">
      <Button variant="default">Default</Button>
      <Button variant="secondary">Secondary</Button>
      <Button variant="outline">Outline</Button>
      <Button variant="ghost">Ghost</Button>
      <Button variant="link">Link</Button>
      <Button variant="destructive">Destructive</Button>
    </div>
  ),
};

export const Sizes: Story = {
  render: () => (
    <div className="flex items-center gap-2">
      <Button size="sm">Small</Button>
      <Button size="default">Default</Button>
      <Button size="lg">Large</Button>
      <Button size="icon">
        <Plus className="h-4 w-4" />
      </Button>
    </div>
  ),
};

export const WithIcons: Story = {
  render: () => (
    <div className="flex gap-2">
      <Button>
        <Plus className="mr-2 h-4 w-4" />
        Add Item
      </Button>
      <Button variant="outline">
        <Download className="mr-2 h-4 w-4" />
        Download
      </Button>
    </div>
  ),
};

export const Loading: Story = {
  args: {
    isLoading: true,
    children: 'Please wait',
  },
};
```

## Accessibility Guidelines

### ARIA Attributes

All components include proper ARIA attributes:

```typescript
// Example: Accessible button with loading state
<Button
  aria-label={isLoading ? "Saving document" : "Save document"}
  aria-disabled={isLoading}
  disabled={isLoading}
>
  {isLoading ? "Saving..." : "Save"}
</Button>

// Example: Accessible form field
<div>
  <Label htmlFor="email">Email Address</Label>
  <Input
    id="email"
    type="email"
    aria-describedby="email-error"
    aria-invalid={!!error}
  />
  {error && (
    <p id="email-error" role="alert" className="text-destructive text-sm">
      {error.message}
    </p>
  )}
</div>
```

### Keyboard Navigation

Components support keyboard navigation:

```typescript
// Example: Keyboard-accessible dropdown
const DropdownMenu = () => {
  const [isOpen, setIsOpen] = useState(false);
  const [focusedIndex, setFocusedIndex] = useState(-1);

  const handleKeyDown = (e: KeyboardEvent) => {
    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault();
        setFocusedIndex(prev => Math.min(prev + 1, items.length - 1));
        break;
      case 'ArrowUp':
        e.preventDefault();
        setFocusedIndex(prev => Math.max(prev - 1, 0));
        break;
      case 'Enter':
        e.preventDefault();
        if (focusedIndex >= 0) {
          selectItem(items[focusedIndex]);
        }
        break;
      case 'Escape':
        setIsOpen(false);
        break;
    }
  };

  return (
    <div onKeyDown={handleKeyDown}>
      {/* Dropdown implementation */}
    </div>
  );
};
```

## Component Testing

### Unit Tests for Components

```typescript
// src/components/ui/__tests__/status-badge.test.tsx
import { render, screen } from '@testing-library/react';
import { StatusBadge } from '../status-badge';

describe('StatusBadge', () => {
  it('renders with correct status', () => {
    render(<StatusBadge status="approved" />);
    expect(screen.getByText('Approved')).toBeInTheDocument();
  });

  it('applies correct styling for each status', () => {
    const { rerender } = render(<StatusBadge status="pending" />);
    expect(screen.getByText('Pending')).toHaveClass('bg-yellow-100', 'text-yellow-800');

    rerender(<StatusBadge status="approved" />);
    expect(screen.getByText('Approved')).toHaveClass('bg-green-100', 'text-green-800');

    rerender(<StatusBadge status="rejected" />);
    expect(screen.getByText('Rejected')).toHaveClass('bg-red-100', 'text-red-800');
  });

  it('accepts custom className', () => {
    render(<StatusBadge status="draft" className="custom-class" />);
    expect(screen.getByText('Draft')).toHaveClass('custom-class');
  });
});
```

## Best Practices

### Component Design Principles

1. **Single Responsibility**: Each component should have one clear purpose
2. **Composition over Inheritance**: Use composition patterns for flexibility
3. **Accessibility First**: Include ARIA attributes and keyboard support
4. **Type Safety**: Use TypeScript for all component props and state
5. **Consistent API**: Follow consistent patterns across components

### Performance Considerations

1. **Memoization**: Use React.memo for expensive components
2. **Lazy Loading**: Dynamically import heavy components
3. **Bundle Size**: Keep component dependencies minimal
4. **Re-renders**: Optimize to prevent unnecessary re-renders

### Styling Guidelines

1. **Design Tokens**: Use CSS variables for consistent theming
2. **Responsive Design**: Include responsive utilities
3. **Dark Mode**: Support both light and dark themes
4. **Animation**: Use subtle animations for better UX

This component library provides a solid foundation for building consistent, accessible, and maintainable user interfaces across the Liyali Gateway Frontend application.