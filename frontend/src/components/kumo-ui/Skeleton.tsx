type SkeletonProps = {
  className?: string
}

export default function Skeleton({ className = "" }: SkeletonProps) {
  return (
    <div
      aria-hidden="true"
      className={`animate-pulse rounded-md bg-surface-100 ${className}`}
    />
  )
}

type SkeletonLinesProps = {
  count?: number
  className?: string
}

export function SkeletonLines({ count = 3, className = "" }: SkeletonLinesProps) {
  return (
    <div className={`space-y-3 ${className}`}>
      {Array.from({ length: count }).map((_, i) => (
        <Skeleton
          key={i}
          className={`h-4 ${i % 2 === 0 ? "w-full" : "w-4/5"}`}
        />
      ))}
    </div>
  )
}
