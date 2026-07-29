import type { ReactNode } from "react"
import Container from "@/components/layout/Container"

type SectionProps = {
  title?: string
  description?: string
  children: ReactNode
  className?: string
}

export default function Section({ title, description, children, className = "" }: SectionProps) {
  return (
    <section className={`py-16 lg:py-24 ${className}`}>
      <Container>
        {title && (
          <div className="mb-12 max-w-2xl">
            <h2 className="text-3xl font-semibold tracking-tight text-surface-900 sm:text-4xl">
              {title}
            </h2>
            {description && (
              <p className="mt-4 text-lg text-surface-500 leading-relaxed">
                {description}
              </p>
            )}
          </div>
        )}
        {children}
      </Container>
    </section>
  )
}
