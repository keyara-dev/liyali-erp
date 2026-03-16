package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"github.com/liyali/liyali-gateway/models"
)

type DepartmentService struct {
	db *gorm.DB
}

func NewDepartmentService(db *gorm.DB) *DepartmentService {
	return &DepartmentService{db: db}
}

// GetAllDepartments retrieves all departments for an organization with pagination
func (s *DepartmentService) GetAllDepartments(organizationID string, page, pageSize int) ([]interface{}, int64, error) {
	var departments []models.OrganizationDepartment
	var total int64

	offset := (page - 1) * pageSize

	// Count total departments
	countQ := s.db.Model(&models.OrganizationDepartment{})
	if organizationID != "" {
		countQ = countQ.Where("organization_id = ?", organizationID)
	}
	if err := countQ.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count departments: %w", err)
	}

	// Get departments with pagination
	listQ := s.db.Model(&models.OrganizationDepartment{})
	if organizationID != "" {
		listQ = listQ.Where("organization_id = ?", organizationID)
	}
	if err := listQ.Order("name ASC").Limit(pageSize).Offset(offset).Find(&departments).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch departments: %w", err)
	}

	// Convert to interface slice
	result := make([]interface{}, len(departments))
	for i, dept := range departments {
		result[i] = dept
	}

	return result, total, nil
}

// GetActiveDepartments retrieves only active departments
func (s *DepartmentService) GetActiveDepartments(organizationID string, page, pageSize int) ([]interface{}, int64, error) {
	var departments []models.OrganizationDepartment
	var total int64

	offset := (page - 1) * pageSize

	// Count total active departments
	countQ := s.db.Model(&models.OrganizationDepartment{}).Where("is_active = ?", true)
	if organizationID != "" {
		countQ = countQ.Where("organization_id = ?", organizationID)
	}
	if err := countQ.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count active departments: %w", err)
	}

	// Get active departments with pagination
	listQ := s.db.Model(&models.OrganizationDepartment{}).Where("is_active = ?", true)
	if organizationID != "" {
		listQ = listQ.Where("organization_id = ?", organizationID)
	}
	if err := listQ.Order("name ASC").Limit(pageSize).Offset(offset).Find(&departments).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch active departments: %w", err)
	}

	// Convert to interface slice
	result := make([]interface{}, len(departments))
	for i, dept := range departments {
		result[i] = dept
	}

	return result, total, nil
}

// GetInactiveDepartments retrieves only inactive departments
func (s *DepartmentService) GetInactiveDepartments(organizationID string, page, pageSize int) ([]interface{}, int64, error) {
	var departments []models.OrganizationDepartment
	var total int64

	offset := (page - 1) * pageSize

	// Count total inactive departments
	countQ := s.db.Model(&models.OrganizationDepartment{}).Where("is_active = ?", false)
	if organizationID != "" {
		countQ = countQ.Where("organization_id = ?", organizationID)
	}
	if err := countQ.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count inactive departments: %w", err)
	}

	// Get inactive departments with pagination
	listQ := s.db.Model(&models.OrganizationDepartment{}).Where("is_active = ?", false)
	if organizationID != "" {
		listQ = listQ.Where("organization_id = ?", organizationID)
	}
	if err := listQ.Order("name ASC").Limit(pageSize).Offset(offset).Find(&departments).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch inactive departments: %w", err)
	}

	// Convert to interface slice
	result := make([]interface{}, len(departments))
	for i, dept := range departments {
		result[i] = dept
	}

	return result, total, nil
}

// GetDepartmentByID retrieves a specific department by ID
func (s *DepartmentService) GetDepartmentByID(organizationID, departmentID string) (*models.OrganizationDepartment, error) {
	var department models.OrganizationDepartment

	if err := s.db.Where("id = ? AND organization_id = ?", departmentID, organizationID).
		First(&department).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("department not found")
		}
		return nil, fmt.Errorf("failed to fetch department: %w", err)
	}

	return &department, nil
}

// CreateDepartment creates a new department
func (s *DepartmentService) CreateDepartment(organizationID, name, code string, description, managerName, parentID *string) (*models.OrganizationDepartment, error) {
	department := models.OrganizationDepartment{
		ID:             uuid.New().String(),
		OrganizationID: organizationID,
		Name:           name,
		Code:           code,
		ParentID:       parentID,
		IsActive:       true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	
	// Set description if provided
	if description != nil {
		department.Description = *description
	}

	// Set manager name if provided
	if managerName != nil {
		department.ManagerName = *managerName
	}

	if err := s.db.Create(&department).Error; err != nil {
		return nil, fmt.Errorf("failed to create department: %w", err)
	}

	return &department, nil
}

// UpdateDepartment updates an existing department
func (s *DepartmentService) UpdateDepartment(organizationID, departmentID string, name, code, description, managerName, parentID *string, isActive *bool) (*models.OrganizationDepartment, error) {
	var department models.OrganizationDepartment

	// Find the department
	if err := s.db.Where("id = ? AND organization_id = ?", departmentID, organizationID).
		First(&department).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("department not found")
		}
		return nil, fmt.Errorf("failed to find department: %w", err)
	}

	// Update fields if provided
	updates := make(map[string]interface{})
	
	if name != nil {
		updates["name"] = *name
	}
	if code != nil {
		updates["code"] = *code
	}
	if description != nil {
		updates["description"] = *description
	}
	if managerName != nil {
		updates["manager_name"] = *managerName
	}
	if parentID != nil {
		updates["parent_id"] = *parentID
	}
	if isActive != nil {
		updates["is_active"] = *isActive
	}
	
	updates["updated_at"] = time.Now()

	// Perform update
	if err := s.db.Model(&department).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update department: %w", err)
	}

	// Reload the department to get updated values
	if err := s.db.Where("id = ? AND organization_id = ?", departmentID, organizationID).
		First(&department).Error; err != nil {
		return nil, fmt.Errorf("failed to reload department: %w", err)
	}

	return &department, nil
}

// DeleteDepartment soft deletes a department
func (s *DepartmentService) DeleteDepartment(organizationID, departmentID string) error {
	result := s.db.Model(&models.OrganizationDepartment{}).
		Where("id = ? AND organization_id = ?", departmentID, organizationID).
		Updates(map[string]interface{}{
			"is_active":  false,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to delete department: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("department not found")
	}

	return nil
}

// RestoreDepartment restores a soft deleted department
func (s *DepartmentService) RestoreDepartment(organizationID, departmentID string) (*models.OrganizationDepartment, error) {
	var department models.OrganizationDepartment

	// Update the department to active
	result := s.db.Model(&models.OrganizationDepartment{}).
		Where("id = ? AND organization_id = ?", departmentID, organizationID).
		Updates(map[string]interface{}{
			"is_active":  true,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return nil, fmt.Errorf("failed to restore department: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("department not found")
	}

	// Reload the department
	if err := s.db.Where("id = ? AND organization_id = ?", departmentID, organizationID).
		First(&department).Error; err != nil {
		return nil, fmt.Errorf("failed to reload department: %w", err)
	}

	return &department, nil
}

// DepartmentExists checks if a department exists in the organization
func (s *DepartmentService) DepartmentExists(organizationID, departmentID string) (bool, error) {
	var count int64
	
	if err := s.db.Model(&models.OrganizationDepartment{}).
		Where("id = ? AND organization_id = ?", departmentID, organizationID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check department existence: %w", err)
	}

	return count > 0, nil
}

// DepartmentCodeExists checks if a department code already exists in the organization
func (s *DepartmentService) DepartmentCodeExists(organizationID, code string) (bool, error) {
	var count int64
	
	if err := s.db.Model(&models.OrganizationDepartment{}).
		Where("organization_id = ? AND code = ?", organizationID, code).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check department code: %w", err)
	}

	return count > 0, nil
}

// DepartmentCodeExistsExcluding checks if a department code exists excluding a specific department
func (s *DepartmentService) DepartmentCodeExistsExcluding(organizationID, code, excludeDepartmentID string) (bool, error) {
	var count int64
	
	if err := s.db.Model(&models.OrganizationDepartment{}).
		Where("organization_id = ? AND code = ? AND id != ?", organizationID, code, excludeDepartmentID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check department code: %w", err)
	}

	return count > 0, nil
}

// GetDepartmentModules retrieves modules assigned to a department
func (s *DepartmentService) GetDepartmentModules(organizationID, departmentID string) ([]interface{}, error) {
	// This is a placeholder implementation
	// You'll need to implement the actual module assignment logic based on your module system
	
	// For now, return an empty slice
	// In a real implementation, you would:
	// 1. Query the department_modules table (if it exists)
	// 2. Join with modules table to get module details
	// 3. Return the list of modules assigned to this department
	
	modules := make([]interface{}, 0)
	return modules, nil
}

// AssignModuleToDepartment assigns a module to a department
func (s *DepartmentService) AssignModuleToDepartment(organizationID, departmentID, moduleID string) error {
	// This is a placeholder implementation
	// You'll need to implement the actual module assignment logic based on your module system
	
	// In a real implementation, you would:
	// 1. Check if the module exists
	// 2. Check if the assignment already exists
	// 3. Create a new record in department_modules table
	
	return nil
}

// RemoveModuleFromDepartment removes a module assignment from a department
func (s *DepartmentService) RemoveModuleFromDepartment(organizationID, departmentID, moduleID string) error {
	// This is a placeholder implementation
	// You'll need to implement the actual module removal logic based on your module system
	
	// In a real implementation, you would:
	// 1. Check if the assignment exists
	// 2. Delete the record from department_modules table
	
	return nil
}