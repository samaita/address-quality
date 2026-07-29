import type { ReactNode } from "react"

type CardProps = {
  children: ReactNode
  className?: string
  hover?: boolean
}

export default function Card({ children, className = "", hover = false }: CardProps) {
  return (
    <div
      className={`rounded-2xl border border-surface-200 bg-white p-6 shadow-card ${
        hover ? "transition-shadow duration-200 hover:shadow-elevated" : ""
      } ${className}`}
    >
      {children}
    </div>
  )
}
