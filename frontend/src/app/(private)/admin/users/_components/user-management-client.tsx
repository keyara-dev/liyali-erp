"use client";

import { useState, ReactNode } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Users, Building2, Shield } from "lucide-react";
import DepartmentsConfig from "./departments-config";
import UserRolesConfig from "./user-roles-config";

interface UserManagementClientProps {
  userId: string;
  userRole: string;
  usersTabContent?: ReactNode;
}

export function UserManagementClient({
  userId,
  userRole,
  usersTabContent,
}: UserManagementClientProps) {
  const [activeTab, setActiveTab] = useState<"users" | "departments" | "roles">(
    "users"
  );

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-2xl font-bold tracking-tight">User Management</h1>
        <p className="text-sm text-muted-foreground">
          Manage users, departments, roles and access permissions across the
          system
        </p>
      </div>

      {/* Tabbed Interface */}
      <Tabs
        value={activeTab}
        onValueChange={(value) => setActiveTab(value as any)}
        className="w-full"
      >
        <TabsList className="grid w-full grid-cols-3">
          <TabsTrigger value="users" className="gap-2">
            <Users className="h-4 w-4" />
            <span className="hidden sm:inline">Users</span>
          </TabsTrigger>
          <TabsTrigger value="departments" className="gap-2">
            <Building2 className="h-4 w-4" />
            <span className="hidden sm:inline">Departments</span>
          </TabsTrigger>
          <TabsTrigger value="roles" className="gap-2">
            <Shield className="h-4 w-4" />
            <span className="hidden sm:inline">Manage Roles</span>
          </TabsTrigger>
        </TabsList>

        {/* Users Tab - Content loaded from SSR page */}
        <TabsContent value="users" className="space-y-4">
          {usersTabContent ? (
            usersTabContent
          ) : (
            <div className="text-sm text-muted-foreground">
              <p>
                User management table is rendered server-side for optimal
                performance with pagination and filtering.
              </p>
            </div>
          )}
        </TabsContent>

        {/* Departments Tab */}
        <TabsContent value="departments" className="space-y-4">
          <DepartmentsConfig />
        </TabsContent>

        {/* Manage Roles Tab */}
        <TabsContent value="roles" className="space-y-4">
          <UserRolesConfig />
        </TabsContent>
      </Tabs>
    </div>
  );
}
