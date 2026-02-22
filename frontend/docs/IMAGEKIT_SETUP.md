# ImageKit Integration Setup

This document explains how to set up ImageKit for organization logo uploads in the Liyali application.

## Overview

ImageKit is used as the CDN and image storage solution for organization logos. It provides:

- Fast image uploads
- Automatic image optimization
- Real-time image transformations
- CDN delivery

## Setup Instructions

### 1. Create an ImageKit Account

1. Go to [ImageKit.io](https://imagekit.io/) and sign up for a free account
2. After signing up, you'll get access to your dashboard

### 2. Get Your Credentials

From your ImageKit dashboard, you'll need three pieces of information:

1. **Public Key**: Found in Developer Options → API Keys
2. **Private Key**: Found in Developer Options → API Keys (keep this secret!)
3. **URL Endpoint**: Found in Developer Options → URL-endpoint (looks like `https://ik.imagekit.io/your_imagekit_id`)

### 3. Configure Environment Variables

Add the following to your `.env.local` file:

```env
# ImageKit Configuration
NEXT_PUBLIC_IMAGEKIT_PUBLIC_KEY=your_public_key_here
NEXT_PUBLIC_IMAGEKIT_URL_ENDPOINT=https://ik.imagekit.io/your_imagekit_id
IMAGEKIT_PRIVATE_KEY=your_private_key_here
NEXT_PUBLIC_IMAGEKIT_AUTH_ENDPOINT=/api/imagekit-auth
```

**Important**:

- The `IMAGEKIT_PRIVATE_KEY` should NEVER be exposed to the client
- Only `NEXT_PUBLIC_*` variables are accessible in the browser
- The private key is only used in the `/api/imagekit-auth` endpoint

### 4. Test the Integration

1. Start your development server:

   ```bash
   npm run dev
   ```

2. Navigate to the workspace creation page or organization settings
3. Try uploading a logo image
4. Verify the image appears correctly

## File Structure

```
frontend/
├── src/
│   ├── lib/
│   │   └── imagekit.ts                    # ImageKit utilities and upload logic
│   ├── components/
│   │   └── ui/
│   │       ├── organization-logo-upload.tsx  # Logo upload component
│   │       └── organization-avatar.tsx       # Display component with optimization
│   └── app/
│       └── api/
│           └── imagekit-auth/
│               └── route.ts                  # Authentication endpoint
```

## Usage

### Upload Component

```tsx
import { OrganizationLogoUpload } from "@/components/ui/organization-logo-upload";

function MyComponent() {
  const [logoUrl, setLogoUrl] = useState("");

  return (
    <OrganizationLogoUpload
      currentLogoUrl={logoUrl}
      organizationName="My Organization"
      onLogoChange={setLogoUrl}
      size="md"
    />
  );
}
```

### Display Component

```tsx
import { OrganizationAvatar } from "@/components/ui/organization-avatar";

function MyComponent() {
  return (
    <OrganizationAvatar
      name="My Organization"
      logoUrl="https://ik.imagekit.io/..."
      size="md"
    />
  );
}
```

## Features

### Automatic Image Optimization

The `OrganizationAvatar` component automatically optimizes images using ImageKit transformations:

- Resizes to appropriate dimensions based on size prop
- Converts to optimal format (WebP when supported)
- Applies quality compression (80%)
- Maintains aspect ratio

### Upload Validation

The upload component validates:

- File type (JPG, PNG, GIF, WebP only)
- File size (max 10MB)
- Provides user-friendly error messages

### Progress Tracking

Upload progress is displayed in real-time with a progress bar and percentage indicator.

## Folder Structure in ImageKit

Uploaded logos are organized in the following folder structure:

- `/organizations/` - All organization logos

You can customize this in the `uploadToImageKit` function in `lib/imagekit.ts`.

## Security

- Private key is never exposed to the client
- Authentication tokens expire after 1 hour
- Signature-based authentication prevents unauthorized uploads
- All uploads go through the backend authentication endpoint

## Troubleshooting

### Upload fails with "Authentication failed"

- Check that your `IMAGEKIT_PRIVATE_KEY` is correct
- Verify the `/api/imagekit-auth` endpoint is accessible
- Check browser console for detailed error messages

### Images not displaying

- Verify the `NEXT_PUBLIC_IMAGEKIT_URL_ENDPOINT` is correct
- Check that the image URL is properly stored in the database
- Ensure CORS is configured in ImageKit dashboard (usually automatic)

### Slow uploads

- Check your internet connection
- Verify ImageKit service status
- Consider reducing image size before upload

## Cost Considerations

ImageKit free tier includes:

- 20GB bandwidth per month
- 20GB storage
- Unlimited image transformations

For production use, monitor your usage and upgrade if needed.

## Additional Resources

- [ImageKit Documentation](https://docs.imagekit.io/)
- [ImageKit Next.js Integration](https://imagekit.io/docs/integration/nextjs)
- [ImageKit API Reference](https://docs.imagekit.io/api-reference/api-introduction)
