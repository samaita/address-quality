import useCopyToClipboard from "@/hooks/useCopyToClipboard"
import CopyButton from "@/components/common/CopyButton"

type JsonViewerProps = {
  data: unknown
  title?: string
}

export default function JsonViewer({ data, title = "Raw JSON" }: JsonViewerProps) {
  const { copied, copy } = useCopyToClipboard()
  const code = JSON.stringify(data, null, 2)

  return (
    <div className="group relative overflow-hidden rounded-xl border border-surface-200 bg-surface-50">
      <div className="flex items-center justify-between border-b border-surface-200 px-4 py-2">
        <span className="text-xs font-medium text-surface-500">{title}</span>
        <CopyButton
          copied={copied}
          onCopy={() => copy(code)}
          className="text-surface-400 hover:text-surface-700"
        />
      </div>
      <pre className="overflow-x-auto p-4 text-xs leading-relaxed text-surface-700">
        <code>{code}</code>
      </pre>
    </div>
  )
}
