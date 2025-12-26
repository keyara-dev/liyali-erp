"use client";

import { useState } from "react";
import { useOrganizationContext } from "@/contexts/organization-context";
import { useSession } from "@/hooks/use-session";
import {
  useSelectOrganization,
  useLogout,
} from "@/hooks/use-organization-mutations";
import { LogOut, Loader2, ArrowRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { Spinner } from "@/components";

export default function WelcomePage() {
  const { user } = useSession();
  const { userOrganizations, currentOrganization, isLoading } =
    useOrganizationContext();
  const { selectOrganization, isPending: isNavigating } =
    useSelectOrganization();
  const { logout } = useLogout();
  const [selectedOrgId, setSelectedOrgId] = useState<string | null>(
    currentOrganization?.id ?? null
  );

  const handleSelectOrganization = async (orgId: string) => {
    if (isNavigating) return;
    setSelectedOrgId(orgId);
    await selectOrganization(orgId);
  };

  const handleLogout = async () => {
    await logout();
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center">
        <div className="flex flex-col items-center gap-3">
          <Spinner className="h-8 w-8 animate-spin text-blue-600" />
          <p className="text-slate-600">Loading organizations...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
      {/* Header */}
      <div className="border-b border-slate-200 bg-white/50 backdrop-blur-sm sticky top-0">
        <div className="max-w-4xl mx-auto px-6 py-4 flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-slate-900">
              Liyali Gateway
            </h1>
            <p className="text-sm text-slate-600">
              Select a workspace to continue
            </p>
          </div>
          <Button
            variant="ghost"
            size="sm"
            onClick={handleLogout}
            className="text-slate-600 hover:text-slate-900"
          >
            <LogOut className="h-4 w-4 mr-2" />
            Sign Out
          </Button>
        </div>
      </div>

      {/* Main Content */}
      <div className="max-w-4xl mx-auto px-6 py-12">
        {/* User Info */}
        <div className="mb-10">
          <p className="text-slate-600 text-sm">Signed in as</p>
          <p className="text-lg font-medium text-slate-900">
            {user?.email || "User"}
          </p>
        </div>

        {/* Workspaces Section */}
        <div>
          <div className="mb-6">
            <h2 className="text-lg font-semibold text-slate-900">
              Select a workspace
            </h2>
            <p className="text-sm text-slate-600">
              Choose where you'd like to work
            </p>
          </div>

          {userOrganizations.length === 0 ? (
            // No organizations state
            <div className="text-center py-12 bg-white rounded-lg border border-slate-200">
              <p className="text-slate-600 mb-4">No workspaces available</p>
              <Button onClick={handleLogout} variant="outline">
                Sign Out
              </Button>
            </div>
          ) : (
            // Organizations Grid
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {userOrganizations.map((org) => {
                const isDefault = currentOrganization?.id === org.id;
                const isSelected = selectedOrgId === org.id;

                return (
                  <button
                    key={org.id}
                    onClick={() => handleSelectOrganization(org.id)}
                    disabled={isNavigating}
                    className={cn(
                      "relative overflow-hidden rounded-lg border-2 transition-all duration-200",
                      "hover:shadow-lg hover:border-blue-400",
                      "disabled:opacity-50 disabled:cursor-not-allowed",
                      "text-left p-6 bg-white",
                      isDefault &&
                        "border-blue-500 ring-2 ring-blue-200 ring-offset-2",
                      !isDefault && "border-slate-200",
                      isNavigating && isSelected && "ring-2 ring-blue-400"
                    )}
                  >
                    {/* Default Badge */}
                    {isDefault && (
                      <div className="absolute top-3 right-3">
                        <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                          Default
                        </span>
                      </div>
                    )}

                    {/* Organization Logo and Name */}
                    <div className="flex items-start gap-4 mb-3">
                      <div>
                        {org.logoUrl ? (
                          <img
                            src={org.logoUrl}
                            alt={org.name}
                            className="w-12 h-12 rounded-lg object-cover flex-shrink-0"
                          />
                        ) : (
                          <div
                            className="w-12 h-12 rounded-lg flex items-center justify-center text-lg font-bold text-white flex-shrink-0"
                            style={{
                              backgroundColor: org.primaryColor || "#0066CC",
                            }}
                          >
                            {org.name[0].toUpperCase()}
                          </div>
                        )}
                      </div>
                      <div className="flex-1">
                        <h3 className="font-semibold text-slate-900 text-base">
                          {org.name}
                        </h3>
                        {org.description && (
                          <p className="text-sm text-slate-600 mt-1 line-clamp-2">
                            {org.description}
                          </p>
                        )}
                      </div>
                    </div>

                    {/* Organization Details */}
                    <div className="flex items-center justify-between mt-4 pt-4 border-t border-slate-100">
                      <div className="flex items-center gap-4">
                        <div>
                          <p className="text-xs text-slate-500 uppercase tracking-wide">
                            Tier
                          </p>
                          <p className="text-sm font-medium text-slate-900 capitalize">
                            {org.tier}
                          </p>
                        </div>
                        <div>
                          <p className="text-xs text-slate-500 uppercase tracking-wide">
                            Status
                          </p>
                          <p className="text-sm font-medium text-green-600">
                            Active
                          </p>
                        </div>
                      </div>
                      <div>
                        {isNavigating && isSelected ? (
                          <Loader2 className="h-5 w-5 animate-spin text-blue-600" />
                        ) : (
                          <ArrowRight className="h-5 w-5 text-slate-400 group-hover:text-blue-600 transition-colors" />
                        )}
                      </div>
                    </div>
                  </button>
                );
              })}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="mt-12 pt-8 border-t border-slate-200">
          <p className="text-xs text-slate-500 text-center">
            Need help? Contact support at support@liyali.com
          </p>
        </div>
      </div>
    </div>
  );
}
