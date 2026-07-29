import { Link } from "react-router-dom"
import Container from "@/components/layout/Container"
import Badge from "@/components/common/Badge"
import CodeBlock from "@/components/docs/CodeBlock"
import { ChevronRightIcon } from "@/components/icons"

const heroCurlSnippet = `curl -X POST https://app.samaita.com/address-quality/v1/validate \\
  -H "Content-Type: application/json" \\
  -H "X-API-Key: \\$AQ_API_KEY" \\
  -d '{"address":"JL MERDEKA NO 56 CITARUM BANDUNG 40115"}'`

export default function Hero() {
  return (
    <section className="py-14 lg:py-28">
      <Container>
        <div className="grid gap-12 lg:grid-cols-2 lg:items-center">
          <div>
            <Badge variant="info" className="mb-4">
              Indonesian Address Validation API · Public Alpha
            </Badge>
            <h1 className="text-3xl font-semibold tracking-tight text-surface-900 sm:text-5xl lg:text-6xl">
              Turn raw Indonesian addresses into structured, verified data.
            </h1>
            <p className="mt-6 text-lg leading-relaxed text-surface-500 sm:text-xl">
              Parse any free-text Indonesian address. Match it against the official
              Kemendagri hierarchy. Get a structured result with a confidence score and
              the evidence behind it.
            </p>
            <div className="mt-10 flex flex-col items-start gap-4 sm:flex-row">
              <Link
                to="/playground"
                className="inline-flex items-center justify-center gap-2 rounded-lg bg-accent-600 px-6 py-3 text-base font-medium text-white shadow-sm transition-all duration-150 hover:bg-accent-700 focus:outline-none focus-visible:ring-2 focus-visible:ring-accent-500 focus-visible:ring-offset-2"
              >
                Try Playground
                <ChevronRightIcon className="h-4 w-4" />
              </Link>
              <Link
                to="/docs"
                className="inline-flex items-center justify-center gap-2 rounded-lg border border-surface-200 bg-white px-6 py-3 text-base font-medium text-surface-700 shadow-sm transition-all duration-150 hover:bg-surface-50 hover:text-surface-900 focus:outline-none focus-visible:ring-2 focus-visible:ring-accent-500 focus-visible:ring-offset-2"
              >
                Read the Docs
              </Link>
            </div>
            <p className="mt-4 text-sm text-surface-400">
              Public Alpha.{" "}
              <a
                href="mailto:garysamaita@gmail.com?subject=Address%20Quality%20API%20Key%20Request"
                className="underline transition-colors hover:text-surface-600"
              >
                Request an API key
              </a>
            </p>
          </div>
          <div>
            <CodeBlock
              code={heroCurlSnippet}
              language="bash"
              title="Your first request"
            />
          </div>
        </div>
      </Container>
    </section>
  )
}
