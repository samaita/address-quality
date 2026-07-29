import type { ButtonHTMLAttributes } from "react"
import { ClipboardDocumentIcon, CheckIcon } from "@/components/icons"

type CopyButtonProps = {
  copied: boolean
  onCopy: () => void
  label?: string
} & Omit<ButtonHTMLAttributes<HTMLButtonElement>, "onClick" | "children">

export default function CopyButton({
  copied,
  onCopy,
  label,
  className = "",
  ...props
}: CopyButtonProps) {
  return (
    <button
      type="button"
      onClick={onCopy}
      className={`inline-flex items-center gap-1.5 rounded-md text-xs transition-colors ${className}`}
      aria-label={copied ? "Copied" : "Copy"}
      {...props}
    >
      {copied ? (
        <>
          <CheckIcon className="h-3.5 w-3.5 text-emerald-500" />
          {label ?? "Copied"}
        </>
      ) : (
        <>
          <ClipboardDocumentIcon className="h-3.5 w-3.5" />
          {label ?? "Copy"}
        </>
      )}
    </button>
  )
}
