import { redirect } from "next/navigation";
import { verifySession } from "@/lib/auth";
import {
  Navbar,
  Hero,
  Features,
  HowItWorks,
  Pricing,
  About,
  Footer,
} from "@/components/landing-page";
import {
  StructuredData,
  softwareApplicationSchema,
  faqSchema,
} from "@/components/seo/structured-data";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Liyali Suite - Modern Business Operations Platform",
  description:
    "Transform your business operations with Liyali Suite. Streamline procurement, automate workflows, and boost team collaboration. Trusted by 500+ organizations worldwide. Start your free trial today.",
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
  ],
  openGraph: {
    title: "Liyali Suite - Modern Business Operations Platform",
    description:
      "Transform your business operations with Liyali Suite. Streamline procurement, automate workflows, and boost team collaboration.",
    type: "website",
    url: "/",
    images: [
      {
        url: "/images/og-image.png",
        width: 1200,
        height: 630,
        alt: "Liyali Suite Dashboard Preview",
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: "Liyali Suite - Modern Business Operations Platform",
    description:
      "Transform your business operations with Liyali Suite. Streamline procurement, automate workflows, and boost team collaboration.",
    images: ["/images/twitter-image.png"],
  },
  alternates: {
    canonical: "/",
  },
};

export default async function HomePage({
  searchParams,
}: {
  searchParams: Promise<{ [key: string]: string | string[] | undefined }>;
}) {
  // Check if user is authenticated
  const { isAuthenticated, session } = await verifySession();
  const params = await searchParams;

  // Allow authenticated users to view landing page if they explicitly request it
  const showLanding = params?.landing === "true" || params?.view === "landing";

  if (isAuthenticated && session?.user_id && !showLanding) {
    // User is authenticated and not explicitly requesting landing page - redirect based on their role
    const userRole = (session.role || "requester").toLowerCase();

    // Map roles to their respective pages (using lowercase to match backend)
    const roleRoutes: Record<string, string> = {
      admin: "/home",
      finance_officer: "/home",
      director: "/home",
      cfo: "/home",
      department_manager: "/home/requisitions",
      requester: "/requisitions",
      compliance_officer: "/admin/compliance/tracking",
    };

    const redirectUrl = roleRoutes[userRole] ?? "/home";
    redirect(redirectUrl);
  }

  // User is not authenticated OR explicitly requested landing page - show landing page
  return (
    <>
      {/* Structured Data for SEO */}
      <StructuredData data={softwareApplicationSchema} />
      <StructuredData data={faqSchema} />

      <main className="min-h-screen bg-slate-50">
        <Navbar isAuthenticated={isAuthenticated} />
        <Hero />
        <Features />
        <HowItWorks />
        <Pricing />
        <About />
        <Footer />
      </main>
    </>
  );
}
