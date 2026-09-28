import { Button } from '@/components/Button'
import { Container } from '@/components/Container'

const downloadOptions = [
  {
    name: 'macOS & Desktop App',
    badge: 'v2.0.1 Native Build',
    description:
      'Native desktop application for Apple Silicon & Intel macOS. Features live mirror failover, background queue, and customizable destination settings.',
    specs: [
      'Universal macOS .app bundle & zip',
      'Instant debounced multi-field search',
      'Configurable download directory',
      'Resumable background queue',
      'Zero trackers & zero telemetry',
    ],
    primaryAction: {
      label: 'Download macOS (.zip)',
      href: 'https://github.com/anilpdv/libgen-gui/releases/download/v2.0.1/LibGen.Downloader-v2.0.1-macos.zip',
    },
    featured: true,
  },
  {
    name: 'Android App (APK)',
    badge: 'Scoped SAF Ready',
    description:
      'Native mobile client with Android Storage Access Framework (SAF) integration for Android 10 through Android 15 phones & tablets.',
    specs: [
      'DocumentFile Scoped Storage (SAF)',
      'Touch-optimized responsive list cards',
      'Direct save to Books & e-reader folders',
      'ARM64 & Universal APK packages',
      '100% Free & Open Source',
    ],
    primaryAction: {
      label: 'Download Android APK',
      href: 'https://github.com/anilpdv/libgen-gui/releases/download/v2.0.1/LibGen-Downloader-Android-arm64.apk',
    },
    featured: false,
  },
  {
    name: 'Source & Developer',
    badge: 'MIT Open Source',
    description:
      'Clone and build from source for Linux, Windows, or macOS. Full access to internal/download, internal/storage, and pkg/libgen packages.',
    specs: [
      'Clean Go 1.22+ architecture',
      'Fyne v2 native vector UI toolkit',
      'Automated test suite with race detector',
      'Cross-compilation build scripts included',
      'Permissive MIT open source license',
    ],
    primaryAction: {
      label: 'View on GitHub',
      href: 'https://github.com/anilpdv/libgen-gui',
    },
    featured: false,
  },
]

function CheckIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 20 20" fill="currentColor" aria-hidden="true" {...props}>
      <path
        fillRule="evenodd"
        d="M16.704 4.153a.75.75 0 01.143 1.052l-8 10.5a.75.75 0 01-1.127.075l-4.5-4.5a.75.75 0 011.06-1.06l3.894 3.893 7.48-9.817a.75.75 0 011.05-.143z"
        clipRule="evenodd"
      />
    </svg>
  )
}

export function Pricing() {
  return (
    <section
      id="downloads"
      aria-label="Download Options"
      className="border-t border-gray-200/80 bg-gray-100 py-20 sm:py-28"
    >
      <Container>
        <div className="mx-auto max-w-2xl text-center">
          <div className="inline-flex items-center gap-2 rounded-full bg-cyan-100 px-3.5 py-1 text-xs font-semibold text-cyan-800 ring-1 ring-cyan-700/20 mb-4">
            <span>GET LIBGEN GUI</span>
          </div>
          <h2 className="text-3xl font-bold tracking-tight text-gray-900 sm:text-4xl">
            Free and open source. No ads. No telemetry.
          </h2>
          <p className="mt-4 text-base text-gray-600 leading-relaxed sm:text-lg">
            Choose your platform below. All releases are self-contained with zero runtime dependencies.
          </p>
        </div>

        <div className="mt-14 grid grid-cols-1 gap-8 lg:grid-cols-3 items-stretch">
          {downloadOptions.map((option) => (
            <div
              key={option.name}
              className={`rounded-2xl p-7 flex flex-col justify-between transition-all duration-200 ${
                option.featured
                  ? 'bg-gray-950 text-white shadow-xl ring-2 ring-cyan-500'
                  : 'bg-white text-gray-900 shadow-2xs border border-gray-200/90'
              }`}
            >
              <div>
                <div className="flex items-center justify-between gap-2">
                  <h3 className={`text-lg font-bold ${option.featured ? 'text-white' : 'text-gray-900'}`}>
                    {option.name}
                  </h3>
                  <span
                    className={`text-[11px] font-semibold px-2 py-0.5 rounded-full ${
                      option.featured
                        ? 'bg-cyan-950 text-cyan-400 border border-cyan-800/60'
                        : 'bg-cyan-50 text-cyan-700 border border-cyan-200'
                    }`}
                  >
                    {option.badge}
                  </span>
                </div>

                <p
                  className={`mt-3 text-sm leading-relaxed ${
                    option.featured ? 'text-gray-300' : 'text-gray-600'
                  }`}
                >
                  {option.description}
                </p>

                <div className={`mt-6 pt-6 border-t ${option.featured ? 'border-gray-800' : 'border-gray-100'}`}>
                  <ul role="list" className="space-y-3">
                    {option.specs.map((spec) => (
                      <li key={spec} className="flex items-start gap-3 text-sm">
                        <CheckIcon
                          className={`h-5 w-5 shrink-0 mt-0.5 ${
                            option.featured ? 'text-cyan-400' : 'text-cyan-600'
                          }`}
                        />
                        <span className={option.featured ? 'text-gray-200' : 'text-gray-700'}>
                          {spec}
                        </span>
                      </li>
                    ))}
                  </ul>
                </div>
              </div>

              <div className="mt-8 pt-4">
                <Button
                  href={option.primaryAction.href}
                  variant="solid"
                  color={option.featured ? 'cyan' : 'gray'}
                  className="w-full text-center"
                >
                  {option.primaryAction.label}
                </Button>
              </div>
            </div>
          ))}
        </div>
      </Container>
    </section>
  )
}
