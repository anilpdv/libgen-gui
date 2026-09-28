import Link from 'next/link'
import { Container } from '@/components/Container'
import { Logomark } from '@/components/Logo'

const footerLinks = {
  product: [
    { label: 'Features', href: '#features' },
    { label: 'Architecture', href: '#architecture' },
    { label: 'Use Cases', href: '#use-cases' },
    { label: 'Downloads', href: '#downloads' },
    { label: 'FAQs', href: '#faqs' },
  ],
  resources: [
    { label: 'GitHub Releases (v2.0.1)', href: 'https://github.com/anilpdv/libgen-gui/releases' },
    { label: 'Issue Tracker', href: 'https://github.com/anilpdv/libgen-gui/issues' },
    { label: 'Discussions', href: 'https://github.com/anilpdv/libgen-gui/discussions' },
    { label: 'Release Notes', href: 'https://github.com/anilpdv/libgen-gui/releases/tag/v2.0.1' },
  ],
  project: [
    { label: 'Documentation', href: 'https://github.com/anilpdv/libgen-gui#readme' },
    { label: 'MIT License', href: 'https://github.com/anilpdv/libgen-gui/blob/main/LICENSE' },
    { label: 'Source Repository', href: 'https://github.com/anilpdv/libgen-gui' },
  ],
}

function GithubIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true" {...props}>
      <path
        fillRule="evenodd"
        clipRule="evenodd"
        d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.53 1.032 1.53 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"
      />
    </svg>
  )
}

export function Footer() {
  return (
    <footer className="border-t border-gray-800 bg-gray-950 text-gray-400">
      <Container className="py-16 sm:py-20">
        <div className="grid grid-cols-1 gap-10 sm:grid-cols-2 lg:grid-cols-5">
          <div className="lg:col-span-2">
            <div className="flex items-center gap-3 text-white">
              <Logomark className="h-8 w-8 shrink-0 rounded-lg shadow-sm" />
              <span className="text-lg font-bold tracking-tight">
                LibGen <span className="text-cyan-400">GUI</span>
              </span>
            </div>
            <p className="mt-4 text-[13px] sm:text-sm text-gray-300 leading-relaxed max-w-sm">
              Fast, lightweight native client for searching, discovering, and downloading research papers and books with automatic mirror failover and queue recovery.
            </p>
            <div className="mt-6">
              <Link
                href="https://github.com/anilpdv/libgen-gui"
                className="inline-flex items-center gap-2 rounded-xl border border-gray-700 bg-gray-900 px-4 py-2 text-xs font-semibold text-gray-200 hover:bg-gray-800 hover:text-white transition-colors"
              >
                <GithubIcon className="h-4 w-4" />
                <span>View on GitHub</span>
              </Link>
            </div>
          </div>

          <div>
            <p className="text-[11px] font-bold tracking-wider uppercase text-gray-300">
              Product
            </p>
            <ul role="list" className="mt-4 space-y-2">
              {footerLinks.product.map((link) => (
                <li key={link.label}>
                  <Link
                    href={link.href}
                    className="min-h-[28px] inline-flex items-center text-[13px] text-gray-400 hover:text-cyan-400 transition-colors"
                  >
                    {link.label}
                  </Link>
                </li>
              ))}
            </ul>
          </div>

          <div>
            <p className="text-[11px] font-bold tracking-wider uppercase text-gray-300">
              Resources
            </p>
            <ul role="list" className="mt-4 space-y-2">
              {footerLinks.resources.map((link) => (
                <li key={link.label}>
                  <Link
                    href={link.href}
                    className="min-h-[28px] inline-flex items-center text-[13px] text-gray-400 hover:text-cyan-400 transition-colors"
                  >
                    {link.label}
                  </Link>
                </li>
              ))}
            </ul>
          </div>

          <div>
            <p className="text-[11px] font-bold tracking-wider uppercase text-gray-300">
              Project
            </p>
            <ul role="list" className="mt-4 space-y-2">
              {footerLinks.project.map((link) => (
                <li key={link.label}>
                  <Link
                    href={link.href}
                    className="min-h-[28px] inline-flex items-center text-[13px] text-gray-400 hover:text-cyan-400 transition-colors"
                  >
                    {link.label}
                  </Link>
                </li>
              ))}
            </ul>
          </div>
        </div>

        <div className="mt-14 pt-8 border-t border-gray-800/80 flex flex-col sm:flex-row items-center justify-between gap-4 text-xs text-gray-400">
          <p>
            &copy; {new Date().getFullYear()} LibGen GUI Contributors. Distributed under the MIT License.
          </p>
          <p className="text-gray-400">
            Independent search client designed for research and academic study.
          </p>
        </div>
      </Container>
    </footer>
  )
}
