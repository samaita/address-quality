import { useCallback, useMemo, useState } from "react"
import Button from "@/components/common/Button"
import { Alert } from "@/components/kumo-ui"
import { ArrowPathIcon } from "@/components/icons"
import { EXAMPLE_ADDRESSES } from "@/data/mock"
import type { ApiError } from "@/types/api"

type PlaygroundInputProps = {
  onValidate: (address: string) => void
  onClear: () => void
  loading: boolean
  error?: Pick<ApiError, "kind" | "message"> | null
}

export default function PlaygroundInput({
  onValidate,
  onClear,
  loading,
  error,
}: PlaygroundInputProps) {
  const [value, setValue] = useState("")
  const trimmed = useMemo(() => value.trim(), [value])

  const handleSubmit = useCallback(() => {
    if (trimmed && !loading) {
      onValidate(trimmed)
    }
  }, [trimmed, loading, onValidate])

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) {
        handleSubmit()
      }
    },
    [handleSubmit],
  )

  return (
    <div className="flex h-full flex-col gap-5">
      <div className="flex flex-col gap-3">
        <label htmlFor="address-input" className="text-sm font-semibold text-surface-900">
          Address
        </label>
        <textarea
          id="address-input"
          value={value}
          onChange={(e) => setValue(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Paste an Indonesian address, e.g. Jl. Asia Afrika No.56, Braga, Bandung 40111"
          className="h-40 w-full resize-y rounded-xl border border-surface-200 bg-white p-4 text-sm leading-relaxed text-surface-900 placeholder:text-surface-400 focus:border-accent-500 focus:outline-none focus:ring-2 focus:ring-accent-500/20 disabled:opacity-50"
          disabled={loading}
        />
        <p className="text-xs text-surface-400">
          Press Ctrl+Enter or Cmd+Enter to validate
        </p>
      </div>

      <div className="flex gap-3">
        <Button
          onClick={handleSubmit}
          disabled={!trimmed || loading}
          className="flex-1"
        >
          {loading ? (
            <>
              <ArrowPathIcon className="h-4 w-4 animate-spin" />
              Validating...
            </>
          ) : (
            "Validate Address"
          )}
        </Button>
        <Button
          variant="secondary"
          onClick={() => {
            setValue("")
            onClear()
          }}
          disabled={loading}
        >
          Clear
        </Button>
      </div>

      {error && (
        <Alert
          variant="danger"
          title="Validation failed"
          action={
            <Button
              size="sm"
              variant="ghost"
              onClick={handleSubmit}
              className="text-red-600 hover:bg-red-50 hover:text-red-700"
            >
              Try again
            </Button>
          }
        >
          {error.message}
        </Alert>
      )}

      <div className="mt-auto pt-2">
        <p className="mb-2 text-sm font-semibold text-surface-900">
          Example addresses
        </p>
        <ul className="space-y-1">
          {EXAMPLE_ADDRESSES.map((addr) => (
            <li key={addr}>
              <button
                type="button"
                onClick={() => setValue(addr)}
                disabled={loading}
                className="w-full rounded-lg px-3 py-1.5 text-left text-xs text-surface-500 transition-colors hover:bg-surface-100 hover:text-surface-900 disabled:opacity-50"
              >
                {addr}
              </button>
            </li>
          ))}
        </ul>
      </div>
    </div>
  )
}
