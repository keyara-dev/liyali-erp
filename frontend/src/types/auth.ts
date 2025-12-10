/**
 * Authentication and Account Types
 * Consolidated from src/lib/types/index.ts
 */

// User Types
export type UserType =
  | "REQUESTER"
  | "DEPARTMENT_MANAGER"
  | "FINANCE_OFFICER"
  | "DIRECTOR"
  | "CFO"
  | "COMPLIANCE_OFFICER"
  | "ADMIN";

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
  user_type?: UserType;
  mfa_enabled?: boolean;
  is_ldap_user?: boolean;
  created_at?: Date | string;
  updated_at?: Date | string;
  last_login?: Date | string;
  expiresAt?: Date | string;
};

export interface AuthSession {
  accessToken: string;
  user: User;
  user_type?: UserType;
  user_id?: string;
  change_password?: boolean;
  mfa_required?: boolean;
  institution_id?: string;
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
