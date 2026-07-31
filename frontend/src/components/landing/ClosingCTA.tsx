import Container from "@/components/layout/Container"
import { Badge } from "@cloudflare/kumo/components/badge"
import { LinkButton } from "@cloudflare/kumo/components/button"
import { ChevronRightIcon } from "@/components/icons"

export default function ClosingCTA() {
  return (
    <section data-mode="dark" className="bg-surface-900 py-20">
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
            <LinkButton href="/playground" variant="primary" size="lg">
              Try Playground
              <ChevronRightIcon className="h-4 w-4" />
            </LinkButton>
            <LinkButton
              href="mailto:garysamaita@gmail.com?subject=Address%20Quality%20API%20Key%20Request"
              variant="secondary"
              size="lg"
            >
              Request an API key
            </LinkButton>
          </div>
          <p className="mt-6 text-sm text-surface-500">
            BSL 1.1 · Converts to Apache 2.0 on 2030-03-01
          </p>
        </div>
      </Container>
    </section>
  )
}
