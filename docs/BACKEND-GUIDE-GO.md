# Backend Implementation Guide - Go Fiber

**Status**: Implementation Guide for Phase 12
**Database**: PostgreSQL
**Framework**: Go Fiber
**ORM**: GORM

## Table of Contents

1. [Data Models](#data-models)
2. [Database Setup](#database-setup)
3. [Fiber Application Setup](#fiber-application-setup)
4. [API Routes](#api-routes)
5. [Handler Implementation](#handler-implementation)
6. [Authentication & Middleware](#authentication--middleware)
7. [Error Handling](#error-handling)
8. [Database Optimization](#database-optimization)
9. [Performance Tips](#performance-tips)
10. [NoSQL Considerations](#nosql-considerations)

---

## Data Models

### 1. User Model

```go
package models

import (
	"database/sql/driver"
	"encoding/json"
	"gorm.io/datatypes"
	"time"
)

// UserRole represents user role in the system
type UserRole string

const (
	RoleDepartmentManager UserRole = "DEPARTMENT_MANAGER"
	RoleFinanceOfficer    UserRole = "FINANCE_OFFICER"
	RoleDirector          UserRole = "DIRECTOR"
	RoleCFO               UserRole = "CFO"
	RoleComplianceOfficer UserRole = "COMPLIANCE_OFFICER"
	RoleAdmin             UserRole = "ADMIN"
	RoleUser              UserRole = "USER"
)

type User struct {
	ID           string       `gorm:"primaryKey" json:"id"`
	Email        string       `gorm:"uniqueIndex" json:"email"`
	Name         string       `json:"name"`
	Role         UserRole     `gorm:"index" json:"role"`
	Department   string       `json:"department"`
	IsActive     bool         `json:"is_active"`
	LastLogin    *time.Time   `json:"last_login"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`

	// Relations
	ApprovalTasks    []ApprovalTask    `gorm:"foreignKey:ApproverUserID"`
	ApprovalHistory  []ApprovalHistory `gorm:"foreignKey:ApproverUserID"`
	AuditLogs        []AuditLog        `gorm:"foreignKey:UserID"`
	Sessions         []Session         `gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}
```

### 2. Session Model

```go
type Session struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"index" json:"user_id"`
	Token     string    `gorm:"uniqueIndex" json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`

	// Relations
	User User `gorm:"foreignKey:UserID"`
}

func (Session) TableName() string {
	return "sessions"
}
```

### 3. ApprovalTask Model

```go
// EntityType represents the document type
type EntityType string

const (
	EntityRequisition   EntityType = "REQUISITION"
	EntityBudget        EntityType = "BUDGET"
	EntityPO            EntityType = "PO"
	EntityPV            EntityType = "PV"
	EntityGRN           EntityType = "GRN"
)

// TaskStatus represents approval task status
type TaskStatus string

const (
	StatusPending  TaskStatus = "pending"
	StatusApproved TaskStatus = "approved"
	StatusRejected TaskStatus = "rejected"
)

type ApprovalTask struct {
	ID              string          `gorm:"primaryKey" json:"id"`
	EntityID        string          `gorm:"index" json:"entity_id"`
	EntityType      EntityType      `gorm:"index" json:"entity_type"`
	EntityNumber    string          `json:"entity_number"`
	Status          TaskStatus      `gorm:"index" json:"status"`
	StageName       string          `json:"stage_name"`
	StageIndex      int             `json:"stage_index"`
	Importance      string          `json:"importance"` // LOW, MEDIUM, HIGH
	ApproverUserID  string          `gorm:"index" json:"approver_user_id"`
	CreatedAt       time.Time       `json:"created_at"`
	DueDate         time.Time       `json:"due_date"`
	WorkflowID      string          `gorm:"index" json:"workflow_id"`
	WorkflowName    string          `json:"workflow_name"`

	// Relations
	ApproverUser    User                `gorm:"foreignKey:ApproverUserID"`
	History         []ApprovalHistory   `gorm:"foreignKey:TaskID"`
	Document        datatypes.JSONType  `gorm:"type:jsonb" json:"document"`
}

func (ApprovalTask) TableName() string {
	return "approval_tasks"
}

// GetAllTasks retrieves all approval tasks with optional filtering
func (db *gorm.DB) GetAllTasks(filters map[string]interface{}) ([]ApprovalTask, error) {
	var tasks []ApprovalTask
	query := db.Preload("ApproverUser")

	if status, ok := filters["status"]; ok {
		query = query.Where("status = ?", status)
	}
	if approverID, ok := filters["approver_id"]; ok {
		query = query.Where("approver_user_id = ?", approverID)
	}

	return tasks, query.Find(&tasks).Error
}
```

### 4. ApprovalHistory Model

```go
// ActionType represents the action taken
type ActionType string

const (
	ActionApproved   ActionType = "approved"
	ActionRejected   ActionType = "rejected"
	ActionReassigned ActionType = "reassigned"
	ActionSubmitted  ActionType = "submitted"
)

type ApprovalHistory struct {
	ID              string          `gorm:"primaryKey" json:"id"`
	TaskID          string          `gorm:"index" json:"task_id"`
	Action          ActionType      `gorm:"index" json:"action"`
	ApproverUserID  string          `gorm:"index" json:"approver_user_id"`
	Timestamp       time.Time       `json:"timestamp"`
	Signature       string          `gorm:"type:text" json:"signature"`       // Base64 encoded
	Remarks         string          `gorm:"type:text" json:"remarks"`
	PreviousApprover *string        `json:"previous_approver"`

	// Relations
	Task            ApprovalTask    `gorm:"foreignKey:TaskID"`
	ApproverUser    User            `gorm:"foreignKey:ApproverUserID"`
}

func (ApprovalHistory) TableName() string {
	return "approval_history"
}
```

### 5. Document Models (Base)

```go
type Document struct {
	ID          string          `gorm:"primaryKey" json:"id"`
	Type        EntityType      `gorm:"index" json:"type"`
	Number      string          `gorm:"uniqueIndex" json:"number"`
	Status      string          `json:"status"`
	CreatorID   string          `json:"creator_id"`
	Data        datatypes.JSONType `gorm:"type:jsonb" json:"data"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

func (Document) TableName() string {
	return "documents"
}

// Requisition specific fields
type Requisition struct {
	Document
	Description   string    `json:"description"`
	Amount        float64   `json:"amount"`
	DepartmentID  string    `json:"department_id"`
	RequesterName string    `json:"requester_name"`
}

// PurchaseOrder specific fields
type PurchaseOrder struct {
	Document
	VendorID      string    `json:"vendor_id"`
	VendorName    string    `json:"vendor_name"`
	Amount        float64   `json:"amount"`
	Items         datatypes.JSONType `gorm:"type:jsonb" json:"items"`
}

// PaymentVoucher specific fields
type PaymentVoucher struct {
	Document
	InvoiceID     string    `json:"invoice_id"`
	VendorID      string    `json:"vendor_id"`
	Amount        float64   `json:"amount"`
	GLCodes       datatypes.JSONType `gorm:"type:jsonb" json:"gl_codes"`
	PaymentMethod string    `json:"payment_method"`
}

// GRN (Goods Received Note) specific fields
type GRN struct {
	Document
	POID          string    `json:"po_id"`
	WarehouseID   string    `json:"warehouse_id"`
	Items         datatypes.JSONType `gorm:"type:jsonb" json:"items"`
	DamageNotes   string    `gorm:"type:text" json:"damage_notes"`
	Variances     datatypes.JSONType `gorm:"type:jsonb" json:"variances"`
}
```

### 6. Workflow Model

```go
type WorkflowStage struct {
	Name            string   `json:"name"`
	Order           int      `json:"order"`
	ApproverRoles   []string `json:"approver_roles"` // Stored as JSON array
	AllowReassign   bool     `json:"allow_reassign"`
}

type Workflow struct {
	ID          string              `gorm:"primaryKey" json:"id"`
	Name        string              `json:"name"`
	Description string              `gorm:"type:text" json:"description"`
	EntityType  EntityType          `gorm:"index" json:"entity_type"`
	Status      string              `json:"status"` // published, draft
	Stages      datatypes.JSONType  `gorm:"type:jsonb" json:"stages"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	CreatedBy   string              `json:"created_by"`
}

func (Workflow) TableName() string {
	return "workflows"
}
```

### 7. AuditLog Model

```go
type AuditLog struct {
	ID          string          `gorm:"primaryKey" json:"id"`
	UserID      string          `gorm:"index" json:"user_id"`
	Action      string          `gorm:"index" json:"action"`
	EntityID    string          `gorm:"index" json:"entity_id"`
	EntityType  string          `json:"entity_type"`
	OldValue    datatypes.JSONType `gorm:"type:jsonb" json:"old_value"`
	NewValue    datatypes.JSONType `gorm:"type:jsonb" json:"new_value"`
	Timestamp   time.Time       `gorm:"index" json:"timestamp"`
	IPAddress   string          `json:"ip_address"`
	UserAgent   string          `json:"user_agent"`

	// Relations
	User        User            `gorm:"foreignKey:UserID"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
```

### 8. Notification Model

```go
type NotificationType string

const (
	NotificationTaskAssigned NotificationType = "task_assigned"
	NotificationApproved     NotificationType = "approved"
	NotificationRejected     NotificationType = "rejected"
)

type Notification struct {
	ID          string             `gorm:"primaryKey" json:"id"`
	UserID      string             `gorm:"index" json:"user_id"`
	Type        NotificationType   `json:"type"`
	Title       string             `json:"title"`
	Message     string             `gorm:"type:text" json:"message"`
	TaskID      string             `json:"task_id"`
	IsRead      bool               `gorm:"index" json:"is_read"`
	ReadAt      *time.Time         `json:"read_at"`
	CreatedAt   time.Time          `json:"created_at"`

	// Relations
	User        User               `gorm:"foreignKey:UserID"`
}

func (Notification) TableName() string {
	return "notifications"
}
```

---

## Database Setup

### PostgreSQL Connection

```go
package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB initializes the database connection
func InitDB() error {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = db
	return nil
}

// RunMigrations runs all database migrations
func RunMigrations() error {
	return DB.AutoMigrate(
		&User{},
		&Session{},
		&ApprovalTask{},
		&ApprovalHistory{},
		&Document{},
		&Requisition{},
		&PurchaseOrder{},
		&PaymentVoucher{},
		&GRN{},
		&Workflow{},
		&AuditLog{},
		&Notification{},
	)
}
```

### Create Tables SQL

```sql
-- Users table
CREATE TABLE users (
  id VARCHAR(36) PRIMARY KEY,
  email VARCHAR(255) UNIQUE NOT NULL,
  name VARCHAR(255) NOT NULL,
  role VARCHAR(50) NOT NULL,
  department VARCHAR(255),
  is_active BOOLEAN DEFAULT true,
  last_login TIMESTAMP NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_role (role),
  INDEX idx_email (email)
);

-- Sessions table
CREATE TABLE sessions (
  id VARCHAR(36) PRIMARY KEY,
  user_id VARCHAR(36) NOT NULL,
  token VARCHAR(500) UNIQUE NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  INDEX idx_user_id (user_id),
  INDEX idx_expires_at (expires_at)
);

-- Approval tasks table
CREATE TABLE approval_tasks (
  id VARCHAR(36) PRIMARY KEY,
  entity_id VARCHAR(36) NOT NULL,
  entity_type VARCHAR(50) NOT NULL,
  entity_number VARCHAR(100) NOT NULL,
  status VARCHAR(50) NOT NULL,
  stage_name VARCHAR(255),
  stage_index INT,
  importance VARCHAR(50),
  approver_user_id VARCHAR(36) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  due_date TIMESTAMP,
  workflow_id VARCHAR(36),
  workflow_name VARCHAR(255),
  document JSONB,
  FOREIGN KEY (approver_user_id) REFERENCES users(id),
  INDEX idx_status (status),
  INDEX idx_entity_id (entity_id),
  INDEX idx_approver (approver_user_id),
  INDEX idx_created (created_at)
);

-- Approval history table
CREATE TABLE approval_history (
  id VARCHAR(36) PRIMARY KEY,
  task_id VARCHAR(36) NOT NULL,
  action VARCHAR(50) NOT NULL,
  approver_user_id VARCHAR(36) NOT NULL,
  timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  signature TEXT,
  remarks TEXT,
  previous_approver VARCHAR(36),
  FOREIGN KEY (task_id) REFERENCES approval_tasks(id) ON DELETE CASCADE,
  FOREIGN KEY (approver_user_id) REFERENCES users(id),
  INDEX idx_task_id (task_id),
  INDEX idx_action (action),
  INDEX idx_timestamp (timestamp)
);

-- Documents table (base)
CREATE TABLE documents (
  id VARCHAR(36) PRIMARY KEY,
  type VARCHAR(50) NOT NULL,
  number VARCHAR(100) UNIQUE NOT NULL,
  status VARCHAR(50),
  creator_id VARCHAR(36),
  data JSONB,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_type (type),
  INDEX idx_number (number),
  INDEX idx_created (created_at)
);

-- Audit logs table
CREATE TABLE audit_logs (
  id VARCHAR(36) PRIMARY KEY,
  user_id VARCHAR(36) NOT NULL,
  action VARCHAR(100) NOT NULL,
  entity_id VARCHAR(36),
  entity_type VARCHAR(50),
  old_value JSONB,
  new_value JSONB,
  timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  ip_address VARCHAR(45),
  user_agent TEXT,
  FOREIGN KEY (user_id) REFERENCES users(id),
  INDEX idx_user_id (user_id),
  INDEX idx_action (action),
  INDEX idx_timestamp (timestamp)
);

-- Notifications table
CREATE TABLE notifications (
  id VARCHAR(36) PRIMARY KEY,
  user_id VARCHAR(36) NOT NULL,
  type VARCHAR(50) NOT NULL,
  title VARCHAR(255),
  message TEXT,
  task_id VARCHAR(36),
  is_read BOOLEAN DEFAULT false,
  read_at TIMESTAMP NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  INDEX idx_user_id (user_id),
  INDEX idx_is_read (is_read),
  INDEX idx_created (created_at)
);

-- Workflows table
CREATE TABLE workflows (
  id VARCHAR(36) PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  description TEXT,
  entity_type VARCHAR(50) NOT NULL,
  status VARCHAR(50),
  stages JSONB,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  created_by VARCHAR(36),
  INDEX idx_entity_type (entity_type),
  INDEX idx_status (status)
);
```

---

## Fiber Application Setup

### Main Application

```go
package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"

	"liyali-api/database"
	"liyali-api/routes"
	"liyali-api/middleware"
)

func main() {
	// Initialize database
	if err := database.InitDB(); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	// Run migrations
	if err := database.RunMigrations(); err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Liyali Gateway API",
		ErrorHandler: middleware.ErrorHandler,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: os.Getenv("ALLOWED_ORIGINS"),
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Content-Type,Authorization",
	}))

	// Routes
	routes.SetupRoutes(app)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	log.Printf("Server starting on :%s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
```

### Routes Setup

```go
package routes

import (
	"github.com/gofiber/fiber/v3"

	"liyali-api/handlers"
	"liyali-api/middleware"
)

func SetupRoutes(app *fiber.App) {
	// Health check
	app.Get("/health", handlers.HealthCheck)

	// Auth routes (no auth required)
	authGroup := app.Group("/api/auth")
	authGroup.Post("/login", handlers.Login)
	authGroup.Post("/logout", handlers.Logout)
	authGroup.Post("/refresh", handlers.RefreshToken)

	// Protected routes - require authentication
	api := app.Group("/api", middleware.AuthMiddleware)

	// Approval tasks routes
	approvalsGroup := api.Group("/approvals")
	approvalsGroup.Get("/tasks", handlers.GetApprovalTasks)
	approvalsGroup.Get("/tasks/:id", handlers.GetApprovalTaskDetail)
	approvalsGroup.Post("/tasks/:id/approve", handlers.ApproveTask)
	approvalsGroup.Post("/tasks/:id/reject", handlers.RejectTask)
	approvalsGroup.Post("/tasks/:id/reassign", handlers.ReassignTask)

	// Bulk operations
	bulkGroup := api.Group("/approvals/bulk")
	bulkGroup.Post("/approve", handlers.BulkApprove)
	bulkGroup.Post("/reject", handlers.BulkReject)
	bulkGroup.Post("/reassign", handlers.BulkReassign)

	// Analytics routes
	analyticsGroup := api.Group("/analytics")
	analyticsGroup.Get("/metrics", handlers.GetAnalyticsMetrics)
	analyticsGroup.Get("/trends", handlers.GetWorkflowTrends)
	analyticsGroup.Get("/bottleneck", handlers.GetBottleneckAnalysis)

	// Workflow routes
	workflowGroup := api.Group("/workflows")
	workflowGroup.Get("", handlers.GetWorkflows)
	workflowGroup.Get("/:id", handlers.GetWorkflowDetail)
	workflowGroup.Post("", handlers.CreateWorkflow)
	workflowGroup.Put("/:id", handlers.UpdateWorkflow)

	// Notifications
	notifyGroup := api.Group("/notifications")
	notifyGroup.Get("", handlers.GetNotifications)
	notifyGroup.Post("/:id/read", handlers.MarkNotificationRead)
	notifyGroup.Delete("/:id", handlers.DeleteNotification)
}
```

---

## Handler Implementation

### Approval Handler Example

```go
package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"liyali-api/database"
	"liyali-api/models"
	"liyali-api/middleware"
)

// ApproveTaskRequest represents the request body for approving a task
type ApproveTaskRequest struct {
	AssignmentID string `json:"assignment_id" validate:"required"`
	StageNumber  int    `json:"stage_number" validate:"required,min=0"`
	Signature    string `json:"signature" validate:"required"`
	Comments     string `json:"comments"`
}

// ApproveTaskResponse represents the response after approving
type ApproveTaskResponse struct {
	Success   bool      `json:"success"`
	TaskID    string    `json:"task_id"`
	Action    string    `json:"action"`
	NewStatus string    `json:"new_status"`
	NextStage string    `json:"next_stage,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// ApproveTask handles task approval
func ApproveTask(c fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	taskID := c.Params("id")
	var req ApproveTaskRequest

	if err := c.BindJSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Validate request
	if req.AssignmentID == "" || req.Signature == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Missing required fields",
		})
	}

	// Start transaction
	tx := database.DB.BeginTx(c.Context(), nil)

	// Get task
	var task models.ApprovalTask
	if err := tx.First(&task, "id = ?", taskID).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"error":   "Task not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Database error",
		})
	}

	// Verify user is the approver
	if task.ApproverUserID != user.ID {
		tx.Rollback()
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   "User is not the assigned approver",
		})
	}

	// Create history record
	history := models.ApprovalHistory{
		ID:             uuid.New().String(),
		TaskID:         taskID,
		Action:         models.ActionApproved,
		ApproverUserID: user.ID,
		Timestamp:      time.Now(),
		Signature:      req.Signature,
		Remarks:        req.Comments,
	}

	if err := tx.Create(&history).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create history record",
		})
	}

	// Update task status
	task.Status = models.StatusApproved
	if err := tx.Save(&task).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to update task",
		})
	}

	// Log audit
	auditLog := models.AuditLog{
		ID:         uuid.New().String(),
		UserID:     user.ID,
		Action:     "approve_task",
		EntityID:   taskID,
		EntityType: string(task.EntityType),
		Timestamp:  time.Now(),
		IPAddress:  c.IP(),
	}

	if err := tx.Create(&auditLog).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to log action",
		})
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to commit transaction",
		})
	}

	return c.Status(fiber.StatusOK).JSON(ApproveTaskResponse{
		Success:   true,
		TaskID:    taskID,
		Action:    "approved",
		NewStatus: string(models.StatusApproved),
		Timestamp: time.Now(),
	})
}

// GetApprovalTasks retrieves all approval tasks with optional filtering
func GetApprovalTasks(c fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	// Get query parameters
	status := c.Query("status", "")
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Build query
	query := database.DB.Preload("ApproverUser")

	// Filter by user's tasks
	query = query.Where("approver_user_id = ?", user.ID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Get total count
	var total int64
	if err := query.Model(&models.ApprovalTask{}).Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to count tasks",
		})
	}

	// Get paginated results
	var tasks []models.ApprovalTask
	if err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&tasks).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to fetch tasks",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"tasks":     tasks,
			"total":     total,
			"page":      page,
			"limit":     limit,
			"pageCount": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// RejectTask handles task rejection
func RejectTask(c fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	taskID := c.Params("id")

	var req struct {
		Signature string `json:"signature" validate:"required"`
		Remarks   string `json:"remarks" validate:"required"`
	}

	if err := c.BindJSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	tx := database.DB.BeginTx(c.Context(), nil)

	var task models.ApprovalTask
	if err := tx.First(&task, "id = ?", taskID).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Task not found",
		})
	}

	if task.ApproverUserID != user.ID {
		tx.Rollback()
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   "User is not the assigned approver",
		})
	}

	history := models.ApprovalHistory{
		ID:             uuid.New().String(),
		TaskID:         taskID,
		Action:         models.ActionRejected,
		ApproverUserID: user.ID,
		Timestamp:      time.Now(),
		Signature:      req.Signature,
		Remarks:        req.Remarks,
	}

	if err := tx.Create(&history).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create history record",
		})
	}

	task.Status = models.StatusRejected
	task.StageIndex = 0 // Reset to first stage

	if err := tx.Save(&task).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to update task",
		})
	}

	tx.Commit()

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"task_id":   taskID,
			"action":    "rejected",
			"new_status": string(models.StatusRejected),
			"reason":    req.Remarks,
			"timestamp": time.Now(),
		},
	})
}

// ReassignTask handles task reassignment
func ReassignTask(c fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	taskID := c.Params("id")

	var req struct {
		NewApproverID   string `json:"new_approver_id" validate:"required"`
		NewApproverName string `json:"new_approver_name" validate:"required"`
		Reason          string `json:"reason"`
	}

	if err := c.BindJSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	tx := database.DB.BeginTx(c.Context(), nil)

	var task models.ApprovalTask
	if err := tx.First(&task, "id = ?", taskID).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Task not found",
		})
	}

	if task.ApproverUserID != user.ID {
		tx.Rollback()
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   "User is not the assigned approver",
		})
	}

	previousApprover := task.ApproverUserID
	task.ApproverUserID = req.NewApproverID

	history := models.ApprovalHistory{
		ID:               uuid.New().String(),
		TaskID:           taskID,
		Action:           models.ActionReassigned,
		ApproverUserID:   req.NewApproverID,
		Timestamp:        time.Now(),
		Remarks:          req.Reason,
		PreviousApprover: &previousApprover,
	}

	if err := tx.Create(&history).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create history record",
		})
	}

	if err := tx.Save(&task).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to update task",
		})
	}

	tx.Commit()

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"task_id":      taskID,
			"action":       "reassigned",
			"new_approver": req.NewApproverName,
			"timestamp":    time.Now(),
		},
	})
}
```

### Bulk Operations Handler

```go
// BulkApprove approves multiple tasks
func BulkApprove(c fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	var req struct {
		TaskIDs []string `json:"task_ids" validate:"required,min=1"`
		Remarks string   `json:"remarks"`
	}

	if err := c.BindJSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	tx := database.DB.BeginTx(c.Context(), nil)
	approved := 0
	failed := 0

	for _, taskID := range req.TaskIDs {
		var task models.ApprovalTask
		if err := tx.First(&task, "id = ?", taskID).Error; err != nil {
			failed++
			continue
		}

		if task.ApproverUserID != user.ID {
			failed++
			continue
		}

		history := models.ApprovalHistory{
			ID:             uuid.New().String(),
			TaskID:         taskID,
			Action:         models.ActionApproved,
			ApproverUserID: user.ID,
			Timestamp:      time.Now(),
			Remarks:        req.Remarks,
		}

		if err := tx.Create(&history).Error; err != nil {
			failed++
			continue
		}

		task.Status = models.StatusApproved
		if err := tx.Save(&task).Error; err != nil {
			failed++
			continue
		}

		approved++
	}

	tx.Commit()

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"approved":  approved,
			"failed":    failed,
			"message":   fmt.Sprintf("Successfully approved %d tasks", approved),
			"timestamp": time.Now(),
		},
	})
}
```

### Analytics Handler

```go
// GetAnalyticsMetrics returns dashboard metrics
func GetAnalyticsMetrics(c fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	var pending, approved, rejected int64
	database.DB.Model(&models.ApprovalTask{}).
		Where("approver_user_id = ?", user.ID).
		Where("status = ?", models.StatusPending).
		Count(&pending)

	database.DB.Model(&models.ApprovalTask{}).
		Where("approver_user_id = ?", user.ID).
		Where("status = ?", models.StatusApproved).
		Count(&approved)

	database.DB.Model(&models.ApprovalTask{}).
		Where("approver_user_id = ?", user.ID).
		Where("status = ?", models.StatusRejected).
		Count(&rejected)

	// Calculate SLA compliance (tasks completed within 2 days)
	var completedInTime int64
	database.DB.Model(&models.ApprovalTask{}).
		Where("approver_user_id = ?", user.ID).
		Where("status = ?", models.StatusApproved).
		Where("EXTRACT(EPOCH FROM (updated_at - created_at))/86400 <= ?", 2).
		Count(&completedInTime)

	slaCompliance := 0
	if approved > 0 {
		slaCompliance = int((completedInTime * 100) / approved)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"total_pending":     pending,
			"total_approved":    approved,
			"total_rejected":    rejected,
			"avg_approval_time": "2.5 days",
			"sla_compliance":    slaCompliance,
		},
	})
}

// GetWorkflowTrends returns 7-day approval trends
func GetWorkflowTrends(c fiber.Ctx) error {
	user := middleware.GetUserFromContext(c)
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized",
		})
	}

	type TrendData struct {
		Date     string `json:"date"`
		Approved int64  `json:"approved"`
		Rejected int64  `json:"rejected"`
		Pending  int64  `json:"pending"`
	}

	var trends []TrendData

	// Query last 7 days of data
	for i := 6; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		startOfDay := time.Now().AddDate(0, 0, -i).Truncate(24 * time.Hour)
		endOfDay := startOfDay.Add(24 * time.Hour)

		var approved, rejected, pending int64

		database.DB.Model(&models.ApprovalTask{}).
			Where("approver_user_id = ? AND created_at >= ? AND created_at < ? AND status = ?",
				user.ID, startOfDay, endOfDay, models.StatusApproved).
			Count(&approved)

		database.DB.Model(&models.ApprovalTask{}).
			Where("approver_user_id = ? AND created_at >= ? AND created_at < ? AND status = ?",
				user.ID, startOfDay, endOfDay, models.StatusRejected).
			Count(&rejected)

		database.DB.Model(&models.ApprovalTask{}).
			Where("approver_user_id = ? AND created_at >= ? AND created_at < ? AND status = ?",
				user.ID, startOfDay, endOfDay, models.StatusPending).
			Count(&pending)

		trends = append(trends, TrendData{
			Date:     date,
			Approved: approved,
			Rejected: rejected,
			Pending:  pending,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    trends,
	})
}
```

---

## Authentication & Middleware

### Auth Middleware

```go
package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"

	"liyali-api/database"
	"liyali-api/models"
)

// UserContext key for storing user in context
type UserContextKey string

const UserContextKeyValue UserContextKey = "user"

// AuthMiddleware validates JWT tokens
func AuthMiddleware(c fiber.Ctx) error {
	authHeader := c.Get("Authorization")

	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Missing authorization header",
		})
	}

	// Extract token from "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid authorization format",
		})
	}

	tokenString := parts[1]

	// Verify token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid token",
		})
	}

	claims := token.Claims.(jwt.MapClaims)
	userID := claims["user_id"].(string)

	// Get user from database
	var user models.User
	if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "User not found",
		})
	}

	// Store user in context
	c.Locals(string(UserContextKeyValue), &user)

	return c.Next()
}

// GetUserFromContext retrieves user from context
func GetUserFromContext(c fiber.Ctx) *models.User {
	user := c.Locals(string(UserContextKeyValue))
	if user == nil {
		return nil
	}
	return user.(*models.User)
}

// RoleMiddleware checks if user has required role
func RoleMiddleware(requiredRoles ...models.UserRole) fiber.Handler {
	return func(c fiber.Ctx) error {
		user := GetUserFromContext(c)
		if user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Unauthorized",
			})
		}

		allowed := false
		for _, role := range requiredRoles {
			if user.Role == role {
				allowed = true
				break
			}
		}

		if !allowed {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error":   "Insufficient permissions",
			})
		}

		return c.Next()
	}
}
```

---

## Error Handling

### Global Error Handler

```go
package middleware

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

// AppError represents an application error
type AppError struct {
	Code    int
	Message string
	Details interface{}
}

// ErrorHandler is the global error handler
func ErrorHandler(c fiber.Ctx, err error) error {
	log.Printf("Error: %v", err)

	// Fiber error
	var fe *fiber.Error
	if errors.As(err, &fe) {
		return c.Status(fe.Code).JSON(fiber.Map{
			"success": false,
			"error":   fe.Message,
		})
	}

	// GORM errors
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Record not found",
		})
	}

	// Default error
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"success": false,
		"error":   "Internal server error",
	})
}
```

---

## Database Optimization

### Indexes

```sql
-- Performance indexes for common queries
CREATE INDEX idx_approval_tasks_approver_status
ON approval_tasks(approver_user_id, status);

CREATE INDEX idx_approval_tasks_entity
ON approval_tasks(entity_type, entity_id);

CREATE INDEX idx_approval_history_task
ON approval_history(task_id, action);

CREATE INDEX idx_audit_logs_user_action
ON audit_logs(user_id, action, timestamp);

CREATE INDEX idx_notifications_user_read
ON notifications(user_id, is_read, created_at);
```

### Query Optimization

```go
// Use Preload for eager loading instead of N+1 queries
query := database.DB.
	Preload("ApproverUser").
	Preload("History").
	Where("status = ?", models.StatusPending).
	Find(&tasks)

// Use Select to fetch only needed columns
query := database.DB.
	Select("id", "entity_id", "status", "created_at").
	Where("approver_user_id = ?", userID).
	Find(&tasks)

// Batch operations
var taskIDs []string
for _, id := range ids {
	taskIDs = append(taskIDs, id)
}
database.DB.Where("id IN ?", taskIDs).Updates(map[string]interface{}{
	"status": models.StatusApproved,
})
```

---

## Performance Tips

1. **Connection Pooling**: Configured with max 100 open connections
2. **Caching**: Consider Redis for frequently accessed metrics
3. **Pagination**: Always use LIMIT/OFFSET for large result sets
4. **Batch Operations**: Process multiple updates in single transaction
5. **Indexes**: Created on foreign keys and commonly filtered columns
6. **N+1 Prevention**: Use Preload for related data

---

## NoSQL Considerations

While PostgreSQL is recommended for this workflow system, consider NoSQL (MongoDB) in these scenarios:

1. **High-Volume Event Logging**: Use MongoDB for immutable audit logs
2. **Time-Series Data**: Store analytics data in a time-series collection
3. **Document Variants**: If document structures vary significantly by type
4. **Real-time Updates**: Consider MongoDB change streams for live dashboards

**Best Practice**: Use PostgreSQL as primary database for ACID transactions, add MongoDB for high-volume, unstructured data (audit logs, notifications).

```go
// Example MongoDB connection for audit logs
import "go.mongodb.org/mongo-driver/mongo"

var mongoClient *mongo.Client

func InitMongoDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().
		ApplyURI(os.Getenv("MONGODB_URI")))

	mongoClient = client
	return err
}

// Store audit logs in MongoDB for long-term retention
func LogToMongoDB(log *models.AuditLog) error {
	collection := mongoClient.Database("liyali").Collection("audit_logs")
	_, err := collection.InsertOne(context.Background(), log)
	return err
}
```

---

**Status**: Ready for Phase 12 Implementation
**Next**: Create Node.js/Prisma ORM equivalent guide
