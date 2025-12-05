export interface PremiumPlan {
  id?: string
  name: string
  description?: string
  price: number
  billingCycle: 'monthly' | 'yearly'
  features: string[]
  active?: boolean
}
