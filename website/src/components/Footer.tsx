import Link from 'next/link'

import { Button } from '@/components/Button'
import { Container } from '@/components/Container'
import { Logomark } from '@/components/Logo'
import { NavLinks } from '@/components/NavLinks'

export function Footer() {
  return (
    <footer className="border-t border-gray-200 bg-white">
      <Container>
        <div className="flex flex-col items-start justify-between gap-y-12 pt-16 pb-8 lg:flex-row lg:items-center lg:py-16">
          <div>
            <div className="flex items-center text-gray-900">
              <Logomark className="h-10 w-10 flex-none fill-cyan-500" />
              <div className="ml-4">
                <p className="text-lg font-bold tracking-tight">LibGen GUI</p>
                <p className="mt-0.5 text-xs text-gray-500">
                  Fast, native book & paper search and downloader.
                </p>
              </div>
            </div>
            <nav className="mt-8 flex flex-wrap gap-x-8 gap-y-3">
              <NavLinks />
            </nav>
          </div>
          <div className="flex flex-col sm:flex-row gap-4">
            <Button
              href="https://github.com/anilpdv/libgen-gui/releases"
              color="cyan"
            >
              Download Latest Release (v2.0.0)
            </Button>
            <Button
              href="https://github.com/anilpdv/libgen-gui"
              variant="outline"
            >
              GitHub Repository
            </Button>
          </div>
        </div>
        <div className="flex flex-col items-center justify-between border-t border-gray-100 py-8 md:flex-row">
          <p className="text-xs text-gray-500">
            &copy; {new Date().getFullYear()} LibGen GUI Contributors. MIT Licensed.
          </p>
          <p className="mt-4 text-xs text-gray-400 md:mt-0">
            Built with Go & Fyne • Designed for academic research and fair use.
          </p>
        </div>
      </Container>
    </footer>
  )
}
