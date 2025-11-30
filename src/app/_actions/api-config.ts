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
