import { Document, Page, Text, View, StyleSheet, PDFDownloadLink } from '@react-pdf/renderer'
import { WorkflowDocument } from '@/types/workflow'

// Define styles for the PDF
const styles = StyleSheet.create({
  page: {
    padding: 40,
    fontFamily: 'Helvetica',
    fontSize: 11,
    lineHeight: 1.5,
  },
  header: {
    marginBottom: 30,
    borderBottomWidth: 2,
    borderBottomColor: '#000',
    paddingBottom: 10,
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 5,
  },
  subtitle: {
    fontSize: 12,
    color: '#666',
    marginBottom: 10,
  },
  section: {
    marginBottom: 20,
    padding: 10,
    borderTopWidth: 1,
    borderTopColor: '#ccc',
  },
  sectionTitle: {
    fontSize: 12,
    fontWeight: 'bold',
    marginBottom: 10,
    color: '#333',
  },
  infoGrid: {
    display: 'flex',
    flexDirection: 'row',
    flexWrap: 'wrap',
    marginBottom: 10,
  },
  infoItem: {
    width: '50%',
    marginBottom: 10,
  },
  label: {
    fontSize: 9,
    color: '#666',
    marginBottom: 3,
  },
  value: {
    fontSize: 11,
    fontWeight: 'bold',
    color: '#000',
  },
  table: {
    marginTop: 10,
    marginBottom: 10,
  },
  tableRow: {
    display: 'flex',
    flexDirection: 'row',
    borderBottomWidth: 1,
    borderBottomColor: '#ddd',
    paddingTop: 8,
    paddingBottom: 8,
  },
  tableHeader: {
    display: 'flex',
    flexDirection: 'row',
    borderBottomWidth: 2,
    borderBottomColor: '#000',
    paddingTop: 8,
    paddingBottom: 8,
    fontWeight: 'bold',
  },
  tableCell: {
    flex: 1,
    paddingRight: 10,
    fontSize: 10,
  },
  tableCellRight: {
    flex: 1,
    paddingRight: 10,
    fontSize: 10,
    textAlign: 'right',
  },
  footerRow: {
    display: 'flex',
    flexDirection: 'row',
    borderTopWidth: 2,
    borderTopColor: '#000',
    paddingTop: 8,
    paddingBottom: 8,
    fontWeight: 'bold',
    backgroundColor: '#f5f5f5',
  },
  approvalSection: {
    marginTop: 40,
    padding: 10,
  },
  approvalGrid: {
    display: 'flex',
    flexDirection: 'row',
    gap: 20,
  },
  approvalItem: {
    flex: 1,
    borderWidth: 1,
    borderColor: '#ddd',
    padding: 10,
    minHeight: 100,
  },
  approvalName: {
    marginTop: 30,
    borderTopWidth: 1,
    borderTopColor: '#000',
    paddingTop: 5,
  },
})

interface POPDFProps {
  po: WorkflowDocument
}

/**
 * PurchaseOrderPDF Component
 * Renders a PDF document for a purchase order using @react-pdf/renderer
 */
function PurchaseOrderPDF({ po }: POPDFProps) {
  const items = po.metadata?.items || []
  const totalAmount = po.metadata?.totalAmount || 0
  const vendorName = po.metadata?.vendorName || 'N/A'

  return (
    <Document>
      <Page style={styles.page}>
        {/* Header */}
        <View style={styles.header}>
          <Text style={styles.title}>PURCHASE ORDER</Text>
          <Text style={styles.subtitle}>Mitete Town Council</Text>
          <View style={styles.infoGrid}>
            <View style={styles.infoItem}>
              <Text style={styles.label}>PO Number</Text>
              <Text style={styles.value}>{po.documentNumber}</Text>
            </View>
            <View style={styles.infoItem}>
              <Text style={styles.label}>Date</Text>
              <Text style={styles.value}>
                {new Date(po.createdAt).toLocaleDateString()}
              </Text>
            </View>
          </View>
        </View>

        {/* Vendor Information */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Vendor Information</Text>
          <View style={styles.infoGrid}>
            <View style={styles.infoItem}>
              <Text style={styles.label}>Vendor Name</Text>
              <Text style={styles.value}>{vendorName}</Text>
            </View>
            <View style={styles.infoItem}>
              <Text style={styles.label}>Vendor ID</Text>
              <Text style={styles.value}>{po.metadata?.vendorId || 'N/A'}</Text>
            </View>
            <View style={styles.infoItem}>
              <Text style={styles.label}>Delivery Type</Text>
              <Text style={styles.value}>{po.metadata?.deliveryType || 'Standard'}</Text>
            </View>
          </View>
        </View>

        {/* Items Table */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Order Items</Text>
          <View style={styles.table}>
            {/* Table Header */}
            <View style={styles.tableHeader}>
              <Text style={{ ...styles.tableCell, flex: 2 }}>Description</Text>
              <Text style={styles.tableCellRight}>Quantity</Text>
              <Text style={styles.tableCellRight}>Unit Cost</Text>
              <Text style={styles.tableCellRight}>Total Cost</Text>
            </View>

            {/* Table Rows */}
            {items.map((item: any, index: number) => (
              <View key={index} style={styles.tableRow}>
                <Text style={{ ...styles.tableCell, flex: 2 }}>
                  {item.description}
                </Text>
                <Text style={styles.tableCellRight}>{item.quantity}</Text>
                <Text style={styles.tableCellRight}>
                  K {item.unitCost.toLocaleString()}
                </Text>
                <Text style={styles.tableCellRight}>
                  K {item.totalCost.toLocaleString()}
                </Text>
              </View>
            ))}

            {/* Total Row */}
            <View style={styles.footerRow}>
              <Text style={{ ...styles.tableCell, flex: 2 }}></Text>
              <Text style={styles.tableCellRight}></Text>
              <Text style={styles.tableCellRight}>Total:</Text>
              <Text style={styles.tableCellRight}>
                K {totalAmount.toLocaleString()}
              </Text>
            </View>
          </View>
        </View>

        {/* Approval Section */}
        <View style={styles.approvalSection}>
          <Text style={styles.sectionTitle}>APPROVAL SIGNATURES</Text>

          {/* Signature lines for 4 stages */}
          <View style={styles.approvalGrid}>
            {/* Stage 1 */}
            <View style={styles.approvalItem}>
              <Text style={{ fontSize: 10, fontWeight: 'bold' }}>
                Stage 1: Department Head
              </Text>
              <Text style={{ fontSize: 9, color: '#666', marginTop: 5 }}>
                Approved by: _______________
              </Text>
              <Text style={{ fontSize: 9, color: '#666', marginTop: 20 }}>
                Date: _______________
              </Text>
            </View>

            {/* Stage 2 */}
            <View style={styles.approvalItem}>
              <Text style={{ fontSize: 10, fontWeight: 'bold' }}>
                Stage 2: Auditor Review
              </Text>
              <Text style={{ fontSize: 9, color: '#666', marginTop: 5 }}>
                Approved by: _______________
              </Text>
              <Text style={{ fontSize: 9, color: '#666', marginTop: 20 }}>
                Date: _______________
              </Text>
            </View>
          </View>

          <View style={{ ...styles.approvalGrid, marginTop: 10 }}>
            {/* Stage 3 */}
            <View style={styles.approvalItem}>
              <Text style={{ fontSize: 10, fontWeight: 'bold' }}>
                Stage 3: Finance Director
              </Text>
              <Text style={{ fontSize: 9, color: '#666', marginTop: 5 }}>
                Approved by: _______________
              </Text>
              <Text style={{ fontSize: 9, color: '#666', marginTop: 20 }}>
                Date: _______________
              </Text>
            </View>

            {/* Stage 4 */}
            <View style={styles.approvalItem}>
              <Text style={{ fontSize: 10, fontWeight: 'bold' }}>
                Stage 4: Principal Officer
              </Text>
              <Text style={{ fontSize: 9, color: '#666', marginTop: 5 }}>
                Approved by: _______________
              </Text>
              <Text style={{ fontSize: 9, color: '#666', marginTop: 20 }}>
                Date: _______________
              </Text>
            </View>
          </View>
        </View>
      </Page>
    </Document>
  )
}

/**
 * Generate and download the PDF for a purchase order
 */
export async function generatePurchaseOrderPDF(po: WorkflowDocument) {
  try {
    // Create a temporary link to download the PDF
    const element = <PurchaseOrderPDF po={po} />

    // Use PDFDownloadLink to trigger download
    const link = document.createElement('a')
    const filename = `PO-${po.documentNumber}-${new Date().getTime()}.pdf`

    // Note: In a real implementation, you would render the PDF and get the blob
    // For now, we'll use a library wrapper

    console.log(`Generating PDF: ${filename}`)
    // The actual PDF generation will be handled by @react-pdf/renderer
    // This is a placeholder for the download logic
  } catch (error) {
    console.error('Error generating PDF:', error)
    throw error
  }
}

export { PurchaseOrderPDF }
