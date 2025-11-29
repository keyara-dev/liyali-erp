# Color Implementation Examples

## Real Component Examples Using New Color Scheme

### 1. Approval Status Badges

```jsx
// Status Badge Component
export function StatusBadge({ status }) {
  const statusConfig = {
    DRAFT: {
      bg: 'bg-gray-100',
      text: 'text-gray-800',
      label: 'Draft',
    },
    SUBMITTED: {
      bg: 'bg-primary/20',
      text: 'text-primary',
      label: 'Submitted',
    },
    IN_APPROVAL: {
      bg: 'bg-accent/20',
      text: 'text-accent-foreground',
      label: 'In Approval',
    },
    APPROVED: {
      bg: 'bg-secondary/20',
      text: 'text-secondary',
      label: 'Approved',
    },
    REJECTED: {
      bg: 'bg-destructive/20',
      text: 'text-destructive',
      label: 'Rejected',
    },
  }

  const config = statusConfig[status] || statusConfig.DRAFT

  return (
    <span className={`px-3 py-1 rounded-full text-sm font-medium ${config.bg} ${config.text}`}>
      {config.label}
    </span>
  )
}

// Usage
<StatusBadge status="APPROVED" />
// Renders: Green background with green text "Approved"

<StatusBadge status="IN_APPROVAL" />
// Renders: Amber background with dark text "In Approval"
```

---

### 2. Action Buttons in Approval Dialog

```jsx
// Approval Dialog Buttons
export function ApprovalActionButtons({ onApprove, onReject, onPending }) {
  return (
    <div className="flex gap-3">
      {/* Approve Button - Primary Green */}
      <button
        onClick={onApprove}
        className="flex-1 bg-secondary text-secondary-foreground hover:bg-secondary/90
                   px-4 py-2 rounded-lg font-semibold transition-colors
                   flex items-center justify-center gap-2"
      >
        <CheckCircle className="w-4 h-4" />
        Approve
      </button>

      {/* Pending/Review Button - Primary Blue */}
      <button
        onClick={onPending}
        className="flex-1 bg-primary text-primary-foreground hover:bg-primary/90
                   px-4 py-2 rounded-lg font-semibold transition-colors
                   flex items-center justify-center gap-2"
      >
        <Clock className="w-4 h-4" />
        Send for Review
      </button>

      {/* Reject Button - Destructive Red */}
      <button
        onClick={onReject}
        className="flex-1 bg-destructive text-destructive-foreground hover:bg-destructive/90
                   px-4 py-2 rounded-lg font-semibold transition-colors
                   flex items-center justify-center gap-2"
      >
        <XCircle className="w-4 h-4" />
        Reject
      </button>
    </div>
  )
}
```

---

### 3. Document Status Timeline

```jsx
// Workflow Stage Indicator
export function WorkflowTimeline({ currentStage, totalStages, stageName }) {
  const stages = ['Submitted', 'Department Head', 'Auditor', 'Finance Director', 'Principal Officer']

  return (
    <div className="space-y-4">
      {stages.map((stage, index) => {
        const stageNum = index + 1
        const isCompleted = stageNum < currentStage
        const isCurrent = stageNum === currentStage
        const isPending = stageNum > currentStage

        return (
          <div key={stage} className="flex items-center gap-4">
            {/* Stage Circle */}
            <div
              className={`w-10 h-10 rounded-full flex items-center justify-center font-bold text-sm
                          ${isCompleted ? 'bg-secondary text-secondary-foreground' : ''}
                          ${isCurrent ? 'bg-primary text-primary-foreground ring-2 ring-primary ring-offset-2' : ''}
                          ${isPending ? 'bg-muted text-muted-foreground' : ''}`}
            >
              {isCompleted ? '✓' : stageNum}
            </div>

            {/* Stage Label */}
            <div className="flex-1">
              <p
                className={`font-semibold
                          ${isCompleted ? 'text-secondary' : ''}
                          ${isCurrent ? 'text-primary' : ''}
                          ${isPending ? 'text-muted-foreground' : ''}`}
              >
                {stage}
              </p>
              {isCurrent && <p className="text-sm text-accent">Currently here</p>}
            </div>

            {/* Connecting Line */}
            {stageNum < stages.length && (
              <div
                className={`absolute left-5 w-0.5 h-12 mt-12
                           ${isCompleted ? 'bg-secondary' : 'bg-muted'}`}
              />
            )}
          </div>
        )
      })}
    </div>
  )
}
```

---

### 4. Alert/Warning Component

```jsx
// Alert Component with Color System
export function Alert({ type, title, message, action }) {
  const alertConfig = {
    success: {
      bg: 'bg-secondary/10',
      border: 'border-secondary/30',
      icon: 'text-secondary',
      title: 'text-secondary',
      button: 'bg-secondary text-secondary-foreground hover:bg-secondary/90',
    },
    warning: {
      bg: 'bg-accent/10',
      border: 'border-accent/30',
      icon: 'text-accent',
      title: 'text-accent-foreground',
      button: 'bg-accent text-accent-foreground hover:bg-accent/90',
    },
    error: {
      bg: 'bg-destructive/10',
      border: 'border-destructive/30',
      icon: 'text-destructive',
      title: 'text-destructive',
      button: 'bg-destructive text-destructive-foreground hover:bg-destructive/90',
    },
    info: {
      bg: 'bg-primary/10',
      border: 'border-primary/30',
      icon: 'text-primary',
      title: 'text-primary',
      button: 'bg-primary text-primary-foreground hover:bg-primary/90',
    },
  }

  const config = alertConfig[type] || alertConfig.info

  return (
    <div className={`p-4 rounded-lg border ${config.bg} ${config.border}`}>
      <div className="flex gap-3">
        <div className={`w-5 h-5 ${config.icon} flex-shrink-0`}>
          {type === 'success' && <CheckCircle2 />}
          {type === 'warning' && <AlertCircle />}
          {type === 'error' && <XCircle />}
          {type === 'info' && <Info />}
        </div>

        <div className="flex-1">
          <h3 className={`font-semibold ${config.title}`}>{title}</h3>
          <p className="text-sm text-muted-foreground mt-1">{message}</p>
        </div>

        {action && (
          <button className={`px-3 py-1 rounded text-sm font-medium ${config.button}`}>
            {action.label}
          </button>
        )}
      </div>
    </div>
  )
}

// Usage Examples
<Alert
  type="success"
  title="Document Approved"
  message="Your requisition has been successfully approved by all parties."
  action={{ label: 'View' }}
/>

<Alert
  type="warning"
  title="Action Required"
  message="This purchase order is waiting for your approval at Stage 2."
  action={{ label: 'Review' }}
/>

<Alert
  type="error"
  title="Submission Failed"
  message="Please check your form and try again."
/>

<Alert
  type="info"
  title="Workflow Status"
  message="Your document is currently being reviewed by the auditor."
/>
```

---

### 5. Workflow Document Card

```jsx
// Document List Card with Color System
export function DocumentCard({ document, onView, onDownload, onApprove }) {
  const getStatusColor = (status) => {
    switch (status) {
      case 'APPROVED':
        return { badge: 'bg-secondary/20 text-secondary', text: 'text-secondary' }
      case 'IN_APPROVAL':
        return { badge: 'bg-accent/20 text-accent-foreground', text: 'text-accent' }
      case 'REJECTED':
        return { badge: 'bg-destructive/20 text-destructive', text: 'text-destructive' }
      default:
        return { badge: 'bg-muted text-muted-foreground', text: 'text-muted-foreground' }
    }
  }

  const colors = getStatusColor(document.status)

  return (
    <div className="border border-border rounded-lg p-4 hover:bg-muted/30 transition">
      <div className="flex items-start justify-between mb-3">
        <div>
          <h3 className="font-semibold text-lg">{document.documentNumber}</h3>
          <p className="text-sm text-muted-foreground">{document.type}</p>
        </div>

        {/* Status Badge */}
        <span className={`px-2 py-1 rounded-full text-xs font-medium ${colors.badge}`}>
          {document.status}
        </span>
      </div>

      {/* Details */}
      <div className="grid grid-cols-2 gap-2 mb-4 text-sm">
        <div>
          <p className="text-muted-foreground">Amount</p>
          <p className="font-semibold">K {document.amount.toLocaleString()}</p>
        </div>
        <div>
          <p className="text-muted-foreground">Stage</p>
          <p className="font-semibold">
            {document.currentStage} of {document.totalStages}
          </p>
        </div>
      </div>

      {/* Action Buttons */}
      <div className="flex gap-2">
        {/* View Button - Primary */}
        <button
          onClick={onView}
          className="flex-1 bg-primary text-primary-foreground hover:bg-primary/90
                     px-3 py-2 rounded text-sm font-medium transition"
        >
          View
        </button>

        {/* Download Button - Secondary */}
        <button
          onClick={onDownload}
          className="flex-1 bg-secondary/20 text-secondary hover:bg-secondary/30
                     px-3 py-2 rounded text-sm font-medium transition"
        >
          Download
        </button>

        {/* Approve Button (if needed) - Green */}
        {document.status === 'IN_APPROVAL' && (
          <button
            onClick={onApprove}
            className="flex-1 bg-secondary text-secondary-foreground hover:bg-secondary/90
                       px-3 py-2 rounded text-sm font-medium transition"
          >
            Approve
          </button>
        )}
      </div>
    </div>
  )
}
```

---

### 6. Form Input with Validation

```jsx
// Form Input with Color-Based Validation
export function FormInput({ label, value, error, success, status = 'normal' }) {
  const inputStyles = {
    normal: 'border-border focus:border-primary focus:ring-primary',
    error: 'border-destructive focus:border-destructive focus:ring-destructive',
    success: 'border-secondary focus:border-secondary focus:ring-secondary',
    warning: 'border-accent focus:border-accent focus:ring-accent',
  }

  const statusColor = {
    normal: 'text-muted-foreground',
    error: 'text-destructive',
    success: 'text-secondary',
    warning: 'text-accent',
  }

  return (
    <div className="space-y-2">
      <label className="block text-sm font-medium text-foreground">
        {label}
      </label>

      <input
        type="text"
        value={value}
        className={`w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2
                    transition ${inputStyles[status]}`}
        placeholder="Enter value..."
      />

      {error && (
        <p className="text-sm text-destructive flex items-center gap-1">
          <XCircle className="w-4 h-4" />
          {error}
        </p>
      )}

      {success && (
        <p className="text-sm text-secondary flex items-center gap-1">
          <CheckCircle className="w-4 h-4" />
          {success}
        </p>
      )}
    </div>
  )
}

// Usage
<FormInput
  label="Email"
  value={email}
  status={emailError ? 'error' : 'normal'}
  error={emailError}
/>

<FormInput
  label="Document Number"
  value={docNumber}
  status="success"
  success="Document number verified"
/>
```

---

### 7. Data Table with Color Coding

```jsx
// Table Row with Status Color Coding
export function DocumentTable({ documents }) {
  const getRowColor = (status) => {
    switch (status) {
      case 'APPROVED':
        return 'hover:bg-secondary/5 border-l-4 border-secondary'
      case 'IN_APPROVAL':
        return 'hover:bg-accent/5 border-l-4 border-accent'
      case 'REJECTED':
        return 'hover:bg-destructive/5 border-l-4 border-destructive'
      default:
        return 'hover:bg-muted/30'
    }
  }

  return (
    <table className="w-full">
      <thead className="bg-muted/50 border-b border-border">
        <tr>
          <th className="px-4 py-3 text-left text-sm font-semibold">Document</th>
          <th className="px-4 py-3 text-left text-sm font-semibold">Amount</th>
          <th className="px-4 py-3 text-left text-sm font-semibold">Status</th>
          <th className="px-4 py-3 text-left text-sm font-semibold">Actions</th>
        </tr>
      </thead>
      <tbody>
        {documents.map((doc) => (
          <tr key={doc.id} className={`border-b border-border ${getRowColor(doc.status)}`}>
            <td className="px-4 py-3">
              <span className="font-medium text-primary hover:underline cursor-pointer">
                {doc.documentNumber}
              </span>
            </td>
            <td className="px-4 py-3">K {doc.amount.toLocaleString()}</td>
            <td className="px-4 py-3">
              <span
                className={`px-2 py-1 rounded-full text-xs font-medium
                          ${doc.status === 'APPROVED' ? 'bg-secondary/20 text-secondary' : ''}
                          ${doc.status === 'IN_APPROVAL' ? 'bg-accent/20 text-accent-foreground' : ''}
                          ${doc.status === 'REJECTED' ? 'bg-destructive/20 text-destructive' : ''}`}
              >
                {doc.status}
              </span>
            </td>
            <td className="px-4 py-3">
              <button className="text-primary hover:underline text-sm font-medium">
                View
              </button>
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}
```

---

### 8. Progress Indicator

```jsx
// Multi-Stage Approval Progress
export function ApprovalProgress({ current, total, stages }) {
  return (
    <div className="space-y-2">
      {/* Progress Bar */}
      <div className="w-full bg-muted rounded-full h-2 overflow-hidden">
        <div
          className="bg-primary h-full transition-all duration-500 ease-out"
          style={{ width: `${(current / total) * 100}%` }}
        />
      </div>

      {/* Stage Labels */}
      <div className="flex justify-between text-xs">
        {stages.map((stage, index) => {
          const stageNum = index + 1
          const isCompleted = stageNum < current
          const isCurrent = stageNum === current

          return (
            <div
              key={stage}
              className={`font-medium
                        ${isCompleted ? 'text-secondary' : ''}
                        ${isCurrent ? 'text-primary' : ''}
                        ${stageNum > current ? 'text-muted-foreground' : ''}`}
            >
              {isCompleted && '✓ '}
              {stage}
            </div>
          )
        })}
      </div>

      {/* Status Text */}
      <p className="text-sm text-muted-foreground">
        Stage <span className="font-semibold text-primary">{current}</span> of{' '}
        <span className="font-semibold">{total}</span>
      </p>
    </div>
  )
}

// Usage
<ApprovalProgress
  current={2}
  total={4}
  stages={['Submitted', 'Department', 'Auditor', 'Finance']}
/>
```

---

## Color Accessibility Checklist

When implementing these components:

- [ ] **Contrast**: All text has minimum 4.5:1 contrast ratio
- [ ] **Focus States**: All interactive elements have visible focus rings in primary color
- [ ] **Color Plus Meaning**: Status indicated by both color AND text/icon
- [ ] **Dark Mode**: Test all colors in dark mode
- [ ] **Colorblind Safe**: No blue-red combinations, sufficient saturation
- [ ] **Hover States**: All buttons have distinct hover states
- [ ] **Disabled States**: Clearly differentiated from enabled states

---

## Dark Mode Considerations

All examples above automatically support dark mode because they use CSS custom properties:

```css
/* Light Mode */
--primary: oklch(52.4% 0.21 265.5);

/* Dark Mode (automatically applied) */
.dark {
  --primary: oklch(64% 0.22 265.5);
}
```

No additional dark mode classes needed in most cases. The Tailwind color system handles it automatically.

---

**Last Updated**: November 29, 2024
**Tailwind Version**: v4.x
**Color Space**: OKLCH
