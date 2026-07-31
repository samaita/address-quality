import { useCallback, useRef } from "react"
import Container from "@/components/layout/Container"
import { PageHeader } from "@/components/kumo-ui"
import PlaygroundInput from "@/components/playground/PlaygroundInput"
import PlaygroundResults from "@/components/playground/PlaygroundResults"
import useValidateAddress from "@/hooks/useValidateAddress"

export default function Playground() {
  const { state, validate, clear } = useValidateAddress()
  const lastAddress = useRef<string>("")

  const handleValidate = useCallback(
    (address: string) => {
      lastAddress.current = address
      void validate(address)
    },
    [validate],
  )

  const handleRetry = useCallback(() => {
    if (lastAddress.current) {
      void validate(lastAddress.current)
    }
  }, [validate])

  return (
    <Container className="pb-16">
      <PageHeader
        title="Playground"
        description="Paste an Indonesian address and validate it against official data. See confidence, parsed hierarchy, evidence, and candidate matches in real time."
      />

      <div className="flex flex-col gap-8 lg:flex-row">
        <div className="w-full lg:w-[440px] xl:w-[480px]">
          <PlaygroundInput
            onValidate={handleValidate}
            onClear={clear}
            loading={state.status === "loading"}
            error={state.status === "error" ? state.error : null}
          />
        </div>

        <div className="min-w-0 flex-1">
          <PlaygroundResults state={state} onRetry={handleRetry} />
        </div>
      </div>
    </Container>
  )
}
