import { Container } from '@/components/Container'

const useCases = [
  {
    title: 'Researchers & Academics',
    role: 'Literature Review & Citation Sourcing',
    description:
      'Search across live mirrors simultaneously and retrieve high-resolution papers, journals, and monographs without manual website switching or CAPTCHAs.',
    icon: AcademicIcon,
  },
  {
    title: 'Students & Educators',
    role: 'Course Textbooks & Reference Packs',
    description:
      'Queue multiple textbooks and syllabus readings in the background. If Wi-Fi drops on campus, resumable chunk transfers pick up automatically without data loss.',
    icon: StudentIcon,
  },
  {
    title: 'Technical Readers & Archivists',
    role: 'Offline Digital Libraries',
    description:
      'Manage multi-gigabyte digital collections via a lightweight native Go client with zero Electron bloat, HTTP Range resumption, and MD5 checksum validation.',
    icon: ArchivistIcon,
  },
]

function AcademicIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 24 24" fill="none" aria-hidden="true" {...props}>
      <path d="M12 3L1 9l11 6 9-4.91V17h2V9L12 3zM5 13.18v4L12 21l7-3.82v-4L12 17l-7-3.82z" fill="#0891b2" />
    </svg>
  )
}

function StudentIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 24 24" fill="none" aria-hidden="true" {...props}>
      <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20" stroke="#0891b2" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
      <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z" stroke="#0891b2" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}

function ArchivistIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 24 24" fill="none" aria-hidden="true" {...props}>
      <path d="M21 8v13H3V8M1 3h22v5H1V3zM10 12h4" stroke="#0891b2" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}

export function UseCases() {
  return (
    <section
      id="use-cases"
      aria-label="Built for Research Workflows"
      className="py-20 sm:py-28 bg-gray-50 border-t border-gray-200/80"
    >
      <Container>
        <div className="mx-auto max-w-2xl sm:text-center">
          <div className="inline-flex items-center gap-2 rounded-full bg-cyan-50 px-3.5 py-1 text-xs font-semibold text-cyan-800 ring-1 ring-cyan-700/15 mb-4">
            <span>REAL-WORLD WORKFLOWS</span>
          </div>
          <h2 className="text-3xl font-bold tracking-tight text-gray-900 sm:text-4xl">
            Built for research-heavy workflows.
          </h2>
          <p className="mt-4 text-base text-gray-600 leading-relaxed sm:text-lg">
            Purpose-built to replace tedious browser scraping with a dependable, local-first search and download client.
          </p>
        </div>

        <div className="mt-14 grid grid-cols-1 gap-8 sm:grid-cols-2 lg:grid-cols-3">
          {useCases.map((item) => {
            let Icon = item.icon
            return (
              <div
                key={item.title}
                className="rounded-2xl border border-gray-200/90 bg-white p-7 shadow-2xs hover:shadow-md hover:border-cyan-400/80 transition-all duration-200 flex flex-col"
              >
                <div className="rounded-xl bg-cyan-50 p-3 w-fit">
                  <Icon className="h-6 w-6" />
                </div>
                <h3 className="mt-5 text-lg font-bold text-gray-900">
                  {item.title}
                </h3>
                <p className="text-xs font-semibold text-cyan-700 uppercase tracking-wider mt-1">
                  {item.role}
                </p>
                <p className="mt-3 text-sm text-gray-600 leading-relaxed flex-auto">
                  {item.description}
                </p>
              </div>
            )
          })}
        </div>
      </Container>
    </section>
  )
}
