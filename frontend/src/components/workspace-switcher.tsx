'use client';

import { useState } from 'react';
import { useOrganizationContext } from '@/contexts/organization-context';
import { Check, ChevronsUpDown, Plus, Loader } from 'lucide-react';
import { cn } from '@/lib/utils';

export function WorkspaceSwitcher() {
  const {
    currentOrganization,
    userOrganizations,
    switchWorkspace,
    isLoading: contextLoading,
  } = useOrganizationContext();

  const [isOpen, setIsOpen] = useState(false);
  const [isSwitching, setIsSwitching] = useState(false);

  const handleSwitchWorkspace = async (orgId: string) => {
    if (orgId === currentOrganization?.id) {
      setIsOpen(false);
      return;
    }

    setIsSwitching(true);
    try {
      await switchWorkspace(orgId);
      setIsOpen(false);
    } catch (error) {
      console.error('Failed to switch workspace:', error);
    } finally {
      setIsSwitching(false);
    }
  };

  if (contextLoading) {
    return (
      <div className="w-full h-10 bg-slate-100 rounded-md animate-pulse flex items-center px-3">
        <Loader className="h-4 w-4 animate-spin" />
      </div>
    );
  }

  if (!currentOrganization) {
    return (
      <div className="w-full h-10 bg-slate-100 rounded-md flex items-center px-3 text-sm text-slate-500">
        No organization selected
      </div>
    );
  }

  return (
    <div className="relative w-full">
      <button
        onClick={() => setIsOpen(!isOpen)}
        disabled={isSwitching}
        className={cn(
          'w-full h-10 px-3 py-2 text-sm font-medium rounded-md border border-slate-200 bg-white',
          'hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500',
          'flex items-center justify-between transition-colors',
          'disabled:opacity-50 disabled:cursor-not-allowed'
        )}
      >
        <div className="flex items-center gap-2 min-w-0">
          {currentOrganization.logoUrl ? (
            <img
              src={currentOrganization.logoUrl}
              alt={currentOrganization.name}
              className="w-5 h-5 rounded flex-shrink-0"
            />
          ) : (
            <div
              className="w-5 h-5 rounded flex items-center justify-center text-xs font-semibold text-white flex-shrink-0"
              style={{ backgroundColor: currentOrganization.primaryColor || '#0066CC' }}
            >
              {currentOrganization.name[0].toUpperCase()}
            </div>
          )}
          <span className="truncate">{currentOrganization.name}</span>
        </div>
        <ChevronsUpDown className="h-4 w-4 flex-shrink-0 opacity-50" />
      </button>

      {isOpen && (
        <div className="absolute top-full left-0 right-0 mt-2 bg-white rounded-md border border-slate-200 shadow-lg z-50">
          <div className="p-2">
            {/* Header */}
            <div className="px-2 py-1.5 text-xs font-semibold text-slate-500 uppercase tracking-wide">
              Workspaces
            </div>

            {/* Organization list */}
            {userOrganizations.length > 0 ? (
              <div className="space-y-1">
                {userOrganizations.map((org) => (
                  <button
                    key={org.id}
                    onClick={() => handleSwitchWorkspace(org.id)}
                    disabled={isSwitching}
                    className={cn(
                      'w-full px-2 py-2 rounded text-sm text-left flex items-center gap-2 transition-colors',
                      'hover:bg-slate-100 disabled:opacity-50 disabled:cursor-not-allowed',
                      currentOrganization.id === org.id && 'bg-blue-50'
                    )}
                  >
                    {org.logoUrl ? (
                      <img
                        src={org.logoUrl}
                        alt={org.name}
                        className="w-4 h-4 rounded flex-shrink-0"
                      />
                    ) : (
                      <div
                        className="w-4 h-4 rounded flex items-center justify-center text-xs font-semibold text-white flex-shrink-0"
                        style={{ backgroundColor: org.primaryColor || '#0066CC' }}
                      >
                        {org.name[0].toUpperCase()}
                      </div>
                    )}
                    <span className="truncate flex-1">{org.name}</span>
                    {currentOrganization.id === org.id && (
                      <Check className="h-4 w-4 flex-shrink-0 text-blue-600" />
                    )}
                  </button>
                ))}
              </div>
            ) : (
              <div className="px-2 py-2 text-sm text-slate-500">
                No organizations available
              </div>
            )}

            {/* Divider */}
            {userOrganizations.length > 0 && (
              <div className="my-2 border-t border-slate-200" />
            )}

            {/* Create new workspace */}
            <button
              className={cn(
                'w-full px-2 py-2 rounded text-sm text-left flex items-center gap-2',
                'text-slate-700 hover:bg-slate-100 transition-colors'
              )}
            >
              <Plus className="h-4 w-4" />
              Create workspace
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
