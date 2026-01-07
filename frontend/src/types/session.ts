/**
 * Session and Authentication Types
 * Types for session management and authentication state
 */

import { User } from './user';
import { UserRole } from './api';

// ================== SESSION TYPES ==================

export interface SessionData {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  token?: string;
  expiresAt?: Date;
  data?: any;                  // Generic data property for backward compatibility
}

export interface AuthSession {
  access_token: string;
  refresh_token?: string;
  role: UserRole;
  user_id?: string;
  change_password?: boolean;
  mfa_required?: boolean;
  organization_id?: string;
  expiresAt?: Date;
  user?: User;
}

export interface AuthUser {
  id: string;
  name: string;
  email: string;
  role: UserRole;
  organizationId?: string;
  isSuperAdmin?: boolean;
}

// Re-export UserType from core to avoid duplication
export type { UserType } from './core';

// ================== PERMISSION TYPES ==================

export interface PermissionCheck {
  resource: string;
  action: string;
  context?: Record<string, any>;
}

export interface PermissionsData {
  /** Check if user has a specific permission */
  hasPermission: (resource: string, action: string) => boolean;
  /** Get all permissions for the current user */
  permissions: string[];
  /** Check if user has any of the specified roles */
  hasRole: (roles: string | string[]) => boolean;
  /** Get current user role */
  role: string;
}

// ================== SESSION CONFIG ==================

export interface SessionConfig {
  tokenRefreshBuffer: number; // milliseconds before expiry to refresh
  maxRetries: number;
  retryDelay: number;
  autoLogoutWarning: number; // minutes before auto logout to show warning
}

// ================== LOGIN ATTEMPT TYPES ==================

export interface LoginAttempt {
  id: string;
  userId?: string;
  email: string;
  ipAddress: string;
  userAgent: string;
  success: boolean;
  failureReason?: string;
  createdAt: Date;
}

export interface AccountLockout {
  id: string;
  userId: string;
  reason: string;
  lockedAt: Date;
  expiresAt?: Date;
  unlocked: boolean;
  unlockedAt?: Date;
  unlockedBy?: string;
}