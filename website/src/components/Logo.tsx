import React from 'react'

export function Logomark(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 40 40" fill="none" aria-hidden="true" {...props}>
      <rect width="40" height="40" rx="10" fill="#0284c7" />
      <path
        d="M12 12C12 10.8954 12.8954 10 14 10H26C27.1046 10 28 10.8954 28 12V28C28 29.1046 27.1046 30 26 30H14C12.8954 30 12 29.1046 12 28V12Z"
        fill="#0369a1"
      />
      <path
        d="M14 13H26V27H14V13Z"
        fill="#ffffff"
      />
      <path
        d="M17 18L20 21L23 18"
        stroke="#0284c7"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M20 15V21"
        stroke="#0284c7"
        strokeWidth="2"
        strokeLinecap="round"
      />
      <path
        d="M17 24H23"
        stroke="#0284c7"
        strokeWidth="2"
        strokeLinecap="round"
      />
    </svg>
  )
}

export function Logo({ className = '', ...props }: React.ComponentPropsWithoutRef<'div'>) {
  return (
    <div className={`flex items-center gap-3 ${className}`} {...props}>
      <Logomark className="h-9 w-9 shrink-0 shadow-sm rounded-lg" />
      <div className="flex flex-col">
        <span className="text-lg font-bold tracking-tight text-gray-900 leading-tight">
          LibGen <span className="text-cyan-600 font-extrabold">GUI</span>
        </span>
        <span className="text-[10px] font-medium tracking-wider uppercase text-gray-500">
          Downloader v2.0
        </span>
      </div>
    </div>
  )
}
