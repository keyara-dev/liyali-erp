# PDF Generation System

The Liyali Gateway frontend includes a comprehensive PDF generation system for creating professional documents like requisitions, purchase orders, payment vouchers, and reports.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    PDF Generation Stack                     │
├─────────────────────────────────────────────────────────────┤
│ 1. React Components → PDF Templates                        │
│    ↕ Reusable templates with data binding                  │
│                                                             │
│ 2. @react-pdf/renderer → PDF Engine                        │
│    ↕ React-to-PDF conversion with styling                  │
│                                                             │
│ 3. Template System → Dynamic Layouts                       │
│    ↕ Configurable templates per document type              │
│                                                             │
│ 4. Preview System → Live Preview                           │
│    ↕ Real-time preview before generation                   │
│                                                             │
│ 5. Download/Print → Output Options                         │
│    ↕ Multiple output formats and delivery methods          │
└─────────────────────────────────────────────────────────────┘
```

## Core PDF Components

### Base PDF Document Structure

```typescript
// src/components/pdf/base-pdf-document.tsx
import {
  Document,
  Page,
  Text,
  View,
  StyleSheet,
  Font,
  Image,
} from '@react-pdf/renderer';

// Register fonts for better typography
Font.register({
  family: 'Inter',
  fonts: [
    { src: '/fonts/Inter-Regular.ttf' },
    { src: '/fonts/Inter-Bold.ttf', fontWeight: 'bold' },
    { src: '/fonts/Inter-Medium.ttf', fontWeight: 'medium' },
  ],
});

// Common styles for all PDF documents
export const commonStyles = StyleSheet.create({
  page: {
    flexDirection: 'column',
    backgroundColor: '#FFFFFF',
    padding: 40,
    fontFamily: 'Inter',
    fontSize: 10,
    lineHeight: 1.4,
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 30,
    paddingBottom: 20,
    borderBottomWidth: 2,
    borderBottomColor: '#E5E7EB',
  },
  logo: {
    width: 120,
    height: 40,
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#1F2937',
  },
  subtitle: {
    fontSize: 14,
    color: '#6B7280',
    marginTop: 4,
  },
  section: {
    marginBottom: 20,
  },
  sectionTitle: {
    fontSize: 14,
    fontWeight: 'bold',
    color: '#374151',
    marginBottom: 10,
    paddingBottom: 5,
    borderBottomWidth: 1,
    borderBottomColor: '#E5E7EB',
  },
  row: {
    flexDirection: 'row',
    marginBottom: 8,
  },
  label: {
    width: '30%',
    fontSize: 10,
    fontWeight: 'medium',
    color: '#6B7280',
  },
  value: {
    width: '70%',
    fontSize: 10,
    color: '#1F2937',
  },
  table: {
    display: 'table',
    width: 'auto',
    borderStyle: 'solid',
    borderWidth: 1,
    borderColor: '#E5E7EB',
    borderRightWidth: 0,
    borderBottomWidth: 0,
  },
  tableRow: {
    margin: 'auto',
    flexDirection: 'row',
  },
  tableColHeader: {
    width: '25%',
    borderStyle: 'solid',
    borderWidth: 1,
    borderColor: '#E5E7EB',
    borderLeftWidth: 0,
    borderTopWidth: 0,
    backgroundColor: '#F9FAFB',
    padding: 8,
  },
  tableCol: {
    width: '25%',
    borderStyle: 'solid',
    borderWidth: 1,
    borderColor: '#E5E7EB',
    borderLeftWidth: 0,
    borderTopWidth: 0,
    padding: 8,
  },
  tableCellHeader: {
    fontSize: 10,
    fontWeight: 'bold',
    color: '#374151',
  },
  tableCell: {
    fontSize: 9,
    color: '#1F2937',
  },
  footer: {
    position: 'absolute',
    bottom: 30,
    left: 40,
    right: 40,
    textAlign: 'center',
    color: '#9CA3AF',
    fontSize: 8,
    borderTopWidth: 1,
    borderTopColor: '#E5E7EB',
    paddingTop: 10,
  },
});

// Base document component
interface BasePDFDocumentProps {
  title: string;
  subtitle?: string;
  children: React.ReactNode;
  showHeader?: boolean;
  showFooter?: boolean;
}

export function BasePDFDocument({
  title,
  subtitle,
  children,
  showHeader = true,
  showFooter = true,
}: BasePDFDocumentProps) {
  return (
    <Document>
      <Page size="A4" style={commonStyles.page}>
        {showHeader && (
          <View style={commonStyles.header}>
            <View>
              <Image
                style={commonStyles.logo}
                src="/images/liyali-logo.png"
              />
            </View>
            <View style={{ alignItems: 'flex-end' }}>
              <Text style={commonStyles.title}>{title}</Text>
              {subtitle && (
                <Text style={commonStyles.subtitle}>{subtitle}</Text>
              )}
            </View>
          </View>
        )}

        <View style={{ flex: 1 }}>
          {children}
        </View>

        {showFooter && (
          <Text style={commonStyles.footer}>
            Generated on {new Date().toLocaleDateString()} | Liyali Gateway System
          </Text>
        )}
      </Page>
    </Document>
  );
}
```

### Requisition PDF Template

```typescript
// src/components/pdf/requisition-pdf.tsx
import { BasePDFDocument, commonStyles } from './base-pdf-document';
import { View, Text } from '@react-pdf/renderer';

interface RequisitionPDFProps {
  requisition: Requisition;
  approvalHistory?: ApprovalLogEntry[];
}

export function RequisitionPDF({ 
  requisition, 
  approvalHistory = [] 
}: RequisitionPDFProps) {
  return (
    <BasePDFDocument
      title="Purchase Requisition"
      subtitle={`REQ-${requisition.documentNumber}`}
    >
      {/* Document Information */}
      <View style={commonStyles.section}>
        <Text style={commonStyles.sectionTitle}>Document Information</Text>
        <View style={commonStyles.row}>
          <Text style={commonStyles.label}>Requisition Number:</Text>
          <Text style={commonStyles.value}>REQ-{requisition.documentNumber}</Text>
        </View>
        <View style={commonStyles.row}>
          <Text style={commonStyles.label}>Date Created:</Text>
          <Text style={commonStyles.value}>
            {new Date(requisition.createdAt).toLocaleDateString()}
          </Text>
        </View>
        <View style={commonStyles.row}>
          <Text style={commonStyles.label}>Status:</Text>
          <Text style={commonStyles.value}>{requisition.status}</Text>
        </View>
        <View style={commonStyles.row}>
          <Text style={commonStyles.label}>Priority:</Text>
          <Text style={commonStyles.value}>{requisition.priority || 'Normal'}</Text>
        </View>
      </View>

      {/* Requester Information */}
      <View style={commonStyles.section}>
        <Text style={commonStyles.sectionTitle}>Requester Information</Text>
        <View style={commonStyles.row}>
          <Text style={commonStyles.label}>Requested By:</Text>
          <Text style={commonStyles.value}>{requisition.requestedBy}</Text>
        </View>
        <View style={commonStyles.row}>
          <Text style={commonStyles.label}>Department:</Text>
          <Text style={commonStyles.value}>{requisition.department}</Text>
        </View>
        <View style={commonStyles.row}>
          <Text style={commonStyles.label}>Budget Code:</Text>
          <Text style={commonStyles.value}>{requisition.budgetCode || 'N/A'}</Text>
        </View>
      </View>

      {/* Items Table */}
      <View style={commonStyles.section}>
        <Text style={commonStyles.sectionTitle}>Requested Items</Text>
        <View style={commonStyles.table}>
          {/* Table Header */}
          <View style={commonStyles.tableRow}>
            <View style={[commonStyles.tableColHeader, { width: '40%' }]}>
              <Text style={commonStyles.tableCellHeader}>Description</Text>
            </View>
            <View style={[commonStyles.tableColHeader, { width: '15%' }]}>
              <Text style={commonStyles.tableCellHeader}>Quantity</Text>
            </View>
            <View style={[commonStyles.tableColHeader, { width: '15%' }]}>
              <Text style={commonStyles.tableCellHeader}>Unit</Text>
            </View>
            <View style={[commonStyles.tableColHeader, { width: '15%' }]}>
              <Text style={commonStyles.tableCellHeader}>Unit Price</Text>
            </View>
            <View style={[commonStyles.tableColHeader, { width: '15%' }]}>
              <Text style={commonStyles.tableCellHeader}>Total</Text>
            </View>
          </View>

          {/* Table Rows */}
          {requisition.items.map((item, index) => (
            <View key={index} style={commonStyles.tableRow}>
              <View style={[commonStyles.tableCol, { width: '40%' }]}>
                <Text style={commonStyles.tableCell}>{item.description}</Text>
              </View>
              <View style={[commonStyles.tableCol, { width: '15%' }]}>
                <Text style={commonStyles.tableCell}>{item.quantity}</Text>
              </View>
              <View style={[commonStyles.tableCol, { width: '15%' }]}>
                <Text style={commonStyles.tableCell}>{item.unit}</Text>
              </View>
              <View style={[commonStyles.tableCol, { width: '15%' }]}>
                <Text style={commonStyles.tableCell}>
                  ${item.unitPrice?.toFixed(2) || '0.00'}
                </Text>
              </View>
              <View style={[commonStyles.tableCol, { width: '15%' }]}>
                <Text style={commonStyles.tableCell}>
                  ${((item.quantity || 0) * (item.unitPrice || 0)).toFixed(2)}
                </Text>
              </View>
            </View>
          ))}

          {/* Total Row */}
          <View style={commonStyles.tableRow}>
            <View style={[commonStyles.tableCol, { width: '70%' }]}>
              <Text style={[commonStyles.tableCell, { fontWeight: 'bold' }]}>
                Total Amount:
              </Text>
            </View>
            <View style={[commonStyles.tableCol, { width: '30%' }]}>
              <Text style={[commonStyles.tableCell, { fontWeight: 'bold' }]}>
                ${requisition.totalAmount?.toFixed(2) || '0.00'}
              </Text>
            </View>
          </View>
        </View>
      </View>

      {/* Justification */}
      {requisition.justification && (
        <View style={commonStyles.section}>
          <Text style={commonStyles.sectionTitle}>Justification</Text>
          <Text style={commonStyles.value}>{requisition.justification}</Text>
        </View>
      )}

      {/* Approval History */}
      {approvalHistory.length > 0 && (
        <View style={commonStyles.section}>
          <Text style={commonStyles.sectionTitle}>Approval History</Text>
          <View style={commonStyles.table}>
            <View style={commonStyles.tableRow}>
              <View style={[commonStyles.tableColHeader, { width: '25%' }]}>
                <Text style={commonStyles.tableCellHeader}>Date</Text>
              </View>
              <View style={[commonStyles.tableColHeader, { width: '25%' }]}>
                <Text style={commonStyles.tableCellHeader}>Approver</Text>
              </View>
              <View style={[commonStyles.tableColHeader, { width: '20%' }]}>
                <Text style={commonStyles.tableCellHeader}>Action</Text>
              </View>
              <View style={[commonStyles.tableColHeader, { width: '30%' }]}>
                <Text style={commonStyles.tableCellHeader}>Comments</Text>
              </View>
            </View>

            {approvalHistory.map((entry, index) => (
              <View key={index} style={commonStyles.tableRow}>
                <View style={[commonStyles.tableCol, { width: '25%' }]}>
                  <Text style={commonStyles.tableCell}>
                    {new Date(entry.timestamp).toLocaleDateString()}
                  </Text>
                </View>
                <View style={[commonStyles.tableCol, { width: '25%' }]}>
                  <Text style={commonStyles.tableCell}>{entry.approverName}</Text>
                </View>
                <View style={[commonStyles.tableCol, { width: '20%' }]}>
                  <Text style={commonStyles.tableCell}>{entry.action}</Text>
                </View>
                <View style={[commonStyles.tableCol, { width: '30%' }]}>
                  <Text style={commonStyles.tableCell}>
                    {entry.comments || 'No comments'}
                  </Text>
                </View>
              </View>
            ))}
          </View>
        </View>
      )}
    </BasePDFDocument>
  );
}
```

### Purchase Order PDF Template

```typescript
// src/components/pdf/purchase-order-pdf.tsx
export function PurchaseOrderPDF({ 
  purchaseOrder, 
  vendor 
}: { 
  purchaseOrder: PurchaseOrder; 
  vendor?: Vendor; 
}) {
  return (
    <BasePDFDocument
      title="Purchase Order"
      subtitle={`PO-${purchaseOrder.documentNumber}`}
    >
      {/* PO Information */}
      <View style={commonStyles.section}>
        <Text style={commonStyles.sectionTitle}>Purchase Order Information</Text>
        <View style={commonStyles.row}>
          <Text style={commonStyles.label}>PO Number:</Text>
          <Text style={commonStyles.value}>PO-{purchaseOrder.documentNumber}</Text>
        </View>
        <View style={commonStyles.row}>
          <Text style={commonStyles.label}>Date:</Text>
          <Text style={commonStyles.value}>
            {new Date(purchaseOrder.createdAt).toLocaleDateString()}
          </Text>
        </View>
        <View style={commonStyles.row}>
          <Text style={commonStyles.label}>Expected Delivery:</Text>
          <Text style={commonStyles.value}>
            {purchaseOrder.expectedDeliveryDate 
              ? new Date(purchaseOrder.expectedDeliveryDate).toLocaleDateString()
              : 'TBD'
            }
          </Text>
        </View>
      </View>

      {/* Vendor Information */}
      {vendor && (
        <View style={commonStyles.section}>
          <Text style={commonStyles.sectionTitle}>Vendor Information</Text>
          <View style={commonStyles.row}>
            <Text style={commonStyles.label}>Vendor Name:</Text>
            <Text style={commonStyles.value}>{vendor.name}</Text>
          </View>
          <View style={commonStyles.row}>
            <Text style={commonStyles.label}>Contact Person:</Text>
            <Text style={commonStyles.value}>{vendor.contactPerson || 'N/A'}</Text>
          </View>
          <View style={commonStyles.row}>
            <Text style={commonStyles.label}>Email:</Text>
            <Text style={commonStyles.value}>{vendor.email || 'N/A'}</Text>
          </View>
          <View style={commonStyles.row}>
            <Text style={commonStyles.label}>Phone:</Text>
            <Text style={commonStyles.value}>{vendor.phone || 'N/A'}</Text>
          </View>
          {vendor.address && (
            <View style={commonStyles.row}>
              <Text style={commonStyles.label}>Address:</Text>
              <Text style={commonStyles.value}>{vendor.address}</Text>
            </View>
          )}
        </View>
      )}

      {/* Items Table - Similar to requisition but with vendor-specific details */}
      <View style={commonStyles.section}>
        <Text style={commonStyles.sectionTitle}>Order Items</Text>
        <View style={commonStyles.table}>
          <View style={commonStyles.tableRow}>
            <View style={[commonStyles.tableColHeader, { width: '35%' }]}>
              <Text style={commonStyles.tableCellHeader}>Description</Text>
            </View>
            <View style={[commonStyles.tableColHeader, { width: '15%' }]}>
              <Text style={commonStyles.tableCellHeader}>SKU</Text>
            </View>
            <View style={[commonStyles.tableColHeader, { width: '10%' }]}>
              <Text style={commonStyles.tableCellHeader}>Qty</Text>
            </View>
            <View style={[commonStyles.tableColHeader, { width: '15%' }]}>
              <Text style={commonStyles.tableCellHeader}>Unit Price</Text>
            </View>
            <View style={[commonStyles.tableColHeader, { width: '10%' }]}>
              <Text style={commonStyles.tableCellHeader}>Tax</Text>
            </View>
            <View style={[commonStyles.tableColHeader, { width: '15%' }]}>
              <Text style={commonStyles.tableCellHeader}>Total</Text>
            </View>
          </View>

          {purchaseOrder.items.map((item, index) => (
            <View key={index} style={commonStyles.tableRow}>
              <View style={[commonStyles.tableCol, { width: '35%' }]}>
                <Text style={commonStyles.tableCell}>{item.description}</Text>
              </View>
              <View style={[commonStyles.tableCol, { width: '15%' }]}>
                <Text style={commonStyles.tableCell}>{item.sku || 'N/A'}</Text>
              </View>
              <View style={[commonStyles.tableCol, { width: '10%' }]}>
                <Text style={commonStyles.tableCell}>{item.quantity}</Text>
              </View>
              <View style={[commonStyles.tableCol, { width: '15%' }]}>
                <Text style={commonStyles.tableCell}>
                  ${item.unitPrice?.toFixed(2) || '0.00'}
                </Text>
              </View>
              <View style={[commonStyles.tableCol, { width: '10%' }]}>
                <Text style={commonStyles.tableCell}>
                  {item.taxRate ? `${item.taxRate}%` : '0%'}
                </Text>
              </View>
              <View style={[commonStyles.tableCol, { width: '15%' }]}>
                <Text style={commonStyles.tableCell}>
                  ${((item.quantity || 0) * (item.unitPrice || 0) * (1 + (item.taxRate || 0) / 100)).toFixed(2)}
                </Text>
              </View>
            </View>
          ))}
        </View>
      </View>

      {/* Terms and Conditions */}
      <View style={commonStyles.section}>
        <Text style={commonStyles.sectionTitle}>Terms and Conditions</Text>
        <Text style={commonStyles.value}>
          1. Payment terms: Net 30 days from invoice date{'\n'}
          2. Delivery must be made to the specified address{'\n'}
          3. All items must meet specified quality standards{'\n'}
          4. Vendor must provide delivery confirmation{'\n'}
          5. Any changes to this order must be approved in writing
        </Text>
      </View>
    </BasePDFDocument>
  );
}
```

## PDF Generation Hooks

### Core PDF Generation Hook

```typescript
// src/hooks/use-pdf-generation.ts
import { pdf } from '@react-pdf/renderer';
import { saveAs } from 'file-saver';

export function usePDFGeneration() {
  const [isGenerating, setIsGenerating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const generatePDF = async (
    component: React.ReactElement,
    filename: string,
    options: {
      download?: boolean;
      preview?: boolean;
      print?: boolean;
    } = { download: true }
  ) => {
    setIsGenerating(true);
    setError(null);

    try {
      const blob = await pdf(component).toBlob();

      if (options.download) {
        saveAs(blob, `${filename}.pdf`);
      }

      if (options.preview) {
        const url = URL.createObjectURL(blob);
        window.open(url, '_blank');
        // Clean up the URL after a delay
        setTimeout(() => URL.revokeObjectURL(url), 1000);
      }

      if (options.print) {
        const url = URL.createObjectURL(blob);
        const printWindow = window.open(url, '_blank');
        printWindow?.addEventListener('load', () => {
          printWindow.print();
        });
      }

      return blob;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to generate PDF';
      setError(errorMessage);
      throw new Error(errorMessage);
    } finally {
      setIsGenerating(false);
    }
  };

  return {
    generatePDF,
    isGenerating,
    error,
  };
}
```

### Document-Specific Hooks

```typescript
// src/hooks/use-requisition-pdf.ts
export function useRequisitionPDF() {
  const { generatePDF, isGenerating, error } = usePDFGeneration();

  const generateRequisitionPDF = async (
    requisition: Requisition,
    options?: {
      includeApprovalHistory?: boolean;
      download?: boolean;
      preview?: boolean;
    }
  ) => {
    const approvalHistory = options?.includeApprovalHistory 
      ? await fetchApprovalHistory(requisition.id)
      : [];

    const component = (
      <RequisitionPDF 
        requisition={requisition}
        approvalHistory={approvalHistory}
      />
    );

    const filename = `requisition-${requisition.documentNumber}`;
    
    return generatePDF(component, filename, {
      download: options?.download ?? true,
      preview: options?.preview ?? false,
    });
  };

  return {
    generateRequisitionPDF,
    isGenerating,
    error,
  };
}

// Similar hooks for other document types
export function usePurchaseOrderPDF() {
  const { generatePDF, isGenerating, error } = usePDFGeneration();

  const generatePurchaseOrderPDF = async (
    purchaseOrder: PurchaseOrder,
    options?: { download?: boolean; preview?: boolean }
  ) => {
    const vendor = await fetchVendorById(purchaseOrder.vendorId);
    
    const component = (
      <PurchaseOrderPDF 
        purchaseOrder={purchaseOrder}
        vendor={vendor}
      />
    );

    const filename = `purchase-order-${purchaseOrder.documentNumber}`;
    
    return generatePDF(component, filename, options);
  };

  return {
    generatePurchaseOrderPDF,
    isGenerating,
    error,
  };
}
```

## PDF Preview Component

### Live Preview with Real-time Updates

```typescript
// src/components/pdf/pdf-preview.tsx
import { PDFViewer } from '@react-pdf/renderer';

interface PDFPreviewProps {
  component: React.ReactElement;
  width?: string | number;
  height?: string | number;
  showToolbar?: boolean;
}

export function PDFPreview({
  component,
  width = '100%',
  height = 600,
  showToolbar = true,
}: PDFPreviewProps) {
  return (
    <div className="border rounded-lg overflow-hidden">
      <PDFViewer
        width={width}
        height={height}
        showToolbar={showToolbar}
        style={{
          border: 'none',
        }}
      >
        {component}
      </PDFViewer>
    </div>
  );
}

// Usage in forms for live preview
export function RequisitionFormWithPreview() {
  const [formData, setFormData] = useState<Partial<Requisition>>({});
  const [showPreview, setShowPreview] = useState(false);

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
      {/* Form */}
      <div>
        <RequisitionForm
          data={formData}
          onChange={setFormData}
        />
      </div>

      {/* Preview */}
      {showPreview && (
        <div>
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold">PDF Preview</h3>
            <Button
              variant="outline"
              onClick={() => setShowPreview(false)}
            >
              Hide Preview
            </Button>
          </div>
          
          <PDFPreview
            component={
              <RequisitionPDF 
                requisition={formData as Requisition}
              />
            }
            height={500}
          />
        </div>
      )}

      {!showPreview && (
        <div className="flex items-center justify-center border-2 border-dashed border-gray-300 rounded-lg h-96">
          <Button onClick={() => setShowPreview(true)}>
            Show PDF Preview
          </Button>
        </div>
      )}
    </div>
  );
}
```

## Bulk PDF Generation

### Batch Processing for Multiple Documents

```typescript
// src/hooks/use-bulk-pdf-generation.ts
export function useBulkPDFGeneration() {
  const [progress, setProgress] = useState(0);
  const [isGenerating, setIsGenerating] = useState(false);
  const [errors, setErrors] = useState<string[]>([]);

  const generateBulkPDFs = async <T>(
    items: T[],
    generateComponent: (item: T) => React.ReactElement,
    getFilename: (item: T) => string,
    options: {
      zipFilename?: string;
      onProgress?: (current: number, total: number) => void;
    } = {}
  ) => {
    setIsGenerating(true);
    setProgress(0);
    setErrors([]);

    const JSZip = (await import('jszip')).default;
    const zip = new JSZip();
    const generatedErrors: string[] = [];

    try {
      for (let i = 0; i < items.length; i++) {
        const item = items[i];
        
        try {
          const component = generateComponent(item);
          const blob = await pdf(component).toBlob();
          const filename = `${getFilename(item)}.pdf`;
          
          zip.file(filename, blob);
          
          const currentProgress = ((i + 1) / items.length) * 100;
          setProgress(currentProgress);
          options.onProgress?.(i + 1, items.length);
          
        } catch (error) {
          const errorMessage = `Failed to generate PDF for ${getFilename(item)}: ${
            error instanceof Error ? error.message : 'Unknown error'
          }`;
          generatedErrors.push(errorMessage);
        }
      }

      // Generate and download zip file
      const zipBlob = await zip.generateAsync({ type: 'blob' });
      const zipFilename = options.zipFilename || `bulk-pdfs-${Date.now()}.zip`;
      saveAs(zipBlob, zipFilename);

      setErrors(generatedErrors);
      return {
        success: true,
        totalGenerated: items.length - generatedErrors.length,
        errors: generatedErrors,
      };

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Bulk generation failed';
      setErrors([errorMessage]);
      throw new Error(errorMessage);
    } finally {
      setIsGenerating(false);
      setProgress(0);
    }
  };

  return {
    generateBulkPDFs,
    progress,
    isGenerating,
    errors,
  };
}

// Usage for bulk requisition PDFs
export function BulkRequisitionPDFGenerator({ 
  requisitions 
}: { 
  requisitions: Requisition[] 
}) {
  const { generateBulkPDFs, progress, isGenerating, errors } = useBulkPDFGeneration();

  const handleBulkGenerate = async () => {
    await generateBulkPDFs(
      requisitions,
      (requisition) => <RequisitionPDF requisition={requisition} />,
      (requisition) => `requisition-${requisition.documentNumber}`,
      {
        zipFilename: `requisitions-${new Date().toISOString().split('T')[0]}.zip`,
        onProgress: (current, total) => {
          console.log(`Generated ${current} of ${total} PDFs`);
        },
      }
    );
  };

  return (
    <div className="space-y-4">
      <Button
        onClick={handleBulkGenerate}
        disabled={isGenerating || requisitions.length === 0}
        className="w-full"
      >
        {isGenerating ? (
          <>
            <Loader2 className="w-4 h-4 mr-2 animate-spin" />
            Generating PDFs... ({Math.round(progress)}%)
          </>
        ) : (
          <>
            <Download className="w-4 h-4 mr-2" />
            Generate {requisitions.length} PDFs
          </>
        )}
      </Button>

      {isGenerating && (
        <div className="space-y-2">
          <div className="w-full bg-gray-200 rounded-full h-2">
            <div
              className="bg-blue-600 h-2 rounded-full transition-all duration-300"
              style={{ width: `${progress}%` }}
            />
          </div>
          <p className="text-sm text-gray-600 text-center">
            {Math.round(progress)}% complete
          </p>
        </div>
      )}

      {errors.length > 0 && (
        <Alert variant="destructive">
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>Generation Errors</AlertTitle>
          <AlertDescription>
            <ul className="list-disc list-inside space-y-1">
              {errors.map((error, index) => (
                <li key={index} className="text-sm">{error}</li>
              ))}
            </ul>
          </AlertDescription>
        </Alert>
      )}
    </div>
  );
}
```

## Template Customization

### Dynamic Template System

```typescript
// src/lib/pdf-templates.ts
export interface PDFTemplate {
  id: string;
  name: string;
  description: string;
  documentType: 'requisition' | 'purchase_order' | 'payment_voucher';
  component: React.ComponentType<any>;
  settings: {
    showLogo: boolean;
    showHeader: boolean;
    showFooter: boolean;
    includeApprovalHistory: boolean;
    colorScheme: 'default' | 'blue' | 'green' | 'purple';
    fontSize: 'small' | 'medium' | 'large';
  };
}

export const PDF_TEMPLATES: PDFTemplate[] = [
  {
    id: 'requisition-standard',
    name: 'Standard Requisition',
    description: 'Standard requisition template with all details',
    documentType: 'requisition',
    component: RequisitionPDF,
    settings: {
      showLogo: true,
      showHeader: true,
      showFooter: true,
      includeApprovalHistory: true,
      colorScheme: 'default',
      fontSize: 'medium',
    },
  },
  {
    id: 'requisition-minimal',
    name: 'Minimal Requisition',
    description: 'Simplified requisition template',
    documentType: 'requisition',
    component: MinimalRequisitionPDF,
    settings: {
      showLogo: false,
      showHeader: true,
      showFooter: false,
      includeApprovalHistory: false,
      colorScheme: 'default',
      fontSize: 'small',
    },
  },
];

export function getTemplate(templateId: string): PDFTemplate | undefined {
  return PDF_TEMPLATES.find(template => template.id === templateId);
}

export function getTemplatesForDocumentType(
  documentType: PDFTemplate['documentType']
): PDFTemplate[] {
  return PDF_TEMPLATES.filter(template => template.documentType === documentType);
}
```

### Template Selection Component

```typescript
// src/components/pdf/template-selector.tsx
interface TemplateSelectorProps {
  documentType: PDFTemplate['documentType'];
  selectedTemplateId?: string;
  onTemplateSelect: (templateId: string) => void;
}

export function TemplateSelector({
  documentType,
  selectedTemplateId,
  onTemplateSelect,
}: TemplateSelectorProps) {
  const templates = getTemplatesForDocumentType(documentType);

  return (
    <div className="space-y-4">
      <Label className="text-sm font-medium">PDF Template</Label>
      
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {templates.map(template => (
          <Card
            key={template.id}
            className={cn(
              "cursor-pointer transition-all hover:shadow-md",
              selectedTemplateId === template.id && "ring-2 ring-primary"
            )}
            onClick={() => onTemplateSelect(template.id)}
          >
            <CardContent className="p-4">
              <div className="flex items-start justify-between">
                <div className="space-y-2">
                  <h4 className="font-medium">{template.name}</h4>
                  <p className="text-sm text-muted-foreground">
                    {template.description}
                  </p>
                </div>
                
                {selectedTemplateId === template.id && (
                  <CheckCircle className="w-5 h-5 text-primary" />
                )}
              </div>

              <div className="mt-3 flex flex-wrap gap-1">
                {template.settings.showLogo && (
                  <Badge variant="outline" className="text-xs">Logo</Badge>
                )}
                {template.settings.includeApprovalHistory && (
                  <Badge variant="outline" className="text-xs">History</Badge>
                )}
                <Badge variant="outline" className="text-xs">
                  {template.settings.fontSize}
                </Badge>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
```

## Performance Optimization

### PDF Generation Optimization

```typescript
// src/lib/pdf-optimization.ts
export class PDFOptimizer {
  private static cache = new Map<string, Blob>();

  // Cache generated PDFs to avoid regeneration
  static async getCachedPDF(
    cacheKey: string,
    generator: () => Promise<Blob>
  ): Promise<Blob> {
    if (this.cache.has(cacheKey)) {
      return this.cache.get(cacheKey)!;
    }

    const blob = await generator();
    this.cache.set(cacheKey, blob);
    
    // Clean up cache after 5 minutes
    setTimeout(() => {
      this.cache.delete(cacheKey);
    }, 5 * 60 * 1000);

    return blob;
  }

  // Generate cache key based on document data
  static generateCacheKey(documentType: string, documentId: string, version: string): string {
    return `${documentType}-${documentId}-${version}`;
  }

  // Optimize images for PDF
  static async optimizeImage(imageUrl: string): Promise<string> {
    // In a real implementation, this would compress/resize images
    return imageUrl;
  }
}

// Usage in PDF generation
export function useOptimizedPDFGeneration() {
  const { generatePDF, isGenerating, error } = usePDFGeneration();

  const generateOptimizedPDF = async (
    component: React.ReactElement,
    filename: string,
    cacheKey?: string,
    options?: any
  ) => {
    if (cacheKey) {
      return PDFOptimizer.getCachedPDF(cacheKey, async () => {
        return pdf(component).toBlob();
      });
    }

    return generatePDF(component, filename, options);
  };

  return {
    generateOptimizedPDF,
    isGenerating,
    error,
  };
}
```

The PDF generation system provides a comprehensive solution for creating professional documents with customizable templates, real-time previews, and efficient batch processing capabilities.