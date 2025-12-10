"use client";

import { useState, useEffect, useCallback } from "react";
import { Plus, Edit, Trash2, RotateCcw, AlertCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { ConfirmationModal } from "@/components/confirmation-modal";
import { toast } from "sonner";
import {
  Department,
  getAllDepartments,
  createDepartment,
  updateDepartment,
  deleteDepartment,
  restoreDepartment,
} from "@/lib/mock-departments";

const INITIAL_FORM_STATE: Omit<Department, "id" | "created_at" | "updated_at"> = {
  name: "",
  code: "",
  description: "",
  manager_name: "",
  is_active: true,
};

export default function DepartmentsMockConfig() {
  const [departments, setDepartments] = useState<Department[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [openModal, setOpenModal] = useState(false);
  const [editingDept, setEditingDept] = useState<Department | null>(null);
  const [formData, setFormData] = useState(INITIAL_FORM_STATE);
  const [error, setError] = useState<string>("");
  const [deleteConfirm, setDeleteConfirm] = useState<{ open: boolean; deptId: string | null }>({
    open: false,
    deptId: null,
  });
  const [restoreConfirm, setRestoreConfirm] = useState<{ open: boolean; deptId: string | null }>({
    open: false,
    deptId: null,
  });

  useEffect(() => {
    const loadDepartments = () => {
      try {
        const data = getAllDepartments();
        setDepartments(data);
      } catch (err) {
        console.error("Failed to load departments:", err);
        toast.error("Failed to load departments");
      } finally {
        setIsLoading(false);
      }
    };

    loadDepartments();
  }, []);

  const resetForm = useCallback(() => {
    setFormData(INITIAL_FORM_STATE);
    setError("");
    setEditingDept(null);
  }, []);

  const handleOpenModal = (dept: Department | null = null) => {
    if (dept) {
      setEditingDept(dept);
      setFormData({
        name: dept.name,
        code: dept.code,
        description: dept.description,
        manager_name: dept.manager_name,
        is_active: dept.is_active,
      });
    } else {
      resetForm();
    }
    setOpenModal(true);
  };

  const handleCloseModal = () => {
    setOpenModal(false);
    resetForm();
  };

  const validateForm = (): boolean => {
    if (!formData.name.trim()) {
      setError("Department name is required");
      return false;
    }
    if (!formData.code.trim()) {
      setError("Department code is required");
      return false;
    }
    if (!formData.manager_name.trim()) {
      setError("Manager name is required");
      return false;
    }

    const isDuplicate = departments.some(
      (dept) => dept.code.toUpperCase() === formData.code.toUpperCase() && dept.id !== editingDept?.id
    );
    if (isDuplicate) {
      setError("A department with this code already exists");
      return false;
    }

    return true;
  };

  const handleSave = () => {
    if (!validateForm()) {
      return;
    }

    try {
      if (editingDept) {
        const updated = updateDepartment(editingDept.id, formData);
        if (updated) {
          setDepartments(getAllDepartments());
          toast.success("Department updated successfully");
        } else {
          toast.error("Failed to update department");
        }
      } else {
        createDepartment(formData);
        setDepartments(getAllDepartments());
        toast.success("Department created successfully");
      }
      handleCloseModal();
    } catch (err) {
      console.error("Error saving department:", err);
      toast.error("Failed to save department");
    }
  };

  const handleDeleteConfirm = () => {
    if (!deleteConfirm.deptId) return;

    try {
      const deleted = deleteDepartment(deleteConfirm.deptId);
      if (deleted) {
        setDepartments(getAllDepartments());
        toast.success("Department deleted successfully");
      } else {
        toast.error("Failed to delete department");
      }
    } catch (err) {
      console.error("Error deleting department:", err);
      toast.error("Failed to delete department");
    } finally {
      setDeleteConfirm({ open: false, deptId: null });
    }
  };

  const handleRestoreConfirm = () => {
    if (!restoreConfirm.deptId) return;

    try {
      const restored = restoreDepartment(restoreConfirm.deptId);
      if (restored) {
        setDepartments(getAllDepartments());
        toast.success("Department restored successfully");
      } else {
        toast.error("Failed to restore department");
      }
    } catch (err) {
      console.error("Error restoring department:", err);
      toast.error("Failed to restore department");
    } finally {
      setRestoreConfirm({ open: false, deptId: null });
    }
  };

  const activeDepts = departments.filter((d) => d.is_active);
  const inactiveDepts = departments.filter((d) => !d.is_active);

  if (isLoading) {
    return (
      <Card className="p-4">
        <div className="flex items-center justify-center py-8">
          <div className="text-center">
            <div className="mx-auto mb-4 h-12 w-12 animate-spin rounded-full border-4 border-gray-200 border-t-blue-600" />
            <p className="text-sm text-gray-500">Loading departments...</p>
          </div>
        </div>
      </Card>
    );
  }

  return (
    <>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle>Departments Management</CardTitle>
            <CardDescription>Create and manage departments in your organization</CardDescription>
          </div>
          <Button onClick={() => handleOpenModal()} size="sm" className="gap-2">
            <Plus className="h-4 w-4" />
            Add Department
          </Button>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="space-y-4">
            <div>
              <h3 className="text-sm font-semibold text-foreground">Active Departments ({activeDepts.length})</h3>
              <p className="text-xs text-muted-foreground">Departments currently in use</p>
            </div>

            {activeDepts.length === 0 ? (
              <div className="rounded-lg border border-dashed border-border bg-muted/30 p-8 text-center">
                <p className="text-sm text-muted-foreground">No active departments yet. Create one to get started.</p>
              </div>
            ) : (
              <div className="overflow-x-auto rounded-lg border">
                <Table>
                  <TableHeader>
                    <TableRow className="bg-muted/50">
                      <TableHead>Name</TableHead>
                      <TableHead>Code</TableHead>
                      <TableHead>Manager</TableHead>
                      <TableHead>Description</TableHead>
                      <TableHead className="text-right">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {activeDepts.map((dept) => (
                      <TableRow key={dept.id}>
                        <TableCell className="font-medium">{dept.name}</TableCell>
                        <TableCell>
                          <Badge variant="outline">{dept.code}</Badge>
                        </TableCell>
                        <TableCell className="text-sm">{dept.manager_name}</TableCell>
                        <TableCell className="max-w-xs truncate text-sm text-muted-foreground">
                          {dept.description || "—"}
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex justify-end gap-2">
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => handleOpenModal(dept)}
                              className="gap-1 text-xs">
                              <Edit className="h-3 w-3" />
                              Edit
                            </Button>
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => setDeleteConfirm({ open: true, deptId: dept.id })}
                              className="gap-1 text-xs text-destructive hover:text-destructive">
                              <Trash2 className="h-3 w-3" />
                              Delete
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            )}
          </div>

          {inactiveDepts.length > 0 && (
            <div className="space-y-4 border-t pt-6">
              <div>
                <h3 className="text-sm font-semibold text-foreground">Inactive Departments ({inactiveDepts.length})</h3>
                <p className="text-xs text-muted-foreground">Deleted departments can be restored</p>
              </div>

              <div className="overflow-x-auto rounded-lg border border-dashed">
                <Table>
                  <TableHeader>
                    <TableRow className="bg-muted/30">
                      <TableHead>Name</TableHead>
                      <TableHead>Code</TableHead>
                      <TableHead>Manager</TableHead>
                      <TableHead>Description</TableHead>
                      <TableHead className="text-right">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {inactiveDepts.map((dept) => (
                      <TableRow key={dept.id} className="opacity-60">
                        <TableCell className="font-medium line-through">{dept.name}</TableCell>
                        <TableCell>
                          <Badge variant="outline" className="line-through">
                            {dept.code}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-sm line-through">{dept.manager_name}</TableCell>
                        <TableCell className="max-w-xs truncate text-sm text-muted-foreground line-through">
                          {dept.description || "—"}
                        </TableCell>
                        <TableCell className="text-right">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => setRestoreConfirm({ open: true, deptId: dept.id })}
                            className="gap-1 text-xs">
                            <RotateCcw className="h-3 w-3" />
                            Restore
                          </Button>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      <Dialog open={openModal} onOpenChange={setOpenModal}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{editingDept ? "Edit Department" : "Create New Department"}</DialogTitle>
          </DialogHeader>

          <form onSubmit={(e) => { e.preventDefault(); handleSave(); }} className="space-y-4">
            {error && (
              <div className="flex items-center gap-2 rounded-md bg-destructive/10 p-3 text-sm text-destructive">
                <AlertCircle className="h-4 w-4" />
                {error}
              </div>
            )}

            <Input
              label="Department Name"
              placeholder="e.g., Operations, HR, Finance"
              value={formData.name}
              onChange={(e) => {
                setFormData((p) => ({ ...p, name: e.target.value }));
                setError("");
              }}
              required
            />

            <Input
              label="Department Code"
              placeholder="e.g., OPS, HR, FIN"
              value={formData.code}
              onChange={(e) => {
                setFormData((p) => ({ ...p, code: e.target.value.toUpperCase() }));
                setError("");
              }}
              required
            />

            <Input
              label="Manager Name"
              placeholder="Full name of department manager"
              value={formData.manager_name}
              onChange={(e) => {
                setFormData((p) => ({ ...p, manager_name: e.target.value }));
                setError("");
              }}
              required
            />

            <Textarea
              label="Description (Optional)"
              placeholder="Brief description of the department's role and responsibilities"
              value={formData.description}
              onChange={(e) => {
                setFormData((p) => ({ ...p, description: e.target.value }));
                setError("");
              }}
              rows={3}
            />

            <div className="flex justify-end gap-3 pt-2">
              <DialogClose asChild>
                <Button type="button" variant="outline">
                  Cancel
                </Button>
              </DialogClose>
              <Button type="submit">{editingDept ? "Update Department" : "Create Department"}</Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>

      <ConfirmationModal
        open={deleteConfirm.open}
        onOpenChange={(open) => setDeleteConfirm({ open, deptId: null })}
        onConfirm={handleDeleteConfirm}
        type="delete"
        title="Delete Department"
        description={`Are you sure you want to delete the "${
          departments.find((d) => d.id === deleteConfirm.deptId)?.name || "Department"
        }" department? This action can be undone later.`}
        confirmText="Delete"
        cancelText="Cancel"
      />

      <ConfirmationModal
        open={restoreConfirm.open}
        onOpenChange={(open) => setRestoreConfirm({ open, deptId: null })}
        onConfirm={handleRestoreConfirm}
        type="default"
        title="Restore Department"
        description={`Are you sure you want to restore the "${
          departments.find((d) => d.id === restoreConfirm.deptId)?.name || "Department"
        }" department?`}
        confirmText="Restore"
        cancelText="Cancel"
      />
    </>
  );
}
