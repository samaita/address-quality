import { Loader } from "@cloudflare/kumo/components/loader"
import Card from "@/components/common/Card"

export default function LoadingState() {
  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2 text-sm font-medium text-surface-700">
        <Loader size="sm" aria-label="Validating" />
        Validating address...
      </div>

      <Card>
        <div className="flex items-center justify-between">
          <SkeletonBlock className="h-10 w-28" />
          <SkeletonBlock className="h-6 w-20" />
        </div>
        <SkeletonBlock className="mt-4 h-2 w-full" />
        <div className="mt-6 space-y-3">
          <SkeletonBlock className="h-4 w-full" />
          <SkeletonBlock className="h-4 w-4/5" />
          <SkeletonBlock className="h-4 w-full" />
          <SkeletonBlock className="h-4 w-2/3" />
          <SkeletonBlock className="h-4 w-4/5" />
        </div>
      </Card>

      <Card>
        <SkeletonBlock className="h-4 w-40" />
        <div className="mt-4 space-y-2">
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="flex gap-4">
              <SkeletonBlock className="h-4 w-24" />
              <SkeletonBlock className="h-4 w-32" />
              <SkeletonBlock className="h-4 w-16" />
              <SkeletonBlock className="h-4 w-20" />
            </div>
          ))}
        </div>
      </Card>

      <Card>
        <SkeletonBlock className="h-4 w-36" />
        <div className="mt-4 space-y-3">
          <SkeletonBlock className="h-4 w-full" />
          <SkeletonBlock className="h-4 w-4/5" />
          <SkeletonBlock className="h-4 w-2/3" />
          <SkeletonBlock className="h-4 w-4/5" />
        </div>
      </Card>
    </div>
  )
}

function SkeletonBlock({ className = "" }: { className?: string }) {
  return (
    <div aria-hidden="true" className={`animate-pulse rounded-md bg-surface-100 ${className}`} />
  )
}
