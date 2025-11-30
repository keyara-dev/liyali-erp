'use client'

import { useEffect, useState } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { getUserSessions, revokeSession } from '@/app/_actions/settings'
import { AlertCircle, CheckCircle, Loader2, Globe, Smartphone, LogOut } from 'lucide-react'

interface Session {
  id: string
  device: string
  location: string
  ipAddress: string
  lastActive: string
  createdAt: string
  isCurrent: boolean
}

export function SessionsManagement() {
  const [isLoading, setIsLoading] = useState(true)
  const [revoking, setRevoking] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  const [sessions, setSessions] = useState<Session[]>([])

  useEffect(() => {
    loadSessions()
  }, [])

  const loadSessions = async () => {
    try {
      setIsLoading(true)
      const result = await getUserSessions()
      if (result.success && result.data) {
        setSessions(result.data)
      } else {
        setError(result.message || 'Failed to load sessions')
      }
    } catch (err) {
      setError('An error occurred while loading sessions')
      console.error(err)
    } finally {
      setIsLoading(false)
    }
  }

  const handleRevokeSession = async (sessionId: string) => {
    if (confirm('Are you sure you want to revoke this session?')) {
      try {
        setRevoking(sessionId)
        const result = await revokeSession(sessionId)
        if (result.success) {
          setSessions((prev) => prev.filter((s) => s.id !== sessionId))
          setSuccess('Session revoked successfully')
        } else {
          setError(result.message || 'Failed to revoke session')
        }
      } catch (err) {
        setError('An error occurred while revoking session')
        console.error(err)
      } finally {
        setRevoking(null)
      }
    }
  }

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Active Sessions</CardTitle>
          <CardDescription>
            Manage your active login sessions and security
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center py-8">
            <Loader2 className="h-4 w-4 animate-spin" />
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Active Sessions</CardTitle>
        <CardDescription>
          View and manage your active login sessions. Revoke any sessions you don't recognize.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {error && (
          <div className="flex items-center gap-2 p-3 rounded-lg bg-red-50 text-red-700 border border-red-200">
            <AlertCircle className="h-4 w-4 flex-shrink-0" />
            <p className="text-sm">{error}</p>
          </div>
        )}

        {success && (
          <div className="flex items-center gap-2 p-3 rounded-lg bg-green-50 text-green-700 border border-green-200">
            <CheckCircle className="h-4 w-4 flex-shrink-0" />
            <p className="text-sm">{success}</p>
          </div>
        )}

        {sessions.length === 0 ? (
          <p className="text-center text-muted-foreground py-8">No active sessions</p>
        ) : (
          <div className="space-y-3">
            {sessions.map((session) => (
              <div
                key={session.id}
                className="flex items-start justify-between p-4 border rounded-lg hover:bg-muted/50 transition-colors"
              >
                <div className="flex items-start gap-3 flex-1">
                  <div className="mt-1">
                    {session.device.toLowerCase().includes('mobile') ? (
                      <Smartphone className="h-5 w-5 text-muted-foreground" />
                    ) : (
                      <Globe className="h-5 w-5 text-muted-foreground" />
                    )}
                  </div>
                  <div className="space-y-1 flex-1">
                    <div className="flex items-center gap-2">
                      <p className="font-medium text-sm">{session.device}</p>
                      {session.isCurrent && (
                        <Badge variant="secondary" className="text-xs">
                          Current Session
                        </Badge>
                      )}
                    </div>
                    <p className="text-xs text-muted-foreground">
                      {session.location} • {session.ipAddress}
                    </p>
                    <div className="text-xs text-muted-foreground space-y-0.5">
                      <p>
                        Active since{' '}
                        {formatDate(session.createdAt)}
                      </p>
                      <p>
                        Last active{' '}
                        {formatDate(session.lastActive)}
                      </p>
                    </div>
                  </div>
                </div>
                {!session.isCurrent && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => handleRevokeSession(session.id)}
                    disabled={revoking === session.id}
                    className="ml-4"
                  >
                    {revoking === session.id ? (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    ) : (
                      <>
                        <LogOut className="h-4 w-4 mr-2" />
                        Revoke
                      </>
                    )}
                  </Button>
                )}
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
