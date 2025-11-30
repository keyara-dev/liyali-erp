'use client'

import { useState } from 'react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { User } from '@/types/auth'
import { AccountSettings } from './account-settings'
import { ChangePassword } from './change-password'
import { GeneralSettings } from './general-settings'
import { SessionsManagement } from './sessions-management'
import { Users, Lock, Settings as SettingsIcon, Globe } from 'lucide-react'

interface SettingsClientProps {
  user: User | null
}

export function SettingsClient({ user }: SettingsClientProps) {
  const [profileUser, setProfileUser] = useState<User | null>(user)

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Settings</h1>
        <p className="text-muted-foreground mt-2">
          Manage your account, security, and preferences
        </p>
      </div>

      {/* Tabs */}
      <Tabs defaultValue="account" className="w-full">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="account" className="flex items-center gap-2">
            <Users className="h-4 w-4" />
            <span className="hidden sm:inline">Account</span>
          </TabsTrigger>
          <TabsTrigger value="password" className="flex items-center gap-2">
            <Lock className="h-4 w-4" />
            <span className="hidden sm:inline">Password</span>
          </TabsTrigger>
          <TabsTrigger value="general" className="flex items-center gap-2">
            <SettingsIcon className="h-4 w-4" />
            <span className="hidden sm:inline">General</span>
          </TabsTrigger>
          <TabsTrigger value="sessions" className="flex items-center gap-2">
            <Globe className="h-4 w-4" />
            <span className="hidden sm:inline">Sessions</span>
          </TabsTrigger>
        </TabsList>

        {/* Account Tab */}
        <TabsContent value="account" className="space-y-4">
          <AccountSettings
            user={profileUser}
            onProfileUpdate={(updatedUser) => setProfileUser(updatedUser)}
          />
        </TabsContent>

        {/* Password Tab */}
        <TabsContent value="password" className="space-y-4">
          <ChangePassword />
        </TabsContent>

        {/* General Tab */}
        <TabsContent value="general" className="space-y-4">
          <GeneralSettings />
        </TabsContent>

        {/* Sessions Tab */}
        <TabsContent value="sessions" className="space-y-4">
          <SessionsManagement />
        </TabsContent>
      </Tabs>
    </div>
  )
}
