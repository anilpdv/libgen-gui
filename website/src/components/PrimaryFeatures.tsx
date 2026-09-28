'use client'

import { useState } from 'react'
import Image from 'next/image'
import clsx from 'clsx'

import { Container } from '@/components/Container'
import mirrorHealthImg from '@/images/screenshots/mirror-health.webp'
import desktopSearchImg from '@/images/screenshots/desktop-search.webp'

const features = [
  {
    name: 'Multi-Mirror Health Checking',
    category: 'RESILIENCE',
    description:
      'Periodically measures endpoint availability and round-trip latency across official mirrors, exposing real-time health indicators without manual intervention.',
    icon: MirrorProbeIcon,
    badge: 'Live Latency Prober',
    panel: 'mirror',
  },
  {
    name: 'Automatic Failover & Retry Policy',
    category: 'NETWORKING',
    description:
      'Seamlessly switches to another healthy mirror when an active endpoint times out or drops packets, adhering to strict user-configurable retry limits.',
    icon: FailoverIcon,
    badge: 'Zero Dead Mirrors',
    panel: 'failover',
  },
  {
    name: 'Serialized Resumable Queue',
    category: 'DOWNLOADS',
    description:
      'Organizes active and pending downloads with HTTP Range byte-level resumption, atomic .part safety, and graceful cancellation controls.',
    icon: QueueDownloadIcon,
    badge: 'Atomic .part Safety',
    panel: 'queue',
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

function FailoverIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 24 24" fill="none" aria-hidden="true" {...props}>
      <path
        d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
        stroke="#22d3ee"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
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

export function PrimaryFeatures() {
  let [selectedFeature, setSelectedFeature] = useState(0)
  let activeFeature = features[selectedFeature]

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

        <div className="mt-14 grid grid-cols-1 gap-10 lg:grid-cols-12 lg:items-center">
          {/* Feature Selector Cards */}
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
                    <div
                      className={clsx(
                        'rounded-xl p-2.5 shrink-0',
                        isSelected ? 'bg-cyan-500/20 text-cyan-400' : 'bg-gray-800 text-gray-400',
                      )}
                    >
                      <Icon className="h-6 w-6" />
                    </div>
                    <div className="flex-auto">
                      <div className="flex items-center justify-between gap-2">
                        <div className="flex items-center gap-2">
                          <span className="text-[10px] font-bold text-cyan-400 tracking-wider uppercase">
                            {feature.category}
                          </span>
                          <span className="text-gray-600">•</span>
                          <h3 className="text-base font-semibold text-white sm:text-lg">
                            {feature.name}
                          </h3>
                        </div>
                        <span className="text-[11px] font-semibold text-cyan-400 bg-cyan-950/80 px-2.5 py-0.5 rounded-full border border-cyan-800/50 shrink-0">
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

          {/* Authentic Feature Preview Container */}
          <div className="lg:col-span-5 flex justify-center">
            <div className="relative w-full max-w-[460px] rounded-2xl border border-gray-800 bg-gray-900/90 p-4 shadow-2xl shadow-cyan-950/30 backdrop-blur-md">
              {/* Glow */}
              <div className="absolute inset-0 bg-cyan-500/10 rounded-2xl blur-xl pointer-events-none" />

              <div className="relative z-10">
                <div className="flex items-center justify-between pb-3 mb-3 border-b border-gray-800 text-xs">
                  <div className="flex items-center gap-2">
                    <span className="h-2 w-2 rounded-full bg-emerald-400 animate-pulse" />
                    <span className="font-semibold text-gray-200">
                      {selectedFeature === 0
                        ? 'MIRROR HEALTH & PROBER'
                        : selectedFeature === 1
                        ? 'FAILOVER SIMULATION'
                        : 'SERIALIZED QUEUE STATE'}
                    </span>
                  </div>
                  <span className="text-[11px] font-mono text-cyan-400 bg-cyan-950/80 px-2 py-0.5 rounded border border-cyan-800/40">
                    Real Native Dialog
                  </span>
                </div>

                {selectedFeature === 0 ? (
                  <div className="space-y-3">
                    <div className="overflow-hidden rounded-xl border border-gray-700/80 bg-gray-950">
                      <Image
                        src={mirrorHealthImg}
                        alt="Authentic LibGen GUI Mirror Status Dialog showing live endpoint latencies and health"
                        className="w-full h-auto object-contain mx-auto"
                      />
                    </div>
                    <div className="p-2.5 rounded-lg bg-gray-950/80 border border-gray-800 text-[11px] text-gray-400 flex items-center justify-between">
                      <span>Probing interval: <strong className="text-gray-200">30s</strong></span>
                      <span className="text-emerald-400 font-medium">Automatic Latency Sorting</span>
                    </div>
                  </div>
                ) : (
                  <div className="space-y-3">
                    <div className="overflow-hidden rounded-xl border border-gray-700/80 bg-gray-950">
                      <Image
                        src={desktopSearchImg}
                        alt="LibGen GUI Desktop interface demonstrating automated search and queue management"
                        className="w-full h-auto object-cover rounded-lg"
                      />
                    </div>
                    <div className="p-2.5 rounded-lg bg-gray-950/80 border border-gray-800 text-[11px] text-gray-400 flex items-center justify-between">
                      <span>Queue engine: <strong className="text-gray-200">Thread-Safe FIFO</strong></span>
                      <span className="text-cyan-400 font-medium">HTTP Range Byte-Level Resume</span>
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      </Container>
    </section>
  )
}
