# SEO Implementation Guide

## Overview

This document outlines the SEO optimization and favicon implementation for the Liyali Suite platform.

## Implemented Features

### 1. Metadata Configuration

#### Frontend App (`frontend/src/app/layout.tsx`)

- ✅ Comprehensive metadata with title templates
- ✅ Open Graph tags for social sharing
- ✅ Twitter Card configuration
- ✅ Structured data (JSON-LD) for Organization
- ✅ Favicon and icon configuration
- ✅ Robots meta tags
- ✅ Site verification tags (Google, Yandex, Yahoo)

#### Admin Console (`admin-console/src/app/layout.tsx`)

- ✅ Basic metadata configuration
- ✅ Favicon setup
- ✅ Robots noindex (admin pages shouldn't be indexed)

### 2. Dynamic Meta Files

#### Frontend

- `manifest.ts` - PWA manifest configuration
- `robots.ts` - Robots.txt generation
- `sitemap.ts` - XML sitemap generation
- `opengraph-image.tsx` - Dynamic OG image generation
- `twitter-image.tsx` - Dynamic Twitter card image
- `icon.tsx` - Dynamic favicon generation
- `apple-icon.tsx` - Apple touch icon generation

#### Admin Console

- `icon.tsx` - Admin favicon (red "A")
- `apple-icon.tsx` - Admin apple touch icon

### 3. SEO Components

#### PageSEO Component (`frontend/src/components/seo/page-seo.tsx`)

Client-side component for dynamic meta tag updates in client components.

**Usage:**

```tsx
import { PageSEO } from "@/components/seo/page-seo";

export default function MyPage() {
  return (
    <>
      <PageSEO
        title="Page Title"
        description="Page description"
        image="/images/custom-og-image.png"
      />
      {/* Page content */}
    </>
  );
}
```

### 4. Favicon Setup

#### Automatic Generation

Next.js 14+ automatically generates favicons from:

- `app/icon.tsx` - Main favicon (32x32)
- `app/apple-icon.tsx` - Apple touch icon (180x180)

#### Manual Generation (Optional)

Use the provided script to generate PNG icons from SVG:

```bash
# Install dependencies
npm install sharp

# Run generation script
node scripts/generate-icons.js
```

This creates:

- `icon-16.png`, `icon-32.png`, `icon-192.png`, `icon-512.png`
- `apple-touch-icon.png`
- `favicon.ico`

### 5. Open Graph Images

#### Dynamic Generation

Next.js automatically generates OG images from:

- `app/opengraph-image.tsx` (1200x630)
- `app/twitter-image.tsx` (1200x675)

#### Manual Creation (Optional)

1. Open `scripts/og-image-template.html` in browser
2. Screenshot the templates at exact dimensions
3. Save as `frontend/public/images/og-image.png` and `twitter-image.png`

Or use design tools:

- Figma
- Canva
- Adobe Photoshop
- Online generators: [OpenGraph.xyz](https://www.opengraph.xyz/)

## Testing

### 1. Local Testing

```bash
# Start development server
npm run dev

# Check metadata
# Visit: http://localhost:3000
# Open DevTools > Application > Manifest
```

### 2. Favicon Testing

Test favicons on:

- Chrome DevTools (Application > Manifest)
- [Real Favicon Generator Checker](https://realfavicongenerator.net/favicon_checker)
- Multiple browsers (Chrome, Firefox, Safari, Edge)
- Mobile devices (iOS, Android)

### 3. Social Media Preview Testing

Test link previews on:

- [Facebook Sharing Debugger](https://developers.facebook.com/tools/debug/)
- [Twitter Card Validator](https://cards-dev.twitter.com/validator)
- [LinkedIn Post Inspector](https://www.linkedin.com/post-inspector/)
- [Open Graph Debugger](https://www.opengraph.xyz/)

### 4. SEO Testing Tools

- [Google Search Console](https://search.google.com/search-console)
- [Google Rich Results Test](https://search.google.com/test/rich-results)
- [Lighthouse](https://developers.google.com/web/tools/lighthouse) (in Chrome DevTools)
- [PageSpeed Insights](https://pagespeed.web.dev/)

## Best Practices

### Page-Level SEO

For server components (recommended):

```tsx
import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Page Title",
  description: "Page description",
  openGraph: {
    title: "Page Title",
    description: "Page description",
    images: ["/images/page-og-image.png"],
  },
};

export default function Page() {
  return <div>Content</div>;
}
```

For client components:

```tsx
"use client";
import { PageSEO } from "@/components/seo/page-seo";

export default function Page() {
  return (
    <>
      <PageSEO title="Page Title" description="Page description" />
      <div>Content</div>
    </>
  );
}
```

### Dynamic Routes

```tsx
import { Metadata } from "next";

type Props = {
  params: { id: string };
};

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const data = await fetchData(params.id);

  return {
    title: data.title,
    description: data.description,
    openGraph: {
      title: data.title,
      description: data.description,
      images: [data.image],
    },
  };
}
```

### Structured Data

Add JSON-LD structured data for rich snippets:

```tsx
<script
  type="application/ld+json"
  dangerouslySetInnerHTML={{
    __html: JSON.stringify({
      "@context": "https://schema.org",
      "@type": "Product",
      name: "Product Name",
      description: "Product description",
      image: "https://example.com/image.jpg",
      offers: {
        "@type": "Offer",
        price: "99.99",
        priceCurrency: "USD",
      },
    }),
  }}
/>
```

## Environment Variables

Add to `.env.local`:

```env
# SEO Configuration
NEXT_PUBLIC_APP_URL=https://liyali.com
GOOGLE_SITE_VERIFICATION=your-verification-code
YANDEX_VERIFICATION=your-verification-code
YAHOO_VERIFICATION=your-verification-code
```

## Checklist

### Pre-Launch

- [ ] All meta tags configured
- [ ] Favicons generated and tested
- [ ] OG images created
- [ ] Sitemap accessible at `/sitemap.xml`
- [ ] Robots.txt accessible at `/robots.txt`
- [ ] Manifest accessible at `/manifest.json`
- [ ] Structured data validated
- [ ] Mobile-friendly test passed
- [ ] Page speed optimized (>90 score)

### Post-Launch

- [ ] Submit sitemap to Google Search Console
- [ ] Submit sitemap to Bing Webmaster Tools
- [ ] Verify site ownership
- [ ] Monitor search console for errors
- [ ] Test social media previews
- [ ] Set up Google Analytics
- [ ] Configure canonical URLs
- [ ] Monitor Core Web Vitals

## Resources

### Documentation

- [Next.js Metadata](https://nextjs.org/docs/app/building-your-application/optimizing/metadata)
- [Next.js SEO](https://nextjs.org/learn/seo/introduction-to-seo)
- [Google Search Central](https://developers.google.com/search)

### Tools

- [Real Favicon Generator](https://realfavicongenerator.net/)
- [Meta Tags](https://metatags.io/)
- [Schema.org](https://schema.org/)
- [JSON-LD Generator](https://technicalseo.com/tools/schema-markup-generator/)

### Testing

- [Google Mobile-Friendly Test](https://search.google.com/test/mobile-friendly)
- [Google Rich Results Test](https://search.google.com/test/rich-results)
- [Facebook Debugger](https://developers.facebook.com/tools/debug/)
- [Twitter Card Validator](https://cards-dev.twitter.com/validator)

## Troubleshooting

### Favicon Not Updating

1. Clear browser cache (Ctrl+Shift+Delete)
2. Hard refresh (Ctrl+F5)
3. Check browser DevTools > Application > Storage
4. Verify file exists at `/favicon.ico`
5. Check Next.js build output

### OG Image Not Showing

1. Verify image dimensions (1200x630 for OG, 1200x675 for Twitter)
2. Check image file size (<1MB recommended)
3. Use absolute URLs in meta tags
4. Clear social media cache:
   - Facebook: Use Sharing Debugger
   - Twitter: Use Card Validator
   - LinkedIn: Use Post Inspector

### Sitemap Not Accessible

1. Check `app/sitemap.ts` exists
2. Verify build completed successfully
3. Test locally: `http://localhost:3000/sitemap.xml`
4. Check server configuration (nginx, Apache)

### Metadata Not Updating

1. Ensure using server components for static metadata
2. Use `generateMetadata` for dynamic routes
3. Check metadata precedence (page > layout > root)
4. Verify no client-side overrides
5. Clear Next.js cache: `rm -rf .next`

## Support

For issues or questions:

- Check Next.js documentation
- Review implementation in this repository
- Contact development team
