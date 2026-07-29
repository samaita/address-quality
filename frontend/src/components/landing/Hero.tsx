import Container from "@/components/layout/Container"
import Button from "@/components/common/Button"
import { ArrowTopRightOnSquareIcon } from "@/components/icons"

export default function Hero() {
  return (
    <section className="py-20 lg:py-32">
      <Container>
        <div className="mx-auto max-w-3xl text-center">
          <h1 className="text-4xl font-semibold tracking-tight text-surface-900 sm:text-5xl lg:text-6xl">
            Validate Indonesian Addresses with Confidence
          </h1>
          <p className="mt-6 text-lg leading-relaxed text-surface-500 sm:text-xl">
            Address Quality API parses, validates, and resolves Indonesian addresses
            against the official administrative hierarchy. Get structured data,
            confidence scores, and explainable results from raw text input.
          </p>
          <div className="mt-10 flex flex-col items-center justify-center gap-4 sm:flex-row">
            <Button as="a" href="/playground" size="lg">
              Try Playground
              <ChevronRightIcon className="h-4 w-4" />
            </Button>
            <Button as="a" href="/docs" variant="secondary" size="lg">
              API Docs
              <ArrowTopRightOnSquareIcon className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </Container>
    </section>
  )
}

function ChevronRightIcon({ className }: { className?: string }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 24 24"
      strokeWidth={2}
      stroke="currentColor"
      className={className}
    >
      <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5" />
    </svg>
  )
}
