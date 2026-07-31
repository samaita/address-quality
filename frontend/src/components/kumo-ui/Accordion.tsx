import { useState, type ReactNode } from "react"
import { ChevronDownIcon } from "@/components/icons"

export type AccordionItem = {
  id: string
  trigger: ReactNode
  actions?: ReactNode
  children: ReactNode
  defaultOpen?: boolean
}

type AccordionProps = {
  items: AccordionItem[]
  className?: string
}

export default function Accordion({ items, className = "" }: AccordionProps) {
  const initialOpen = items.find((item) => item.defaultOpen)?.id
  const [openId, setOpenId] = useState<string | null>(initialOpen ?? null)

  return (
    <div className={`divide-y divide-surface-200 rounded-xl border border-surface-200 ${className}`}>
      {items.map((item) => {
        const isOpen = openId === item.id

        return (
          <div key={item.id}>
            <div className="flex items-center gap-2 pr-2">
              <button
                type="button"
                aria-expanded={isOpen}
                aria-controls={`${item.id}-panel`}
                onClick={() => setOpenId(isOpen ? null : item.id)}
                className="flex flex-1 items-center justify-between gap-4 px-4 py-3 text-left text-sm font-medium text-surface-900 transition-colors hover:bg-surface-50 focus:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-accent-500"
              >
                <span>{item.trigger}</span>
                <ChevronDownIcon
                  className={`h-4 w-4 flex-shrink-0 text-surface-400 transition-transform duration-200 ${
                    isOpen ? "rotate-180" : ""
                  }`}
                />
              </button>
              {item.actions}
            </div>
            {isOpen && (
              <div
                id={`${item.id}-panel`}
                className="border-t border-surface-200 px-4 py-4"
              >
                {item.children}
              </div>
            )}
          </div>
        )
      })}
    </div>
  )
}
