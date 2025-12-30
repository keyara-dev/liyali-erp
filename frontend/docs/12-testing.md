# Testing Strategy

Comprehensive testing approach for the Liyali Gateway Frontend, covering unit tests, integration tests, end-to-end tests, and testing best practices.

## Testing Philosophy

The testing strategy follows the **Testing Pyramid** principle:

```
                    /\
                   /  \
                  / E2E \     ← Few, high-value tests
                 /______\
                /        \
               /Integration\ ← Some integration tests
              /__________\
             /            \
            /  Unit Tests  \   ← Many, fast tests
           /________________\
```

### Testing Principles

1. **Fast Feedback**: Tests should run quickly to provide immediate feedback
2. **Reliable**: Tests should be deterministic and not flaky
3. **Maintainable**: Tests should be easy to understand and modify
4. **Comprehensive**: Critical paths should have good test coverage
5. **Realistic**: Tests should simulate real user interactions

## Testing Stack

### Core Testing Libraries

```json
{
  "devDependencies": {
    "@testing-library/react": "^14.0.0",
    "@testing-library/jest-dom": "^6.1.0",
    "@testing-library/user-event": "^14.5.0",
    "jest": "^29.7.0",
    "jest-environment-jsdom": "^29.7.0",
    "@types/jest": "^29.5.0",
    "msw": "^2.0.0",
    "playwright": "^1.40.0"
  }
}
```

### Test Configuration

```javascript
// jest.config.js
const nextJest = require('next/jest');

const createJestConfig = nextJest({
  dir: './',
});

const customJestConfig = {
  setupFilesAfterEnv: ['<rootDir>/jest.setup.js'],
  testEnvironment: 'jest-environment-jsdom',
  testPathIgnorePatterns: ['<rootDir>/.next/', '<rootDir>/node_modules/'],
  collectCoverageFrom: [
    'src/**/*.{js,jsx,ts,tsx}',
    '!src/**/*.d.ts',
    '!src/**/*.stories.{js,jsx,ts,tsx}',
    '!src/**/__tests__/**',
  ],
  coverageThreshold: {
    global: {
      branches: 70,
      functions: 70,
      lines: 70,
      statements: 70,
    },
  },
  moduleNameMapping: {
    '^@/(.*)$': '<rootDir>/src/$1',
  },
};

module.exports = createJestConfig(customJestConfig);
```

```javascript
// jest.setup.js
import '@testing-library/jest-dom';
import { server } from './src/mocks/server';

// Mock Next.js router
jest.mock('next/navigation', () => ({
  useRouter: () => ({
    push: jest.fn(),
    replace: jest.fn(),
    back: jest.fn(),
    forward: jest.fn(),
    refresh: jest.fn(),
  }),
  useSearchParams: () => new URLSearchParams(),
  usePathname: () => '/',
}));

// Setup MSW
beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

// Mock IntersectionObserver
global.IntersectionObserver = jest.fn().mockImplementation(() => ({
  observe: jest.fn(),
  unobserve: jest.fn(),
  disconnect: jest.fn(),
}));
```

## Unit Testing

### Component Testing

#### Basic Component Tests

```typescript
// src/components/ui/__tests__/button.test.tsx
import { render, screen, fireEvent } from '@testing-library/react';
import { Button } from '../button';

describe('Button', () => {
  it('renders with correct text', () => {
    render(<Button>Click me</Button>);
    expect(screen.getByRole('button', { name: /click me/i })).toBeInTheDocument();
  });

  it('applies variant styles correctly', () => {
    render(<Button variant="destructive">Delete</Button>);
    const button = screen.getByRole('button');
    expect(button).toHaveClass('bg-destructive');
  });

  it('handles click events', () => {
    const handleClick = jest.fn();
    render(<Button onClick={handleClick}>Click me</Button>);
    
    fireEvent.click(screen.getByRole('button'));
    expect(handleClick).toHaveBeenCalledTimes(1);
  });

  it('shows loading state', () => {
    render(<Button isLoading>Loading</Button>);
    const button = screen.getByRole('button');
    
    expect(button).toBeDisabled();
    expect(screen.getByTestId('loading-spinner')).toBeInTheDocument();
  });

  it('forwards ref correctly', () => {
    const ref = React.createRef<HTMLButtonElement>();
    render(<Button ref={ref}>Button</Button>);
    
    expect(ref.current).toBeInstanceOf(HTMLButtonElement);
  });
});
```

#### Form Component Testing

```typescript
// src/components/forms/__tests__/requisition-form.test.tsx
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { RequisitionForm } from '../requisition-form';

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });

  return render(
    <QueryClientProvider client={queryClient}>
      {ui}
    </QueryClientProvider>
  );
}

describe('RequisitionForm', () => {
  const user = userEvent.setup();

  it('validates required fields', async () => {
    const onSubmit = jest.fn();
    renderWithProviders(<RequisitionForm onSubmit={onSubmit} />);

    // Try to submit empty form
    await user.click(screen.getByRole('button', { name: /submit/i }));

    // Check for validation errors
    expect(screen.getByText(/title is required/i)).toBeInTheDocument();
    expect(screen.getByText(/department is required/i)).toBeInTheDocument();
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it('submits form with valid data', async () => {
    const onSubmit = jest.fn();
    renderWithProviders(<RequisitionForm onSubmit={onSubmit} />);

    // Fill form fields
    await user.type(screen.getByLabelText(/title/i), 'Office Supplies');
    await user.selectOptions(screen.getByLabelText(/department/i), 'IT');
    await user.type(screen.getByLabelText(/justification/i), 'Need supplies for new project');

    // Add an item
    await user.click(screen.getByRole('button', { name: /add item/i }));
    await user.type(screen.getByLabelText(/item description/i), 'Laptop');
    await user.type(screen.getByLabelText(/quantity/i), '2');
    await user.type(screen.getByLabelText(/unit price/i), '1000');

    // Submit form
    await user.click(screen.getByRole('button', { name: /submit/i }));

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith({
        title: 'Office Supplies',
        department: 'IT',
        justification: 'Need supplies for new project',
        items: [{
          description: 'Laptop',
          quantity: 2,
          unitPrice: 1000,
          total: 2000,
        }],
        totalAmount: 2000,
      });
    });
  });

  it('calculates totals correctly', async () => {
    renderWithProviders(<RequisitionForm />);

    // Add multiple items
    await user.click(screen.getByRole('button', { name: /add item/i }));
    await user.type(screen.getByLabelText(/item description/i), 'Laptop');
    await user.type(screen.getByLabelText(/quantity/i), '2');
    await user.type(screen.getByLabelText(/unit price/i), '1000');

    await user.click(screen.getByRole('button', { name: /add item/i }));
    const descriptions = screen.getAllByLabelText(/item description/i);
    const quantities = screen.getAllByLabelText(/quantity/i);
    const prices = screen.getAllByLabelText(/unit price/i);

    await user.type(descriptions[1], 'Mouse');
    await user.type(quantities[1], '5');
    await user.type(prices[1], '25');

    // Check total calculation
    await waitFor(() => {
      expect(screen.getByText(/total: \$2,125\.00/i)).toBeInTheDocument();
    });
  });
});
```

### Hook Testing

#### Custom Hook Tests

```typescript
// src/hooks/__tests__/use-requisitions.test.ts
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useRequisitions } from '../use-requisitions';
import { server } from '../../mocks/server';
import { rest } from 'msw';

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });

  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
}

describe('useRequisitions', () => {
  it('fetches requisitions successfully', async () => {
    const { result } = renderHook(() => useRequisitions(), {
      wrapper: createWrapper(),
    });

    expect(result.current.isLoading).toBe(true);

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data).toHaveLength(3);
    expect(result.current.data[0]).toMatchObject({
      id: 'req-1',
      title: 'Office Supplies',
      status: 'PENDING',
    });
  });

  it('handles error states', async () => {
    // Mock API error
    server.use(
      rest.get('/api/requisitions', (req, res, ctx) => {
        return res(ctx.status(500), ctx.json({ message: 'Server error' }));
      })
    );

    const { result } = renderHook(() => useRequisitions(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });

    expect(result.current.error).toBeTruthy();
  });

  it('filters requisitions by status', async () => {
    const { result } = renderHook(
      () => useRequisitions({ status: 'APPROVED' }),
      { wrapper: createWrapper() }
    );

    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(result.current.data).toHaveLength(1);
    expect(result.current.data[0].status).toBe('APPROVED');
  });
});
```

#### State Management Hook Tests

```typescript
// src/hooks/__tests__/use-approval-store.test.ts
import { renderHook, act } from '@testing-library/react';
import { useApprovalStore } from '../use-approval-store';

describe('useApprovalStore', () => {
  beforeEach(() => {
    // Reset store before each test
    useApprovalStore.getState().reset();
  });

  it('manages selected items', () => {
    const { result } = renderHook(() => useApprovalStore());

    expect(result.current.selectedItems).toEqual([]);

    act(() => {
      result.current.setSelectedItems(['item-1', 'item-2']);
    });

    expect(result.current.selectedItems).toEqual(['item-1', 'item-2']);
  });

  it('sets bulk action', () => {
    const { result } = renderHook(() => useApprovalStore());

    act(() => {
      result.current.setBulkAction('approve');
    });

    expect(result.current.bulkAction).toBe('approve');
  });

  it('clears selection', () => {
    const { result } = renderHook(() => useApprovalStore());

    act(() => {
      result.current.setSelectedItems(['item-1', 'item-2']);
      result.current.setBulkAction('approve');
    });

    act(() => {
      result.current.clearSelection();
    });

    expect(result.current.selectedItems).toEqual([]);
    expect(result.current.bulkAction).toBeNull();
  });
});
```

## Integration Testing

### Page Component Tests

```typescript
// src/app/(private)/requisitions/__tests__/page.test.tsx
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import RequisitionsPage from '../page';

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });

  return render(
    <QueryClientProvider client={queryClient}>
      {ui}
    </QueryClientProvider>
  );
}

describe('RequisitionsPage', () => {
  const user = userEvent.setup();

  it('displays requisitions list', async () => {
    renderWithProviders(<RequisitionsPage />);

    // Wait for data to load
    await waitFor(() => {
      expect(screen.getByText('Office Supplies')).toBeInTheDocument();
    });

    // Check that requisitions are displayed
    expect(screen.getByText('REQ-001')).toBeInTheDocument();
    expect(screen.getByText('PENDING')).toBeInTheDocument();
  });

  it('filters requisitions by status', async () => {
    renderWithProviders(<RequisitionsPage />);

    await waitFor(() => {
      expect(screen.getByText('Office Supplies')).toBeInTheDocument();
    });

    // Apply status filter
    await user.click(screen.getByRole('button', { name: /filter/i }));
    await user.click(screen.getByText('Approved'));

    await waitFor(() => {
      expect(screen.queryByText('Office Supplies')).not.toBeInTheDocument();
      expect(screen.getByText('Approved Requisition')).toBeInTheDocument();
    });
  });

  it('creates new requisition', async () => {
    renderWithProviders(<RequisitionsPage />);

    // Click create button
    await user.click(screen.getByRole('button', { name: /create requisition/i }));

    // Fill form in modal
    await user.type(screen.getByLabelText(/title/i), 'New Requisition');
    await user.selectOptions(screen.getByLabelText(/department/i), 'HR');

    // Submit form
    await user.click(screen.getByRole('button', { name: /create/i }));

    // Verify success message
    await waitFor(() => {
      expect(screen.getByText(/requisition created successfully/i)).toBeInTheDocument();
    });
  });

  it('handles bulk operations', async () => {
    renderWithProviders(<RequisitionsPage />);

    await waitFor(() => {
      expect(screen.getByText('Office Supplies')).toBeInTheDocument();
    });

    // Select multiple items
    const checkboxes = screen.getAllByRole('checkbox');
    await user.click(checkboxes[1]); // First item
    await user.click(checkboxes[2]); // Second item

    // Perform bulk action
    await user.click(screen.getByRole('button', { name: /bulk approve/i }));

    // Confirm action
    await user.click(screen.getByRole('button', { name: /confirm/i }));

    await waitFor(() => {
      expect(screen.getByText(/2 requisitions approved/i)).toBeInTheDocument();
    });
  });
});
```

### Workflow Integration Tests

```typescript
// src/components/workflows/__tests__/approval-workflow.test.tsx
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ApprovalWorkflow } from '../approval-workflow';
import { TestProviders } from '../../../test-utils';

describe('ApprovalWorkflow', () => {
  const user = userEvent.setup();
  
  const mockRequisition = {
    id: 'req-1',
    title: 'Office Supplies',
    status: 'PENDING_APPROVAL',
    totalAmount: 1500,
    currentStage: 'department_manager',
  };

  it('displays approval stages correctly', () => {
    render(
      <TestProviders>
        <ApprovalWorkflow requisition={mockRequisition} />
      </TestProviders>
    );

    expect(screen.getByText('Department Manager')).toBeInTheDocument();
    expect(screen.getByText('Finance Officer')).toBeInTheDocument();
    expect(screen.getByText('Director')).toBeInTheDocument();
  });

  it('handles approval action', async () => {
    const onApprove = jest.fn();
    
    render(
      <TestProviders>
        <ApprovalWorkflow 
          requisition={mockRequisition} 
          onApprove={onApprove}
        />
      </TestProviders>
    );

    // Click approve button
    await user.click(screen.getByRole('button', { name: /approve/i }));

    // Add comments
    await user.type(
      screen.getByLabelText(/comments/i), 
      'Approved for business needs'
    );

    // Confirm approval
    await user.click(screen.getByRole('button', { name: /confirm approval/i }));

    await waitFor(() => {
      expect(onApprove).toHaveBeenCalledWith({
        requisitionId: 'req-1',
        action: 'APPROVE',
        comments: 'Approved for business needs',
      });
    });
  });

  it('handles rejection with reason', async () => {
    const onReject = jest.fn();
    
    render(
      <TestProviders>
        <ApprovalWorkflow 
          requisition={mockRequisition} 
          onReject={onReject}
        />
      </TestProviders>
    );

    // Click reject button
    await user.click(screen.getByRole('button', { name: /reject/i }));

    // Select rejection reason
    await user.selectOptions(
      screen.getByLabelText(/reason/i), 
      'insufficient_budget'
    );

    // Add comments
    await user.type(
      screen.getByLabelText(/comments/i), 
      'Budget exceeded for this quarter'
    );

    // Confirm rejection
    await user.click(screen.getByRole('button', { name: /confirm rejection/i }));

    await waitFor(() => {
      expect(onReject).toHaveBeenCalledWith({
        requisitionId: 'req-1',
        action: 'REJECT',
        reason: 'insufficient_budget',
        comments: 'Budget exceeded for this quarter',
      });
    });
  });
});
```

## Mock Service Worker (MSW)

### API Mocking Setup

```typescript
// src/mocks/handlers.ts
import { rest } from 'msw';

export const handlers = [
  // Requisitions API
  rest.get('/api/requisitions', (req, res, ctx) => {
    const status = req.url.searchParams.get('status');
    
    let requisitions = [
      {
        id: 'req-1',
        documentNumber: '001',
        title: 'Office Supplies',
        status: 'PENDING',
        totalAmount: 1500,
        createdAt: '2024-01-15T10:00:00Z',
      },
      {
        id: 'req-2',
        documentNumber: '002',
        title: 'IT Equipment',
        status: 'APPROVED',
        totalAmount: 5000,
        createdAt: '2024-01-16T14:30:00Z',
      },
    ];

    if (status) {
      requisitions = requisitions.filter(req => req.status === status);
    }

    return res(
      ctx.status(200),
      ctx.json({
        success: true,
        data: requisitions,
      })
    );
  }),

  rest.post('/api/requisitions', (req, res, ctx) => {
    return res(
      ctx.status(201),
      ctx.json({
        success: true,
        data: {
          id: 'req-new',
          documentNumber: '003',
          ...req.body,
          status: 'DRAFT',
          createdAt: new Date().toISOString(),
        },
      })
    );
  }),

  rest.put('/api/requisitions/:id/approve', (req, res, ctx) => {
    const { id } = req.params;
    
    return res(
      ctx.status(200),
      ctx.json({
        success: true,
        data: {
          id,
          status: 'APPROVED',
          approvedAt: new Date().toISOString(),
        },
      })
    );
  }),

  // Error scenarios
  rest.get('/api/requisitions/error', (req, res, ctx) => {
    return res(
      ctx.status(500),
      ctx.json({
        success: false,
        message: 'Internal server error',
      })
    );
  }),
];
```

```typescript
// src/mocks/server.ts
import { setupServer } from 'msw/node';
import { handlers } from './handlers';

export const server = setupServer(...handlers);
```

### Dynamic Mock Responses

```typescript
// src/test-utils/mock-utils.ts
import { server } from '../mocks/server';
import { rest } from 'msw';

export function mockApiSuccess<T>(endpoint: string, data: T) {
  server.use(
    rest.get(endpoint, (req, res, ctx) => {
      return res(
        ctx.status(200),
        ctx.json({ success: true, data })
      );
    })
  );
}

export function mockApiError(endpoint: string, status = 500, message = 'Server error') {
  server.use(
    rest.get(endpoint, (req, res, ctx) => {
      return res(
        ctx.status(status),
        ctx.json({ success: false, message })
      );
    })
  );
}

// Usage in tests
describe('RequisitionsPage', () => {
  it('handles empty state', async () => {
    mockApiSuccess('/api/requisitions', []);
    
    render(<RequisitionsPage />);
    
    await waitFor(() => {
      expect(screen.getByText(/no requisitions found/i)).toBeInTheDocument();
    });
  });

  it('handles API errors', async () => {
    mockApiError('/api/requisitions', 500, 'Failed to fetch requisitions');
    
    render(<RequisitionsPage />);
    
    await waitFor(() => {
      expect(screen.getByText(/failed to fetch requisitions/i)).toBeInTheDocument();
    });
  });
});
```

## End-to-End Testing

### Playwright Configuration

```typescript
// playwright.config.ts
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'html',
  use: {
    baseURL: 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },
    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    },
    {
      name: 'Mobile Chrome',
      use: { ...devices['Pixel 5'] },
    },
  ],

  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:3000',
    reuseExistingServer: !process.env.CI,
  },
});
```

### E2E Test Examples

```typescript
// e2e/requisition-workflow.spec.ts
import { test, expect } from '@playwright/test';

test.describe('Requisition Workflow', () => {
  test.beforeEach(async ({ page }) => {
    // Login before each test
    await page.goto('/login');
    await page.fill('[data-testid="email"]', 'admin@liyali.com');
    await page.fill('[data-testid="password"]', 'admin123');
    await page.click('[data-testid="login-button"]');
    await expect(page).toHaveURL('/dashboard');
  });

  test('creates and approves requisition', async ({ page }) => {
    // Navigate to requisitions
    await page.click('[data-testid="nav-requisitions"]');
    await expect(page).toHaveURL('/requisitions');

    // Create new requisition
    await page.click('[data-testid="create-requisition"]');
    
    // Fill form
    await page.fill('[data-testid="title"]', 'E2E Test Requisition');
    await page.selectOption('[data-testid="department"]', 'IT');
    await page.fill('[data-testid="justification"]', 'Testing purposes');

    // Add item
    await page.click('[data-testid="add-item"]');
    await page.fill('[data-testid="item-description-0"]', 'Test Item');
    await page.fill('[data-testid="item-quantity-0"]', '1');
    await page.fill('[data-testid="item-price-0"]', '100');

    // Submit
    await page.click('[data-testid="submit-requisition"]');

    // Verify creation
    await expect(page.locator('[data-testid="success-message"]')).toContainText(
      'Requisition created successfully'
    );

    // Navigate to approvals
    await page.click('[data-testid="nav-approvals"]');
    
    // Find and approve the requisition
    const requisitionRow = page.locator('[data-testid="requisition-row"]').first();
    await requisitionRow.click('[data-testid="approve-button"]');
    
    // Add approval comments
    await page.fill('[data-testid="approval-comments"]', 'Approved via E2E test');
    await page.click('[data-testid="confirm-approval"]');

    // Verify approval
    await expect(page.locator('[data-testid="success-message"]')).toContainText(
      'Requisition approved successfully'
    );
  });

  test('handles bulk operations', async ({ page }) => {
    await page.goto('/requisitions');

    // Select multiple requisitions
    await page.check('[data-testid="select-requisition-1"]');
    await page.check('[data-testid="select-requisition-2"]');

    // Perform bulk approval
    await page.click('[data-testid="bulk-approve"]');
    await page.click('[data-testid="confirm-bulk-action"]');

    // Verify bulk operation
    await expect(page.locator('[data-testid="success-message"]')).toContainText(
      '2 requisitions approved'
    );
  });

  test('searches and filters requisitions', async ({ page }) => {
    await page.goto('/requisitions');

    // Search
    await page.fill('[data-testid="search-input"]', 'Office Supplies');
    await expect(page.locator('[data-testid="requisition-row"]')).toHaveCount(1);

    // Clear search
    await page.fill('[data-testid="search-input"]', '');

    // Filter by status
    await page.click('[data-testid="status-filter"]');
    await page.click('[data-testid="status-pending"]');
    
    // Verify filtered results
    const statusBadges = page.locator('[data-testid="status-badge"]');
    await expect(statusBadges.first()).toContainText('PENDING');
  });
});
```

### Visual Regression Testing

```typescript
// e2e/visual-regression.spec.ts
import { test, expect } from '@playwright/test';

test.describe('Visual Regression Tests', () => {
  test('requisitions page layout', async ({ page }) => {
    await page.goto('/login');
    await page.fill('[data-testid="email"]', 'admin@liyali.com');
    await page.fill('[data-testid="password"]', 'admin123');
    await page.click('[data-testid="login-button"]');
    
    await page.goto('/requisitions');
    await page.waitForLoadState('networkidle');
    
    // Take screenshot
    await expect(page).toHaveScreenshot('requisitions-page.png');
  });

  test('requisition form modal', async ({ page }) => {
    await page.goto('/requisitions');
    await page.click('[data-testid="create-requisition"]');
    
    // Wait for modal to be visible
    await expect(page.locator('[data-testid="requisition-modal"]')).toBeVisible();
    
    // Take screenshot of modal
    await expect(page.locator('[data-testid="requisition-modal"]')).toHaveScreenshot(
      'requisition-form-modal.png'
    );
  });
});
```

## Test Utilities

### Custom Render Function

```typescript
// src/test-utils/index.tsx
import React from 'react';
import { render, RenderOptions } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ThemeProvider } from 'next-themes';

interface CustomRenderOptions extends Omit<RenderOptions, 'wrapper'> {
  queryClient?: QueryClient;
}

export function renderWithProviders(
  ui: React.ReactElement,
  {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    }),
    ...renderOptions
  }: CustomRenderOptions = {}
) {
  function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        <ThemeProvider attribute="class" defaultTheme="light">
          {children}
        </ThemeProvider>
      </QueryClientProvider>
    );
  }

  return render(ui, { wrapper: Wrapper, ...renderOptions });
}

// Re-export everything
export * from '@testing-library/react';
export { renderWithProviders as render };
```

### Test Data Factories

```typescript
// src/test-utils/factories.ts
import { faker } from '@faker-js/faker';

export function createMockRequisition(overrides: Partial<Requisition> = {}): Requisition {
  return {
    id: faker.string.uuid(),
    documentNumber: faker.string.numeric(3),
    title: faker.commerce.productName(),
    description: faker.lorem.sentence(),
    status: 'PENDING',
    totalAmount: faker.number.float({ min: 100, max: 10000, fractionDigits: 2 }),
    requestedBy: faker.person.fullName(),
    department: faker.helpers.arrayElement(['IT', 'HR', 'Finance', 'Operations']),
    createdAt: faker.date.recent().toISOString(),
    updatedAt: faker.date.recent().toISOString(),
    items: [createMockRequisitionItem()],
    ...overrides,
  };
}

export function createMockRequisitionItem(overrides: Partial<RequisitionItem> = {}): RequisitionItem {
  const quantity = faker.number.int({ min: 1, max: 10 });
  const unitPrice = faker.number.float({ min: 10, max: 1000, fractionDigits: 2 });
  
  return {
    id: faker.string.uuid(),
    description: faker.commerce.productName(),
    quantity,
    unit: faker.helpers.arrayElement(['pcs', 'kg', 'liters', 'boxes']),
    unitPrice,
    totalPrice: quantity * unitPrice,
    ...overrides,
  };
}

export function createMockUser(overrides: Partial<User> = {}): User {
  return {
    id: faker.string.uuid(),
    name: faker.person.fullName(),
    email: faker.internet.email(),
    role: faker.helpers.arrayElement(['ADMIN', 'MANAGER', 'USER']),
    department: faker.helpers.arrayElement(['IT', 'HR', 'Finance', 'Operations']),
    createdAt: faker.date.past().toISOString(),
    updatedAt: faker.date.recent().toISOString(),
    ...overrides,
  };
}
```

## Testing Best Practices

### Test Organization

1. **Group related tests** using `describe` blocks
2. **Use descriptive test names** that explain what is being tested
3. **Follow AAA pattern**: Arrange, Act, Assert
4. **Keep tests focused** on a single behavior
5. **Use data-testid** attributes for reliable element selection

### Async Testing

```typescript
// Good: Wait for specific conditions
await waitFor(() => {
  expect(screen.getByText('Data loaded')).toBeInTheDocument();
});

// Good: Wait for element to appear
await screen.findByText('Data loaded');

// Avoid: Arbitrary timeouts
await new Promise(resolve => setTimeout(resolve, 1000));
```

### Mock Management

```typescript
// Reset mocks between tests
beforeEach(() => {
  jest.clearAllMocks();
});

// Mock only what you need
jest.mock('../api/client', () => ({
  getRequisitions: jest.fn(),
  createRequisition: jest.fn(),
}));

// Use MSW for HTTP mocking instead of mocking fetch
```

### Accessibility Testing

```typescript
// Test keyboard navigation
await user.tab();
expect(screen.getByRole('button')).toHaveFocus();

// Test ARIA attributes
expect(screen.getByRole('button')).toHaveAttribute('aria-expanded', 'false');

// Test screen reader content
expect(screen.getByLabelText('Search requisitions')).toBeInTheDocument();
```

## Continuous Integration

### GitHub Actions Workflow

```yaml
# .github/workflows/test.yml
name: Test

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'
          cache: 'pnpm'
      
      - name: Install dependencies
        run: pnpm install
      
      - name: Run linting
        run: pnpm lint
      
      - name: Run type checking
        run: pnpm type-check
      
      - name: Run unit tests
        run: pnpm test --coverage
      
      - name: Upload coverage reports
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage/lcov.info
      
      - name: Run E2E tests
        run: pnpm test:e2e
        env:
          CI: true

  build:
    runs-on: ubuntu-latest
    needs: test

    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'
          cache: 'pnpm'
      
      - name: Install dependencies
        run: pnpm install
      
      - name: Build application
        run: pnpm build
```

This comprehensive testing strategy ensures high code quality, reliability, and maintainability while providing fast feedback during development.