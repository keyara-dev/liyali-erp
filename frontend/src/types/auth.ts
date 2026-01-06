/**
 * Authentication and Account Types
 * Consolidated from src/lib/types/index.ts
 */

// User Types
export type UserType =
  | "requester"
  | "department_manager"
  | "finance_officer"
  | "director"
  | "cfo"
  | "compliance_officer"
  | "admin";

export type User = {
  id: string;
  first_name?: string;
  last_name?: string;
  name: string;
  email: string;
  username?: string;
  role: UserType;
  department_id?: string;
  department?: string;
  avatar?: string;
  is_active?: boolean;
  role?: UserType;
  mfa_enabled?: boolean;
  is_ldap_user?: boolean;
  created_at?: Date | string;
  updated_at?: Date | string;
  last_login?: Date | string;
  expiresAt?: Date | string;
};

export interface AuthSession {
  access_token: string;
  refresh_token?: string; // Add refresh token support
  user: User;
  role?: UserType;
  user_id?: string;
  change_password?: boolean;
  mfa_required?: boolean;
  institution_id?: string;
  organization_id?: string;
  expiresAt?: Date | string;
  permissions?: Permission[];
}

export interface Permission {
  id: string;
  name: string;
  description?: string;
  resource?: string;
  action?: string;
}

export interface SessionResponse {
  success: boolean;
  message: string;
  data?: any;
  status?: number;
  statusText?: string;
}

export interface Organization {
  id: string;
  name: string;
  slug: string;
  description?: string;
  active: boolean;
  tier: string;
  createdAt: string;
  updatedAt?: string;
}

export interface RegistrationResponse {
  user: User;
  organization: Organization;
}
