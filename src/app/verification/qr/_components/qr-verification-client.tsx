'use client'

import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { CheckCircle2, AlertCircle, QrCode } from 'lucide-react'

interface QRVerificationClientProps {
  userId: string
  userRole: string
}

interface VerifiedDocument {
  id: string
  documentNumber: string
  qrCode: string
  verifiedAt: string
  verifiedBy: string
  status: 'valid' | 'invalid' | 'expired'
  documentType: string
  hashValue: string
}

const VERIFIED_DOCUMENTS: VerifiedDocument[] = [
  {
    id: '1',
    documentNumber: 'REQ-2024-001',
    qrCode: 'QR-REQ-2024-001-ABC123',
    verifiedAt: new Date(Date.now() - 2 * 60 * 60 * 1000).toLocaleString(),
    verifiedBy: 'John Mwale',
    status: 'valid',
    documentType: 'REQUISITION',
    hashValue: 'f8a9c2b1d4e7f3a9c2b1d4e7f3a9c2b1',
  },
  {
    id: '2',
    documentNumber: 'PO-2024-042',
    qrCode: 'QR-PO-2024-042-DEF456',
    verifiedAt: new Date(Date.now() - 5 * 60 * 60 * 1000).toLocaleString(),
    verifiedBy: 'James Chileshe',
    status: 'valid',
    documentType: 'PURCHASE_ORDER',
    hashValue: 'a3b2c1d4e7f3a9c2b1d4e7f3a9c2b1d4',
  },
  {
    id: '3',
    documentNumber: 'PV-2024-015',
    qrCode: 'QR-PV-2024-015-GHI789',
    verifiedAt: new Date(Date.now() - 24 * 60 * 60 * 1000).toLocaleString(),
    verifiedBy: 'Paul Nkosi',
    status: 'expired',
    documentType: 'PAYMENT_VOUCHER',
    hashValue: 'c1b2a3d4e7f3a9c2b1d4e7f3a9c2b1d4',
  },
  {
    id: '4',
    documentNumber: 'REQ-2024-025',
    qrCode: 'QR-REQ-2024-025-JKL012',
    verifiedAt: new Date(Date.now() - 1 * 60 * 60 * 1000).toLocaleString(),
    verifiedBy: 'Sarah Banda',
    status: 'invalid',
    documentType: 'REQUISITION',
    hashValue: 'e7f3a9c2b1d4e7f3a9c2b1d4e7f3a9c2',
  },
]

export function QRVerificationClient({ userId, userRole }: QRVerificationClientProps) {
  const [activeTab, setActiveTab] = useState('scan')
  const [qrInput, setQrInput] = useState('')
  const [scanResult, setScanResult] = useState<VerifiedDocument | null>(null)
  const [scanned, setScanned] = useState(false)

  const handleScan = () => {
    if (!qrInput) return

    const found = VERIFIED_DOCUMENTS.find((doc) => doc.qrCode === qrInput || doc.documentNumber === qrInput)
    setScanResult(found || null)
    setScanned(true)
  }

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-xl font-bold tracking-tight lg:text-2xl">QR Code Verification</h1>
        <p className="text-sm text-muted-foreground">
          Verify and authenticate documents using QR codes
        </p>
      </div>

      {/* Tabs */}
      <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
        <TabsList className="grid w-full grid-cols-2 lg:w-auto">
          <TabsTrigger value="scan">Scan QR Code</TabsTrigger>
          <TabsTrigger value="history">Verification History</TabsTrigger>
        </TabsList>

        {/* Scan Tab */}
        <TabsContent value="scan" className="space-y-6">
          {/* QR Scanner Card */}
          <Card>
            <CardHeader>
              <CardTitle className="text-lg flex items-center gap-2">
                <QrCode className="h-5 w-5" />
                Scan Document QR Code
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {/* Input */}
              <div className="space-y-2">
                <label className="text-sm font-medium">QR Code or Document Number</label>
                <div className="flex gap-2">
                  <Input
                    placeholder="Enter QR code (e.g., QR-REQ-2024-001-ABC123) or document number"
                    value={qrInput}
                    onChange={(e) => {
                      setQrInput(e.target.value)
                      setScanned(false)
                      setScanResult(null)
                    }}
                    onKeyPress={(e) => e.key === 'Enter' && handleScan()}
                  />
                  <Button onClick={handleScan}>
                    Verify
                  </Button>
                </div>
              </div>

              {/* Scan Result */}
              {scanned && (
                <div className={`border rounded-lg p-6 ${
                  scanResult
                    ? scanResult.status === 'valid'
                      ? 'bg-secondary/10 border-secondary/30'
                      : scanResult.status === 'expired'
                        ? 'bg-accent/10 border-accent/30'
                        : 'bg-destructive/10 border-destructive/30'
                    : 'bg-muted/10 border-muted/30'
                }`}>
                  {scanResult ? (
                    <div className="space-y-4">
                      {/* Status Header */}
                      <div className="flex items-center gap-3">
                        {scanResult.status === 'valid' ? (
                          <CheckCircle2 className="h-8 w-8 text-secondary" />
                        ) : (
                          <AlertCircle className="h-8 w-8 text-destructive" />
                        )}
                        <div>
                          <h3 className="font-semibold text-lg">
                            {scanResult.status === 'valid'
                              ? 'Document Verified'
                              : scanResult.status === 'expired'
                                ? 'Verification Expired'
                                : 'Invalid Document'}
                          </h3>
                          <p className="text-sm text-muted-foreground">
                            {scanResult.status === 'valid'
                              ? 'This document has been authenticated'
                              : scanResult.status === 'expired'
                                ? 'Verification has expired, re-authentication required'
                                : 'QR code does not match document'}
                          </p>
                        </div>
                      </div>

                      {/* Document Details */}
                      <div className="grid grid-cols-2 gap-4">
                        <div>
                          <p className="text-sm text-muted-foreground">Document Number</p>
                          <p className="font-semibold">{scanResult.documentNumber}</p>
                        </div>
                        <div>
                          <p className="text-sm text-muted-foreground">Document Type</p>
                          <p className="font-semibold">{scanResult.documentType}</p>
                        </div>
                        <div>
                          <p className="text-sm text-muted-foreground">Verified At</p>
                          <p className="font-semibold text-sm">{scanResult.verifiedAt}</p>
                        </div>
                        <div>
                          <p className="text-sm text-muted-foreground">Verified By</p>
                          <p className="font-semibold">{scanResult.verifiedBy}</p>
                        </div>
                      </div>

                      {/* Hash Value */}
                      <div className="bg-muted/50 p-3 rounded-lg">
                        <p className="text-xs font-medium text-muted-foreground mb-1">Document Hash</p>
                        <p className="font-mono text-xs break-all">{scanResult.hashValue}</p>
                      </div>

                      {/* Status Badge */}
                      <div className="flex items-center justify-between">
                        <span className="text-sm font-medium">Verification Status</span>
                        <Badge variant={
                          scanResult.status === 'valid' ? 'default' : 'destructive'
                        } className={scanResult.status === 'valid' ? 'bg-secondary' : ''}>
                          {scanResult.status.toUpperCase()}
                        </Badge>
                      </div>

                      {/* Actions */}
                      <div className="flex gap-2 pt-2">
                        <Button variant="outline" className="flex-1">
                          Download Certificate
                        </Button>
                        <Button variant="outline" className="flex-1">
                          View Document
                        </Button>
                      </div>
                    </div>
                  ) : (
                    <div className="text-center py-4">
                      <p className="text-destructive font-semibold">Document Not Found</p>
                      <p className="text-sm text-muted-foreground mt-1">
                        The QR code or document number you entered could not be found in the system.
                      </p>
                    </div>
                  )}
                </div>
              )}

              {/* Info Box */}
              <div className="bg-primary/5 border border-primary/20 rounded-lg p-4">
                <p className="text-sm">
                  <span className="font-medium">How it works:</span> Scan the QR code on your document or
                  enter the document number. The system will verify the document's authenticity and show
                  verification details.
                </p>
              </div>
            </CardContent>
          </Card>

          {/* Mock QR Codes Info */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Test QR Codes</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-2 text-sm">
                <p className="text-muted-foreground">Try scanning one of these:</p>
                <ul className="list-disc list-inside space-y-1 text-muted-foreground">
                  <li>QR-REQ-2024-001-ABC123 (Valid)</li>
                  <li>QR-PO-2024-042-DEF456 (Valid)</li>
                  <li>QR-PV-2024-015-GHI789 (Expired)</li>
                  <li>QR-REQ-2024-025-JKL012 (Invalid)</li>
                </ul>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* History Tab */}
        <TabsContent value="history" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Verification History</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="rounded-md border overflow-hidden">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Document Number</TableHead>
                      <TableHead>Type</TableHead>
                      <TableHead>Verified At</TableHead>
                      <TableHead>Verified By</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>QR Code</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {VERIFIED_DOCUMENTS.map((doc) => (
                      <TableRow key={doc.id}>
                        <TableCell className="font-medium">{doc.documentNumber}</TableCell>
                        <TableCell className="text-sm">{doc.documentType}</TableCell>
                        <TableCell className="text-sm">{doc.verifiedAt}</TableCell>
                        <TableCell className="text-sm">{doc.verifiedBy}</TableCell>
                        <TableCell>
                          <Badge variant={
                            doc.status === 'valid' ? 'default' : doc.status === 'expired' ? 'outline' : 'destructive'
                          } className={doc.status === 'valid' ? 'bg-secondary' : ''}>
                            {doc.status}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <code className="text-xs bg-muted px-2 py-1 rounded">
                            {doc.qrCode.substring(0, 15)}...
                          </code>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
