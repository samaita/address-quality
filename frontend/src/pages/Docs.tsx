import Container from "@/components/layout/Container"
import Badge from "@/components/common/Badge"
import { PageHeader } from "@/components/kumo-ui"
import Sidebar from "@/components/docs/Sidebar"
import EndpointCard from "@/components/docs/EndpointCard"
import CodeBlock from "@/components/docs/CodeBlock"

const reqCode = JSON.stringify({ address: "JL MERDEKA NO 56 CITARUM BANDUNG 40115" }, null, 2)

const reqWithSource = JSON.stringify(
  { address: "JL MERDEKA NO 56 CITARUM BANDUNG 40115", source_code: "kemendagri" },
  null,
  2,
)

const successCode = JSON.stringify(
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
      location: { province: "Jawa Barat", city: "Kota Bandung", district: "Bandung Wetan", sub_district: "Citarum", postal_code: "40115" },
      assessment: { matched: ["province", "city", "district", "sub_district", "postal_code"], missing: ["road_name"], conflicts: [], ambiguous: [] },
      resolution: { strategy: ["top_down", "postal"], candidate_count: 1, candidates: [{ uuid: "019fac45-...", score: 0.97, location: { province: "Jawa Barat", city: "Kota Bandung", district: "Bandung Wetan", sub_district: "Citarum", postal_code: "40115" }, reasons: ["exact_match", "match_postal_code_exact"] }] },
      metadata: { location_source: "kemendagri", location_version: "2025" },
    },
  },
  null,
  2,
)

const errorCode = JSON.stringify(
  { timestamp: "2026-07-29T05:06:43Z", request_id: "019fac44-95c0-79cb-b1d4-649463403ea7", error: "invalid request body" },
  null,
  2,
)

const authErrorCode = JSON.stringify(
  { timestamp: "2026-07-29T05:06:43Z", request_id: "019fac44-95c0-79cb-b1d4-649463403ea7", error: "missing or invalid API key" },
  null,
  2,
)

function SectionTitle({ children }: { children: React.ReactNode }) {
  return <h3 className="text-lg font-semibold text-surface-900">{children}</h3>
}

function Table({ children }: { children: React.ReactNode }) {
  return (
    <div className="overflow-x-auto rounded-xl border border-surface-200">
      <table className="w-full text-left text-sm">{children}</table>
    </div>
  )
}

export default function Docs() {
  return (
    <Container className="pb-24">
      <PageHeader
        title="Documentation"
        description="Learn how to integrate the Address Quality API into your application. Get started with authentication, endpoints, request schemas, and examples."
      />

      <div className="flex gap-12">
        <Sidebar />

        <div className="flex-1 min-w-0 max-w-3xl space-y-16">
          {/* Introduction */}
          <section id="introduction" className="scroll-mt-20">
            <h2 className="text-2xl font-semibold tracking-tight text-surface-900">
              Introduction
            </h2>
            <p className="mt-4 leading-relaxed text-surface-600">
              Address Quality is a REST API that validates Indonesian addresses. Given a raw address
              string, the API parses it, cross-references components against the official Kemendagri
              administrative database, and returns structured results with confidence scores,
              validation status, and explainable evidence.
            </p>
            <p className="mt-3 leading-relaxed text-surface-600">
              The API is designed for logistics platforms, KYC systems, geocoding pipelines, and any
              application that needs reliable Indonesian address data.
            </p>
          </section>

          {/* Quick Start */}
          <section id="quickstart" className="scroll-mt-20">
            <h2 className="text-2xl font-semibold tracking-tight text-surface-900">
              Quick Start
            </h2>
            <p className="mt-4 leading-relaxed text-surface-600">
              Get started in under a minute. Sign up for an API key, then make your first request.
            </p>
            <div className="mt-6 space-y-4">
              <SectionTitle>1. Get your API key</SectionTitle>
              <p className="text-sm leading-relaxed text-surface-600">
                Sign up at the developer dashboard to receive your API key. Include it in the
                <code className="mx-1 rounded bg-surface-100 px-1.5 py-0.5 font-mono text-xs text-surface-700">X-API-Key</code>
                header of every request.
              </p>
              <SectionTitle>2. Make your first request</SectionTitle>
              <CodeBlock
                code={`curl -X POST https://api.addressquality.dev/v1/validate \\\n  -H "Content-Type: application/json" \\\n  -H "X-API-Key: your-api-key" \\\n  -d '{"address": "JL MERDEKA NO 56 CITARUM BANDUNG 40115"}'`}
                language="bash"
                title="Terminal"
              />
            </div>
          </section>

          {/* Authentication */}
          <section id="authentication" className="scroll-mt-20">
            <h2 className="text-2xl font-semibold tracking-tight text-surface-900">
              Authentication
            </h2>
            <p className="mt-4 leading-relaxed text-surface-600">
              All API requests require authentication via an API key. Pass your key in the
              <code className="mx-1 rounded bg-surface-100 px-1.5 py-0.5 font-mono text-xs text-surface-700">X-API-Key</code>
              HTTP header.
            </p>
            <div className="mt-6 space-y-4">
              <Table>
                <thead>
                  <tr className="border-b border-surface-200 bg-surface-50">
                    <th className="px-4 py-2.5 text-xs font-medium uppercase tracking-wider text-surface-500">Header</th>
                    <th className="px-4 py-2.5 text-xs font-medium uppercase tracking-wider text-surface-500">Value</th>
                    <th className="px-4 py-2.5 text-xs font-medium uppercase tracking-wider text-surface-500">Required</th>
                  </tr>
                </thead>
                <tbody>
                  <tr className="border-b border-surface-100">
                    <td className="px-4 py-3 font-mono text-xs text-surface-900">X-API-Key</td>
                    <td className="px-4 py-3 text-sm text-surface-700">Your API key</td>
                    <td className="px-4 py-3"><Badge>Yes</Badge></td>
                  </tr>
                </tbody>
              </Table>
              <p className="text-sm leading-relaxed text-surface-600">
                Requests without a valid API key return a
                <code className="mx-1 rounded bg-surface-100 px-1.5 py-0.5 font-mono text-xs text-surface-700">401 Unauthorized</code>
                status.
              </p>
            </div>
          </section>

          {/* POST /v1/validate */}
          <section id="validate" className="scroll-mt-20">
            <h2 className="text-2xl font-semibold tracking-tight text-surface-900">
              POST /v1/validate
            </h2>
            <p className="mt-4 leading-relaxed text-surface-600">
              The primary endpoint for address validation. Accepts a raw address string and returns
              structured validation results.
            </p>
            <div className="mt-6">
              <EndpointCard
                method="POST"
                path="/v1/validate"
                description="Validate and resolve an Indonesian address against the official location database."
                requestCode={reqCode}
                responseCode={successCode}
                errorCode={errorCode}
              />
            </div>
          </section>

          {/* Request Schema */}
          <section id="request-schema" className="scroll-mt-20">
            <h2 className="text-2xl font-semibold tracking-tight text-surface-900">
              Request Schema
            </h2>
            <p className="mt-4 leading-relaxed text-surface-600">
              Send a JSON object with the address to validate. All other fields are optional.
            </p>
            <div className="mt-6 space-y-6">
              <Table>
                <thead>
                  <tr className="border-b border-surface-200 bg-surface-50">
                    <th className="px-4 py-2.5 text-xs font-medium uppercase tracking-wider text-surface-500">Field</th>
                    <th className="px-4 py-2.5 text-xs font-medium uppercase tracking-wider text-surface-500">Type</th>
                    <th className="px-4 py-2.5 text-xs font-medium uppercase tracking-wider text-surface-500">Required</th>
                    <th className="px-4 py-2.5 text-xs font-medium uppercase tracking-wider text-surface-500">Description</th>
                  </tr>
                </thead>
                <tbody>
                  <tr className="border-b border-surface-100">
                    <td className="px-4 py-3 font-mono text-xs font-medium text-surface-900">address</td>
                    <td className="px-4 py-3 font-mono text-xs text-surface-600">string</td>
                    <td className="px-4 py-3"><Badge>Yes</Badge></td>
                    <td className="px-4 py-3 text-sm text-surface-600">Raw Indonesian address text</td>
                  </tr>
                  <tr>
                    <td className="px-4 py-3 font-mono text-xs font-medium text-surface-900">source_code</td>
                    <td className="px-4 py-3 font-mono text-xs text-surface-600">string</td>
                    <td className="px-4 py-3"><Badge variant="default">No</Badge></td>
                    <td className="px-4 py-3 text-sm text-surface-600">Location data source (default: kemendagri)</td>
                  </tr>
                </tbody>
              </Table>
              <SectionTitle>Example with source code</SectionTitle>
              <CodeBlock code={reqWithSource} title="POST /v1/validate" />
            </div>
          </section>

          {/* Response Schema */}
          <section id="response-schema" className="scroll-mt-20">
            <h2 className="text-2xl font-semibold tracking-tight text-surface-900">
              Response Schema
            </h2>
            <p className="mt-4 leading-relaxed text-surface-600">
              A successful response returns the validation results wrapped in a standard envelope.
            </p>
            <div className="mt-6 space-y-6">
              <Table>
                <thead>
                  <tr className="border-b border-surface-200 bg-surface-50">
                    <th className="px-4 py-2.5 text-xs font-medium uppercase tracking-wider text-surface-500">Field</th>
                    <th className="px-4 py-2.5 text-xs font-medium uppercase tracking-wider text-surface-500">Type</th>
                    <th className="px-4 py-2.5 text-xs font-medium uppercase tracking-wider text-surface-500">Description</th>
                  </tr>
                </thead>
                <tbody>
                  <tr className="border-b border-surface-100">
                    <td className="px-4 py-3 font-mono text-xs font-medium text-surface-900">timestamp</td>
                    <td className="px-4 py-3 font-mono text-xs text-surface-600">string</td>
                    <td className="px-4 py-3 text-sm text-surface-600">ISO 8601 UTC timestamp</td>
                  </tr>
                  <tr className="border-b border-surface-100">
                    <td className="px-4 py-3 font-mono text-xs font-medium text-surface-900">request_id</td>
                    <td className="px-4 py-3 font-mono text-xs text-surface-600">string</td>
                    <td className="px-4 py-3 text-sm text-surface-600">Unique request identifier (UUIDv7)</td>
                  </tr>
                  <tr className="border-b border-surface-100">
                    <td className="px-4 py-3 font-mono text-xs font-medium text-surface-900">data.address_id</td>
                    <td className="px-4 py-3 font-mono text-xs text-surface-600">string</td>
                    <td className="px-4 py-3 text-sm text-surface-600">Unique address record identifier</td>
                  </tr>
                  <tr className="border-b border-surface-100">
                    <td className="px-4 py-3 font-mono text-xs font-medium text-surface-900">data.status</td>
                    <td className="px-4 py-3 font-mono text-xs text-surface-600">string</td>
                    <td className="px-4 py-3 text-sm text-surface-600">VALID, INCOMPLETE, AMBIGUOUS, CONFLICT, or UNKNOWN</td>
                  </tr>
                  <tr className="border-b border-surface-100">
                    <td className="px-4 py-3 font-mono text-xs font-medium text-surface-900">data.confidence</td>
                    <td className="px-4 py-3 font-mono text-xs text-surface-600">number</td>
                    <td className="px-4 py-3 text-sm text-surface-600">Confidence score from 0.0 to 1.0</td>
                  </tr>
                  <tr className="border-b border-surface-100">
                    <td className="px-4 py-3 font-mono text-xs font-medium text-surface-900">data.raw_input</td>
                    <td className="px-4 py-3 font-mono text-xs text-surface-600">string</td>
                    <td className="px-4 py-3 text-sm text-surface-600">Original address as submitted</td>
                  </tr>
                  <tr className="border-b border-surface-100">
                    <td className="px-4 py-3 font-mono text-xs font-medium text-surface-900">data.normalized_input</td>
                    <td className="px-4 py-3 font-mono text-xs text-surface-600">string</td>
                    <td className="px-4 py-3 text-sm text-surface-600">Normalized version of the input</td>
                  </tr>
                  <tr className="border-b border-surface-100">
                    <td className="px-4 py-3 font-mono text-xs font-medium text-surface-900">data.formatted_address</td>
                    <td className="px-4 py-3 font-mono text-xs text-surface-600">string</td>
                    <td className="px-4 py-3 text-sm text-surface-600">Standardized formatted address</td>
                  </tr>
                  <tr className="border-b border-surface-100">
                    <td className="px-4 py-3 font-mono text-xs font-medium text-surface-900">data.location</td>
                    <td className="px-4 py-3 font-mono text-xs text-surface-600">object</td>
                    <td className="px-4 py-3 text-sm text-surface-600">Parsed administrative hierarchy</td>
                  </tr>
                  <tr className="border-b border-surface-100">
                    <td className="px-4 py-3 font-mono text-xs font-medium text-surface-900">data.assessment</td>
                    <td className="px-4 py-3 font-mono text-xs text-surface-600">object</td>
                    <td className="px-4 py-3 text-sm text-surface-600">Matched, missing, conflicts, ambiguous</td>
                  </tr>
                  <tr className="border-b border-surface-100">
                    <td className="px-4 py-3 font-mono text-xs font-medium text-surface-900">data.resolution</td>
                    <td className="px-4 py-3 font-mono text-xs text-surface-600">object</td>
                    <td className="px-4 py-3 text-sm text-surface-600">Resolution strategy and candidate matches</td>
                  </tr>
                  <tr>
                    <td className="px-4 py-3 font-mono text-xs font-medium text-surface-900">data.metadata</td>
                    <td className="px-4 py-3 font-mono text-xs text-surface-600">object</td>
                    <td className="px-4 py-3 text-sm text-surface-600">Data source and version information</td>
                  </tr>
                </tbody>
              </Table>
            </div>
          </section>

          {/* Error Responses */}
          <section id="errors" className="scroll-mt-20">
            <h2 className="text-2xl font-semibold tracking-tight text-surface-900">
              Error Responses
            </h2>
            <p className="mt-4 leading-relaxed text-surface-600">
              The API uses standard HTTP status codes to indicate success or failure.
            </p>
            <div className="mt-6 space-y-6">
              <Table>
                <thead>
                  <tr className="border-b border-surface-200 bg-surface-50">
                    <th className="px-4 py-2.5 text-xs font-medium uppercase tracking-wider text-surface-500">Status</th>
                    <th className="px-4 py-2.5 text-xs font-medium uppercase tracking-wider text-surface-500">Description</th>
                  </tr>
                </thead>
                <tbody>
                  <tr className="border-b border-surface-100">
                    <td className="px-4 py-3"><Badge variant="warning">400</Badge></td>
                    <td className="px-4 py-3 text-sm text-surface-600">Bad Request — invalid or missing address field</td>
                  </tr>
                  <tr className="border-b border-surface-100">
                    <td className="px-4 py-3"><Badge variant="danger">401</Badge></td>
                    <td className="px-4 py-3 text-sm text-surface-600">Unauthorized — missing or invalid API key</td>
                  </tr>
                  <tr className="border-b border-surface-100">
                    <td className="px-4 py-3"><Badge variant="warning">429</Badge></td>
                    <td className="px-4 py-3 text-sm text-surface-600">Too Many Requests — rate limit exceeded</td>
                  </tr>
                  <tr>
                    <td className="px-4 py-3"><Badge variant="danger">500</Badge></td>
                    <td className="px-4 py-3 text-sm text-surface-600">Internal Server Error — unexpected server failure</td>
                  </tr>
                </tbody>
              </Table>
              <SectionTitle>Error response format</SectionTitle>
              <CodeBlock code={authErrorCode} title="401 Unauthorized" />
            </div>
          </section>

          {/* Rate Limits */}
          <section id="rate-limits" className="scroll-mt-20">
            <h2 className="text-2xl font-semibold tracking-tight text-surface-900">
              Rate Limits
            </h2>
            <p className="mt-4 leading-relaxed text-surface-600">
              Rate limits are enforced per API key. Exceeding the limit returns a
              <code className="mx-1 rounded bg-surface-100 px-1.5 py-0.5 font-mono text-xs text-surface-700">429 Too Many Requests</code>
              response.
            </p>
            <div className="mt-6">
              <Table>
                <thead>
                  <tr className="border-b border-surface-200 bg-surface-50">
                    <th className="px-4 py-2.5 text-xs font-medium uppercase tracking-wider text-surface-500">Tier</th>
                    <th className="px-4 py-2.5 text-xs font-medium uppercase tracking-wider text-surface-500">Requests / minute</th>
                  </tr>
                </thead>
                <tbody>
                  <tr className="border-b border-surface-100">
                    <td className="px-4 py-3 text-sm font-medium text-surface-900">Free</td>
                    <td className="px-4 py-3 text-sm text-surface-600">100</td>
                  </tr>
                  <tr className="border-b border-surface-100">
                    <td className="px-4 py-3 text-sm font-medium text-surface-900">Pro</td>
                    <td className="px-4 py-3 text-sm text-surface-600">1,000</td>
                  </tr>
                  <tr>
                    <td className="px-4 py-3 text-sm font-medium text-surface-900">Enterprise</td>
                    <td className="px-4 py-3 text-sm text-surface-600">Custom</td>
                  </tr>
                </tbody>
              </Table>
            </div>
          </section>

          {/* FAQ */}
          <section id="faq" className="scroll-mt-20">
            <h2 className="text-2xl font-semibold tracking-tight text-surface-900">
              FAQ
            </h2>
            <div className="mt-6 space-y-6">
              <div>
                <SectionTitle>What address formats does the API support?</SectionTitle>
                <p className="mt-2 text-sm leading-relaxed text-surface-600">
                  The API handles most Indonesian address formats, including structured addresses with
                  road names, landmarks, administrative areas, and postal codes. It normalizes input
                  before parsing to handle variations in spelling and abbreviations.
                </p>
              </div>
              <div>
                <SectionTitle>How is the confidence score calculated?</SectionTitle>
                <p className="mt-2 text-sm leading-relaxed text-surface-600">
                  The confidence score reflects the proportion of address components that were
                  successfully matched against the official database. A score of 1.0 (100%) means all
                  components matched. The score decreases when components are missing or conflicting.
                </p>
              </div>
              <div>
                <SectionTitle>What location data source is used?</SectionTitle>
                <p className="mt-2 text-sm leading-relaxed text-surface-600">
                  The API uses the official Kemendagri (Ministry of Home Affairs) administrative
                  database, which includes all provinces, cities, districts, subdistricts, and
                  postal codes in Indonesia.
                </p>
              </div>
              <div>
                <SectionTitle>What does the AMBIGUOUS status mean?</SectionTitle>
                <p className="mt-2 text-sm leading-relaxed text-surface-600">
                  AMBIGUOUS means the API identified multiple possible locations that match parts of
                  the input. The response includes candidate matches with individual confidence
                  scores to help disambiguate.
                </p>
              </div>
              <div>
                <SectionTitle>Can I use the API in production?</SectionTitle>
                <p className="mt-2 text-sm leading-relaxed text-surface-600">
                  Yes. The API is designed for production use with high availability, rate limiting,
                  and comprehensive error handling. Contact us for enterprise SLAs and custom rate
                  limits.
                </p>
              </div>
              <div>
                <SectionTitle>How do I report incorrect data?</SectionTitle>
                <p className="mt-2 text-sm leading-relaxed text-surface-600">
                  If you encounter incorrect or outdated administrative data, please open an issue on
                  the GitHub repository. We regularly update our database from official sources.
                </p>
              </div>
            </div>
          </section>
        </div>
      </div>
    </Container>
  )
}
