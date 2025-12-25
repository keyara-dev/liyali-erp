import { NextResponse, type NextRequest } from "next/server";
import { verifySession } from "./lib/auth";
import { AUTH_SESSION } from "./lib/constants";

export default async function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const url = request.nextUrl.clone();
  const response = NextResponse.next();

  // Add security headers to all responses
  response.headers.set("X-Frame-Options", "DENY");
  response.headers.set("X-Content-Type-Options", "nosniff");
  response.headers.set("Referrer-Policy", "strict-origin-when-cross-origin");
  response.headers.set(
    "Permissions-Policy",
    "camera=(), microphone=(), geolocation=()"
  );

  if (process.env.NODE_ENV === "production") {
    response.headers.set(
      "Strict-Transport-Security",
      "max-age=31536000; includeSubDomains; preload"
    );
  }

  // Exclude public assets like icons, manifest, and images
  if (
    pathname.startsWith("/web-app-manifest") ||
    pathname.startsWith("/favicon") ||
    pathname.startsWith("/_next") ||
    pathname.startsWith("/static") ||
    pathname.startsWith("/public") ||
    pathname.startsWith("/manifest.json")
  ) {
    return response;
  }

  // ✅ FAST: Check cookie existence only (no JWT decryption for most routes)
  const hasAuthCookie = request.cookies.has(AUTH_SESSION);

  // Define authentication pages (login, register, OTP)
  const isAuthPage =
    pathname.startsWith("/login") ||
    pathname.startsWith("/register") ||
    pathname.startsWith("/otp");

  // Check if accessing admin routes
  const isAdminRoute = pathname.startsWith("/admin");

  // If no auth cookie and not on auth page, redirect to login
  if (!hasAuthCookie && !isAuthPage) {
    url.pathname = "/login";
    return NextResponse.redirect(url);
  }

  // If has auth cookie and on auth page, let the (auth)/layout.tsx handle routing
  // The layout will check role and redirect appropriately
  // We don't redirect here to avoid conflicts

  // ✅ NEW: Admin route protection
  // Verify role for admin routes (requires JWT decode - acceptable for security)
  if (isAdminRoute && hasAuthCookie) {
    try {
      const { session, isAuthenticated, role } = await verifySession();

      console.log("[Proxy] Admin route check:", "session");

      // If not authenticated or not an ADMIN, redirect to access denied page
      if (!isAuthenticated || role !== "ADMIN") {
        console.log(
          "[Proxy] Non-admin user attempting to access admin route, redirecting to /access-denied"
        );
        url.pathname = "/access-denied";
        return NextResponse.redirect(url);
      }
    } catch (error) {
      // If decryption fails, let it through (layout will handle)
      console.error("[Proxy] Admin route check failed:", error);
    }
  }

  return response;
}

export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - api (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - public (static assets)
     * - favicon.ico, manifest.json (static files)
     * - icon*.svg, *.png, *.jpg, *.gif, *.webp, etc. (image files)
     */
    "/((?!api|_next/static|_next/image|public|favicon\\.ico|manifest\\.json|icon.*\\.svg|.*\\.png|.*\\.jpg|.*\\.jpeg|.*\\.gif|.*\\.webp|.*\\.svg|.*\\.ico|.*\\.woff|.*\\.woff2|.*\\.ttf|.*\\.eot).*)",
  ],
};
