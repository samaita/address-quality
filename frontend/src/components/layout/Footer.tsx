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
          <div className="flex flex-col gap-3 sm:items-end">
            <div className="flex flex-wrap gap-x-6 gap-y-2">
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
              <a
                href="https://github.com/samaita/address-quality"
                target="_blank"
                rel="noopener noreferrer"
                className="text-sm text-surface-500 transition-colors hover:text-surface-900"
              >
                GitHub
              </a>
            </div>
          </div>
        </div>
        <div className="mt-10 border-t border-surface-100 pt-6">
          <p className="text-xs text-surface-400">
            &copy; {new Date().getFullYear()} Samaita. All rights reserved.
          </p>
        </div>
      </Container>
    </footer>
  )
}
