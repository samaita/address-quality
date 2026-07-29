import { useEffect, useRef, useState, useCallback } from "react"
import Prism from "prismjs"
import "prismjs/components/prism-json"
import "prismjs/themes/prism-tomorrow.css"
import { ClipboardDocumentIcon, CheckIcon } from "@/components/icons"

type CodeBlockProps = {
  code: string
  language?: string
  title?: string
}

export default function CodeBlock({ code, language = "json", title }: CodeBlockProps) {
  const ref = useRef<HTMLElement>(null)
  const [copied, setCopied] = useState(false)

  useEffect(() => {
    if (ref.current) {
      Prism.highlightElement(ref.current)
    }
  }, [code, language])

  const handleCopy = useCallback(async () => {
    await navigator.clipboard.writeText(code)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }, [code])

  return (
    <div className="group relative overflow-hidden rounded-xl border border-surface-700 bg-[#2d2d2d]">
      {title && (
        <div className="flex items-center justify-between border-b border-surface-700 px-4 py-2">
          <span className="text-xs font-medium text-surface-400">{title}</span>
        </div>
      )}
      <div className="relative">
        <button
          type="button"
          onClick={handleCopy}
          className="absolute right-3 top-3 z-10 rounded-md p-1.5 text-surface-500 opacity-0 transition-opacity hover:text-white group-hover:opacity-100"
          aria-label="Copy code"
        >
          {copied ? (
            <CheckIcon className="h-4 w-4 text-emerald-400" />
          ) : (
            <ClipboardDocumentIcon className="h-4 w-4" />
          )}
        </button>
        <pre className="overflow-x-auto p-4 text-sm leading-relaxed">
          <code ref={ref} className={`language-${language}`}>
            {code}
          </code>
        </pre>
      </div>
    </div>
  )
}
