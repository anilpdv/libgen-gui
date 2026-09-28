'use client'

import { useState, useEffect } from 'react'
import clsx from 'clsx'
import { motion } from 'framer-motion'
import { AppScreen } from '@/components/AppScreen'

export function AppDemo() {
  const [downloadProgress, setDownloadProgress] = useState(64)
  const [selectedBooks, setSelectedBooks] = useState<Record<string, boolean>>({
    book1: true,
    book2: true,
  })

  useEffect(() => {
    const interval = setInterval(() => {
      setDownloadProgress((prev) => (prev >= 100 ? 25 : prev + 3))
    }, 400)
    return () => clearInterval(interval)
  }, [])

  return (
    <AppScreen>
      <AppScreen.Body>
        <div className="flex flex-col h-full bg-slate-900 text-slate-100 text-xs font-sans select-none overflow-hidden">
          {/* Top Window Bar */}
          <div className="flex items-center justify-between px-3 py-2 bg-slate-950 border-b border-slate-800">
            <div className="flex items-center gap-1.5">
              <div className="w-2.5 h-2.5 rounded-full bg-rose-500/80" />
              <div className="w-2.5 h-2.5 rounded-full bg-amber-500/80" />
              <div className="w-2.5 h-2.5 rounded-full bg-emerald-500/80" />
            </div>
            <div className="text-[11px] font-medium text-slate-400">
              LibGen Downloader v2.0
            </div>
            <div className="flex items-center gap-1 px-1.5 py-0.5 rounded bg-emerald-950/80 border border-emerald-500/30 text-[10px] text-emerald-400 font-mono">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
              Mirrors: 2/2 (38ms)
            </div>
          </div>

          {/* Search Header */}
          <div className="p-3 bg-slate-900 border-b border-slate-800 space-y-2">
            <div className="flex gap-2">
              <div className="flex-1 flex items-center bg-slate-950 rounded-lg px-2.5 py-1.5 border border-slate-700 text-slate-200">
                <svg viewBox="0 0 20 20" fill="currentColor" className="w-3.5 h-3.5 mr-1.5 text-slate-400">
                  <path fillRule="evenodd" d="M9 3.5a5.5 5.5 0 100 11 5.5 5.5 0 000-11zM2 9a7 7 0 1112.452 4.391l3.328 3.329a.75.75 0 11-1.06 1.06l-3.329-3.328A7 7 0 012 9z" clipRule="evenodd" />
                </svg>
                <span className="text-slate-100 font-medium">Distributed Systems</span>
              </div>
              <div className="px-2 py-1 bg-slate-800 border border-slate-700 rounded-lg text-slate-300 font-medium text-[11px] flex items-center">
                PDF ▾
              </div>
              <div className="px-3 py-1 bg-cyan-600 rounded-lg text-white font-semibold text-[11px] flex items-center">
                Search
              </div>
            </div>
            <div className="flex items-center justify-between text-[10px] text-slate-400 px-0.5">
              <span>Found 25 results — Page 1</span>
              <span className="text-cyan-400 font-medium">Auto-failover enabled</span>
            </div>
          </div>

          {/* Results List */}
          <div className="flex-1 overflow-y-auto p-2.5 space-y-2">
            {/* Book 1 */}
            <div
              onClick={() => setSelectedBooks((prev) => ({ ...prev, book1: !prev.book1 }))}
              className={clsx(
                'p-2.5 rounded-lg border transition-all cursor-pointer flex gap-2.5 items-start',
                selectedBooks.book1
                  ? 'bg-cyan-950/40 border-cyan-500/60 shadow-xs'
                  : 'bg-slate-950 border-slate-800',
              )}
            >
              <div className={clsx(
                'w-4 h-4 rounded mt-0.5 flex items-center justify-center border text-[10px]',
                selectedBooks.book1
                  ? 'bg-cyan-600 border-cyan-500 text-white'
                  : 'border-slate-700 bg-slate-900',
              )}>
                {selectedBooks.book1 ? '✓' : ''}
              </div>
              <div className="flex-1 min-w-0">
                <div className="font-semibold text-slate-100 text-[11px] truncate">
                  Designing Data-Intensive Applications
                </div>
                <div className="text-[10px] text-slate-400 truncate">
                  Martin Kleppmann • O&apos;Reilly Media (2017)
                </div>
                <div className="flex items-center gap-2 mt-1.5">
                  <span className="px-1.5 py-0.5 rounded bg-cyan-950 border border-cyan-800/80 text-cyan-300 text-[9px] font-bold uppercase">
                    PDF
                  </span>
                  <span className="text-[10px] text-slate-400">14.8 MB</span>
                  <span className="text-[10px] text-emerald-400 font-medium ml-auto">
                    Active Download
                  </span>
                </div>
              </div>
            </div>

            {/* Book 2 */}
            <div
              onClick={() => setSelectedBooks((prev) => ({ ...prev, book2: !prev.book2 }))}
              className={clsx(
                'p-2.5 rounded-lg border transition-all cursor-pointer flex gap-2.5 items-start',
                selectedBooks.book2
                  ? 'bg-cyan-950/40 border-cyan-500/60 shadow-xs'
                  : 'bg-slate-950 border-slate-800',
              )}
            >
              <div className={clsx(
                'w-4 h-4 rounded mt-0.5 flex items-center justify-center border text-[10px]',
                selectedBooks.book2
                  ? 'bg-cyan-600 border-cyan-500 text-white'
                  : 'border-slate-700 bg-slate-900',
              )}>
                {selectedBooks.book2 ? '✓' : ''}
              </div>
              <div className="flex-1 min-w-0">
                <div className="font-semibold text-slate-100 text-[11px] truncate">
                  Distributed Systems: Principles & Paradigms
                </div>
                <div className="text-[10px] text-slate-400 truncate">
                  Andrew S. Tanenbaum, Maarten van Steen (2016)
                </div>
                <div className="flex items-center gap-2 mt-1.5">
                  <span className="px-1.5 py-0.5 rounded bg-slate-800 text-slate-300 text-[9px] font-bold uppercase">
                    PDF
                  </span>
                  <span className="text-[10px] text-slate-400">8.2 MB</span>
                  <span className="text-[10px] text-amber-400 font-medium ml-auto">
                    Queued (#1)
                  </span>
                </div>
              </div>
            </div>

            {/* Book 3 */}
            <div className="p-2.5 rounded-lg border border-slate-800 bg-slate-950 flex gap-2.5 items-start opacity-70">
              <div className="w-4 h-4 rounded mt-0.5 border border-slate-700 bg-slate-900" />
              <div className="flex-1 min-w-0">
                <div className="font-semibold text-slate-100 text-[11px] truncate">
                  Database Internals: A Deep-Dive into Storage Engines
                </div>
                <div className="text-[10px] text-slate-400 truncate">
                  Alex Petrov • O&apos;Reilly (2019)
                </div>
                <div className="flex items-center gap-2 mt-1.5">
                  <span className="px-1.5 py-0.5 rounded bg-slate-800 text-slate-300 text-[9px] font-bold uppercase">
                    EPUB
                  </span>
                  <span className="text-[10px] text-slate-400">6.4 MB</span>
                </div>
              </div>
            </div>
          </div>

          {/* Bottom Download Control Bar */}
          <div className="p-3 bg-slate-950 border-t border-slate-800 space-y-2">
            <div className="flex items-center justify-between text-[10px]">
              <div className="text-slate-300 font-medium truncate flex-1 mr-2">
                Downloading: <span className="text-cyan-400 font-semibold">{((14.8 * downloadProgress) / 100).toFixed(1)} MB / 14.8 MB</span> ({downloadProgress}%)
              </div>
              <div className="px-1.5 py-0.5 rounded bg-slate-800 text-slate-300 font-mono text-[9px]">
                3.4 MB/s
              </div>
            </div>

            {/* Progress Bar */}
            <div className="h-1.5 w-full bg-slate-800 rounded-full overflow-hidden">
              <motion.div
                className="h-full bg-linear-to-r from-cyan-500 to-emerald-400"
                style={{ width: `${downloadProgress}%` }}
                transition={{ duration: 0.3 }}
              />
            </div>

            <div className="flex items-center justify-between pt-1">
              <span className="text-[10px] text-slate-400 truncate">
                Save to: ~/Downloads/Books
              </span>
              <div className="flex gap-1.5">
                <div className="px-2 py-0.5 bg-slate-800 text-slate-300 rounded text-[10px] font-medium">
                  Queue (1)
                </div>
                <div className="px-2 py-0.5 bg-rose-950 text-rose-300 border border-rose-800/80 rounded text-[10px] font-medium">
                  Cancel
                </div>
              </div>
            </div>
          </div>
        </div>
      </AppScreen.Body>
    </AppScreen>
  )
}
