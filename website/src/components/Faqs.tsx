import { Container } from '@/components/Container'

const faqs = [
  [
    {
      question: 'How does multi-mirror failover work?',
      answer:
        'LibGen GUI probes configured official mirrors (e.g. libgen.is, libgen.rs, libgen.st) in the background with adaptive latency measurements. If an active mirror responds with an error or times out, the client automatically re-routes the request to the fastest available healthy mirror.',
    },
    {
      question: 'What formats are supported?',
      answer:
        'All major publication formats including PDF, EPUB, MOBI, AZW3, DJVU, and CBZ. You can filter search results directly by format and sort by year or file size.',
    },
    {
      question: 'Are downloads resumable?',
      answer:
        'Yes. LibGen GUI uses HTTP Range requests and temporary .part files. If your network connection drops or you pause the download, it resumes from the exact byte position without restarting.',
    },
  ],
  [
    {
      question: 'How does it handle Android permissions?',
      answer:
        'On Android 10 through 15, LibGen GUI uses the native Storage Access Framework (SAF) and DocumentFile APIs. You select a destination folder once (such as your e-reader or Books folder), and downloads stream directly into that directory without requiring broad storage permissions.',
    },
    {
      question: 'Why build with Go and Fyne instead of Electron?',
      answer:
        'Electron bundles an entire Chromium browser and Node.js runtime, taking hundreds of megabytes of RAM. Go + Fyne compiles into a single compact native binary that starts in under 200ms and runs smoothly even on low-spec laptops and mobile devices.',
    },
    {
      question: 'Is it completely free and open source?',
      answer:
        'Yes! LibGen GUI is licensed under the MIT License. The code is 100% open source on GitHub with zero trackers, zero ads, and zero telemetry.',
    },
  ],
  [
    {
      question: 'What is the queue persistence mechanism?',
      answer:
        'The download queue writes its state to an atomic ledger on disk. If the app closes unexpectedly or your computer restarts, your incomplete downloads and queue order are restored on launch.',
    },
    {
      question: 'How can I build it from source?',
      answer:
        'Make sure you have Go installed, clone the repository, and run `go build -o "LibGen Downloader" ./cmd/libgen-gui`. You can also package it for macOS, Windows, Linux, or Android using `fyne package`.',
    },
    {
      question: 'Legal & Fair Access Notice',
      answer:
        'LibGen GUI is a search client tool. Users are responsible for ensuring that their downloads comply with all applicable copyright and intellectual property laws in their respective jurisdictions.',
    },
  ],
]

export function Faqs() {
  return (
    <section
      id="faqs"
      aria-labelledby="faqs-title"
      className="border-t border-gray-200 py-20 sm:py-32"
    >
      <Container>
        <div className="mx-auto max-w-2xl lg:mx-0">
          <h2
            id="faqs-title"
            className="text-3xl font-medium tracking-tight text-gray-900 sm:text-4xl"
          >
            Frequently asked questions
          </h2>
          <p className="mt-3 text-lg text-gray-600">
            Have questions about LibGen GUI? Here are the most common answers. For more help, explore our{' '}
            <a
              href="https://github.com/anilpdv/libgen-gui"
              className="text-cyan-600 font-medium underline hover:text-cyan-700"
            >
              GitHub repository
            </a>
            .
          </p>
        </div>
        <ul
          role="list"
          className="mx-auto mt-16 grid max-w-2xl grid-cols-1 gap-8 sm:mt-20 lg:max-w-none lg:grid-cols-3"
        >
          {faqs.map((column, columnIndex) => (
            <li key={columnIndex}>
              <ul role="list" className="space-y-10">
                {column.map((faq, faqIndex) => (
                  <li key={faqIndex}>
                    <h3 className="text-base font-semibold text-gray-900">
                      {faq.question}
                    </h3>
                    <p className="mt-3 text-sm text-gray-600 leading-relaxed">{faq.answer}</p>
                  </li>
                ))}
              </ul>
            </li>
          ))}
        </ul>
      </Container>
    </section>
  )
}
