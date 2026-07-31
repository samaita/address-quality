import { useEffect, useState } from "react"
import { Bars3Icon, XMarkIcon } from "@/components/icons"

const sections = [
  { id: "introduction", label: "Introduction" },
  { id: "how-it-works", label: "How It Works" },
  { id: "quickstart", label: "Quick Start" },
  { id: "authentication", label: "Authentication" },
  { id: "validate", label: "POST /v1/validate" },
  { id: "request-schema", label: "Request Fields" },
  { id: "successful-response", label: "Successful Response" },
  { id: "response-schema", label: "Understanding the Response" },
  { id: "errors", label: "Error Responses" },
  { id: "rate-limits", label: "Rate Limits" },
  { id: "data-source", label: "Data Source" },
  { id: "faq", label: "FAQ" },
]

export default function Sidebar() {
  const [activeId, setActiveId] = useState("introduction")
  const [mobileOpen, setMobileOpen] = useState(false)

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting) {
            setActiveId(entry.target.id)
          }
        }
      },
      { rootMargin: "-80px 0px -60% 0px" },
    )

    for (const { id } of sections) {
      const el = document.getElementById(id)
      if (el) observer.observe(el)
    }

    return () => observer.disconnect()
  }, [])

  return (
    <>
      <button
        type="button"
        onClick={() => setMobileOpen(!mobileOpen)}
        className="sticky top-16 z-40 flex w-full items-center justify-between border-b border-surface-200 bg-white px-4 py-3 text-sm font-medium text-surface-700 lg:hidden"
      >
        On this page
        {mobileOpen ? (
          <XMarkIcon className="h-4 w-4" />
        ) : (
          <Bars3Icon className="h-4 w-4" />
        )}
      </button>

      {mobileOpen && (
        <div className="border-b border-surface-200 bg-white px-4 pb-4 lg:hidden">
          {sections.map(({ id, label }) => (
            <a
              key={id}
              href={`#${id}`}
              onClick={() => {
                setMobileOpen(false)
                setActiveId(id)
              }}
              className={`block rounded-lg px-3 py-2 text-sm transition-colors ${
                activeId === id
                  ? "bg-accent-50 font-medium text-accent-700"
                  : "text-surface-600 hover:text-surface-900"
              }`}
            >
              {label}
            </a>
          ))}
        </div>
      )}

      <nav className="hidden w-56 flex-shrink-0 lg:block">
        <div className="sticky top-24">
          <p className="mb-3 text-xs font-medium uppercase tracking-wider text-surface-400">
            On this page
          </p>
          <ul className="space-y-0.5">
            {sections.map(({ id, label }) => (
              <li key={id}>
                <a
                  href={`#${id}`}
                  className={`block rounded-lg px-3 py-2 text-sm transition-colors ${
                    activeId === id
                      ? "bg-accent-50 font-medium text-accent-700"
                      : "text-surface-500 hover:text-surface-900"
                  }`}
                >
                  {label}
                </a>
              </li>
            ))}
          </ul>
        </div>
      </nav>
    </>
  )
}
