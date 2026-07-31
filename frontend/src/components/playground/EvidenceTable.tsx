import Card from "@/components/common/Card"
import Badge from "@/components/common/Badge"
import { Alert, Table, TBody, Td, Th, THead, Tr } from "@/components/kumo-ui"
import { buildEvidenceRows } from "@/lib/evidence"
import type { ResponseData } from "@/types/api"

type EvidenceTableProps = {
  data: ResponseData
}

const statusBadge: Record<
  "matched" | "partial" | "missing",
  { variant: "default" | "success" | "warning" | "danger" | "info"; label: string }
> = {
  matched: { variant: "success", label: "Matched" },
  partial: { variant: "warning", label: "Partial" },
  missing: { variant: "default", label: "Missing" },
}

export default function EvidenceTable({ data }: EvidenceTableProps) {
  const rows = buildEvidenceRows(data)

  if (rows.length === 0 && data.assessment.conflicts.length === 0) return null

  return (
    <Card>
      <h3 className="mb-4 text-sm font-semibold text-surface-900">
        Evidence Breakdown
      </h3>

      {rows.length > 0 && (
        <Table>
          <THead>
            <Tr>
              <Th>Evidence</Th>
              <Th>Extracted Value</Th>
              <Th>Confidence</Th>
              <Th>Status</Th>
            </Tr>
          </THead>
          <TBody>
            {rows.map((row) => {
              const badge = statusBadge[row.status]
              return (
                <Tr key={row.field}>
                  <Td className="font-medium text-surface-900">{row.label}</Td>
                  <Td className={row.value ? "" : "text-surface-400"}>
                    {row.value ?? "—"}
                  </Td>
                  <Td className="font-mono">
                    {row.confidence != null ? row.confidence.toFixed(2) : "—"}
                  </Td>
                  <Td>
                    <Badge variant={badge.variant}>{badge.label}</Badge>
                  </Td>
                </Tr>
              )
            })}
          </TBody>
        </Table>
      )}

      {data.assessment.conflicts.length > 0 && (
        <Alert variant="danger" title="Conflicts" className="mt-4">
          <ul className="space-y-1">
            {data.assessment.conflicts.map((c) => (
              <li key={c.type} className="text-sm">
                {c.message}
              </li>
            ))}
          </ul>
        </Alert>
      )}
    </Card>
  )
}
