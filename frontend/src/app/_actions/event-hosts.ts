'use server'

import { APIResponse, EventHost } from '@/types'

const NOT_IMPLEMENTED = 'Event hosts feature is not yet implemented.'

export async function fetchEventHosts(): Promise<APIResponse<EventHost[] | null>> {
  return {
    success: false,
    message: NOT_IMPLEMENTED,
    data: null,
    status: 501,
  }
}

export async function createEventHost(
  _data: EventHost
): Promise<APIResponse<EventHost | null>> {
  return {
    success: false,
    message: NOT_IMPLEMENTED,
    data: null,
    status: 501,
  }
}

export async function updateEventHost(
  _hostId: string,
  _data: Partial<EventHost>
): Promise<APIResponse<EventHost | null>> {
  return {
    success: false,
    message: NOT_IMPLEMENTED,
    data: null,
    status: 501,
  }
}

export async function deleteEventHost(_hostId: string): Promise<APIResponse<null>> {
  return {
    success: false,
    message: NOT_IMPLEMENTED,
    data: null,
    status: 501,
  }
}
