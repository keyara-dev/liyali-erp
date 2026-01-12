"use client";

import { useRouter } from "next/navigation";
import { useState, useEffect, useMemo, useCallback } from "react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Switch } from "@/components/ui/switch";
import { InfoIcon, ShieldIcon, Plus, Edit } from "lucide-react";
import { Spinner } from "@/components/ui/spinner";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { QUERY_KEYS } from "@/lib/constants";

import { toast } from "sonner";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { Skeleton } from "@/components/ui/skeleton";
import {
  bulkUpdateRolePermissions,
  getRolePermissions,
} from "@/app/_actions/roles-permissions";
import { getDepartmentModules } from "@/app/_actions/departments";
import { createRole, updateRole } from "@/app/_actions/roles-permissions";
import { ConfirmationModal } from "@/components/modals/confirmation-modal";

interface RolesPermissionsProps {
  departmentId: string;
}

interface Role {
  id: string;
  name: string;
  code: string;
  department_id: string;
  description: string;
  is_active: boolean;
  is_department_head: boolean;
}

interface Module {
  id: string;
  name: string;
  description?: string;
  parent_module_id?: string | null;
  module_code?: string;
}

type PermissionType =
  | "can_view"
  | "can_create"
  | "can_edit"
  | "can_delete"
  | "can_approve"
  | "can_export"
  | "can_assign"
  | "can_configure";

const PERMISSION_LABELS: Record<PermissionType, string> = {
  can_view: "View",
  can_create: "Create",
  can_edit: "Edit",
  can_delete: "Delete",
  can_approve: "Approve",
  can_export: "Export",
  can_assign: "Assign",
  can_configure: "Configure",
};

// Format module name - use module_code for "Overview" modules
const formatModuleName = (name: string, moduleCode?: string): string => {
  // If name is "Overview", use module_code instead
  if (name.toLowerCase().trim() === "overview" && moduleCode) {
    // Format: "RISK_OVERVIEW" -> "Risk Overview"
    return moduleCode
      .split("_")
      .map((word) => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
      .join(" ");
  }
  return name;
};

export default function UserRolesConfig({
  departmentId,
}: RolesPermissionsProps) {
  const router = useRouter();

  const queryClient = useQueryClient();
  const [selectedRole, setSelectedRole] = useState<string | null>(null);
  const [permissionsMatrix, setPermissionsMatrix] = useState<
    Record<string, Record<PermissionType, boolean>>
  >({});
  const [hasChanges, setHasChanges] = useState(false);
  const [openRoleModal, setOpenRoleModal] = useState(false);
  const [editingRole, setEditingRole] = useState<Role | null>(null);
  const [confirmSwitchRole, setConfirmSwitchRole] = useState(false);
  const [pendingRoleId, setPendingRoleId] = useState<string | null>(null);

  // Fetch roles for this department
  const { data: rolesData, isLoading: rolesLoading } = useQuery({
    queryKey: [QUERY_KEYS.ROLES, departmentId],
    queryFn: async () => {
      // Fetch roles from backend via action
      const response = await fetch(`/api/roles?departmentId=${departmentId}`);
      if (!response.ok) throw new Error("Failed to fetch roles");
      return response.json();
    },
    enabled: !!departmentId,
    staleTime: 5 * 60 * 1000,
  });

  const rolesResponse = rolesData || { success: true, data: { data: [] } };

  const roles: Role[] = useMemo(
    () =>
      rolesResponse?.success && rolesResponse.data.data
        ? rolesResponse.data.data
        : [],
    [rolesResponse]
  );

  // Fetch modules assigned to this department
  const { data: modulesResponse, isLoading: modulesLoading } = useQuery({
    queryKey: [QUERY_KEYS.DEPARTMENT_MODULES, departmentId],
    queryFn: () => getDepartmentModules(departmentId!),
    enabled: !!departmentId,
    staleTime: 5 * 60 * 1000,
  });

  const modules: Module[] = useMemo(() => {
    if (!modulesResponse?.success || !modulesResponse?.data) return [];
    // Handle both modulesResponse.data.data and modulesResponse.data as array
    const data = modulesResponse.data.data || modulesResponse.data;
    return Array.isArray(data) ? data : [];
  }, [modulesResponse]);

  // Fetch permissions for selected role
  const { data: permissionsResponse, isLoading: permissionsLoading } = useQuery(
    {
      queryKey: [QUERY_KEYS.ROLE_PERMISSIONS, selectedRole],
      queryFn: () => getRolePermissions(selectedRole!),
      enabled: !!selectedRole,
      staleTime: 5 * 60 * 1000,
    }
  );

  // Build permissions matrix when data loads
  useEffect(() => {
    if (
      permissionsResponse?.success &&
      permissionsResponse?.data &&
      selectedRole
    ) {
      const matrix: Record<string, Record<PermissionType, boolean>> = {};

      // Initialize all modules with false permissions
      modules.forEach((module) => {
        matrix[module.id] = {
          can_view: false,
          can_create: false,
          can_edit: false,
          can_delete: false,
          can_approve: false,
          can_export: false,
          can_assign: false,
          can_configure: false,
        };
      });

      // Fill in actual permissions from nested structure
      const responseData = permissionsResponse.data as any[];

      responseData.forEach((moduleData: any) => {
        // Handle parent module permissions
        if (moduleData.permissions && matrix[moduleData.id]) {
          matrix[moduleData.id] = {
            can_view: moduleData.permissions.can_view || false,
            can_create: moduleData.permissions.can_create || false,
            can_edit: moduleData.permissions.can_edit || false,
            can_delete: moduleData.permissions.can_delete || false,
            can_approve: moduleData.permissions.can_approve || false,
            can_export: moduleData.permissions.can_export || false,
            can_assign: moduleData.permissions.can_assign || false,
            can_configure: moduleData.permissions.can_configure || false,
          };
        }

        // Handle sub_modules permissions
        if (moduleData.sub_modules && Array.isArray(moduleData.sub_modules)) {
          moduleData.sub_modules.forEach((subModule: any) => {
            if (subModule.permissions && matrix[subModule.id]) {
              matrix[subModule.id] = {
                can_view: subModule.permissions.can_view || false,
                can_create: subModule.permissions.can_create || false,
                can_edit: subModule.permissions.can_edit || false,
                can_delete: subModule.permissions.can_delete || false,
                can_approve: subModule.permissions.can_approve || false,
                can_export: subModule.permissions.can_export || false,
                can_assign: subModule.permissions.can_assign || false,
                can_configure: subModule.permissions.can_configure || false,
              };
            }
          });
        }
      });

      setPermissionsMatrix(matrix);
      setHasChanges(false);
    }
  }, [permissionsResponse, selectedRole, modules]);

  // Set first role as selected when roles load
  useEffect(() => {
    if (roles.length > 0 && !selectedRole) {
      setSelectedRole(roles[0].id);
    }
  }, [roles, selectedRole]);

  // Toggle permission
  const togglePermission = (
    moduleId: string,
    permissionType: PermissionType
  ) => {
    setPermissionsMatrix((prev) => {
      const currentModulePerms = prev[moduleId] || {
        can_view: false,
        can_create: false,
        can_edit: false,
        can_delete: false,
        can_approve: false,
        can_export: false,
        can_assign: false,
        can_configure: false,
      };

      const newValue = !currentModulePerms[permissionType];

      // console.log(`🔄 Toggling ${permissionType} for module ${moduleId}:`, {
      //   oldValue: currentModulePerms[permissionType],
      //   newValue
      // });

      return {
        ...prev,
        [moduleId]: {
          ...currentModulePerms,
          [permissionType]: newValue,
        },
      };
    });
    setHasChanges(true);
  };

  // Save permissions mutation
  const savePermissionsMutation = useMutation({
    mutationFn: async () => {
      if (!selectedRole) {
        throw new Error("No role selected");
      }

      const permissions = modules
        .map((module) => {
          const modulePerms = permissionsMatrix[module.id];
          return {
            moduleId: module.id,
            parentModuleId: module.parent_module_id || null,
            canView: modulePerms?.can_view || false,
            canCreate: modulePerms?.can_create || false,
            canEdit: modulePerms?.can_edit || false,
            canDelete: modulePerms?.can_delete || false,
            canApprove: modulePerms?.can_approve || false,
            canExport: modulePerms?.can_export || false,
            canAssign: modulePerms?.can_assign || false,
            canConfigure: modulePerms?.can_configure || false,
          };
        })
        .filter((perm) => {
          // Filter out modules where all permissions are false
          // API requires at least one permission to be true
          return (
            perm.canView ||
            perm.canCreate ||
            perm.canEdit ||
            perm.canDelete ||
            perm.canApprove ||
            perm.canExport ||
            perm.canAssign ||
            perm.canConfigure
          );
        });

      console.log("💾 Saving permissions for role:", selectedRole);
      console.log(
        "📋 Permissions to save:",
        JSON.stringify(permissions, null, 2)
      );

      if (permissions.length === 0) {
        throw new Error("At least one module must have permissions enabled");
      }

      const response = await bulkUpdateRolePermissions({
        roleId: selectedRole,
        permissions,
      });

      if (!response.success) {
        throw new Error(response.message);
      }

      return response;
    },
    onSuccess: (response) => {
      console.log("✅ Permissions saved successfully:", response);

      // Show detailed success message
      const { successCount, failureCount } = response.data || {};
      if (failureCount && failureCount > 0) {
        toast.warning(
          `Updated ${successCount || 0} permissions successfully, ${failureCount} failed`
        );
      } else {
        toast.success(
          `Successfully updated permissions for ${successCount || modules.length} module(s)`
        );
      }

      // Reset the hasChanges flag immediately to show changes are saved
      setHasChanges(false);

      // Refetch permissions from server to confirm the changes persisted
      // This ensures the UI reflects what's actually in the database
      queryClient.invalidateQueries({
        queryKey: [QUERY_KEYS.ROLE_PERMISSIONS, selectedRole],
      });
    },
    onError: (error: Error) => {
      toast.error(error.message || "Failed to update permissions");
    },
  });

  const handleSave = () => {
    savePermissionsMutation.mutate();
  };

  const handleConfirmRoleSwitch = () => {
    if (pendingRoleId) {
      setSelectedRole(pendingRoleId);
      setHasChanges(false);
      setPendingRoleId(null);
    }
    setConfirmSwitchRole(false);
  };

  const isLoading = false || modulesLoading;
  const isSaving = savePermissionsMutation.isPending;

  if (!departmentId) {
    return (
      <div className="grid place-items-center">
        <Empty>
          <EmptyHeader>
            <EmptyMedia variant="icon">
              <InfoIcon className="h-6 w-6" />
            </EmptyMedia>
            <EmptyTitle>No Department Selected</EmptyTitle>
            <EmptyDescription>
              Please select a department to view roles and permissions.
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      </div>
    );
  }

  if (isLoading) {
    return (
      <Card className="p-4">
        <div className="mb-4 flex items-center justify-between">
          <div className="space-y-2">
            <Skeleton className="h-7 w-48" />
            <Skeleton className="h-4 w-96" />
          </div>
          <Skeleton className="h-9 w-32" />
        </div>

        <div className="space-y-6">
          {/* Roles Overview Skeleton */}
          <Card>
            <CardHeader>
              <Skeleton className="mb-2 h-6 w-56" />
              <Skeleton className="h-4 w-72" />
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 gap-3 md:grid-cols-2 lg:grid-cols-3">
                {[...Array(6)].map((_, i) => (
                  <div key={i} className="rounded-md border p-4">
                    <div className="mb-2 flex items-center gap-2">
                      <Skeleton className="h-5 w-5 rounded-full" />
                      <Skeleton className="h-5 w-32" />
                    </div>
                    <Skeleton className="h-4 w-full" />
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>

          {/* Permissions Matrix Skeleton */}
          <Card>
            <CardHeader>
              <Skeleton className="mb-2 h-6 w-64" />
              <Skeleton className="h-4 w-80" />
            </CardHeader>
            <CardContent>
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>
                        <Skeleton className="h-4 w-20" />
                      </TableHead>
                      {[...Array(8)].map((_, i) => (
                        <TableHead key={i} className="text-center">
                          <Skeleton className="mx-auto h-4 w-16" />
                        </TableHead>
                      ))}
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {[...Array(5)].map((_, rowIndex) => (
                      <TableRow key={rowIndex}>
                        <TableCell>
                          <Skeleton className="h-4 w-32" />
                        </TableCell>
                        {[...Array(8)].map((_, colIndex) => (
                          <TableCell key={colIndex} className="text-center">
                            <Skeleton className="mx-auto h-6 w-10 rounded-full" />
                          </TableCell>
                        ))}
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </CardContent>
          </Card>

          {/* Save Button Area Skeleton */}
          <div className="rounded-md border p-4">
            <div className="flex items-center justify-between">
              <div className="space-y-2">
                <Skeleton className="h-5 w-48" />
                <Skeleton className="h-4 w-64" />
              </div>
              <Skeleton className="h-9 w-36" />
            </div>
          </div>
        </div>
      </Card>
    );
  }

  if (!roles?.length || roles?.length === 0) {
    return (
      <>
        <div className="col-span-full rounded-lg border border-dashed">
          <Empty>
            <EmptyHeader>
              <EmptyMedia variant="icon">
                <ShieldIcon className="h-6 w-6" />
              </EmptyMedia>
              <EmptyTitle>No User Roles</EmptyTitle>
              <EmptyDescription>
                No roles found for this department.
              </EmptyDescription>
            </EmptyHeader>
            <EmptyContent>
              <div className="flex gap-2">
                <Button
                  size="sm"
                  onClick={() => {
                    setEditingRole(null);
                    setOpenRoleModal(true);
                  }}
                >
                  <Plus className="mr-2 h-4 w-4" />
                  Create First Role
                </Button>
              </div>
            </EmptyContent>
          </Empty>
        </div>
        <CreateOrUpdateRoleDialog
          openModal={openRoleModal}
          setOpenModal={setOpenRoleModal}
          departmentId={departmentId}
          initialData={editingRole}
          setInitialData={setEditingRole}
        />
      </>
    );
  }

  if (modules?.length === 0) {
    return (
      <div className="col-span-full rounded-lg border border-dashed">
        <Empty>
          <EmptyHeader>
            <EmptyMedia variant="icon">
              <InfoIcon className="h-6 w-6" />
            </EmptyMedia>
            <EmptyTitle>No Modules Assigned</EmptyTitle>
            <EmptyDescription>
              No modules have been assigned to this department. Please assign
              modules first.
            </EmptyDescription>
          </EmptyHeader>
        </Empty>
      </div>
    );
  }

  return (
    <>
      <Card className="p-4">
        <div className="mb-4 flex items-center justify-between">
          <div className="space-y-1">
            <h3 className="text-lg font-semibold">Roles & Permissions</h3>
            <p className="text-muted-foreground text-sm">
              List of all the roles and permissions in this department
            </p>
          </div>
          <Button
            size="sm"
            onClick={() => {
              setEditingRole(null);
              setOpenRoleModal(true);
            }}
          >
            <Plus className="mr-2 h-4 w-4" />
            Add New Role
          </Button>
        </div>

        <div className="space-y-6">
          {isSaving && (
            <div className="flex items-center gap-2 rounded-md border border-blue-200 bg-blue-50 p-3">
              <Spinner className="h-4 w-4 text-blue-600" />
              <span className="text-sm text-blue-600">
                Saving permissions...
              </span>
            </div>
          )}

          {/* Roles Overview */}
          <Card>
            <CardHeader>
              <CardTitle>Roles in This Department</CardTitle>
              <CardDescription>
                Select a role to configure its permissions for assigned modules
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-1 gap-3 md:grid-cols-2 lg:grid-cols-3">
                {roles.map((role) => (
                  <div
                    key={role.id}
                    onClick={() => {
                      if (hasChanges) {
                        setPendingRoleId(role.id);
                        setConfirmSwitchRole(true);
                        return;
                      }
                      setSelectedRole(role.id);
                      setHasChanges(false);
                    }}
                    className={cn(
                      "hover:bg-accent group relative cursor-pointer rounded-md border p-4 text-left transition-colors",
                      selectedRole === role.id ? "border-primary bg-accent" : ""
                    )}
                  >
                    <Button
                      variant="ghost"
                      size="icon"
                      className="absolute top-2 right-2 h-7 w-7 opacity-0 transition-opacity group-hover:opacity-100"
                      onClick={(e) => {
                        e.stopPropagation();
                        setEditingRole(role);
                        setOpenRoleModal(true);
                      }}
                    >
                      <Edit className="h-4 w-4" />
                    </Button>
                    <div
                      className="mb-2 flex items-center gap-2"
                      title={role.is_active ? "Active" : "Inactive"}
                    >
                      <ShieldIcon className="h-5 w-5" />
                      <h3 className="font-medium">{role.name}</h3>
                    </div>
                    <p className="text-muted-foreground text-sm">
                      {role.description || `Code: ${role.code}`}
                    </p>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>

          {/* Permissions Matrix */}
          {selectedRole && (
            <Card>
              <CardHeader>
                <CardTitle>
                  Permissions for{" "}
                  {roles.find((r) => r.id === selectedRole)?.name}
                </CardTitle>
                <CardDescription>
                  Configure which modules and actions this role can access
                </CardDescription>
              </CardHeader>
              <CardContent>
                {permissionsLoading ? (
                  <div className="flex items-center justify-center py-8">
                    <Spinner className="h-6 w-6" />
                    <span className="text-muted-foreground ml-2">
                      Loading permissions...
                    </span>
                  </div>
                ) : (
                  <div className="overflow-x-auto">
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead>Module</TableHead>
                          {Object.entries(PERMISSION_LABELS).map(
                            ([key, label]) => (
                              <TableHead key={key} className="text-center">
                                <TooltipProvider>
                                  <Tooltip>
                                    <TooltipTrigger>
                                      <div className="flex items-center justify-center gap-1">
                                        {label}
                                        <InfoIcon className="text-muted-foreground h-3 w-3" />
                                      </div>
                                    </TooltipTrigger>
                                    <TooltipContent>
                                      <p>{key}</p>
                                    </TooltipContent>
                                  </Tooltip>
                                </TooltipProvider>
                              </TableHead>
                            )
                          )}
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {modules.map((module) => (
                          <TableRow key={module.id}>
                            <TableCell className="font-medium">
                              {formatModuleName(
                                module.name,
                                module.module_code
                              )}
                            </TableCell>
                            {Object.keys(PERMISSION_LABELS).map((permType) => (
                              <TableCell key={permType} className="text-center">
                                <Switch
                                  checked={
                                    permissionsMatrix[module.id]?.[
                                      permType as PermissionType
                                    ] || false
                                  }
                                  onCheckedChange={() =>
                                    togglePermission(
                                      module.id,
                                      permType as PermissionType
                                    )
                                  }
                                  disabled={isSaving}
                                />
                              </TableCell>
                            ))}
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  </div>
                )}
              </CardContent>
            </Card>
          )}

          {/* Save Button */}
          {selectedRole && (
            <div className="flex items-center justify-between rounded-md border border-amber-200 bg-amber-50 p-4">
              <div>
                <p className="text-sm font-medium text-amber-900">
                  {hasChanges
                    ? "You have unsaved changes"
                    : "All changes saved"}
                </p>
                <p className="text-muted-foreground text-sm">
                  {hasChanges
                    ? "Click 'Save Permissions' to apply your changes"
                    : "Permissions are up to date"}
                </p>
              </div>
              <Button
                onClick={handleSave}
                disabled={!hasChanges || isSaving}
                size="sm"
              >
                {isSaving ? "Saving..." : "Save Permissions"}
              </Button>
            </div>
          )}
        </div>
      </Card>

      <CreateOrUpdateRoleDialog
        openModal={openRoleModal}
        setOpenModal={setOpenRoleModal}
        departmentId={departmentId}
        initialData={editingRole}
        setInitialData={setEditingRole}
      />

      <ConfirmationModal
        open={confirmSwitchRole}
        onOpenChange={setConfirmSwitchRole}
        onConfirm={handleConfirmRoleSwitch}
        type="close"
        title="Unsaved Changes"
        description="You have unsaved changes. Are you sure you want to switch roles? All unsaved changes will be lost."
        confirmText="Switch Role"
        cancelText="Cancel"
      />
    </>
  );
}

const ROLE_INITIAL_STATE: Omit<Role, "id" | "department_id"> = {
  name: "",
  code: "",
  description: "",
  is_active: true,
  is_department_head: false,
};

interface CreateOrUpdateRoleDialogProps {
  openModal: boolean;
  setOpenModal: React.Dispatch<React.SetStateAction<boolean>>;
  departmentId: string;
  initialData: Role | null;
  setInitialData: React.Dispatch<React.SetStateAction<Role | null>>;
}

function CreateOrUpdateRoleDialog({
  openModal,
  setOpenModal,
  departmentId,
  initialData,
  setInitialData,
}: CreateOrUpdateRoleDialogProps) {
  const queryClient = useQueryClient();
  const [formData, setFormData] = useState(initialData || ROLE_INITIAL_STATE);
  const [error, setError] = useState<{ status: boolean; message: string }>({
    status: false,
    message: "",
  });

  useEffect(() => {
    if (initialData) {
      setFormData(initialData);
    } else {
      setFormData(ROLE_INITIAL_STATE);
    }
  }, [initialData, openModal]);

  const saveMutation = useMutation({
    mutationFn: (data: any) => {
      // const payload = {
      //   name: data.name,
      //   code: data.code,
      //   description: data.description,
      //   departmentId: departmentId,
      //   isActive: data.is_active
      // };
      return initialData
        ? updateRole({
            ...data,
            id: initialData.id,
            department_id: departmentId,
          })
        : createRole({ ...data, department_id: departmentId }, "", []);
    },
    onSuccess: (response) => {
      if (response.success) {
        toast.success(
          `Role ${initialData ? "updated" : "created"} successfully`
        );
        // Invalidate roles query with matching params object
        queryClient.invalidateQueries({
          queryKey: [QUERY_KEYS.ROLES, { departmentId }],
        });
        setOpenModal(false);
      } else {
        toast.error(response.message);
        setError({
          status: true,
          message: response.message || "An error occurred",
        });
      }
    },
    onError: (err: Error) => {
      toast.error(err.message || "An error occurred");
      setError({ status: true, message: err.message });
    },
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    saveMutation.mutate(formData);
  };

  const handleOpenChange = useCallback(
    (open: boolean) => {
      setOpenModal(open);
      if (!open) {
        setInitialData(null);
        setFormData(ROLE_INITIAL_STATE);
        setError({ status: false, message: "" });
      }
    },
    [setOpenModal, setInitialData]
  );

  return (
    <Dialog open={openModal} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>
            {initialData ? "Update Role" : "Create New Role"}
          </DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <Input
            label="Role Name"
            placeholder="e.g., Auditor, Manager"
            value={formData.name}
            onChange={(e) =>
              setFormData((p) => ({ ...p, name: e.target.value }))
            }
            required
          />
          <Input
            label="Role Code"
            placeholder="e.g., AUD, MGR"
            value={formData.code}
            onChange={(e) =>
              setFormData((p) => ({ ...p, code: e.target.value.toUpperCase() }))
            }
            required
          />
          <Textarea
            label="Description"
            placeholder="A short description of the role (optional)"
            value={formData.description}
            onChange={(e) =>
              setFormData((p) => ({ ...p, description: e.target.value }))
            }
          />
          <div className="flex items-center space-x-2 rounded-lg border bg-slate-50/5 p-4 py-2 transition-colors hover:bg-slate-50">
            <Checkbox
              // type="checkbox"
              id="is_department_head"
              checked={Boolean(formData.is_department_head || false)}
              onCheckedChange={(checked) =>
                setFormData((p) => ({
                  ...p,
                  is_department_head: Boolean(checked || false),
                }))
              }
              className="text-primary focus:ring-primary h-4 w-4 cursor-pointer rounded border-gray-300 focus:ring-2 focus:ring-offset-2"
            />
            <Label
              htmlFor="is_department_head"
              className="flex w-full flex-1 cursor-pointer flex-col items-start gap-0 text-sm font-medium select-none"
            >
              Department Head
              <span className="text-muted-foreground block text-xs font-normal">
                One department can only have one department head.
              </span>
            </Label>
          </div>

          <div className="flex justify-end gap-3 pt-2">
            <DialogClose asChild>
              <Button type="button" variant="destructive">
                Cancel
              </Button>
            </DialogClose>
            <Button
              type="submit"
              disabled={
                saveMutation.isPending || !formData.name || !formData.code
              }
              isLoading={saveMutation.isPending}
            >
              {initialData ? "Update Role" : "Create Role"}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
