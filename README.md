# Liyali Gateway

A modern procurement management system with approval workflows, document management, and comprehensive reporting.

---

## 🚀 Quick Start

```bash
# Backend
cd backend
cp .env.example .env
go run cmd/migrate/main.go
go run main.go  # http://localhost:8081

# Frontend
cd frontend
cp .env.example .env
npm install
npm run dev  # http://localhost:3000
```

---

## 📚 Documentation

- **[QUICK_REFERENCE.md](QUICK_REFERENCE.md)** - Quick commands and patterns
- **[DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md)** - Detailed development guide
- **[DEPLOYMENT_GUIDE.md](DEPLOYMENT_GUIDE.md)** - Complete deployment guide
- **[FEATURES_IMPLEMENTED.md](FEATURES_IMPLEMENTED.md)** - Feature documentation
- **[.kiro/specs/](.kiro/specs/)** - Feature specifications

---

## 🛠 Tech Stack

**Backend**: Go 1.21+ • Fiber v2 • PostgreSQL 15+  
**Frontend**: Next.js 16 • TypeScript 5+ • React Query • Tailwind CSS  
**Deployment**: Fly.io

---

## 📦 Key Features

- ✅ Admin Reports & Analytics (live data)
- ✅ Workflow Selection System
- ✅ Configuration Checklist
- ✅ Organization Logo Upload (ImageKit)
- ✅ Session Management & Auto-refresh
- ✅ PDF Generation
- ✅ Subscription Management
- ✅ Document Management (Requisitions, POs, PVs, GRNs, Budgets)

---

## 🏗 Project Structure

```
liyali-gateway/
├── backend/           # Go/Fiber backend
│   ├── handlers/     # HTTP endpoints
│   ├── services/     # Business logic
│   ├── repository/   # Database layer
│   └── models/       # Data structures
│
├── frontend/         # Next.js frontend
│   └── src/
│       ├── app/      # Pages & server actions
│       ├── components/ # UI components
│       ├── hooks/    # React Query hooks
│       └── types/    # TypeScript types
│
└── .kiro/specs/      # Feature specifications
```

---

## 🔐 Environment Variables

### Backend (.env)

```env
DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=require
JWT_SECRET=your-secret-key
APP_PORT=8081
FRONTEND_URL=https://your-frontend.com
```

### Frontend (.env)

```env
NEXT_PUBLIC_API_URL=https://your-backend.com
NEXT_PUBLIC_IMAGEKIT_PUBLIC_KEY=your_key
IMAGEKIT_PRIVATE_KEY=your_private_key
```

---

## 🧪 Testing

```bash
# Backend
cd backend
go test ./...

# Frontend
cd frontend
npm run build  # Type checking
npm run lint
```

---

## 🚢 Deployment

### Using Makefile (Recommended)

```bash
# Deploy all apps
make deploy

# Deploy individual apps
make deploy-backend    # Backend only
make deploy-web        # Web frontend only
make deploy-admin      # Admin console only

# Pre-deployment checks
make pre-deploy        # Verify env, build, test, migrate
```

### Fly.io (Direct)

```bash
cd backend && fly deploy
cd frontend && fly deploy
cd admin-console && fly deploy
```

### Manual

1. Build: `make build` or individual builds
2. Set environment variables
3. Run migrations: `make migrate`
4. Deploy and restart services

---

## 📖 Development Workflow

### Adding a New Feature

1. **Backend**: Model → Repository → Service → Handler → Route
2. **Frontend**: Type → Server Action → Hook → Component
3. **Database**: Create migration files
4. **Documentation**: Update relevant docs

See [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md) for detailed examples.

---

## 🤝 Contributing

1. Create feature branch: `git checkout -b feat/feature-name`
2. Make changes following code patterns
3. Test thoroughly
4. Commit with clear messages
5. Push and create PR

---

## 📝 License

Proprietary - All rights reserved

---

## 🆘 Support

- Check documentation in root directory
- Review `.kiro/specs/` for feature details
- Check existing code for examples

---

**Last Updated**: February 23, 2026
