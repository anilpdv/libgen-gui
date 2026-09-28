'use client'

import Link from 'next/link'
import {
  Popover,
  PopoverButton,
  PopoverBackdrop,
  PopoverPanel,
} from '@headlessui/react'
import { AnimatePresence, motion } from 'framer-motion'

import { Button } from '@/components/Button'
import { Container } from '@/components/Container'
import { Logo } from '@/components/Logo'
import { NavLinks } from '@/components/NavLinks'

function MenuIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 24 24" fill="none" aria-hidden="true" {...props}>
      <path
        d="M4 6h16M4 12h16M4 18h16"
        strokeWidth={2}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  )
}

function CloseIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 24 24" fill="none" aria-hidden="true" {...props}>
      <path
        d="M6 18L18 6M6 6l12 12"
        strokeWidth={2}
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  )
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

function MobileNavLink(
  props: Omit<
    React.ComponentPropsWithoutRef<typeof PopoverButton<typeof Link>>,
    'as' | 'className'
  >,
) {
  return (
    <PopoverButton
      as={Link}
      className="block text-base font-medium py-2 text-gray-800 hover:text-cyan-600 transition-colors"
      {...props}
    />
  )
}

export function Header() {
  return (
    <header className="sticky top-0 z-50 bg-white/90 backdrop-blur-md border-b border-gray-200/80 transition-all">
      <nav aria-label="Main Navigation">
        <Container className="flex h-[70px] items-center justify-between">
          <div className="flex items-center gap-8 lg:gap-10">
            <Link href="/" aria-label="LibGen GUI Home">
              <Logo />
            </Link>
            <div className="hidden lg:flex lg:items-center lg:gap-7">
              <NavLinks />
            </div>
          </div>

          <div className="flex items-center gap-3">
            <div className="hidden sm:flex sm:items-center sm:gap-3">
              <Button
                href="https://github.com/anilpdv/libgen-gui"
                variant="outline"
                color="gray"
                className="gap-2"
                aria-label="View source repository on GitHub"
              >
                <GithubIcon className="h-4 w-4 text-gray-700" />
                <span>GitHub</span>
              </Button>
              <Button
                href="#downloads"
                variant="solid"
                color="cyan"
                aria-label="Download LibGen GUI application"
              >
                Download App
              </Button>
            </div>

            <Popover className="lg:hidden">
              {({ open }) => (
                <>
                  <PopoverButton
                    className="relative z-10 inline-flex items-center justify-center rounded-lg p-2.5 text-gray-700 hover:bg-gray-100 hover:text-gray-900 focus:outline-hidden"
                    aria-expanded={open}
                    aria-controls="mobile-navigation-panel"
                    aria-label={open ? 'Close main menu' : 'Open main menu'}
                  >
                    {open ? (
                      <CloseIcon className="h-6 w-6 stroke-current" />
                    ) : (
                      <MenuIcon className="h-6 w-6 stroke-current" />
                    )}
                  </PopoverButton>
                  <AnimatePresence initial={false}>
                    {open && (
                      <>
                        <PopoverBackdrop
                          static
                          as={motion.div}
                          initial={{ opacity: 0 }}
                          animate={{ opacity: 1 }}
                          exit={{ opacity: 0 }}
                          className="fixed inset-0 z-40 bg-gray-900/40 backdrop-blur-xs"
                        />
                        <PopoverPanel
                          static
                          id="mobile-navigation-panel"
                          as={motion.div}
                          initial={{ opacity: 0, y: -16 }}
                          animate={{ opacity: 1, y: 0 }}
                          exit={{
                            opacity: 0,
                            y: -16,
                            transition: { duration: 0.15 },
                          }}
                          className="absolute inset-x-4 top-20 z-50 origin-top rounded-2xl bg-white p-6 shadow-xl ring-1 ring-gray-900/10"
                        >
                          <div className="divide-y divide-gray-100">
                            <div className="space-y-1 pb-4">
                              <MobileNavLink href="#features">Features</MobileNavLink>
                              <MobileNavLink href="#architecture">Architecture</MobileNavLink>
                              <MobileNavLink href="#use-cases">Use Cases</MobileNavLink>
                              <MobileNavLink href="#downloads">Downloads</MobileNavLink>
                              <MobileNavLink href="#faqs">FAQs</MobileNavLink>
                            </div>
                            <div className="pt-4 flex flex-col gap-3">
                              <Button
                                href="#downloads"
                                variant="solid"
                                color="cyan"
                                className="w-full"
                              >
                                Download App
                              </Button>
                              <Button
                                href="https://github.com/anilpdv/libgen-gui"
                                variant="outline"
                                color="gray"
                                className="w-full gap-2"
                              >
                                <GithubIcon className="h-4 w-4" />
                                <span>Star on GitHub</span>
                              </Button>
                            </div>
                          </div>
                        </PopoverPanel>
                      </>
                    )}
                  </AnimatePresence>
                </>
              )}
            </Popover>
          </div>
        </Container>
      </nav>
    </header>
  )
}
