'use client'

import { useState } from 'react'
import clsx from 'clsx'

import { AppScreen } from '@/components/AppScreen'
import { Container } from '@/components/Container'
import { PhoneFrame } from '@/components/PhoneFrame'

const features = [
  {
    name: 'Multi-Mirror Failover & Live Prober',
    description:
      'Continuous background latency probing and automatic failover across official mirrors ensure your searches and downloads never stall or drop packets.',
    icon: MirrorProbeIcon,
    badge: 'Zero Dead Mirrors',
    screen: MirrorScreen,
  },
  {
    name: 'Resumable Queue & Range Requests',
    description:
      'A robust serialized FIFO queue with pause, resume, cancel, and HTTP Range support. Interrupted downloads seamlessly pick up right where they left off.',
    icon: QueueDownloadIcon,
    badge: 'Atomic .part Files',
    screen: QueueScreen,
  },
  {
    name: 'Instant Multi-Field Book Search',
    description:
      'Debounced search across title, author, series, publisher, ISBN, and MD5 with live extension and size badges, direct 1-click downloads, and metadata inspection.',
    icon: SearchBookIcon,
    badge: 'Sub-50ms Response',
    screen: SearchScreen,
  },
]

function MirrorProbeIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 24 24" fill="none" aria-hidden="true" {...props}>
      <circle cx={12} cy={12} r={9} stroke="#22d3ee" strokeWidth="2" />
      <circle cx={12} cy={12} r={3} fill="#22d3ee" />
      <path
        d="M12 3v3M12 18v3M3 12h3M18 12h3"
        stroke="#22d3ee"
        strokeWidth="2"
        strokeLinecap="round"
      />
    </svg>
  )
}

function QueueDownloadIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 24 24" fill="none" aria-hidden="true" {...props}>
      <path
        d="M8 12l4 4 4-4M12 4v12M5 20h14"
        stroke="#22d3ee"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  )
}

function SearchBookIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 24 24" fill="none" aria-hidden="true" {...props}>
      <path
        d="M11 19a8 8 0 1 0 0-16 8 8 0 0 0 0 16zm6-2l4 4"
        stroke="#22d3ee"
        strokeWidth="2"
        strokeLinecap="round"
      />
    </svg>
  )
}

function MirrorScreen() {
  return (
    <AppScreen className="w-full">
      <AppScreen.Header>
        <AppScreen.Title>Mirror Health</AppScreen.Title>
        <AppScreen.Subtitle>
          <span className="text-emerald-400 font-semibold">3 of 3 Active</span> • 115ms avg latency
        </AppScreen.Subtitle>
      </AppScreen.Header>
      <AppScreen.Body>
        <div className="divide-y divide-gray-100">
          {[
            { url: 'https://libgen.is', latency: '82 ms', status: 'Optimal (Primary)', color: 'bg-emerald-500' },
            { url: 'https://libgen.rs', latency: '115 ms', status: 'Healthy (Failover 1)', color: 'bg-emerald-500' },
            { url: 'https://libgen.st', latency: '178 ms', status: 'Healthy (Failover 2)', color: 'bg-emerald-500' },
            { url: 'https://libgen.li', latency: '240 ms', status: 'Active (Fallback)', color: 'bg-cyan-500' },
          ].map((mirror) => (
            <div key={mirror.url} className="flex items-center gap-3 px-4 py-3">
              <span className={`inline-block h-2 w-2 rounded-full ${mirror.color}`} />
              <div className="flex-auto min-w-0">
                <div className="text-xs font-semibold text-gray-900 truncate">{mirror.url}</div>
                <div className="text-[10px] text-gray-500">{mirror.status}</div>
              </div>
              <div className="flex-none text-right">
                <div className="text-xs font-mono font-medium text-gray-900">{mirror.latency}</div>
                <div className="text-[10px] text-emerald-600 font-medium">99.9%</div>
              </div>
            </div>
          ))}
        </div>
        <div className="p-3 bg-gray-50 border-t border-gray-100 flex items-center justify-between text-[11px] text-gray-600">
          <span>Background Auto-Probe</span>
          <span className="font-semibold text-cyan-600">Every 30s</span>
        </div>
      </AppScreen.Body>
    </AppScreen>
  )
}

function QueueScreen() {
  return (
    <AppScreen className="w-full">
      <AppScreen.Header>
        <AppScreen.Title>Download Queue</AppScreen.Title>
        <AppScreen.Subtitle>
          <span className="text-cyan-400 font-semibold">2 Downloading</span> • 1 Queued
        </AppScreen.Subtitle>
      </AppScreen.Header>
      <AppScreen.Body>
        <div className="p-3.5 space-y-3">
          <div className="rounded-xl border border-gray-100 bg-white p-3 shadow-2xs">
            <div className="flex justify-between items-start">
              <div>
                <h4 className="text-xs font-semibold text-gray-900 line-clamp-1">
                  Computer Systems: A Programmer’s Perspective
                </h4>
                <p className="text-[10px] text-gray-500 mt-0.5">PDF • 14.8 MB / 19.4 MB</p>
              </div>
              <span className="text-[10px] font-semibold text-cyan-700 bg-cyan-50 px-2 py-0.5 rounded-full">
                76%
              </span>
            </div>
            <div className="mt-2 h-1.5 w-full overflow-hidden rounded-full bg-gray-100">
              <div className="h-full bg-cyan-500 rounded-full" style={{ width: '76%' }} />
            </div>
            <div className="mt-1.5 flex justify-between text-[9px] text-gray-500">
              <span>3.4 MB/s</span>
              <span>Chunk 4/5 • Resumable</span>
            </div>
          </div>

          <div className="rounded-xl border border-gray-100 bg-white p-3 shadow-2xs">
            <div className="flex justify-between items-start">
              <div>
                <h4 className="text-xs font-semibold text-gray-900 line-clamp-1">
                  Designing Data-Intensive Applications
                </h4>
                <p className="text-[10px] text-gray-500 mt-0.5">EPUB • 8.2 MB / 8.2 MB</p>
              </div>
              <span className="text-[10px] font-semibold text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded-full">
                Completed
              </span>
            </div>
            <div className="mt-2 h-1.5 w-full overflow-hidden rounded-full bg-emerald-100">
              <div className="h-full bg-emerald-500 rounded-full w-full" />
            </div>
            <div className="mt-1.5 flex justify-between text-[9px] text-gray-500">
              <span>Saved to ~/Documents/LibgenBooks</span>
              <span>MD5 Verified</span>
            </div>
          </div>

          <div className="rounded-xl border border-gray-100 bg-gray-50 p-2.5 flex justify-between items-center">
            <span className="text-xs font-medium text-gray-700 line-clamp-1">
              Structure and Interpretation of Computer Programs
            </span>
            <span className="text-[10px] font-medium text-gray-500 bg-gray-200/80 px-2 py-0.5 rounded-full shrink-0 ml-2">
              Queued
            </span>
          </div>
        </div>
      </AppScreen.Body>
    </AppScreen>
  )
}

function SearchScreen() {
  return (
    <AppScreen className="w-full">
      <AppScreen.Header>
        <AppScreen.Title>Book Search</AppScreen.Title>
        <AppScreen.Subtitle>Query: &quot;Distributed Systems&quot;</AppScreen.Subtitle>
      </AppScreen.Header>
      <AppScreen.Body>
        <div className="p-3">
          <div className="rounded-lg bg-gray-100 px-3 py-1.5 text-xs text-gray-700 flex items-center gap-2 mb-2.5">
            <svg viewBox="0 0 20 20" fill="currentColor" className="w-3.5 h-3.5 text-gray-400">
              <path fillRule="evenodd" d="M9 3.5a5.5 5.5 0 100 11 5.5 5.5 0 000-11zM2 9a7 7 0 1112.452 4.391l3.328 3.329a.75.75 0 11-1.06 1.06l-3.329-3.328A7 7 0 012 9z" clipRule="evenodd" />
            </svg>
            <span className="text-gray-900 font-medium">Distributed Systems</span>
          </div>

          <div className="space-y-2">
            {[
              { title: 'Distributed Systems: Principles & Paradigms', author: 'Tanenbaum, van Steen', year: '2023', format: 'PDF', size: '18.4 MB' },
              { title: 'Understanding Distributed Systems, 2nd Edition', author: 'Roberto Vitillo', year: '2022', format: 'EPUB', size: '5.1 MB' },
              { title: 'Database Internals: Distributed Data Systems', author: 'Alex Petrov', year: '2019', format: 'PDF', size: '14.2 MB' },
            ].map((book) => (
              <div key={book.title} className="rounded-lg border border-gray-100 p-2.5 bg-white">
                <h5 className="text-xs font-semibold text-gray-900 line-clamp-1">{book.title}</h5>
                <p className="text-[10px] text-gray-500 mt-0.5">{book.author}</p>
                <div className="mt-1.5 flex items-center justify-between">
                  <div className="flex gap-1 text-[9px]">
                    <span className="bg-gray-100 px-1.5 py-0.5 rounded font-mono font-medium">{book.format}</span>
                    <span className="bg-gray-100 px-1.5 py-0.5 rounded text-gray-600">{book.size}</span>
                    <span className="bg-gray-100 px-1.5 py-0.5 rounded text-gray-600">{book.year}</span>
                  </div>
                  <span className="bg-cyan-600 text-white text-[9px] font-medium px-2 py-0.5 rounded">
                    Download
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>
      </AppScreen.Body>
    </AppScreen>
  )
}

export function PrimaryFeatures() {
  let [selectedFeature, setSelectedFeature] = useState(0)
  let ActiveScreen = features[selectedFeature].screen

  return (
    <section
      id="features"
      aria-label="Engineered for Reliability"
      className="bg-gray-950 py-20 sm:py-28 text-white relative overflow-hidden"
    >
      <Container>
        <div className="max-w-2xl">
          <div className="inline-flex items-center gap-2 rounded-full bg-cyan-950/80 px-3.5 py-1 text-xs font-semibold text-cyan-400 ring-1 ring-cyan-500/30 mb-4">
            <span>RESILIENT INFRASTRUCTURE</span>
          </div>
          <h2 className="text-3xl font-bold tracking-tight text-white sm:text-4xl">
            Engineered for reliability, speed, and clean code.
          </h2>
          <p className="mt-4 text-base text-gray-300 leading-relaxed sm:text-lg">
            Unlike slow browser scraping or heavyweight wrappers, LibGen GUI is compiled into a lightweight native binary with production-grade networking, resilient mirror rotation, and thread-safe queue management.
          </p>
        </div>

        <div className="mt-14 grid grid-cols-1 gap-12 lg:grid-cols-12 lg:items-center">
          {/* 3 Consistent Cards */}
          <div className="lg:col-span-7 space-y-4">
            {features.map((feature, idx) => {
              let isSelected = selectedFeature === idx
              let Icon = feature.icon
              return (
                <button
                  key={feature.name}
                  type="button"
                  onClick={() => setSelectedFeature(idx)}
                  className={clsx(
                    'w-full text-left rounded-2xl p-6 transition-all duration-200 border cursor-pointer',
                    isSelected
                      ? 'bg-gray-900 border-cyan-500/80 ring-1 ring-cyan-500/40 shadow-lg shadow-cyan-950/40'
                      : 'bg-gray-900/50 border-gray-800/80 hover:bg-gray-900/80 hover:border-gray-700 text-gray-300',
                  )}
                >
                  <div className="flex items-start gap-4">
                    <div className={clsx('rounded-xl p-2.5 shrink-0', isSelected ? 'bg-cyan-500/20 text-cyan-400' : 'bg-gray-800 text-gray-400')}>
                      <Icon className="h-6 w-6" />
                    </div>
                    <div className="flex-auto">
                      <div className="flex items-center justify-between gap-2">
                        <h3 className="text-base font-semibold text-white sm:text-lg">
                          {feature.name}
                        </h3>
                        <span className="text-[11px] font-semibold text-cyan-400 bg-cyan-950/80 px-2 py-0.5 rounded-full border border-cyan-800/50 shrink-0">
                          {feature.badge}
                        </span>
                      </div>
                      <p className="mt-2 text-sm text-gray-300 leading-relaxed">
                        {feature.description}
                      </p>
                    </div>
                  </div>
                </button>
              )
            })}
          </div>

          {/* Device Screen Preview */}
          <div className="lg:col-span-5 flex justify-center">
            <div className="relative w-full max-w-[340px] drop-shadow-2xl">
              <div className="absolute inset-0 bg-cyan-500/15 rounded-3xl blur-2xl pointer-events-none" />
              <PhoneFrame>
                <ActiveScreen />
              </PhoneFrame>
            </div>
          </div>
        </div>
      </Container>
    </section>
  )
}
