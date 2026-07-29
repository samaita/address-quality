import { useState, useCallback } from "react"
import { validateAddress as callApi } from "@/services/api"
import type { AddressResponse } from "@/types/api"

type State =
  | { status: "idle" }
  | { status: "loading" }
  | { status: "success"; data: AddressResponse }
  | { status: "error"; error: string }

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
      setState({ status: "error", error: message })
    }
  }, [])

  const reset = useCallback(() => {
    setState({ status: "idle" })
  }, [])

  return { state, validate, reset } as const
}
