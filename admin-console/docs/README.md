# Liyali Admin Console

## Overview

Administrative portal for managing the Liyali Gateway system.

## Features

### User Management

- View and manage all users
- Edit user details and roles
- Suspend/activate accounts
- Reset passwords

### Organization Management

- View all organizations
- Manage subscriptions
- Monitor usage and limits
- Organization settings

### Subscription Management

- View all subscriptions
- Manage plans and features
- Track usage and billing
- Trial management

### System Settings

- Configure system-wide settings
- Environment variables
- Feature flags
- Security settings

### Analytics

- System metrics
- User activity
- Subscription analytics
- Performance monitoring

## Getting Started

### Installation

```bash
cd admin-console
npm install
```

### Development

```bash
npm run dev
```

Access at: http://localhost:3001

### Build

```bash
npm run build
npm start
```

## Configuration

### Environment Variables

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_APP_NAME=Liyali Admin Console
```

## Authentication

Admin console requires admin-level authentication:

1. Login with admin credentials
2. Token stored in localStorage
3. Auto-refresh on expiry
4. Secure API communication

## Database Integration

✅ All data fetched from database via API  
✅ No mock data  
✅ Real-time updates  
✅ Type-safe with TypeScript

## Tech Stack

- **Framework**: Next.js 14
- **UI**: Tailwind CSS + shadcn/ui
- **State**: React Query
- **Forms**: React Hook Form
- **Validation**: Zod
- **Icons**: Lucide React

## Project Structure

```
admin-console/
├── src/
│   ├── app/
│   │   ├── admin/          # Admin pages
│   │   ├── login/          # Auth pages
│   │   └── _actions/       # Server actions
│   ├── components/
│   │   ├── layout/         # Layout components
│   │   └── ui/             # UI components
│   ├── hooks/              # Custom hooks
│   ├── lib/                # Utilities
│   └── types/              # TypeScript types
└── public/                 # Static assets
```

## Development

### Adding New Features

1. Create page in `src/app/admin/`
2. Add server actions in `src/app/_actions/`
3. Create components in `src/components/`
4. Add types in `src/types/`

### Code Style

- Use TypeScript
- Follow ESLint rules
- Use Prettier for formatting
- Component naming: PascalCase
- File naming: kebab-case

## Deployment

### Build

```bash
npm run build
```

### Docker

```bash
docker build -t admin-console .
docker run -p 3001:3001 admin-console
```

## Troubleshooting

### API Connection Issues

- Check `NEXT_PUBLIC_API_URL` is correct
- Verify backend is running
- Check CORS settings

### Authentication Issues

- Clear localStorage
- Check token expiry
- Verify admin permissions

### Build Errors

- Clear `.next` folder
- Delete `node_modules` and reinstall
- Check TypeScript errors: `npx tsc --noEmit`

## Support

For issues or questions, contact the development team.
