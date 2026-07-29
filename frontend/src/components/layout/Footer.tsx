import { Link } from "react-router-dom"
import Container from "@/components/layout/Container"

export default function Footer() {
  return (
    <footer className="border-t border-surface-200">
      <Container className="py-12">
        <div className="flex flex-col items-start justify-between gap-8 sm:flex-row">
          <div>
            <Link
              to="/"
              className="text-base font-semibold tracking-tight text-surface-900"
            >
              Address Quality
            </Link>
            <p className="mt-1 text-sm text-surface-500">
              Indonesian address validation API.
            </p>
          </div>
          <div className="flex flex-row flex-wrap gap-x-8 gap-y-3 sm:items-end">
            <div className="flex flex-col gap-2">
              <span className="text-xs font-semibold uppercase tracking-wider text-surface-400">
                Product
              </span>
              <Link
                to="/playground"
                className="text-sm text-surface-500 transition-colors hover:text-surface-900"
              >
                Playground
              </Link>
              <Link
                to="/docs"
                className="text-sm text-surface-500 transition-colors hover:text-surface-900"
              >
                Docs
              </Link>
            </div>
            <div className="flex flex-col gap-2">
              <span className="text-xs font-semibold uppercase tracking-wider text-surface-400">
                Resources
              </span>
              <a
                href="https://github.com/samaita/address-quality"
                target="_blank"
                rel="noopener noreferrer"
                className="text-sm text-surface-500 transition-colors hover:text-surface-900"
              >
                GitHub
              </a>
              <a
                href="/swagger.yaml"
                download
                className="text-sm text-surface-500 transition-colors hover:text-surface-900"
              >
                OpenAPI
              </a>
            </div>
            <div className="flex flex-col gap-2">
              <span className="text-xs font-semibold uppercase tracking-wider text-surface-400">
                Legal
              </span>
              <span className="text-sm text-surface-500">
                BSL 1.1
              </span>
              <span className="text-xs text-surface-400">
                Converts to Apache 2.0 on 2030-03-01
              </span>
            </div>
          </div>
        </div>
        <div className="mt-10 border-t border-surface-100 pt-6">
          <p className="text-xs text-surface-500">
            &copy; {new Date().getFullYear()} Samaita. All rights reserved.
          </p>
        </div>
      </Container>
    </footer>
  )
}
