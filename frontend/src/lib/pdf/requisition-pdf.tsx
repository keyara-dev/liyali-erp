import React from 'react'
import {
  Document,
  Page,
  Text,
  View,
  Image,
} from '@react-pdf/renderer'
import { Requisition } from '@/types/requisition'
import { pdfStyles } from './pdf-styles'
import { generateDocumentQRData, generateTrackingCode } from './qr-utils'

interface RequisitionPDFProps {
  requisition: Requisition
  qrCodeUrl?: string
}

const getStatusColor = (status: string) => {
  switch (status) {
    case 'DRAFT':
      return pdfStyles.statusDraft
    case 'SUBMITTED':
      return pdfStyles.statusSubmitted
    case 'IN_REVIEW':
      return pdfStyles.statusInReview
    case 'APPROVED':
      return pdfStyles.statusApproved
    case 'REJECTED':
      return pdfStyles.statusRejected
    default:
      return pdfStyles.statusDraft
  }
}

const RequisitionPDF: React.FC<RequisitionPDFProps> = ({ requisition, qrCodeUrl }) => {
  const trackingCode = generateTrackingCode('REQUISITION', requisition.requisitionNumber || requisition.reqNumber)
  const qrData = generateDocumentQRData(
    'REQUISITION',
    requisition.requisitionNumber || requisition.reqNumber,
    requisition.id,
    new Date(requisition.createdAt)
  )

  return (
    <Document>
      <Page size="A4" style={pdfStyles.page}>
        {/* Header with Republic of Zambia and Logo */}
        <View style={{ marginBottom: 15, textAlign: 'center', display: 'flex', flexDirection: 'row', justifyContent: 'center', alignItems: 'center', gap: 10 }}>
          {/* Logo */}
          <View style={{ width: 40, height: 40 }}>
            <Image
              src="/icon1.png"
              style={{ width: 40, height: 40 }}
            />
          </View>
          {/* Text */}
          <View style={{ textAlign: 'center' }}>
            <Text style={{ fontSize: 11, fontWeight: 'bold', marginBottom: 3 }}>
              REPUBLIC OF ZAMBIA
            </Text>
            <Text style={{ fontSize: 13, fontWeight: 'bold' }}>
              PURCHASE REQUISITION
            </Text>
          </View>
        </View>

        {/* Main Header Section */}
        <View style={[pdfStyles.header, { marginBottom: 15, flexDirection: 'row', justifyContent: 'space-between' }]}>
          <View style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-start' }}>
            <View style={{ marginBottom: 5 }}>
              <Image
                src="/icon1.png"
                style={{ width: 30, height: 30 }}
              />
            </View>
            <Text style={{ fontSize: 13, fontWeight: 'bold', marginBottom: 2, color: '#1e40af' }}>Liyali</Text>
            <Text style={{ fontSize: 8, color: '#666' }}>Finance & Procurement System</Text>
          </View>
          <View style={{ textAlign: 'right' }}>
            <Text style={{ fontSize: 10, fontWeight: 'bold', marginBottom: 2 }}>
              Requisition No: {requisition.requisitionNumber}
            </Text>
            <Text style={{ fontSize: 8, color: '#666', marginBottom: 3 }}>
              Date: {new Date(requisition.createdAt).toLocaleDateString()}
            </Text>
            <View style={{
              borderWidth: 1,
              borderColor: '#ddd',
              padding: 4,
              width: 75,
              textAlign: 'center',
              marginLeft: 'auto'
            }}>
              <Text style={{ fontSize: 7, fontWeight: 'bold', marginBottom: 1 }}>TRACKING CODE</Text>
              <Text style={{ fontSize: 6 }}>{trackingCode}</Text>
            </View>
          </View>
        </View>

        {/* Status Badges */}
        <View style={{ marginBottom: 20, flexDirection: 'row', gap: 10 }}>
          <View style={[pdfStyles.statusBadge, getStatusColor(requisition.status)]}>
            <Text style={{ fontSize: 9 }}>{requisition.status}</Text>
          </View>
          {requisition.priority && (
            <View
              style={[
                pdfStyles.statusBadge,
                {
                  backgroundColor:
                    requisition.priority === 'urgent'
                      ? '#fee2e2'
                      : requisition.priority === 'high'
                      ? '#fed7aa'
                      : '#dbeafe',
                  color:
                    requisition.priority === 'urgent'
                      ? '#991b1b'
                      : requisition.priority === 'high'
                      ? '#92400e'
                      : '#1e40af',
                },
              ]}
            >
              <Text style={{ fontSize: 9 }}>{requisition.priority}</Text>
            </View>
          )}
        </View>

        {/* SECTION 1: FOR REQUESTING OFFICE USE ONLY */}
        <View style={{ marginBottom: 20, borderWidth: 1, borderColor: '#1e40af', padding: 10 }}>
          <Text style={{ fontSize: 11, fontWeight: 'bold', backgroundColor: '#dbeafe', padding: 5, marginBottom: 10 }}>
            SECTION 1: FOR REQUESTING OFFICE USE ONLY
          </Text>

          {/* Requisition Info */}
          <View style={{ marginBottom: 12, display: 'flex', flexDirection: 'row', gap: 20 }}>
            <View style={{ flex: 1 }}>
              <Text style={{ fontSize: 8, fontWeight: 'bold', marginBottom: 2, color: '#666' }}>DEPARTMENT</Text>
              <Text style={{ fontSize: 10 }}>{requisition.department || '—'}</Text>
            </View>
            <View style={{ flex: 1 }}>
              <Text style={{ fontSize: 8, fontWeight: 'bold', marginBottom: 2, color: '#666' }}>PRIORITY</Text>
              <Text style={{ fontSize: 10 }}>{requisition.priority || '—'}</Text>
            </View>
            <View style={{ flex: 1 }}>
              <Text style={{ fontSize: 8, fontWeight: 'bold', marginBottom: 2, color: '#666' }}>BUDGET CODE</Text>
              <Text style={{ fontSize: 9, fontFamily: 'Courier' }}>{requisition.budgetCode || '—'}</Text>
            </View>
          </View>

          {/* Description */}
          {requisition.description && (
            <View style={{ marginBottom: 12 }}>
              <Text style={{ fontSize: 8, fontWeight: 'bold', marginBottom: 2, color: '#666' }}>DESCRIPTION</Text>
              <Text style={{ fontSize: 9 }}>{requisition.description}</Text>
            </View>
          )}

          {/* Requester Info */}
          <View style={{ display: 'flex', flexDirection: 'row', gap: 20 }}>
            <View style={{ flex: 1 }}>
              <Text style={{ fontSize: 8, fontWeight: 'bold', marginBottom: 2, color: '#666' }}>REQUESTED BY</Text>
              <Text style={{ fontSize: 10 }}>{requisition.requestedByName || '—'}</Text>
              <Text style={{ fontSize: 8, color: '#999' }}>{requisition.requestedByRole || '—'}</Text>
            </View>
            <View style={{ flex: 1 }}>
              <Text style={{ fontSize: 8, fontWeight: 'bold', marginBottom: 2, color: '#666' }}>DATE REQUESTED</Text>
              <Text style={{ fontSize: 10 }}>
                {new Date(requisition.createdAt).toLocaleDateString()}
              </Text>
            </View>
          </View>
        </View>

        {/* Line Items Table */}
        {requisition.items && requisition.items.length > 0 && (
          <View style={{ marginBottom: 20 }}>
            <Text style={{ fontSize: 10, fontWeight: 'bold', marginBottom: 8 }}>
              PLEASE PURCHASE THE FOLLOWING GOODS/SERVICES:
            </Text>

            {/* Table Header */}
            <View style={{ borderWidth: 1, borderColor: '#1e40af', marginBottom: 0 }}>
              <View style={{ display: 'flex', flexDirection: 'row', backgroundColor: '#f3f4f6', borderBottomWidth: 1, borderBottomColor: '#1e40af' }}>
                <Text style={{ flex: 0.5, padding: 5, fontSize: 8, fontWeight: 'bold', color: '#1e40af', textAlign: 'center' }}>Item</Text>
                <Text style={{ flex: 2, padding: 5, fontSize: 8, fontWeight: 'bold', color: '#1e40af', borderLeftWidth: 1, borderLeftColor: '#1e40af' }}>Description</Text>
                <Text style={{ flex: 1, padding: 5, fontSize: 8, fontWeight: 'bold', color: '#1e40af', textAlign: 'center', borderLeftWidth: 1, borderLeftColor: '#1e40af' }}>Qty</Text>
                <Text style={{ flex: 1, padding: 5, fontSize: 8, fontWeight: 'bold', color: '#1e40af', textAlign: 'right', borderLeftWidth: 1, borderLeftColor: '#1e40af' }}>Unit Price</Text>
                <Text style={{ flex: 1, padding: 5, fontSize: 8, fontWeight: 'bold', color: '#1e40af', textAlign: 'right', borderLeftWidth: 1, borderLeftColor: '#1e40af' }}>Total</Text>
              </View>

              {/* Table Rows */}
              {requisition.items.map((item: any, index: number) => {
                // Handle both naming conventions (description/itemDescription, unitPrice/estimatedCost)
                const itemDescription = item.description || item.itemDescription || ''
                const unitPrice = item.unitPrice || item.estimatedCost || 0
                const totalPrice = (item.totalPrice) || (item.quantity * unitPrice) || 0

                return (
                  <View key={item.id} style={{ display: 'flex', flexDirection: 'row', borderBottomWidth: 1, borderBottomColor: '#e5e7eb' }}>
                    <Text style={{ flex: 0.5, padding: 5, fontSize: 8, color: '#1f2937', textAlign: 'center' }}>{index + 1}</Text>
                    <Text style={{ flex: 2, padding: 5, fontSize: 8, color: '#1f2937', borderLeftWidth: 1, borderLeftColor: '#e5e7eb' }}>
                      {itemDescription}
                    </Text>
                    <Text style={{ flex: 1, padding: 5, fontSize: 8, color: '#1f2937', textAlign: 'center', borderLeftWidth: 1, borderLeftColor: '#e5e7eb' }}>
                      {item.quantity}
                    </Text>
                    <Text style={{ flex: 1, padding: 5, fontSize: 8, color: '#1f2937', textAlign: 'right', borderLeftWidth: 1, borderLeftColor: '#e5e7eb' }}>
                      {requisition.currency} {unitPrice?.toLocaleString() || '0'}
                    </Text>
                    <Text style={{ flex: 1, padding: 5, fontSize: 8, color: '#1f2937', textAlign: 'right', borderLeftWidth: 1, borderLeftColor: '#e5e7eb' }}>
                      {requisition.currency} {totalPrice?.toLocaleString() || '0'}
                    </Text>
                  </View>
                )
              })}
            </View>

            {/* Totals */}
            <View style={{ display: 'flex', flexDirection: 'row', justifyContent: 'flex-end', marginTop: 15, paddingTop: 10 }}>
              <View style={{ width: '35%' }}>
                <View style={{ display: 'flex', flexDirection: 'row', justifyContent: 'space-between', paddingBottom: 5, borderBottomWidth: 2, borderBottomColor: '#1e40af' }}>
                  <Text style={{ fontSize: 9, fontWeight: 'bold', color: '#1f2937' }}>TOTAL AMOUNT:</Text>
                  <Text style={{ fontSize: 11, fontWeight: 'bold', color: '#166534' }}>
                    {requisition.currency} {requisition.totalAmount?.toLocaleString() || '0'}
                  </Text>
                </View>
              </View>
            </View>
          </View>
        )}

        {/* SECTION 2: APPROVAL SIGNATURES */}
        {requisition.approvalChain && requisition.approvalChain.length > 0 && (
          <View style={{ marginBottom: 20, borderWidth: 1, borderColor: '#1e40af', padding: 10 }}>
            <Text style={{ fontSize: 11, fontWeight: 'bold', backgroundColor: '#dbeafe', padding: 5, marginBottom: 10 }}>
              APPROVAL CHAIN
            </Text>

            {/* Dynamic approval stages based on actual workflow */}
            <View style={{ display: 'flex', flexDirection: 'row', gap: 10, flexWrap: 'wrap' }}>
              {requisition.approvalChain.map((stage: any, index: number) => (
                <View
                  key={index}
                  style={{
                    flex: index % 2 === 0 ? 1 : 1,
                    minWidth: '45%',
                    borderWidth: 1,
                    borderColor: '#ddd',
                    padding: 8,
                    marginBottom: 8,
                  }}
                >
                  <Text style={{ fontSize: 8, fontWeight: 'bold', marginBottom: 3, color: '#1e40af' }}>
                    {stage.stageName || `Stage ${stage.stageNumber}`}
                  </Text>
                  <Text style={{ fontSize: 8, marginBottom: 2 }}>Assigned to: {stage.assignedTo}</Text>
                  <Text style={{ fontSize: 8, marginBottom: 4 }}>Status: {stage.status}</Text>
                  {stage.actionTakenAt && (
                    <Text style={{ fontSize: 7, color: '#666' }}>
                      Approved: {new Date(stage.actionTakenAt).toLocaleDateString()}
                    </Text>
                  )}
                  {stage.signature && (
                    <Text style={{ fontSize: 7, fontStyle: 'italic', color: '#999', marginTop: 3 }}>
                      Signature: {stage.signature}
                    </Text>
                  )}
                </View>
              ))}
            </View>
          </View>
        )}

        {/* QR Code and Tracking Information */}
        <View style={{
          marginTop: 20,
          paddingTop: 10,
          borderTopWidth: 1,
          borderTopColor: '#ddd',
          display: 'flex',
          flexDirection: 'row',
          gap: 15,
          alignItems: 'flex-start'
        }}>
          {/* QR Code Section */}
          {qrCodeUrl && (
            <View style={{ width: 80, height: 80 }}>
              <Image source={qrCodeUrl} style={{ width: 80, height: 80 }} />
            </View>
          )}

          {/* Tracking Information */}
          <View style={{ flex: 1 }}>
            <Text style={{ fontSize: 8, fontWeight: 'bold', marginBottom: 4 }}>DOCUMENT TRACKING</Text>
            <Text style={{ fontSize: 7, marginBottom: 2 }}>Tracking Code: {trackingCode}</Text>
            <Text style={{ fontSize: 7, marginBottom: 2 }}>Document ID: {requisition.id}</Text>
            <Text style={{ fontSize: 7, marginBottom: 2 }}>Status: {requisition.status}</Text>
            <Text style={{ fontSize: 7 }}>
              Generated: {new Date().toLocaleDateString()} {new Date().toLocaleTimeString()}
            </Text>
          </View>
        </View>

        {/* Footer */}
        <View
          style={{
            marginTop: 'auto',
            paddingTop: 10,
            borderTopWidth: 1,
            borderTopColor: '#ddd',
            textAlign: 'center',
          }}
        >
          <Text style={{ fontSize: 7, color: '#999' }}>
            This is a system-generated document. Digital signatures and QR codes verify authenticity.
          </Text>
          <Text style={{ fontSize: 7, color: '#999', marginTop: 2 }}>
            Scan the QR code above to verify this document.
          </Text>
        </View>
      </Page>
    </Document>
  )
}

export default RequisitionPDF
