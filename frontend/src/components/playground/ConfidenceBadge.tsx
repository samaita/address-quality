import type { QualityStatus } from "@/types/api"

type ConfidenceBadgeProps = {
  confidence: number
  status: QualityStatus
}

const statusConfig: Record<
  QualityStatus,
  { label: string; ring: string; text: string; bg: string }
> = {
  VALID: {
    label: "Valid",
    ring: "border-emerald-500",
    text: "text-emerald-600",
    bg: "bg-emerald-50",
  },
  INCOMPLETE: {
    label: "Incomplete",
    ring: "border-amber-500",
    text: "text-amber-600",
    bg: "bg-amber-50",
  },
  AMBIGUOUS: {
    label: "Ambiguous",
    ring: "border-orange-500",
    text: "text-orange-600",
    bg: "bg-orange-50",
  },
  CONFLICT: {
    label: "Conflict",
    ring: "border-red-500",
    text: "text-red-600",
    bg: "bg-red-50",
  },
  UNKNOWN: {
    label: "Unknown",
    ring: "border-surface-300",
    text: "text-surface-600",
    bg: "bg-surface-100",
  },
}

export default function ConfidenceBadge({ confidence, status }: ConfidenceBadgeProps) {
  const config = statusConfig[status]
  const percentage = Math.round(confidence * 100)

  return (
    <div className="flex items-center gap-4">
      <div
        className={`flex h-20 w-20 flex-col items-center justify-center rounded-full border-2 ${config.ring} ${config.bg}`}
      >
        <span className={`text-xl font-bold ${config.text}`}>{percentage}%</span>
      </div>
      <div>
        <p className={`text-lg font-semibold ${config.text}`}>{config.label}</p>
        <p className="text-sm text-surface-500">
          Confidence score based on matched evidence
        </p>
      </div>
    </div>
  )
}
