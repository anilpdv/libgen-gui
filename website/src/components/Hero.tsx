import Image from 'next/image'
import Link from 'next/link'
import { Button } from '@/components/Button'
import { Container } from '@/components/Container'
import desktopSearchImg from '@/images/screenshots/desktop-search.webp'

function AppleIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true" {...props}>
      <path d="M18.71 19.5c-.83 1.24-1.71 2.45-3.05 2.47-1.34.03-1.77-.79-3.29-.79-1.53 0-2 .77-3.27.82-1.31.05-2.3-1.32-3.14-2.53C4.25 17 2.94 12.45 4.7 9.39c.87-1.52 2.43-2.48 4.12-2.51 1.28-.02 2.5.87 3.29.87.78 0 2.26-1.07 3.81-.91.65.03 2.47.26 3.64 1.98-.09.06-2.17 1.28-2.15 3.81.03 3.02 2.65 4.03 2.68 4.04-.03.07-.42 1.44-1.38 2.83M15.97 6.37c.61-.75 1.04-1.8 1.01-2.87-.96.04-2.13.64-2.79 1.41-.58.68-1.1 1.76-.96 2.81 1.08.08 2.13-.56 2.74-1.35z" />
    </svg>
  )
}

function AndroidIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true" {...props}>
      <path d="M17.523 15.3414c-.5511 0-.9993-.4486-.9993-.9997s.4482-.9993.9993-.9993c.551 0 .9993.4482.9993.9993.0001.5511-.4483.9997-.9993.9997m-11.046 0c-.5511 0-.9993-.4486-.9993-.9997s.4482-.9993.9993-.9993c.5511 0 .9993.4482.9993.9993 0 .5511-.4482.9997-.9993.9997m11.4045-6.02l1.996-3.4567c.1086-.188.0441-.4285-.1438-.5371-.1879-.1085-.428-.0439-.5366.1439l-2.023 3.5042c-1.5037-.6873-3.1895-1.0746-4.9741-1.0746s-3.4704.3873-4.9741 1.0746L5.201 5.4716c-.1086-.1878-.3487-.2524-.5366-.1439-.1879.1086-.2524.3491-.1438.5371l1.996 3.4567C3.1258 11.233 1 14.5445 1 18.3975h22c0-3.853-2.1258-7.1645-5.5185-9.0761" />
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

export function Hero() {
  return (
    <section className="relative overflow-hidden pt-10 pb-16 sm:pt-14 sm:pb-24 lg:pt-16 lg:pb-28">
      <Container className="w-[min(100%-48px,1240px)] max-sm:w-[min(100%-32px,1240px)]">
        <div className="grid grid-cols-1 gap-12 lg:grid-cols-[minmax(420px,0.9fr)_minmax(560px,1.25fr)] lg:gap-x-14 lg:items-center">
          {/* Left Column: Copy & Actions */}
          <div className="relative z-10 mx-auto max-w-2xl lg:max-w-none">
            <div className="inline-flex items-center gap-x-2 rounded-full bg-cyan-50 px-3.5 py-1 text-xs font-semibold text-cyan-800 ring-1 ring-cyan-700/15 mb-6">
              <span className="h-1.5 w-1.5 rounded-full bg-cyan-600 animate-pulse" />
              <span>OPEN SOURCE · NATIVE · CROSS-PLATFORM</span>
            </div>

            <h1 className="max-w-[650px] text-[2rem] sm:text-5xl lg:text-[3.25rem] font-bold tracking-tight text-gray-900 leading-[1.12] lg:leading-[1.02]">
              A fast, native downloader for books and research papers.
            </h1>

            <p className="mt-5 text-base sm:text-[1.0625rem] text-gray-600 leading-[1.65] max-w-[590px]">
              Search multiple sources, recover automatically from failing mirrors, and manage downloads from one lightweight native application built with Go and Fyne.
            </p>

            {/* CTAs */}
            <div className="mt-8 flex flex-wrap items-center gap-3.5">
              <Button
                href="https://github.com/anilpdv/libgen-gui/releases/download/v2.0.1/LibGen.Downloader-v2.0.1-macos.zip"
                variant="solid"
                color="cyan"
                className="gap-2.5 shadow-sm min-h-[44px] px-5"
              >
                <AppleIcon className="h-4 w-4" />
                <span>Download macOS (v2.0.1)</span>
              </Button>

              <Button
                href="https://github.com/anilpdv/libgen-gui/releases/download/v2.0.1/LibGen-Downloader-Android-arm64.apk"
                variant="outline"
                color="gray"
                className="gap-2.5 min-h-[44px] px-5"
              >
                <AndroidIcon className="h-4 w-4 text-emerald-600" />
                <span>Download Android APK</span>
              </Button>

              <Link
                href="https://github.com/anilpdv/libgen-gui"
                className="min-h-[44px] inline-flex items-center gap-2 px-3 py-2 text-sm font-semibold text-gray-700 hover:text-gray-900 transition-colors"
                aria-label="View LibGen GUI repository on GitHub"
              >
                <GithubIcon className="h-4 w-4 text-gray-700" />
                <span>GitHub</span>
              </Link>
            </div>

            {/* Simplified Platform Support Status */}
            <div className="mt-10 pt-6 border-t border-gray-200/80">
              <div className="flex flex-col sm:flex-row sm:items-center gap-4 sm:gap-8">
                {/* Tested Releases */}
                <div>
                  <span className="text-[11px] font-bold uppercase tracking-wider text-gray-500 block mb-1.5">
                    Tested Releases
                  </span>
                  <div className="flex items-center gap-2">
                    <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-emerald-50 text-emerald-800 text-xs font-semibold ring-1 ring-emerald-600/20">
                      <span className="h-1.5 w-1.5 rounded-full bg-emerald-600" />
                      macOS
                    </span>
                    <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-emerald-50 text-emerald-800 text-xs font-semibold ring-1 ring-emerald-600/20">
                      <span className="h-1.5 w-1.5 rounded-full bg-emerald-600" />
                      Android
                    </span>
                  </div>
                </div>

                <div className="hidden sm:block h-8 w-px bg-gray-200" />

                {/* Build Support */}
                <div>
                  <span className="text-[11px] font-bold uppercase tracking-wider text-gray-500 block mb-1.5">
                    Build Support
                  </span>
                  <div className="flex items-center gap-2">
                    <span className="inline-flex items-center px-2 py-1 rounded-md bg-slate-100 text-slate-700 text-xs font-medium ring-1 ring-slate-400/20">
                      Linux
                    </span>
                    <span className="inline-flex items-center px-2 py-1 rounded-md bg-slate-100 text-slate-700 text-xs font-medium ring-1 ring-slate-400/20">
                      Windows
                    </span>
                    <span className="inline-flex items-center px-2 py-1 rounded-md bg-slate-100 text-slate-700 text-xs font-medium ring-1 ring-slate-400/20">
                      iOS Simulator
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Right Column: Dominant Authentic Desktop Application Screenshot */}
          <div className="relative mt-4 lg:mt-0">
            {/* Subtle Ambient Radial Glow */}
            <div className="absolute -inset-4 bg-cyan-500/12 rounded-3xl blur-2xl -z-10 pointer-events-none" />

            <figure className="relative m-0">
              <div className="overflow-hidden rounded-[14px] border border-slate-900/10 bg-[#0b1020] shadow-[0_28px_70px_rgba(15,23,42,0.16),0_8px_24px_rgba(15,23,42,0.1)]">
                <Image
                  src={desktopSearchImg}
                  alt="Actual LibGen GUI desktop interface showing live search results, format filters, and download queue controls"
                  priority
                  className="w-full h-auto block"
                />
              </div>
              <figcaption className="mt-3 text-center text-xs text-gray-500">
                Actual desktop application interface · macOS release shown
              </figcaption>
            </figure>
          </div>
        </div>
      </Container>
    </section>
  )
}
