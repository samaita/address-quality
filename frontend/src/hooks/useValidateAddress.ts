import { useState, useCallback } from "react"
import { validateAddress as callApi, RequestError } from "@/services/api"
import type { AddressResponse, ApiError } from "@/types/api"

type ErrorState = {
  kind: ApiError["kind"]
  message: string
  status?: number
}

type State =
  | { status: "idle" }
  | { status: "loading" }
  | { status: "success"; data: AddressResponse }
  | { status: "error"; error: ErrorState }

export default function useValidateAddress() {
  const [state, setState] = useState<State>({ status: "idle" })

  const validate = useCallback(async (address: string) => {
    setState({ status: "loading" })
    try {
      const data = await callApi({ address })
      setState({ status: "success", data })
    } catch (err) {
      const message =
        err instanceof Error ? err.message : "An unexpected error occurred"
      if (err instanceof RequestError) {
        setState({ status: "error", error: { kind: err.kind, message, status: err.status } })
      } else {
        setState({ status: "error", error: { kind: "unknown", message } })
      }
    }
  }, [])

  const clear = useCallback(() => {
    setState({ status: "idle" })
  }, [])

  return { state, validate, clear } as const
}
