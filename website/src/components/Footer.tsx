import Link from 'next/link'
import { Button } from '@/components/Button'
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
    { label: 'GitHub Repository', href: 'https://github.com/anilpdv/libgen-gui' },
    { label: 'Releases (v2.0.1)', href: 'https://github.com/anilpdv/libgen-gui/releases' },
    { label: 'Issue Tracker', href: 'https://github.com/anilpdv/libgen-gui/issues' },
    { label: 'Discussions', href: 'https://github.com/anilpdv/libgen-gui/discussions' },
  ],
  project: [
    { label: 'Documentation', href: 'https://github.com/anilpdv/libgen-gui#readme' },
    { label: 'MIT License', href: 'https://github.com/anilpdv/libgen-gui/blob/main/LICENSE' },
    { label: 'Legal Notice', href: '#faqs' },
  ],
}

export function Footer() {
  return (
    <footer className="border-t border-gray-200/90 bg-gray-900 text-gray-400">
      <Container className="py-16 sm:py-20">
        <div className="grid grid-cols-1 gap-10 sm:grid-cols-2 lg:grid-cols-5">
          <div className="lg:col-span-2">
            <div className="flex items-center gap-3 text-white">
              <Logomark className="h-8 w-8 shrink-0 rounded-lg shadow-sm" />
              <span className="text-lg font-bold tracking-tight">
                LibGen <span className="text-cyan-400">GUI</span>
              </span>
            </div>
            <p className="mt-4 text-sm text-gray-300 leading-relaxed max-w-sm">
              Fast, lightweight native client for searching, discovering, and downloading research papers and books with automatic mirror failover and queue recovery.
            </p>
            <div className="mt-6">
              <Button
                href="#downloads"
                variant="solid"
                color="cyan"
                className="text-xs font-semibold py-2 px-4"
              >
                Download App (v2.0.1)
              </Button>
            </div>
          </div>

          <div>
            <p className="text-xs font-semibold uppercase tracking-wider text-gray-200">
              Product
            </p>
            <ul role="list" className="mt-4 space-y-2.5">
              {footerLinks.product.map((link) => (
                <li key={link.label}>
                  <Link
                    href={link.href}
                    className="text-sm text-gray-400 hover:text-white transition-colors"
                  >
                    {link.label}
                  </Link>
                </li>
              ))}
            </ul>
          </div>

          <div>
            <p className="text-xs font-semibold uppercase tracking-wider text-gray-200">
              Resources
            </p>
            <ul role="list" className="mt-4 space-y-2.5">
              {footerLinks.resources.map((link) => (
                <li key={link.label}>
                  <Link
                    href={link.href}
                    className="text-sm text-gray-400 hover:text-white transition-colors"
                  >
                    {link.label}
                  </Link>
                </li>
              ))}
            </ul>
          </div>

          <div>
            <p className="text-xs font-semibold uppercase tracking-wider text-gray-200">
              Project
            </p>
            <ul role="list" className="mt-4 space-y-2.5">
              {footerLinks.project.map((link) => (
                <li key={link.label}>
                  <Link
                    href={link.href}
                    className="text-sm text-gray-400 hover:text-white transition-colors"
                  >
                    {link.label}
                  </Link>
                </li>
              ))}
            </ul>
          </div>
        </div>

        <div className="mt-14 pt-8 border-t border-gray-800 flex flex-col sm:flex-row items-center justify-between gap-4 text-xs text-gray-500">
          <p>
            &copy; {new Date().getFullYear()} LibGen GUI Contributors. Distributed under the MIT License.
          </p>
          <p className="text-gray-500">
            Independent search client designed for research and academic study.
          </p>
        </div>
      </Container>
    </footer>
  )
}
