// API Response type
export interface APIResponse<T = any> {
  success: boolean;
  data?: T;
  message?: string;
  meta?: {
    total?: number;
    page?: number;
    limit?: number;
    totalPages?: number;
  };
  status?: number;
}

// Organization types
export interface Organization {
  id: string;
  name: string;
  domain: string;
  created_at: string;
  status: "active" | "suspended" | "pending";
  subscription_tier: "basic" | "professional" | "enterprise";
  user_count: number;
  trial_status: "trial" | "subscribed" | "expired";
  trial_start_date?: string;
  trial_end_date?: string;
  days_remaining?: number;
}

// User types
export interface AdminUser {
  id: string;
  email: string;
  name: string;
  role: "super_admin" | "admin" | "compliance_officer";
  permissions: string[];
  created_at: string;
  last_login?: string;
}

// Dashboard types
export interface DashboardMetrics {
  total_organizations: number;
  active_organizations: number;
  trial_organizations: number;
  expired_trials: number;
  total_users: number;
  active_users: number;
  system_health: {
    uptime: string;
    cpu_usage: number;
    memory_usage: number;
    disk_usage: number;
  };
}

// Trial management types
export interface TrialResetRequest {
  trialDays: number;
  reason: string;
}

export interface TrialStatus {
  organization_id: string;
  trial_start_date: string;
  trial_end_date: string;
  days_remaining: number;
  status: "active" | "expired" | "extended";
}

// Subscription types
export interface SubscriptionPlan {
  id: string;
  name: string;
  price: number;
  features: string[];
  max_users: number;
  storage_limit: string;
}

export interface SubscriptionTier {
  id: string;
  name: string;
  display_name: string;
  description: string;
  price_monthly: number;
  price_yearly: number;
  max_users: number;
  max_organizations?: number;
  storage_limit_gb: number;
  features: string[];
  is_active: boolean;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

export interface SubscriptionFeature {
  id: string;
  name: string;
  display_name: string;
  description: string;
  category: string;
  is_active: boolean;
  created_at: string;
}

export interface CreateTierRequest {
  name: string;
  display_name: string;
  description: string;
  price_monthly: number;
  price_yearly: number;
  max_users: number;
  max_organizations?: number;
  storage_limit_gb: number;
  features: string[];
  is_active: boolean;
  sort_order: number;
}

export interface UpdateTierRequest extends Partial<CreateTierRequest> {
  id: string;
}

// Audit log types
export interface AuditLog {
  id: string;
  action: string;
  user_id: string;
  user_name: string;
  organization_id?: string;
  organization_name?: string;
  details: Record<string, any>;
  timestamp: string;
  ip_address?: string;
  user_agent?: string;
}
