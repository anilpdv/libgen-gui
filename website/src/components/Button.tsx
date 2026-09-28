import Link from 'next/link'
import clsx from 'clsx'

const baseStyles = {
  solid:
    'inline-flex items-center justify-center rounded-lg px-5 py-2.5 min-h-[44px] text-sm font-semibold transition-all duration-150 active:scale-[0.98]',
  outline:
    'inline-flex items-center justify-center rounded-lg border px-5 py-2.5 min-h-[44px] text-sm font-semibold transition-all duration-150 active:scale-[0.98]',
}

const variantStyles = {
  solid: {
    cyan: 'bg-cyan-600 text-white hover:bg-cyan-700 active:bg-cyan-800 shadow-sm shadow-cyan-600/20',
    white:
      'bg-white text-gray-900 hover:bg-gray-100 active:bg-gray-200 shadow-sm',
    gray: 'bg-gray-900 text-white hover:bg-gray-800 active:bg-gray-950 shadow-sm',
  },
  outline: {
    gray: 'border-gray-300 text-gray-700 hover:border-gray-400 hover:bg-gray-50 active:bg-gray-100',
    white: 'border-white/40 text-white hover:border-white hover:bg-white/10 active:bg-white/20',
    cyan: 'border-cyan-500/40 text-cyan-600 hover:border-cyan-600 hover:bg-cyan-50 active:bg-cyan-100',
  },
}

type ButtonProps = (
  | {
      variant?: 'solid'
      color?: keyof typeof variantStyles.solid
    }
  | {
      variant: 'outline'
      color?: keyof typeof variantStyles.outline
    }
) &
  (
    | Omit<React.ComponentPropsWithoutRef<typeof Link>, 'color'>
    | (Omit<React.ComponentPropsWithoutRef<'button'>, 'color'> & {
        href?: undefined
      })
  )

export function Button({ className, ...props }: ButtonProps) {
  props.variant ??= 'solid'
  props.color ??= 'gray'

  className = clsx(
    baseStyles[props.variant],
    props.variant === 'outline'
      ? variantStyles.outline[props.color]
      : props.variant === 'solid'
        ? variantStyles.solid[props.color]
        : undefined,
    className,
  )

  return typeof props.href === 'undefined' ? (
    <button className={className} {...props} />
  ) : (
    <Link className={className} {...props} />
  )
}
