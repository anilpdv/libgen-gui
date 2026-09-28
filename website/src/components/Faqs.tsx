'use client'

import { useState } from 'react'
import { Container } from '@/components/Container'

const faqs = [
  {
    question: 'How does multi-mirror search and failover work?',
    answer:
      'LibGen GUI continuously probes configured mirrors (such as libgen.is, libgen.rs, libgen.st, and libgen.li) in the background with adaptive latency measurements. If an active mirror responds with an error, times out, or gets DNS blocked, the client automatically re-routes subsequent searches and downloads to the fastest available healthy mirror without user intervention.',
  },
  {
    question: 'What file formats and metadata fields are supported?',
    answer:
      'The client parses and displays all major academic formats including PDF, EPUB, MOBI, AZW3, DJVU, and CBZ. You can filter queries by format and sort live results by Title, Author, Year, Format, or File Size. Search operates across title, author, series, publisher, ISBN-10/13, and MD5 hashes.',
  },
  {
    question: 'Are downloads resumable if connection drops?',
    answer:
      'Yes. LibGen GUI isolates in-flight downloads into temporary `.part` files and utilizes HTTP Range request headers. If your connection is interrupted or you pause the transfer, resuming picks up from the exact byte offset rather than downloading from the beginning. Upon completion, files are atomically renamed to their final path.',
  },
  {
    question: 'How does Android Storage Access Framework (SAF) work?',
    answer:
      'On Android 10 through 15, Google enforces Scoped Storage which restricts standard POSIX file writes. LibGen GUI integrates directly with Android native DocumentFile and Storage Access Framework (SAF) APIs. You select a target folder once (such as your Books or e-reader directory), and downloads stream directly into that folder with verified write permissions.',
  },
  {
    question: 'Why build with Go and Fyne instead of Electron?',
    answer:
      'Electron bundles an entire Chromium browser and Node.js runtime, typically consuming 300MB–600MB of RAM for a basic utility. LibGen GUI is compiled into a single compact native binary using Go and the Fyne vector toolkit. It starts in under 200ms, consumes less than 35MB of RAM, and uses native GPU rendering on desktop and mobile.',
  },
  {
    question: 'Where are downloads stored and how do I change the directory?',
    answer:
      'On desktop, downloads default to your standard system Downloads/LibgenBooks folder. You can customize this anytime via the footer [Change] button or by opening Settings → Downloads. Each queued download captures its destination at enqueue time, allowing active downloads to finish undisturbed while new tasks use your updated destination.',
  },
  {
    question: 'How does queue persistence handle application crashes?',
    answer:
      'The FIFO download manager writes its task ledger atomically to a local `.queue.json` file. If the application is closed or the OS restarts, incomplete tasks are restored in their proper order on the next startup with destination metadata intact.',
  },
  {
    question: 'Is LibGen GUI free and open source?',
    answer:
      'Yes. LibGen GUI is licensed under the permissive MIT Open Source license. The complete source code is hosted publicly on GitHub with zero trackers, zero advertisements, and zero telemetry.',
  },
]

function ChevronDownIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 20 20" fill="currentColor" aria-hidden="true" {...props}>
      <path
        fillRule="evenodd"
        d="M5.23 7.21a.75.75 0 011.06.02L10 11.168l3.71-3.938a.75.75 0 111.08 1.04l-4.25 4.5a.75.75 0 01-1.08 0l-4.25-4.5a.75.75 0 01.02-1.06z"
        clipRule="evenodd"
      />
    </svg>
  )
}

export function Faqs() {
  let [openIndex, setOpenIndex] = useState<number | null>(0)

  let toggle = (index: number) => {
    setOpenIndex((current) => (current === index ? null : index))
  }

  return (
    <section
      id="faqs"
      aria-labelledby="faqs-title"
      className="border-t border-gray-200/80 bg-white py-20 sm:py-28"
    >
      <Container>
        <div className="mx-auto max-w-2xl text-center">
          <div className="inline-flex items-center gap-2 rounded-full bg-cyan-50 px-3.5 py-1 text-xs font-semibold text-cyan-800 ring-1 ring-cyan-700/15 mb-4">
            <span>FREQUENTLY ASKED QUESTIONS</span>
          </div>
          <h2
            id="faqs-title"
            className="text-3xl font-bold tracking-tight text-gray-900 sm:text-4xl"
          >
            Everything you need to know.
          </h2>
          <p className="mt-4 text-base text-gray-600 leading-relaxed sm:text-lg">
            Have questions about architecture, mirrors, or storage? Here are the most common answers.
          </p>
        </div>

        <div className="mx-auto mt-14 max-w-[850px] divide-y divide-gray-200/80 rounded-2xl border border-gray-200/90 bg-white shadow-2xs">
          {faqs.map((faq, index) => {
            let isOpen = openIndex === index
            let answerId = `faq-answer-${index}`
            let buttonId = `faq-button-${index}`

            return (
              <div key={faq.question} className="px-6 py-5 sm:px-8">
                <button
                  id={buttonId}
                  type="button"
                  onClick={() => toggle(index)}
                  aria-expanded={isOpen}
                  aria-controls={answerId}
                  className="flex w-full items-center justify-between text-left gap-4 font-semibold text-gray-900 hover:text-cyan-700 transition-colors cursor-pointer"
                >
                  <span className="text-base sm:text-lg">{faq.question}</span>
                  <ChevronDownIcon
                    className={`h-5 w-5 shrink-0 text-gray-500 transition-transform duration-200 ${
                      isOpen ? 'rotate-180 text-cyan-600' : ''
                    }`}
                  />
                </button>
                {isOpen && (
                  <div id={answerId} role="region" aria-labelledby={buttonId} className="mt-3.5 pr-6">
                    <p className="text-sm text-gray-600 leading-relaxed sm:text-base">
                      {faq.answer}
                    </p>
                  </div>
                )}
              </div>
            )
          })}
        </div>
      </Container>
    </section>
  )
}
