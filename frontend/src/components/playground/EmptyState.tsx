import { Empty } from "@cloudflare/kumo/components/empty"
import { MapPinIcon } from "@/components/icons"

export default function EmptyState() {
  return (
    <Empty
      size="lg"
      icon={<MapPinIcon className="h-10 w-10 text-surface-300" />}
      title="Enter an address to validate"
      description="Paste an Indonesian address on the left and click Validate Address to see parsed fields, confidence score, and candidate matches."
    />
  )
}
