import Container from "@/components/layout/Container"
import Badge from "@/components/common/Badge"
import { PageHeader } from "@/components/kumo-ui"
import Sidebar from "@/components/docs/Sidebar"
import CodeBlock from "@/components/docs/CodeBlock"
import Mermaid from "@/components/docs/Mermaid"
import "prismjs/components/prism-http"

const reqCode = JSON.stringify({ address: "JL MERDEKA NO 56 CITARUM BANDUNG 40115" }, null, 2)

const reqHttp = `POST /v1/validate
Content-Type: application/json

${reqCode}`

const reqWithSource = JSON.stringify(
  { address: "JL MERDEKA NO 56 CITARUM BANDUNG 40115", source_code: "kemendagri" },
  null,
  2,
)

const successCode = `{
  "timestamp": "2026-07-29T05:08:05Z",
  "request_id": "019fac45-d6cb-7101-9159-76bd7c25867b",
  "data": {
    "address_id": "019fac45-d6cb-7153-aeef-742c66db6d18",
    "status": "VALID",
    "confidence": 0.97,
    "raw_input": "JL MERDEKA NO 56 CITARUM BANDUNG 40115",
    "normalized_input": "jl merdeka no citarum bandung 40115",
    "formatted_address": "Citarum, Bandung Wetan, Kota Bandung, Jawa Barat 40115",
    "location": { "province": "Jawa Barat", "city": "Kota Bandung", "district": "Bandung Wetan", "sub_district": "Citarum", "postal_code": "40115" },
    "assessment": { "matched": ["province", "city", "district", "sub_district", "postal_code"], "missing": ["road_name"], "conflicts": [], "ambiguous": [] },
    "resolution": { "strategy": ["top_down", "postal"], "candidate_count": 1, "candidates": [{ "score": 0.97, "reasons": ["exact_match", "match_postal_code_exact"] }] },
    "metadata": { "location_source": "kemendagri", "location_version": "2025" }
  }
}`

const errorCode = `{
  "timestamp": "2026-07-29T05:06:43Z",
  "request_id": "019fac44-95c0-79cb-b1d4-649463403ea7",
  "error": "missing or invalid API key"
}`

const pipelineSteps = [
  {
    step: "Normalize Input",
    description:
      "Cleans and standardizes the input by handling casing, whitespace, abbreviations, and common variations.",
  },
  {
    step: "Extract Address Components",
    description:
      "Identifies administrative components such as province, city, district, subdistrict, and postal code.",
  },
  {
    step: "Validate Administrative Hierarchy",
    description:
      "Verifies that the extracted components form a valid administrative hierarchy using the selected location dataset.",
  },
  {
    step: "Resolve Candidate Locations",
    description:
      "When multiple valid matches exist, ranks candidates and selects the most likely location.",
  },
  {
    step: "Compute Confidence",
    description: "Calculates a confidence score based on the available validation evidence.",
  },
  {
    step: "Generate Structured Response",
    description:
      "Returns the resolved address, confidence score, assessment, candidate information, and metadata.",
  },
]

const pipelineFlow = `flowchart LR
    A["Raw Address"]
    --> B["Normalize Input"]
    --> C["Extract Address Components"]
    --> D["Validate Administrative Hierarchy"]
    --> E["Resolve Candidate Locations"]
    --> F["Compute Confidence"]
    --> G["Generate Structured Response"]`

function SectionTitle({ children }: { children: React.ReactNode }) {
  return <h3 className="text-lg font-semibold text-surface-900">{children}</h3>
}

function Callout({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-800">
      <div className="flex items-center gap-2">
        <Badge variant="warning">Warning</Badge>
        <span className="font-semibold">{title}</span>
      </div>
      <div className="mt-1 leading-relaxed">{children}</div>
    </div>
  )
}

function Table({ children }: { children: React.ReactNode }) {
  return (
    <div className="overflow-x-auto rounded-xl border border-surface-200">
      <table className="w-full text-left text-sm">{children}</table>
    </div>
  )
}

function Th({ children }: { children: React.ReactNode }) {
  return (
    <th className="px-4 py-2.5 text-xs font-medium uppercase tracking-wider text-surface-500">
      {children}
    </th>
  )
}

function Td({ children, className = "" }: { children: React.ReactNode; className?: string }) {
  return <td className={`px-4 py-3 ${className}`}>{children}</td>
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
            <h2 className="text-2xl font-semibold tracking-tight text-surface-900">Introduction</h2>
            <p className="mt-4 leading-relaxed text-surface-600">
              Address Quality is an API for validating and resolving Indonesian addresses.
            </p>
            <p className="mt-3 leading-relaxed text-surface-600">
              Instead of simply parsing text, the API validates administrative hierarchy against
              official location data, resolves ambiguous matches, and returns structured results
              with confidence scores and explainable evidence.
            </p>
            <p className="mt-3 leading-relaxed text-surface-600">
              Whether you're building a logistics platform, KYC workflow, customer onboarding, or
              data cleaning pipeline, Address Quality helps transform inconsistent address text into
              reliable structured data.
            </p>
            <div className="mt-6">
              <Callout title="Alpha Release">
                Address Quality is currently in <strong>Alpha</strong> and is intended for
                evaluation and testing. The API is not yet recommended for production workloads.
              </Callout>
            </div>
          </section>

          {/* How It Works */}
          <section id="how-it-works" className="scroll-mt-20">
            <h2 className="text-2xl font-semibold tracking-tight text-surface-900">How It Works</h2>
            <p className="mt-4 leading-relaxed text-surface-600">
              Every request follows the same validation pipeline.
            </p>
            <div className="mt-6">
              <Mermaid chart={pipelineFlow} />
            </div>
            <div className="mt-6 space-y-4">
              <SectionTitle>Validation Pipeline</SectionTitle>
              <Table>
                <thead>
                  <tr className="border-b border-surface-200 bg-surface-50">
                    <Th>Step</Th>
                    <Th>Description</Th>
                  </tr>
                </thead>
                <tbody>
                  {pipelineSteps.map(({ step, description }) => (
                    <tr key={step} className="border-b border-surface-100 last:border-b-0">
                      <Td className="text-sm font-medium text-surface-900">{step}</Td>
                      <Td className="text-sm text-surface-600">{description}</Td>
                    </tr>
                  ))}
                </tbody>
              </Table>
            </div>
          </section>

          {/* Quick Start */}
          <section id="quickstart" className="scroll-mt-20">
            <h2 className="text-2xl font-semibold tracking-tight text-surface-900">Quick Start</h2>
            <p className="mt-4 leading-relaxed text-surface-600">
              Get started in less than a minute.
            </p>
            <div className="mt-6 space-y-4">
              <SectionTitle>1. Get an API Key</SectionTitle>
              <Callout title="Alpha Access">
                Self-service account registration is not available yet.
                <p className="mt-1">
                  API keys are currently issued manually during the Alpha program.
                </p>
              </Callout>
              <p className="text-sm leading-relaxed text-surface-600">
                Include the API key in every request using the
                <code className="mx-1 rounded bg-surface-100 px-1.5 py-0.5 font-mono text-xs text-surface-700">X-API-Key</code>
                header.
              </p>
              <SectionTitle>2. Validate an Address</SectionTitle>
              <CodeBlock
                code={`curl -X POST https://api.addressquality.dev/v1/validate \\\n  -H "Content-Type: application/json" \\\n  -H "X-API-Key: your-api-key" \\\n  -d '{\n    "address":"JL MERDEKA NO 56 CITARUM BANDUNG 40115"\n  }'`}
                language="bash"
                title="Terminal"
              />
            </div>
          </section>

          {/* Authentication */}
          <section id="authentication" className="scroll-mt-20">
            <h2 className="text-2xl font-semibold tracking-tight text-surface-900">Authentication</h2>
            <p className="mt-4 leading-relaxed text-surface-600">
              All API requests require authentication.
            </p>
            <div className="mt-6 space-y-4">
              <Table>
                <thead>
                  <tr className="border-b border-surface-200 bg-surface-50">
                    <Th>Header</Th>
                    <Th>Required</Th>
                  </tr>
                </thead>
                <tbody>
                  <tr className="border-b border-surface-100">
                    <Td className="font-mono text-xs text-surface-900">X-API-Key</Td>
                    <Td><Badge>Yes</Badge></Td>
                  </tr>
                </tbody>
              </Table>
              <p className="text-sm leading-relaxed text-surface-600">
                Requests without a valid API key return
                <code className="mx-1 rounded bg-surface-100 px-1.5 py-0.5 font-mono text-xs text-surface-700">401 Unauthorized</code>.
              </p>
            </div>
          </section>

          {/* POST /v1/validate */}
          <section id="validate" className="scroll-mt-20">
            <h2 className="text-2xl font-semibold tracking-tight text-surface-900">
              POST /v1/validate
            </h2>
            <p className="mt-4 leading-relaxed text-surface-600">
              Validates an Indonesian address and returns the most likely administrative hierarchy.
            </p>
            <div className="mt-6 space-y-6">
              <SectionTitle>Request</SectionTitle>
              <CodeBlock code={reqHttp} language="http" title="POST /v1/validate" />
              <SectionTitle>Request Fields</SectionTitle>
              <Table>
                <thead>
                  <tr className="border-b border-surface-200 bg-surface-50">
                    <Th>Field</Th>
                    <Th>Type</Th>
                    <Th>Required</Th>
                    <Th>Description</Th>
                  </tr>
                </thead>
                <tbody>
                  <tr className="border-b border-surface-100">
                    <Td className="font-mono text-xs font-medium text-surface-900">address</Td>
                    <Td className="font-mono text-xs text-surface-600">string</Td>
                    <Td><Badge>Yes</Badge></Td>
                    <Td className="text-sm text-surface-600">Raw Indonesian address.</Td>
                  </tr>
                  <tr>
                    <Td className="font-mono text-xs font-medium text-surface-900">source_code</Td>
                    <Td className="font-mono text-xs text-surface-600">string</Td>
                    <Td><Badge variant="default">No</Badge></Td>
                    <Td className="text-sm text-surface-600">
                      Location dataset used for validation. Default: <code className="font-mono text-xs text-surface-700">kemendagri</code>.
                    </Td>
                  </tr>
                </tbody>
              </Table>
              <SectionTitle>Example</SectionTitle>
              <CodeBlock code={reqWithSource} title="Request body" />
            </div>
          </section>

          {/* Successful Response */}
          <section id="successful-response" className="scroll-mt-20">
            <h2 className="text-2xl font-semibold tracking-tight text-surface-900">
              Successful Response
            </h2>
            <div className="mt-6">
              <CodeBlock code={successCode} title="200 OK" />
            </div>
          </section>

          {/* Understanding the Response */}
          <section id="response-schema" className="scroll-mt-20">
            <h2 className="text-2xl font-semibold tracking-tight text-surface-900">
              Understanding the Response
            </h2>
            <div className="mt-6 space-y-6">
              <div className="space-y-4">
                <SectionTitle>Status</SectionTitle>
                <Table>
                  <thead>
                    <tr className="border-b border-surface-200 bg-surface-50">
                      <Th>Status</Th>
                      <Th>Description</Th>
                    </tr>
                  </thead>
                  <tbody>
                    {[
                      ["VALID", "A single confident match was found."],
                      ["INCOMPLETE", "Some address components are missing."],
                      ["AMBIGUOUS", "Multiple valid candidates were found."],
                      ["CONFLICT", "The supplied components contradict each other."],
                      ["UNKNOWN", "No valid administrative hierarchy could be resolved."],
                    ].map(([status, description]) => (
                      <tr key={status} className="border-b border-surface-100 last:border-b-0">
                        <Td className="font-mono text-xs font-medium text-surface-900">{status}</Td>
                        <Td className="text-sm text-surface-600">{description}</Td>
                      </tr>
                    ))}
                  </tbody>
                </Table>
              </div>

              <div className="space-y-4">
                <SectionTitle>Confidence</SectionTitle>
                <p className="text-sm leading-relaxed text-surface-600">
                  The confidence score ranges from <strong>0.0</strong> to <strong>1.0</strong>.
                </p>
                <p className="text-sm leading-relaxed text-surface-600">
                  Higher scores indicate stronger evidence that the resolved location matches the
                  supplied address.
                </p>
                <p className="text-sm leading-relaxed text-surface-600">
                  The score is derived from multiple validation signals, including:
                </p>
                <ul className="list-disc space-y-1 pl-6 text-sm leading-relaxed text-surface-600">
                  <li>Administrative hierarchy consistency</li>
                  <li>Postal code validation</li>
                  <li>Candidate comparison</li>
                  <li>Validation evidence</li>
                </ul>
                <p className="text-sm leading-relaxed text-surface-600">
                  The scoring algorithm may evolve over time without changing the API response
                  format.
                </p>
              </div>

              <div className="space-y-4">
                <SectionTitle>Assessment</SectionTitle>
                <p className="text-sm leading-relaxed text-surface-600">
                  The <code className="font-mono text-xs text-surface-700">assessment</code> object
                  explains how the supplied address compares to the resolved location.
                </p>
                <Table>
                  <thead>
                    <tr className="border-b border-surface-200 bg-surface-50">
                      <Th>Field</Th>
                      <Th>Description</Th>
                    </tr>
                  </thead>
                  <tbody>
                    {[
                      ["matched", "Components successfully validated."],
                      ["missing", "Components that were not found."],
                      ["conflicts", "Components that contradict the resolved hierarchy."],
                      ["ambiguous", "Components with multiple possible matches."],
                    ].map(([field, description]) => (
                      <tr key={field} className="border-b border-surface-100 last:border-b-0">
                        <Td className="font-mono text-xs font-medium text-surface-900">{field}</Td>
                        <Td className="text-sm text-surface-600">{description}</Td>
                      </tr>
                    ))}
                  </tbody>
                </Table>
              </div>

              <div className="space-y-4">
                <SectionTitle>Resolution</SectionTitle>
                <p className="text-sm leading-relaxed text-surface-600">
                  The <code className="font-mono text-xs text-surface-700">resolution</code> object
                  explains how the final location was selected.
                </p>
                <Table>
                  <thead>
                    <tr className="border-b border-surface-200 bg-surface-50">
                      <Th>Field</Th>
                      <Th>Description</Th>
                    </tr>
                  </thead>
                  <tbody>
                    {[
                      ["strategy", "Validation strategies used during resolution."],
                      ["candidate_count", "Number of valid candidate locations found."],
                      [
                        "candidates",
                        "Ranked candidate locations with confidence scores and supporting evidence.",
                      ],
                    ].map(([field, description]) => (
                      <tr key={field} className="border-b border-surface-100 last:border-b-0">
                        <Td className="font-mono text-xs font-medium text-surface-900">{field}</Td>
                        <Td className="text-sm text-surface-600">{description}</Td>
                      </tr>
                    ))}
                  </tbody>
                </Table>
              </div>
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
                    <Th>Status</Th>
                    <Th>Description</Th>
                  </tr>
                </thead>
                <tbody>
                  <tr className="border-b border-surface-100">
                    <Td><Badge variant="warning">400</Badge></Td>
                    <Td className="text-sm text-surface-600">Invalid request body.</Td>
                  </tr>
                  <tr className="border-b border-surface-100">
                    <Td><Badge variant="danger">401</Badge></Td>
                    <Td className="text-sm text-surface-600">Missing or invalid API key.</Td>
                  </tr>
                  <tr className="border-b border-surface-100">
                    <Td><Badge variant="warning">429</Badge></Td>
                    <Td className="text-sm text-surface-600">Rate limit exceeded.</Td>
                  </tr>
                  <tr>
                    <Td><Badge variant="danger">500</Badge></Td>
                    <Td className="text-sm text-surface-600">Internal server error.</Td>
                  </tr>
                </tbody>
              </Table>
              <SectionTitle>Example</SectionTitle>
              <CodeBlock code={errorCode} title="401 Unauthorized" />
            </div>
          </section>

          {/* Rate Limits */}
          <section id="rate-limits" className="scroll-mt-20">
            <h2 className="text-2xl font-semibold tracking-tight text-surface-900">Rate Limits</h2>
            <p className="mt-4 leading-relaxed text-surface-600">
              Rate limits are enforced per API key and IP address.
            </p>
            <div className="mt-6">
              <Table>
                <thead>
                  <tr className="border-b border-surface-200 bg-surface-50">
                    <Th>Plan</Th>
                    <Th>Requests / Hour</Th>
                  </tr>
                </thead>
                <tbody>
                  <tr className="border-b border-surface-100">
                    <Td className="text-sm font-medium text-surface-900">Free</Td>
                    <Td className="text-sm text-surface-600">10</Td>
                  </tr>
                  <tr>
                    <Td className="text-sm font-medium text-surface-900">Pro</Td>
                    <Td className="text-sm text-surface-600">TBA</Td>
                  </tr>
                </tbody>
              </Table>
            </div>
          </section>

          {/* Data Source */}
          <section id="data-source" className="scroll-mt-20">
            <h2 className="text-2xl font-semibold tracking-tight text-surface-900">Data Source</h2>
            <p className="mt-4 leading-relaxed text-surface-600">
              Address Quality validates addresses using official Indonesian administrative datasets.
            </p>
            <div className="mt-6 space-y-4">
              <Table>
                <thead>
                  <tr className="border-b border-surface-200 bg-surface-50">
                    <Th>Dataset</Th>
                    <Th>Coverage</Th>
                  </tr>
                </thead>
                <tbody>
                  <tr>
                    <Td className="text-sm text-surface-900">
                      Kepmendagri No. 300.2.2-2430 Tahun 2025
                    </Td>
                    <Td className="text-sm text-surface-600">
                      Province, City/Regency, District, Subdistrict, Postal Code
                    </Td>
                  </tr>
                </tbody>
              </Table>
              <p className="text-sm leading-relaxed text-surface-600">Reference dataset:</p>
              <ul className="list-disc space-y-1 pl-6 text-sm leading-relaxed text-surface-600">
                <li>
                  <a
                    href="https://github.com/cahyadsn/wilayah_ref"
                    target="_blank"
                    rel="noreferrer"
                    className="text-accent-700 underline underline-offset-2 hover:text-accent-900"
                  >
                    github.com/cahyadsn/wilayah_ref
                  </a>
                </li>
              </ul>
              <p className="text-sm leading-relaxed text-surface-600">
                Datasets are updated periodically as new administrative data becomes available.
              </p>
            </div>
          </section>

          {/* FAQ */}
          <section id="faq" className="scroll-mt-20">
            <h2 className="text-2xl font-semibold tracking-tight text-surface-900">FAQ</h2>
            <div className="mt-6 space-y-6">
              <div>
                <SectionTitle>What address formats are supported?</SectionTitle>
                <p className="mt-2 text-sm leading-relaxed text-surface-600">
                  Most Indonesian address formats are supported, including multiline addresses,
                  abbreviations, postal codes, and mixed administrative components.
                </p>
              </div>
              <div>
                <SectionTitle>Does the API return multiple candidates?</SectionTitle>
                <p className="mt-2 text-sm leading-relaxed text-surface-600">
                  Yes. When an address cannot be resolved uniquely, the response contains ranked
                  candidate locations together with confidence scores and supporting evidence.
                </p>
              </div>
              <div>
                <SectionTitle>How is confidence calculated?</SectionTitle>
                <p className="mt-2 text-sm leading-relaxed text-surface-600">
                  Confidence is derived from multiple validation signals rather than a simple
                  component count. The scoring model may evolve as additional validation evidence is
                  introduced.
                </p>
              </div>
              <div>
                <SectionTitle>Is the API deterministic?</SectionTitle>
                <p className="mt-2 text-sm leading-relaxed text-surface-600">
                  Yes. Given the same input and dataset version, the API returns consistent results.
                </p>
              </div>
              <div>
                <SectionTitle>Can I use the API in production?</SectionTitle>
                <p className="mt-2 text-sm leading-relaxed text-surface-600">
                  Not yet. The current Alpha release is intended for evaluation and integration
                  testing. While the API already supports authentication, rate limiting, automated
                  testing, and periodic evaluation, it is{" "}
                  <strong>not yet recommended for production workloads</strong>.
                </p>
              </div>
            </div>
          </section>
        </div>
      </div>
    </Container>
  )
}
