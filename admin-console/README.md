# Liyali Admin Console

Administrative portal for managing the Liyali Gateway system. Built with Next.js 15, React 19, and Shadcn/ui components.

## Features

- **Dashboard Overview**: System health, metrics, and recent activity
- **Organization Management**: View and manage all organizations
- **Trial Management**: Reset trial periods, monitor expiring trials
- **User Management**: Admin users, roles, and permissions
- **System Monitoring**: Analytics, audit logs, API monitoring
- **Configuration**: System settings, feature flags, notifications

## Getting Started

### Prerequisites

- Node.js 18+
- npm, yarn, or pnpm

### Installation

1. Install dependencies:

```bash
npm install
# or
yarn install
# or
pnpm install
```

2. Copy environment variables:

```bash
cp .env.example .env
```

3. Update the API URL in `.env`:

```
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### Development

Run the development server:

```bash
npm run dev
# or
yarn dev
# or
pnpm dev
```

The admin console will be available at [http://localhost:3001](http://localhost:3001).

### Build

Build for production:

```bash
npm run build
npm start
```

## Project Structure

```
admin-console/
├── src/
│   ├── app/                    # Next.js app router
│   │   ├── admin/             # Admin routes
│   │   │   ├── dashboard/     # Dashboard page
│   │   │   ├── organizations/ # Organization management
│   │   │   ├── trial-management/ # Trial reset functionality
│   │   │   └── layout.tsx     # Admin layout
│   │   ├── globals.css        # Global styles
│   │   └── layout.tsx         # Root layout
│   ├── components/            # Reusable components
│   │   ├── ui/               # Shadcn/ui components
│   │   └── layout/           # Layout components
│   └── lib/                  # Utilities and configuration
│       ├── utils.ts          # Utility functions
│       └── routes-config.tsx # Navigation configuration
├── package.json
├── tailwind.config.js
└── next.config.ts
```

## Key Features

### Subscription Management

The subscription management page (`/admin/subscriptions`) provides comprehensive control over:

#### **Subscription Tiers Tab**

- Create, edit, and delete subscription tiers (Basic, Professional, Enterprise, etc.)
- Set pricing (monthly/yearly), user limits, and storage quotas
- Assign features to each tier using checkboxes
- Activate/deactivate tiers
- Visual tier comparison with feature lists

#### **Features Management Tab**

- Define available features for subscription tiers
- Organize features by categories (Core, Advanced, Integrations, etc.)
- Enable/disable features globally
- Bulk feature assignment to tiers

#### **Trial Management Tab** (Key Requirement)

- Overview of active, expiring, and expired trials
- One-click trial reset functionality with audit trail
- Search and filter organizations by trial status
- Integration with backend API endpoint: `POST /api/v1/organizations/{id}/trial/reset`
- Proper error handling and success notifications

#### **Analytics Tab**

- Revenue metrics (MRR, ARR, ARPU, LTV)
- Subscription tier distribution
- Trial conversion rates and analytics
- Churn rate monitoring
- Growth trends visualization

### Organization Management

The organizations page (`/admin/organizations`) shows:

- Complete list of all organizations
- Status indicators (active, suspended, trial, subscribed)
- User counts and subscription tiers
- Quick actions for viewing and editing

### Admin Authentication

The admin console integrates with the existing backend authentication system and requires admin-level permissions to access.

## API Integration

The admin console communicates with the Liyali Gateway backend API:

- Base URL: Configured via `NEXT_PUBLIC_API_URL`
- Authentication: Uses existing JWT token system with admin-specific session management
- **Subscription Management APIs:**
  - `GET /api/v1/subscriptions/tiers` - Get all subscription tiers
  - `POST /api/v1/subscriptions/tiers` - Create new tier
  - `PUT /api/v1/subscriptions/tiers/{id}` - Update tier
  - `DELETE /api/v1/subscriptions/tiers/{id}` - Delete tier
  - `GET /api/v1/subscriptions/features` - Get all features
  - `POST /api/v1/subscriptions/features` - Create new feature
  - `PUT /api/v1/subscriptions/features/{id}` - Update feature
  - `DELETE /api/v1/subscriptions/features/{id}` - Delete feature
- **Trial Management APIs:**
  - `POST /api/v1/organizations/{id}/trial/reset` - Reset trial period
  - `GET /api/v1/organizations?filter=trial` - Get trial organizations
- **Analytics APIs:**
  - `GET /api/v1/subscriptions/analytics` - Get subscription analytics
- **Organization APIs:**
  - `GET /api/v1/organizations` - Get all organizations

## Deployment

The admin console can be deployed alongside the main application or as a separate service:

1. Build the application: `npm run build`
2. Deploy to your hosting platform
3. Ensure the API URL environment variable points to your backend
4. Configure admin authentication and authorization

## Development Notes

- Built using the same UI patterns as the main frontend application
- Uses Shadcn/ui component library for consistency
- Responsive design works on desktop and mobile
- TypeScript for type safety
- Tailwind CSS for styling
