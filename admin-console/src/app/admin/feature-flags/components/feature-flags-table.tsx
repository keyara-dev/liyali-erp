"use client";

import { useState } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Checkbox } from "@/components/ui/checkbox";
import { Switch } from "@/components/ui/switch";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  MoreHorizontal,
  Edit,
  Trash2,
  Copy,
  Archive,
  BarChart3,
  Eye,
  Flag,
  Beaker,
  Shield,
  AlertTriangle,
  Users,
  Clock,
  Target,
  TrendingUp,
} from "lucide-react";
import { format } from "date-fns";
import { cn } from "@/lib/utils";
import type { FeatureFlag } from "@/app/_actions/feature-flags";

interface FeatureFlagsTableProps {
  flags: FeatureFlag[];
  selectedFlags: string[];
  onSelectionChange: (flagIds: string[]) => void;
  onEdit: (flag: FeatureFlag) => void;
  onDelete: (flagId: string) => void;
  onToggle: (flagId: string) => void;
  onArchive: (flagId: string) => void;
  onDuplicate: (flag: FeatureFlag) => void;
  onViewAnalytics: (flag: FeatureFlag) => void;
  isLoading?: boolean;
}

export function FeatureFlagsTable({
  flags,
  selectedFlags,
  onSelectionChange,
  onEdit,
  onDelete,
  onToggle,
  onArchive,
  onDuplicate,
  onViewAnalytics,
  isLoading = false,
}: FeatureFlagsTableProps) {
  const categoryIcons = {
    feature: Flag,
    experiment: Beaker,
    operational: Shield,
    killswitch: AlertTriangle,
    permission: Users,
  };

  const categoryColors = {
    feature: "bg-blue-100 text-blue-800",
    experiment: "bg-green-100 text-green-800",
    operational: "bg-yellow-100 text-yellow-800",
    killswitch: "bg-red-100 text-red-800",
    permission: "bg-purple-100 text-purple-800",
  };

  const typeColors = {
    boolean: "bg-blue-100 text-blue-800",
    string: "bg-green-100 text-green-800",
    number: "bg-orange-100 text-orange-800",
    json: "bg-purple-100 text-purple-800",
  };

  const environmentColors = {
    all: "bg-gray-100 text-gray-800",
    production: "bg-red-100 text-red-800",
    staging: "bg-yellow-100 text-yellow-800",
    development: "bg-green-100 text-green-800",
  };

  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      onSelectionChange(flags.map((flag) => flag.id));
    } else {
      onSelectionChange([]);
    }
  };

  const handleSelectFlag = (flagId: string, checked: boolean) => {
    if (checked) {
      onSelectionChange([...selectedFlags, flagId]);
    } else {
      onSelectionChange(selectedFlags.filter((id) => id !== flagId));
    }
  };

  const isAllSelected =
    flags.length > 0 && selectedFlags.length === flags.length;
  const isPartiallySelected =
    selectedFlags.length > 0 && selectedFlags.length < flags.length;

  const isExpiringSoon = (expiresAt?: string) => {
    if (!expiresAt) return false;
    const expiryDate = new Date(expiresAt);
    const thirtyDaysFromNow = new Date();
    thirtyDaysFromNow.setDate(thirtyDaysFromNow.getDate() + 30);
    return expiryDate <= thirtyDaysFromNow;
  };

  if (isLoading) {
    return (
      <div className="border rounded-lg">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-12">
                <div className="h-4 w-4 bg-muted animate-pulse rounded" />
              </TableHead>
              <TableHead>Flag</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Category</TableHead>
              <TableHead>Environment</TableHead>
              <TableHead>Targeting</TableHead>
              <TableHead>Evaluations</TableHead>
              <TableHead>Updated</TableHead>
              <TableHead className="w-12"></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {Array.from({ length: 5 }).map((_, i) => (
              <TableRow key={i}>
                <TableCell>
                  <div className="h-4 w-4 bg-muted animate-pulse rounded" />
                </TableCell>
                <TableCell>
                  <div className="h-4 bg-muted animate-pulse rounded w-32" />
                </TableCell>
                <TableCell>
                  <div className="h-6 bg-muted animate-pulse rounded w-16" />
                </TableCell>
                <TableCell>
                  <div className="h-6 bg-muted animate-pulse rounded w-20" />
                </TableCell>
                <TableCell>
                  <div className="h-6 bg-muted animate-pulse rounded w-24" />
                </TableCell>
                <TableCell>
                  <div className="h-4 bg-muted animate-pulse rounded w-16" />
                </TableCell>
                <TableCell>
                  <div className="h-4 bg-muted animate-pulse rounded w-12" />
                </TableCell>
                <TableCell>
                  <div className="h-4 bg-muted animate-pulse rounded w-20" />
                </TableCell>
                <TableCell>
                  <div className="h-8 w-8 bg-muted animate-pulse rounded" />
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    );
  }

  return (
    <div className="border rounded-lg">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-12">
              <Checkbox
                checked={isAllSelected}
                onCheckedChange={handleSelectAll}
                ref={(el) => {
                  if (el) {
                    const input = el.querySelector(
                      'input[type="checkbox"]',
                    ) as HTMLInputElement;
                    if (input) input.indeterminate = isPartiallySelected;
                  }
                }}
              />
            </TableHead>
            <TableHead>Flag</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Category</TableHead>
            <TableHead>Environment</TableHead>
            <TableHead>Targeting</TableHead>
            <TableHead>Evaluations</TableHead>
            <TableHead>Updated</TableHead>
            <TableHead className="w-12"></TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {flags.map((flag) => {
            const CategoryIcon =
              categoryIcons[flag.category as keyof typeof categoryIcons] ||
              Flag;
            const hasTargeting =
              flag.targeting.enabled || flag.targeting.rolloutPercentage > 0;

            return (
              <TableRow
                key={flag.id}
                className={cn(flag.is_archived && "opacity-60")}
              >
                <TableCell>
                  <Checkbox
                    checked={selectedFlags.includes(flag.id)}
                    onCheckedChange={(checked) =>
                      handleSelectFlag(flag.id, checked as boolean)
                    }
                  />
                </TableCell>

                <TableCell>
                  <div className="space-y-1">
                    <div className="flex items-center space-x-2">
                      <span className="font-medium">{flag.name}</span>
                      {flag.is_archived && (
                        <Archive className="h-3 w-3 text-muted-foreground" />
                      )}
                      {isExpiringSoon(flag.expires_at) && (
                        <Clock className="h-3 w-3 text-amber-500" />
                      )}
                    </div>
                    <div className="text-xs text-muted-foreground font-mono">
                      {flag.key}
                    </div>
                    {flag.description && (
                      <p className="text-xs text-muted-foreground max-w-xs truncate">
                        {flag.description}
                      </p>
                    )}
                    {flag.tags.length > 0 && (
                      <div className="flex flex-wrap gap-1 mt-1">
                        {flag.tags.slice(0, 3).map((tag, index) => (
                          <Badge
                            key={index}
                            variant="outline"
                            className="text-xs"
                          >
                            {tag}
                          </Badge>
                        ))}
                        {flag.tags.length > 3 && (
                          <Badge variant="outline" className="text-xs">
                            +{flag.tags.length - 3}
                          </Badge>
                        )}
                      </div>
                    )}
                  </div>
                </TableCell>

                <TableCell>
                  <div className="flex items-center space-x-2">
                    <Switch
                      checked={flag.enabled}
                      onCheckedChange={() => onToggle(flag.id)}
                      disabled={flag.is_archived}
                    />
                    <Badge
                      variant={flag.enabled ? "default" : "secondary"}
                      className={cn(
                        flag.enabled
                          ? "bg-green-100 text-green-800"
                          : "bg-gray-100 text-gray-800",
                      )}
                    >
                      {flag.enabled ? "Enabled" : "Disabled"}
                    </Badge>
                  </div>
                  <div className="flex items-center space-x-1 mt-1">
                    <Badge
                      variant="outline"
                      className={
                        typeColors[flag.type as keyof typeof typeColors]
                      }
                    >
                      {flag.type}
                    </Badge>
                  </div>
                </TableCell>

                <TableCell>
                  <div className="flex items-center space-x-2">
                    <CategoryIcon className="h-4 w-4 text-muted-foreground" />
                    <Badge
                      variant="secondary"
                      className={
                        categoryColors[
                          flag.category as keyof typeof categoryColors
                        ]
                      }
                    >
                      {flag.category}
                    </Badge>
                  </div>
                </TableCell>

                <TableCell>
                  <Badge
                    variant="outline"
                    className={
                      environmentColors[
                        flag.environment as keyof typeof environmentColors
                      ]
                    }
                  >
                    {flag.environment === "all" ? "All" : flag.environment}
                  </Badge>
                </TableCell>

                <TableCell>
                  <div className="space-y-1">
                    {hasTargeting ? (
                      <div className="flex items-center space-x-1">
                        <Target className="h-3 w-3 text-blue-500" />
                        <span className="text-xs text-blue-600">Active</span>
                      </div>
                    ) : (
                      <span className="text-xs text-muted-foreground">
                        None
                      </span>
                    )}
                    {flag.targeting.rolloutPercentage > 0 && (
                      <div className="text-xs text-muted-foreground">
                        {flag.targeting.rolloutPercentage}% rollout
                      </div>
                    )}
                    {flag.targeting.rules.length > 0 && (
                      <div className="text-xs text-muted-foreground">
                        {flag.targeting.rules.length} rule
                        {flag.targeting.rules.length > 1 ? "s" : ""}
                      </div>
                    )}
                  </div>
                </TableCell>

                <TableCell>
                  <div className="space-y-1">
                    <div className="flex items-center space-x-1">
                      <TrendingUp className="h-3 w-3 text-muted-foreground" />
                      <span className="text-sm font-medium">
                        {flag.evaluation_count.toLocaleString()}
                      </span>
                    </div>
                    {flag.last_evaluated && (
                      <div className="text-xs text-muted-foreground">
                        Last: {format(new Date(flag.last_evaluated), "HH:mm")}
                      </div>
                    )}
                  </div>
                </TableCell>

                <TableCell>
                  <div className="text-sm">
                    <div>
                      {format(new Date(flag.updated_at), "MMM dd, yyyy")}
                    </div>
                    <div className="text-xs text-muted-foreground">
                      {format(new Date(flag.updated_at), "HH:mm")}
                    </div>
                    <div className="text-xs text-muted-foreground">
                      by {flag.updated_by.split("@")[0]}
                    </div>
                  </div>
                  {flag.expires_at && (
                    <div
                      className={cn(
                        "text-xs mt-1",
                        isExpiringSoon(flag.expires_at)
                          ? "text-amber-600"
                          : "text-muted-foreground",
                      )}
                    >
                      Expires: {format(new Date(flag.expires_at), "MMM dd")}
                    </div>
                  )}
                </TableCell>

                <TableCell>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="ghost" className="h-8 w-8 p-0">
                        <MoreHorizontal className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem onClick={() => onEdit(flag)}>
                        <Edit className="mr-2 h-4 w-4" />
                        Edit Flag
                      </DropdownMenuItem>

                      <DropdownMenuItem onClick={() => onDuplicate(flag)}>
                        <Copy className="mr-2 h-4 w-4" />
                        Duplicate
                      </DropdownMenuItem>

                      <DropdownMenuItem onClick={() => onViewAnalytics(flag)}>
                        <BarChart3 className="mr-2 h-4 w-4" />
                        View Analytics
                      </DropdownMenuItem>

                      <DropdownMenuSeparator />

                      {!flag.is_archived ? (
                        <DropdownMenuItem onClick={() => onArchive(flag.id)}>
                          <Archive className="mr-2 h-4 w-4" />
                          Archive Flag
                        </DropdownMenuItem>
                      ) : (
                        <DropdownMenuItem onClick={() => onArchive(flag.id)}>
                          <Eye className="mr-2 h-4 w-4" />
                          Unarchive Flag
                        </DropdownMenuItem>
                      )}

                      <DropdownMenuSeparator />

                      <DropdownMenuItem
                        onClick={() => onDelete(flag.id)}
                        className="text-red-600"
                        disabled={flag.enabled && !flag.is_archived}
                      >
                        <Trash2 className="mr-2 h-4 w-4" />
                        Delete Flag
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </TableCell>
              </TableRow>
            );
          })}
        </TableBody>
      </Table>

      {flags.length === 0 && (
        <div className="text-center py-12">
          <Flag className="mx-auto h-12 w-12 text-muted-foreground" />
          <h3 className="mt-4 text-lg font-semibold">No feature flags found</h3>
          <p className="text-muted-foreground">
            No flags match your current filters.
          </p>
        </div>
      )}
    </div>
  );
}
