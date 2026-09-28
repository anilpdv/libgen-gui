import { useId } from 'react'

import { Container } from '@/components/Container'

const features = [
  {
    name: 'Multi-Field Advanced Search',
    description:
      'Search seamlessly across Title, Author, Series, Publisher, Year, ISBN, or MD5 checksum with immediate debounced queries.',
    icon: SearchIcon,
  },
  {
    name: 'Atomic Resumable Downloads',
    description:
      'Download safely with temporary .part file isolation, HTTP Range chunk resumption, and automatic cleanup on error or cancellation.',
    icon: DownloadIcon,
  },
  {
    name: 'Android Scoped Storage (SAF)',
    description:
      'Native DocumentFile and Storage Access Framework integration ensuring smooth downloads on Android 10 through Android 15.',
  icon: AndroidIcon,
  },
  {
    name: 'Zero Electron Overhead',
    description:
      'Compiled with pure Go and Fyne into a single standalone native executable that uses less than 35MB of RAM.',
    icon: SpeedIcon,
  },
  {
    name: 'Continuous Health Probing',
    description:
      'Probes mirror response times in the background with exponential backoff and automatic cooldown for degraded nodes.',
    icon: ShieldIcon,
  },
  {
    name: 'Persistent Crash-Safe Queue',
    description:
      'Atomic JSON ledger keeps your download queue safe across app restarts and sudden network disconnections.',
    icon: QueueIcon,
  },
]

function SearchIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 32 32" fill="none" aria-hidden="true" {...props}>
      <circle cx={16} cy={16} r={16} fill="#06B6D4" fillOpacity={0.12} />
      <path
        d="M14 19a5 5 0 1 0 0-10 5 5 0 0 0 0 10zm3.5-1.5L22 22"
        stroke="#0891B2"
        strokeWidth="2"
        strokeLinecap="round"
      />
    </svg>
  )
}

function DownloadIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 32 32" fill="none" aria-hidden="true" {...props}>
      <circle cx={16} cy={16} r={16} fill="#06B6D4" fillOpacity={0.12} />
      <path
        d="M11 15l5 5 5-5M16 9v11M10 23h12"
        stroke="#0891B2"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  )
}

function AndroidIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 32 32" fill="none" aria-hidden="true" {...props}>
      <circle cx={16} cy={16} r={16} fill="#06B6D4" fillOpacity={0.12} />
      <path
        d="M10 18v5M22 18v5M13 23v4M19 23v4M12 14a4 4 0 0 1 8 0v8H12v-8z"
        stroke="#0891B2"
        strokeWidth="2"
        strokeLinecap="round"
      />
    </svg>
  )
}

function SpeedIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 32 32" fill="none" aria-hidden="true" {...props}>
      <circle cx={16} cy={16} r={16} fill="#06B6D4" fillOpacity={0.12} />
      <path
        d="M16 11v5l3 3M9 16a7 7 0 1 1 14 0 7 7 0 0 1-14 0z"
        stroke="#0891B2"
        strokeWidth="2"
        strokeLinecap="round"
      />
    </svg>
  )
}

function ShieldIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 32 32" fill="none" aria-hidden="true" {...props}>
      <circle cx={16} cy={16} r={16} fill="#06B6D4" fillOpacity={0.12} />
      <path
        d="M16 9l6 3v4c0 4-3 7-6 8-3-1-6-4-6-8v-4l6-3z"
        stroke="#0891B2"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  )
}

function QueueIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 32 32" fill="none" aria-hidden="true" {...props}>
      <circle cx={16} cy={16} r={16} fill="#06B6D4" fillOpacity={0.12} />
      <path
        d="M10 11h12M10 16h12M10 21h8"
        stroke="#0891B2"
        strokeWidth="2"
        strokeLinecap="round"
      />
    </svg>
  )
}

export function SecondaryFeatures() {
  return (
    <section
      id="secondary-features"
      aria-label="Technical Highlights"
      className="py-20 sm:py-32"
    >
      <Container>
        <div className="mx-auto max-w-2xl sm:text-center">
          <h2 className="text-3xl font-medium tracking-tight text-gray-900 sm:text-4xl">
            Modern architecture. High performance.
          </h2>
          <p className="mt-3 text-lg text-gray-600 leading-relaxed">
            Designed for researchers, students, and bibliophiles who need dependable access without frustrating browser CAPTCHAs, stalled tabs, or failed downloads.
          </p>
        </div>
        <ul
          role="list"
          className="mx-auto mt-16 grid max-w-2xl grid-cols-1 gap-6 text-sm sm:mt-20 sm:grid-cols-2 md:gap-y-10 lg:max-w-none lg:grid-cols-3"
        >
          {features.map((feature) => (
            <li
              key={feature.name}
              className="rounded-2xl border border-gray-200/80 p-8 transition-shadow hover:shadow-md hover:border-cyan-200"
            >
              <feature.icon className="h-8 w-8" />
              <h3 className="mt-6 font-semibold text-gray-900 text-base">
                {feature.name}
              </h3>
              <p className="mt-2 text-gray-600 leading-relaxed">{feature.description}</p>
            </li>
          ))}
        </ul>
      </Container>
    </section>
  )
}
