import type { ReactNode } from "react"
import Section from "@/components/kumo-ui/Section"
import {
  MagnifyingGlassIcon,
  ArrowsRightLeftIcon,
  SparklesIcon,
  ShieldCheckIcon,
} from "@/components/icons"

const steps: { number: number; icon: ReactNode; title: string; description: string }[] = [
  {
    number: 1,
    icon: <MagnifyingGlassIcon className="h-5 w-5" />,
    title: "Parse",
    description:
      "Split raw text into road, province, city, district, subdistrict, and postal code.",
  },
  {
    number: 2,
    icon: <ArrowsRightLeftIcon className="h-5 w-5" />,
    title: "Match",
    description:
      "Verify each component against the Kemendagri 2025 administrative hierarchy.",
  },
  {
    number: 3,
    icon: <SparklesIcon className="h-5 w-5" />,
    title: "Score",
    description:
      "Compute a 0–100% confidence from match precision and evidence consistency.",
  },
  {
    number: 4,
    icon: <ShieldCheckIcon className="h-5 w-5" />,
    title: "Resolve",
    description:
      "Return ranked candidates with evidence for each match and any conflicts found.",
  },
]

export default function HowItWorks() {
  return (
    <Section
      title="From raw text to verified result"
      description="A single API call runs the full pipeline — no multiple endpoints, no state to manage."
    >
      <div className="mb-4 flex items-center gap-2">
        <span className="text-xs font-semibold uppercase tracking-wider text-accent-600">
          How it works
        </span>
      </div>
      <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
        {steps.map((step) => (
          <div key={step.number} className="flex flex-col gap-3 rounded-xl border border-surface-200 bg-white p-6">
            <div className="flex items-center gap-3">
              <span className="flex h-8 w-8 items-center justify-center rounded-full bg-accent-100 text-sm font-semibold text-accent-700">
                {step.number}
              </span>
              <span className="text-surface-600">{step.icon}</span>
            </div>
            <h3 className="text-base font-semibold text-surface-900">{step.title}</h3>
            <p className="text-sm leading-relaxed text-surface-500">{step.description}</p>
          </div>
        ))}
      </div>
    </Section>
  )
}
