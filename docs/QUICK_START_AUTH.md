# Quick Start - Authentication System

## Login to the Application

1. **Start the development server**
   ```bash
   pnpm dev
   ```

2. **Navigate to login page**
   - Go to: http://localhost:3000/login

3. **Choose a demo account**

   **Option A: Use the Login Form**
   - Email: `admin@liyali.com`
   - Password: `password123`
   - Click "Sign In"

   **Option B: Use Quick Demo Buttons**
   - Click any demo user button (👤 Requester, ⚙️ Admin, etc.)
   - Automatically logged in and redirected

4. **You're logged in!**
   - Dashboard loads automatically
   - See your user info in the sidebar

## Available Demo Accounts

| Account | Email | Password | Best For |
|---------|-------|----------|----------|
| 👤 Requester | requester@liyali.com | password123 | Creating requisitions |
| 👥 Manager | manager@liyali.com | password123 | Approving requests |
| 💼 Finance | finance@liyali.com | password123 | Finance operations |
| 👔 Director | director@liyali.com | password123 | Director approvals |
| 💎 CFO | cfo@liyali.com | password123 | CFO level actions |
| ✅ Compliance | compliance@liyali.com | password123 | Compliance & monitoring |
| ⚙️ Admin | admin@liyali.com | password123 | Full admin access |

## What Each Role Can Do

### Requester
- ✓ View dashboard
- ✓ Search transactions
- ✓ Create requisitions
- ✓ View QR codes

### Manager
- ✓ All requester features
- ✓ Approve/reject items

### Finance Officer
- ✓ All requester features
- ✓ Financial operations

### Director
- ✓ All requester features
- ✓ Executive decisions

### CFO
- ✓ All requester features
- ✓ CFO-level approvals

### Compliance Officer
- ✓ All requester features
- ✓ View reports & analytics
- ✓ View activity logs
- ✓ Track compliance
- ✓ System monitoring

### Admin
- ✓ **All features**
- ✓ Manage users
- ✓ View all reports
- ✓ Manage system settings

## Key Features by Role

| Feature | Public | Requester | Compliance | Admin |
|---------|--------|-----------|-----------|-------|
| Dashboard | ✗ | ✓ | ✓ | ✓ |
| Search | ✗ | ✓ | ✓ | ✓ |
| Create Requisition | ✗ | ✓ | ✓ | ✓ |
| Reports | ✗ | ✗ | ✓ | ✓ |
| Manage Users | ✗ | ✗ | ✗ | ✓ |
| Activity Logs | ✗ | ✗ | ✓ | ✓ |
| Compliance | ✗ | ✗ | ✓ | ✓ |
| Monitoring | ✗ | ✗ | ✓ | ✓ |
| QR Verification | ✗ | ✓ | ✓ | ✓ |

## Logging Out

1. **Click your avatar** in the top right corner
2. **Select "Logout"** from the dropdown menu
3. **Redirected to login page**

## Session Details

- **Duration**: 24 hours
- **Automatic Logout**: After 24 hours of inactivity
- **Logout on Browser Close**: No (uses cookies)
- **Remember Me**: Not available (disabled for security)

## Testing Different Roles

### Test as Admin
1. Go to `/login`
2. Click ⚙️ Admin button
3. Access `/admin/users` to manage users
4. Access `/admin/reports` to view reports

### Test as Requester
1. Go to `/login`
2. Click 👤 Requester button
3. Try accessing `/admin/users` → You'll be redirected
4. Access `/workflows/requisitions/create` → Works ✓

### Test as Compliance Officer
1. Go to `/login`
2. Click ✅ Compliance button
3. Access `/compliance/tracking` → Works ✓
4. Access `/admin/users` → You'll be redirected

## Debugging

### Check Current Session
Open browser DevTools → Application → Cookies → Look for `auth_session`

### Check User Info
In server components:
```typescript
const user = await getCurrentUser()
console.log(user) // See user details
```

### Check Role
```typescript
const isAdmin = await isAdmin()
console.log(isAdmin) // true or false
```

## Common Issues

### Issue: Can't Log In
**Solution:**
- Check email spelling (lowercase: `admin@liyali.com`)
- Verify password is exactly `password123`
- Clear browser cookies and try again

### Issue: Logged In But Redirected to Login
**Solution:**
- Session expired (after 24 hours)
- Cookie was cleared
- Log in again

### Issue: Can Access Admin Page as Non-Admin
**Solution:**
- This shouldn't happen (page checks role)
- Clear cookies and log back in
- Check browser console for errors

### Issue: Lost Session After Page Refresh
**Solution:**
- Session should persist (stored in cookies)
- Check if cookies are enabled in browser
- Not a normal behavior - report as bug

## Next Steps

1. **Explore the Application**
   - Try different roles
   - Test each feature
   - Create requisitions

2. **Understand the System**
   - Review [AUTH_SYSTEM.md](AUTH_SYSTEM.md) for technical details
   - Check [IMPLEMENTATION_GUIDE.md](IMPLEMENTATION_GUIDE.md) for feature overview

3. **Make It Production-Ready**
   - See "Migration to Production" in [AUTH_SYSTEM.md](AUTH_SYSTEM.md)
   - Add proper database
   - Implement password hashing
   - Add security features

## More Information

- **Full Documentation**: See [AUTH_SYSTEM.md](AUTH_SYSTEM.md)
- **Implementation Guide**: See [IMPLEMENTATION_GUIDE.md](IMPLEMENTATION_GUIDE.md)
- **Source Code**: See `src/lib/auth.ts` and `src/auth.ts`
