# ImageKit Quick Start Guide

## 🚀 Get Started in 3 Steps

### Step 1: Get ImageKit Credentials (2 minutes)

1. Sign up at https://imagekit.io/
2. Go to **Developer Options** → **API Keys**
3. Copy these three values:
   - **Public Key**: `public_xxx...`
   - **Private Key**: `private_xxx...`
   - **URL Endpoint**: `https://ik.imagekit.io/your_id`

### Step 2: Add to Environment (1 minute)

Create/edit `frontend/.env.local`:

```env
NEXT_PUBLIC_IMAGEKIT_PUBLIC_KEY=public_xxx
NEXT_PUBLIC_IMAGEKIT_URL_ENDPOINT=https://ik.imagekit.io/your_id
IMAGEKIT_PRIVATE_KEY=private_xxx
NEXT_PUBLIC_IMAGEKIT_AUTH_ENDPOINT=/api/imagekit-auth
```

### Step 3: Restart Server (30 seconds)

```bash
cd frontend
npm run dev
```

## ✅ Test It

1. Go to http://localhost:3000/welcome
2. Click "Create workspace"
3. Upload a logo
4. Create the workspace
5. See your logo in the sidebar! 🎉

## 📖 Usage Examples

### Add to Settings Page

```tsx
import { OrganizationLogoSection } from "@/components/organization/organization-logo-section";

<OrganizationLogoSection
  organizationId={org.id}
  organizationName={org.name}
  currentLogoUrl={org.logoUrl}
/>;
```

### Display Logo

```tsx
import { OrganizationAvatar } from "@/components/ui/organization-avatar";

<OrganizationAvatar name={org.name} logoUrl={org.logoUrl} size="md" />;
```

## 📚 Full Documentation

- **Setup**: `frontend/docs/IMAGEKIT_SETUP.md`
- **Examples**: `frontend/docs/LOGO_UPLOAD_USAGE_EXAMPLES.md`
- **Testing**: `frontend/docs/IMAGEKIT_TESTING_GUIDE.md`
- **Implementation**: `ORGANIZATION_LOGO_IMPLEMENTATION.md`

## 💰 Cost

**Free tier includes:**

- 20GB bandwidth/month
- 20GB storage
- Unlimited transformations

Perfect for most applications!

## 🔒 Security

- ✅ Private key never exposed
- ✅ Secure token-based uploads
- ✅ File validation
- ✅ Automatic optimization

## ❓ Troubleshooting

**Upload fails?**

- Check environment variables are set correctly
- Restart dev server after adding variables
- Check browser console for errors

**Images don't show?**

- Verify URL endpoint is correct
- Check logoUrl is saved in database
- Clear browser cache

## 🎯 What's Included

- ✅ Upload component with drag-and-drop
- ✅ Display component with optimization
- ✅ Ready-to-use settings section
- ✅ Automatic image optimization
- ✅ Progress tracking
- ✅ File validation
- ✅ Security built-in

## 🚀 Next Steps

1. Set up ImageKit account ← **Start here**
2. Add environment variables
3. Test the upload
4. Add to your settings pages
5. Enjoy! 🎉

---

**Need help?** Check the full documentation in `frontend/docs/`
