import { useState, useCallback } from "react"
import Button from "@/components/common/Button"
import { ArrowPathIcon } from "@/components/icons"
import { EXAMPLE_ADDRESS } from "@/data/mock"

type AddressInputProps = {
  onValidate: (address: string) => void
  loading: boolean
}

export default function AddressInput({ onValidate, loading }: AddressInputProps) {
  const [value, setValue] = useState("")

  const handleSubmit = useCallback(() => {
    const trimmed = value.trim()
    if (trimmed && !loading) {
      onValidate(trimmed)
    }
  }, [value, loading, onValidate])

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) {
        handleSubmit()
      }
    },
    [handleSubmit],
  )

  const handleExample = useCallback(() => {
    setValue(EXAMPLE_ADDRESS)
  }, [])

  return (
    <div className="flex h-full flex-col gap-4">
      <div className="relative flex-1">
        <textarea
          value={value}
          onChange={(e) => setValue(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder={EXAMPLE_ADDRESS}
          className="h-full w-full resize-none rounded-xl border border-surface-200 bg-white p-4 text-sm leading-relaxed text-surface-900 placeholder:text-surface-400 focus:border-accent-500 focus:outline-none focus:ring-2 focus:ring-accent-500/20 disabled:opacity-50"
          disabled={loading}
          rows={6}
        />
      </div>
      <div className="flex gap-3">
        <Button
          onClick={handleSubmit}
          disabled={!value.trim() || loading}
          className="flex-1"
        >
          {loading ? (
            <>
              <ArrowPathIcon className="h-4 w-4 animate-spin" />
              Validating...
            </>
          ) : (
            "Validate"
          )}
        </Button>
        <Button variant="secondary" onClick={handleExample} disabled={loading}>
          Try Example
        </Button>
      </div>
      <p className="text-xs text-surface-400">
        Press Ctrl+Enter or Cmd+Enter to submit
      </p>
    </div>
  )
}
