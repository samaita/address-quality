import Section from "@/components/kumo-ui/Section"
import FeatureCard from "@/components/landing/FeatureCard"

const features = [
  {
    icon: "🔍",
    title: "Address Parsing",
    description:
      "Tokenize raw Indonesian addresses into structured components — road names, administrative areas, and postal codes.",
  },
  {
    icon: "🗺️",
    title: "Administrative Validation",
    description:
      "Cross-reference parsed components against the official Kemendagri database of provinces, cities, districts, and subdistricts.",
  },
  {
    icon: "📊",
    title: "Confidence Scoring",
    description:
      "Each result includes a 0–100% confidence score based on the precision and consistency of matched evidence.",
  },
  {
    icon: "⚠️",
    title: "Ambiguity Detection",
    description:
      "Identifies conflicting or overlapping address components and surfaces alternative candidate locations ranked by score.",
  },
  {
    icon: "✅",
    title: "Explainable Results",
    description:
      "Every match includes evidence — what was matched, what was missing, and any conflicts found during resolution.",
  },
]

export default function Features() {
  return (
    <Section
      title="Everything you need for address validation"
      description="From raw text to structured, verified data — the API handles the complexity of Indonesian address formats."
    >
      <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
        {features.map((feature) => (
          <FeatureCard key={feature.title} {...feature} />
        ))}
      </div>
    </Section>
  )
}
