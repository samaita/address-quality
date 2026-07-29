import type { ReactNode } from "react"
import Card from "@/components/common/Card"

type FeatureCardProps = {
  icon: ReactNode
  title: string
  description: string
}

export default function FeatureCard({ icon, title, description }: FeatureCardProps) {
  return (
    <Card hover className="flex flex-col gap-3">
      <span className="text-accent-600">{icon}</span>
      <h3 className="text-base font-semibold text-surface-900">{title}</h3>
      <p className="text-sm leading-relaxed text-surface-500">{description}</p>
    </Card>
  )
}
