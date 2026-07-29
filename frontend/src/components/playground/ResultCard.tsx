import Card from "@/components/common/Card"
import type { Location } from "@/types/api"

type ResultCardProps = {
  location: Location
  formattedAddress: string
  normalizedInput: string
}

const fields: { label: string; key: keyof Location }[] = [
  { label: "Province", key: "province" },
  { label: "City", key: "city" },
  { label: "District", key: "district" },
  { label: "Subdistrict", key: "sub_district" },
  { label: "Postal Code", key: "postal_code" },
]

export default function ResultCard({
  location,
  formattedAddress,
  normalizedInput,
}: ResultCardProps) {
  return (
    <Card>
      <div className="mb-4">
        <p className="text-sm font-medium text-surface-900">{formattedAddress}</p>
        <p className="mt-0.5 text-xs text-surface-400">{normalizedInput}</p>
      </div>
      <dl className="divide-y divide-surface-100">
        {fields.map(({ label, key }) => (
          <div key={key} className="flex items-baseline justify-between py-2.5">
            <dt className="text-sm text-surface-500">{label}</dt>
            <dd className="text-sm font-medium text-surface-900">{location[key]}</dd>
          </div>
        ))}
      </dl>
    </Card>
  )
}
