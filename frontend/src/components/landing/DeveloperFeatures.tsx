import Section from "@/components/kumo-ui/Section"
import Badge from "@/components/common/Badge"

const rows = [
  { label: "Endpoint", value: "POST /v1/validate" },
  { label: "Base URL", value: "https://app.samaita.com/address-quality" },
  { label: "Auth", value: "X-API-Key header" },
  { label: "Rate limit", value: "10 req / hour per key" },
  { label: "Content type", value: "application/json" },
  { label: "OpenAPI", value: "/swagger.yaml", href: "/swagger.yaml", download: true },
  { label: "License", value: "Business Source License 1.1", extra: "Converts to Apache 2.0 on 2030-03-01" },
]

export default function DeveloperFeatures() {
  return (
    <Section
      title="Developer details"
      description="Everything you need to integrate the API into your stack."
    >
      <div className="mb-4 flex items-center gap-2">
        <span className="text-xs font-semibold uppercase tracking-wider text-accent-600">
          Integration
        </span>
      </div>
      <div className="overflow-hidden rounded-xl border border-surface-200">
        {rows.map((row, i) => (
          <div
            key={row.label}
            className={`flex flex-col gap-1 px-5 py-4 sm:flex-row sm:items-center sm:gap-6 ${
              i !== 0 ? "border-t border-surface-100" : ""
            }`}
          >
            <span className="min-w-[120px] text-sm font-medium text-surface-500">
              {row.label}
            </span>
            <div className="min-w-0 flex-1">
              {row.href ? (
                <a
                  href={row.href}
                  download={row.download}
                  className="font-mono text-sm text-surface-900 underline transition-colors hover:text-accent-600"
                >
                  {row.value}
                </a>
              ) : row.label === "Auth" ? (
                <div className="flex items-center gap-3">
                  <code className="rounded bg-surface-100 px-1.5 py-0.5 font-mono text-sm text-surface-800">
                    X-API-Key
                  </code>
                  <a
                    href="mailto:garysamaita@gmail.com?subject=Address%20Quality%20API%20Key%20Request"
                    className="text-xs text-surface-400 underline transition-colors hover:text-surface-600"
                  >
                    Request a key
                  </a>
                </div>
              ) : row.label === "Rate limit" ? (
                <div className="flex items-center gap-3">
                  <span className="font-mono text-sm text-surface-900">
                    10 req / hour
                  </span>
                  <a
                    href="mailto:garysamaita@gmail.com?subject=Address%20Quality%20-%20Dedicated%20Rate%20Limit"
                    className="text-xs text-surface-400 underline transition-colors hover:text-surface-600"
                  >
                    Request dedicated limit
                  </a>
                </div>
              ) : (
                <span className="font-mono text-sm text-surface-900">
                  {row.value}
                </span>
              )}
              {row.extra && (
                <p className="mt-1 text-xs text-surface-400">{row.extra}</p>
              )}
            </div>
          </div>
        ))}
      </div>
      <div className="mt-6 flex flex-wrap items-center gap-3 text-xs text-surface-400">
        <Badge variant="danger">400</Badge>
        <span>Bad Request</span>
        <span className="text-surface-300">·</span>
        <Badge variant="danger">401</Badge>
        <span>Unauthorized</span>
        <span className="text-surface-300">·</span>
        <Badge variant="danger">404</Badge>
        <span>Not Found</span>
        <span className="text-surface-300">·</span>
        <Badge variant="warning">429</Badge>
        <span>Rate Limited</span>
      </div>
    </Section>
  )
}
