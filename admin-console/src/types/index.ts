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
  subscription_status: "trial" | "active" | "expired" | "cancelled";
  user_count: number;
  trial_status: "trial" | "subscribed" | "expired";
  trial_start_date?: string;
  trial_end_date?: string;
  days_remaining?: number;
  settings?: {
    max_users?: number;
    custom_branding?: boolean;
  };
}

// User types
export interface AdminUser {
  id: string;
  email: string;
  name: string;
  role: "super_admin";
  permissions: string[];
  created_at: string;
  last_login?: string;
}

// Dashboard types
export interface DashboardMetrics {
  total_organizations: number;
  active_organizations: number;
  trial_organizations: number;
  expiring_trials: number;
  total_users: number;
  active_users: number;
  recent_organizations: Array<{
    id: string;
    name: string;
    created_at: string;
    status: string;
  }>;
  system_health: {
    uptime: string;
    cpu_usage: number;
    memory_usage: number;
    disk_usage: number;
  };
}

// Trial management types
export interface TrialResetRequest {
  trial_days: number;
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
  displayName: string;
  description: string;
  priceMonthly: number;
  priceYearly: number;
  maxWorkspaces: number;
  maxTeamMembers: number;
  maxDocuments: number;
  maxWorkflows: number;
  maxCustomRoles: number;
  maxRequisitions: number;
  maxBudgets: number;
  maxPurchaseOrders: number;
  maxPaymentVouchers: number;
  maxGRNs: number;
  maxDepartments: number;
  maxVendors: number;
  features: string[];
  isActive: boolean;
  sortOrder: number;
  createdAt: string;
  updatedAt: string;
  featureCount?: number;
  organizationCount?: number;
}

export interface SubscriptionFeature {
  id: string;
  name: string;
  displayName: string;
  description: string;
  category: string;
  isActive: boolean;
  createdAt: string;
}

export interface CreateTierRequest {
  name: string;
  displayName: string;
  description: string;
  priceMonthly: number;
  priceYearly: number;
  maxWorkspaces: number;
  maxTeamMembers: number;
  maxDocuments: number;
  maxWorkflows: number;
  maxCustomRoles: number;
  maxRequisitions: number;
  maxBudgets: number;
  maxPurchaseOrders: number;
  maxPaymentVouchers: number;
  maxGRNs: number;
  maxDepartments: number;
  maxVendors: number;
  features: string[];
  isActive: boolean;
  sortOrder: number;
}

export interface UpdateTierRequest extends Partial<Omit<CreateTierRequest, "name">> {
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
