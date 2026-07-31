type ProgressTone = "green" | "yellow" | "red" | "accent"

type ProgressProps = {
  value: number
  tone?: ProgressTone
  className?: string
  label?: string
}

const toneStyles: Record<ProgressTone, string> = {
  green: "bg-emerald-500",
  yellow: "bg-amber-500",
  red: "bg-red-500",
  accent: "bg-accent-600",
}

export default function Progress({
  value,
  tone = "accent",
  className = "",
  label,
}: ProgressProps) {
  const clamped = Math.min(100, Math.max(0, Math.round(value * 100)))

  return (
    <div
      role="progressbar"
      aria-valuenow={clamped}
      aria-valuemin={0}
      aria-valuemax={100}
      aria-label={label}
      className={`h-2 w-full overflow-hidden rounded-full bg-surface-100 ${className}`}
    >
      <div
        className={`h-full rounded-full transition-all duration-500 ${toneStyles[tone]}`}
        style={{ width: `${clamped}%` }}
      />
    </div>
  )
}
