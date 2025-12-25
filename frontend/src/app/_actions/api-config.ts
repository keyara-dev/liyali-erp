import { verifySession } from "@/lib/auth";
import { AUTH_SESSION } from "@/lib/constants";
import axiosClient, { AxiosRequestConfig, AxiosRequestHeaders } from "axios";

export const axios = axiosClient.create({
  baseURL: process.env.BASE_URL || "http://localhost:8080",
});

// Reusable error handler following DRY principle
const createErrorHandler = () => async (error: Error | any) => {
  // Timeout error
  if (error.code === "ECONNABORTED" || error.code === "ETIMEDOUT") {
    throw {
      ...error,
      type: "Timeout Error",
      message: "Request timed out! Please try again",
    };
  }

  // Network error
  if (
    error.code === "ECONNREFUSED" ||
    error.code === "ECONNRESET" ||
    error.code === "ENOTFOUND"
  ) {
    return Promise.reject({
      ...error,
      type: "Network Error",
      message: "Please check your internet connection.",
    });
  }

  // No response error
  if (!error.response) {
    return Promise.reject({
      ...error,
      type: "No Response Error",
      message: "No response from server.",
    });
  }

  const { status, data } = error.response;

  // Handle token expiration (403 with "token has expired" message)
  // if (
  //   status === 403 &&
  //   (data?.error === "token has expired" || data?.message === "token has expired")
  // ) {
  //   console.log("[AUTH] Token expired - logging out user");
  //   const response = await logUserOut("expired token");

  //   if (response.success) {
  //     redirect("/auth/login");
  //   }
  // }

  // Handle specific error codes
  const errorMap: { [x: string]: string } = {
    400: "Bad request",
    403: "Forbidden",
    404: "Resource not found",
    500: "Internal server error",
    502: "Bad gateway",
    503: "Service unavailable",
  };

  return Promise.reject({
    ...error,
    // details: data?.errors || {},
    type: "API",
    status,
    message:
      data?.message || data?.error || errorMap[status] || "Request failed",
  });
};

// Shared response and error handlers
const responseHandler = (response: any) => response;
const errorHandler = createErrorHandler();

// Apply the same interceptors to both API clients
axios.interceptors.response.use(responseHandler, errorHandler);
// externalApiClient.interceptors.response.use(responseHandler, errorHandler);

export type RequestType = AxiosRequestConfig & {
  contentType?: AxiosRequestHeaders["Content-Type"];
};

const authenticatedApiClient = async (request: RequestType) => {
  const { session } = await verifySession();

  if (!session?.access_token) {
    throw new Error("No valid session found");
  }

  const headers: any = {
    "Content-type": request.contentType
      ? request.contentType
      : "application/json",
    Authorization: `Bearer ${session?.access_token}`,
    Cookie: `${AUTH_SESSION}=${session.access_token}`, // Forward the session cookie to API
  };

  // Add organization context if available
  if (session.organization_id) {
    headers["X-Organization-ID"] = session.organization_id;
  }

  const config = {
    method: "GET",
    headers,
    withCredentials: true,
    ...request,
  };

  return await axios(config);
};

export default authenticatedApiClient;

// Re-export response helpers from the library
// This file maintains backward compatibility for imports
// Note: This file does NOT have 'use server' because it only exports utility functions
export {
  successResponse,
  unauthorizedResponse,
  notFoundResponse,
  methodNotAllowedResponse,
  handleError,
  badRequestResponse,
} from "@/lib/response-helpers";
