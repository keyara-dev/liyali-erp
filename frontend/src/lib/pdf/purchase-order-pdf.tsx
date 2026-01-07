import React from 'react'
import {
  Document,
  Page,
  Text,
  View,
  Image,
} from '@react-pdf/renderer'
import { PurchaseOrder } from '@/types/purchase-order'
import { pdfStyles } from './pdf-styles'
import { generateDocumentQRData, generateTrackingCode } from './qr-utils'

interface PurchaseOrderPDFProps {
  purchaseOrder: PurchaseOrder
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

const PurchaseOrderPDF: React.FC<PurchaseOrderPDFProps> = ({ purchaseOrder, qrCodeUrl }) => {
  const trackingCode = generateTrackingCode('PURCHASE_ORDER', purchaseOrder.poNumber)
  const qrData = generateDocumentQRData(
    'PURCHASE_ORDER',
    purchaseOrder.poNumber,
    purchaseOrder.id,
    new Date(purchaseOrder.createdAt)
  )

  return (
    <Document>
      <Page size="A4" style={pdfStyles.page}>
        {/* Header with Republic of Zambia and Logo */}
        <View style={{ marginBottom: 20, textAlign: 'center' }}>
          <Text style={{ fontSize: 11, fontWeight: 'bold', marginBottom: 5 }}>
            REPUBLIC OF ZAMBIA
          </Text>
          <Text style={{ fontSize: 14, fontWeight: 'bold', marginBottom: 8 }}>
            PURCHASE ORDER
          </Text>
        </View>

        {/* Main Header Section */}
        <View style={[pdfStyles.header, { marginBottom: 20, flexDirection: 'row', justifyContent: 'space-between' }]}>
          <View>
            <Text style={{ fontSize: 14, fontWeight: 'bold', marginBottom: 3 }}>Liyali</Text>
            <Text style={{ fontSize: 9, color: '#666' }}>Finance & Procurement System</Text>
          </View>
          <View style={{ textAlign: 'right' }}>
            <Text style={{ fontSize: 11, fontWeight: 'bold', marginBottom: 2 }}>
              Purchase Order No: {purchaseOrder.poNumber}
            </Text>
            <Text style={{ fontSize: 9, color: '#666', marginBottom: 4 }}>
              Date: {new Date(purchaseOrder.createdAt).toLocaleDateString()}
            </Text>
            <View style={{
              borderWidth: 1,
              borderColor: '#ddd',
              padding: 6,
              width: 80,
              textAlign: 'center',
              marginLeft: 'auto'
            }}>
              <Text style={{ fontSize: 8, fontWeight: 'bold', marginBottom: 2 }}>TRACKING CODE</Text>
              <Text style={{ fontSize: 7 }}>{trackingCode}</Text>
            </View>
          </View>
        </View>

        {/* Status Badges */}
        <View style={{ marginBottom: 20, flexDirection: 'row', gap: 10 }}>
          <View style={[pdfStyles.statusBadge, getStatusColor(purchaseOrder.status)]}>
            <Text style={{ fontSize: 9 }}>{purchaseOrder.status}</Text>
          </View>
        </View>

        {/* Important Notice */}
        <View style={{ marginBottom: 15, padding: 8, backgroundColor: '#fff3cd', borderWidth: 1, borderColor: '#ffc107' }}>
          <Text style={{ fontSize: 8, fontWeight: 'bold', marginBottom: 3 }}>IMPORTANT INSTRUCTIONS:</Text>
          <Text style={{ fontSize: 7, lineHeight: 1.4 }}>
            • All packages and invoices must be marked with this PO number
            • Supply services in accordance with the contract/quotation
            • Invoice must be submitted with original copy of this order for payment
            • Delivery should be made to the address specified below
          </Text>
        </View>

        {/* Vendor Information Section */}
        <View style={{ marginBottom: 15, borderWidth: 1, borderColor: '#1e40af', padding: 10 }}>
          <Text style={{ fontSize: 11, fontWeight: 'bold', backgroundColor: '#dbeafe', padding: 5, marginBottom: 10 }}>
            TO (VENDOR/SUPPLIER)
          </Text>

          <View>
            <Text style={{ fontSize: 8, fontWeight: 'bold', marginBottom: 2, color: '#666' }}>VENDOR NAME</Text>
            <Text style={{ fontSize: 11, fontWeight: 'bold', marginBottom: 12 }}>{purchaseOrder.vendorName || '—'}</Text>

            <Text style={{ fontSize: 8, fontWeight: 'bold', marginBottom: 2, color: '#666' }}>DEPARTMENT</Text>
            <Text style={{ fontSize: 10 }}>{purchaseOrder.department || '—'}</Text>
          </View>
        </View>

        {/* Order Details */}
        <View style={{ marginBottom: 15, display: 'flex', flexDirection: 'row', gap: 20 }}>
          <View style={{ flex: 1, borderWidth: 1, borderColor: '#ddd', padding: 8 }}>
            <Text style={{ fontSize: 8, fontWeight: 'bold', marginBottom: 2, color: '#666' }}>REQUEST DATE</Text>
            <Text style={{ fontSize: 10 }}>{new Date(purchaseOrder.createdAt).toLocaleDateString()}</Text>
          </View>
          <View style={{ flex: 1, borderWidth: 1, borderColor: '#ddd', padding: 8 }}>
            <Text style={{ fontSize: 8, fontWeight: 'bold', marginBottom: 2, color: '#666' }}>REQUIRED BY DATE</Text>
            <Text style={{ fontSize: 10 }}>{purchaseOrder.requiredByDate ? new Date(purchaseOrder.requiredByDate).toLocaleDateString() : '—'}</Text>
          </View>
          <View style={{ flex: 1, borderWidth: 1, borderColor: '#ddd', padding: 8 }}>
            <Text style={{ fontSize: 8, fontWeight: 'bold', marginBottom: 2, color: '#666' }}>PRIORITY</Text>
            <Text style={{ fontSize: 10 }}>{purchaseOrder.priority || 'MEDIUM'}</Text>
          </View>
        </View>

        {/* Line Items Table */}
        {purchaseOrder.items && purchaseOrder.items.length > 0 && (
          <View style={{ marginBottom: 15 }}>
            {/* Table Header */}
            <View style={[pdfStyles.tableHeaderRow, { paddingVertical: 6 }]}>
              <Text style={{ ...pdfStyles.tableHeaderCell, width: '8%' }}>#</Text>
              <Text style={{ ...pdfStyles.tableHeaderCell, width: '25%' }}>Description</Text>
              <Text style={{ ...pdfStyles.tableHeaderCell, width: '12%', textAlign: 'right' }}>Quantity</Text>
              <Text style={{ ...pdfStyles.tableHeaderCell, width: '15%', textAlign: 'right' }}>Unit Price</Text>
              <Text style={{ ...pdfStyles.tableHeaderCell, width: '20%', textAlign: 'right' }}>Amount</Text>
              <Text style={{ ...pdfStyles.tableHeaderCell, width: '20%' }}>Remarks</Text>
            </View>

            {/* Table Rows */}
            {purchaseOrder.items.map((item: any, index: number) => (
              <View key={item.id} style={[pdfStyles.tableRow, { paddingVertical: 5 }]}>
                <Text style={{ ...pdfStyles.tableCell, width: '8%' }}>{index + 1}</Text>
                <Text style={{ ...pdfStyles.tableCell, width: '25%' }}>
                  {item.description}
                </Text>
                <Text style={{ ...pdfStyles.tableCell, width: '12%', textAlign: 'right' }}>
                  {item.quantity} {item.unit}
                </Text>
                <Text style={{ ...pdfStyles.tableCell, width: '15%', textAlign: 'right' }}>
                  {purchaseOrder.currency} {item.unitPrice?.toLocaleString() || '0'}
                </Text>
                <Text style={{ ...pdfStyles.tableCell, width: '20%', textAlign: 'right' }}>
                  {purchaseOrder.currency} {item.totalPrice?.toLocaleString() || '0'}
                </Text>
                <Text style={{ ...pdfStyles.tableCell, width: '20%', fontSize: 7 }}>
                  {item.remarks || '—'}
                </Text>
              </View>
            ))}

            {/* Financial Summary */}
            <View style={{ marginTop: 10, display: 'flex', flexDirection: 'row', gap: 20, justifyContent: 'flex-end' }}>
              <View style={{ width: '35%' }}>
                <View style={{ display: 'flex', flexDirection: 'row', justifyContent: 'space-between', borderTopWidth: 2, borderTopColor: '#1e40af', paddingTop: 5 }}>
                  <Text style={{ fontSize: 10, fontWeight: 'bold' }}>TOTAL ORDER VALUE:</Text>
                  <Text style={{ fontSize: 11, fontWeight: 'bold', color: '#1e40af' }}>
                    {purchaseOrder.currency} {purchaseOrder.totalAmount?.toLocaleString() || '0'}
                  </Text>
                </View>
              </View>
            </View>
          </View>
        )}

        {/* Financial Information */}
        {(purchaseOrder.budgetCode || purchaseOrder.costCenter) && (
          <View style={{ marginBottom: 15, borderWidth: 1, borderColor: '#1e40af', padding: 10 }}>
            <Text style={{ fontSize: 11, fontWeight: 'bold', backgroundColor: '#dbeafe', padding: 5, marginBottom: 10 }}>
              FINANCIAL INFORMATION
            </Text>
            <View style={{ display: 'flex', flexDirection: 'row', gap: 20 }}>
              {purchaseOrder.budgetCode && (
                <View style={{ flex: 1 }}>
                  <Text style={{ fontSize: 8, fontWeight: 'bold', marginBottom: 2, color: '#666' }}>BUDGET CODE</Text>
                  <Text style={{ fontSize: 9, fontFamily: 'Courier' }}>{purchaseOrder.budgetCode}</Text>
                </View>
              )}
              {purchaseOrder.costCenter && (
                <View style={{ flex: 1 }}>
                  <Text style={{ fontSize: 8, fontWeight: 'bold', marginBottom: 2, color: '#666' }}>COST CENTER</Text>
                  <Text style={{ fontSize: 9, fontFamily: 'Courier' }}>{purchaseOrder.costCenter}</Text>
                </View>
              )}
              {purchaseOrder.projectCode && (
                <View style={{ flex: 1 }}>
                  <Text style={{ fontSize: 8, fontWeight: 'bold', marginBottom: 2, color: '#666' }}>PROJECT CODE</Text>
                  <Text style={{ fontSize: 9, fontFamily: 'Courier' }}>{purchaseOrder.projectCode}</Text>
                </View>
              )}
            </View>
          </View>
        )}

        {/* APPROVAL SIGNATURES - Dynamic based on actual workflow */}
        {purchaseOrder.approvalChain && purchaseOrder.approvalChain.length > 0 && (
          <View style={{ marginBottom: 15, borderWidth: 1, borderColor: '#1e40af', padding: 10 }}>
            <Text style={{ fontSize: 11, fontWeight: 'bold', backgroundColor: '#dbeafe', padding: 5, marginBottom: 10 }}>
              APPROVAL SIGNATURES
            </Text>

            <View style={{ display: 'flex', flexDirection: 'row', gap: 10, flexWrap: 'wrap' }}>
              {purchaseOrder.approvalChain!.map((stage: any, index: number) => (
                <View
                  key={index}
                  style={{
                    flex: 1,
                    minWidth: index === purchaseOrder.approvalChain!.length - 1 && index % 2 === 0 ? '48%' : '48%',
                    borderWidth: 1,
                    borderColor: '#ddd',
                    padding: 8,
                    marginBottom: 8,
                  }}
                >
                  <Text style={{ fontSize: 8, fontWeight: 'bold', marginBottom: 2, color: '#1e40af' }}>
                    {stage.stageName || `Approval Stage ${stage.stageNumber}`}
                  </Text>
                  <Text style={{ fontSize: 7, marginBottom: 2 }}>Assigned to: {stage.assignedTo}</Text>
                  <Text style={{ fontSize: 7, marginBottom: 3 }}>Status: {stage.status}</Text>
                  {stage.actionTakenAt && (
                    <Text style={{ fontSize: 7, color: '#666' }}>
                      Date: {new Date(stage.actionTakenAt).toLocaleDateString()}
                    </Text>
                  )}
                  <View style={{ marginTop: 6, minHeight: 30, borderTopWidth: 1, borderTopColor: '#999', paddingTop: 3 }}>
                    <Text style={{ fontSize: 6, color: '#999' }}>Signature</Text>
                  </View>
                </View>
              ))}
            </View>
          </View>
        )}

        {/* Source Requisition (if applicable) */}
        {purchaseOrder.sourceRequisitionNumber && (
          <View style={{
            marginBottom: 15,
            padding: 10,
            backgroundColor: '#f0f7ff',
            borderLeftWidth: 4,
            borderLeftColor: '#7c3aed'
          }}>
            <Text style={{ fontSize: 8, fontWeight: 'bold', marginBottom: 3 }}>SOURCE REQUISITION</Text>
            <Text style={{ fontSize: 9 }}>
              Requisition: {purchaseOrder.sourceRequisitionNumber}
            </Text>
          </View>
        )}

        {/* QR Code and Tracking Information */}
        <View style={{
          marginTop: 15,
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
            <Text style={{ fontSize: 7, marginBottom: 2 }}>Document ID: {purchaseOrder.id}</Text>
            <Text style={{ fontSize: 7, marginBottom: 2 }}>Status: {purchaseOrder.status}</Text>
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

export default PurchaseOrderPDF
