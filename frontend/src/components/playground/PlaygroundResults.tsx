import EmptyState from "@/components/playground/EmptyState"
import LoadingState from "@/components/playground/LoadingState"
import ErrorState from "@/components/playground/ErrorState"
import SummaryCard from "@/components/playground/SummaryCard"
import EvidenceTable from "@/components/playground/EvidenceTable"
import CandidateList from "@/components/playground/CandidateList"
import RawResponse from "@/components/playground/RawResponse"
import type { ApiError } from "@/types/api"
import type { AddressResponse } from "@/types/api"

export type PlaygroundState =
  | { status: "idle" }
  | { status: "loading" }
  | { status: "success"; data: AddressResponse }
  | { status: "error"; error: Pick<ApiError, "kind" | "message" | "status"> }

type PlaygroundResultsProps = {
  state: PlaygroundState
  onRetry: () => void
}

export default function PlaygroundResults({ state, onRetry }: PlaygroundResultsProps) {
  switch (state.status) {
    case "idle":
      return <EmptyState />
    case "loading":
      return <LoadingState />
    case "error":
      return <ErrorState error={state.error} onRetry={onRetry} />
    case "success": {
      const { data: response } = state
      const d = response.data
      return (
        <div className="space-y-6">
          <SummaryCard data={d} requestId={response.request_id} />
          <EvidenceTable data={d} />
          <CandidateList candidates={d.resolution.candidates} />
          <RawResponse data={response} />
        </div>
      )
    }
  }
}
