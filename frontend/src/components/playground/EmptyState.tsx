import Card from "@/components/common/Card"
import { MapPinIcon } from "@/components/icons"

export default function EmptyState() {
  return (
    <Card className="flex flex-col items-center justify-center py-20 text-center">
      <div className="flex h-12 w-12 items-center justify-center rounded-full bg-surface-100">
        <MapPinIcon className="h-6 w-6 text-surface-400" />
      </div>
      <p className="mt-4 text-base font-medium text-surface-900">
        Enter an address and click <span className="font-semibold text-accent-600">Validate Address</span>.
      </p>
      <p className="mt-2 max-w-sm text-sm text-surface-500">
        Paste an Indonesian address on the left to see parsed fields, confidence
        score, and candidate matches.
      </p>
    </Card>
  )
}
