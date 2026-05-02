"use client";

import { useState, useEffect } from "react";
import { useSearchParams, useRouter, usePathname } from "next/navigation";
import { PageHeader } from "@/components/base/page-header";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { TasksTable } from "./tasks-table";
import { TaskStatsCards } from "./task-stats-cards";
import { ApprovalsList } from "./approvals-list";

interface TasksClientProps {
  userId: string;
  userRole: string;
}

type TabValue = "tasks" | "approvals";

export function TasksClient({ userId, userRole }: TasksClientProps) {
  const searchParams = useSearchParams();
  const router = useRouter();
  const pathname = usePathname();
  const [refreshTrigger] = useState(0);
  const [activeTab, setActiveTab] = useState<TabValue>("tasks");

  useEffect(() => {
    const tabParam = searchParams.get("tab");
    if (tabParam === "approvals" || tabParam === "tasks") {
      setActiveTab(tabParam);
    }
  }, [searchParams]);

  const handleTabChange = (value: string) => {
    const v = value as TabValue;
    setActiveTab(v);
    // Keep URL in sync so refreshes/back-button preserve tab state.
    const params = new URLSearchParams(searchParams.toString());
    params.set("tab", v);
    router.replace(`${pathname}?${params.toString()}`, { scroll: false });
  };

  return (
    <div className="space-y-5">
      <PageHeader
        title="Workflows"
        subtitle="Tasks and approvals assigned to you"
        showBackButton={false}
      />

      <Tabs value={activeTab} onValueChange={handleTabChange} className="space-y-5">
        <TabsList className="inline-flex h-9 w-full sm:w-auto bg-muted/60 p-1 rounded-lg">
          <TabsTrigger
            value="tasks"
            className="flex-1 sm:flex-initial sm:px-6 data-[state=active]:bg-background data-[state=active]:text-foreground data-[state=active]:shadow-sm rounded-md"
          >
            Tasks
          </TabsTrigger>
          <TabsTrigger
            value="approvals"
            className="flex-1 sm:flex-initial sm:px-6 data-[state=active]:bg-background data-[state=active]:text-foreground data-[state=active]:shadow-sm rounded-md"
          >
            Approvals
          </TabsTrigger>
        </TabsList>

        <TabsContent value="tasks" className="space-y-4 mt-0">
          <TaskStatsCards userId={userId} refreshTrigger={refreshTrigger} />
          <TasksTable />
        </TabsContent>

        <TabsContent value="approvals" className="space-y-4 mt-0">
          <ApprovalsList userId={userId} userRole={userRole} />
        </TabsContent>
      </Tabs>
    </div>
  );
}
