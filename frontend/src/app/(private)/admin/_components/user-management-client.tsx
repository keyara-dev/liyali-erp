"use client";

import { useState, ReactNode } from "react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Users, Building2, Shield, Logs } from "lucide-react";
import DepartmentsConfig from "./departments-config";
import UserRolesConfig from "./user-roles-config";
import { ActivityLogsClient } from "./activity-logs-client";
import AccessDeniedPage from "@/app/(private)/access-denied/page";

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
    "users",
  );

  return (
    <div className="space-y-6">
      <Tabs
        value={activeTab}
        onValueChange={(value) => setActiveTab(value as any)}
        className="w-full"
      >
        <TabsList className="grid w-full grid-cols-4">
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
          <TabsTrigger value="logs" className="gap-2">
            <Logs className="h-4 w-4" />
            <span className="hidden sm:inline">Activity Logs</span>
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
        {/* Activity Logs Tab */}
        <TabsContent value="logs" className="space-y-4">
          {userRole == "admin" ? (
            <ActivityLogsClient userId={userId} userRole={userRole} />
          ) : (
            <AccessDeniedPage />
          )}
        </TabsContent>
      </Tabs>
    </div>
  );
}
