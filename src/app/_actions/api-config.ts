import { auth } from "@/auth";
import { APIResponse } from "@/types";
import axiosClient, { AxiosRequestConfig, AxiosRequestHeaders } from "axios";
import { cookies } from "next/headers";

export type RequestType = AxiosRequestConfig & {
  contentType?: AxiosRequestHeaders["Content-Type"];
};

const axios = axiosClient.create({
  baseURL: process.env.NEXT_PUBLIC_SERVER_URL || process.env.SERVER_URL,
});

// Add request interceptor to include authentication headers
axios.interceptors.request.use(
  async (config) => {
    try {
      // Get NextAuth session
      const session = await auth();

      // Get admin token from cookies
      const cookieStore = await cookies();
      const adminToken = cookieStore.get("admin-token")?.value;

      // For server-to-server calls within the same app, we need to ensure the session is properly forwarded
      // The issue is that server actions run in a different context and may not have access to the original request cookies

      // Get all cookies from the cookie store and forward them
      const allCookies = cookieStore.getAll();
      let cookiesToSend = "";

      // Forward ALL cookies to ensure NextAuth session is properly passed
      allCookies.forEach((cookie, index) => {
        if (index === 0) {
          cookiesToSend = `${cookie.name}=${cookie.value}`;
        } else {
          cookiesToSend += `; ${cookie.name}=${cookie.value}`;
        }
      });

      if (cookiesToSend) {
        config.headers.Cookie = cookiesToSend;
      }

      // Add session verification header if we have a session
      if (session?.user) {
        config.headers["X-Auth-User-ID"] = session.user.id;
        config.headers["X-Auth-User-Role"] = (session.user as any).role || "";
        config.headers["X-Debug-User"] = JSON.stringify({
          id: session.user.id,
          username: (session.user as any).username,
          role: (session.user as any).role,
        });
      }

      console.log("🔐 Request headers set:", {
        hasCookies: !!cookiesToSend,
        cookieCount: allCookies.length,
        hasAdminToken: !!adminToken,
        hasSession: !!session?.user,
        sessionUser: session?.user
          ? {
              id: session.user.id,
              role: (session.user as any).role,
              username: (session.user as any).username,
            }
          : null,
        cookies: cookiesToSend
          ? cookiesToSend.substring(0, 300) + "..."
          : "none",
        endpoint: `${config.method?.toUpperCase()} ${config.url}`,
      });
    } catch (error) {
      console.warn("⚠️ Failed to set auth headers:", error);
      // Even if session retrieval fails, still try to forward cookies
      try {
        const cookieStore = await cookies();
        const allCookies = cookieStore.getAll();
        let cookiesToSend = "";
        allCookies.forEach((cookie, index) => {
          if (index === 0) {
            cookiesToSend = `${cookie.name}=${cookie.value}`;
          } else {
            cookiesToSend += `; ${cookie.name}=${cookie.value}`;
          }
        });
        if (cookiesToSend) {
          config.headers.Cookie = cookiesToSend;
        }
      } catch (fallbackError) {
        console.error("⚠️ Could not even forward cookies:", fallbackError);
      }
    }

    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Note: authenticatedService function removed as it was not being used
// and was causing build issues with missing verifySession function

// Add response interceptor for error handling
axios.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    // Network error (no connection)
    if (!error.response) {
      return Promise.reject({
        ...error,
        type: "Network Error",
        message: "Please check your internet connection.",
      });
    }

    const { status, data } = error.response;

    // Handle specific error codes
    const errorMap: { [x: string]: string } = {
      400: "Bad request",
      401: "Unauthorized - Please check your authentication credentials",
      403: "Forbidden - Insufficient permissions",
      404: "Resource not found",
      500: "Internal server error",
      502: "Bad gateway",
      503: "Service unavailable",
    };

    // Log authentication errors for debugging
    if (status === 401 || status === 403) {
      console.error("🔐 Authentication/Authorization error:", {
        status,
        message: data?.message,
        endpoint: `${originalRequest.method?.toUpperCase()} ${originalRequest.url}`,
        requestHeaders: {
          Cookie: originalRequest.headers.Cookie
            ? originalRequest.headers.Cookie.substring(0, 200) + "..."
            : "none",
          "X-Debug-User": originalRequest.headers["X-Debug-User"] || "none",
        },
        responseData: data,
      });
    }

    return Promise.reject({
      ...error,
      type: "api",
      status,
      message: data?.message || errorMap[status] || "Request failed",
      details: data?.errors || {},
    });
  }
);

// Response helpers
export function successResponse(
  data: any | null,
  message: string = "Action completed successfully"
): APIResponse {
  return {
    success: true,
    message,
    data,
    status: 200,
    statusText: "OK",
  };
}

export function unauthorizedResponse(
  message: string = "Unauthorized"
): APIResponse {
  return {
    success: false,
    message,
    data: null,
    status: 401,
    statusText: "UNAUTHORIZED",
  };
}

export function notFoundResponse(message: string): APIResponse {
  return {
    success: false,
    message,
    data: null,
    status: 404,
    statusText: "NOT FOUND",
  };
}

export function methodNotAllowedResponse(): APIResponse {
  return {
    success: false,
    message: "Method not allowed",
    data: null,
    status: 405,
    statusText: "METHOD NOT ALLOWED",
  };
}

export function handleError(
  error: any,
  method = "GET",
  url: string
): APIResponse {
  console.error({
    endpoint: `${method} |  ~ ${url}`,
    status: error?.response?.status,
    statusText: error?.response?.statusText,
    headers: error?.response?.headers,
    config: error?.response?.config,
    data: error?.response?.data || error,
  });

  // Handle authentication errors specifically
  const status = error?.response?.status || 500;
  if (status === 401) {
    return unauthorizedResponse(
      error?.response?.data?.message ||
        "Authentication required. Please log in again."
    );
  }

  if (status === 403) {
    return {
      success: false,
      message:
        error?.response?.data?.message ||
        "You don't have permission to perform this action.",
      data: null,
      status: 403,
      statusText: "FORBIDDEN",
    };
  }

  return {
    success: false,
    message:
      error?.response?.data?.message ||
      error?.response?.data?.error ||
      error?.response?.message ||
      error?.message ||
      "Oops! Something went wrong. Please try again.",
    data: null,
    status: status,
    statusText: error?.response?.statusText || "INTERNAL SERVER ERROR",
  };
}

export function badRequestResponse(message: string): APIResponse {
  return {
    success: false,
    message,
    data: null,
    status: 400,
    statusText: "BAD REQUEST",
  };
}

export default axios;
