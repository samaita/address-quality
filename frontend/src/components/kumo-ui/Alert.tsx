import type { ReactNode } from "react"
import { CheckIcon, ExclamationTriangleIcon } from "@/components/icons"

type AlertVariant = "info" | "success" | "warning" | "danger"

type AlertProps = {
  variant?: AlertVariant
  title?: ReactNode
  icon?: ReactNode
  action?: ReactNode
  children?: ReactNode
  className?: string
}

const variantStyles: Record<AlertVariant, string> = {
  info: "border-accent-200 bg-accent-50 text-accent-800",
  success: "border-emerald-200 bg-emerald-50 text-emerald-800",
  warning: "border-amber-200 bg-amber-50 text-amber-800",
  danger: "border-red-200 bg-red-50 text-red-800",
}

const iconStyles: Record<AlertVariant, string> = {
  info: "text-accent-500",
  success: "text-emerald-500",
  warning: "text-amber-500",
  danger: "text-red-500",
}

export default function Alert({
  variant = "info",
  title,
  icon,
  action,
  children,
  className = "",
}: AlertProps) {
  const defaultIcon =
    variant === "success" ? (
      <CheckIcon className="h-5 w-5" />
    ) : (
      <ExclamationTriangleIcon className="h-5 w-5" />
    )

  return (
    <div
      role="alert"
      className={`rounded-xl border px-4 py-3 ${variantStyles[variant]} ${className}`}
    >
      <div className="flex items-start gap-3">
        <span className={`mt-0.5 flex-shrink-0 ${iconStyles[variant]}`}>
          {icon ?? defaultIcon}
        </span>
        <div className="min-w-0 flex-1">
          {title && (
            <p className="text-sm font-semibold text-current">{title}</p>
          )}
          {children && (
            <div className={`text-sm ${title ? "mt-1" : ""} text-current opacity-90`}>
              {children}
            </div>
          )}
        </div>
        {action && <div className="flex-shrink-0">{action}</div>}
      </div>
    </div>
  )
}
