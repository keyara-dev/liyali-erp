# Property-Based Tests for Purchase Order Mutations

This directory contains property-based tests for the Purchase Order submission workflow.

## Testing Framework Setup Required

These tests require the following dependencies to be installed:

```bash
# Install testing dependencies
pnpm add -D vitest @testing-library/react @testing-library/react-hooks @vitest/ui
pnpm add -D fast-check @types/node
pnpm add -D jsdom @testing-library/jest-dom
```

## Vitest Configuration

Create a `vitest.config.ts` file in the frontend root:

```typescript
import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";
import path from "path";

export default defineConfig({
  plugins: [react()],
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: ["./src/test/setup.ts"],
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
});
```

## Test Setup File

Create `frontend/src/test/setup.ts`:

```typescript
import "@testing-library/jest-dom";
import { expect, afterEach, vi } from "vitest";
import { cleanup } from "@testing-library/react";

// Cleanup after each test
afterEach(() => {
  cleanup();
});

// Mock Next.js router
vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: vi.fn(),
    replace: vi.fn(),
    prefetch: vi.fn(),
  }),
  usePathname: () => "/",
  useSearchParams: () => new URLSearchParams(),
}));
```

## Package.json Scripts

Add these scripts to `frontend/package.json`:

```json
{
  "scripts": {
    "test": "vitest",
    "test:ui": "vitest --ui",
    "test:run": "vitest run",
    "test:coverage": "vitest run --coverage"
  }
}
```

## Running the Tests

Once the setup is complete, run the tests with:

```bash
# Run all tests in watch mode
pnpm test

# Run tests once
pnpm test:run

# Run with UI
pnpm test:ui

# Run with coverage
pnpm test:coverage
```

## Property-Based Testing with fast-check

The tests use [fast-check](https://github.com/dubzzz/fast-check) for property-based testing. This approach:

1. **Generates random test cases**: Instead of writing specific examples, we define properties that should hold for all valid inputs
2. **Finds edge cases automatically**: fast-check explores the input space to find counterexamples
3. **Provides better coverage**: 100+ random test cases per property ensure comprehensive testing

### Example Property

```typescript
// Property: For any PO and workflow ID, the system should call the API with workflowId
it("should call API with workflowId for any valid data", async () => {
  await fc.assert(
    fc.asyncProperty(submitRequestArbitrary, async (submitData) => {
      // Test that the property holds for this random input
      await result.current.mutateAsync(submitData);
      expect(mockAPI).toHaveBeenCalledWith(
        expect.objectContaining({ workflowId: submitData.workflowId }),
      );
    }),
    { numRuns: 100 }, // Run 100 random test cases
  );
});
```

## Test Files

- `use-purchase-order-mutations.property.test.ts` - Property-based tests for submit workflow API integration
  - **Property 2**: Submit workflow API integration (Requirements 1.4, 9.5, 14.1)
  - Validates that POST /purchase-orders/:id/submit is called with correct workflowId
  - Tests all valid combinations of submission data
  - Verifies error handling and user context preservation

- `use-purchase-order-detail.property.test.ts` - Property-based tests for permission calculations
  - **Property 1**: Submit button visibility for draft POs (Requirements 1.1, 1.2, 7.1, 7.2)
  - Validates that canSubmit is true only for creator in DRAFT status
  - Tests all combinations of PO status, creator, and user
  - Verifies permission consistency and invariants
  - Additional tests for canEdit and canWithdraw permissions

- `use-purchase-order-detail-approval-controls.property.test.ts` - Property-based tests for approval action controls visibility
  - **Property 5**: Approval action controls visibility (Requirements 3.3, 3.8, 7.3)
  - Validates that approval controls are visible if and only if status is PENDING and canApprove is true
  - Tests all combinations of PO status and canApprove values
  - Verifies that non-approvers cannot see approval controls
  - Verifies that non-PENDING POs never show approval controls
  - Additional tests for loading states and status case-insensitivity

## Coverage

The property-based tests validate:

### Submit Workflow (use-purchase-order-mutations.property.test.ts)

- ✅ API integration with correct endpoint and request body
- ✅ workflowId is always included in the request
- ✅ All required fields are passed correctly
- ✅ Optional fields (comments) are handled properly
- ✅ User context (ID, name, role) is preserved for audit trail
- ✅ Error handling works for all input combinations
- ✅ Edge cases (minimal fields, long comments, etc.)

### Permission Calculations (use-purchase-order-detail.property.test.ts)

- ✅ canSubmit is true only for creator in DRAFT status
- ✅ canSubmit is false for all non-DRAFT statuses
- ✅ canSubmit is false when user is not creator
- ✅ isCreator is calculated independently of status
- ✅ Permissions are consistent for the same inputs
- ✅ canEdit is true only for creator in DRAFT or REJECTED status
- ✅ canWithdraw is true only for creator in PENDING status
- ✅ Non-creators have no modification permissions
- ✅ Status comparison is case-insensitive

### Approval Action Controls Visibility (use-purchase-order-detail-approval-controls.property.test.ts)

- ✅ Approval controls visible if and only if status is PENDING and canApprove is true
- ✅ Approval controls always visible for PENDING POs when canApprove is true
- ✅ Approval controls never visible for PENDING POs when canApprove is false
- ✅ Approval controls never visible for non-PENDING POs regardless of canApprove
- ✅ DRAFT POs show "Ready to Submit" message instead of approval controls
- ✅ Visibility is consistent for the same inputs
- ✅ Status comparison is case-insensitive
- ✅ Loading state shows loading indicator, not approval controls
- ✅ Approval controls require both PENDING status AND canApprove=true (AND logic)

## Troubleshooting

### Module Resolution Issues

If you encounter module resolution errors, ensure:

1. `tsconfig.json` includes the test files
2. Path aliases are configured in both `tsconfig.json` and `vitest.config.ts`
3. All dependencies are installed

### React Query Issues

If tests fail with React Query errors:

1. Ensure each test creates a fresh QueryClient
2. Use `waitFor` for async assertions
3. Set `retry: false` in QueryClient options for tests

### Fast-check Issues

If property tests fail:

1. Check the counterexample in the error message
2. Use `verbose: true` option for detailed output
3. Reduce `numRuns` for faster debugging
4. Add `.filter()` to arbitraries to exclude invalid inputs
