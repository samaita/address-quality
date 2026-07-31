import Card from "@/components/common/Card"
import { ArrowPathIcon } from "@/components/icons"
import { Skeleton, SkeletonLines } from "@/components/kumo-ui"

export default function LoadingState() {
  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2 text-sm font-medium text-surface-700">
        <ArrowPathIcon className="h-4 w-4 animate-spin text-accent-500" />
        Validating address...
      </div>

      <Card>
        <div className="flex items-center justify-between">
          <Skeleton className="h-10 w-28" />
          <Skeleton className="h-6 w-20" />
        </div>
        <Skeleton className="mt-4 h-2 w-full" />
        <div className="mt-6 space-y-3">
          <SkeletonLines count={5} />
        </div>
      </Card>

      <Card>
        <Skeleton className="h-4 w-40" />
        <div className="mt-4 space-y-2">
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="flex gap-4">
              <Skeleton className="h-4 w-24" />
              <Skeleton className="h-4 w-32" />
              <Skeleton className="h-4 w-16" />
              <Skeleton className="h-4 w-20" />
            </div>
          ))}
        </div>
      </Card>

      <Card>
        <Skeleton className="h-4 w-36" />
        <SkeletonLines className="mt-4" count={4} />
      </Card>
    </div>
  )
}
