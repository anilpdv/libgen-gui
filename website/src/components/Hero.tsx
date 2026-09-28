import Image from 'next/image'
import { Button } from '@/components/Button'
import { Container } from '@/components/Container'
import desktopSearchImg from '@/images/screenshots/desktop-search.webp'
import androidSearchImg from '@/images/screenshots/android-search.webp'

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

const platforms = [
  { name: 'macOS', spec: 'Apple Silicon + Intel', status: 'TESTED', badgeColor: 'bg-emerald-50 text-emerald-700 ring-emerald-600/20' },
  { name: 'Android', spec: 'Android 8.0+, API 26–35', status: 'TESTED', badgeColor: 'bg-emerald-50 text-emerald-700 ring-emerald-600/20' },
  { name: 'Linux', spec: 'Build from source', status: 'BUILD SUPPORT', badgeColor: 'bg-slate-100 text-slate-700 ring-slate-500/20' },
  { name: 'Windows', spec: 'Build from source', status: 'BUILD SUPPORT', badgeColor: 'bg-slate-100 text-slate-700 ring-slate-500/20' },
  { name: 'iOS', spec: 'Simulator support', status: 'BUILD SUPPORT', badgeColor: 'bg-slate-100 text-slate-700 ring-slate-500/20' },
]

export function Hero() {
  return (
    <div className="relative overflow-hidden pt-10 pb-16 sm:pt-14 sm:pb-24 lg:pb-28">
      <Container>
        <div className="lg:grid lg:grid-cols-12 lg:gap-x-10 lg:items-center">
          {/* Left Column: Copy & Actions */}
          <div className="relative z-10 mx-auto max-w-2xl lg:col-span-6 lg:max-w-none">
            <div className="inline-flex items-center gap-x-2 rounded-full bg-cyan-50 px-3.5 py-1 text-xs font-semibold text-cyan-800 ring-1 ring-cyan-700/15 mb-6">
              <span className="h-1.5 w-1.5 rounded-full bg-cyan-600 animate-pulse" />
              <span>OPEN SOURCE · NATIVE · CROSS-PLATFORM</span>
            </div>

            <h1 className="text-3xl font-bold tracking-tight text-gray-900 sm:text-5xl lg:text-[3.15rem] lg:leading-[1.12]">
              A fast, native downloader for books and research papers.
            </h1>

            <p className="mt-5 text-[15px] sm:text-base lg:text-[1.0625rem] text-gray-600 leading-[1.65] max-w-[580px]">
              Search multiple sources, recover automatically from failing mirrors, and manage downloads from one lightweight native application built with Go and Fyne.
            </p>

            <div className="mt-8 flex flex-wrap items-center gap-3.5">
              <Button
                href="https://github.com/anilpdv/libgen-gui/releases/download/v2.0.1/LibGen.Downloader-v2.0.1-macos.zip"
                variant="solid"
                color="cyan"
                className="gap-2.5 shadow-sm"
              >
                <AppleIcon className="h-4 w-4" />
                <span>Download macOS (v2.0.1)</span>
              </Button>

              <Button
                href="https://github.com/anilpdv/libgen-gui/releases/download/v2.0.1/LibGen-Downloader-Android-arm64.apk"
                variant="outline"
                color="gray"
                className="gap-2.5"
              >
                <AndroidIcon className="h-4 w-4 text-emerald-600" />
                <span>Get Android APK</span>
              </Button>

              <Button
                href="https://github.com/anilpdv/libgen-gui"
                variant="outline"
                color="gray"
                className="gap-2"
                aria-label="View LibGen GUI on GitHub"
              >
                <GithubIcon className="h-4 w-4 text-gray-700" />
                <span>GitHub</span>
              </Button>
            </div>

            {/* Platform Status Grid */}
            <div className="mt-10 pt-6 border-t border-gray-200/80">
              <div className="flex items-center justify-between mb-3">
                <p className="text-xs font-semibold uppercase tracking-wider text-gray-500">
                  Platform Verification & Support
                </p>
                <span className="text-[11px] text-gray-400 font-medium">v2.0.1 Release</span>
              </div>
              <div className="grid grid-cols-2 gap-2.5 sm:grid-cols-3 lg:grid-cols-5">
                {platforms.map((platform) => (
                  <div
                    key={platform.name}
                    className="rounded-xl border border-gray-200/80 bg-white p-3 shadow-2xs flex flex-col justify-between max-sm:last:col-span-2"
                  >
                    <div className="flex items-center justify-between gap-1">
                      <span className="text-xs font-bold text-gray-900">{platform.name}</span>
                      <span
                        className={`inline-block text-[9px] font-semibold px-1.5 py-0.5 rounded-sm ring-1 ${platform.badgeColor}`}
                      >
                        {platform.status}
                      </span>
                    </div>
                    <p className="text-[12px] text-gray-500 mt-1.5 leading-snug">
                      {platform.spec}
                    </p>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Right Column: Authentic Application Screenshots */}
          <div className="relative mt-12 lg:col-span-6 lg:mt-0">
            {/* Subtle Cyan Ambient Glow */}
            <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 h-[420px] w-[500px] rounded-full bg-cyan-500/10 blur-3xl pointer-events-none" />

            <div className="relative rounded-2xl border border-gray-200/90 bg-white p-2 shadow-2xl shadow-slate-900/10">
              <div className="flex items-center justify-between px-3 py-1.5 border-b border-gray-100 bg-gray-50/80 rounded-t-xl mb-1 text-[11px] text-gray-500">
                <div className="flex items-center gap-1.5">
                  <div className="w-2.5 h-2.5 rounded-full bg-red-400" />
                  <div className="w-2.5 h-2.5 rounded-full bg-amber-400" />
                  <div className="w-2.5 h-2.5 rounded-full bg-emerald-400" />
                </div>
                <span className="font-semibold text-gray-700">LibGen Downloader — Desktop UI</span>
                <span className="font-mono text-[10px] text-cyan-700 bg-cyan-50 px-2 py-0.5 rounded font-medium">
                  Mirrors: 2/2 Active
                </span>
              </div>

              {/* Real Desktop Application Screenshot */}
              <div className="relative overflow-hidden rounded-lg bg-slate-900">
                <Image
                  src={desktopSearchImg}
                  alt="LibGen GUI desktop application interface showing search results, file formats, and download destination bar"
                  priority
                  className="w-full h-auto object-cover rounded-lg"
                />
              </div>

              {/* Overlapping Authentic Android / Mobile Preview Badge */}
              <div className="absolute -bottom-6 -right-4 sm:-right-6 w-36 sm:w-44 rounded-xl border border-gray-200/90 bg-white p-1.5 shadow-xl shadow-slate-950/20 backdrop-blur-md">
                <div className="text-[10px] font-bold text-gray-800 px-1.5 py-0.5 flex items-center justify-between border-b border-gray-100 mb-1">
                  <span>Android Mobile UI</span>
                  <span className="text-[8px] bg-emerald-100 text-emerald-800 px-1 rounded font-medium">SAF</span>
                </div>
                <Image
                  src={androidSearchImg}
                  alt="LibGen GUI Android mobile interface with scoped storage and touch controls"
                  className="w-full h-auto rounded-lg"
                />
              </div>
            </div>

            {/* Captions */}
            <div className="mt-8 flex items-center justify-between text-xs text-gray-500 px-1">
              <div className="flex items-center gap-1.5">
                <span className="h-2 w-2 rounded-full bg-cyan-600" />
                <span className="font-medium text-gray-700">Real Application UI</span>
                <span className="text-gray-400">— Desktop Search & Download Queue</span>
              </div>
              <span className="text-gray-400 hidden sm:inline">No Electron · Pure Go + Fyne</span>
            </div>
          </div>
        </div>
      </Container>
    </div>
  )
}
