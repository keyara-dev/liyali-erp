"use client";

import { useState, useEffect, useTransition } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from "@/components/ui/table";
import { Plus, Trash2, Building, Pencil, View } from "lucide-react";
import { toast } from "sonner";
import { ConfirmationModal } from "@/components/confirmation-modal";
import { Department } from "@/lib/mock-departments";
import { deleteDepartment } from "@/app/_actions/config-actions";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { QUERY_KEYS } from "@/lib/constants";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle
} from "@/components/ui/empty";
import { CustomPagination } from "@/components/ui/custom-pagination";
import { CreateOrUpdateDepartment } from "./department-users";
import Link from "next/link";

type Pagination = {
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
};

type DepartmentsConfigProps = {
  initialDepartments: Department[];
  pagination: Pagination;
};

export default function DepartmentsConfig({
  initialDepartments,
  pagination
}: DepartmentsConfigProps) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [isPending, startTransition] = useTransition();
  const queryClient = useQueryClient();

  const [openModal, setOpenModal] = useState(false);
  const [editingDepartment, setEditingDepartment] = useState<Department | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [departmentToDelete, setDepartmentToDelete] = useState<string | null>(null);
  const [departments, setDepartments] = useState<Department[]>(initialDepartments);

  useEffect(() => {
    setDepartments(initialDepartments);
  }, [initialDepartments]);

  // Pagination handler
  const updatePagination = ({ page, page_size }: { page?: number; page_size?: number }) => {
    const params = new URLSearchParams(searchParams.toString());

    if (page !== undefined) {
      params.set("page", String(page));
    }

    if (page_size !== undefined) {
      params.set("page_size", String(page_size));
      params.set("page", "1");
    }

    startTransition(() => {
      router.push(`?${params.toString()}`);
    });
  };

  // Delete department mutation
  const deleteMutation = useMutation({
    mutationFn: (id: string) => deleteDepartment(id),
    onSuccess: (response) => {
      if (response.success) {
        toast.success("Department deleted successfully");
        queryClient.invalidateQueries({ queryKey: [QUERY_KEYS.DEPARTMENTS] });
        router.refresh();
      } else {
        toast.error(response.message || "Failed to delete department");
      }
    },
    onError: (error) => {
      toast.error("Failed to delete department");
      console.error("Error deleting department:", error);
    },
    onSettled: () => {
      setDeleteDialogOpen(false);
      setDepartmentToDelete(null);
    }
  });

  const handleDeleteClick = (id: string) => {
    setDepartmentToDelete(id);
    setDeleteDialogOpen(true);
  };

  const handleDeleteConfirm = async () => {
    if (!departmentToDelete) return;
    deleteMutation.mutate(departmentToDelete);
  };

  // Transform pagination for CustomPagination
  const customPaginationData = {
    page: pagination.page,
    page_size: pagination.page_size,
    total_pages: pagination.total_pages,
    totalCount: pagination.total,
    has_prev: pagination.has_prev,
    has_next: pagination.has_next
  };

  return (
    <>
      <Card className="p-4">
        <div className="mb-4 flex items-center justify-between">
          <h3 className="text-lg font-semibold">Departments</h3>
          <Button
            size="sm"
            onClick={() => {
              setEditingDepartment(null);
              setOpenModal(true);
            }}>
            <Plus className="h-4 w-4" />
            New Department
          </Button>
        </div>

        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Code</TableHead>
              <TableHead className="w-24" align="center">
                Actions
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {departments.length === 0 ? (
              <TableRow>
                <TableCell colSpan={4} align="center">
                  <Empty>
                    <EmptyHeader>
                      <EmptyMedia variant="icon">
                        <Building />
                      </EmptyMedia>
                      <EmptyTitle>No Departments Yet</EmptyTitle>
                      <EmptyDescription>
                        You haven&apos;t created any departments yet. Get started by creating your
                        first department.
                      </EmptyDescription>
                    </EmptyHeader>
                    <EmptyContent>
                      <div className="flex gap-2">
                        <Button
                          size="sm"
                          onClick={() => {
                            setEditingDepartment(null);
                            setOpenModal(true);
                          }}>
                          <Plus className="h-4 w-4" /> Create New Department
                        </Button>
                      </div>
                    </EmptyContent>
                  </Empty>
                </TableCell>
              </TableRow>
            ) : (
              departments.map((department) => (
                <TableRow
                  key={department.id}
                  className="cursor-pointer"
                  onClick={() => {
                    router.push(`/dashboard/system-configs/departments/${department.id}`);
                  }}>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <Building className="text-muted-foreground h-4 w-4" />
                      <span className="font-medium">{department.name}</span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <span className="font-mono text-sm">
                      {department.description || "No description provided"}
                    </span>
                  </TableCell>
                  <TableCell>
                    <span className="font-mono text-sm">{department.code}</span>
                  </TableCell>
                  <TableCell align="center">
                    <div className="flex justify-end gap-2">
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={(e) => e.stopPropagation()}
                        className="h-8 gap-1.5">
                        <Link
                          href={`/dashboard/system-configs/departments/${department.id}`}
                          className="flex cursor-pointer items-center gap-2">
                          <View className="h-3.5 w-3.5" />
                          View
                        </Link>
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={(e) => {
                          setEditingDepartment(department);
                          setOpenModal(true);
                          e.stopPropagation();
                        }}
                        className="h-8 gap-1.5">
                        <Pencil className="h-3.5 w-3.5" />
                        Edit
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={(e) => {
                          handleDeleteClick(String(department.id));
                          e.stopPropagation();
                        }}
                        className="text-destructive hover:text-destructive hover:bg-destructive/10 h-8 gap-1.5">
                        <Trash2 className="h-4 w-4" /> Delete
                      </Button>
                    </div>
                    <div className="flex gap-4"></div>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>

        {/* CustomPagination */}
        {departments.length > 0 && (
          <CustomPagination
            pagination={customPaginationData}
            updatePagination={updatePagination}
            allowSetPageSize={true}
            showDetails={true}
            className="mt-4 border-t"
          />
        )}
      </Card>

      <CreateOrUpdateDepartment
        openModal={openModal}
        setOpenModal={setOpenModal}
        initialData={editingDepartment}
        departmentId={editingDepartment?.id}
        setInitialData={setEditingDepartment}
      />

      <ConfirmationModal
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
        title="Delete Department"
        description="Are you sure you want to delete this department? This action cannot be undone and may affect related data."
        onConfirm={handleDeleteConfirm}
        isLoading={deleteMutation.isPending}
      />
    </>
  );
}
