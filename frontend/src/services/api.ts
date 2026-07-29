import type { AddressRequest, AddressResponse } from "@/types/api"
import { MOCK_ADDRESS_RESPONSE } from "@/data/mock"

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? ""

export async function validateAddress(request: AddressRequest): Promise<AddressResponse> {
  if (API_BASE_URL) {
    const res = await fetch(`${API_BASE_URL}/v1/validate`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(request),
    })
    if (!res.ok) {
      const error = await res.json().catch(() => ({ error: "Request failed" }))
      throw new Error(error.error ?? `HTTP ${res.status}`)
    }
    return res.json()
  }

  await new Promise((r) => setTimeout(r, 800))

  return {
    ...MOCK_ADDRESS_RESPONSE,
    data: {
      ...MOCK_ADDRESS_RESPONSE.data,
      raw_input: request.address,
    },
  }
}
