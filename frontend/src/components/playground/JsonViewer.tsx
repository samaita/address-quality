import { useState, useCallback } from "react"
import { ClipboardDocumentIcon, CheckIcon } from "@/components/icons"

type JsonViewerProps = {
  data: unknown
  title?: string
}

export default function JsonViewer({ data, title = "Raw JSON" }: JsonViewerProps) {
  const [copied, setCopied] = useState(false)

  const code = JSON.stringify(data, null, 2)

  const handleCopy = useCallback(async () => {
    await navigator.clipboard.writeText(code)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }, [code])

  return (
    <div className="group relative overflow-hidden rounded-xl border border-surface-200 bg-surface-50">
      <div className="flex items-center justify-between border-b border-surface-200 px-4 py-2">
        <span className="text-xs font-medium text-surface-500">{title}</span>
        <button
          type="button"
          onClick={handleCopy}
          className="flex items-center gap-1.5 rounded-md px-2 py-1 text-xs text-surface-400 transition-colors hover:text-surface-700"
          aria-label="Copy JSON"
        >
          {copied ? (
            <>
              <CheckIcon className="h-3.5 w-3.5 text-emerald-500" />
              Copied
            </>
          ) : (
            <>
              <ClipboardDocumentIcon className="h-3.5 w-3.5" />
              Copy
            </>
          )}
        </button>
      </div>
      <pre className="overflow-x-auto p-4 text-xs leading-relaxed text-surface-700">
        <code>{code}</code>
      </pre>
    </div>
  )
}
