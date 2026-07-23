'use server'

import { APIResponse, WhatsAppConfig } from '@/types'

const NOT_IMPLEMENTED = 'WhatsApp integration is not yet implemented. Contact your administrator.'

export async function fetchWhatsAppConfig(): Promise<APIResponse<WhatsAppConfig | null>> {
  return {
    success: false,
    message: NOT_IMPLEMENTED,
    data: null,
    status: 501,
  }
}

export async function updateWhatsAppConfig(
  _configId: string,
  _data: Partial<WhatsAppConfig>
): Promise<APIResponse<WhatsAppConfig | null>> {
  return {
    success: false,
    message: NOT_IMPLEMENTED,
    data: null,
    status: 501,
  }
}

export async function testWhatsAppConnection(
  _configId: string
): Promise<APIResponse<{ connected: boolean } | null>> {
  return {
    success: false,
    message: NOT_IMPLEMENTED,
    data: null,
    status: 501,
  }
}
