import Card from "@/components/common/Card"
import { Badge } from "@cloudflare/kumo/components/badge"
import { Banner } from "@cloudflare/kumo/components/banner"
import { Table } from "@cloudflare/kumo/components/table"
import { buildEvidenceRows } from "@/lib/evidence"
import type { ResponseData } from "@/types/api"

type EvidenceTableProps = {
  data: ResponseData
}

const statusBadge: Record<
  "matched" | "partial" | "missing",
  {
    variant:
      | "primary"
      | "secondary"
      | "error"
      | "warning"
      | "success"
      | "info"
      | "neutral"
    label: string
  }
> = {
  matched: { variant: "success", label: "Matched" },
  partial: { variant: "warning", label: "Partial" },
  missing: { variant: "neutral", label: "Missing" },
}

export default function EvidenceTable({ data }: EvidenceTableProps) {
  const rows = buildEvidenceRows(data)
  const conflicts = data.assessment.conflicts ?? []

  if (rows.length === 0 && conflicts.length === 0) return null

  return (
    <Card>
      <h3 className="mb-4 text-sm font-semibold text-surface-900">
        Evidence Breakdown
      </h3>

      {rows.length > 0 && (
        <Table>
          <Table.Header>
            <Table.Row>
              <Table.Head>Evidence</Table.Head>
              <Table.Head>Extracted Value</Table.Head>
              <Table.Head>Confidence</Table.Head>
              <Table.Head>Status</Table.Head>
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {rows.map((row) => {
              const badge = statusBadge[row.status]
              return (
                <Table.Row key={row.field}>
                  <Table.Cell className="font-medium text-surface-900">
                    {row.label}
                  </Table.Cell>
                  <Table.Cell className={row.value ? "" : "text-surface-400"}>
                    {row.value ?? "—"}
                  </Table.Cell>
                  <Table.Cell className="font-mono">
                    {row.confidence != null ? row.confidence.toFixed(2) : "—"}
                  </Table.Cell>
                  <Table.Cell>
                    <Badge variant={badge.variant}>{badge.label}</Badge>
                  </Table.Cell>
                </Table.Row>
              )
            })}
          </Table.Body>
        </Table>
      )}

      {conflicts.length > 0 && (
        <Banner
          variant="error"
          title="Conflicts"
          className="mt-4"
          description={
            <ul className="space-y-1">
              {conflicts.map((c) => (
                <li key={c.type} className="text-sm">
                  {c.message}
                </li>
              ))}
            </ul>
          }
        />
      )}
    </Card>
  )
}
