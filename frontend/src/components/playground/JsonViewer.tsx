import { useEffect, useRef } from "react"
import Prism from "prismjs"
import "prismjs/components/prism-json"
import "prismjs/themes/prism-tomorrow.css"
import useCopyToClipboard from "@/hooks/useCopyToClipboard"

type JsonViewerProps = {
  data: unknown
  title?: string
}

export default function JsonViewer({ data, title = "Raw JSON" }: JsonViewerProps) {
  const code = JSON.stringify(data, null, 2)
  const ref = useRef<HTMLElement>(null)
  const { copied, copy } = useCopyToClipboard()

  useEffect(() => {
    if (ref.current) {
      Prism.highlightElement(ref.current)
    }
  }, [code])

  return (
    <div className="group relative overflow-hidden rounded-xl border border-surface-700 bg-[#2d2d2d]">
      <div className="flex items-center justify-between border-b border-surface-700 px-4 py-2">
        <span className="text-xs font-medium text-surface-400">{title}</span>
        <button
          type="button"
          onClick={() => copy(code)}
          className="rounded-md p-1 text-surface-500 opacity-0 transition-opacity hover:text-white group-hover:opacity-100 focus:outline-none focus-visible:opacity-100 focus-visible:ring-2 focus-visible:ring-white/30"
          aria-label="Copy code"
        >
          {copied ? (
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="h-4 w-4 text-emerald-400">
              <path strokeLinecap="round" strokeLinejoin="round" d="M4.5 12.75l6 6 9-13.5" />
            </svg>
          ) : (
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="h-4 w-4">
              <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 17.25v3.375c0 .621-.504 1.125-1.125 1.125h-9.75a1.125 1.125 0 01-1.125-1.125V7.875c0-.621.504-1.125 1.125-1.125H6.75a9.06 9.06 0 011.5.124m7.5 10.376h3.375c.621 0 1.125-.504 1.125-1.125V11.25c0-4.46-3.243-8.161-7.5-8.876a9.06 9.06 0 00-1.5-.124H9.375c-.621 0-1.125.504-1.125 1.125v3.5m7.5 10.375H9.375a1.125 1.125 0 01-1.125-1.125v-9.25m12 6.625v-1.875a3.375 3.375 0 00-3.375-3.375h-1.5a1.125 1.125 0 01-1.125-1.125v-1.5a3.375 3.375 0 00-3.375-3.375H9.75" />
            </svg>
          )}
        </button>
      </div>
      <pre className="max-h-[300px] overflow-y-auto overflow-x-auto p-4 text-sm leading-relaxed">
        <code ref={ref} className="language-json">
          {code}
        </code>
      </pre>
    </div>
  )
}
