package main

import (
	"context"
	"encoding/json"
	"log"
	"math/big"

	"github.com/cozyCodr/liyali-gateway/internal/config"
	"github.com/cozyCodr/liyali-gateway/internal/db"
	"github.com/cozyCodr/liyali-gateway/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	log.Println("🌱 Starting database seeding...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Test database connection
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("✅ Connected to database")

	// Initialize sqlc queries
	queries := db.New(pool)
	ctx := context.Background()

	// Get the manager user (should exist from previous testing)
	manager, err := queries.GetUserByEmail(ctx, "manager@liyali.com")
	if err != nil {
		log.Fatalf("Manager user not found. Please ensure users are created first: %v", err)
	}
	managerID := utils.PgtypeToUUID(manager.ID)

	log.Printf("👤 Using manager user: %s (%s)", manager.Name, manager.Email)

	// Create sample workflows
	log.Println("\n📋 Creating sample workflows...")
	workflows := createSampleWorkflows(ctx, queries, managerID)
	log.Printf("✅ Created %d workflows", len(workflows))

	// Create sample documents
	log.Println("\n📄 Creating sample documents...")
	documents := createSampleDocuments(ctx, queries, managerID, workflows)
	log.Printf("✅ Created %d documents", len(documents))

	// Create sample approval tasks
	log.Println("\n✔️  Creating sample approval tasks...")
	tasks := createSampleApprovalTasks(ctx, queries, managerID, documents)
	log.Printf("✅ Created %d approval tasks", len(tasks))

	// Create sample notifications
	log.Println("\n🔔 Creating sample notifications...")
	notifications := createSampleNotifications(ctx, queries, managerID, tasks)
	log.Printf("✅ Created %d notifications", len(notifications))

	log.Println("\n🎉 Database seeding completed successfully!")
}

func createSampleWorkflows(ctx context.Context, queries *db.Queries, managerID uuid.UUID) []db.Workflow {
	workflows := []db.Workflow{}

	// Workflow 1: Requisition Approval (2 stages)
	requisitionStages := []map[string]interface{}{
		{
			"stage":      1,
			"name":       "Department Head Approval",
			"approvers":  []string{"MANAGER"},
			"required":   1,
			"sla_hours":  24,
		},
		{
			"stage":      2,
			"name":       "Finance Approval",
			"approvers":  []string{"FINANCE_MANAGER"},
			"required":   1,
			"sla_hours":  48,
		},
	}
	requisitionStagesJSON, _ := json.Marshal(requisitionStages)

	requisitionWorkflow, err := queries.CreateWorkflow(ctx, db.CreateWorkflowParams{
		Name:         "Standard Requisition Workflow",
		Description:  pgtype.Text{String: "Two-stage approval for purchase requisitions", Valid: true},
		DocumentType: "REQUISITION",
		Stages:       requisitionStagesJSON,
		IsActive:     pgtype.Bool{Bool: true, Valid: true},
		CreatedBy:    utils.UUIDToPgtype(managerID),
	})
	if err != nil {
		log.Printf("Warning: Failed to create requisition workflow: %v", err)
	} else {
		workflows = append(workflows, requisitionWorkflow)
		log.Printf("  ✓ Created: %s", requisitionWorkflow.Name)
	}

	// Workflow 2: Budget Approval (3 stages)
	budgetStages := []map[string]interface{}{
		{
			"stage":      1,
			"name":       "Department Head Review",
			"approvers":  []string{"MANAGER"},
			"required":   1,
			"sla_hours":  48,
		},
		{
			"stage":      2,
			"name":       "Finance Review",
			"approvers":  []string{"FINANCE_MANAGER"},
			"required":   1,
			"sla_hours":  72,
		},
		{
			"stage":      3,
			"name":       "Executive Approval",
			"approvers":  []string{"ADMIN"},
			"required":   1,
			"sla_hours":  96,
		},
	}
	budgetStagesJSON, _ := json.Marshal(budgetStages)

	budgetWorkflow, err := queries.CreateWorkflow(ctx, db.CreateWorkflowParams{
		Name:         "Annual Budget Approval",
		Description:  pgtype.Text{String: "Three-stage approval for annual budgets", Valid: true},
		DocumentType: "BUDGET",
		Stages:       budgetStagesJSON,
		IsActive:     pgtype.Bool{Bool: true, Valid: true},
		CreatedBy:    utils.UUIDToPgtype(managerID),
	})
	if err != nil {
		log.Printf("Warning: Failed to create budget workflow: %v", err)
	} else {
		workflows = append(workflows, budgetWorkflow)
		log.Printf("  ✓ Created: %s", budgetWorkflow.Name)
	}

	// Workflow 3: Purchase Order Approval (2 stages)
	poStages := []map[string]interface{}{
		{
			"stage":      1,
			"name":       "Procurement Review",
			"approvers":  []string{"MANAGER"},
			"required":   1,
			"sla_hours":  24,
		},
		{
			"stage":      2,
			"name":       "Finance Approval",
			"approvers":  []string{"FINANCE_MANAGER"},
			"required":   1,
			"sla_hours":  48,
		},
	}
	poStagesJSON, _ := json.Marshal(poStages)

	poWorkflow, err := queries.CreateWorkflow(ctx, db.CreateWorkflowParams{
		Name:         "Purchase Order Workflow",
		Description:  pgtype.Text{String: "Standard approval for purchase orders", Valid: true},
		DocumentType: "PURCHASE_ORDER",
		Stages:       poStagesJSON,
		IsActive:     pgtype.Bool{Bool: true, Valid: true},
		CreatedBy:    utils.UUIDToPgtype(managerID),
	})
	if err != nil {
		log.Printf("Warning: Failed to create PO workflow: %v", err)
	} else {
		workflows = append(workflows, poWorkflow)
		log.Printf("  ✓ Created: %s", poWorkflow.Name)
	}

	return workflows
}

func createSampleDocuments(ctx context.Context, queries *db.Queries, managerID uuid.UUID, workflows []db.Workflow) []db.Document {
	documents := []db.Document{}

	if len(workflows) == 0 {
		log.Println("Warning: No workflows available, skipping document creation")
		return documents
	}

	// Get workflow IDs
	var requisitionWorkflowID, budgetWorkflowID, poWorkflowID uuid.UUID
	for _, wf := range workflows {
		switch wf.DocumentType {
		case "REQUISITION":
			requisitionWorkflowID = utils.PgtypeToUUID(wf.ID)
		case "BUDGET":
			budgetWorkflowID = utils.PgtypeToUUID(wf.ID)
		case "PURCHASE_ORDER":
			poWorkflowID = utils.PgtypeToUUID(wf.ID)
		}
	}

	// Document 1: Office Supplies Requisition
	docData1 := map[string]interface{}{
		"items": []map[string]interface{}{
			{"name": "Printer Paper", "quantity": 10, "unit_price": 25.00},
			{"name": "Pens (Box)", "quantity": 5, "unit_price": 15.00},
			{"name": "Notebooks", "quantity": 20, "unit_price": 3.50},
		},
		"justification": "Monthly office supplies replenishment",
	}
	docDataJSON1, _ := json.Marshal(docData1)

	doc1, err := queries.CreateDocument(ctx, db.CreateDocumentParams{
		DocumentType:   "REQUISITION",
		DocumentNumber: "REQ-2025-001",
		Title:          "Office Supplies - January 2025",
		Description:    pgtype.Text{String: "Monthly requisition for office supplies", Valid: true},
		Amount:         pgtype.Numeric{Int: big.NewInt(39500), Exp: -2, Valid: true}, // 395.00
		Currency:       pgtype.Text{String: "USD", Valid: true},
		Status:         "DRAFT",
		CreatedBy:      utils.UUIDToPgtype(managerID),
		Department:     pgtype.Text{String: "Operations", Valid: true},
		WorkflowID:     utils.UUIDToPgtype(requisitionWorkflowID),
		Data:           docDataJSON1,
	})
	if err != nil {
		log.Printf("Warning: Failed to create document 1: %v", err)
	} else {
		documents = append(documents, doc1)
		log.Printf("  ✓ Created: %s", doc1.Title)
	}

	// Document 2: Q1 2025 Budget
	docData2 := map[string]interface{}{
		"categories": []map[string]interface{}{
			{"name": "Personnel", "amount": 150000},
			{"name": "Operations", "amount": 50000},
			{"name": "Marketing", "amount": 30000},
			{"name": "Technology", "amount": 40000},
		},
		"period": "Q1 2025",
	}
	docDataJSON2, _ := json.Marshal(docData2)

	doc2, err := queries.CreateDocument(ctx, db.CreateDocumentParams{
		DocumentType:   "BUDGET",
		DocumentNumber: "BUD-2025-Q1",
		Title:          "Q1 2025 Operating Budget",
		Description:    pgtype.Text{String: "First quarter budget proposal", Valid: true},
		Amount:         pgtype.Numeric{Int: big.NewInt(27000000), Exp: -2, Valid: true}, // 270,000.00
		Currency:       pgtype.Text{String: "USD", Valid: true},
		Status:         "DRAFT",
		CreatedBy:      utils.UUIDToPgtype(managerID),
		Department:     pgtype.Text{String: "Finance", Valid: true},
		WorkflowID:     utils.UUIDToPgtype(budgetWorkflowID),
		Data:           docDataJSON2,
	})
	if err != nil {
		log.Printf("Warning: Failed to create document 2: %v", err)
	} else {
		documents = append(documents, doc2)
		log.Printf("  ✓ Created: %s", doc2.Title)
	}

	// Document 3: IT Equipment Purchase Order
	docData3 := map[string]interface{}{
		"vendor": "TechSupply Inc.",
		"items": []map[string]interface{}{
			{"name": "Dell Laptop", "model": "XPS 15", "quantity": 5, "unit_price": 1500.00},
			{"name": "Monitor", "model": "27-inch 4K", "quantity": 5, "unit_price": 400.00},
		},
		"delivery_date": "2025-02-15",
	}
	docDataJSON3, _ := json.Marshal(docData3)

	doc3, err := queries.CreateDocument(ctx, db.CreateDocumentParams{
		DocumentType:   "PURCHASE_ORDER",
		DocumentNumber: "PO-2025-0042",
		Title:          "IT Equipment - New Employee Setup",
		Description:    pgtype.Text{String: "Laptops and monitors for new hires", Valid: true},
		Amount:         pgtype.Numeric{Int: big.NewInt(950000), Exp: -2, Valid: true}, // 9,500.00
		Currency:       pgtype.Text{String: "USD", Valid: true},
		Status:         "SUBMITTED",
		CreatedBy:      utils.UUIDToPgtype(managerID),
		Department:     pgtype.Text{String: "IT", Valid: true},
		WorkflowID:     utils.UUIDToPgtype(poWorkflowID),
		Data:           docDataJSON3,
	})
	if err != nil {
		log.Printf("Warning: Failed to create document 3: %v", err)
	} else {
		documents = append(documents, doc3)
		log.Printf("  ✓ Created: %s", doc3.Title)
	}

	return documents
}

func createSampleApprovalTasks(ctx context.Context, queries *db.Queries, managerID uuid.UUID, documents []db.Document) []db.ApprovalTask {
	tasks := []db.ApprovalTask{}

	if len(documents) == 0 {
		log.Println("Warning: No documents available, skipping approval task creation")
		return tasks
	}

	// Create approval tasks for submitted documents
	for _, doc := range documents {
		if doc.Status == "SUBMITTED" {
			task, err := queries.CreateApprovalTask(ctx, db.CreateApprovalTaskParams{
				DocumentID:   doc.ID,
				AssignedTo:   utils.UUIDToPgtype(managerID),
				AssignedBy:   utils.UUIDToPgtype(managerID),
				Status:       "PENDING",
				CurrentStage: 1,
				TotalStages:  2,
				Priority:     pgtype.Text{String: "MEDIUM", Valid: true},
				Notes:        pgtype.Text{String: "Please review and approve", Valid: true},
			})
			if err != nil {
				log.Printf("Warning: Failed to create approval task: %v", err)
			} else {
				tasks = append(tasks, task)
				log.Printf("  ✓ Created task for: %s", doc.Title)
			}
		}
	}

	return tasks
}

func createSampleNotifications(ctx context.Context, queries *db.Queries, managerID uuid.UUID, tasks []db.ApprovalTask) []db.Notification {
	notifications := []db.Notification{}

	// Create notifications for approval tasks
	for _, task := range tasks {
		notif, err := queries.CreateNotification(ctx, db.CreateNotificationParams{
			UserID:       utils.UUIDToPgtype(managerID),
			Type:         "TASK_ASSIGNED",
			Title:        "New Approval Task Assigned",
			Message:      "You have been assigned a new approval task",
			RelatedID:    task.ID,
			SentViaEmail: pgtype.Bool{Bool: false, Valid: true},
		})
		if err != nil {
			log.Printf("Warning: Failed to create notification: %v", err)
		} else {
			notifications = append(notifications, notif)
		}
	}

	log.Printf("  ✓ Created %d notifications", len(notifications))
	return notifications
}
