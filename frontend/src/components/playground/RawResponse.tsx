import { useEffect, useRef, useState } from "react"
import Prism from "prismjs"
import "prismjs/components/prism-json"
import "prismjs/themes/prism-tomorrow.css"
import { Accordion } from "@cloudflare/kumo/primitives/accordion"
import { ChevronDownIcon } from "@/components/icons"
import CopyButton from "@/components/common/CopyButton"
import useCopyToClipboard from "@/hooks/useCopyToClipboard"

type RawResponseProps = {
  data: unknown
}

export default function RawResponse({ data }: RawResponseProps) {
  const code = JSON.stringify(data, null, 2)
  const ref = useRef<HTMLElement>(null)
  const { copied, copy } = useCopyToClipboard()
  const [open, setOpen] = useState(false)

  useEffect(() => {
    if (ref.current) {
      Prism.highlightElement(ref.current)
    }
  }, [code])

  return (
    <Accordion.Root
      value={open ? ["raw-response"] : []}
      onValueChange={(v) => setOpen(v.length > 0)}
      multiple={false}
      className="divide-y divide-surface-200 rounded-xl border border-surface-200"
    >
      <Accordion.Item value="raw-response">
        <Accordion.Header className="flex items-center gap-2 pr-2">
          <Accordion.Trigger className="flex flex-1 items-center justify-between gap-4 px-4 py-3 text-left text-sm font-medium text-surface-900 transition-colors hover:bg-surface-50 focus:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-accent-500">
            <span>Raw API Response</span>
            <ChevronDownIcon
              className={`h-4 w-4 flex-shrink-0 text-surface-400 transition-transform duration-200 ${
                open ? "rotate-180" : ""
              }`}
            />
          </Accordion.Trigger>
          <CopyButton
            copied={copied}
            onCopy={() => copy(code)}
            className="flex-shrink-0 text-xs text-surface-400 hover:text-surface-700"
          />
        </Accordion.Header>
        <Accordion.Panel className="border-t border-surface-200 px-4 py-4">
          <div className="overflow-hidden rounded-xl border border-surface-700 bg-[#2d2d2d]">
            <pre className="max-h-[300px] overflow-y-auto overflow-x-auto p-4 text-sm leading-relaxed">
              <code ref={ref} className="language-json">
                {code}
              </code>
            </pre>
          </div>
        </Accordion.Panel>
      </Accordion.Item>
    </Accordion.Root>
  )
}
