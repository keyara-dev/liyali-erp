package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/liyali/liyali-gateway/config"
)

// SeedSubscriptionData seeds the database with subscription plans and feature flags
func main() {
	// Initialize database connection
	config.InitDatabase()

	// Seed subscription plans
	if err := seedSubscriptionPlans(); err != nil {
		log.Fatalf("Failed to seed subscription plans: %v", err)
	}

	// Seed feature flags
	if err := seedFeatureFlags(); err != nil {
		log.Fatalf("Failed to seed feature flags: %v", err)
	}

	log.Println("✅ Subscription data seeded successfully!")
}

func seedSubscriptionPlans() error {
	log.Println("🌱 Seeding subscription plans...")

	plans := []struct {
		Name         string                 `json:"name"`
		Slug         string                 `json:"slug"`
		Description  string                 `json:"description"`
		PriceMonthly float64                `json:"price_monthly"`
		PriceYearly  float64                `json:"price_yearly"`
		Features     []string               `json:"features"`
		MaxUsers     int32                  `json:"max_users"`
		SortOrder    int32                  `json:"sort_order"`
		Metadata     map[string]interface{} `json:"metadata"`
	}{
		{
			Name:         "Starter Plan",
			Slug:         "STARTER_PLAN",
			Description:  "Perfect for small teams getting started with procurement workflows",
			PriceMonthly: 0.00,
			PriceYearly:  0.00,
			Features: []string{
				"Core procurement workflows",
				"Up to 50 users",
				"Single Workspace",
				"Document Verification (QR Codes and Doc Numbers)",
				"Standard analytics",
				"Notifications (Email & In-App)",
			},
			MaxUsers:  50,
			SortOrder: 1,
			Metadata: map[string]interface{}{
				"offline_capabilities": false,
				"api_access":          false,
				"custom_roles":        false,
				"priority_support":    false,
				"dedicated_instance":  false,
				"sla_guarantees":      false,
			},
		},
		{
			Name:         "Pro Plan",
			Slug:         "PRO_PLAN",
			Description:  "Advanced features for growing organizations",
			PriceMonthly: 99.00,
			PriceYearly:  990.00,
			Features: []string{
				"Everything in Starter Plan",
				"Up to 200 users",
				"Custom Role management",
				"Offline capabilities",
				"Priority support",
				"Advanced analytics",
				"API Access",
			},
			MaxUsers:  200,
			SortOrder: 2,
			Metadata: map[string]interface{}{
				"offline_capabilities": true,
				"api_access":          true,
				"custom_roles":        true,
				"priority_support":    true,
				"dedicated_instance":  false,
				"sla_guarantees":      false,
			},
		},
		{
			Name:         "Enterprise",
			Slug:         "ENTERPRISE",
			Description:  "Complete solution for large organizations",
			PriceMonthly: 0.00,
			PriceYearly:  0.00,
			Features: []string{
				"Everything in Pro Plan",
				"Unlimited users",
				"Dedicated instance",
				"Custom integrations",
				"SLA guarantees",
				"Dedicated success manager",
				"Models Creation/Modifications",
			},
			MaxUsers:  -1,
			SortOrder: 3,
			Metadata: map[string]interface{}{
				"offline_capabilities": true,
				"api_access":          true,
				"custom_roles":        true,
				"priority_support":    true,
				"dedicated_instance":  true,
				"sla_guarantees":      true,
				"custom_pricing":      true,
			},
		},
	}

	for _, plan := range plans {
		featuresJSON, _ := json.Marshal(plan.Features)
		metadataJSON, _ := json.Marshal(plan.Metadata)

		query := `
			INSERT INTO subscription_plans (name, slug, description, price_monthly, price_yearly, features, max_users, sort_order, metadata)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (slug) DO UPDATE SET
				name = EXCLUDED.name,
				description = EXCLUDED.description,
				price_monthly = EXCLUDED.price_monthly,
				price_yearly = EXCLUDED.price_yearly,
				features = EXCLUDED.features,
				max_users = EXCLUDED.max_users,
				sort_order = EXCLUDED.sort_order,
				metadata = EXCLUDED.metadata,
				updated_at = CURRENT_TIMESTAMP
		`

		result := config.DB.Exec(query,
			plan.Name,
			plan.Slug,
			plan.Description,
			plan.PriceMonthly,
			plan.PriceYearly,
			featuresJSON,
			plan.MaxUsers,
			plan.SortOrder,
			metadataJSON,
		)
		if result.Error != nil {
			return fmt.Errorf("failed to insert plan %s: %w", plan.Slug, result.Error)
		}

		log.Printf("  ✓ Seeded plan: %s", plan.Name)
	}

	return nil
}

func seedFeatureFlags() error {
	log.Println("🌱 Seeding feature flags...")

	features := []struct {
		Name              string   `json:"name"`
		Description       string   `json:"description"`
		PlanRequirements  []string `json:"plan_requirements"`
		IsTrialAllowed    bool     `json:"is_trial_allowed"`
		IsEnterpriseOnly  bool     `json:"is_enterprise_only"`
	}{
		{
			Name:              "custom_roles",
			Description:       "Create and manage custom user roles",
			PlanRequirements:  []string{"PRO_PLAN", "ENTERPRISE"},
			IsTrialAllowed:    false,
			IsEnterpriseOnly:  false,
		},
		{
			Name:              "offline_capabilities",
			Description:       "Work offline and sync when connected",
			PlanRequirements:  []string{"PRO_PLAN", "ENTERPRISE"},
			IsTrialAllowed:    false,
			IsEnterpriseOnly:  false,
		},
		{
			Name:              "api_access",
			Description:       "Access to REST API endpoints",
			PlanRequirements:  []string{"PRO_PLAN", "ENTERPRISE"},
			IsTrialAllowed:    false,
			IsEnterpriseOnly:  false,
		},
		{
			Name:              "priority_support",
			Description:       "Priority customer support",
			PlanRequirements:  []string{"PRO_PLAN", "ENTERPRISE"},
			IsTrialAllowed:    false,
			IsEnterpriseOnly:  false,
		},
		{
			Name:              "dedicated_instance",
			Description:       "Dedicated server instance",
			PlanRequirements:  []string{"ENTERPRISE"},
			IsTrialAllowed:    false,
			IsEnterpriseOnly:  true,
		},
		{
			Name:              "sla_guarantees",
			Description:       "Service Level Agreement guarantees",
			PlanRequirements:  []string{"ENTERPRISE"},
			IsTrialAllowed:    false,
			IsEnterpriseOnly:  true,
		},
		{
			Name:              "custom_integrations",
			Description:       "Custom third-party integrations",
			PlanRequirements:  []string{"ENTERPRISE"},
			IsTrialAllowed:    false,
			IsEnterpriseOnly:  true,
		},
		{
			Name:              "models_modification",
			Description:       "Create and modify data models",
			PlanRequirements:  []string{"ENTERPRISE"},
			IsTrialAllowed:    false,
			IsEnterpriseOnly:  true,
		},
		{
			Name:              "advanced_analytics",
			Description:       "Advanced reporting and analytics",
			PlanRequirements:  []string{"PRO_PLAN", "ENTERPRISE"},
			IsTrialAllowed:    false,
			IsEnterpriseOnly:  false,
		},
		{
			Name:              "unlimited_users",
			Description:       "No user limit restrictions",
			PlanRequirements:  []string{"ENTERPRISE"},
			IsTrialAllowed:    false,
			IsEnterpriseOnly:  true,
		},
		{
			Name:              "core_workflows",
			Description:       "Basic procurement workflows",
			PlanRequirements:  []string{"STARTER_PLAN", "PRO_PLAN", "ENTERPRISE"},
			IsTrialAllowed:    true,
			IsEnterpriseOnly:  false,
		},
		{
			Name:              "document_verification",
			Description:       "QR codes and document number verification",
			PlanRequirements:  []string{"STARTER_PLAN", "PRO_PLAN", "ENTERPRISE"},
			IsTrialAllowed:    true,
			IsEnterpriseOnly:  false,
		},
		{
			Name:              "standard_analytics",
			Description:       "Basic reporting and analytics",
			PlanRequirements:  []string{"STARTER_PLAN", "PRO_PLAN", "ENTERPRISE"},
			IsTrialAllowed:    true,
			IsEnterpriseOnly:  false,
		},
	}

	for _, feature := range features {
		planRequirementsJSON, _ := json.Marshal(feature.PlanRequirements)

		query := `
			INSERT INTO feature_flags (name, description, plan_requirements, is_trial_allowed, is_enterprise_only)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (name) DO UPDATE SET
				description = EXCLUDED.description,
				plan_requirements = EXCLUDED.plan_requirements,
				is_trial_allowed = EXCLUDED.is_trial_allowed,
				is_enterprise_only = EXCLUDED.is_enterprise_only,
				updated_at = CURRENT_TIMESTAMP
		`

		result := config.DB.Exec(query,
			feature.Name,
			feature.Description,
			planRequirementsJSON,
			feature.IsTrialAllowed,
			feature.IsEnterpriseOnly,
		)
		if result.Error != nil {
			return fmt.Errorf("failed to insert feature %s: %w", feature.Name, result.Error)
		}

		log.Printf("  ✓ Seeded feature: %s", feature.Name)
	}

	return nil
}