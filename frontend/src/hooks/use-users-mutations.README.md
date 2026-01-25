# User Mutations Hook

A comprehensive set of reusable React Query mutation hooks for user management operations.

## Overview

The `use-users-mutations.ts` file provides a complete set of mutation hooks for all user-related operations including creation, updates, deletion, status management, and bulk operations.

## Features

- ✅ **Automatic Error Handling**: All hooks include built-in error handling with toast notifications
- ✅ **Success Callbacks**: Optional success callbacks for custom actions after operations
- ✅ **Query Invalidation**: Automatic cache invalidation to keep UI in sync
- ✅ **Loading States**: Built-in loading states via `isPending` property
- ✅ **TypeScript Support**: Full TypeScript support with proper types
- ✅ **Bulk Operations**: Support for bulk user operations
- ✅ **Consistent API**: All hooks follow the same pattern for easy usage

## Available Hooks

### 1. `useCreateUser(onSuccess?)`

Creates a new user in the system.

```typescript
const createUser = useCreateUser(() => {
  console.log("User created successfully");
  router.push("/admin/users");
});

await createUser.mutateAsync({
  email: "user@example.com",
  password: "password123",
  first_name: "John",
  last_name: "Doe",
  role: "requester",
  department_id: "dept-123",
});
```

### 2. `useUpdateUser(onSuccess?)`

Updates an existing user's information.

```typescript
const updateUser = useUpdateUser(() => {
  console.log("User updated successfully");
});

await updateUser.mutateAsync({
  userId: "user-123",
  data: {
    first_name: "Jane",
    last_name: "Smith",
    is_active: true,
  },
});
```

### 3. `useDeleteUser(onSuccess?)`

Deletes a user from the system.

```typescript
const deleteUser = useDeleteUser(() => {
  console.log("User deleted successfully");
});

await deleteUser.mutateAsync("user-123");
```

### 4. `useToggleUserStatus(onSuccess?)`

Toggles a user's active/inactive status.

```typescript
const toggleStatus = useToggleUserStatus();

await toggleStatus.mutateAsync({
  userId: "user-123",
  isActive: false,
});
```

### 5. `useActivateUser(onSuccess?)`

Activates a user account.

```typescript
const activateUser = useActivateUser();
await activateUser.mutateAsync("user-123");
```

### 6. `useDeactivateUser(onSuccess?)`

Deactivates a user account.

```typescript
const deactivateUser = useDeactivateUser();
await deactivateUser.mutateAsync("user-123");
```

### 7. `useResetUserPassword(onSuccess?)`

Resets a user's password.

```typescript
const resetPassword = useResetUserPassword();

await resetPassword.mutateAsync({
  userId: "user-123",
  password: "newPassword123",
});
```

### 8. `useToggleUserMFA(onSuccess?)`

Toggles Multi-Factor Authentication for a user.

```typescript
const toggleMFA = useToggleUserMFA();

await toggleMFA.mutateAsync({
  userId: "user-123",
  enabled: true,
});
```

### 9. `useBulkUserOperations(onSuccess?)`

Performs bulk operations on multiple users.

```typescript
const bulkOperations = useBulkUserOperations();

await bulkOperations.mutateAsync({
  operation: "activate",
  userIds: ["user-1", "user-2", "user-3"],
});
```

## Usage Patterns

### Basic Usage

```typescript
import { useCreateUser } from '@/hooks/use-users-mutations';

function CreateUserForm() {
  const createUser = useCreateUser();

  const handleSubmit = async (formData) => {
    try {
      await createUser.mutateAsync(formData);
      // Success is handled automatically
    } catch (error) {
      // Error is handled automatically
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      {/* form fields */}
      <button
        type="submit"
        disabled={createUser.isPending}
      >
        {createUser.isPending ? 'Creating...' : 'Create User'}
      </button>
    </form>
  );
}
```

### With Success Callback

```typescript
const createUser = useCreateUser((data) => {
  console.log("User created:", data);
  router.push("/admin/users");
  // Custom success actions
});
```

### In User List Component

```typescript
function UserList({ users }) {
  const deleteUser = useDeleteUser();
  const toggleStatus = useToggleUserStatus();

  return (
    <div>
      {users.map(user => (
        <div key={user.id}>
          <span>{user.name}</span>
          <button
            onClick={() => toggleStatus.mutateAsync({
              userId: user.id,
              isActive: !user.is_active
            })}
            disabled={toggleStatus.isPending}
          >
            {user.is_active ? 'Deactivate' : 'Activate'}
          </button>
          <button
            onClick={() => deleteUser.mutateAsync(user.id)}
            disabled={deleteUser.isPending}
          >
            Delete
          </button>
        </div>
      ))}
    </div>
  );
}
```

## Error Handling

All hooks include automatic error handling:

- Toast notifications for errors
- Console logging for debugging
- Proper error propagation for custom handling

```typescript
const createUser = useCreateUser();

try {
  await createUser.mutateAsync(userData);
  // Success handled automatically
} catch (error) {
  // Error already shown to user via toast
  // Additional custom error handling can go here
}
```

## Loading States

All hooks provide loading states via the `isPending` property:

```typescript
const createUser = useCreateUser();

return (
  <button disabled={createUser.isPending}>
    {createUser.isPending ? 'Creating...' : 'Create User'}
  </button>
);
```

## Cache Management

All hooks automatically invalidate relevant queries to keep the UI in sync:

- User lists are refreshed after create/update/delete operations
- Individual user data is refreshed after updates
- Related queries are invalidated as needed

## TypeScript Support

All hooks are fully typed with proper TypeScript interfaces:

```typescript
import type {
  CreateUserRequest,
  UpdateUserRequest,
} from "@/app/_actions/user-actions";

// Types are automatically inferred
const createUser = useCreateUser();
await createUser.mutateAsync(userData); // userData is typed as CreateUserRequest
```

## Best Practices

1. **Use Success Callbacks**: Provide success callbacks for navigation or custom actions
2. **Handle Loading States**: Always show loading states in your UI
3. **Batch Operations**: Use bulk operations for multiple users when possible
4. **Error Boundaries**: Consider using error boundaries for additional error handling
5. **Optimistic Updates**: Consider implementing optimistic updates for better UX

## Integration

The hooks are automatically exported from the main hooks index:

```typescript
import { useCreateUser, useUpdateUser } from "@/hooks";
// or
import { useCreateUser, useUpdateUser } from "@/hooks/use-users-mutations";
```

## Dependencies

- `@tanstack/react-query` - For mutation management
- `sonner` - For toast notifications
- User action functions from `@/app/_actions/user-actions`
- Query keys from `@/lib/constants`
