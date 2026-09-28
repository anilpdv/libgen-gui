import { type Metadata, type Viewport } from 'next'
import { Inter } from 'next/font/google'
import clsx from 'clsx'

import '@/styles/tailwind.css'

const inter = Inter({
  subsets: ['latin'],
  display: 'swap',
  variable: '--font-inter',
})

export const viewport: Viewport = {
  themeColor: '#0891b2',
  width: 'device-width',
  initialScale: 1,
}

export const metadata: Metadata = {
  title: 'LibGen GUI | Native Desktop & Android Research Downloader',
  description:
    'A fast, lightweight, native client for searching, discovering, and downloading books and research papers from Library Genesis mirrors with automatic failover and queue recovery.',
  openGraph: {
    title: 'LibGen GUI | Native Desktop & Android Research Downloader',
    description:
      'Fast, native, multi-mirror scientific paper & book search client built with Go and Fyne with resumable downloads and Android SAF support.',
    url: 'https://anilpdv.github.io/libgen-gui',
    siteName: 'LibGen GUI',
    type: 'website',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'LibGen GUI | Native Desktop & Android Research Downloader',
    description:
      'Fast, native, multi-mirror book & scientific paper downloader built with Go and Fyne.',
  },
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en" className={clsx('bg-gray-50 text-gray-900 antialiased', inter.variable)}>
      <body>
        <a
          href="#features"
          className="sr-only focus:not-sr-only focus:fixed focus:top-4 focus:left-4 focus:z-50 focus:rounded-md focus:bg-cyan-600 focus:px-4 focus:py-2 focus:text-white focus:shadow-lg focus:outline-none"
        >
          Skip to main content
        </a>
        {children}
      </body>
    </html>
  )
}
