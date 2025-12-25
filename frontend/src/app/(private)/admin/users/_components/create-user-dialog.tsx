"use client";

import { PencilLine, Plus, Check, Copy, UserCog } from "lucide-react";
import { Dispatch, SetStateAction, useEffect, useMemo, useState } from "react";
import { toast } from "sonner";
import { useRouter } from "next/navigation";

import { Input } from "@/components/ui/input";
import { SelectField } from "@/components/ui/select-field";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { DialogClose } from "@radix-ui/react-dialog";

import { User, UserType } from "@/types";
import { generateRandomString } from "@/lib/utils";
import { useCreateUser, useUpdateUser } from "@/hooks/use-users-query";
import { getAllDepartments } from "@/lib/mock-departments";

type FormData = {
  username?: string | number | readonly string[] | undefined;
  first_name: string;
  last_name: string;
  email: string;
  phone?: string;
  role: string;
  department_id?: string;
  department?: string;
  is_active: boolean;
  role: UserType;
  password?: string;
};

type Role = {
  id: string;
  name: string;
  code: string;
  department_id: string;
};

export default function CreateUserForm({
  role,
  user,
  showTrigger,
  isOpenModal,
  setIsOpenModal,
}: {
  showTrigger?: boolean;
  role: UserType;
  user: User | null;
  isOpenModal?: boolean;
  setIsOpenModal?: Dispatch<SetStateAction<boolean>>;
}) {
  const router = useRouter();
  const [copied, setCopied] = useState(false);
  const [internalOpen, setInternalOpen] = useState<boolean>(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const isEditMode = !!user;

  // Use internal state for trigger mode, external state for controlled mode
  const dialogOpen = showTrigger ? internalOpen : isOpenModal;
  const setDialogOpen = showTrigger ? setInternalOpen : setIsOpenModal;

  // TanStack Query mutations
  const createUserMutation = useCreateUser();
  const updateUserMutation = useUpdateUser();

  // Initialize form state
  const initialFormState: FormData = useMemo(() => {
    if (isEditMode && user) {
      return {
        first_name: user.first_name || "",
        last_name: user.last_name || "",
        email: user.email || "",
        phone: (user as any).phone || "",
        role: user.role || "",
        department_id: user.department_id || "",
        is_active: user.is_active ?? true,
        role,
        password: "",
      };
    }
    return {
      first_name: "",
      last_name: "",
      email: "",
      phone: "",
      role: "",
      department_id: "",
      is_active: true,
      role,
      password: generateRandomString(),
    };
  }, [isEditMode, user?.id, role]);

  const [formData, setFormData] = useState<FormData>(
    initialFormState as FormData
  );

  // Get departments from mock data
  const departments = useMemo(() => {
    return getAllDepartments().filter((d) => d.is_active);
  }, []);

  // Reset form when user changes or dialog opens/closes
  useEffect(() => {
    if (isEditMode && user) {
      setFormData({
        first_name: user.first_name || "",
        last_name: user.last_name || "",
        email: user.email || "",
        phone: (user as any).phone || "",
        role: user.role || "",
        department: user.department || "",
        department_id: user.department_id || "",
        is_active: user.is_active ?? true,
        role,
      });
    } else if (!isEditMode && dialogOpen) {
      setFormData({
        first_name: "",
        last_name: "",
        email: "",
        phone: "",
        role: "",
        department: "",
        department_id: "",
        is_active: true,
        role,
      });
    }
  }, [user?.id, dialogOpen, isEditMode, role]);

  const handleCopyPassword = async () => {
    try {
      await navigator.clipboard.writeText(formData.password || "");
      setCopied(true);
      toast.info("Password copied to clipboard.");
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      toast.error("Failed to copy password");
    }
  };

  const handleGenerateNewPassword = () => {
    setFormData((prev) => ({
      ...prev,
      password: generateRandomString(),
    }));
    setCopied(false);
  };

  const resetForm = () => {
    setFormData(initialFormState);
    setCopied(false);
  };

  const handleCloseModal = () => {
    resetForm();
    setDialogOpen?.(false);
  };

  const validateForm = (): boolean => {
    if (!formData.first_name.trim()) {
      toast.error("First name is required");
      return false;
    }
    if (!formData.last_name.trim()) {
      toast.error("Last name is required");
      return false;
    }

    if (!formData.email.trim()) {
      toast.error("Email is required");
      return false;
    }
    if (!isEditMode && !formData.password?.trim()) {
      toast.error("Password is required");
      return false;
    }
    return true;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    setIsSubmitting(true);

    try {
      let response;
      if (isEditMode) {
        const updateData = {
          email: formData.email,
          phone: formData.phone,
          first_name: formData.first_name,
          last_name: formData.last_name,
          department_id: formData.department_id,
          is_active: formData.is_active,
          role,
        };
        response = await updateUserMutation.mutateAsync({
          userId: user!.id,
          data: updateData,
        });
      } else {
        response = await createUserMutation.mutateAsync({
          email: formData.email,
          phone: formData.phone || "",
          password: formData.password || generateRandomString(12),
          first_name: formData.first_name,
          last_name: formData.last_name,
          department_id: formData.department_id || "",
          role,
          username: String(formData.username || ""),
          branch_id: "",
          role_id: "",
        });
      }

      if (response.success) {
        toast.success(
          `User ${isEditMode ? "updated" : "created"} successfully`
        );
        handleCloseModal();
        router.refresh();
      } else {
        toast.error(
          response.message ||
            `Failed to ${isEditMode ? "update" : "create"} user`
        );
      }
    } catch (error) {
      console.error("Form submission error:", error);
      toast.error("An unexpected error occurred");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog
      open={dialogOpen}
      onOpenChange={(open) => {
        if (showTrigger) {
          if (open) {
            setInternalOpen(true);
          } else {
            handleCloseModal();
          }
        } else {
          if (!open) {
            handleCloseModal();
          }
        }
      }}
    >
      {showTrigger && (
        <DialogTrigger asChild>
          <Button size="sm">
            {user ? (
              <>
                <PencilLine className="mr-2 h-4 w-4" /> Update User
              </>
            ) : (
              <>
                <Plus className="mr-2 h-4 w-4" />
                Create New User
              </>
            )}
          </Button>
        </DialogTrigger>
      )}

      <DialogContent className="max-h-[90vh] w-full overflow-hidden p-0">
        <DialogHeader className="border-b px-6 py-4">
          <div className="flex items-center gap-3">
            <div className="bg-primary/5 text-primary hover:bg-primary/10 flex h-7 w-7 items-center justify-center rounded-full">
              <UserCog className="h-4 w-4" />
            </div>
            <DialogTitle>
              {isEditMode ? "Edit User" : "Create New User"}
            </DialogTitle>
          </div>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="overflow-y-auto px-6 py-6">
            <div>
              <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                <div>
                  <Label htmlFor="first_name">
                    First Name <span className="text-destructive">*</span>
                  </Label>
                  <Input
                    id="first_name"
                    placeholder="Bob"
                    value={formData.first_name}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        first_name: e.target.value,
                      }))
                    }
                    disabled={isSubmitting}
                    required
                    className="mt-1"
                  />
                </div>

                <div>
                  <Label htmlFor="last_name">
                    Last Name <span className="text-destructive">*</span>
                  </Label>
                  <Input
                    id="last_name"
                    placeholder="Mwale"
                    value={formData.last_name}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        last_name: e.target.value,
                      }))
                    }
                    disabled={isSubmitting}
                    required
                    className="mt-1"
                  />
                </div>

                <div>
                  <Label htmlFor="username">
                    Username <span className="text-destructive">*</span>
                  </Label>
                  <Input
                    id="username"
                    placeholder="bmwale"
                    value={formData.username}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        username: e.target.value,
                      }))
                    }
                    disabled={isSubmitting}
                    required
                    className="mt-1"
                  />
                </div>

                <div>
                  <Label htmlFor="email">
                    Email Address <span className="text-destructive">*</span>
                  </Label>
                  <Input
                    id="email"
                    type="email"
                    placeholder="mail@company.com"
                    value={formData.email}
                    onChange={(e) =>
                      setFormData((prev) => ({
                        ...prev,
                        email: e.target.value,
                      }))
                    }
                    disabled={isSubmitting}
                    required
                    className="mt-1"
                  />
                </div>
              </div>
            </div>

            <div className="mt-4">
              <div className="grid grid-cols-1 gap-4">
                <div>
                  <Label htmlFor="department_id">
                    Department <span className="text-destructive">*</span>
                  </Label>
                  <SelectField
                    value={formData.department_id}
                    onValueChange={(value) =>
                      setFormData((prev) => ({
                        ...prev,
                        department_id: value,
                      }))
                    }
                    isDisabled={isSubmitting}
                    placeholder="Select department"
                    options={[
                      { id: "", name: "Select department" },
                      ...departments.map((dept) => ({
                        id: dept.id,
                        name: dept.name,
                      })),
                    ]}
                    classNames={{
                      wrapper: "w-full",
                      input: "!h-10",
                    }}
                  />
                </div>

                <div>
                  <Label htmlFor="role">
                    Role <span className="text-destructive">*</span>
                  </Label>
                  <SelectField
                    value={formData.role}
                    onValueChange={(value) =>
                      setFormData((prev) => ({
                        ...prev,
                        role: value,
                      }))
                    }
                    isDisabled={isSubmitting}
                    placeholder="Select role"
                    options={[
                      { id: "", name: "Select role" },
                      { id: "role-1", name: "Administrator" },
                      { id: "role-2", name: "Manager" },
                      { id: "role-3", name: "Viewer" },
                    ]}
                    classNames={{
                      wrapper: "w-full",
                      input: "!h-10",
                    }}
                  />
                </div>

                {isEditMode && (
                  <div className="flex flex-row items-center justify-between rounded-lg border p-4">
                    <div className="space-y-0.5">
                      <Label className="text-base">Account Status</Label>
                      <p className="text-sm text-muted-foreground">
                        {formData.is_active
                          ? "Account is active"
                          : "Account is deactivated"}
                      </p>
                    </div>
                    <Switch
                      checked={formData.is_active}
                      onCheckedChange={(checked) =>
                        setFormData((prev) => ({
                          ...prev,
                          is_active: checked,
                        }))
                      }
                      disabled={isSubmitting}
                    />
                  </div>
                )}

                {!isEditMode && (
                  <div>
                    <Label htmlFor="password">
                      Password <span className="text-destructive">*</span>
                    </Label>
                    <div className="mt-1 flex w-full flex-col items-center gap-2 sm:flex-row">
                      <div className="relative flex w-full items-center gap-2">
                        <Input
                          id="password"
                          value={formData.password}
                          readOnly
                          className="cursor-default font-mono text-sm"
                          disabled={isSubmitting}
                        />
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          onClick={handleCopyPassword}
                          className="hover:bg-muted/5 absolute right-1 shrink-0"
                          disabled={isSubmitting}
                        >
                          {copied ? (
                            <Check className="h-4 w-4 text-green-600" />
                          ) : (
                            <Copy className="h-4 w-4" />
                          )}
                        </Button>
                      </div>
                      <Button
                        type="button"
                        onClick={handleGenerateNewPassword}
                        disabled={isSubmitting}
                      >
                        Generate new password
                      </Button>
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>

          <DialogFooter className="flex justify-end gap-3 border-t p-4">
            <div className="flex w-full items-center justify-end gap-3">
              <DialogClose asChild>
                <Button
                  type="button"
                  variant="outline"
                  onClick={handleCloseModal}
                  disabled={isSubmitting}
                >
                  Cancel
                </Button>
              </DialogClose>
              <Button
                type="submit"
                disabled={isSubmitting}
                isLoading={isSubmitting}
                loadingText={isEditMode ? "Updating..." : "Creating..."}
              >
                {isEditMode ? "Update User" : "Create User"}
              </Button>
            </div>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

// Convenience wrapper component for just showing the button trigger
export function CreateUserButton({ role }: { role: UserType }) {
  return <CreateUserForm showTrigger={true} role={role} user={null} />;
}
