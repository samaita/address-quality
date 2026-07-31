import type { AddressRequest, AddressResponse, ApiError } from "@/types/api"
import { MOCK_ADDRESS_RESPONSE } from "@/data/mock"

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? ""
const API_KEY = import.meta.env.VITE_API_KEY ?? ""
const REQUEST_TIMEOUT_MS = 15_000

export class RequestError extends Error implements ApiError {
  kind: ApiError["kind"]
  status?: number

  constructor(kind: ApiError["kind"], message: string, status?: number) {
    super(message)
    this.name = "RequestError"
    this.kind = kind
    this.status = status
  }
}

function classifyStatus(status: number): ApiError["kind"] {
  if (status === 429) return "rate_limited"
  if (status === 401 || status === 403) return "unauthorized"
  if (status >= 500) return "server"
  return "client"
}

async function parseErrorMessage(response: Response): Promise<string> {
  try {
    const body = await response.json()
    if (body?.error && typeof body.error === "string") return body.error
  } catch {
    // ignore non-JSON error bodies
  }
  return `HTTP ${response.status}`
}

export async function validateAddress(
  request: AddressRequest,
): Promise<AddressResponse> {
  if (API_BASE_URL) {
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS)

    try {
      const headers: Record<string, string> = {
        "Content-Type": "application/json",
      }
      if (API_KEY) {
        headers["X-API-Key"] = API_KEY
      }

      let res: Response
      try {
        res = await fetch(`${API_BASE_URL}/v1/validate`, {
          method: "POST",
          headers,
          body: JSON.stringify(request),
          signal: controller.signal,
        })
      } catch (err) {
        if (err instanceof DOMException && err.name === "AbortError") {
          throw new RequestError("timeout", "Request timed out")
        }
        throw new RequestError("network", "Network request failed")
      }

      if (!res.ok) {
        const message = await parseErrorMessage(res)
        throw new RequestError(classifyStatus(res.status), message, res.status)
      }

      return (await res.json()) as AddressResponse
    } finally {
      clearTimeout(timeoutId)
    }
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
