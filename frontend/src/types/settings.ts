/**
 * Settings and Configuration Types
 * User settings, account tiers, and application configuration
 */

export type Currency = "USD" | "ZMW";

export type AccountTier = "FREE" | "PRO" | "ENTERPRISE";

export type SignupSettings = {
  allowSignups: boolean;
  requireEmailVerification: boolean;
  defaultAccountTier: AccountTier;
  defaultCurrency: Currency;
  autoApproveUsers?: boolean;  // Auto-approve new user registrations
  defaultRole?: string;        // Default role for new users
};

export type SettingsData = {
  id: string;
  userId?: string;
  theme?: "light" | "dark" | "auto";
  language?: string;
  currency?: Currency;
  notifications?: {
    email?: boolean;
    push?: boolean;
    sms?: boolean;
  };
  privacy?: {
    profileVisible?: boolean;
    showActivity?: boolean;
  };
  createdAt?: Date;
  updatedAt?: Date;
};
