import { Button } from '@/components/Button'
import { Container } from '@/components/Container'

export function CallToAction() {
  return (
    <section
      id="get-started"
      aria-label="Download Call to Action"
      className="bg-gray-900 py-16 sm:py-20 text-white relative overflow-hidden"
    >
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[300px] bg-cyan-600/10 rounded-full blur-3xl pointer-events-none" />

      <Container className="relative">
        <div className="mx-auto max-w-2xl text-center">
          <h2 className="text-3xl font-bold tracking-tight text-white sm:text-4xl">
            Ready to try LibGen GUI?
          </h2>
          <p className="mt-4 text-base text-gray-300 leading-relaxed sm:text-lg">
            Download a tested macOS or Android release, or inspect the source before installing.
          </p>
          <div className="mt-8 flex flex-wrap justify-center items-center gap-4">
            <Button
              href="https://github.com/anilpdv/libgen-gui/releases"
              variant="solid"
              color="cyan"
              className="gap-2 px-6"
            >
              <span>View Releases</span>
            </Button>
            <Button
              href="https://github.com/anilpdv/libgen-gui"
              variant="outline"
              color="white"
              className="gap-2 px-6"
            >
              <span>View Source</span>
            </Button>
          </div>
        </div>
      </Container>
    </section>
  )
}
