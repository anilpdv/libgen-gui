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
      'Download safely with temporary .part file isolation, HTTP Range chunk resumption, and atomic rename upon verified completion.',
    icon: DownloadIcon,
  },
  {
    name: 'Android Scoped Storage (SAF)',
    description:
      'Native DocumentFile and Storage Access Framework integration ensuring direct folder selection on Android 10 through Android 15.',
    icon: AndroidIcon,
  },
  {
    name: 'Decoupled Clean Architecture',
    description:
      'Modular Go packages separating network mirrors, download queues, storage targets, and Fyne vector UI into testable layers.',
    icon: ArchitectureIcon,
  },
  {
    name: 'Continuous Mirror Health Prober',
    description:
      'Probes mirror response times in the background with exponential backoff and automatic failover for degraded or blocked nodes.',
    icon: ShieldIcon,
  },
  {
    name: 'Persistent Crash-Safe Queue',
    description:
      'Atomic JSON ledger preserves your FIFO download queue and captures per-task destinations across application restarts.',
    icon: QueueIcon,
  },
]

function SearchIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 24 24" fill="none" aria-hidden="true" {...props}>
      <circle cx={11} cy={11} r={7} stroke="#0891b2" strokeWidth="2" />
      <path d="M16 16l5 5" stroke="#0891b2" strokeWidth="2" strokeLinecap="round" />
    </svg>
  )
}

function DownloadIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 24 24" fill="none" aria-hidden="true" {...props}>
      <path d="M12 4v12M7 11l5 5 5-5M4 19h16" stroke="#0891b2" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}

function AndroidIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 24 24" fill="none" aria-hidden="true" {...props}>
      <path d="M4 10v6M20 10v6M9 17v4M15 17v4M7 9a5 5 0 0 1 10 0v8H7V9z" stroke="#0891b2" strokeWidth="2" strokeLinecap="round" />
      <circle cx={10} cy={11} r={1} fill="#0891b2" />
      <circle cx={14} cy={11} r={1} fill="#0891b2" />
    </svg>
  )
}

function ArchitectureIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 24 24" fill="none" aria-hidden="true" {...props}>
      <rect x={3} y={3} width={7} height={7} rx={1.5} stroke="#0891b2" strokeWidth="2" />
      <rect x={14} y={3} width={7} height={7} rx={1.5} stroke="#0891b2" strokeWidth="2" />
      <rect x={3} y={14} width={7} height={7} rx={1.5} stroke="#0891b2" strokeWidth="2" />
      <rect x={14} y={14} width={7} height={7} rx={1.5} stroke="#0891b2" strokeWidth="2" />
    </svg>
  )
}

function ShieldIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 24 24" fill="none" aria-hidden="true" {...props}>
      <path d="M12 3l8 3.5v5c0 5-3.5 8.5-8 9.5-4.5-1-8-4.5-8-9.5v-5L12 3z" stroke="#0891b2" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
      <path d="M9 12l2 2 4-4" stroke="#0891b2" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}

function QueueIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 24 24" fill="none" aria-hidden="true" {...props}>
      <path d="M4 6h16M4 12h16M4 18h10" stroke="#0891b2" strokeWidth="2" strokeLinecap="round" />
      <circle cx={18} cy={18} r={3} stroke="#0891b2" strokeWidth="2" />
    </svg>
  )
}

export function SecondaryFeatures() {
  return (
    <section
      id="architecture"
      aria-label="Architecture & Capabilities"
      className="py-20 sm:py-28 bg-white border-t border-gray-200/80"
    >
      <Container>
        <div className="mx-auto max-w-2xl sm:text-center">
          <div className="inline-flex items-center gap-2 rounded-full bg-cyan-50 px-3.5 py-1 text-xs font-semibold text-cyan-800 ring-1 ring-cyan-700/15 mb-4">
            <span>CORE ARCHITECTURE</span>
          </div>
          <h2 className="text-3xl font-bold tracking-tight text-gray-900 sm:text-4xl">
            Modern architecture. High performance.
          </h2>
          <p className="mt-4 text-base text-gray-600 leading-relaxed sm:text-lg">
            Built for researchers, students, and technical readers who need dependable access without browser CAPTCHAs, stalled tabs, or failed downloads.
          </p>
        </div>

        <div className="mt-14 grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {features.map((feature) => {
            let Icon = feature.icon
            return (
              <div
                key={feature.name}
                className="rounded-2xl border border-gray-200/80 bg-white p-6 shadow-2xs hover:border-cyan-400/80 hover:shadow-md transition-all duration-200"
              >
                <div className="rounded-xl bg-cyan-50 p-3 w-fit text-cyan-600">
                  <Icon className="h-6 w-6" />
                </div>
                <h3 className="mt-5 text-base font-semibold text-gray-900">
                  {feature.name}
                </h3>
                <p className="mt-2 text-sm text-gray-600 leading-relaxed">
                  {feature.description}
                </p>
              </div>
            )
          })}
        </div>
      </Container>
    </section>
  )
}
