import Section from "@/components/kumo-ui/Section"

const limitations = [
  {
    title: "Road-level data",
    description:
      "Kemendagri covers subdistrict-level only. `road_name` in the response is best-effort from the input text and is not verified against a street database.",
  },
  {
    title: "Geocoding",
    description:
      "The API returns administrative hierarchy, not geographic coordinates. For lat/long, combine with a geocoding service.",
  },
  {
    title: "Rate limit",
    description:
      "10 requests per hour per key during Public Alpha. Request a dedicated limit for production use.",
  },
  {
    title: "License",
    description:
      "Business Source License 1.1. Source-available, not open source. Converts to Apache 2.0 on 2030-03-01.",
  },
  {
    title: "Input language",
    description:
      "Indonesian and English transliterations supported. Mixed scripts are handled on a best-effort basis.",
  },
  {
    title: "Alpha scope",
    description:
      "Single endpoint only. Batch processing and async validation are planned for post-Alpha.",
  },
]

export default function Limitations() {
  return (
    <Section
      title="Current limitations"
      description="Public Alpha means the API is under active development. Here's what you should know before trying it."
    >
      <div className="mb-4 flex items-center gap-2">
        <span className="text-xs font-semibold uppercase tracking-wider text-accent-600">
          Limitations
        </span>
      </div>
      <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
        {limitations.map((item) => (
          <div
            key={item.title}
            className="rounded-xl border border-amber-200 bg-amber-50 px-5 py-4"
          >
            <h3 className="text-base font-semibold text-amber-800">{item.title}</h3>
            <p className="mt-2 text-sm leading-relaxed text-amber-700">
              {item.description}
            </p>
          </div>
        ))}
      </div>
    </Section>
  )
}
