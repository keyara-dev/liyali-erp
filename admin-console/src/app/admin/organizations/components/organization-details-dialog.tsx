"use client";

import { useState, useEffect } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Separator } from "@/components/ui/separator";
import {
  Building2,
  Users,
  Mail,
  Phone,
  Calendar,
  Activity,
  CreditCard,
  Clock,
  MapPin,
  Monitor,
  AlertTriangle,
  CheckCircle,
  XCircle,
  Globe,
  Settings,
} from "lucide-react";
import { toast } from "sonner";
import {
  getOrganizationUsers,
  getOrganizationActivity,
  getOrganizationSubscription,
  resetOrganizationTrial,
  type Organization,
  type OrganizationUser,
  type OrganizationActivity,
  type TrialResetRequest,
} from "@/app/_actions/organizations";
import { ChangeTierDialog } from "./change-tier-dialog";
import { OverrideLimitsDialog } from "./override-limits-dialog";

interface OrganizationDetailsDialogProps {
  organization: Organization;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onOrganizationUpdated: () => void;
}

export function OrganizationDetailsDialog({
  organization,
  open,
  onOpenChange,
  onOrganizationUpdated,
}: OrganizationDetailsDialogProps) {
  const [activeTab, setActiveTab] = useState("overview");
  const [users, setUsers] = useState<OrganizationUser[]>([]);
  const [activities, setActivities] = useState<OrganizationActivity[]>([]);
  const [subscription, setSubscription] = useState<any>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [showChangeTier, setShowChangeTier] = useState(false);
  const [showOverrideLimits, setShowOverrideLimits] = useState(false);

  useEffect(() => {
    if (open && organization) {
      loadOrganizationDetails();
    }
  }, [open, organization]);

  const loadOrganizationDetails = async () => {
    setIsLoading(true);
    try {
      // Load organization users
      const usersResult = await getOrganizationUsers(organization.id, 1, 20);
      if (usersResult.success && usersResult.data) {
        setUsers(usersResult.data.users || []);
      }

      // Load organization activity
      const activityResult = await getOrganizationActivity(organization.id, 1, 20);
      if (activityResult.success && activityResult.data) {
        setActivities(activityResult.data.activities || []);
      }

      // Load subscription details
      const subscriptionResult = await getOrganizationSubscription(organization.id);
      if (subscriptionResult.success && subscriptionResult.data) {
        setSubscription(subscriptionResult.data);
      }
    } catch (error) {
      console.error("Failed to load organization details:", error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleTrialReset = async (request: TrialResetRequest) => {
    try {
      const result = await resetOrganizationTrial(organization.id, request);
      if (result.success) {
        toast.success("Trial reset successfully");
        onOrganizationUpdated();
        loadOrganizationDetails();
      } else {
        toast.error(result.message || "Failed to reset trial");
      }
    } catch (error) {
      toast.error("Failed to reset trial");
    }
  };

  const getStatusBadge = (status: string) => {
    if (status === "suspended") {
      return <Badge variant="destructive">Suspended</Badge>;
    }
    if (status === "pending") {
      return <Badge variant="secondary">Pending</Badge>;
    }
    return <Badge variant="default">Active</Badge>;
  };

  const getTrialStatusBadge = (trialStatus: string) => {
    switch (trialStatus) {
      case "trial":
        return <Badge variant="secondary">Trial</Badge>;
      case "subscribed":
        return <Badge variant="default">Subscribed</Badge>;
      case "expired":
        return <Badge variant="destructive">Expired</Badge>;
      default:
        return <Badge variant="outline">{trialStatus}</Badge>;
    }
  };

  const getTierBadge = (tier: string) => {
    switch (tier) {
      case "enterprise":
        return <Badge variant="default">Enterprise</Badge>;
      case "professional":
        return <Badge className="bg-blue-100 text-blue-800">Professional</Badge>;
      case "basic":
        return <Badge variant="secondary">Basic</Badge>;
      default:
        return <Badge variant="outline">{tier}</Badge>;
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Building2 className="h-5 w-5" />
            Organization Details: {organization.name}
          </DialogTitle>
          <DialogDescription>
            Comprehensive view of organization information, users, and activity
          </DialogDescription>
        </DialogHeader>

        <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
          <TabsList className="grid w-full grid-cols-4">
            <TabsTrigger value="overview">Overview</TabsTrigger>
            <TabsTrigger value="users">Users</TabsTrigger>
            <TabsTrigger value="subscription">Subscription</TabsTrigger>
            <TabsTrigger value="activity">Activity</TabsTrigger>
          </TabsList>

          <TabsContent value="overview" className="space-y-4">
            <div className="grid gap-4 md:grid-cols-2">
              <Card>
                <CardHeader>
                  <CardTitle className="text-lg flex items-center gap-2">
                    <Building2 className="h-4 w-4" />
                    Basic Information
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium">Status:</span>
                    {getStatusBadge(organization.status)}
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium">Trial Status:</span>
                    {getTrialStatusBadge(organization.trial_status)}
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium">Subscription:</span>
                    {getTierBadge(organization.subscription_tier)}
                  </div>
                  <Separator />
                  <div className="space-y-2">
                    <div className="flex items-center gap-2">
                      <Globe className="h-4 w-4 text-muted-foreground" />
                      <span className="text-sm">{organization.domain}</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <Users className="h-4 w-4 text-muted-foreground" />
                      <span className="text-sm">{organization.user_count} users</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <Calendar className="h-4 w-4 text-muted-foreground" />
                      <span className="text-sm">
                        Created {new Date(organization.created_at).toLocaleDateString()}
                      </span>
                    </div>
                    {organization.trial_end_date && (
                      <div className="flex items-center gap-2">
                        <Clock className="h-4 w-4 text-muted-foreground" />
                        <span className="text-sm">
                          Trial ends {new Date(organization.trial_end_date).toLocaleDateString()}
                          {organization.days_remaining !== undefined && (
                            <span className="ml-1 text-muted-foreground">
                              ({organization.days_remaining} days remaining)
                            </span>
                          )}
                        </span>
                      </div>
                    )}
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="text-lg flex items-center gap-2">
                    <Settings className="h-4 w-4" />
                    Settings & Features
                  </CardTitle>
                </CardHeader>
                <CardContent className="space-y-3">
                  {organization.settings?.max_users && (
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-medium">Max Users:</span>
                      <span className="text-sm">{organization.settings.max_users}</span>
                    </div>
                  )}
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium">Custom Branding:</span>
                    {organization.settings?.custom_branding ? (
                      <CheckCircle className="h-4 w-4 text-green-500" />
                    ) : (
                      <XCircle className="h-4 w-4 text-red-500" />
                    )}
                  </div>
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium">API Access:</span>
                    {organization.settings?.api_access ? (
                      <CheckCircle className="h-4 w-4 text-green-500" />
                    ) : (
                      <XCircle className="h-4 w-4 text-red-500" />
                    )}
                  </div>
                  {organization.settings?.features_enabled && (
                    <div className="space-y-2">
                      <span className="text-sm font-medium">Enabled Features:</span>
                      <div className="flex flex-wrap gap-1">
                        {organization.settings.features_enabled.map((feature) => (
                          <Badge key={feature} variant="outline" className="text-xs">
                            {feature}
                          </Badge>
                        ))}
                      </div>
                    </div>
                  )}
                </CardContent>
              </Card>
            </div>

            {organization.contact_info && (
              <Card>
                <CardHeader>
                  <CardTitle className="text-lg">Contact Information</CardTitle>
                </CardHeader>
                <CardContent className="grid gap-4 md:grid-cols-2">
                  {organization.contact_info.admin_name && (
                    <div>
                      <span className="text-sm font-medium">Admin Name:</span>
                      <p className="text-sm text-muted-foreground">
                        {organization.contact_info.admin_name}
                      </p>
                    </div>
                  )}
                  {organization.contact_info.admin_email && (
                    <div>
                      <span className="text-sm font-medium">Admin Email:</span>
                      <p className="text-sm text-muted-foreground">
                        {organization.contact_info.admin_email}
                      </p>
                    </div>
                  )}
                  {organization.contact_info.phone && (
                    <div>
                      <span className="text-sm font-medium">Phone:</span>
                      <p className="text-sm text-muted-foreground">
                        {organization.contact_info.phone}
                      </p>
                    </div>
                  )}
                  {organization.contact_info.address && (
                    <div>
                      <span className="text-sm font-medium">Address:</span>
                      <p className="text-sm text-muted-foreground">
                        {organization.contact_info.address}
                      </p>
                    </div>
                  )}
                </CardContent>
              </Card>
            )}
          </TabsContent>

          <TabsContent value="users" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Users className="h-4 w-4" />
                  Organization Users ({users.length})
                </CardTitle>
              </CardHeader>
              <CardContent>
                {isLoading ? (
                  <div className="text-center py-8">Loading users...</div>
                ) : users.length === 0 ? (
                  <div className="text-center py-8 text-muted-foreground">
                    No users found in this organization
                  </div>
                ) : (
                  <div className="space-y-4">
                    {users.map((user) => (
                      <div
                        key={user.id}
                        className="flex items-center justify-between rounded-lg border p-4"
                      >
                        <div className="space-y-1">
                          <div className="flex items-center gap-2">
                            <h4 className="font-medium">{user.name}</h4>
                            {user.is_admin && (
                              <Badge variant="default" className="text-xs">
                                Admin
                              </Badge>
                            )}
                          </div>
                          <div className="flex items-center gap-4 text-sm text-muted-foreground">
                            <span className="flex items-center gap-1">
                              <Mail className="h-3 w-3" />
                              {user.email}
                            </span>
                            <span>Role: {user.role}</span>
                            <span>
                              Joined: {new Date(user.joined_at).toLocaleDateString()}
                            </span>
                          </div>
                          {user.last_login && (
                            <div className="text-xs text-muted-foreground">
                              Last login: {new Date(user.last_login).toLocaleString()}
                            </div>
                          )}
                        </div>
                        <div className="flex items-center gap-2">
                          <Badge
                            variant={user.status === "active" ? "default" : "secondary"}
                          >
                            {user.status}
                          </Badge>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="subscription" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <CreditCard className="h-4 w-4" />
                  Subscription Details
                </CardTitle>
              </CardHeader>
              <CardContent>
                {subscription ? (
                  <div className="space-y-4">
                    <div className="grid gap-4 md:grid-cols-2">
                      <div>
                        <span className="text-sm font-medium">Plan:</span>
                        <p className="text-sm text-muted-foreground">
                          {subscription.plan_name || organization.subscription_tier}
                        </p>
                      </div>
                      {subscription.monthly_cost && (
                        <div>
                          <span className="text-sm font-medium">Monthly Cost:</span>
                          <p className="text-sm text-muted-foreground">
                            ${subscription.monthly_cost}
                          </p>
                        </div>
                      )}
                      {subscription.next_billing_date && (
                        <div>
                          <span className="text-sm font-medium">Next Billing:</span>
                          <p className="text-sm text-muted-foreground">
                            {new Date(subscription.next_billing_date).toLocaleDateString()}
                          </p>
                        </div>
                      )}
                      {subscription.payment_status && (
                        <div>
                          <span className="text-sm font-medium">Payment Status:</span>
                          <Badge
                            variant={
                              subscription.payment_status === "active"
                                ? "default"
                                : "destructive"
                            }
                          >
                            {subscription.payment_status}
                          </Badge>
                        </div>
                      )}
                    </div>
                  </div>
                ) : (
                  <div className="text-center py-8 text-muted-foreground">
                    No subscription details available
                  </div>
                )}
              </CardContent>
            </Card>

            {/* Subscription Action Buttons */}
            <Card>
              <CardHeader>
                <CardTitle className="text-sm">Subscription Actions</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex flex-wrap gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setShowChangeTier(true)}
                  >
                    <Activity className="mr-2 h-4 w-4" />
                    Change Tier
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setShowOverrideLimits(true)}
                  >
                    <Settings className="mr-2 h-4 w-4" />
                    Override Limits
                  </Button>
                </div>
              </CardContent>
            </Card>

            {/* Change Tier Dialog */}
            <ChangeTierDialog
              organization={organization}
              open={showChangeTier}
              onOpenChange={setShowChangeTier}
              onSuccess={() => {
                onOrganizationUpdated();
                loadOrganizationDetails();
              }}
            />

            {/* Override Limits Dialog */}
            <OverrideLimitsDialog
              organization={organization}
              open={showOverrideLimits}
              onOpenChange={setShowOverrideLimits}
              onSuccess={() => {
                onOrganizationUpdated();
                loadOrganizationDetails();
              }}
            />
          </TabsContent>

          <TabsContent value="activity" className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <Activity className="h-4 w-4" />
                  Recent Activity ({activities.length})
                </CardTitle>
              </CardHeader>
              <CardContent>
                {isLoading ? (
                  <div className="text-center py-8">Loading activity...</div>
                ) : activities.length === 0 ? (
                  <div className="text-center py-8 text-muted-foreground">
                    No recent activity found
                  </div>
                ) : (
                  <div className="space-y-4">
                    {activities.map((activity) => (
                      <div
                        key={activity.id}
                        className="flex items-start gap-3 rounded-lg border p-4"
                      >
                        <div className="flex h-8 w-8 items-center justify-center rounded-full bg-primary/10">
                          <Activity className="h-4 w-4 text-primary" />
                        </div>
                        <div className="flex-1 space-y-1">
                          <div className="flex items-center justify-between">
                            <h4 className="font-medium">{activity.action}</h4>
                            <span className="text-xs text-muted-foreground">
                              {new Date(activity.timestamp).toLocaleString()}
                            </span>
                          </div>
                          <p className="text-sm text-muted-foreground">
                            {activity.description}
                          </p>
                          <div className="flex items-center gap-4 text-xs text-muted-foreground">
                            {activity.user_name && (
                              <span className="flex items-center gap-1">
                                <Users className="h-3 w-3" />
                                {activity.user_name}
                              </span>
                            )}
                            {activity.ip_address && (
                              <span className="flex items-center gap-1">
                                <MapPin className="h-3 w-3" />
                                {activity.ip_address}
                              </span>
                            )}
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  );
}