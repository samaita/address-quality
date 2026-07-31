import { Progress } from "@cloudflare/kumo/primitives/progress"
import type { ConfidenceTier } from "@/lib/confidence"

type ConfidenceBarProps = {
  value: number
  tone: ConfidenceTier
  label?: string
  className?: string
}

const toneBar: Record<ConfidenceTier, string> = {
  green: "bg-emerald-500",
  yellow: "bg-amber-500",
  red: "bg-red-500",
}

export default function ConfidenceBar({
  value,
  tone,
  label,
  className = "",
}: ConfidenceBarProps) {
  const pct = Math.min(100, Math.max(0, Math.round(value * 100)))

  return (
    <Progress.Root
      value={pct}
      max={100}
      aria-label={label}
      className={`h-2 w-full overflow-hidden rounded-full bg-surface-100 ${className}`}
    >
      <Progress.Track className="h-full w-full">
        <Progress.Indicator
          className={`h-full rounded-full ${toneBar[tone]}`}
          style={{ width: `${pct}%` }}
        />
      </Progress.Track>
    </Progress.Root>
  )
}
