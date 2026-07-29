import Container from "@/components/layout/Container"
import Card from "@/components/common/Card"
import { PageHeader } from "@/components/kumo-ui"
import AddressInput from "@/components/playground/AddressInput"
import ConfidenceBadge from "@/components/playground/ConfidenceBadge"
import ResultCard from "@/components/playground/ResultCard"
import EvidenceList from "@/components/playground/EvidenceList"
import CandidateTable from "@/components/playground/CandidateTable"
import JsonViewer from "@/components/playground/JsonViewer"
import useValidateAddress from "@/hooks/useValidateAddress"
import { ArrowPathIcon } from "@/components/icons"

export default function Playground() {
  const { state, validate, reset } = useValidateAddress()

  return (
    <Container className="pb-16">
      <PageHeader
        title="Playground"
        description="Test the API with example addresses. Submit an address and see the parsed result, confidence score, and candidate matches in real time."
      />

      <div className="flex flex-col gap-8 lg:flex-row">
        <div className="w-full lg:w-[440px] xl:w-[480px]">
          <AddressInput
            onValidate={validate}
            loading={state.status === "loading"}
          />
        </div>

        <div className="flex-1 min-w-0">
          {state.status === "idle" && (
            <Card className="flex flex-col items-center justify-center py-16 text-center">
              <p className="text-base font-medium text-surface-900">
                Ready to validate
              </p>
              <p className="mt-2 text-sm text-surface-500">
                Type an Indonesian address or click "Try Example" to get started.
              </p>
            </Card>
          )}

          {state.status === "loading" && (
            <Card className="flex flex-col items-center justify-center py-16 text-center">
              <ArrowPathIcon className="h-8 w-8 animate-spin text-accent-500" />
              <p className="mt-4 text-sm font-medium text-surface-700">
                Validating address...
              </p>
              <p className="mt-1 text-sm text-surface-500">
                Parsing and resolving against official data.
              </p>
            </Card>
          )}

          {state.status === "error" && (
            <Card className="border-red-200 bg-red-50">
              <p className="text-sm font-medium text-red-700">Validation failed</p>
              <p className="mt-1 text-sm text-red-600">{state.error}</p>
              <button
                type="button"
                onClick={reset}
                className="mt-3 text-sm font-medium text-red-700 underline hover:text-red-800"
              >
                Try again
              </button>
            </Card>
          )}

          {state.status === "success" && (() => {
            const { data: response } = state
            const d = response.data
            return (
              <div className="space-y-6">
                <ConfidenceBadge confidence={d.confidence} status={d.status} />
                <ResultCard
                  location={d.location}
                  formattedAddress={d.formatted_address}
                  normalizedInput={d.normalized_input}
                />
                <EvidenceList assessment={d.assessment} />
                <CandidateTable candidates={d.resolution.candidates} />
                <JsonViewer data={response} title="Raw JSON Response" />
              </div>
            )
          })()}
        </div>
      </div>
    </Container>
  )
}
