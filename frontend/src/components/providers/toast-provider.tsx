'use client'

import { Toaster } from 'sonner'

/**
 * Toast Provider Component
 * Provides Sonner toasts throughout the application
 * Add this to your root layout
 */
export function ToastProvider() {
  return (
    <Toaster
      position="top-right"
      expand
      richColors
      theme="system"
      closeButton
    />
  )
}
