import { Link } from "react-router-dom"
import Container from "@/components/layout/Container"
import Badge from "@/components/common/Badge"
import { ChevronRightIcon } from "@/components/icons"

export default function ClosingCTA() {
  return (
    <section className="bg-surface-900 py-20">
      <Container>
        <div className="mx-auto max-w-2xl text-center">
          <Badge variant="info" className="mb-4">
            Public Alpha
          </Badge>
          <h2 className="text-3xl font-semibold tracking-tight text-white sm:text-4xl">
            Ready to try it?
          </h2>
          <p className="mt-4 text-lg leading-relaxed text-surface-400">
            Run an address through the Playground, or request an API key to start
            integrating in minutes.
          </p>
          <div className="mt-10 flex flex-col items-center justify-center gap-4 sm:flex-row">
            <Link
              to="/playground"
              className="inline-flex items-center justify-center gap-2 rounded-lg bg-white px-6 py-3 text-base font-medium text-surface-900 shadow-sm transition-all duration-150 hover:bg-surface-100 focus:outline-none focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-offset-2 focus-visible:ring-offset-surface-900"
            >
              Try Playground
              <ChevronRightIcon className="h-4 w-4" />
            </Link>
            <a
              href="mailto:garysamaita@gmail.com?subject=Address%20Quality%20API%20Key%20Request"
              className="inline-flex items-center justify-center gap-2 rounded-lg border border-surface-600 px-6 py-3 text-base font-medium text-surface-200 transition-all duration-150 hover:border-surface-400 hover:text-white focus:outline-none focus-visible:ring-2 focus-visible:ring-white focus-visible:ring-offset-2 focus-visible:ring-offset-surface-900"
            >
              Request an API key
            </a>
          </div>
          <p className="mt-6 text-sm text-surface-500">
            BSL 1.1 · Converts to Apache 2.0 on 2030-03-01
          </p>
        </div>
      </Container>
    </section>
  )
}
