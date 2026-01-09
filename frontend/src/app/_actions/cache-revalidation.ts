"use server";

import { revalidatePath, revalidateTag } from "next/cache";

/**
 * Server action to revalidate Next.js cache when switching organizations
 * This ensures server-side data is fresh for the new organization
 */

/**
 * Organization-scoped paths that should be revalidated on org switch
 */
const ORGANIZATION_SCOPED_PATHS = [
  "/dashboard",
  "/requisitions",
  "/purchase-orders",
  "/grn",
  "/budgets",
  "/analytics",
  "/workflows",
  "/approvals",
  "/settings",
  "/reports",
  "/vendors",
  "/categories",
  "/users",
  "/members",
  "/notifications",
  "/audit-logs",
  "/documents",
  "/templates",
  "/payment-vouchers",
  "/invoices",
  "/contracts",
  "/projects",
  "/departments",
  "/cost-centers",
  "/inventory",
  "/assets",
] as const;

/**
 * Organization-scoped cache tags that should be revalidated
 */
const ORGANIZATION_SCOPED_TAGS = [
  "requisitions",
  "purchase-orders",
  "grn",
  "budgets",
  "analytics",
  "dashboard",
  "workflows",
  "approvals",
  "categories",
  "vendors",
  "users",
  "members",
  "settings",
  "reports",
  "notifications",
  "audit-logs",
  "documents",
  "templates",
  "payment-vouchers",
  "invoices",
  "contracts",
  "projects",
  "departments",
  "cost-centers",
  "inventory",
  "assets",
  "organization-data",
] as const;

/**
 * Revalidate all organization-scoped paths and tags
 */
export async function revalidateOrganizationCache(organizationId?: string) {
  try {
    console.log(
      `[Server Cache] Revalidating cache for organization: ${organizationId || "all"}`
    );

    // Revalidate all organization-scoped paths
    for (const path of ORGANIZATION_SCOPED_PATHS) {
      try {
        revalidatePath(path);
        revalidatePath(`${path}/[...slug]`); // Revalidate dynamic routes
        console.log(`[Server Cache] Revalidated path: ${path}`);
      } catch (error) {
        console.warn(
          `[Server Cache] Failed to revalidate path ${path}:`,
          error
        );
      }
    }

    // Revalidate all organization-scoped cache tags
    for (const tag of ORGANIZATION_SCOPED_TAGS) {
      try {
        // Note: revalidateTag in Next.js 16+ might have different signature
        (revalidateTag as any)(tag);
        if (organizationId) {
          (revalidateTag as any)(`${tag}-${organizationId}`); // Organization-specific tags
        }
        console.log(`[Server Cache] Revalidated tag: ${tag}`);
      } catch (error) {
        console.warn(`[Server Cache] Failed to revalidate tag ${tag}:`, error);
      }
    }

    // Revalidate the entire app layout to ensure fresh navigation data
    revalidatePath("/(private)", "layout");
    revalidatePath("/(private)/(main)", "layout");

    console.log("[Server Cache] Organization cache revalidation completed");
    return { success: true, message: "Cache revalidated successfully" };
  } catch (error) {
    console.error(
      "[Server Cache] Failed to revalidate organization cache:",
      error
    );
    return {
      success: false,
      message: "Failed to revalidate cache",
      error: error instanceof Error ? error.message : "Unknown error",
    };
  }
}

/**
 * Revalidate specific paths
 */
export async function revalidateSpecificPaths(paths: string[]) {
  try {
    console.log(
      `[Server Cache] Revalidating specific paths: ${paths.join(", ")}`
    );

    for (const path of paths) {
      try {
        revalidatePath(path);
        console.log(`[Server Cache] Revalidated path: ${path}`);
      } catch (error) {
        console.warn(
          `[Server Cache] Failed to revalidate path ${path}:`,
          error
        );
      }
    }

    return { success: true, message: "Paths revalidated successfully" };
  } catch (error) {
    console.error("[Server Cache] Failed to revalidate specific paths:", error);
    return {
      success: false,
      message: "Failed to revalidate paths",
      error: error instanceof Error ? error.message : "Unknown error",
    };
  }
}

/**
 * Revalidate specific cache tags
 */
export async function revalidateSpecificTags(tags: string[]) {
  try {
    console.log(
      `[Server Cache] Revalidating specific tags: ${tags.join(", ")}`
    );

    for (const tag of tags) {
      try {
        (revalidateTag as any)(tag);
        console.log(`[Server Cache] Revalidated tag: ${tag}`);
      } catch (error) {
        console.warn(`[Server Cache] Failed to revalidate tag ${tag}:`, error);
      }
    }

    return { success: true, message: "Tags revalidated successfully" };
  } catch (error) {
    console.error("[Server Cache] Failed to revalidate specific tags:", error);
    return {
      success: false,
      message: "Failed to revalidate tags",
      error: error instanceof Error ? error.message : "Unknown error",
    };
  }
}

/**
 * Force revalidate the current page
 */
export async function revalidateCurrentPage() {
  try {
    // This will be called from the client with the current pathname
    revalidatePath("/", "page");
    return { success: true, message: "Current page revalidated" };
  } catch (error) {
    console.error("[Server Cache] Failed to revalidate current page:", error);
    return {
      success: false,
      message: "Failed to revalidate current page",
      error: error instanceof Error ? error.message : "Unknown error",
    };
  }
}

/**
 * Get organization-scoped paths and tags for reference
 */
export async function getOrganizationScopedCacheKeys() {
  return {
    paths: ORGANIZATION_SCOPED_PATHS,
    tags: ORGANIZATION_SCOPED_TAGS,
  };
}
