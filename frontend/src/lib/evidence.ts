import type { EvidenceRow, ResponseData } from "@/types/api"

const FIELD_DEFS: {
  field: string
  label: string
  locationKey?: keyof ResponseData["location"]
}[] = [
  { field: "province", label: "Province", locationKey: "province" },
  { field: "city", label: "City", locationKey: "city" },
  { field: "district", label: "District", locationKey: "district" },
  { field: "sub_district", label: "Sub District", locationKey: "sub_district" },
  { field: "postal_code", label: "Postal Code", locationKey: "postal_code" },
  { field: "road_name", label: "Road" },
]

function resolveStatus(data: ResponseData, field: string): EvidenceRow["status"] | null {
  if (data.assessment.matched.includes(field)) return "matched"
  if (data.assessment.ambiguous.includes(field)) return "partial"
  if (data.assessment.missing.includes(field)) return "missing"
  return null
}

export function buildEvidenceRows(data: ResponseData): EvidenceRow[] {
  const rows: EvidenceRow[] = []

  for (const def of FIELD_DEFS) {
    const status = resolveStatus(data, def.field)
    if (!status) continue

    const locationKey = def.locationKey
    const locationValue = locationKey != null ? data.location[locationKey] : undefined
    const hasValue =
      typeof locationValue === "string" && locationValue.length > 0
    const value =
      status === "missing" || !hasValue ? null : locationValue

    rows.push({
      field: def.field,
      label: def.label,
      value,
      confidence:
        status === "matched" ? 1 : status === "partial" ? 0.7 : null,
      status,
    })
  }

  return rows
}
