import type { ReactNode } from "react"

type ContainerProps = {
  children: ReactNode
  className?: string
}

export default function Container({ children, className = "" }: ContainerProps) {
  return (
    <div className={`mx-auto max-w-page px-4 sm:px-6 lg:px-8 ${className}`}>
      {children}
    </div>
  )
}
