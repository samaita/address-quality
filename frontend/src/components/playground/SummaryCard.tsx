import { useMemo } from "react"
import Card from "@/components/common/Card"
import Badge from "@/components/common/Badge"
import { Progress } from "@/components/kumo-ui"
import { getConfidenceTier, getStatusTone } from "@/lib/confidence"
import type { ResponseData } from "@/types/api"

type SummaryCardProps = {
  data: ResponseData
  requestId: string
}

const hierarchyFields: { key: keyof ResponseData["location"]; label: string }[] = [
  { key: "province", label: "Province" },
  { key: "city", label: "City" },
  { key: "district", label: "District" },
  { key: "sub_district", label: "Sub District" },
  { key: "postal_code", label: "Postal Code" },
]

export default function SummaryCard({ data, requestId }: SummaryCardProps) {
  const tier = getConfidenceTier(data.confidence)
  const percentage = Math.round(data.confidence * 100)

  const sourceLabel = useMemo(
    () =>
      [data.metadata.location_source, data.metadata.location_version]
        .filter(Boolean)
        .join(" · "),
    [data.metadata],
  )

  return (
    <Card>
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <p className="text-sm font-medium text-surface-500">Overall confidence</p>
          <div className="mt-1 flex items-baseline gap-3">
            <span className="text-4xl font-bold tracking-tight text-surface-900">
              {percentage}%
            </span>
            <Badge variant={getStatusTone(data.status)}>
              {data.status.toLowerCase()}
            </Badge>
          </div>
        </div>
        {sourceLabel && (
          <Badge variant="info" className="mt-1">
            {sourceLabel}
          </Badge>
        )}
      </div>

      <Progress
        className="mt-4"
        value={data.confidence}
        tone={tier}
        label="Overall confidence"
      />

      <dl className="mt-6 divide-y divide-surface-100">
        {hierarchyFields.map(({ key, label }) => (
          <div key={key} className="flex items-baseline justify-between py-2.5">
            <dt className="text-sm text-surface-500">{label}</dt>
            <dd className="text-sm font-medium text-surface-900">
              {data.location[key]}
            </dd>
          </div>
        ))}
      </dl>

      <div className="mt-6 flex items-center gap-2 border-t border-surface-100 pt-4">
        <span className="text-xs text-surface-400">request_id</span>
        <code className="truncate font-mono text-xs text-surface-600">
          {requestId}
        </code>
      </div>
    </Card>
  )
}
