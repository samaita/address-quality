import Section from "@/components/kumo-ui/Section"
import CodeBlock from "@/components/docs/CodeBlock"

const requestCode = JSON.stringify(
  { address: "JL MERDEKA NO 56 CITARUM BANDUNG 40115" },
  null,
  2,
)

const responseCode = JSON.stringify(
  {
    timestamp: "2026-07-29T05:08:05Z",
    request_id: "019fac45-d6cb-7101-9159-76bd7c25867b",
    data: {
      address_id: "019fac45-d6cb-7153-aeef-742c66db6d18",
      status: "VALID",
      confidence: 0.97,
      raw_input: "JL MERDEKA NO 56 CITARUM BANDUNG 40115",
      normalized_input: "jl merdeka no citarum bandung 40115",
      formatted_address: "Citarum, Bandung Wetan, Kota Bandung, Jawa Barat 40115",
      location: {
        province: "Jawa Barat",
        city: "Kota Bandung",
        district: "Bandung Wetan",
        sub_district: "Citarum",
        postal_code: "40115",
      },
      assessment: {
        matched: ["province", "city", "district", "sub_district", "postal_code"],
        missing: ["road_name"],
        conflicts: [],
        ambiguous: [],
      },
      resolution: {
        strategy: ["top_down", "postal"],
        candidate_count: 1,
        candidates: [
          {
            uuid: "019fac45-d6d0-7e53-8e0d-a44f30d72a53",
            score: 0.97,
            location: {
              province: "Jawa Barat",
              city: "Kota Bandung",
              district: "Bandung Wetan",
              sub_district: "Citarum",
              postal_code: "40115",
            },
            reasons: ["exact_match", "match_postal_code_exact"],
          },
        ],
      },
      metadata: {
        location_source: "kemendagri",
        location_version: "2025",
      },
    },
  },
  null,
  2,
)

export default function ExampleSection() {
  return (
    <Section
      title="See it in action"
      description="Send an address and receive structured validation results with confidence scoring and resolution details."
    >
      <div className="grid gap-6 lg:grid-cols-2">
        <div>
          <p className="mb-3 text-sm font-medium text-surface-500">Request</p>
          <CodeBlock code={requestCode} language="json" title="POST /v1/validate" />
        </div>
        <div>
          <p className="mb-3 text-sm font-medium text-surface-500">Response</p>
          <CodeBlock code={responseCode} language="json" title="200 OK" />
        </div>
      </div>
    </Section>
  )
}
