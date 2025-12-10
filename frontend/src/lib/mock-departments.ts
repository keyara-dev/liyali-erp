/**
 * Mock Departments Data with localStorage Support
 * Provides both initial mock data and persistent storage capabilities
 */

export interface Department {
  id: string;
  name: string;
  code: string;
  description: string;
  manager_name: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

// Initial mock departments data
const INITIAL_DEPARTMENTS: Department[] = [
  {
    id: "dept-001",
    name: "Operations",
    code: "OPS",
    description: "Handles day-to-day operational activities and logistics",
    manager_name: "James Chileshe",
    is_active: true,
    created_at: "2025-01-01T08:00:00Z",
    updated_at: "2025-01-01T08:00:00Z",
  },
  {
    id: "dept-002",
    name: "Human Resources",
    code: "HR",
    description: "Manages employee recruitment, development, and relations",
    manager_name: "Maria Chiyanda",
    is_active: true,
    created_at: "2025-01-01T08:00:00Z",
    updated_at: "2025-01-01T08:00:00Z",
  },
  {
    id: "dept-003",
    name: "Finance",
    code: "FIN",
    description: "Manages financial planning, accounting, and reporting",
    manager_name: "David Mwende",
    is_active: true,
    created_at: "2025-01-01T08:00:00Z",
    updated_at: "2025-01-01T08:00:00Z",
  },
  {
    id: "dept-004",
    name: "Compliance & Audit",
    code: "CAU",
    description: "Ensures regulatory compliance and conducts internal audits",
    manager_name: "Alice Nkonde",
    is_active: true,
    created_at: "2025-01-01T08:00:00Z",
    updated_at: "2025-01-01T08:00:00Z",
  },
  {
    id: "dept-005",
    name: "IT & Systems",
    code: "ITS",
    description: "Manages IT infrastructure, systems, and digital transformation",
    manager_name: "Robert Chanda",
    is_active: true,
    created_at: "2025-01-01T08:00:00Z",
    updated_at: "2025-01-01T08:00:00Z",
  },
];

const STORAGE_KEY = "mock_departments";

/**
 * Initialize departments from localStorage or use initial mock data
 */
export function getInitialDepartments(): Department[] {
  if (typeof window === "undefined") {
    return INITIAL_DEPARTMENTS;
  }

  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      return JSON.parse(stored);
    }
  } catch (error) {
    console.error("Failed to load departments from localStorage:", error);
  }

  return INITIAL_DEPARTMENTS;
}

/**
 * Save departments to localStorage
 */
export function saveDepartmentsToStorage(departments: Department[]): void {
  if (typeof window === "undefined") {
    return;
  }

  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(departments));
  } catch (error) {
    console.error("Failed to save departments to localStorage:", error);
  }
}

/**
 * Get all departments (from storage if available)
 */
export function getAllDepartments(): Department[] {
  return getInitialDepartments();
}

/**
 * Get department by ID
 */
export function getDepartmentById(id: string): Department | undefined {
  return getInitialDepartments().find((dept) => dept.id === id);
}

/**
 * Get active departments only
 */
export function getActiveDepartments(): Department[] {
  return getInitialDepartments().filter((dept) => dept.is_active);
}

/**
 * Create a new department
 */
export function createDepartment(
  data: Omit<Department, "id" | "created_at" | "updated_at">
): Department {
  const newDepartment: Department = {
    ...data,
    id: `dept-${Date.now()}`,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  };

  const departments = getInitialDepartments();
  departments.push(newDepartment);
  saveDepartmentsToStorage(departments);

  return newDepartment;
}

/**
 * Update an existing department
 */
export function updateDepartment(
  id: string,
  data: Partial<Omit<Department, "id" | "created_at">>
): Department | null {
  const departments = getInitialDepartments();
  const index = departments.findIndex((dept) => dept.id === id);

  if (index === -1) {
    return null;
  }

  departments[index] = {
    ...departments[index],
    ...data,
    updated_at: new Date().toISOString(),
  };

  saveDepartmentsToStorage(departments);
  return departments[index];
}

/**
 * Delete a department (soft delete - sets is_active to false)
 */
export function deleteDepartment(id: string): boolean {
  const departments = getInitialDepartments();
  const department = departments.find((dept) => dept.id === id);

  if (!department) {
    return false;
  }

  department.is_active = false;
  department.updated_at = new Date().toISOString();
  saveDepartmentsToStorage(departments);

  return true;
}

/**
 * Restore a deleted department
 */
export function restoreDepartment(id: string): boolean {
  const departments = getInitialDepartments();
  const department = departments.find((dept) => dept.id === id);

  if (!department) {
    return false;
  }

  department.is_active = true;
  department.updated_at = new Date().toISOString();
  saveDepartmentsToStorage(departments);

  return true;
}

/**
 * Reset departments to initial mock data
 * Useful for testing
 */
export function resetDepartmentsToInitial(): void {
  saveDepartmentsToStorage(INITIAL_DEPARTMENTS);
}

/**
 * Clear all departments from storage
 */
export function clearDepartmentsStorage(): void {
  if (typeof window === "undefined") {
    return;
  }

  try {
    localStorage.removeItem(STORAGE_KEY);
  } catch (error) {
    console.error("Failed to clear departments storage:", error);
  }
}

/**
 * Export INITIAL_DEPARTMENTS for use in other components
 */
export const MOCK_DEPARTMENTS = INITIAL_DEPARTMENTS;
