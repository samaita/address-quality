import Card from "@/components/common/Card"
import type { ResolutionCandidate } from "@/types/api"

type CandidateTableProps = {
  candidates: ResolutionCandidate[]
}

export default function CandidateTable({ candidates }: CandidateTableProps) {
  if (!candidates || candidates.length === 0) return null

  return (
    <Card>
      <h3 className="mb-4 text-sm font-semibold text-surface-900">
        Candidate Matches ({candidates.length})
      </h3>
      <div className="overflow-x-auto">
        <table className="w-full text-left text-sm">
          <thead>
            <tr className="border-b border-surface-200 text-xs font-medium uppercase tracking-wider text-surface-500">
              <th className="pb-2 pr-4">Score</th>
              <th className="pb-2 pr-4">Province</th>
              <th className="pb-2 pr-4">City</th>
              <th className="pb-2 pr-4">District</th>
              <th className="pb-2 pr-4">Subdistrict</th>
              <th className="pb-2">Reasons</th>
            </tr>
          </thead>
          <tbody>
            {candidates.map((c) => (
              <tr key={c.uuid} className="border-b border-surface-100 last:border-0">
                <td className="py-2.5 pr-4 font-medium text-surface-900">
                  {Math.round(c.score * 100)}%
                </td>
                <td className="py-2.5 pr-4 text-surface-700">{c.location.province}</td>
                <td className="py-2.5 pr-4 text-surface-700">{c.location.city}</td>
                <td className="py-2.5 pr-4 text-surface-700">{c.location.district}</td>
                <td className="py-2.5 pr-4 text-surface-700">{c.location.sub_district}</td>
                <td className="py-2.5 text-surface-500">
                  <span className="inline-flex flex-wrap gap-1">
                    {c.reasons.map((r) => (
                      <span
                        key={r}
                        className="rounded-full bg-surface-100 px-2 py-0.5 text-xs text-surface-600"
                      >
                        {r.replace(/_/g, " ")}
                      </span>
                    ))}
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </Card>
  )
}
