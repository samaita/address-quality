import type { HTMLAttributes } from "react"

type TableProps = HTMLAttributes<HTMLTableElement>

export function Table({ className = "", children, ...props }: TableProps) {
  return (
    <div className="overflow-x-auto rounded-xl border border-surface-200">
      <table className={`w-full text-left text-sm ${className}`} {...props}>
        {children}
      </table>
    </div>
  )
}

type THeadProps = HTMLAttributes<HTMLTableSectionElement>

export function THead({ className = "", children, ...props }: THeadProps) {
  return (
    <thead className={`bg-surface-50 ${className}`} {...props}>
      {children}
    </thead>
  )
}

type TBodyProps = HTMLAttributes<HTMLTableSectionElement>

export function TBody({ className = "", children, ...props }: TBodyProps) {
  return (
    <tbody className={className} {...props}>
      {children}
    </tbody>
  )
}

type TrProps = HTMLAttributes<HTMLTableRowElement>

export function Tr({ className = "", children, ...props }: TrProps) {
  return (
    <tr className={`border-b border-surface-100 last:border-0 ${className}`} {...props}>
      {children}
    </tr>
  )
}

type ThProps = HTMLAttributes<HTMLTableCellElement>

export function Th({ className = "", children, ...props }: ThProps) {
  return (
    <th
      className={`px-4 py-2.5 text-xs font-medium uppercase tracking-wider text-surface-500 ${className}`}
      {...props}
    >
      {children}
    </th>
  )
}

type TdProps = HTMLAttributes<HTMLTableCellElement>

export function Td({ className = "", children, ...props }: TdProps) {
  return (
    <td className={`px-4 py-2.5 text-sm text-surface-700 ${className}`} {...props}>
      {children}
    </td>
  )
}
