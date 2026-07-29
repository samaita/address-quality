import Container from "@/components/layout/Container"
import { PageHeader } from "@/components/kumo-ui"

export default function Playground() {
  return (
    <Container>
      <PageHeader
        title="Playground"
        description="Test the API with example addresses. Submit an address and see the parsed result, confidence score, and candidate matches in real time."
      />
    </Container>
  )
}
