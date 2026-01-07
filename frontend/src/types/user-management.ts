export interface UserRoleAssignment {
  userId: string
  customRoleId: string
  assignedAt: Date
  assignedBy: string
}

export interface CreateUserRequest {
  email: string;
  name?: string;
  role: string;
  department?: string;
  departmentId?: string;
  password?: string;
  active?: boolean;
}

export interface CreateOrganizationRequest {
  name: string;
  slug?: string;
  description?: string;
  address?: string;
  logoUrl?: string;
  primaryColor?: string;
  tier?: string;
}
