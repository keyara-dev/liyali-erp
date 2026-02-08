# SEO & Metadata

## Overview

The frontend implements comprehensive SEO optimization following Next.js 14+ conventions.

## Icon Files (Auto-Detected)

Next.js automatically detects and handles these files:

```
src/app/
├── icon.tsx              # Favicon (32x32)
├── apple-icon.tsx        # Apple touch icon (180x180)
├── icon-192.tsx          # PWA icon (192x192)
├── icon-512.tsx          # PWA icon (512x512)
├── opengraph-image.tsx   # OG image (1200x630)
└── twitter-image.tsx     # Twitter card (1200x675)
```

## Metadata Files

```
src/app/
├── manifest.ts           # PWA manifest
├── robots.ts             # Robots.txt
└── sitemap.ts            # XML sitemap
```

## Usage

### Page-Level SEO (Server Components)

```tsx
import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Page Title",
  description: "Page description",
  openGraph: {
    title: "Page Title",
    description: "Page description",
  },
};
```

### Client Components

```tsx
import { PageSEO } from "@/components/seo/page-seo";

export default function Page() {
  return (
    <>
      <PageSEO title="Page Title" description="Page description" />
      {/* Content */}
    </>
  );
}
```

## Testing

```bash
# Start dev server
npm run dev

# Test routes:
http://localhost:3000/icon
http://localhost:3000/apple-icon
http://localhost:3000/manifest.json
http://localhost:3000/sitemap.xml
http://localhost:3000/robots.txt
http://localhost:3000/opengraph-image
http://localhost:3000/twitter-image
```

## Social Media Testing

- [Facebook Debugger](https://developers.facebook.com/tools/debug/)
- [Twitter Card Validator](https://cards-dev.twitter.com/validator)
- [LinkedIn Inspector](https://www.linkedin.com/post-inspector/)

## Environment Variables

```env
NEXT_PUBLIC_APP_URL=https://liyali.com
GOOGLE_SITE_VERIFICATION=your-code
```

## Resources

- [Next.js Metadata](https://nextjs.org/docs/app/building-your-application/optimizing/metadata)
- [Next.js App Icons](https://nextjs.org/docs/app/api-reference/file-conventions/metadata/app-icons)
