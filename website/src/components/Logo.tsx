import React from 'react'

export function Logomark(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 40 40" fill="none" aria-hidden="true" {...props}>
      <rect width="40" height="40" rx="10" fill="#0891b2" />
      <path
        d="M12 11C12 9.89543 12.8954 9 14 9H26C27.1046 9 28 9.89543 28 11V29C28 30.1046 27.1046 31 26 31H14C12.8954 31 12 30.1046 12 29V11Z"
        fill="#0e7490"
      />
      <path
        d="M14 12H26V28H14V12Z"
        fill="#ffffff"
      />
      <path
        d="M17 17.5L20 20.5L23 17.5"
        stroke="#0891b2"
        strokeWidth="2.2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M20 14V20.5"
        stroke="#0891b2"
        strokeWidth="2.2"
        strokeLinecap="round"
      />
      <path
        d="M16.5 24H23.5"
        stroke="#0891b2"
        strokeWidth="2.2"
        strokeLinecap="round"
      />
    </svg>
  )
}

export function Logo({ className = '', ...props }: React.ComponentPropsWithoutRef<'div'>) {
  return (
    <div className={`flex items-center gap-3 ${className}`} {...props}>
      <Logomark className="h-8 w-8 shrink-0 rounded-lg shadow-sm" />
      <div className="flex flex-col">
        <span className="text-base font-bold tracking-tight text-gray-900 leading-none">
          LibGen <span className="text-cyan-600 font-extrabold">GUI</span>
        </span>
        <span className="text-[11px] font-medium tracking-wider uppercase text-gray-500 mt-1">
          v2.0 Native Client
        </span>
      </div>
    </div>
  )
}
