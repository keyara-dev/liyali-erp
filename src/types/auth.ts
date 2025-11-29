/**
 * Authentication and Account Types
 * Consolidated from src/lib/types/index.ts
 */

// User Types
export type UserType =
  | 'REQUESTER'
  | 'DEPARTMENT_MANAGER'
  | 'FINANCE_OFFICER'
  | 'DIRECTOR'
  | 'CFO'
  | 'COMPLIANCE_OFFICER'
  | 'ADMIN'

export interface User {
  id: string
  name: string
  email: string
  role: UserType
  department?: string
  avatar?: string
  user_type?: UserType
  expiresAt?: Date | string
}

export interface AuthSession {
  accessToken: string
  user_type?: UserType
  user_id?: string
  user?: Partial<User>
  change_password?: boolean
  mfa_required?: boolean
  organization_id?: string
  expiresAt?: Date | string
  permissions?: Permission[]
}

export interface Permission {
  id: string
  name: string
  description?: string
  resource?: string
  action?: string
}

export interface SessionResponse {
  success: boolean
  message: string
  data?: any
  status?: number
  statusText?: string
}
