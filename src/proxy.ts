import { NextResponse } from "next/server";

export default function middleware() {
  // This is a simple pass-through middleware
  // Authentication is handled server-side via getSession() in auth.ts
  return NextResponse.next();
}

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
