import clsx from 'clsx'

export function Container({
  className,
  ...props
}: React.ComponentPropsWithoutRef<'div'>) {
  return (
    <div
      className={clsx('mx-auto w-[min(100%-40px,1200px)] max-sm:w-[min(100%-32px,1200px)]', className)}
      {...props}
    />
  )
}
