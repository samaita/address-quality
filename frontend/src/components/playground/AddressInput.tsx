import { useState, useCallback } from "react"
import Button from "@/components/common/Button"
import { ArrowPathIcon } from "@/components/icons"
import { EXAMPLE_ADDRESS } from "@/data/mock"

const EXAMPLE_ADDRESSES = [
  EXAMPLE_ADDRESS,
  "KOMPLEK BUMI INDAH BLOK A5 NO 12, BEJI, DEPOK",
  "JL SUDIRMAN NO 88 RT 01 RW 03, PISANG BARU, JAKARTA PUSAT",
]

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

  return (
    <div className="flex h-full flex-col gap-4">
      <div className="relative min-h-[160px] flex-1">
        <textarea
          value={value}
          onChange={(e) => setValue(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder={EXAMPLE_ADDRESS}
          className="h-full min-h-[160px] max-h-[200px] w-full resize-none rounded-xl border border-surface-200 bg-white p-4 text-sm leading-relaxed text-surface-900 placeholder:text-surface-400 focus:border-accent-500 focus:outline-none focus:ring-2 focus:ring-accent-500/20 disabled:opacity-50"
          disabled={loading}
        />
      </div>
      <Button
        onClick={handleSubmit}
        disabled={!value.trim() || loading}
        className="w-full"
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
      <p className="text-xs text-surface-400">
        Press Ctrl+Enter or Cmd+Enter to submit
      </p>
    </div>
  )
}
