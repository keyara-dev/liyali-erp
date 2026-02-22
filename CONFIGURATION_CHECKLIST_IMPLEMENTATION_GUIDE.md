# Configuration Checklist - Quick Implementation Guide

## Overview

This guide shows you how to add configuration requirement banners to any document creation or submission form in the application.

## Quick Start

### Step 1: Import Required Components

```typescript
import { useConfigurationStatus } from "@/hooks/use-configuration-status";
import { ConfigurationChecklistBanner } from "@/components/ui/configuration-checklist-banner";
import { WorkflowRequirementBanner } from "@/components/ui/workflow-requirement-banner";
```

### Step 2: Add Configuration Check

```typescript
// For creation forms (departments, categories, budgets)
const configStatus = useConfigurationStatus({
  includeWorkflow: false,
});

// For submission forms (includes workflow requirement)
const configStatus = useConfigurationStatus({
  includeWorkflow: true,
  workflowEntityType: "requisition", // or "budget", "purchase_order", etc.
});
```

### Step 3: Display Banner

```typescript
// In your form JSX
{!configStatus.allConfigured && (
  <ConfigurationChecklistBanner
    requirements={configStatus.requirements}
    variant="creation" // or "submission"
  />
)}
```

### Step 4: Disable Submit Button (Optional)

```typescript
<Button
  onClick={handleSubmit}
  disabled={!configStatus.allConfigured}
>
  Create Document
</Button>
```

## Complete Examples

### Example 1: Budget Creation Dialog

```typescript
"use client";

import { useState } from "react";
import { Dialog, DialogContent } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { useConfigurationStatus } from "@/hooks/use-configuration-status";
import { ConfigurationChecklistBanner } from "@/components/ui/configuration-checklist-banner";

export function CreateBudgetDialog({ open, onOpenChange }) {
  // Check configuration status
  const configStatus = useConfigurationStatus({
    includeWorkflow: false,
  });

  const handleSubmit = () => {
    if (!configStatus.allConfigured) {
      return; // Prevent submission
    }
    // Submit logic here
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <div className="space-y-4">
          {/* Configuration Banner */}
          {!configStatus.allConfigured && (
            <ConfigurationChecklistBanner
              requirements={configStatus.requirements}
              variant="creation"
              title="Budget Configuration Required"
            />
  