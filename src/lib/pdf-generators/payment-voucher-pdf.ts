import { Document, Page, Text, View, StyleSheet } from '@react-pdf/renderer'
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
  amountGrid: {
    display: 'flex',
    flexDirection: 'row',
    justifyContent: 'flex-end',
    marginTop: 20,
  },
  amountItem: {
    width: '50%',
    paddingLeft: 20,
  },
  amountLabel: {
    fontSize: 10,
    color: '#666',
    marginBottom: 5,
  },
  amountValue: {
    fontSize: 14,
    fontWeight: 'bold',
    color: '#000',
    paddingBottom: 5,
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
})

interface PVPDFProps {
  pv: WorkflowDocument
}

/**
 * PaymentVoucherPDF Component
 * Renders a PDF document for a payment voucher using @react-pdf/renderer
 */
function PaymentVoucherPDF({ pv }: PVPDFProps) {
  const vendorName = pv.metadata?.vendorName || 'N/A'
  const grossAmount = pv.metadata?.grossAmount || 0
  const tax = pv.metadata?.tax || 0
  const netAmount = pv.metadata?.netAmount || 0
  const bankInfo = pv.metadata?.bankInfo || {}

  return (
    <Document>
      <Page style={styles.page}>
        {/* Header */}
        <View style={styles.header}>
          <Text style={styles.title}>PAYMENT VOUCHER</Text>
          <Text style={styles.subtitle}>Mitete Town Council</Text>
          <View style={styles.infoGrid}>
            <View style={styles.infoItem}>
              <Text style={styles.label}>Voucher Number</Text>
              <Text style={styles.value}>{pv.documentNumber}</Text>
            </View>
            <View style={styles.infoItem}>
              <Text style={styles.label}>Date</Text>
              <Text style={styles.value}>
                {new Date(pv.createdAt).toLocaleDateString()}
              </Text>
            </View>
          </View>
        </View>

        {/* Vendor Information */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Payee Information</Text>
          <View style={styles.infoGrid}>
            <View style={styles.infoItem}>
              <Text style={styles.label}>Vendor/Payee Name</Text>
              <Text style={styles.value}>{vendorName}</Text>
            </View>
          </View>
        </View>

        {/* Bank Details */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Bank Details</Text>
          <View style={styles.infoGrid}>
            <View style={styles.infoItem}>
              <Text style={styles.label}>Account Name</Text>
              <Text style={styles.value}>{bankInfo.accountName || 'N/A'}</Text>
            </View>
            <View style={styles.infoItem}>
              <Text style={styles.label}>Account Number</Text>
              <Text style={styles.value}>{bankInfo.accountNumber || 'N/A'}</Text>
            </View>
            <View style={styles.infoItem}>
              <Text style={styles.label}>Bank Code</Text>
              <Text style={styles.value}>{bankInfo.bankCode || 'N/A'}</Text>
            </View>
            <View style={styles.infoItem}>
              <Text style={styles.label}>Bank Name</Text>
              <Text style={styles.value}>{bankInfo.bankName || 'N/A'}</Text>
            </View>
          </View>
        </View>

        {/* Amount Summary */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Amount Summary</Text>
          <View style={styles.amountGrid}>
            <View style={styles.amountItem}>
              <Text style={styles.amountLabel}>Gross Amount:</Text>
              <Text style={styles.amountValue}>K {grossAmount.toLocaleString()}</Text>

              <Text style={styles.amountLabel}>Tax:</Text>
              <Text style={styles.amountValue}>K {tax.toLocaleString()}</Text>

              <Text style={{ ...styles.amountLabel, marginTop: 10, fontWeight: 'bold' }}>
                Net Amount (Payable):
              </Text>
              <Text style={{ ...styles.amountValue, color: '#008000', fontSize: 18 }}>
                K {netAmount.toLocaleString()}
              </Text>
            </View>
          </View>
        </View>

        {/* References */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Document References</Text>
          <View style={styles.infoGrid}>
            <View style={styles.infoItem}>
              <Text style={styles.label}>GRN Reference</Text>
              <Text style={styles.value}>GRN-2024-001</Text>
            </View>
            <View style={styles.infoItem}>
              <Text style={styles.label}>PO Reference</Text>
              <Text style={styles.value}>PO-2024-001</Text>
            </View>
          </View>
        </View>

        {/* Approval Section */}
        <View style={styles.approvalSection}>
          <Text style={styles.sectionTitle}>APPROVAL SIGNATURES</Text>

          <View style={{ ...styles.approvalGrid, marginTop: 10 }}>
            {/* Stage 1 */}
            <View style={styles.approvalItem}>
              <Text style={{ fontSize: 10, fontWeight: 'bold' }}>
                Stage 1: Accountant Generation
              </Text>
              <Text style={{ fontSize: 9, color: '#666', marginTop: 5 }}>
                Prepared by: _______________
              </Text>
              <Text style={{ fontSize: 9, color: '#666', marginTop: 20 }}>
                Date: _______________
              </Text>
            </View>

            {/* Stage 2 */}
            <View style={styles.approvalItem}>
              <Text style={{ fontSize: 10, fontWeight: 'bold' }}>
                Stage 2: Department Head Review
              </Text>
              <Text style={{ fontSize: 9, color: '#666', marginTop: 5 }}>
                Reviewed by: _______________
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
                Stage 3: Auditor Review
              </Text>
              <Text style={{ fontSize: 9, color: '#666', marginTop: 5 }}>
                Audited by: _______________
              </Text>
              <Text style={{ fontSize: 9, color: '#666', marginTop: 20 }}>
                Date: _______________
              </Text>
            </View>

            {/* Stage 4 */}
            <View style={styles.approvalItem}>
              <Text style={{ fontSize: 10, fontWeight: 'bold' }}>
                Stage 4: Finance Director Review
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
            {/* Stage 5 */}
            <View style={styles.approvalItem}>
              <Text style={{ fontSize: 10, fontWeight: 'bold' }}>
                Stage 5: Principal Officer Approval
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
 * Generate and download the PDF for a payment voucher
 */
export async function generatePaymentVoucherPDF(pv: WorkflowDocument) {
  try {
    const element = <PaymentVoucherPDF pv={pv} />

    const link = document.createElement('a')
    const filename = `PV-${pv.documentNumber}-${new Date().getTime()}.pdf`

    console.log(`Generating PDF: ${filename}`)
    // The actual PDF generation will be handled by @react-pdf/renderer
    // This is a placeholder for the download logic
  } catch (error) {
    console.error('Error generating Payment Voucher PDF:', error)
    throw error
  }
}

export { PaymentVoucherPDF }
