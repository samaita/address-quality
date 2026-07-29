import { useState } from "react"
import { Link, useLocation } from "react-router-dom"
import { Bars3Icon, XMarkIcon, ArrowTopRightOnSquareIcon } from "@/components/icons"
import Container from "@/components/layout/Container"

const navLinks = [
  { label: "Home", to: "/" },
  { label: "Playground", to: "/playground" },
  { label: "Docs", to: "/docs" },
]

export default function Navbar() {
  const [mobileOpen, setMobileOpen] = useState(false)
  const { pathname } = useLocation()

  return (
    <header className="sticky top-0 z-50 border-b border-surface-200 bg-white/95 backdrop-blur supports-[backdrop-filter]:bg-white/80">
      <Container>
        <nav className="flex h-16 items-center justify-between">
          <Link
            to="/"
            className="text-lg font-semibold tracking-tight text-surface-900"
          >
            Address Quality
          </Link>

          <div className="hidden items-center gap-1 md:flex">
            {navLinks.map((link) => {
              const isActive = pathname === link.to
              return (
                <Link
                  key={link.to}
                  to={link.to}
                  className={`rounded-lg px-3 py-2 text-sm font-medium transition-colors ${
                    isActive
                      ? "text-surface-900"
                      : "text-surface-500 hover:text-surface-900 hover:bg-surface-100"
                  }`}
                >
                  {link.label}
                </Link>
              )
            })}
            <a
              href="https://github.com/samaita/address-quality"
              target="_blank"
              rel="noopener noreferrer"
              className="ml-2 inline-flex items-center gap-1 rounded-lg px-3 py-2 text-sm font-medium text-surface-500 transition-colors hover:text-surface-900 hover:bg-surface-100"
            >
              GitHub
              <ArrowTopRightOnSquareIcon className="h-3.5 w-3.5" />
            </a>
          </div>

          <button
            type="button"
            onClick={() => setMobileOpen(!mobileOpen)}
            className="inline-flex items-center justify-center rounded-lg p-2 text-surface-500 hover:text-surface-900 hover:bg-surface-100 md:hidden"
            aria-label={mobileOpen ? "Close menu" : "Open menu"}
          >
            {mobileOpen ? (
              <XMarkIcon className="h-5 w-5" />
            ) : (
              <Bars3Icon className="h-5 w-5" />
            )}
          </button>
        </nav>

        {mobileOpen && (
          <div className="border-t border-surface-200 pb-4 pt-2 md:hidden">
            {navLinks.map((link) => {
              const isActive = pathname === link.to
              return (
                <Link
                  key={link.to}
                  to={link.to}
                  onClick={() => setMobileOpen(false)}
                  className={`block rounded-lg px-3 py-2 text-sm font-medium transition-colors ${
                    isActive
                      ? "text-surface-900"
                      : "text-surface-500 hover:text-surface-900 hover:bg-surface-100"
                  }`}
                >
                  {link.label}
                </Link>
              )
            })}
            <a
              href="https://github.com/samaita/address-quality"
              target="_blank"
              rel="noopener noreferrer"
              className="mt-1 flex items-center gap-1 rounded-lg px-3 py-2 text-sm font-medium text-surface-500 transition-colors hover:text-surface-900 hover:bg-surface-100"
            >
              GitHub
              <ArrowTopRightOnSquareIcon className="h-3.5 w-3.5" />
            </a>
          </div>
        )}
      </Container>
    </header>
  )
}
