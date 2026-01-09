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
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
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

  const getInitials = (name: string) => {
    return name
      .split(" ")
      .map((n) => n[0])
      .join("")
      .toUpperCase()
      .slice(0, 2);
  };

  const getTierColor = (tier?: string) => {
    switch (tier?.toUpperCase()) {
      case "PRO":
        return "bg-purple-100 text-purple-700";
      case "ENTERPRISE":
        return "bg-emerald-100 text-emerald-700";
      default:
        return "bg-blue-100 text-blue-700";
    }
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
              !sidebarOpen && "px-2"
            )}
          >
            <div className="flex items-center gap-2 min-w-0">
              <Avatar className="h-7 w-7 rounded-lg">
                <AvatarImage
                  src={currentOrganization.logoUrl}
                  alt={currentOrganization.name}
                />
                <AvatarFallback className="rounded-lg  bg-primary text-primary-foreground">
                  <span className="text-base">
                    {getInitials(currentOrganization.name)}
                  </span>
                </AvatarFallback>
              </Avatar>
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
                    <Avatar className="h-7 w-7 rounded-lg">
                      <AvatarImage src={org.logoUrl} alt={org.name} />
                      <AvatarFallback className="rounded-lg bg-muted !text-base">
                        {getInitials(org.name)}
                      </AvatarFallback>
                    </Avatar>
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
                          : "opacity-0"
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
