import { useEffect, useId, useRef, useState } from "react"

type MermaidProps = {
  chart: string
  className?: string
}

export default function Mermaid({ chart, className = "" }: MermaidProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const id = useId().replace(/[:]/g, "")
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const container = containerRef.current
    if (!container) return
    let cancelled = false

    import("mermaid")
      .then(({ default: mermaid }) => {
        if (cancelled) return
        mermaid.initialize({ startOnLoad: false, theme: "base", fontFamily: "inherit" })
        return mermaid.render(id, chart).then(({ svg }) => {
          if (cancelled || !container) return
          container.innerHTML = svg
        })
      })
      .catch((err: unknown) => {
        if (cancelled) return
        setError(err instanceof Error ? err.message : String(err))
      })

    return () => {
      cancelled = true
      container.innerHTML = ""
    }
  }, [id, chart])

  return (
    <div
      className={`overflow-x-auto rounded-xl border border-surface-200 bg-white p-4 ${className}`}
    >
      {error ? (
        <pre className="text-xs text-red-600">{error}</pre>
      ) : (
        <div ref={containerRef} />
      )}
    </div>
  )
}
