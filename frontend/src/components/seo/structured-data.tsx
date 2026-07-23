"use client";

interface StructuredDataProps {
  data: Record<string, any>;
}

export const StructuredData = ({ data }: StructuredDataProps) => {
  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{
        __html: JSON.stringify(data),
      }}
    />
  );
};

// Predefined structured data schemas
export const organizationSchema = {
  "@context": "https://schema.org",
  "@type": "Organization",
  name: "Liyali Suite",
  url: process.env.NEXT_PUBLIC_APP_URL || "https://liyali.com",
  logo: `${process.env.NEXT_PUBLIC_APP_URL || "https://liyali.com"}/images/logo/logo-full.svg`,
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
};

export const softwareApplicationSchema = {
  "@context": "https://schema.org",
  "@type": "SoftwareApplication",
  name: "Liyali Suite",
  applicationCategory: "BusinessApplication",
  operatingSystem: "Web Browser",
  description:
    "Modern business operations platform for procurement, workflow automation, and team collaboration",
  url: process.env.NEXT_PUBLIC_APP_URL || "https://liyali.com",
  screenshot: `${process.env.NEXT_PUBLIC_APP_URL || "https://liyali.com"}/images/dashboard-screenshot.png`,
  offers: {
    "@type": "Offer",
    price: "999",
    priceCurrency: "USD",
    priceValidUntil: "2025-12-31",
    availability: "https://schema.org/InStock",
  },
  aggregateRating: {
    "@type": "AggregateRating",
    ratingValue: "4.8",
    ratingCount: "500",
    bestRating: "5",
    worstRating: "1",
  },
  author: {
    "@type": "Organization",
    name: "Liyali Suite",
  },
};

export const faqSchema = {
  "@context": "https://schema.org",
  "@type": "FAQPage",
  mainEntity: [
    {
      "@type": "Question",
      name: "What is Liyali Suite?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Liyali Suite is a comprehensive business operations platform that streamlines procurement, automates workflows, and enhances team collaboration for modern organizations.",
      },
    },
    {
      "@type": "Question",
      name: "How does Liyali Suite improve business efficiency?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Liyali Suite automates manual processes, provides real-time workflow approvals, enables smart procurement management, and offers advanced analytics to boost operational efficiency by up to 70%.",
      },
    },
    {
      "@type": "Question",
      name: "Is Liyali Suite suitable for small businesses?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Yes, Liyali Suite scales from startups to enterprise organizations with flexible pricing plans starting at $999/month for growing teams.",
      },
    },
    {
      "@type": "Question",
      name: "What security measures does Liyali Suite implement?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Liyali Suite implements bank-grade security with enterprise-level encryption, compliance standards, and 99.9% uptime guarantee to keep your sensitive data secure.",
      },
    },
    {
      "@type": "Question",
      name: "Does Liyali Suite offer a free trial?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Yes, Liyali Suite offers a 14-day free trial with no credit card required, allowing you to explore all features before making a commitment.",
      },
    },
  ],
};

export const breadcrumbSchema = (
  items: Array<{ name: string; url: string }>,
) => ({
  "@context": "https://schema.org",
  "@type": "BreadcrumbList",
  itemListElement: items.map((item, index) => ({
    "@type": "ListItem",
    position: index + 1,
    name: item.name,
    item: item.url,
  })),
});
