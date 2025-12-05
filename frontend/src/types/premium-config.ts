export interface PremiumConfig {
  id?: string
  maxUsers?: number
  maxStorage?: number
  features?: string[]
  price?: number
  billingCycle?: 'monthly' | 'yearly'
  active?: boolean
}
