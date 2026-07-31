import { useEffect, useRef } from "react"
import Prism from "prismjs"
import "prismjs/components/prism-json"
import "prismjs/themes/prism-tomorrow.css"
import { Accordion } from "@/components/kumo-ui"
import CopyButton from "@/components/common/CopyButton"
import useCopyToClipboard from "@/hooks/useCopyToClipboard"

type RawResponseProps = {
  data: unknown
}

export default function RawResponse({ data }: RawResponseProps) {
  const code = JSON.stringify(data, null, 2)
  const ref = useRef<HTMLElement>(null)
  const { copied, copy } = useCopyToClipboard()

  useEffect(() => {
    if (ref.current) {
      Prism.highlightElement(ref.current)
    }
  }, [code])

  return (
    <Accordion
      items={[
        {
          id: "raw-response",
          trigger: <span className="flex items-center gap-2">Raw API Response</span>,
          actions: (
            <CopyButton
              copied={copied}
              onCopy={() => copy(code)}
              className="flex-shrink-0 text-surface-400 hover:text-surface-700"
            />
          ),
          children: (
            <div className="overflow-hidden rounded-xl border border-surface-700 bg-[#2d2d2d]">
              <pre className="max-h-[300px] overflow-y-auto overflow-x-auto p-4 text-sm leading-relaxed">
                <code ref={ref} className="language-json">
                  {code}
                </code>
              </pre>
            </div>
          ),
        },
      ]}
    />
  )
}
