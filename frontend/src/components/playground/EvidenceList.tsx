import Card from "@/components/common/Card"
import type { Assessment } from "@/types/api"
import { CheckIcon, ExclamationTriangleIcon } from "@/components/icons"

type EvidenceListProps = {
  assessment: Assessment
}

export default function EvidenceList({ assessment }: EvidenceListProps) {
  const hasMatched = assessment.matched.length > 0
  const hasMissing = assessment.missing.length > 0
  const hasConflicts = assessment.conflicts.length > 0
  const hasAmbiguous = assessment.ambiguous.length > 0

  if (!hasMatched && !hasMissing && !hasConflicts && !hasAmbiguous) {
    return null
  }

  return (
    <Card>
      <h3 className="mb-4 text-sm font-semibold text-surface-900">Evidence</h3>
      <div className="space-y-4">
        {hasMatched && (
          <div>
            <p className="mb-1.5 text-xs font-medium uppercase tracking-wider text-emerald-600">
              Matched
            </p>
            <ul className="space-y-1">
              {assessment.matched.map((item) => (
                <li key={item} className="flex items-center gap-2 text-sm text-surface-700">
                  <CheckIcon className="h-3.5 w-3.5 flex-shrink-0 text-emerald-500" />
                  {item}
                </li>
              ))}
            </ul>
          </div>
        )}
        {hasMissing && (
          <div>
            <p className="mb-1.5 text-xs font-medium uppercase tracking-wider text-surface-400">
              Missing
            </p>
            <ul className="space-y-1">
              {assessment.missing.map((item) => (
                <li key={item} className="flex items-center gap-2 text-sm text-surface-500">
                  <span className="block h-3.5 w-3.5 flex-shrink-0 rounded-full border border-surface-300" />
                  {item}
                </li>
              ))}
            </ul>
          </div>
        )}
        {hasAmbiguous && (
          <div>
            <p className="mb-1.5 text-xs font-medium uppercase tracking-wider text-orange-600">
              Ambiguous
            </p>
            <ul className="space-y-1">
              {assessment.ambiguous.map((item) => (
                <li key={item} className="flex items-center gap-2 text-sm text-orange-700">
                  <ExclamationTriangleIcon className="h-3.5 w-3.5 flex-shrink-0 text-orange-500" />
                  {item}
                </li>
              ))}
            </ul>
          </div>
        )}
        {hasConflicts && (
          <div>
            <p className="mb-1.5 text-xs font-medium uppercase tracking-wider text-red-600">
              Conflicts
            </p>
            <ul className="space-y-2">
              {assessment.conflicts.map((c, i) => (
                <li key={i} className="flex items-start gap-2 text-sm text-red-700">
                  <ExclamationTriangleIcon className="mt-0.5 h-3.5 w-3.5 flex-shrink-0 text-red-500" />
                  <span>{c.message}</span>
                </li>
              ))}
            </ul>
          </div>
        )}
      </div>
    </Card>
  )
}
