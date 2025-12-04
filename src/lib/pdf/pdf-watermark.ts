/**
 * PDF watermark utilities
 * Add watermarks to PDFs based on document status
 */

export interface WatermarkOptions {
  text: string
  opacity?: number
  fontSize?: number
  fontFamily?: string
  color?: string
  angle?: number
}

/**
 * Get watermark options based on document status
 */
export function getWatermarkByStatus(
  status: string
): WatermarkOptions | null {
  const watermarks: Record<string, WatermarkOptions> = {
    DRAFT: {
      text: 'DRAFT',
      opacity: 0.15,
      fontSize: 72,
      color: '#FF6B6B',
      angle: -45,
    },
    SUBMITTED: {
      text: 'SUBMITTED',
      opacity: 0.12,
      fontSize: 60,
      color: '#FFA500',
      angle: -45,
    },
    IN_REVIEW: {
      text: 'IN REVIEW',
      opacity: 0.12,
      fontSize: 60,
      color: '#FFD93D',
      angle: -45,
    },
    APPROVED: {
      text: 'APPROVED',
      opacity: 0.1,
      fontSize: 60,
      color: '#6BCB77',
      angle: -45,
    },
    PAID: {
      text: 'PAID',
      opacity: 0.1,
      fontSize: 60,
      color: '#4D96FF',
      angle: -45,
    },
    REJECTED: {
      text: 'REJECTED',
      opacity: 0.15,
      fontSize: 60,
      color: '#FF006E',
      angle: -45,
    },
  }

  return watermarks[status] || null
}

/**
 * Create watermark SVG for use in PDFs
 * Returns SVG as data URL
 */
export function createWatermarkSVG(
  options: WatermarkOptions
): string {
  const {
    text,
    opacity = 0.15,
    fontSize = 72,
    fontFamily = 'Arial',
    color = '#CCCCCC',
    angle = -45,
  } = options

  // Create SVG for watermark
  const svg = `
    <svg width="400" height="200" xmlns="http://www.w3.org/2000/svg">
      <defs>
        <style>
          .watermark {
            font-family: ${fontFamily};
            font-size: ${fontSize}px;
            font-weight: bold;
            fill: ${color};
            opacity: ${opacity};
            text-anchor: middle;
            dominant-baseline: middle;
          }
        </style>
      </defs>
      <g transform="translate(200, 100) rotate(${angle})">
        <text class="watermark" x="0" y="0">${text}</text>
      </g>
    </svg>
  `

  // Convert to data URL
  const encoded = encodeURIComponent(svg.trim())
  return `data:image/svg+xml,${encoded}`
}

/**
 * Create watermark canvas element
 * Useful for client-side watermarking
 */
export function createWatermarkCanvas(
  options: WatermarkOptions,
  width: number = 800,
  height: number = 600
): HTMLCanvasElement {
  const canvas = document.createElement('canvas')
  canvas.width = width
  canvas.height = height

  const ctx = canvas.getContext('2d')
  if (!ctx) return canvas

  // Set up canvas
  ctx.clearRect(0, 0, width, height)
  ctx.globalAlpha = options.opacity || 0.15
  ctx.fillStyle = options.color || '#CCCCCC'
  ctx.font = `bold ${options.fontSize || 72}px ${options.fontFamily || 'Arial'}`
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'

  // Rotate and draw watermark
  ctx.save()
  ctx.translate(width / 2, height / 2)
  ctx.rotate(((options.angle || -45) * Math.PI) / 180)
  ctx.fillText(options.text, 0, 0)
  ctx.restore()

  return canvas
}

/**
 * Get watermark configuration for react-pdf
 * Returns style object for use in PDF components
 */
export function getWatermarkStyle(
  status: string
): Record<string, any> | null {
  const watermarkOptions = getWatermarkByStatus(status)
  if (!watermarkOptions) return null

  return {
    position: 'absolute',
    width: '100%',
    height: '100%',
    opacity: watermarkOptions.opacity || 0.15,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontSize: watermarkOptions.fontSize || 72,
    fontWeight: 'bold',
    color: watermarkOptions.color || '#CCCCCC',
    transform: `rotate(${watermarkOptions.angle || -45}deg)`,
    pointerEvents: 'none',
    zIndex: 0,
  }
}

/**
 * List of available watermark statuses
 */
export const AVAILABLE_WATERMARKS = [
  'DRAFT',
  'SUBMITTED',
  'IN_REVIEW',
  'APPROVED',
  'PAID',
  'REJECTED',
] as const

/**
 * Check if a status has a watermark
 */
export function hasWatermark(status: string): boolean {
  return AVAILABLE_WATERMARKS.includes(status as any)
}

/**
 * Get color for watermark based on status
 */
export function getWatermarkColor(status: string): string {
  const options = getWatermarkByStatus(status)
  return options?.color || '#CCCCCC'
}

/**
 * Get watermark text for status
 */
export function getWatermarkText(status: string): string {
  const options = getWatermarkByStatus(status)
  return options?.text || status
}
