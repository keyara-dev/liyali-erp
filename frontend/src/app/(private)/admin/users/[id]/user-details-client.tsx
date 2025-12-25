// =============================================================================
// FILE: app/dashboard/system-configs/users/[id]/user-details-client.tsx
// =============================================================================
"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { User } from "@/types";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import {
  ArrowLeft,
  Mail,
  Building2,
  Briefcase,
  Shield,
  Activity,
  AlertTriangle,
  CheckCircle2,
  Clock,
  FileText,
  TrendingUp,
  User as UserIcon,
  Calendar,
  Edit,
} from "lucide-react";
import { generateAvatarFallback, getAvatarSrc } from "@/lib/utils";
import { Separator } from "@/components/ui/separator";
import CreateUserForm from "../_components/create-user-dialog";

interface UserDetailsClientProps {
  user: User;
}

export function UserDetailsClient({ user }: UserDetailsClientProps) {
  const router = useRouter();
  const [activeTab, setActiveTab] = useState("overview");
  const [isEditingUser, setIsEditingUser] = useState(false);

  const fullName = `${user.first_name} ${user.last_name}`;
  const avatarFallback = generateAvatarFallback(fullName);

  // Mock data - Replace with actual API calls
  const riskMetrics = {
    totalRisks: 12,
    criticalRisks: 2,
    highRisks: 5,
    mediumRisks: 3,
    lowRisks: 2,
    riskScore: 68,
  };

  const auditMetrics = {
    totalAudits: 8,
    completedAudits: 5,
    inProgressAudits: 2,
    upcomingAudits: 1,
    findingsCreated: 24,
    openFindings: 6,
  };

  const recentActivities = [
    {
      id: "1",
      type: "audit",
      action: "Created audit plan",
      details: "ISO 27001:2022 Annual Audit",
      timestamp: new Date(2025, 10, 1, 10, 30),
    },
    {
      id: "2",
      type: "risk",
      action: "Updated risk assessment",
      details: "Cybersecurity Risk - Data Breach",
      timestamp: new Date(2025, 9, 30, 14, 15),
    },
    {
      id: "3",
      type: "finding",
      action: "Resolved finding",
      details: "Missing encryption on database",
      timestamp: new Date(2025, 9, 29, 9, 45),
    },
    {
      id: "4",
      type: "login",
      action: "Logged in",
      details: "From IP: 192.168.1.100",
      timestamp: new Date(2025, 9, 28, 8, 0),
    },
  ];

  const getActivityIcon = (type: string) => {
    switch (type) {
      case "audit":
        return <FileText className="h-4 w-4" />;
      case "risk":
        return <AlertTriangle className="h-4 w-4" />;
      case "finding":
        return <CheckCircle2 className="h-4 w-4" />;
      case "login":
        return <Activity className="h-4 w-4" />;
      default:
        return <Activity className="h-4 w-4" />;
    }
  };

  const formatTimestamp = (date: Date) => {
    const now = new Date();
    const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000);

    if (diffInSeconds < 60) return "Just now";
    if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)}m ago`;
    if (diffInSeconds < 86400)
      return `${Math.floor(diffInSeconds / 3600)}h ago`;
    if (diffInSeconds < 604800)
      return `${Math.floor(diffInSeconds / 86400)}d ago`;

    return date.toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
    });
  };

  return (
    <>
      {/* Header Section */}
      <div className="bg-card border-b">
        <div className="container mx-auto px-4 py-6">
          <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <div className="flex items-center gap-4">
              <Button
                variant="outline"
                size="icon"
                onClick={() => router.back()}
                className="from-primary to-primary/80 flex h-12 w-12 items-center justify-center rounded-xl bg-linear-to-br shadow-lg"
              >
                <ArrowLeft className="text-primary-foreground h-7 w-7" />
              </Button>
              <div>
                <h1 className="text-2xl font-bold tracking-tight">
                  User Details
                </h1>
                <p className="text-muted-foreground text-sm">
                  View and manage user information
                </p>
              </div>
            </div>
            <Button onClick={() => setIsEditingUser(true)}>
              <Edit className="mr-2 h-4 w-4" />
              Edit User
            </Button>
          </div>
        </div>
      </div>

      <div className="container mx-auto space-y-6 px-4 py-6">
        {/* User Profile Card */}
        <Card>
          <CardContent className="pt-6">
            <div className="flex flex-col gap-6 md:flex-row md:items-start">
              <Avatar className="h-24 w-24">
                <AvatarImage src={getAvatarSrc(fullName)} />
                <AvatarFallback className="text-2xl">
                  {avatarFallback}
                </AvatarFallback>
              </Avatar>

              <div className="flex-1 space-y-4">
                <div>
                  <div className="flex items-center gap-2">
                    <h1 className="text-3xl font-bold">{fullName}</h1>
                    <Badge variant={user.is_active ? "default" : "destructive"}>
                      {user.is_active ? "Active" : "Inactive"}
                    </Badge>
                    {user.mfa_enabled && (
                      <Badge variant="outline" className="gap-1">
                        <Shield className="h-3 w-3" />
                        MFA Enabled
                      </Badge>
                    )}
                  </div>
                  <p className="text-muted-foreground mt-1">
                    {user.username || user.email}
                  </p>
                </div>

                <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
                  <div className="flex items-center gap-3">
                    <div className="bg-primary/10 flex h-10 w-10 items-center justify-center rounded-lg">
                      <Mail className="text-primary h-5 w-5" />
                    </div>
                    <div>
                      <p className="text-muted-foreground text-xs">Email</p>
                      <p className="text-sm font-medium">{user.email}</p>
                    </div>
                  </div>

                  <div className="flex items-center gap-3">
                    <div className="bg-primary/10 flex h-10 w-10 items-center justify-center rounded-lg">
                      <Briefcase className="text-primary h-5 w-5" />
                    </div>
                    <div>
                      <p className="text-muted-foreground text-xs">
                        Department
                      </p>
                      <p className="text-sm font-medium">
                        {user.department || "N/A"}
                      </p>
                    </div>
                  </div>

                  <div className="flex items-center gap-3">
                    <div className="bg-primary/10 flex h-10 w-10 items-center justify-center rounded-lg">
                      <UserIcon className="text-primary h-5 w-5" />
                    </div>
                    <div>
                      <p className="text-muted-foreground text-xs">Role</p>
                      <p className="text-sm font-medium">
                        {user.role || "N/A"}
                      </p>
                    </div>
                  </div>
                </div>

                <div className="text-muted-foreground flex items-center gap-4 text-sm">
                  {user.created_at && (
                    <div className="flex items-center gap-1">
                      <Calendar className="h-4 w-4" />
                      <span>
                        Joined{" "}
                        {new Date(user.created_at).toLocaleDateString("en-US", {
                          month: "short",
                          year: "numeric",
                        })}
                      </span>
                    </div>
                  )}
                  {user.last_login && (
                    <div className="flex items-center gap-1">
                      <Clock className="h-4 w-4" />
                      <span>
                        Last login: {formatTimestamp(new Date(user.last_login))}
                      </span>
                    </div>
                  )}
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Tabs for different sections */}
        <Tabs value={activeTab} onValueChange={setActiveTab}>
          <TabsList className="grid w-full grid-cols-3">
            <TabsTrigger value="overview">Overview</TabsTrigger>
            <TabsTrigger value="metrics">Metrics</TabsTrigger>
            <TabsTrigger value="activity">Recent Activity</TabsTrigger>
          </TabsList>

          {/* Overview Tab */}
          <TabsContent value="overview" className="space-y-6">
            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
              {/* Risk Metrics Summary */}
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2 text-lg">
                    <AlertTriangle className="h-5 w-5" />
                    Risk Overview
                  </CardTitle>
                  <CardDescription>Summary of risk assessments</CardDescription>
                </CardHeader>
                <CardContent className="space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-muted-foreground text-sm">
                      Total Risks
                    </span>
                    <span className="text-2xl font-bold">
                      {riskMetrics.totalRisks}
                    </span>
                  </div>
                  <Separator />
                  <div className="space-y-2">
                    <div className="flex items-center justify-between text-sm">
                      <div className="flex items-center gap-2">
                        <div className="h-3 w-3 rounded-full bg-red-500" />
                        <span>Critical</span>
                      </div>
                      <span className="font-medium">
                        {riskMetrics.criticalRisks}
                      </span>
                    </div>
                    <div className="flex items-center justify-between text-sm">
                      <div className="flex items-center gap-2">
                        <div className="h-3 w-3 rounded-full bg-orange-500" />
                        <span>High</span>
                      </div>
                      <span className="font-medium">
                        {riskMetrics.highRisks}
                      </span>
                    </div>
                    <div className="flex items-center justify-between text-sm">
                      <div className="flex items-center gap-2">
                        <div className="h-3 w-3 rounded-full bg-yellow-500" />
                        <span>Medium</span>
                      </div>
                      <span className="font-medium">
                        {riskMetrics.mediumRisks}
                      </span>
                    </div>
                    <div className="flex items-center justify-between text-sm">
                      <div className="flex items-center gap-2">
                        <div className="h-3 w-3 rounded-full bg-green-500" />
                        <span>Low</span>
                      </div>
                      <span className="font-medium">
                        {riskMetrics.lowRisks}
                      </span>
                    </div>
                  </div>
                  <Separator />
                  <div className="flex items-center justify-between">
                    <span className="text-muted-foreground text-sm">
                      Risk Score
                    </span>
                    <div className="flex items-center gap-2">
                      <TrendingUp className="h-4 w-4 text-orange-500" />
                      <span className="text-xl font-bold">
                        {riskMetrics.riskScore}
                      </span>
                    </div>
                  </div>
                </CardContent>
              </Card>

              {/* Audit Metrics Summary */}
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2 text-lg">
                    <FileText className="h-5 w-5" />
                    Audit Overview
                  </CardTitle>
                  <CardDescription>Summary of audit activities</CardDescription>
                </CardHeader>
                <CardContent className="space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-muted-foreground text-sm">
                      Total Audits
                    </span>
                    <span className="text-2xl font-bold">
                      {auditMetrics.totalAudits}
                    </span>
                  </div>
                  <Separator />
                  <div className="space-y-2">
                    <div className="flex items-center justify-between text-sm">
                      <div className="flex items-center gap-2">
                        <CheckCircle2 className="h-4 w-4 text-green-500" />
                        <span>Completed</span>
                      </div>
                      <span className="font-medium">
                        {auditMetrics.completedAudits}
                      </span>
                    </div>
                    <div className="flex items-center justify-between text-sm">
                      <div className="flex items-center gap-2">
                        <Activity className="h-4 w-4 text-blue-500" />
                        <span>In Progress</span>
                      </div>
                      <span className="font-medium">
                        {auditMetrics.inProgressAudits}
                      </span>
                    </div>
                    <div className="flex items-center justify-between text-sm">
                      <div className="flex items-center gap-2">
                        <Clock className="h-4 w-4 text-gray-500" />
                        <span>Upcoming</span>
                      </div>
                      <span className="font-medium">
                        {auditMetrics.upcomingAudits}
                      </span>
                    </div>
                  </div>
                  <Separator />
                  <div className="space-y-2">
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground">
                        Findings Created
                      </span>
                      <span className="font-medium">
                        {auditMetrics.findingsCreated}
                      </span>
                    </div>
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground">
                        Open Findings
                      </span>
                      <span className="font-medium">
                        {auditMetrics.openFindings}
                      </span>
                    </div>
                  </div>
                </CardContent>
              </Card>

              {/* Account Information */}
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2 text-lg">
                    <Shield className="h-5 w-5" />
                    Account Information
                  </CardTitle>
                  <CardDescription>
                    Security and account details
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-3">
                  <div className="space-y-2">
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground">
                        Account Type
                      </span>
                      <Badge variant="outline">
                        {user.role?.replace("_", " ") || "ORGANIZATIONAL USER"}
                      </Badge>
                    </div>
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground">LDAP User</span>
                      <Badge
                        variant={user.is_ldap_user ? "default" : "secondary"}
                      >
                        {user.is_ldap_user ? "Yes" : "No"}
                      </Badge>
                    </div>
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground">MFA Status</span>
                      <Badge
                        variant={user.mfa_enabled ? "default" : "secondary"}
                      >
                        {user.mfa_enabled ? "Enabled" : "Disabled"}
                      </Badge>
                    </div>
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground">
                        Account Status
                      </span>
                      <Badge
                        variant={user.is_active ? "success" : "destructive"}
                      >
                        {user.is_active ? "Active" : "Inactive"}
                      </Badge>
                    </div>
                  </div>
                  <Separator />
                  <div className="space-y-2">
                    {user.created_at && (
                      <div className="flex flex-col gap-1 text-sm">
                        <span className="text-muted-foreground">
                          Created At
                        </span>
                        <span className="font-medium">
                          {new Date(user.created_at).toLocaleDateString(
                            "en-US",
                            {
                              month: "long",
                              day: "numeric",
                              year: "numeric",
                            }
                          )}
                        </span>
                      </div>
                    )}
                    {user.updated_at && (
                      <div className="flex flex-col gap-1 text-sm">
                        <span className="text-muted-foreground">
                          Last Updated
                        </span>
                        <span className="font-medium">
                          {new Date(user.updated_at).toLocaleDateString(
                            "en-US",
                            {
                              month: "long",
                              day: "numeric",
                              year: "numeric",
                            }
                          )}
                        </span>
                      </div>
                    )}
                  </div>
                </CardContent>
              </Card>
            </div>
          </TabsContent>

          {/* Metrics Tab */}
          <TabsContent value="metrics" className="space-y-6">
            <div className="grid gap-6 md:grid-cols-2">
              {/* Detailed Risk Metrics */}
              <Card>
                <CardHeader>
                  <CardTitle>Risk Assessment Metrics</CardTitle>
                  <CardDescription>
                    Detailed breakdown of risk assessments performed
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    <div>
                      <div className="mb-2 flex items-center justify-between text-sm">
                        <span>Critical Risks</span>
                        <span className="font-medium">
                          {riskMetrics.criticalRisks}
                        </span>
                      </div>
                      <div className="h-2 w-full rounded-full bg-gray-200">
                        <div
                          className="h-2 rounded-full bg-red-500"
                          style={{
                            width: `${(riskMetrics.criticalRisks / riskMetrics.totalRisks) * 100}%`,
                          }}
                        />
                      </div>
                    </div>
                    <div>
                      <div className="mb-2 flex items-center justify-between text-sm">
                        <span>High Risks</span>
                        <span className="font-medium">
                          {riskMetrics.highRisks}
                        </span>
                      </div>
                      <div className="h-2 w-full rounded-full bg-gray-200">
                        <div
                          className="h-2 rounded-full bg-orange-500"
                          style={{
                            width: `${(riskMetrics.highRisks / riskMetrics.totalRisks) * 100}%`,
                          }}
                        />
                      </div>
                    </div>
                    <div>
                      <div className="mb-2 flex items-center justify-between text-sm">
                        <span>Medium Risks</span>
                        <span className="font-medium">
                          {riskMetrics.mediumRisks}
                        </span>
                      </div>
                      <div className="h-2 w-full rounded-full bg-gray-200">
                        <div
                          className="h-2 rounded-full bg-yellow-500"
                          style={{
                            width: `${(riskMetrics.mediumRisks / riskMetrics.totalRisks) * 100}%`,
                          }}
                        />
                      </div>
                    </div>
                    <div>
                      <div className="mb-2 flex items-center justify-between text-sm">
                        <span>Low Risks</span>
                        <span className="font-medium">
                          {riskMetrics.lowRisks}
                        </span>
                      </div>
                      <div className="h-2 w-full rounded-full bg-gray-200">
                        <div
                          className="h-2 rounded-full bg-green-500"
                          style={{
                            width: `${(riskMetrics.lowRisks / riskMetrics.totalRisks) * 100}%`,
                          }}
                        />
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>

              {/* Detailed Audit Metrics */}
              <Card>
                <CardHeader>
                  <CardTitle>Audit Performance Metrics</CardTitle>
                  <CardDescription>
                    Detailed breakdown of audit activities and findings
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-4">
                    <div>
                      <div className="mb-2 flex items-center justify-between text-sm">
                        <span>Completion Rate</span>
                        <span className="font-medium">
                          {Math.round(
                            (auditMetrics.completedAudits /
                              auditMetrics.totalAudits) *
                              100
                          )}
                          %
                        </span>
                      </div>
                      <div className="h-2 w-full rounded-full bg-gray-200">
                        <div
                          className="h-2 rounded-full bg-green-500"
                          style={{
                            width: `${(auditMetrics.completedAudits / auditMetrics.totalAudits) * 100}%`,
                          }}
                        />
                      </div>
                    </div>
                    <div>
                      <div className="mb-2 flex items-center justify-between text-sm">
                        <span>In Progress Rate</span>
                        <span className="font-medium">
                          {Math.round(
                            (auditMetrics.inProgressAudits /
                              auditMetrics.totalAudits) *
                              100
                          )}
                          %
                        </span>
                      </div>
                      <div className="h-2 w-full rounded-full bg-gray-200">
                        <div
                          className="h-2 rounded-full bg-blue-500"
                          style={{
                            width: `${(auditMetrics.inProgressAudits / auditMetrics.totalAudits) * 100}%`,
                          }}
                        />
                      </div>
                    </div>
                    <Separator />
                    <div className="space-y-2">
                      <div className="flex items-center justify-between text-sm">
                        <span className="text-muted-foreground">
                          Total Findings Created
                        </span>
                        <span className="font-medium">
                          {auditMetrics.findingsCreated}
                        </span>
                      </div>
                      <div className="flex items-center justify-between text-sm">
                        <span className="text-muted-foreground">
                          Open Findings
                        </span>
                        <span className="font-medium text-orange-600">
                          {auditMetrics.openFindings}
                        </span>
                      </div>
                      <div className="flex items-center justify-between text-sm">
                        <span className="text-muted-foreground">
                          Resolution Rate
                        </span>
                        <span className="font-medium">
                          {Math.round(
                            ((auditMetrics.findingsCreated -
                              auditMetrics.openFindings) /
                              auditMetrics.findingsCreated) *
                              100
                          )}
                          %
                        </span>
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          </TabsContent>

          {/* Activity Tab */}
          <TabsContent value="activity" className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle>Recent Activities</CardTitle>
                <CardDescription>
                  Latest actions and events from this user
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  {recentActivities.map((activity, index) => (
                    <div key={activity.id}>
                      <div className="flex items-start gap-4">
                        <div
                          className={`bg-primary/10 text-primary flex h-8 w-8 items-center justify-center rounded-full`}
                        >
                          {getActivityIcon(activity.type)}
                        </div>
                        <div className="flex-1 space-y-1">
                          <div className="flex items-start justify-between">
                            <div>
                              <p className="text-sm font-medium">
                                {activity.action}
                              </p>
                              <p className="text-muted-foreground text-sm">
                                {activity.details}
                              </p>
                            </div>
                            <span className="text-muted-foreground text-xs">
                              {formatTimestamp(activity.timestamp)}
                            </span>
                          </div>
                        </div>
                      </div>
                      {index < recentActivities.length - 1 && (
                        <Separator className="my-4" />
                      )}
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>

        {/* Edit User Modal */}
        <CreateUserForm
          showTrigger={false}
          role="REQUESTER"
          isOpenModal={isEditingUser}
          user={user}
          setIsOpenModal={(open) => {
            if (!open) {
              setIsEditingUser(false);
              router.refresh();
            }
          }}
        />
      </div>
    </>
  );
}
