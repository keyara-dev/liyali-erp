"use client";

import { useState } from "react";
import { AnimatePresence } from "framer-motion";
import { useOrganizationContext } from "@/hooks/use-organization";
import { CreateWorkspace } from "./_components/create-workspace";
import { WorkspaceSelector } from "./_components/workpace-selector";

export default function WelcomePage() {
  const { refreshOrganizations } = useOrganizationContext();
  const [showCreateForm, setShowCreateForm] = useState(false);

  const handleCreateSuccess = (organization: any) => {
    // Refresh organizations list to include the new one
    refreshOrganizations();
    // Go back to workspace selection
    setShowCreateForm(false);
  };

  return (
    <AnimatePresence mode="wait">
      {showCreateForm ? (
        <CreateWorkspace
          key="create-form"
          onBack={() => setShowCreateForm(false)}
          onSuccess={handleCreateSuccess}
        />
      ) : (
        <WorkspaceSelector
          key="workspace-selector"
          onCreateWorkspace={() => setShowCreateForm(true)}
          showLogo={true}
          showSignOut={true}
        />
      )}
    </AnimatePresence>
  );
}
