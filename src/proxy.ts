import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";
import { auth } from "@/auth";

export default auth(
  (req: NextRequest & { auth: { user: any; [x: string]: any } }) => {
    const { pathname } = req.nextUrl;
    const isAuthenticated = !!req.auth?.user;

    console.log("🛡️ Middleware called for path:", pathname);
    console.log("🛡️ User authenticated:", isAuthenticated);
    console.log("🛡️ User:", req.auth?.user?.username);
    console.log("🛡️ Role:", req.auth?.user?.role);

    // Public routes that don't require authentication
    const publicRoutes = ["/login"];

    // If user is authenticated and tries to access login, redirect to dashboard
    if (isAuthenticated && pathname === "/login") {
      console.log("🔄 Redirecting authenticated user from login to dashboard");
      return NextResponse.redirect(new URL("/", req.url));
    }

    // If accessing a public route, allow access
    if (publicRoutes.includes(pathname)) {
      return NextResponse.next();
    }

    // For all other routes, require authentication
    if (!isAuthenticated) {
      console.log("🔐 Unauthorized access attempt to:", pathname);
      return NextResponse.redirect(new URL("/login", req.url));
    }

    return NextResponse.next();
  }
);

export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - api (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico, manifest.json, icon*.svg (static files)
     */
    "/((?!api|_next/static|_next/image|favicon.ico|manifest.json|icon.*\\.svg).*)",
  ],
};
