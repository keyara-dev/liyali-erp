import type { Metadata } from "next";
// import { Inter } from "next/font/google";
import "./globals.css";
import { Providers } from "./providers";

// const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: {
    default: "Liyali Suite - Modern Business Operations Platform",
    template: "%s | Liyali Suite",
  },
  description:
    "Streamline your business operations with Liyali Suite - the all-in-one platform for procurement, workflow automation, and team collaboration. Trusted by 500+ organizations worldwide.",
  keywords: [
    "business operations platform",
    "procurement software",
    "workflow automation",
    "business process management",
    "enterprise software",
    "procurement management",
    "approval workflows",
    "business efficiency",
    "digital transformation",
    "operational excellence",
    "procurement automation",
    "business workflow software",
    "enterprise procurement",
    "business operations software",
    "procurement platform",
  ],
  authors: [{ name: "Liyali Suite Team" }],
  creator: "Liyali Suite",
  publisher: "Liyali Suite",
  formatDetection: {
    email: false,
    address: false,
    telephone: false,
  },
  metadataBase: new URL(
    process.env.NEXT_PUBLIC_APP_URL || "https://liyali.com",
  ),
  alternates: {
    canonical: "/",
  },
  openGraph: {
    type: "website",
    locale: "en_US",
    url: "/",
    title: "Liyali Suite - Modern Business Operations Platform",
    description:
      "Streamline your business operations with Liyali Suite - the all-in-one platform for procurement, workflow automation, and team collaboration.",
    siteName: "Liyali Suite",
    images: [
      {
        url: "/images/og-image.png",
        width: 1200,
        height: 630,
        alt: "Liyali Suite - Business Operations Platform",
        type: "image/png",
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: "Liyali Suite - Modern Business Operations Platform",
    description:
      "Streamline your business operations with Liyali Suite - the all-in-one platform for procurement, workflow automation, and team collaboration.",
    images: ["/images/twitter-image.png"],
    creator: "@liyalisuite",
    site: "@liyalisuite",
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      "max-video-preview": -1,
      "max-image-preview": "large",
      "max-snippet": -1,
    },
  },
  verification: {
    google: process.env.GOOGLE_SITE_VERIFICATION,
    yandex: process.env.YANDEX_VERIFICATION,
    yahoo: process.env.YAHOO_VERIFICATION,
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <meta name="apple-mobile-web-app-title" content="Liyali Suite" />
        <meta name="application-name" content="Liyali Suite" />
        <meta name="apple-mobile-web-app-capable" content="yes" />
        <meta name="apple-mobile-web-app-status-bar-style" content="default" />
        <meta name="mobile-web-app-capable" content="yes" />
        <meta name="theme-color" content="#0c54e7" />

        {/* Preconnect to external domains */}
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link
          rel="preconnect"
          href="https://fonts.gstatic.com"
          crossOrigin="anonymous"
        />
        <link rel="preconnect" href="https://cdnjs.cloudflare.com" />

        {/* DNS prefetch for performance */}
        <link rel="dns-prefetch" href="https://fonts.googleapis.com" />
        <link rel="dns-prefetch" href="https://fonts.gstatic.com" />
        <link rel="dns-prefetch" href="https://cdnjs.cloudflare.com" />

        {/* Font Awesome with optimized loading */}
        <link
          rel="stylesheet"
          href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css"
          integrity="sha512-iecdLmaskl7CVkqkXNQ/ZH/XLlvWZOJyj7Yy7tcenmpD1ypASozpmT/E0iPtmFIB46ZmdtAc9eNBvH0H/ZpiBw=="
          crossOrigin="anonymous"
          referrerPolicy="no-referrer"
        />

        {/* Structured Data for Organization */}
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{
            __html: JSON.stringify({
              "@context": "https://schema.org",
              "@type": "Organization",
              name: "Liyali Suite",
              url: process.env.NEXT_PUBLIC_APP_URL || "https://liyali.com",
              logo: `${process.env.NEXT_PUBLIC_APP_URL || "https://liyali.com"}/icon?<generated>`,
              description:
                "Modern business operations platform for procurement, workflow automation, and team collaboration",
              foundingDate: "2023",
              contactPoint: {
                "@type": "ContactPoint",
                contactType: "customer service",
                email: "support@liyali.com",
              },
              sameAs: [
                "https://twitter.com/liyalisuite",
                "https://linkedin.com/company/liyali-suite",
              ],
            }),
          }}
        />
      </head>
      <body>
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
