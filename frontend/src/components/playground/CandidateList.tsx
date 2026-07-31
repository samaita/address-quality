import { useMemo } from "react"
import Card from "@/components/common/Card"
import Badge from "@/components/common/Badge"
import { Progress } from "@/components/kumo-ui"
import { getConfidenceTier } from "@/lib/confidence"
import type { ResolutionCandidate } from "@/types/api"

type CandidateListProps = {
  candidates: ResolutionCandidate[]
}

const fields: { key: keyof ResolutionCandidate["location"]; label: string }[] = [
  { key: "province", label: "Province" },
  { key: "city", label: "City" },
  { key: "district", label: "District" },
  { key: "sub_district", label: "Sub District" },
  { key: "postal_code", label: "Postal Code" },
]

export default function CandidateList({ candidates }: CandidateListProps) {
  const sorted = useMemo(
    () => [...candidates].sort((a, b) => b.score - a.score),
    [candidates],
  )

  if (sorted.length === 0) return null

  const bestScore = sorted[0].score

  return (
    <div className="space-y-4">
      <h3 className="text-sm font-semibold text-surface-900">
        Candidate Matches ({sorted.length})
      </h3>

      {sorted.map((candidate) => {
        const isBest = candidate.score === bestScore
        return (
          <Card
            key={candidate.uuid}
            className={`relative ${isBest ? "ring-2 ring-accent-500" : ""}`}
          >
            <div className="flex items-start justify-between gap-4">
              <div className="flex items-center gap-3">
                <span className="text-2xl font-bold tracking-tight text-surface-900">
                  {Math.round(candidate.score * 100)}%
                </span>
                {isBest && (
                  <Badge variant="info">Best match</Badge>
                )}
              </div>
              <div className="w-32">
                <Progress
                  value={candidate.score}
                  tone={getConfidenceTier(candidate.score)}
                  label={`Candidate confidence ${Math.round(candidate.score * 100)}%`}
                />
              </div>
            </div>

            <dl className="mt-5 grid grid-cols-1 gap-x-6 gap-y-3 sm:grid-cols-2">
              {fields.map(({ key, label }) => (
                <div key={key} className="flex items-baseline justify-between gap-4">
                  <dt className="text-xs text-surface-400">{label}</dt>
                  <dd className="text-right text-sm font-medium text-surface-900">
                    {candidate.location[key]}
                  </dd>
                </div>
              ))}
            </dl>

            {candidate.reasons.length > 0 && (
              <div className="mt-5 flex flex-wrap gap-1.5">
                {candidate.reasons.map((reason) => (
                  <span
                    key={reason}
                    className="inline-flex rounded-full bg-surface-100 px-2 py-0.5 text-xs text-surface-600"
                  >
                    {reason.replace(/_/g, " ")}
                  </span>
                ))}
              </div>
            )}
          </Card>
        )
      })}
    </div>
  )
}
