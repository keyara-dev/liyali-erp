'use client'

import { useRouter } from 'next/navigation'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import {
  PlusCircle,
  FileText,
  Search,
  Settings,
} from 'lucide-react'

interface QuickActionsProps {
  userRole: string
}

export function QuickActions({ userRole }: QuickActionsProps) {
  const router = useRouter()

  const actions = [
    {
      title: 'Create Requisition',
      description: 'Start a new requisition',
      icon: PlusCircle,
      href: '/workflows/requisitions/create',
      color: 'text-primary',
      bgColor: 'bg-primary/10',
      visible: true,
    },
    {
      title: 'View Search',
      description: 'Search all transactions',
      icon: Search,
      href: '/workflows/search',
      color: 'text-secondary',
      bgColor: 'bg-secondary/10',
      visible: true,
    },
    {
      title: 'My Documents',
      description: 'View your requisitions',
      icon: FileText,
      href: '/workflows/requisitions',
      color: 'text-accent',
      bgColor: 'bg-accent/10',
      visible: true,
    },
    {
      title: 'Settings',
      description: 'Configure preferences',
      icon: Settings,
      href: '/settings',
      color: 'text-muted-foreground',
      bgColor: 'bg-muted/10',
      visible: userRole === 'ADMIN',
    },
  ]

  const visibleActions = actions.filter((action) => action.visible)

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Quick Actions</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {visibleActions.map((action, index) => {
            const Icon = action.icon
            return (
              <Button
                key={index}
                variant="outline"
                className="w-full justify-start"
                onClick={() => router.push(action.href)}
              >
                <div className={`rounded-lg p-2 mr-3 ${action.bgColor}`}>
                  <Icon className={`h-4 w-4 ${action.color}`} />
                </div>
                <div className="text-left">
                  <p className="text-sm font-medium">{action.title}</p>
                  <p className="text-xs text-muted-foreground">{action.description}</p>
                </div>
              </Button>
            )
          })}
        </div>
      </CardContent>
    </Card>
  )
}
