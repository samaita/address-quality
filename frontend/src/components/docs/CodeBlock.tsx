import { useEffect, useRef, useState } from "react"
import Prism from "prismjs"
import "prismjs/components/prism-json"
import "prismjs/components/prism-bash"
import "prismjs/components/prism-python"
import "prismjs/components/prism-javascript"
import "prismjs/themes/prism-tomorrow.css"
import useCopyToClipboard from "@/hooks/useCopyToClipboard"

type CodeTab = {
  label: string
  code: string
  language?: string
}

type CodeBlockProps = {
  code?: string
  language?: string
  title?: string
  tabs?: CodeTab[]
  collapsible?: boolean
}

export default function CodeBlock({
  code = "",
  language = "json",
  title,
  tabs,
  collapsible = false,
}: CodeBlockProps) {
  const ref = useRef<HTMLElement>(null)
  const { copied, copy } = useCopyToClipboard()
  const [activeTab, setActiveTab] = useState(0)
  const [expanded, setExpanded] = useState(false)

  const effectiveTabs = tabs ?? [{ label: title ?? "", code, language }]
  const active = effectiveTabs[activeTab]
  const activeCode = active.code
  const activeLanguage = active.language ?? language
  const hasMultipleTabs = effectiveTabs.length > 1

  useEffect(() => {
    if (ref.current) {
      Prism.highlightElement(ref.current)
    }
  }, [activeCode, activeLanguage])

  return (
    <div className="group relative overflow-hidden rounded-xl border border-surface-700 bg-[#2d2d2d]">
      {hasMultipleTabs && (
        <div className="flex border-b border-surface-700">
          {effectiveTabs.map((tab, i) => (
            <button
              key={tab.label}
              type="button"
              onClick={() => setActiveTab(i)}
              className={`px-4 py-2 text-xs font-medium transition-colors ${
                i === activeTab
                  ? "bg-surface-800 text-white"
                  : "text-surface-400 hover:text-surface-200"
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>
      )}
      {!hasMultipleTabs && title && (
        <div className="flex items-center justify-between border-b border-surface-700 px-4 py-2">
          <span className="text-xs font-medium text-surface-400">{title}</span>
        </div>
      )}
      <div
        className={`relative ${
          collapsible && !expanded ? "max-h-48 overflow-hidden" : ""
        }`}
      >
        <button
          type="button"
          onClick={() => copy(activeCode)}
          className="absolute right-3 top-3 z-10 rounded-md p-1.5 text-surface-500 transition-opacity hover:text-white max-sm:opacity-100 md:opacity-0 md:group-hover:opacity-100 focus-visible:opacity-100 focus-visible:ring-2 focus-visible:ring-white/30"
          aria-label="Copy code"
        >
          {copied ? (
            <CheckIcon className="h-4 w-4 text-emerald-400" />
          ) : (
            <ClipboardDocumentIcon className="h-4 w-4" />
          )}
        </button>
        <pre className="overflow-x-auto p-4 text-sm leading-relaxed">
          <code ref={ref} className={`language-${activeLanguage}`}>
            {activeCode}
          </code>
        </pre>
      </div>
      {collapsible && (
        <div
          className={`pointer-events-none absolute bottom-0 left-0 right-0 h-16 bg-gradient-to-t from-[#2d2d2d] to-transparent transition-opacity ${
            expanded ? "opacity-0" : "opacity-100"
          }`}
        />
      )}
      {collapsible && (
        <button
          type="button"
          onClick={() => setExpanded(!expanded)}
          className="block w-full border-t border-surface-700 py-2 text-center text-xs font-medium text-surface-400 transition-colors hover:text-surface-200"
        >
          {expanded ? "Show less" : "Show full response"}
        </button>
      )}
    </div>
  )
}

function ClipboardDocumentIcon({ className }: { className?: string }) {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className={className}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 7.5V6.108c0-1.135.845-2.098 1.976-2.192.373-.03.748-.057 1.123-.08M15.75 18H18a2.25 2.25 0 002.25-2.25V6.108c0-1.135-.845-2.098-1.976-2.192a48.424 48.424 0 00-1.123-.08M15.75 18.75v-1.875a3.375 3.375 0 00-3.375-3.375h-1.5a1.125 1.125 0 01-1.125-1.125v-1.5A3.375 3.375 0 006.375 7.5H5.25m11.9-3.664A2.251 2.251 0 0015 2.25h-1.5a2.251 2.251 0 00-2.15 1.586m5.8 0c.065.21.1.433.1.664v.75h-6V4.5c0-.231.035-.454.1-.664M6.75 7.5H4.875c-.621 0-1.125.504-1.125 1.125v12c0 .621.504 1.125 1.125 1.125h9.75c.621 0 1.125-.504 1.125-1.125V16.5a9 9 0 00-9-9z" />
    </svg>
  )
}

function CheckIcon({ className }: { className?: string }) {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className={className}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M4.5 12.75l6 6 9-13.5" />
    </svg>
  )
}
