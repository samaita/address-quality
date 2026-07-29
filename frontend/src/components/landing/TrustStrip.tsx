import Container from "@/components/layout/Container"

const items = [
  { label: "Data source", value: "Kemendagri 2025" },
  { label: "Version", value: "v1 · Public Alpha" },
  { label: "Rate limit", value: "10 req / hour" },
  { label: "License", value: "BSL 1.1 · Source-available" },
]

export default function TrustStrip() {
  return (
    <div className="border-y border-surface-200 bg-surface-50">
      <Container>
        <div className="grid grid-cols-2 gap-4 py-6 sm:grid-cols-4">
          {items.map((item) => (
            <div key={item.label} className="text-center sm:text-left">
              <p className="text-xs font-semibold uppercase tracking-wider text-surface-400">
                {item.label}
              </p>
              <p className="mt-1 text-sm font-medium text-surface-700">
                {item.value}
              </p>
            </div>
          ))}
        </div>
      </Container>
    </div>
  )
}
