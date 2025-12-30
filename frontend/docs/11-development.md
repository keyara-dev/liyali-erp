# Development Guide

Comprehensive guide for developing and contributing to the Liyali Gateway Frontend, including coding standards, development workflows, and best practices.

## Development Environment Setup

### Prerequisites

- **Node.js 18+** (LTS recommended)
- **pnpm** (preferred package manager)
- **Git** for version control
- **VS Code** (recommended IDE)

### Initial Setup

```bash
# Clone the repository
git clone <repository-url>
cd liyali-gateway/frontend

# Install dependencies
pnpm install

# Copy environment file
cp .env.example .env.local

# Start development server
pnpm dev
```

### VS Code Configuration

Install recommended extensions and configure workspace settings:

```json
// .vscode/extensions.json
{
  "recommendations": [
    "bradlc.vscode-tailwindcss",
    "ms-vscode.vscode-typescript-next",
    "esbenp.prettier-vscode",
    "ms-vscode.vscode-eslint",
    "formulahendry.auto-rename-tag",
    "christian-kohler.path-intellisense"
  ]
}
```

```json
// .vscode/settings.json
{
  "typescript.preferences.importModuleSpecifier": "relative",
  "editor.formatOnSave": true,
  "editor.defaultFormatter": "esbenp.prettier-vscode",
  "editor.codeActionsOnSave": {
    "source.fixAll.eslint": "explicit"
  },
  "tailwindCSS.experimental.classRegex": [
    ["cva\\(([^)]*)\\)", "[\"'`]([^\"'`]*).*?[\"'`]"],
    ["cx\\(([^)]*)\\)", "(?:'|\"|`)([^']*)(?:'|\"|`)"]
  ]
}
```

## Project Structure

### Directory Organization

```
src/
├── app/                    # Next.js App Router
│   ├── (auth)/            # Authentication routes
│   ├── (private)/         # Protected routes
│   ├── api/               # API routes (if needed)
│   ├── globals.css        # Global styles
│   ├── layout.tsx         # Root layout
│   └── page.tsx           # Home page
├── components/            # React components
│   ├── ui/               # Base UI components (shadcn/ui)
│   ├── layout/           # Layout components
│   ├── auth/             # Authentication components
│   ├── workflows/        # Workflow-specific components
│   └── [feature]/        # Feature-specific components
├── hooks/                # Custom React hooks
├── lib/                  # Utilities and configurations
│   ├── api/              # API client
│   ├── auth/             # Authentication utilities
│   ├── storage/          # Storage management
│   └── utils/            # General utilities
├── types/                # TypeScript type definitions
└── styles/               # Additional styles
```

### File Naming Conventions

- **Components**: PascalCase (`UserProfile.tsx`)
- **Hooks**: camelCase with `use` prefix (`useUserData.ts`)
- **Utilities**: camelCase (`formatCurrency.ts`)
- **Types**: PascalCase (`UserType.ts`)
- **Constants**: UPPER_SNAKE_CASE (`API_ENDPOINTS.ts`)

## Coding Standards

### TypeScript Guidelines

#### Type Definitions

```typescript
// Define interfaces for all data structures
interface User {
  id: string;
  name: string;
  email: string;
  role: UserRole;
  createdAt: Date;
  updatedAt: Date;
}

// Use enums for fixed sets of values
enum UserRole {
  ADMIN = 'ADMIN',
  MANAGER = 'MANAGER',
  USER = 'USER',
}

// Use union types for flexible options
type Status = 'pending' | 'approved' | 'rejected';

// Generic types for reusable patterns
interface ApiResponse<T> {
  success: boolean;
  data: T;
  message?: string;
  errors?: string[];
}
```

#### Component Props

```typescript
// Always define prop interfaces
interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'outline';
  size?: 'sm' | 'md' | 'lg';
  isLoading?: boolean;
  children: React.ReactNode;
}

// Use React.forwardRef for components that need ref forwarding
const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ variant = 'primary', size = 'md', isLoading, children, ...props }, ref) => {
    return (
      <button
        ref={ref}
        className={cn(buttonVariants({ variant, size }))}
        disabled={isLoading}
        {...props}
      >
        {isLoading ? <Spinner /> : children}
      </button>
    );
  }
);

Button.displayName = 'Button';
```

### React Component Patterns

#### Functional Components

```typescript
// Use function declarations for named components
export function UserProfile({ userId }: { userId: string }) {
  const { data: user, isLoading, error } = useUser(userId);

  if (isLoading) return <UserProfileSkeleton />;
  if (error) return <ErrorDisplay error={error} />;
  if (!user) return <NotFound />;

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-bold">{user.name}</h1>
      <p className="text-muted-foreground">{user.email}</p>
    </div>
  );
}

// Use arrow functions for inline components
const UserCard = ({ user }: { user: User }) => (
  <Card>
    <CardContent>
      <h3>{user.name}</h3>
      <p>{user.email}</p>
    </CardContent>
  </Card>
);
```

#### Custom Hooks

```typescript
// Extract component logic into custom hooks
export function useUserManagement() {
  const [users, setUsers] = useState<User[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchUsers = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    
    try {
      const response = await apiClient.users.getAll();
      setUsers(response.data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch users');
    } finally {
      setIsLoading(false);
    }
  }, []);

  const createUser = useCallback(async (userData: CreateUserRequest) => {
    try {
      const response = await apiClient.users.create(userData);
      setUsers(prev => [...prev, response.data]);
      return response.data;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create user');
      throw err;
    }
  }, []);

  useEffect(() => {
    fetchUsers();
  }, [fetchUsers]);

  return {
    users,
    isLoading,
    error,
    fetchUsers,
    createUser,
  };
}
```

### Styling Guidelines

#### Tailwind CSS Usage

```typescript
// Use Tailwind classes for styling
export function Card({ children, className, ...props }: CardProps) {
  return (
    <div
      className={cn(
        "rounded-lg border bg-card text-card-foreground shadow-sm",
        className
      )}
      {...props}
    >
      {children}
    </div>
  );
}

// Use CSS variables for theme consistency
// globals.css
:root {
  --background: 0 0% 100%;
  --foreground: 222.2 84% 4.9%;
  --primary: 221.2 83.2% 53.3%;
  --primary-foreground: 210 40% 98%;
}

// Use cn() utility for conditional classes
const buttonVariants = cva(
  "inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors",
  {
    variants: {
      variant: {
        default: "bg-primary text-primary-foreground hover:bg-primary/90",
        outline: "border border-input hover:bg-accent hover:text-accent-foreground",
      },
      size: {
        default: "h-10 px-4 py-2",
        sm: "h-9 rounded-md px-3",
        lg: "h-11 rounded-md px-8",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
);
```

## Development Workflow

### Git Workflow

#### Branch Naming

- **Feature branches**: `feature/user-authentication`
- **Bug fixes**: `fix/login-validation-error`
- **Hotfixes**: `hotfix/critical-security-patch`
- **Chores**: `chore/update-dependencies`

#### Commit Messages

Follow conventional commit format:

```bash
# Format: type(scope): description

feat(auth): add user authentication system
fix(ui): resolve button alignment issue
docs(readme): update installation instructions
refactor(api): simplify user data fetching
test(auth): add login form validation tests
chore(deps): update React to v19
```

#### Pull Request Process

1. **Create feature branch** from `main`
2. **Implement changes** following coding standards
3. **Write tests** for new functionality
4. **Update documentation** if needed
5. **Create pull request** with descriptive title and description
6. **Request code review** from team members
7. **Address feedback** and make necessary changes
8. **Merge** after approval and CI passes

### Code Review Guidelines

#### What to Look For

- **Functionality**: Does the code work as intended?
- **Performance**: Are there any performance issues?
- **Security**: Are there any security vulnerabilities?
- **Maintainability**: Is the code easy to understand and modify?
- **Testing**: Are there adequate tests?
- **Documentation**: Is the code properly documented?

#### Review Checklist

```markdown
## Code Review Checklist

### Functionality
- [ ] Code works as intended
- [ ] Edge cases are handled
- [ ] Error handling is appropriate

### Code Quality
- [ ] Code follows project conventions
- [ ] No code duplication
- [ ] Functions are focused and single-purpose
- [ ] Variable names are descriptive

### Performance
- [ ] No unnecessary re-renders
- [ ] Efficient algorithms used
- [ ] Proper memoization where needed

### Security
- [ ] Input validation is present
- [ ] No sensitive data exposed
- [ ] Authentication/authorization checks

### Testing
- [ ] Unit tests cover new functionality
- [ ] Integration tests where appropriate
- [ ] Tests are meaningful and maintainable

### Documentation
- [ ] Code is self-documenting
- [ ] Complex logic is commented
- [ ] README updated if needed
```

## Testing Strategy

### Unit Testing

```typescript
// Component testing with React Testing Library
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

  it('shows loading state', () => {
    render(<Button isLoading>Loading</Button>);
    expect(screen.getByRole('button')).toBeDisabled();
    expect(screen.getByTestId('spinner')).toBeInTheDocument();
  });
});
```

### Hook Testing

```typescript
// Custom hook testing
import { renderHook, waitFor } from '@testing-library/react';
import { useUserData } from './use-user-data';

describe('useUserData', () => {
  it('fetches user data successfully', async () => {
    const { result } = renderHook(() => useUserData('user-123'));

    expect(result.current.isLoading).toBe(true);

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.user).toEqual({
      id: 'user-123',
      name: 'John Doe',
      email: 'john@example.com',
    });
  });

  it('handles error states', async () => {
    // Mock API to return error
    jest.spyOn(apiClient.users, 'getById').mockRejectedValue(new Error('User not found'));

    const { result } = renderHook(() => useUserData('invalid-id'));

    await waitFor(() => {
      expect(result.current.error).toBe('User not found');
    });
  });
});
```

### Integration Testing

```typescript
// Integration test example
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { UserManagementPage } from './user-management-page';

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

describe('UserManagementPage', () => {
  it('creates a new user successfully', async () => {
    renderWithProviders(<UserManagementPage />);

    // Open create user form
    fireEvent.click(screen.getByRole('button', { name: /add user/i }));

    // Fill form
    fireEvent.change(screen.getByLabelText(/name/i), {
      target: { value: 'Jane Doe' },
    });
    fireEvent.change(screen.getByLabelText(/email/i), {
      target: { value: 'jane@example.com' },
    });

    // Submit form
    fireEvent.click(screen.getByRole('button', { name: /create user/i }));

    // Verify success
    await waitFor(() => {
      expect(screen.getByText(/user created successfully/i)).toBeInTheDocument();
    });
  });
});
```

## Performance Optimization

### React Performance

#### Memoization

```typescript
// Memoize expensive calculations
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

#### Code Splitting

```typescript
// Lazy load heavy components
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

### Bundle Optimization

#### Analyze Bundle Size

```bash
# Analyze bundle size
pnpm analyze

# Output shows:
# - Largest modules
# - Duplicate dependencies
# - Optimization opportunities
```

#### Tree Shaking

```typescript
// Import only what you need
import { format } from 'date-fns';
// Instead of: import * as dateFns from 'date-fns';

// Use dynamic imports for conditional code
const loadHeavyLibrary = async () => {
  if (condition) {
    const { heavyFunction } = await import('./heavy-library');
    return heavyFunction;
  }
};
```

## Debugging

### Development Tools

#### React DevTools

- Install React Developer Tools browser extension
- Use Profiler to identify performance bottlenecks
- Inspect component props and state

#### TanStack Query DevTools

```typescript
// Enable in development
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      {children}
      {process.env.NODE_ENV === 'development' && (
        <ReactQueryDevtools initialIsOpen={false} />
      )}
    </QueryClientProvider>
  );
}
```

#### Console Debugging

```typescript
// Use structured logging
const logger = {
  info: (message: string, data?: any) => {
    if (process.env.NODE_ENV === 'development') {
      console.log(`[INFO] ${message}`, data);
    }
  },
  error: (message: string, error?: any) => {
    console.error(`[ERROR] ${message}`, error);
  },
  warn: (message: string, data?: any) => {
    console.warn(`[WARN] ${message}`, data);
  },
};

// Usage
logger.info('User data fetched', { userId, userData });
logger.error('Failed to save document', error);
```

## Deployment

### Build Process

```bash
# Production build
pnpm build

# Test production build locally
pnpm start

# Static export (if needed)
pnpm export
```

### Environment Configuration

```bash
# Production environment variables
NODE_ENV=production
NEXT_PUBLIC_API_URL=https://api.liyali.com
NEXT_PUBLIC_APP_URL=https://app.liyali.com
AUTH_SECRET=production-secret-key-32-characters-long
```

### Performance Monitoring

```typescript
// Core Web Vitals tracking
export function reportWebVitals(metric: any) {
  if (process.env.NODE_ENV === 'production') {
    // Send to analytics service
    analytics.track('web-vital', {
      name: metric.name,
      value: metric.value,
      id: metric.id,
    });
  }
}
```

## Best Practices Summary

### Code Quality

1. **Use TypeScript** for type safety
2. **Follow consistent naming** conventions
3. **Write self-documenting code** with clear variable names
4. **Keep functions small** and focused
5. **Use proper error handling** throughout the application

### Performance

1. **Memoize expensive operations** with useMemo and useCallback
2. **Lazy load components** that aren't immediately needed
3. **Optimize images** and assets
4. **Use proper caching strategies** for API data
5. **Monitor bundle size** and optimize imports

### Maintainability

1. **Write comprehensive tests** for critical functionality
2. **Document complex logic** with comments
3. **Use consistent patterns** across the codebase
4. **Refactor regularly** to improve code quality
5. **Keep dependencies up to date**

### Security

1. **Validate all user inputs** on both client and server
2. **Use proper authentication** and authorization
3. **Sanitize data** before displaying
4. **Keep secrets secure** and never commit them
5. **Regular security audits** of dependencies

This development guide provides the foundation for maintaining high code quality and developer productivity while building scalable, maintainable React applications.