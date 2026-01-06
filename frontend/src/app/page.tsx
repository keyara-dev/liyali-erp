import { redirect } from "next/navigation";
import { verifySession } from "@/lib/auth";
import { 
  Navbar, 
  Hero, 
  Features, 
  HowItWorks, 
  Pricing, 
  About, 
  Footer 
} from "@/components/landing-page";

export default async function HomePage() {
  // Check if user is authenticated
  const { isAuthenticated, session } = await verifySession();

  if (isAuthenticated && session?.user_id) {
    // User is authenticated - redirect based on their role
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

  // User is not authenticated - show landing page
  return (
    <div className="min-h-screen bg-slate-50">
      <Navbar isAuthenticated={isAuthenticated} />
      <Hero />
      <Features />
      <HowItWorks />
      <Pricing />
      <About />
      <Footer />
    </div>
  );
}
