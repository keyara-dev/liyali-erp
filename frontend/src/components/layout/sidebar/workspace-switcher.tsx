"use client";

import { useState } from "react";
import { Check, ChevronsUpDown, Plus, Building2 } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from "@/components/ui/command";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Badge } from "@/components/ui/badge";
import { OrganizationAvatar } from "@/components/ui/organization-avatar";
import { useOrganizationContext } from "@/hooks/use-organization";
import { CreateOrganizationModal } from "@/components/modals/create-organization-modal";
import { useSidebar } from "@/components/ui/sidebar";

export function WorkspaceSwitcher() {
  const [open, setOpen] = useState(false);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const {
    currentOrganization,
    userOrganizations,
    switchWorkspace,
    isLoading,
    refreshOrganizations,
  } = useOrganizationContext();
  const { open: sidebarOpen } = useSidebar();

  const handleSelectWorkspace = async (orgId: string) => {
    if (orgId === currentOrganization?.id) {
      setOpen(false);
      return;
    }

    try {
      await switchWorkspace(orgId);
      setOpen(false);
    } catch (error) {
      console.error("Failed to switch workspace:", error);
    }
  };

  const handleCreateSuccess = (organization: any) => {
    refreshOrganizations();
    // The new organization will be automatically selected
  };

  if (isLoading || !currentOrganization) {
    return (
      <div className="flex items-center gap-2 px-2 py-1.5">
        <div className="h-8 w-8 rounded-lg bg-muted animate-pulse" />
        {sidebarOpen && (
          <div className="flex-1">
            <div className="h-4 w-24 bg-muted animate-pulse rounded mb-1" />
            <div className="h-3 w-16 bg-muted animate-pulse rounded" />
          </div>
        )}
      </div>
    );
  }

  return (
    <>
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="ghost"
            role="combobox"
            aria-expanded={open}
            aria-label="Select workspace"
            className={cn(
              "w-full justify-between h-auto p-2",
              !sidebarOpen && "px-2",
            )}
          >
            <div className="flex items-center gap-2 min-w-0">
              <OrganizationAvatar
                name={currentOrganization.name}
                logoUrl={currentOrganization.logoUrl}
                size="sm"
              />
              {sidebarOpen && (
                <div className="flex-1 text-left min-w-0">
                  <div className="font-medium text-sm truncate">
                    {currentOrganization.name}
                  </div>
                  {/* <div className="flex items-center gap-1">
                    <Badge 
                      variant="secondary" 
                      className={cn("text-xs h-4 px-1.5", getTierColor(currentOrganization.tier))}
                    >
                      {currentOrganization.tier || "STARTER"}
                    </Badge>
                  </div> */}
                </div>
              )}
            </div>
            {sidebarOpen && (
              <ChevronsUpDown className="ml-auto h-4 w-4 shrink-0 opacity-50" />
            )}
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-[240px] p-0" align="start">
          <Command>
            <CommandInput placeholder="Search workspaces..." />
            <CommandList>
              <CommandEmpty>No workspaces found.</CommandEmpty>
              <CommandGroup heading="Workspaces">
                {userOrganizations.map((org) => (
                  <CommandItem
                    key={org.id}
                    value={org.name}
                    onSelect={() => handleSelectWorkspace(org.id)}
                    className="flex items-center gap-2 p-2"
                  >
                    <OrganizationAvatar
                      name={org.name}
                      logoUrl={org.logoUrl}
                      size="sm"
                    />
                    <div className="flex-1 min-w-0">
                      <div className="font-medium text-xs truncate">
                        {org.name}
                      </div>
                      <div className="flex items-center gap-2">
                        {org.description && (
                          <span className="text-xs text-muted-foreground truncate">
                            {org.description}
                          </span>
                        )}
                      </div>
                    </div>
                    <Check
                      className={cn(
                        "ml-auto h-4 w-4",
                        currentOrganization?.id === org.id
                          ? "opacity-100"
                          : "opacity-0",
                      )}
                    />
                  </CommandItem>
                ))}
              </CommandGroup>
              <CommandSeparator />
              <CommandGroup>
                <CommandItem
                  onSelect={() => {
                    setOpen(false);
                    setShowCreateModal(true);
                  }}
                  className="flex items-center gap-2 p-2"
                >
                  <div className="h-8 w-8 rounded-lg border border-dashed border-muted-foreground/50 flex items-center justify-center">
                    <Plus className="h-4 w-4" />
                  </div>
                  <div className="flex-1">
                    <div className="font-medium text-sm">Create workspace</div>
                    <div className="text-xs text-muted-foreground">
                      Start a new organization
                    </div>
                  </div>
                </CommandItem>
              </CommandGroup>
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>

      <CreateOrganizationModal
        open={showCreateModal}
        onOpenChange={setShowCreateModal}
        onSuccess={handleCreateSuccess}
      />
    </>
  );
}
