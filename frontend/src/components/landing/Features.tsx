import Section from "@/components/kumo-ui/Section"
import FeatureCard from "@/components/landing/FeatureCard"
import {
  MagnifyingGlassIcon,
  MapPinIcon,
  SparklesIcon,
  ExclamationTriangleIcon,
  ShieldCheckIcon,
  LinkIcon,
} from "@/components/icons"

const features = [
  {
    icon: <MagnifyingGlassIcon className="h-6 w-6" />,
    title: "Address Parsing",
    description:
      "Split a raw address into road, province, city, district, subdistrict, and postal code.",
  },
  {
    icon: <MapPinIcon className="h-6 w-6" />,
    title: "Administrative Validation",
    description:
      "Verify each component against Indonesia's official Kemendagri administrative database (Ministry of Home Affairs, 2025).",
  },
  {
    icon: <SparklesIcon className="h-6 w-6" />,
    title: "Confidence Scoring",
    description:
      "Every result returns a 0–100% confidence score, computed from match precision and evidence consistency.",
  },
  {
    icon: <ExclamationTriangleIcon className="h-6 w-6" />,
    title: "Ambiguity Detection",
    description:
      "Detects conflicting components. Returns ranked alternative candidates when a match is ambiguous.",
  },
  {
    icon: <ShieldCheckIcon className="h-6 w-6" />,
    title: "Explainable Results",
    description:
      "Every match ships with evidence: what matched, what's missing, and any conflicts found.",
  },
  {
    icon: <LinkIcon className="h-6 w-6" />,
    title: "Kemendagri 2025 Coverage",
    description:
      "All 38 provinces, 514 cities and regencies with districts, subdistricts, and postal codes from the 2025 Kemendagri release.",
  },
]

export default function Features() {
  return (
    <Section
      title="What the API does"
      description="One endpoint handles parsing, matching, scoring, and resolution for Indonesian address formats."
    >
      <div className="mb-4 flex items-center gap-2">
        <span className="text-xs font-semibold uppercase tracking-wider text-accent-600">
          Features
        </span>
      </div>
      <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
        {features.map((feature) => (
          <FeatureCard key={feature.title} {...feature} />
        ))}
      </div>
    </Section>
  )
}
