'use client'

import { useRef, useEffect, useState } from 'react'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { RotateCcw, Download } from 'lucide-react'

interface SignatureCanvasProps {
  onSignatureChange?: (signatureData: string) => void
  disabled?: boolean
}

export function SignatureCanvas({ onSignatureChange, disabled }: SignatureCanvasProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const [isDrawing, setIsDrawing] = useState(false)
  const [hasSignature, setHasSignature] = useState(false)

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return

    // Set canvas size
    const rect = canvas.parentElement?.getBoundingClientRect()
    if (rect) {
      canvas.width = rect.width
      canvas.height = 200
    }

    // Set up canvas styling
    const ctx = canvas.getContext('2d')
    if (ctx) {
      ctx.lineCap = 'round'
      ctx.lineJoin = 'round'
      ctx.lineWidth = 2
      ctx.strokeStyle = '#000000'
    }
  }, [])

  const startDrawing = (e: React.MouseEvent<HTMLCanvasElement>) => {
    if (disabled) return

    const canvas = canvasRef.current
    if (!canvas) return

    const rect = canvas.getBoundingClientRect()
    const x = e.clientX - rect.left
    const y = e.clientY - rect.top

    const ctx = canvas.getContext('2d')
    if (ctx) {
      ctx.beginPath()
      ctx.moveTo(x, y)
      setIsDrawing(true)
    }
  }

  const draw = (e: React.MouseEvent<HTMLCanvasElement>) => {
    if (!isDrawing || disabled) return

    const canvas = canvasRef.current
    if (!canvas) return

    const rect = canvas.getBoundingClientRect()
    const x = e.clientX - rect.left
    const y = e.clientY - rect.top

    const ctx = canvas.getContext('2d')
    if (ctx) {
      ctx.lineTo(x, y)
      ctx.stroke()
      setHasSignature(true)
    }
  }

  const stopDrawing = () => {
    setIsDrawing(false)
  }

  const clearSignature = () => {
    const canvas = canvasRef.current
    if (!canvas) return

    const ctx = canvas.getContext('2d')
    if (ctx) {
      ctx.clearRect(0, 0, canvas.width, canvas.height)
      setHasSignature(false)
      onSignatureChange?.('')
    }
  }

  const saveSignature = () => {
    const canvas = canvasRef.current
    if (!canvas) return

    const signatureData = canvas.toDataURL('image/png')
    onSignatureChange?.(signatureData)
  }

  return (
    <div className="space-y-3">
      <Label>Digital Signature *</Label>
      <div className="border-2 border-dashed rounded-lg bg-gray-50 overflow-hidden">
        <canvas
          ref={canvasRef}
          onMouseDown={startDrawing}
          onMouseMove={draw}
          onMouseUp={stopDrawing}
          onMouseLeave={stopDrawing}
          className={`w-full cursor-${disabled ? 'not-allowed' : 'crosshair'} block`}
          style={{ touchAction: 'none' }}
        />
      </div>
      <p className="text-xs text-muted-foreground">
        Sign above using your mouse. A signature is required to approve.
      </p>
      <div className="flex gap-2">
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={clearSignature}
          disabled={!hasSignature || disabled}
          className="gap-2"
        >
          <RotateCcw className="h-4 w-4" />
          Clear
        </Button>
        <Button
          type="button"
          size="sm"
          onClick={saveSignature}
          disabled={!hasSignature || disabled}
          className="gap-2"
        >
          <Download className="h-4 w-4" />
          Save Signature
        </Button>
      </div>
    </div>
  )
}
