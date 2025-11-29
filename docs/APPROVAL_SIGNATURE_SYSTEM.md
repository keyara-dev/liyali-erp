# Approval Signature & Confirmation System
## Digital Signature Capture for Approvals

**Date**: 2024-11-29
**Status**: Design Specification
**Purpose**: Enable users to sign approvals with remarks, creating a digital audit trail
**Reference**: Mitete Town Council PO No. 0760 (example document)

---

## Overview

When a user approves a document, they must:
1. Enter a **remark** (optional comment about the approval)
2. Provide a **digital signature** (signature pad or typed name)
3. Confirm the approval via **confirmation dialog**
4. System records: signature, timestamp, IP, remarks, approver details

This creates an **immutable approval record** that appears on the final document PDF.

---

## Physical Document Reference

Looking at the Mitete Town Council PO:

```
PURCHASE ORDER
┌─────────────────────────────────┐
│ Prepared by: [Signature]        │
│ Date: 19/11/25                  │
│                                 │
│ Approved by: [Signature]        │
│ Date: 19/11/25                  │
│                                 │
│ Authorised by: ________________ │
│ Date: ________________          │
└─────────────────────────────────┘
```

**In our digital system**, we need to capture:
- ✅ Who approved (user info)
- ✅ When they approved (timestamp)
- ✅ Their signature (digital signature)
- ✅ Their remarks (approval comments)
- ✅ Multi-stage approvals (different approvers at different stages)
- ✅ Digital PDF with all signatures

---

## Approval Confirmation Dialog

### Dialog 1: Standard Approval Confirmation

```
┌──────────────────────────────────────────────────────┐
│ Confirm Approval                                     │ ✕
├──────────────────────────────────────────────────────┤
│                                                      │
│ Document: Purchase Order No. PO-2024-0123          │
│ Stage: 2 of 4 (Auditor Review)                     │
│ Vendor: Broadway Ventures                           │
│ Amount: K 7,500.00                                  │
│                                                      │
├──────────────────────────────────────────────────────┤
│ APPROVAL DETAILS                                     │
├──────────────────────────────────────────────────────┤
│                                                      │
│ Your Remarks (Required if stage requires):          │
│ ┌──────────────────────────────────────────────────┐ │
│ │ Approved after compliance verification. All      │ │
│ │ documents are in order.                          │ │
│ │                                                  │ │
│ │                                                  │ │
│ └──────────────────────────────────────────────────┘ │
│                                                      │
│ Digital Signature:                                  │
│ ┌──────────────────────────────────────────────────┐ │
│ │         [Signature Pad Area]                     │ │
│ │                                                  │ │
│ │    Sign here with your mouse/touch pen          │ │
│ │                                                  │ │
│ └──────────────────────────────────────────────────┘ │
│ [Clear Signature]                                    │
│                                                      │
├──────────────────────────────────────────────────────┤
│ [Cancel]  [Submit Approval]                          │
└──────────────────────────────────────────────────────┘
```

### Dialog 2: Reversal Confirmation

```
┌──────────────────────────────────────────────────────┐
│ Confirm Reversal                                     │ ✕
├──────────────────────────────────────────────────────┤
│                                                      │
│ Document: Purchase Order No. PO-2024-0123          │
│ Stage: 3 of 4 (Finance Director Review)            │
│ Reversing Stage: Going back to Procurement Officer │
│                                                      │
├──────────────────────────────────────────────────────┤
│ REVERSAL DETAILS                                     │
├──────────────────────────────────────────────────────┤
│                                                      │
│ Reversal Reason (Required):                         │
│ ┌──────────────────────────────────────────────────┐ │
│ │ Bank details need verification. Please confirm  │ │
│ │ account information and resubmit.                │ │
│ │                                                  │ │
│ └──────────────────────────────────────────────────┘ │
│                                                      │
│ Digital Signature:                                  │
│ ┌──────────────────────────────────────────────────┐ │
│ │         [Signature Pad Area]                     │ │
│ │                                                  │ │
│ │    Sign here with your mouse/touch pen          │ │
│ │                                                  │ │
│ └──────────────────────────────────────────────────┘ │
│ [Clear Signature]                                    │
│                                                      │
├──────────────────────────────────────────────────────┤
│ [Cancel]  [Confirm Reversal]                         │
└──────────────────────────────────────────────────────┘
```

---

## Data Model

### Approval Record with Signature

```typescript
export type ApprovalRecordWithSignature = {
  // Existing fields
  stageNumber: number
  stageName: string
  assignedTo: string
  assignedRole: string
  status: 'PENDING' | 'APPROVED' | 'REVERSED' | 'REJECTED'
  actionTakenAt?: Date
  actionTakenBy?: string
  comments?: string
  reversedAt?: Date
  reversalReason?: string
  validationsPassed?: string[]
  validationsFailed?: string[]

  // NEW: Signature fields
  signature: {
    // Digital signature (base64 encoded PNG from signature pad)
    imageData: string

    // Alternative: typed signature (name)
    typedName?: string

    // How signature was provided
    signatureType: 'PAD' | 'TYPED'

    // When signed
    signedAt: Date

    // Who signed (user details)
    signedBy: string // user ID
    signedByName: string // user full name
    signedByEmail: string // user email
    signedByRole: string // user role (AUDITOR, etc.)

    // Where signed (security)
    signedFromIP: string
    signedFromBrowser: string

    // Remarks about approval
    remarks: string

    // Hash for verification
    signatureHash: string
  }
}
```

### Reversal Record with Signature

```typescript
export type ReversalRecordWithSignature = {
  // Existing reversal fields
  stageNumber: number
  stageName: string
  assignedRole: string
  assignedTo: string
  status: 'REVERSED'
  reversedAt: Date
  reversalReason: string

  // NEW: Signature fields (same as approval)
  signature: {
    imageData: string
    typedName?: string
    signatureType: 'PAD' | 'TYPED'
    reversedAt: Date
    reversedBy: string
    reversedByName: string
    reversedByEmail: string
    reversedByRole: string
    reversedFromIP: string
    reversedFromBrowser: string
    reversalReasons: string
    signatureHash: string
  }
}
```

---

## Signature Capture Component

### React Component: Signature Pad Integration

```typescript
// src/components/approval-signature-pad.tsx

import { useRef, useState } from 'react'
import SignatureCanvas from 'react-signature-canvas'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

export function ApprovalSignaturePad({
  onSignatureCapture,
  onClear
}: {
  onSignatureCapture: (signatureData: {
    imageData: string
    signatureType: 'PAD' | 'TYPED'
    typedName?: string
  }) => void
  onClear: () => void
}) {
  const signatureRef = useRef<SignatureCanvas>(null)
  const [typedName, setTypedName] = useState('')
  const [signatureType, setSignatureType] = useState<'PAD' | 'TYPED'>('PAD')

  const handleSignaturePad = () => {
    if (signatureRef.current) {
      const imageData = signatureRef.current.toDataURL('image/png')
      onSignatureCapture({
        imageData,
        signatureType: 'PAD'
      })
    }
  }

  const handleTypedName = () => {
    if (typedName.trim()) {
      onSignatureCapture({
        imageData: '', // No image for typed
        signatureType: 'TYPED',
        typedName: typedName.trim()
      })
    }
  }

  const handleClear = () => {
    if (signatureRef.current) {
      signatureRef.current.clear()
    }
    setTypedName('')
    onClear()
  }

  return (
    <Tabs value={signatureType} onValueChange={(v) => setSignatureType(v as 'PAD' | 'TYPED')}>
      <TabsList>
        <TabsTrigger value="PAD">Signature Pad</TabsTrigger>
        <TabsTrigger value="TYPED">Typed Name</TabsTrigger>
      </TabsList>

      <TabsContent value="PAD" className="space-y-4">
        <div className="border-2 border-dashed border-gray-300 rounded-lg bg-white">
          <SignatureCanvas
            ref={signatureRef}
            canvasProps={{
              width: 500,
              height: 150,
              className: 'border-0 w-full'
            }}
            backgroundColor="white"
          />
        </div>
        <p className="text-sm text-gray-600">
          Sign with your mouse or touch pen. Keep within the gray border.
        </p>
        <div className="flex gap-2">
          <Button variant="outline" onClick={handleClear}>
            Clear Signature
          </Button>
          <Button onClick={handleSignaturePad}>
            Use This Signature
          </Button>
        </div>
      </TabsContent>

      <TabsContent value="TYPED" className="space-y-4">
        <div className="space-y-2">
          <label className="text-sm font-medium">Full Name</label>
          <input
            type="text"
            placeholder="Enter your full name"
            value={typedName}
            onChange={(e) => setTypedName(e.target.value)}
            className="w-full border rounded px-3 py-2"
          />
        </div>
        <p className="text-sm text-gray-600">
          Your typed name will serve as your digital signature.
        </p>
        <div className="flex gap-2">
          <Button variant="outline" onClick={handleClear}>
            Clear
          </Button>
          <Button onClick={handleTypedName} disabled={!typedName.trim()}>
            Use This Name
          </Button>
        </div>
      </TabsContent>
    </Tabs>
  )
}
```

---

## Approval Confirmation Dialog Component

### React Component: Approval Dialog

```typescript
// src/components/approval-confirmation-dialog.tsx

'use client'

import { useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { ApprovalSignaturePad } from './approval-signature-pad'
import { ApproveDocumentRequest } from '@/types/workflow'

interface ApprovalDialogProps {
  open: boolean
  documentId: string
  documentType: string
  documentNumber: string
  vendor: string
  amount: string
  stageNumber: number
  totalStages: number
  stageName: string
  isReversal?: boolean
  onApprove: (data: ApproveDocumentRequest & {
    signature: {
      imageData: string
      signatureType: 'PAD' | 'TYPED'
      typedName?: string
    }
  }) => Promise<void>
  onCancel: () => void
}

export function ApprovalConfirmationDialog({
  open,
  documentId,
  documentType,
  documentNumber,
  vendor,
  amount,
  stageNumber,
  totalStages,
  stageName,
  isReversal = false,
  onApprove,
  onCancel
}: ApprovalDialogProps) {
  const [remarks, setRemarks] = useState('')
  const [signature, setSignature] = useState<{
    imageData: string
    signatureType: 'PAD' | 'TYPED'
    typedName?: string
  } | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  const handleApprove = async () => {
    if (!signature) {
      alert('Please provide a signature')
      return
    }

    if (isReversal && !remarks.trim()) {
      alert('Reversal reason is required')
      return
    }

    setIsLoading(true)
    try {
      await onApprove({
        documentId,
        documentType,
        approvingUserId: 'current-user-id', // From auth
        comments: remarks,
        signature
      })
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onCancel}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>
            {isReversal ? 'Confirm Reversal' : 'Confirm Approval'}
          </DialogTitle>
          <DialogDescription>
            Please review the details below and provide your signature
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6">
          {/* Document Info */}
          <div className="bg-gray-50 p-4 rounded-lg space-y-2">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className="text-sm text-gray-600">Document</p>
                <p className="font-medium">{documentNumber}</p>
              </div>
              <div>
                <p className="text-sm text-gray-600">Type</p>
                <p className="font-medium">{documentType}</p>
              </div>
              <div>
                <p className="text-sm text-gray-600">Vendor</p>
                <p className="font-medium">{vendor}</p>
              </div>
              <div>
                <p className="text-sm text-gray-600">Amount</p>
                <p className="font-medium">{amount}</p>
              </div>
              <div>
                <p className="text-sm text-gray-600">Stage</p>
                <p className="font-medium">
                  {stageNumber} of {totalStages}: {stageName}
                </p>
              </div>
            </div>
          </div>

          {/* Remarks/Reason */}
          <div className="space-y-2">
            <label className="font-medium">
              {isReversal ? 'Reversal Reason *' : 'Remarks (Optional)'}
            </label>
            <Textarea
              placeholder={
                isReversal
                  ? 'Explain why you are reversing this approval...'
                  : 'Any remarks about this approval...'
              }
              value={remarks}
              onChange={(e) => setRemarks(e.target.value)}
              className="h-24"
            />
          </div>

          {/* Signature Pad */}
          <div className="space-y-2">
            <label className="font-medium">Digital Signature *</label>
            <ApprovalSignaturePad
              onSignatureCapture={setSignature}
              onClear={() => setSignature(null)}
            />
            {signature && (
              <p className="text-sm text-green-600">
                ✓ Signature captured ({signature.signatureType})
              </p>
            )}
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onCancel} disabled={isLoading}>
            Cancel
          </Button>
          <Button
            onClick={handleApprove}
            disabled={isLoading || !signature}
            className={isReversal ? 'bg-red-600 hover:bg-red-700' : undefined}
          >
            {isLoading
              ? 'Processing...'
              : isReversal
                ? 'Confirm Reversal'
                : 'Submit Approval'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
```

---

## Server Action: Enhanced Approval with Signature

```typescript
// src/app/_actions/approval.ts (UPDATED)

export async function approveDocument(
  request: ApproveDocumentRequest & {
    signature: {
      imageData: string
      signatureType: 'PAD' | 'TYPED'
      typedName?: string
    }
  }
): Promise<ApproveDocumentResponse> {
  try {
    // 1. Load document and state
    const document = store.documents.get(request.documentId)
    let state = store.approvalStates.get(request.documentId)

    if (!document || !state) {
      return {
        success: false,
        message: 'Document not found',
        error: 'NOT_FOUND'
      }
    }

    // 2. Get configuration and current stage
    const config = getApprovalConfig(request.documentType)
    const currentStage = getCurrentApprovalStage(state)

    if (!currentStage) {
      return {
        success: false,
        message: 'Invalid approval stage',
        error: 'BAD_STAGE'
      }
    }

    // 3. Get user info
    const user = store.users.get(request.approvingUserId)
    if (!user) {
      return {
        success: false,
        message: 'User not found',
        error: 'USER_NOT_FOUND'
      }
    }

    // 4. Verify authorization
    const userRoles = user.roleIds || []
    if (!userHasApprovalRole(state, userRoles)) {
      return {
        success: false,
        message: `User does not have required role: ${currentStage.requiredRole}`,
        error: 'UNAUTHORIZED'
      }
    }

    // 5. Verify signature provided
    if (!request.signature || !request.signature.imageData && !request.signature.typedName) {
      return {
        success: false,
        message: 'Digital signature is required',
        error: 'NO_SIGNATURE'
      }
    }

    // 6. Create approval record with signature
    const approvalRecord: ApprovalRecord & { signature: any } = {
      stageNumber: currentStage.stageNumber,
      stageName: currentStage.stageName,
      assignedRole: currentStage.requiredRole,
      assignedTo: request.approvingUserId,
      status: 'APPROVED',
      actionTakenAt: new Date(),
      actionTakenBy: request.approvingUserId,
      comments: request.comments,
      validationsPassed: request.validations
        ? Object.entries(request.validations)
            .filter(([, passed]) => passed)
            .map(([key]) => key)
        : [],

      // NEW: Signature data
      signature: {
        imageData: request.signature.imageData,
        typedName: request.signature.typedName,
        signatureType: request.signature.signatureType,
        signedAt: new Date(),
        signedBy: user.id,
        signedByName: user.name,
        signedByEmail: user.email,
        signedByRole: currentStage.requiredRole,
        signedFromIP: 'client-ip', // From headers in real app
        signedFromBrowser: 'client-browser', // From headers in real app
        remarks: request.comments || '',
        signatureHash: generateSignatureHash(request.signature.imageData)
      }
    }

    state.stageHistory.push(approvalRecord)

    // 7. Check if final stage
    const isFinal = isFinalApprovalStage(state)

    if (isFinal) {
      state.status = 'APPROVED'
      state.approvedAt = new Date()
      state.currentStageNumber = config.totalStages
    } else {
      const nextStage = getNextApprovalStage(state)
      if (nextStage) {
        state.currentStageNumber = nextStage.stageNumber
      }
    }

    state.lastModifiedAt = new Date()
    state.lastModifiedBy = request.approvingUserId
    state.status = isFinal ? 'APPROVED' : 'IN_APPROVAL'

    // 8. Store updated state
    store.documents.set(request.documentId, document)
    store.approvalStates.set(request.documentId, state)

    // 9. Create audit log with signature reference
    const auditLogId = `audit-${Date.now()}`
    store.auditLogs.set(auditLogId, {
      id: auditLogId,
      documentId: document.id,
      action: isFinal ? 'FINAL_APPROVAL' : 'STAGE_APPROVAL',
      userId: request.approvingUserId,
      timestamp: new Date(),
      details: `Document approved at stage ${currentStage.stageNumber} by ${user.name}. Signature: ${request.signature.signatureType}`,
      metadata: {
        stageName: currentStage.stageName,
        stageNumber: currentStage.stageNumber,
        approverRole: currentStage.requiredRole,
        approverName: user.name,
        signatureType: request.signature.signatureType,
        signatureHash: approvalRecord.signature.signatureHash,
        remarks: request.comments
      }
    })

    return {
      success: true,
      message: isFinal ? 'Document fully approved' : 'Approval recorded',
      newStageNumber: state.currentStageNumber,
      isFinalApproval: isFinal
    }
  } catch (error) {
    console.error('Approval failed:', error)
    return {
      success: false,
      message: 'Approval failed',
      error: error instanceof Error ? error.message : 'UNKNOWN_ERROR'
    }
  }
}

// Helper function to generate signature hash
function generateSignatureHash(imageData: string): string {
  // In production, use crypto-js or similar
  return btoa(imageData.substring(0, 100))
}
```

---

## PDF Generation with Signatures

### PDF Document Structure

```typescript
// src/lib/document-pdf-generator.ts

import jsPDF from 'jspdf'

export async function generateApprovalPDF(
  document: any,
  state: ApprovalState
): Promise<Blob> {
  const pdf = new jsPDF()
  const pageWidth = pdf.internal.pageSize.getWidth()
  const pageHeight = pdf.internal.pageSize.getHeight()

  let yPosition = 10

  // Header
  pdf.setFontSize(16)
  pdf.text(`PURCHASE ORDER`, pageWidth / 2, yPosition, { align: 'center' })
  yPosition += 10

  // Document Details
  pdf.setFontSize(10)
  pdf.text(`Document No: ${document.documentNumber}`, 10, yPosition)
  yPosition += 7
  pdf.text(`Vendor: ${document.metadata?.vendorName}`, 10, yPosition)
  yPosition += 7
  pdf.text(`Amount: ${document.metadata?.totalAmount}`, 10, yPosition)
  yPosition += 7
  pdf.text(`Date: ${new Date(document.createdAt).toLocaleDateString()}`, 10, yPosition)
  yPosition += 15

  // Items Table
  pdf.setFontSize(9)
  pdf.text('Items:', 10, yPosition)
  yPosition += 7

  const items = document.metadata?.items || []
  items.forEach((item: any) => {
    pdf.text(`${item.description} - Qty: ${item.quantity} - Amount: ${item.totalCost}`, 15, yPosition)
    yPosition += 6
  })

  yPosition += 10

  // Approval Signatures Section
  pdf.setFontSize(12)
  pdf.text('APPROVAL SIGNATURES', 10, yPosition)
  yPosition += 10

  const approvals = state.stageHistory.filter(r => r.status === 'APPROVED')

  approvals.forEach((approval: any, index: number) => {
    const yStart = yPosition

    // Stage info
    pdf.setFontSize(10)
    pdf.text(`Stage ${approval.stageNumber}: ${approval.stageName}`, 10, yPosition)
    yPosition += 6

    // Approver name
    pdf.text(`Approved by: ${approval.signature?.signedByName}`, 10, yPosition)
    yPosition += 6

    // Signature image (if exists)
    if (approval.signature?.signatureType === 'PAD' && approval.signature?.imageData) {
      try {
        pdf.addImage(
          approval.signature.imageData,
          'PNG',
          10,
          yPosition,
          40,
          15
        )
        yPosition += 20
      } catch (e) {
        console.error('Error adding signature image:', e)
      }
    } else if (approval.signature?.signatureType === 'TYPED') {
      // Typed name as signature
      pdf.setFont(undefined, 'italic')
      pdf.text(`${approval.signature?.typedName}`, 10, yPosition)
      pdf.setFont(undefined, 'normal')
      yPosition += 6
    }

    // Date and remarks
    pdf.setFontSize(9)
    pdf.text(`Date: ${new Date(approval.actionTakenAt).toLocaleString()}`, 10, yPosition)
    yPosition += 5

    if (approval.signature?.remarks) {
      pdf.text(`Remarks: ${approval.signature.remarks}`, 10, yPosition)
      yPosition += 5
    }

    yPosition += 10

    // Add page break if needed
    if (yPosition > pageHeight - 20) {
      pdf.addPage()
      yPosition = 10
    }
  })

  return pdf.output('blob')
}
```

---

## Integration with Approval UI

### Updated Purchase Order Detail Page

```typescript
// src/app/workflows/purchase-orders/[id]/_components/po-detail-client.tsx

'use client'

import { useState } from 'react'
import { approveDocument, reverseDocument } from '@/app/_actions/approval'
import { ApprovalConfirmationDialog } from '@/components/approval-confirmation-dialog'
import { Button } from '@/components/ui/button'

export function PODetailClient({ poId }: { poId: string }) {
  const [showApprovalDialog, setShowApprovalDialog] = useState(false)
  const [showReversalDialog, setShowReversalDialog] = useState(false)
  const [isLoading, setIsLoading] = useState(false)

  const handleApproveClick = () => {
    setShowApprovalDialog(true)
  }

  const handleApproveSubmit = async (data: any) => {
    setIsLoading(true)
    try {
      const result = await approveDocument(data)
      if (result.success) {
        setShowApprovalDialog(false)
        // Refresh page or state
        window.location.reload()
      }
    } finally {
      setIsLoading(false)
    }
  }

  const handleReverseSubmit = async (data: any) => {
    setIsLoading(true)
    try {
      const result = await reverseDocument({
        documentId: poId,
        documentType: 'PURCHASE_ORDER',
        reversingUserId: 'current-user-id',
        reversalReason: data.comments
      })
      if (result.success) {
        setShowReversalDialog(false)
        window.location.reload()
      }
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div>
      {/* PO Details */}
      {/* ... existing content ... */}

      {/* Action Buttons */}
      <div className="flex gap-2 mt-6">
        <Button onClick={handleApproveClick} className="bg-green-600">
          Approve with Signature
        </Button>
        <Button onClick={() => setShowReversalDialog(true)} variant="destructive">
          Reverse with Signature
        </Button>
      </div>

      {/* Approval Dialog */}
      <ApprovalConfirmationDialog
        open={showApprovalDialog}
        documentId={poId}
        documentType="PURCHASE_ORDER"
        documentNumber="PO-2024-0123"
        vendor="Broadway Ventures"
        amount="K 7,500.00"
        stageNumber={2}
        totalStages={4}
        stageName="Auditor Review"
        onApprove={handleApproveSubmit}
        onCancel={() => setShowApprovalDialog(false)}
      />

      {/* Reversal Dialog */}
      <ApprovalConfirmationDialog
        open={showReversalDialog}
        documentId={poId}
        documentType="PURCHASE_ORDER"
        documentNumber="PO-2024-0123"
        vendor="Broadway Ventures"
        amount="K 7,500.00"
        stageNumber={2}
        totalStages={4}
        stageName="Auditor Review"
        isReversal={true}
        onApprove={handleReverseSubmit}
        onCancel={() => setShowReversalDialog(false)}
      />
    </div>
  )
}
```

---

## Signature Libraries to Use

### Option 1: react-signature-canvas (Recommended)
```bash
npm install react-signature-canvas
```

### Option 2: signature_pad
```bash
npm install signature_pad
```

### Option 3: Custom Canvas Implementation
```typescript
// Implement signature capture using HTML5 Canvas API
// More control, no dependencies
```

### PDF Generation
```bash
npm install jspdf
```

---

## Approval Record Display Component

### Show All Approvals with Signatures

```typescript
// src/components/approval-history.tsx

export function ApprovalHistory({ state }: { state: ApprovalState }) {
  const approvals = state.stageHistory.filter(r => r.status === 'APPROVED')

  return (
    <div className="space-y-4">
      <h3 className="font-semibold">Approval History</h3>

      {approvals.map((approval, index) => (
        <div key={index} className="border rounded-lg p-4 bg-gray-50">
          <div className="grid grid-cols-2 gap-4">
            {/* Left column */}
            <div>
              <p className="text-sm text-gray-600">Stage</p>
              <p className="font-medium">{approval.stageName}</p>

              <p className="text-sm text-gray-600 mt-2">Approved By</p>
              <p className="font-medium">{approval.signature?.signedByName}</p>

              <p className="text-sm text-gray-600 mt-2">Date & Time</p>
              <p className="font-medium">
                {new Date(approval.actionTakenAt).toLocaleString()}
              </p>
            </div>

            {/* Right column - Signature */}
            <div>
              <p className="text-sm text-gray-600 mb-2">Signature</p>
              {approval.signature?.signatureType === 'PAD' ? (
                <div className="border rounded bg-white p-2">
                  {approval.signature?.imageData ? (
                    <img
                      src={approval.signature.imageData}
                      alt="Signature"
                      className="h-12"
                    />
                  ) : (
                    <p className="text-gray-400">Signature image</p>
                  )}
                </div>
              ) : (
                <div className="font-italic text-gray-600">
                  {approval.signature?.typedName}
                </div>
              )}
            </div>
          </div>

          {/* Remarks */}
          {approval.signature?.remarks && (
            <div className="mt-3 pt-3 border-t">
              <p className="text-sm text-gray-600">Remarks</p>
              <p className="text-sm">{approval.signature.remarks}</p>
            </div>
          )}
        </div>
      ))}
    </div>
  )
}
```

---

## Security Considerations

### Signature Validation

```typescript
// src/lib/signature-validation.ts

import crypto from 'crypto'

export function validateSignature(
  signatureData: string,
  storedHash: string
): boolean {
  const calculatedHash = crypto
    .createHash('sha256')
    .update(signatureData)
    .digest('hex')

  return calculatedHash === storedHash
}

export function generateSignatureHash(signatureData: string): string {
  return crypto
    .createHash('sha256')
    .update(signatureData)
    .digest('hex')
}
```

### Signature Immutability

```typescript
// Once a signature is recorded:
// 1. Hash is calculated and stored
// 2. Signature data is stored (cannot be modified)
// 3. Timestamp is recorded
// 4. IP and browser info logged
// 5. User cannot edit approval after signing
// 6. Signature can only be removed by admin override
```

---

## Audit Trail with Signatures

### Complete Approval Audit Entry

```typescript
{
  id: 'audit-001',
  documentId: 'po-123',
  action: 'STAGE_APPROVAL',
  userId: 'user-456',
  timestamp: '2024-11-29T14:30:00Z',
  details: 'Purchase Order approved at stage 2 by John Doe',
  metadata: {
    stageName: 'Auditor Review',
    stageNumber: 2,
    approverRole: 'AUDITOR',
    approverName: 'John Doe',
    approverEmail: 'john@example.com',
    signatureType: 'PAD',
    signatureHash: 'abc123def456...',
    remarks: 'Approved after compliance verification',
    signedFromIP: '192.168.1.100',
    signedFromBrowser: 'Chrome 121.0',
    signedAt: '2024-11-29T14:30:00Z'
  }
}
```

---

## Implementation Checklist

### Phase 1: Signature Capture
- [ ] Install signature-canvas library
- [ ] Create ApprovalSignaturePad component
- [ ] Add signature field to ApprovalRecord type
- [ ] Update approval dialog UI

### Phase 2: Server-Side Integration
- [ ] Update approveDocument action with signature handling
- [ ] Update reverseDocument action with signature handling
- [ ] Create signature validation functions
- [ ] Update audit log to include signature info

### Phase 3: PDF Generation
- [ ] Install jsPDF library
- [ ] Create PDF generator function
- [ ] Add signature images to PDF
- [ ] Add approval history to PDF
- [ ] Add download button to UI

### Phase 4: Approval History Display
- [ ] Create ApprovalHistory component
- [ ] Display all approvals with signatures
- [ ] Show remarks for each approval
- [ ] Show timestamps and approver info

### Phase 5: Security & Validation
- [ ] Add signature hash verification
- [ ] Prevent signature tampering
- [ ] Log signature validation attempts
- [ ] Add audit trail entries for all signature actions

---

## Benefits

✅ **Legal Compliance**: Digital signatures satisfy many regulatory requirements
✅ **Audit Trail**: Complete record of who approved and when
✅ **Tamper Protection**: Hash verification ensures signature authenticity
✅ **User Accountability**: Signers can't deny approvals
✅ **Document Authenticity**: Signatures prove approval was authorized
✅ **Mobile-Friendly**: Supports signature pad on tablets/touch devices
✅ **Professional**: Looks like traditional approval document
✅ **Secure**: Encrypted storage of signature data

---

## Example Workflow

```
1. User clicks "Approve" button
   ↓
2. Approval Confirmation Dialog Opens
   - Shows document details
   - Shows current stage
   ↓
3. User Enters Remarks (optional)
   ↓
4. User Signs
   - Either with signature pad (draws)
   - Or types their name
   ↓
5. System Captures:
   - Signature image/name
   - Timestamp
   - User info
   - IP address
   - Browser info
   ↓
6. User Clicks "Submit Approval"
   ↓
7. Server:
   - Validates signature
   - Calculates hash
   - Records approval with signature
   - Creates audit entry
   - Updates document state
   - Sends notifications
   ↓
8. System:
   - Moves document to next stage
   - Updates UI
   - Regenerates PDF with signature
   - Shows success message
```

---

## PDF Output Example

The final PDF will show:

```
PURCHASE ORDER NO. PO-2024-0123
Vendor: Broadway Ventures
Amount: K 7,500.00

APPROVAL SIGNATURES:

Stage 1: Department Head Approval
Approved by: Jane Smith
Signature: [signature image or typed name]
Date: 2024-11-25 09:15 AM
Remarks: Approved for procurement

Stage 2: Auditor Review
Approved by: John Doe
Signature: [signature image or typed name]
Date: 2024-11-29 02:30 PM
Remarks: Approved after compliance verification

Stage 3: Finance Director Approval
Approved by: Robert Johnson
Signature: [signature image or typed name]
Date: 2024-11-29 03:45 PM
Remarks: Budget allocation confirmed
```

---

**Created**: 2024-11-29
**Status**: Ready for Implementation
**Priority**: High (Required for Phase 1)
**Effort**: 2-3 hours for initial implementation
**Dependencies**: jsPDF, react-signature-canvas, existing approval system
