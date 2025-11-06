import * as React from "react"
import { motion } from "framer-motion"
import { cn } from "../../lib/utils"

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "default" | "destructive" | "outline" | "secondary" | "ghost" | "link"
  size?: "default" | "sm" | "lg" | "icon"
  animated?: boolean
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = "default", size = "default", animated = true, ...props }, ref) => {
    const baseClassName = cn(
      "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-lg text-sm font-medium transition-colors",
      "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
      "disabled:pointer-events-none disabled:opacity-50",
      {
        "bg-primary text-primary-foreground hover:bg-primary/90 shadow-md": variant === "default",
        "bg-destructive text-destructive-foreground hover:bg-destructive/90": variant === "destructive",
        "border border-input bg-background hover:bg-accent hover:text-accent-foreground": variant === "outline",
        "bg-secondary text-secondary-foreground hover:bg-secondary/80": variant === "secondary",
        "hover:bg-accent hover:text-accent-foreground": variant === "ghost",
        "text-primary underline-offset-4 hover:underline": variant === "link",
      },
      {
        "h-10 px-4 py-2": size === "default",
        "h-9 rounded-md px-3": size === "sm",
        "h-11 rounded-md px-8": size === "lg",
        "h-10 w-10": size === "icon",
      },
      className
    )

    if (animated) {
      return (
        <motion.button
          className={baseClassName}
          ref={ref}
          whileHover={{ scale: 1.02 }}
          whileTap={{ scale: 0.98 }}
          {...(props as any)}
        />
      )
    }

    return (
      <button
        className={baseClassName}
        ref={ref}
        {...props}
      />
    )
  }
)
Button.displayName = "Button"

export { Button }
