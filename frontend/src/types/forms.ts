/**
 * Form-related Types
 * Types for form data and validation
 */

export interface ChangePassword {
  oldPassword: string
  newPassword: string
  confirmPassword: string
}

export interface ChangePasswordRequest extends ChangePassword {
  userId?: string
}

export interface ChangePasswordResponse {
  success: boolean
  message: string
  error?: string
}
