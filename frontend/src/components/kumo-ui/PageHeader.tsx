type PageHeaderProps = {
  title: string
  description?: string
}

export default function PageHeader({ title, description }: PageHeaderProps) {
  return (
    <div className="py-16 lg:py-24">
      <h1 className="text-4xl font-semibold tracking-tight text-surface-900 sm:text-5xl">
        {title}
      </h1>
      {description && (
        <p className="mt-4 max-w-2xl text-lg text-surface-500 leading-relaxed">
          {description}
        </p>
      )}
    </div>
  )
}
