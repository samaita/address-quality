import Container from "@/components/layout/Container"
import { PageHeader } from "@/components/kumo-ui"

export default function Landing() {
  return (
    <Container>
      <PageHeader
        title="Validate Indonesian Addresses with Confidence"
        description="Address Quality API validates and resolves Indonesian addresses against the official administrative hierarchy. Get structured data, confidence scores, and explainable results from raw address input."
      />
    </Container>
  )
}
