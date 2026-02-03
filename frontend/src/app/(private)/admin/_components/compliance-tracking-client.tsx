'use client'

import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { CheckCircle2, AlertCircle, Clock, FileText } from 'lucide-react'
import { useComplianceRequirements } from '@/hooks/use-compliance-queries'

interface ComplianceItem {
  id: string;
  name: string;
  requirement: string;
  status: 'compliant' | 'non-compliant' | 'pending';
  dueDate: string;
  responsible: string;
  completionDate?: string;
  evidence: string[];
}

interface ComplianceTrackingClientProps {
  userId: string
  userRole: string
}

const STATUS_COLORS: Record<string, { bg: string; text: string; icon: React.ReactNode }> = {
  compliant: {
    bg: 'bg-secondary/10',
    text: 'text-secondary',
    icon: <CheckCircle2 className="h-5 w-5" />,
  },
  'non-compliant': {
    bg: 'bg-destructive/10',
    text: 'text-destructive',
    icon: <AlertCircle className="h-5 w-5" />,
  },
  pending: {
    bg: 'bg-accent/10',
    text: 'text-accent',
    icon: <Clock className="h-5 w-5" />,
  },
}

export function ComplianceTrackingClient({
  userId,
  userRole,
}: ComplianceTrackingClientProps) {
  const [selectedTab, setSelectedTab] = useState('overview')

  // Fetch compliance requirements
  const { data: complianceData, isLoading } = useComplianceRequirements()

  const requirements = (complianceData?.requirements || []) as ComplianceItem[]
  const compliant = requirements.filter((r) => r.status === 'compliant').length
  const nonCompliant = requirements.filter((r) => r.status === 'non-compliant').length
  const pending = requirements.filter((r) => r.status === 'pending').length
  const complianceScore = requirements.length > 0 ? Math.round((compliant / requirements.length) * 100) : 0

  // Suppress unused variable warnings for now
  void userId;
  void userRole;

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-xl font-bold tracking-tight lg:text-2xl">Compliance Tracking</h1>
          <p className="text-sm text-muted-foreground">
            Monitor regulatory compliance and audit requirements
          </p>
        </div>
        <div className="text-center py-12">
          <p className="text-muted-foreground">Loading compliance data...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-xl font-bold tracking-tight lg:text-2xl">Compliance Tracking</h1>
        <p className="text-sm text-muted-foreground">
          Monitor regulatory compliance and audit requirements
        </p>
      </div>

      {/* Compliance Score Card */}
      <Card className="border-primary/20 bg-gradient-to-br from-primary/5 to-transparent">
        <CardContent className="pt-6">
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">Overall Compliance Score</p>
                <p className="text-4xl font-bold text-primary">{complianceScore}%</p>
              </div>
              <div className="text-right">
                <p className="text-2xl font-bold text-secondary">{compliant}</p>
                <p className="text-xs text-muted-foreground">of {requirements.length} requirements</p>
              </div>
            </div>
            <Progress value={complianceScore} className="h-2" />
          </div>
        </CardContent>
      </Card>

      {/* Status Overview */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Compliant
            </CardTitle>
            <CheckCircle2 className="h-5 w-5 text-secondary" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-secondary">{compliant}</div>
            <p className="text-xs text-muted-foreground mt-1">
              {requirements.length > 0 ? Math.round((compliant / requirements.length) * 100) : 0}% complete
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Pending
            </CardTitle>
            <Clock className="h-5 w-5 text-accent" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-accent">{pending}</div>
            <p className="text-xs text-muted-foreground mt-1">
              In progress
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Non-Compliant
            </CardTitle>
            <AlertCircle className="h-5 w-5 text-destructive" />
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-destructive">{nonCompliant}</div>
            <p className="text-xs text-muted-foreground mt-1">
              Requires attention
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Tabs */}
      <Tabs value={selectedTab} onValueChange={setSelectedTab} className="w-full">
        <TabsList className="grid w-full grid-cols-3 lg:w-auto">
          <TabsTrigger value="overview">All Requirements</TabsTrigger>
          <TabsTrigger value="compliant">Compliant</TabsTrigger>
          <TabsTrigger value="issues">Issues</TabsTrigger>
        </TabsList>

        {/* All Requirements Tab */}
        <TabsContent value="overview" className="space-y-4">
          {requirements.length === 0 ? (
            <Card>
              <CardContent className="pt-6 text-center">
                <p className="text-muted-foreground">No compliance requirements found</p>
              </CardContent>
            </Card>
          ) : (
            requirements.map((item) => (
            <Card key={item.id}>
              <CardContent className="pt-6">
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-2">
                      <h3 className="font-semibold">{item.name}</h3>
                      <Badge variant={item.status === 'compliant' ? 'default' : item.status === 'non-compliant' ? 'destructive' : 'outline'}>
                        {item.status}
                      </Badge>
                    </div>
                    <p className="text-sm text-muted-foreground mb-3">{item.requirement}</p>

                    <div className="grid grid-cols-2 gap-3 text-sm">
                      <div>
                        <p className="text-xs text-muted-foreground">Due Date</p>
                        <p className="font-medium">{item.dueDate}</p>
                      </div>
                      <div>
                        <p className="text-xs text-muted-foreground">Responsible</p>
                        <p className="font-medium">{item.responsible}</p>
                      </div>
                      {item.completionDate && (
                        <div className="col-span-2">
                          <p className="text-xs text-muted-foreground">Completed</p>
                          <p className="font-medium text-secondary">{item.completionDate}</p>
                        </div>
                      )}
                    </div>

                    {item.evidence.length > 0 && (
                      <div className="mt-3">
                        <p className="text-xs font-medium text-muted-foreground mb-1">Evidence</p>
                        <div className="flex flex-wrap gap-1">
                          {item.evidence.map((doc: string) => (
                            <Badge key={doc} variant="outline" className="text-xs gap-1">
                              <FileText className="h-3 w-3" />
                              {doc}
                            </Badge>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>
                  <div className={`flex-shrink-0 p-2 rounded-lg ${STATUS_COLORS[item.status].bg}`}>
                    <div className={STATUS_COLORS[item.status].text}>
                      {STATUS_COLORS[item.status].icon}
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
            ))
          )}
        </TabsContent>

        {/* Compliant Tab */}
        <TabsContent value="compliant" className="space-y-4">
          {requirements.filter((r) => r.status === 'compliant').length === 0 ? (
            <Card>
              <CardContent className="pt-6 text-center">
                <p className="text-muted-foreground">No compliant items yet</p>
              </CardContent>
            </Card>
          ) : (
            requirements.filter((r) => r.status === 'compliant').map((item) => (
            <Card key={item.id}>
              <CardContent className="pt-6">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="font-semibold">{item.name}</h3>
                    <p className="text-sm text-muted-foreground mt-1">{item.requirement}</p>
                  </div>
                  <CheckCircle2 className="h-8 w-8 text-secondary flex-shrink-0" />
                </div>
              </CardContent>
            </Card>
            ))
          )}
        </TabsContent>

        {/* Issues Tab */}
        <TabsContent value="issues" className="space-y-4">
          {requirements.filter((r) => r.status !== 'compliant').length === 0 ? (
            <Card>
              <CardContent className="pt-6 text-center">
                <p className="text-muted-foreground">No issues to address</p>
              </CardContent>
            </Card>
          ) : (
            requirements.filter((r) => r.status !== 'compliant').map((item) => (
            <Card key={item.id} className={item.status === 'non-compliant' ? 'border-destructive/50' : ''}>
              <CardContent className="pt-6">
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-2">
                      <h3 className="font-semibold">{item.name}</h3>
                      <Badge variant={item.status === 'non-compliant' ? 'destructive' : 'outline'}>
                        {item.status}
                      </Badge>
                    </div>
                    <p className="text-sm text-muted-foreground">{item.requirement}</p>
                    <p className="text-sm font-medium mt-2">Due: {item.dueDate}</p>
                  </div>
                  <Button variant="outline" size="sm">
                    Update Status
                  </Button>
                </div>
              </CardContent>
            </Card>
            ))
          )}
        </TabsContent>
      </Tabs>
    </div>
  )
}
