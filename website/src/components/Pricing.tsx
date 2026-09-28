'use client'

import clsx from 'clsx'

import { Button } from '@/components/Button'
import { Container } from '@/components/Container'
import { Logomark } from '@/components/Logo'

const plans = [
  {
    name: 'Desktop App (GUI)',
    featured: true,
    price: '$0',
    description:
      'Native graphical interface for macOS, Windows, and Linux. Built with Go and Fyne.',
    button: {
      label: 'Download Desktop',
      href: 'https://github.com/anilpdv/libgen-gui/releases',
    },
    features: [
      'Native macOS, Windows & Linux binaries',
      'Multi-mirror latency health prober',
      'Resumable background download queue',
      'Instant debounced multi-field search',
      'Zero trackers & zero telemetry',
    ],
    logomarkClassName: 'fill-cyan-500',
  },
  {
    name: 'Android App (SAF)',
    featured: false,
    price: '$0',
    description:
      'Native mobile experience with full Android Storage Access Framework (SAF) folder selection.',
    button: {
      label: 'Get Android APK',
      href: 'https://github.com/anilpdv/libgen-gui/releases',
    },
    features: [
      'DocumentFile Scoped Storage (SAF)',
      'Responsive touch-optimized UI',
      'Background network recovery',
      'Direct save to e-reader folders',
      '100% Free & Open Source',
    ],
    logomarkClassName: 'fill-gray-500',
  },
  {
    name: 'CLI & Go Package',
    featured: false,
    price: '$0',
    description:
      'Headless CLI tool and reusable Go SDK for script automation and terminal workflows.',
    button: {
      label: 'View Documentation',
      href: 'https://github.com/anilpdv/libgen-gui#cli--headless-usage',
    },
    features: [
      'Single static binary with zero dependencies',
      'JSON output for shell pipeline scripting',
      'Configurable concurrency & rate limits',
      'Importable Go package (`pkg/libgen`)',
      'Cross-compilation ready for ARM64/x86',
    ],
    logomarkClassName: 'fill-gray-300',
  },
]

function CheckIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true" {...props}>
      <path
        d="M9.307 12.248a.75.75 0 1 0-1.114 1.004l1.114-1.004ZM11 15.25l-.557.502a.75.75 0 0 0 1.15-.043L11 15.25Zm4.844-5.041a.75.75 0 0 0-1.188-.918l1.188.918Zm-7.651 3.043 2.25 2.5 1.114-1.004-2.25-2.5-1.114 1.004Zm3.4 2.457 4.25-5.5-1.187-.918-4.25 5.5 1.188.918Z"
        fill="currentColor"
      />
      <circle
        cx="12"
        cy="12"
        r="8.25"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  )
}

function Plan({
  name,
  price,
  description,
  button,
  features,
  logomarkClassName,
  featured = false,
}: {
  name: string
  price: string
  description: string
  button: {
    label: string
    href: string
  }
  features: Array<string>
  logomarkClassName?: string
  featured?: boolean
}) {
  return (
    <section
      className={clsx(
        'flex flex-col overflow-hidden rounded-3xl p-8 shadow-lg shadow-gray-900/5 transition-transform hover:-translate-y-1',
        featured ? 'bg-gray-900 ring-2 ring-cyan-500' : 'bg-white ring-1 ring-gray-200',
      )}
    >
      <h3
        className={clsx(
          'flex items-center text-base font-semibold',
          featured ? 'text-white' : 'text-gray-900',
        )}
      >
        <Logomark className={clsx('h-6 w-6 flex-none', logomarkClassName)} />
        <span className="ml-3">{name}</span>
      </h3>
      <p
        className={clsx(
          'relative mt-5 flex text-3xl font-bold tracking-tight',
          featured ? 'text-white' : 'text-gray-900',
        )}
      >
        {price}
        <span className={clsx('ml-2 text-sm font-normal self-end mb-1', featured ? 'text-gray-400' : 'text-gray-500')}>
          (100% Free & Open Source)
        </span>
      </p>
      <p
        className={clsx(
          'mt-3 text-sm leading-relaxed',
          featured ? 'text-gray-300' : 'text-gray-600',
        )}
      >
        {description}
      </p>
      <div className="order-last mt-8">
        <ul
          role="list"
          className={clsx(
            '-my-2 divide-y text-sm',
            featured
              ? 'divide-gray-800 text-gray-300'
              : 'divide-gray-100 text-gray-700',
          )}
        >
          {features.map((feature) => (
            <li key={feature} className="flex py-3">
              <CheckIcon
                className={clsx(
                  'h-5 w-5 flex-none',
                  featured ? 'text-cyan-400' : 'text-cyan-600',
                )}
              />
              <span className="ml-3">{feature}</span>
            </li>
          ))}
        </ul>
      </div>
      <Button
        href={button.href}
        color={featured ? 'cyan' : 'gray'}
        className="mt-8"
        aria-label={`Get started with ${name}`}
      >
        {button.label}
      </Button>
    </section>
  )
}

export function Pricing() {
  return (
    <section
      id="pricing"
      aria-labelledby="pricing-title"
      className="border-t border-gray-200 bg-gray-100 py-20 sm:py-32"
    >
      <Container>
        <div className="mx-auto max-w-2xl text-center">
          <h2
            id="pricing-title"
            className="text-3xl font-medium tracking-tight text-gray-900 sm:text-4xl"
          >
            Free forever. No subscriptions. No ads.
          </h2>
          <p className="mt-3 text-lg text-gray-600 leading-relaxed">
            LibGen GUI is distributed under the permissive MIT Open Source license. Choose the version that fits your workflow.
          </p>
        </div>

        <div className="mx-auto mt-16 grid max-w-2xl grid-cols-1 items-stretch gap-x-8 gap-y-10 sm:mt-20 lg:max-w-none lg:grid-cols-3">
          {plans.map((plan) => (
            <Plan key={plan.name} {...plan} />
          ))}
        </div>
      </Container>
    </section>
  )
}
