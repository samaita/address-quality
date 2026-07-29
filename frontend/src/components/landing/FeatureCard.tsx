import Card from "@/components/common/Card"

type FeatureCardProps = {
  icon: string
  title: string
  description: string
}

export default function FeatureCard({ icon, title, description }: FeatureCardProps) {
  return (
    <Card hover className="flex flex-col gap-3">
      <span className="text-2xl" role="img" aria-hidden="true">
        {icon}
      </span>
      <h3 className="text-base font-semibold text-surface-900">{title}</h3>
      <p className="text-sm leading-relaxed text-surface-500">{description}</p>
    </Card>
  )
}
