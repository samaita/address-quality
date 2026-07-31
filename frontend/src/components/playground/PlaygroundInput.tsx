import { useCallback, useMemo, useState } from "react"
import { Button } from "@cloudflare/kumo/components/button"
import { Banner } from "@cloudflare/kumo/components/banner"
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
          type="button"
          variant="primary"
          size="lg"
          loading={loading}
          disabled={!trimmed}
          onClick={handleSubmit}
          className="flex-1"
        >
          Validate Address
        </Button>
        <Button
          type="button"
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

      <div className="pt-2">
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

      {error && (
        <Banner
          variant="error"
          title="Validation failed"
          description={error.message}
          action={
            <Banner.Action variant="secondary" onClick={handleSubmit}>
              Try again
            </Banner.Action>
          }
        />
      )}
    </div>
  )
}
