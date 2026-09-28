import { AppStoreLink } from '@/components/AppStoreLink'
import { Button } from '@/components/Button'
import { CircleBackground } from '@/components/CircleBackground'
import { Container } from '@/components/Container'

export function CallToAction() {
  return (
    <section
      id="get-started-today"
      className="relative overflow-hidden bg-gray-900 py-20 sm:py-28"
    >
      <div className="absolute top-1/2 left-20 -translate-y-1/2 sm:left-1/2 sm:-translate-x-1/2">
        <CircleBackground color="#06B6D4" className="animate-spin-slower opacity-20" />
      </div>
      <Container className="relative">
        <div className="mx-auto max-w-2xl sm:text-center">
          <h2 className="text-3xl font-medium tracking-tight text-white sm:text-4xl">
            Start downloading research papers & books in seconds
          </h2>
          <p className="mt-4 text-lg text-gray-300 leading-relaxed">
            Free, open source, and available for macOS, Windows, Linux, and Android. No registration, no ads, no trackers.
          </p>
          <div className="mt-8 flex flex-wrap justify-center gap-4">
            <AppStoreLink color="white" />
            <Button
              href="https://github.com/anilpdv/libgen-gui"
              variant="outline"
              color="white"
            >
              View GitHub Repo
            </Button>
          </div>
        </div>
      </Container>
    </section>
  )
}
