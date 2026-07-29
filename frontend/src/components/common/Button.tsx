import { forwardRef, type ButtonHTMLAttributes, type AnchorHTMLAttributes } from "react"

type ButtonVariant = "primary" | "secondary" | "ghost"
type ButtonSize = "sm" | "md" | "lg"

type ButtonBaseProps = {
  variant?: ButtonVariant
  size?: ButtonSize
  className?: string
}

type ButtonAsButton = ButtonBaseProps &
  Omit<ButtonHTMLAttributes<HTMLButtonElement>, "className"> & {
    as?: "button"
    href?: never
  }

type ButtonAsLink = ButtonBaseProps &
  Omit<AnchorHTMLAttributes<HTMLAnchorElement>, "className"> & {
    as: "a"
    href: string
  }

type ButtonProps = ButtonAsButton | ButtonAsLink

const variantStyles: Record<ButtonVariant, string> = {
  primary:
    "bg-accent-600 text-white hover:bg-accent-700 focus-visible:ring-accent-500 shadow-sm",
  secondary:
    "bg-white text-surface-700 border border-surface-200 hover:bg-surface-50 hover:text-surface-900 focus-visible:ring-accent-500",
  ghost:
    "text-surface-600 hover:text-surface-900 hover:bg-surface-100 focus-visible:ring-accent-500",
}

const sizeStyles: Record<ButtonSize, string> = {
  sm: "px-3 py-1.5 text-sm",
  md: "px-4 py-2 text-sm",
  lg: "px-6 py-2.5 text-base",
}

const Button = forwardRef<HTMLButtonElement | HTMLAnchorElement, ButtonProps>(
  (props, ref) => {
    const {
      variant = "primary",
      size = "md",
      className = "",
      children,
    } = props

    const baseStyles =
      "inline-flex items-center justify-center gap-2 rounded-lg font-medium transition-all duration-150 focus:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50"

    const classes = `${baseStyles} ${variantStyles[variant]} ${sizeStyles[size]} ${className}`

    if (props.as === "a") {
      const { as: _, variant: _v, size: _s, ...anchorProps } = props as ButtonAsLink
      return (
        <a
          ref={ref as React.Ref<HTMLAnchorElement>}
          className={classes}
          {...anchorProps}
        >
          {children}
        </a>
      )
    }

    const { as: _, variant: _v, size: _s, ...buttonProps } = props as ButtonAsButton
    return (
      <button
        ref={ref as React.Ref<HTMLButtonElement>}
        className={classes}
        {...buttonProps}
      >
        {children}
      </button>
    )
  },
)

Button.displayName = "Button"

export default Button
