import type { QualityStatus } from "@/types/api"
import type { BadgeVariant } from "@/components/common/Badge"

export type ConfidenceTier = "green" | "yellow" | "red"

export function getConfidenceTier(value: number): ConfidenceTier {
  if (value > 0.9) return "green"
  if (value >= 0.7) return "yellow"
  return "red"
}

export function getStatusTone(status: QualityStatus): BadgeVariant {
  switch (status) {
    case "VALID":
      return "success"
    case "INCOMPLETE":
      return "warning"
    case "AMBIGUOUS":
      return "warning"
    case "CONFLICT":
      return "danger"
    case "UNKNOWN":
      return "default"
  }
}
