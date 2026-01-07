'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { useCreateUserMutation } from '@/hooks/use-users-mutations';
import { useCreateOrganizationMutation } from '@/hooks/use-organization-mutations';
import { useSaveRequisition } from '@/hooks/use-requisition-queries';
import { useOfflineStatus, useQueueStats } from '@/hooks/use-offline-queue-processor';
import { Badge } from '@/components/ui/badge';
import { WifiOff, Wifi, Clock, CheckCircle, XCircle, Loader2, RefreshCw } from 'lucide-react';
import { clearQueue } from '@/lib/offline-queue';

/**
 * Demo component to test offline functionality
 * This can be added to any page to test offline operations
 */
export function OfflineDemo() {
  const [userName, setUserName] = useState('');
  const [userEmail, setUserEmail] = useState('');
  const [orgName, setOrgName] = useState('');
  const [reqTitle, setReqTitle] = useState('');
  const [reqDescription, setReqDescription] = useState('');
  
  const isOffline = useOfflineStatus();
  const stats = useQueueStats();
  const createUserMutation = useCreateUserMutation();
  const createOrgMutation = useCreateOrganizationMutation();
  const saveRequisitionMutation = useSaveRequisition();

  const handleCreateUser = async () => {
    if (!userName || !userEmail) return;
    
    try {
      await createUserMutation.createUser({
        name: userName,
        email: userEmail,
        roles: ['user'],
        department: 'IT',
        position: 'Developer'
      });
      setUserName('');
      setUserEmail('');
    } catch (error) {
      console.error('Failed to create user:', error);
    }
  };

  const handleCreateOrg = async () => {
    if (!orgName) return;
    
    try {
      await createOrgMutation.createOrganization({
        name: orgName,
        description: 'Test organization created offline',
        address: '123 Test Street',
        phone: '+1-555-0123',
        email: `contact@${orgName.toLowerCase().replace(/\s+/g, '')}.com`
      });
      setOrgName('');
    } catch (error) {
      console.error('Failed to create organization:', error);
    }
  };

  const handleCreateRequisition = async () => {
    if (!reqTitle) return;
    
    try {
      await saveRequisitionMutation.mutateAsync({
        title: reqTitle,
        description: reqDescription || 'Test requisition created offline',
        department: 'IT',
        priority: 'MEDIUM',
        items: [
          {
            itemNumber: 1,
            description: 'Test Item for offline demo',
            category: 'Office Supplies',
            quantity: 1,
            unitPrice: 100,
            amount: 100,              // Required property
            unit: 'pcs',
            totalPrice: 100,          // Alias for amount
            notes: 'Test item created during offline functionality demo'
          }
        ],
        totalAmount: 100,
        currency: 'USD',
        justification: 'Testing offline functionality',
        expectedDeliveryDate: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
        budgetCode: 'TEST-2025-001',
        createdBy: 'test-user-id',
        createdByName: 'Test User',
        createdByRole: 'requester'
      });
      setReqTitle('');
      setReqDescription('');
    } catch (error) {
      console.error('Failed to create requisition:', error);
    }
  };

  const handleClearQueue = async () => {
    try {
      await clearQueue();
      window.location.reload(); // Refresh to update stats
    } catch (error) {
      console.error('Failed to clear queue:', error);
    }
  };

  return (
    <div className="space-y-6">
      {/* Network Status */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            {isOffline ? (
              <WifiOff className="h-5 w-5 text-orange-600" />
            ) : (
              <Wifi className="h-5 w-5 text-green-600" />
            )}
            Network Status
          </CardTitle>
          <CardDescription>
            Current connection status and offline queue information
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center gap-2">
            <Badge variant={isOffline ? "destructive" : "default"}>
              {isOffline ? "Offline" : "Online"}
            </Badge>
            {isOffline && (
              <span className="text-sm text-muted-foreground">
                Operations will be queued and synced when connection is restored
              </span>
            )}
          </div>

          {stats.total > 0 && (
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div className="flex items-center gap-2">
                <Clock className="h-4 w-4 text-orange-500" />
                <span className="text-sm">Pending: {stats.pending}</span>
              </div>
              <div className="flex items-center gap-2">
                <CheckCircle className="h-4 w-4 text-green-500" />
                <span className="text-sm">Completed: {stats.completed}</span>
              </div>
              <div className="flex items-center gap-2">
                <XCircle className="h-4 w-4 text-red-500" />
                <span className="text-sm">Failed: {stats.failed}</span>
              </div>
              <div className="text-sm font-medium">
                Total: {stats.total}
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Test Operations */}
      <div className="grid md:grid-cols-3 gap-6">
        {/* Create User Test */}
        <Card>
          <CardHeader>
            <CardTitle>Test User Creation</CardTitle>
            <CardDescription>
              Create a user to test offline functionality
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="userName">Name</Label>
              <Input
                id="userName"
                value={userName}
                onChange={(e) => setUserName(e.target.value)}
                placeholder="Enter user name"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="userEmail">Email</Label>
              <Input
                id="userEmail"
                type="email"
                value={userEmail}
                onChange={(e) => setUserEmail(e.target.value)}
                placeholder="Enter user email"
              />
            </div>
            <Button 
              onClick={handleCreateUser}
              disabled={!userName || !userEmail || createUserMutation.isPending}
              className="w-full"
            >
              {createUserMutation.isPending ? 'Creating...' : 'Create User'}
            </Button>
          </CardContent>
        </Card>

        {/* Create Organization Test */}
        <Card>
          <CardHeader>
            <CardTitle>Test Organization Creation</CardTitle>
            <CardDescription>
              Create an organization to test offline functionality
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="orgName">Organization Name</Label>
              <Input
                id="orgName"
                value={orgName}
                onChange={(e) => setOrgName(e.target.value)}
                placeholder="Enter organization name"
              />
            </div>
            <Button 
              onClick={handleCreateOrg}
              disabled={!orgName || createOrgMutation.isPending}
              className="w-full"
            >
              {createOrgMutation.isPending ? 'Creating...' : 'Create Organization'}
            </Button>
          </CardContent>
        </Card>

        {/* Create Requisition Test */}
        <Card>
          <CardHeader>
            <CardTitle>Test Requisition Creation</CardTitle>
            <CardDescription>
              Create a requisition to test offline functionality
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="reqTitle">Title</Label>
              <Input
                id="reqTitle"
                value={reqTitle}
                onChange={(e) => setReqTitle(e.target.value)}
                placeholder="Enter requisition title"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="reqDescription">Description</Label>
              <Textarea
                id="reqDescription"
                value={reqDescription}
                onChange={(e) => setReqDescription(e.target.value)}
                placeholder="Enter description (optional)"
                rows={3}
              />
            </div>
            <Button 
              onClick={handleCreateRequisition}
              disabled={!reqTitle || saveRequisitionMutation.isPending}
              className="w-full"
            >
              {saveRequisitionMutation.isPending ? 'Creating...' : 'Create Requisition'}
            </Button>
          </CardContent>
        </Card>
      </div>

      {/* Queue Management */}
      {stats.total > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center justify-between">
              <span>Queue Management</span>
              <Button
                variant="outline"
                size="sm"
                onClick={handleClearQueue}
              >
                <RefreshCw className="h-4 w-4 mr-2" />
                Clear Queue
              </Button>
            </CardTitle>
            <CardDescription>
              Manage offline operations queue
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="text-sm text-muted-foreground">
              Use "Clear Queue" to remove all pending operations. This is useful for testing or if you want to start fresh.
            </div>
          </CardContent>
        </Card>
      )}

      {/* Instructions */}
      <Card>
        <CardHeader>
          <CardTitle>How to Test Offline Functionality</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3 text-sm text-muted-foreground">
          <div className="grid md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <p><strong>1. Go offline:</strong> Disconnect your internet or use browser dev tools to simulate offline mode</p>
              <p><strong>2. Perform operations:</strong> Try creating users, organizations, or requisitions while offline</p>
              <p><strong>3. Check queue:</strong> Operations will be queued and show in the network status above</p>
            </div>
            <div className="space-y-2">
              <p><strong>4. Go back online:</strong> Reconnect internet to see automatic sync in action</p>
              <p><strong>5. Monitor sync:</strong> Watch the offline indicator in the header for sync progress</p>
              <p><strong>6. Clear queue:</strong> Use the "Clear Queue" button to reset for testing</p>
            </div>
          </div>
          <div className="border-t pt-3 mt-4">
            <p className="font-medium text-foreground">Available Test Operations:</p>
            <ul className="list-disc list-inside space-y-1 mt-2">
              <li><strong>Users:</strong> Create, update, and deactivate user accounts</li>
              <li><strong>Organizations:</strong> Create and update organization profiles</li>
              <li><strong>Requisitions:</strong> Create, update, and submit requisitions for approval</li>
            </ul>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}