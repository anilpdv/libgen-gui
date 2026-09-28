import { useId } from 'react'
import clsx from 'clsx'

import { AppDemo } from '@/components/AppDemo'
import { AppStoreLink } from '@/components/AppStoreLink'
import { Button } from '@/components/Button'
import { Container } from '@/components/Container'
import { PhoneFrame } from '@/components/PhoneFrame'

function BackgroundIllustration(props: React.ComponentPropsWithoutRef<'div'>) {
  let id = useId()

  return (
    <div {...props}>
      <svg
        viewBox="0 0 1026 1026"
        fill="none"
        aria-hidden="true"
        className="absolute inset-0 h-full w-full animate-spin-slow"
      >
        <path
          d="M1025 513c0 282.77-229.23 512-512 512S1 795.77 1 513 230.23 1 513 1s512 229.23 512 512Z"
          stroke="#D4D4D4"
          strokeOpacity="0.7"
        />
        <path
          d="M513 1025C230.23 1025 1 795.77 1 513"
          stroke={`url(#${id}-gradient-1)`}
          strokeLinecap="round"
        />
        <defs>
          <linearGradient
            id={`${id}-gradient-1`}
            x1="1"
            y1="513"
            x2="1"
            y2="1025"
            gradientUnits="userSpaceOnUse"
          >
            <stop stopColor="#06b6d4" />
            <stop offset="1" stopColor="#06b6d4" stopOpacity="0" />
          </linearGradient>
        </defs>
      </svg>
      <svg
        viewBox="0 0 1026 1026"
        fill="none"
        aria-hidden="true"
        className="absolute inset-0 h-full w-full animate-spin-reverse-slower"
      >
        <path
          d="M913 513c0 220.914-179.086 400-400 400S113 733.914 113 513s179.086-400 400-400 400 179.086 400 400Z"
          stroke="#D4D4D4"
          strokeOpacity="0.7"
        />
        <path
          d="M913 513c0 220.914-179.086 400-400 400"
          stroke={`url(#${id}-gradient-2)`}
          strokeLinecap="round"
        />
        <defs>
          <linearGradient
            id={`${id}-gradient-2`}
            x1="913"
            y1="513"
            x2="913"
            y2="913"
            gradientUnits="userSpaceOnUse"
          >
            <stop stopColor="#06b6d4" />
            <stop offset="1" stopColor="#06b6d4" stopOpacity="0" />
          </linearGradient>
        </defs>
      </svg>
    </div>
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

const platforms = [
  { name: 'macOS (Apple Silicon & Intel)', tag: 'v2.0.0 Native App' },
  { name: 'Windows 10/11', tag: 'Native x64' },
  { name: 'Linux (Debian / Arch / RPM)', tag: 'GTK / X11 / Wayland' },
  { name: 'Android 10 - 15', tag: 'SAF Tree Access' },
]

export function Hero() {
  return (
    <div className="overflow-hidden py-20 sm:py-32 lg:pb-32 xl:pb-36">
      <Container>
        <div className="lg:grid lg:grid-cols-12 lg:gap-x-8 lg:gap-y-20">
          <div className="relative z-10 mx-auto max-w-2xl lg:col-span-7 lg:max-w-none lg:pt-6 xl:col-span-6">
            <div className="inline-flex items-center gap-x-2 rounded-full bg-cyan-50 px-3.5 py-1 text-xs font-semibold text-cyan-700 ring-1 ring-cyan-700/10 mb-6">
              <span className="h-1.5 w-1.5 rounded-full bg-cyan-500 animate-pulse" />
              v2.0.0 Architecture Released
            </div>
            <h1 className="text-4xl font-semibold tracking-tight text-gray-900 sm:text-5xl">
              High-speed book and paper downloader. Zero bloat.
            </h1>
            <p className="mt-6 text-lg text-gray-600 leading-relaxed">
              Fast, native, multi-mirror client built with Go and Fyne. Features live mirror health probing, automatic multi-mirror failover, resumable HTTP range downloads, and full Android Storage Access Framework (SAF) integration.
            </p>
            <div className="mt-8 flex flex-wrap gap-x-6 gap-y-4 items-center">
              <AppStoreLink />
              <Button
                href="https://github.com/anilpdv/libgen-gui"
                variant="outline"
                className="group"
              >
                <GithubIcon className="h-5 w-5 flex-none text-gray-700 group-hover:text-gray-900" />
                <span className="ml-2.5">Star on GitHub</span>
              </Button>
            </div>
          </div>
          <div className="relative mt-10 sm:mt-20 lg:col-span-5 lg:row-span-2 lg:mt-0 xl:col-span-6">
            <BackgroundIllustration className="absolute top-4 left-1/2 h-[1026px] w-[1026px] -translate-x-1/3 mask-[linear-gradient(to_bottom,white_20%,transparent_75%)] stroke-gray-300/70 sm:top-16 sm:-translate-x-1/2 lg:-top-16 lg:ml-12 xl:-top-14 xl:ml-0" />
            <div className="-mx-4 h-[448px] mask-[linear-gradient(to_bottom,white_60%,transparent)] px-9 sm:mx-0 lg:absolute lg:-inset-x-10 lg:-top-10 lg:-bottom-20 lg:h-auto lg:px-0 lg:pt-10 xl:-bottom-32">
              <PhoneFrame className="mx-auto max-w-[366px]" priority>
                <AppDemo />
              </PhoneFrame>
            </div>
          </div>
          <div className="relative -mt-4 lg:col-span-7 lg:mt-0 xl:col-span-6">
            <p className="text-center text-xs font-semibold uppercase tracking-wider text-gray-500 lg:text-left">
              Native support across major platforms
            </p>
            <ul
              role="list"
              className="mx-auto mt-6 grid grid-cols-2 gap-4 sm:grid-cols-4 lg:mx-0"
            >
              {platforms.map((platform) => (
                <li
                  key={platform.name}
                  className="flex flex-col justify-center rounded-xl bg-gray-50/80 p-3 ring-1 ring-gray-900/5 text-center sm:text-left"
                >
                  <span className="text-xs font-medium text-gray-900 truncate">
                    {platform.name}
                  </span>
                  <span className="text-[10px] text-cyan-600 font-semibold mt-0.5">
                    {platform.tag}
                  </span>
                </li>
              ))}
            </ul>
          </div>
        </div>
      </Container>
    </div>
  )
}
