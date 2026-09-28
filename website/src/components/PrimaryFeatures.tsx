'use client'

import { Fragment, useEffect, useId, useRef, useState } from 'react'
import { Tab, TabGroup, TabList, TabPanel, TabPanels } from '@headlessui/react'
import clsx from 'clsx'
import {
  type MotionProps,
  type Variant,
  type Variants,
  AnimatePresence,
  motion,
} from 'framer-motion'
import { useDebouncedCallback } from 'use-debounce'

import { AppScreen } from '@/components/AppScreen'
import { CircleBackground } from '@/components/CircleBackground'
import { Container } from '@/components/Container'
import { PhoneFrame } from '@/components/PhoneFrame'

const MotionAppScreenHeader = motion(AppScreen.Header)
const MotionAppScreenBody = motion(AppScreen.Body)

interface CustomAnimationProps {
  isForwards: boolean
  changeCount: number
}

const features = [
  {
    name: 'Multi-Mirror Failover & Live Prober',
    description:
      'Continuous background latency probing and automatic failover across official mirrors ensure your searches and downloads never stall or drop packets.',
    icon: MirrorProbeIcon,
    screen: MirrorScreen,
  },
  {
    name: 'Resumable Queue & Range Requests',
    description:
      'A robust serialized queue with pause, resume, cancel, and HTTP Range support. Interrupted downloads seamlessly pick up right where they left off without restarting.',
    icon: QueueDownloadIcon,
    screen: QueueScreen,
  },
  {
    name: 'Instant Multi-Field Book Search',
    description:
      'Debounced search across title, author, series, publisher, ISBN, and MD5 with live extension and size badges, direct 1-click downloads, and metadata inspection.',
    icon: SearchBookIcon,
    screen: SearchScreen,
  },
]

function MirrorProbeIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 32 32" fill="none" aria-hidden="true" {...props}>
      <circle cx={16} cy={16} r={16} fill="#06B6D4" fillOpacity={0.15} />
      <path
        d="M8 16a8 8 0 1 0 16 0A8 8 0 0 0 8 16zm4 0a4 4 0 1 1 8 0 4 4 0 0 1-8 0z"
        stroke="#06B6D4"
        strokeWidth="2"
      />
      <circle cx="16" cy="16" r="2" fill="#06B6D4" />
      <path
        d="M16 6v3M16 23v3M6 16h3M23 16h3"
        stroke="#06B6D4"
        strokeWidth="2"
        strokeLinecap="round"
      />
    </svg>
  )
}

function QueueDownloadIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 32 32" fill="none" aria-hidden="true" {...props}>
      <circle cx={16} cy={16} r={16} fill="#06B6D4" fillOpacity={0.15} />
      <path
        d="M11 15l5 5 5-5M16 8v12M9 24h14"
        stroke="#06B6D4"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  )
}

function SearchBookIcon(props: React.ComponentPropsWithoutRef<'svg'>) {
  return (
    <svg viewBox="0 0 32 32" fill="none" aria-hidden="true" {...props}>
      <circle cx={16} cy={16} r={16} fill="#06B6D4" fillOpacity={0.15} />
      <path
        d="M14.5 20a5.5 5.5 0 1 0 0-11 5.5 5.5 0 0 0 0 11zm4-1.5l4.5 4.5"
        stroke="#06B6D4"
        strokeWidth="2"
        strokeLinecap="round"
      />
      <path
        d="M8 23h16"
        stroke="#06B6D4"
        strokeWidth="1.5"
        strokeLinecap="round"
      />
    </svg>
  )
}

const headerAnimation: Variants = {
  initial: { opacity: 0, transition: { duration: 0.3 } },
  animate: { opacity: 1, transition: { duration: 0.3, delay: 0.3 } },
  exit: { opacity: 0, transition: { duration: 0.3 } },
}

const maxZIndex = 2147483647

const bodyVariantBackwards: Variant = {
  opacity: 0.4,
  scale: 0.8,
  zIndex: 0,
  filter: 'blur(4px)',
  transition: { duration: 0.4 },
}

const bodyVariantForwards: Variant = (custom: CustomAnimationProps) => ({
  y: '100%',
  zIndex: maxZIndex - custom.changeCount,
  transition: { duration: 0.4 },
})

const bodyAnimation: MotionProps = {
  initial: 'initial',
  animate: 'animate',
  exit: 'exit',
  variants: {
    initial: (custom: CustomAnimationProps, ...props) =>
      custom.isForwards
        ? bodyVariantForwards(custom, ...props)
        : bodyVariantBackwards,
    animate: (custom: CustomAnimationProps) => ({
      y: '0%',
      opacity: 1,
      scale: 1,
      zIndex: maxZIndex / 2 - custom.changeCount,
      filter: 'blur(0px)',
      transition: { duration: 0.4 },
    }),
    exit: (custom: CustomAnimationProps, ...props) =>
      custom.isForwards
        ? bodyVariantBackwards
        : bodyVariantForwards(custom, ...props),
  },
}

type ScreenProps =
  | {
      animated: true
      custom: CustomAnimationProps
    }
  | { animated?: false }

function MirrorScreen(props: ScreenProps) {
  return (
    <AppScreen className="w-full">
      <MotionAppScreenHeader {...(props.animated ? headerAnimation : {})}>
        <AppScreen.Title>Mirror Health</AppScreen.Title>
        <AppScreen.Subtitle>
          <span className="text-emerald-400">3 of 3 Active</span> • 124ms avg latency
        </AppScreen.Subtitle>
      </MotionAppScreenHeader>
      <MotionAppScreenBody
        {...(props.animated ? { ...bodyAnimation, custom: props.custom } : {})}
      >
        <div className="divide-y divide-gray-100">
          {[
            {
              url: 'https://libgen.is',
              latency: '82 ms',
              status: 'Optimal (Primary)',
              healthy: true,
              color: 'bg-emerald-500',
            },
            {
              url: 'https://libgen.rs',
              latency: '115 ms',
              status: 'Healthy (Failover 1)',
              healthy: true,
              color: 'bg-emerald-500',
            },
            {
              url: 'https://libgen.st',
              latency: '178 ms',
              status: 'Healthy (Failover 2)',
              healthy: true,
              color: 'bg-emerald-500',
            },
            {
              url: 'https://libgen.li',
              latency: '240 ms',
              status: 'Active (Fallback)',
              healthy: true,
              color: 'bg-cyan-500',
            },
          ].map((mirror) => (
            <div key={mirror.url} className="flex items-center gap-3 px-4 py-3.5">
              <div className="flex-none">
                <span className={`inline-block h-2.5 w-2.5 rounded-full ${mirror.color}`} />
              </div>
              <div className="flex-auto min-w-0">
                <div className="text-xs font-semibold text-gray-900 truncate">
                  {mirror.url}
                </div>
                <div className="text-[11px] text-gray-500">{mirror.status}</div>
              </div>
              <div className="flex-none text-right">
                <div className="text-xs font-medium font-mono text-gray-900">
                  {mirror.latency}
                </div>
                <div className="text-[10px] text-emerald-600 font-medium">99.9% Up</div>
              </div>
            </div>
          ))}
        </div>
        <div className="p-4 bg-gray-50 border-t border-gray-100">
          <div className="text-[11px] text-gray-600 flex items-center justify-between">
            <span>Adaptive Auto-Probe</span>
            <span className="font-semibold text-cyan-600">Every 30s</span>
          </div>
        </div>
      </MotionAppScreenBody>
    </AppScreen>
  )
}

function QueueScreen(props: ScreenProps) {
  return (
    <AppScreen className="w-full">
      <MotionAppScreenHeader {...(props.animated ? headerAnimation : {})}>
        <AppScreen.Title>Download Queue</AppScreen.Title>
        <AppScreen.Subtitle>
          <span className="text-cyan-400">2 Downloading</span> • 1 Queued
        </AppScreen.Subtitle>
      </MotionAppScreenHeader>
      <MotionAppScreenBody
        {...(props.animated ? { ...bodyAnimation, custom: props.custom } : {})}
      >
        <div className="p-4 space-y-4">
          <div className="rounded-xl border border-gray-100 bg-white p-3.5 shadow-xs">
            <div className="flex justify-between items-start">
              <div>
                <h4 className="text-xs font-semibold text-gray-900 line-clamp-1">
                  Computer Systems: A Programmer’s Perspective
                </h4>
                <p className="text-[11px] text-gray-500 mt-0.5">PDF • 14.8 MB / 19.4 MB</p>
              </div>
              <span className="text-[10px] font-semibold text-cyan-600 bg-cyan-50 px-2 py-0.5 rounded-full">
                76%
              </span>
            </div>
            <div className="mt-2.5 h-1.5 w-full overflow-hidden rounded-full bg-gray-100">
              <div className="h-full bg-cyan-500 rounded-full" style={{ width: '76%' }} />
            </div>
            <div className="mt-2 flex justify-between text-[10px] text-gray-400">
              <span>3.4 MB/s</span>
              <span>Resumable Chunk 4/5</span>
            </div>
          </div>

          <div className="rounded-xl border border-gray-100 bg-white p-3.5 shadow-xs">
            <div className="flex justify-between items-start">
              <div>
                <h4 className="text-xs font-semibold text-gray-900 line-clamp-1">
                  Designing Data-Intensive Applications
                </h4>
                <p className="text-[11px] text-gray-500 mt-0.5">EPUB • 8.2 MB / 8.2 MB</p>
              </div>
              <span className="text-[10px] font-semibold text-emerald-600 bg-emerald-50 px-2 py-0.5 rounded-full">
                Completed
              </span>
            </div>
            <div className="mt-2.5 h-1.5 w-full overflow-hidden rounded-full bg-emerald-100">
              <div className="h-full bg-emerald-500 rounded-full w-full" />
            </div>
            <div className="mt-2 flex justify-between text-[10px] text-gray-400">
              <span>Saved to /Downloads</span>
              <span>Verified MD5</span>
            </div>
          </div>

          <div className="rounded-xl border border-gray-100 bg-gray-50/50 p-3.5">
            <div className="flex justify-between items-start">
              <div>
                <h4 className="text-xs font-medium text-gray-700 line-clamp-1">
                  Structure and Interpretation of Computer Programs
                </h4>
                <p className="text-[11px] text-gray-400 mt-0.5">PDF • 12.1 MB</p>
              </div>
              <span className="text-[10px] font-medium text-gray-500 bg-gray-200/80 px-2 py-0.5 rounded-full">
                Queued
              </span>
            </div>
          </div>
        </div>
      </MotionAppScreenBody>
    </AppScreen>
  )
}

function SearchScreen(props: ScreenProps) {
  return (
    <AppScreen className="w-full">
      <MotionAppScreenHeader {...(props.animated ? headerAnimation : {})}>
        <AppScreen.Title>Book Search</AppScreen.Title>
        <AppScreen.Subtitle>Query: &quot;Distributed Systems&quot;</AppScreen.Subtitle>
      </MotionAppScreenHeader>
      <MotionAppScreenBody
        {...(props.animated ? { ...bodyAnimation, custom: props.custom } : {})}
      >
        <div className="p-3">
          <div className="rounded-lg bg-gray-100 px-3 py-2 text-xs text-gray-600 flex items-center gap-2 mb-3">
            <svg viewBox="0 0 20 20" fill="currentColor" className="w-3.5 h-3.5 text-gray-400">
              <path fillRule="evenodd" d="M9 3.5a5.5 5.5 0 100 11 5.5 5.5 0 000-11zM2 9a7 7 0 1112.452 4.391l3.328 3.329a.75.75 0 11-1.06 1.06l-3.329-3.328A7 7 0 012 9z" clipRule="evenodd" />
            </svg>
            <span className="text-gray-900 font-medium">Distributed Systems 3rd Ed</span>
          </div>

          <div className="space-y-2.5">
            {[
              {
                title: 'Distributed Systems: Principles and Paradigms',
                author: 'Andrew S. Tanenbaum, Maarten van Steen',
                year: '2023',
                format: 'PDF',
                size: '18.4 MB',
              },
              {
                title: 'Understanding Distributed Systems, Second Edition',
                author: 'Roberto Vitillo',
                year: '2022',
                format: 'EPUB',
                size: '5.1 MB',
              },
              {
                title: 'Database Internals: A Deep Dive into Distributed Systems',
                author: 'Alex Petrov',
                year: '2019',
                format: 'PDF',
                size: '14.2 MB',
              },
            ].map((book) => (
              <div key={book.title} className="rounded-lg border border-gray-100 p-2.5 hover:border-cyan-200 transition-colors">
                <h5 className="text-xs font-semibold text-gray-900 line-clamp-1">{book.title}</h5>
                <p className="text-[10px] text-gray-500 mt-0.5">{book.author}</p>
                <div className="mt-2 flex items-center justify-between">
                  <div className="flex gap-1.5 text-[9px]">
                    <span className="bg-gray-100 px-1.5 py-0.5 rounded font-mono font-medium">{book.format}</span>
                    <span className="bg-gray-100 px-1.5 py-0.5 rounded text-gray-600">{book.size}</span>
                    <span className="bg-gray-100 px-1.5 py-0.5 rounded text-gray-600">{book.year}</span>
                  </div>
                  <button className="bg-cyan-500 hover:bg-cyan-600 text-white text-[10px] font-medium px-2 py-0.5 rounded transition-colors">
                    Download
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      </MotionAppScreenBody>
    </AppScreen>
  )
}

function usePrevious<T>(value: T) {
  let ref = useRef<T | undefined>(undefined)

  useEffect(() => {
    ref.current = value
  }, [value])

  return ref.current
}

function FeaturesDesktop() {
  let [changeCount, setChangeCount] = useState(0)
  let [selectedIndex, setSelectedIndex] = useState(0)
  let prevIndex = usePrevious(selectedIndex)
  let isForwards = prevIndex === undefined ? true : selectedIndex > prevIndex

  let onChange = useDebouncedCallback(
    (selectedIndex) => {
      setSelectedIndex(selectedIndex)
      setChangeCount((changeCount) => changeCount + 1)
    },
    100,
    { leading: true },
  )

  return (
    <TabGroup
      className="grid grid-cols-12 items-center gap-8 lg:gap-16 xl:gap-24"
      selectedIndex={selectedIndex}
      onChange={onChange}
      vertical
    >
      <TabList className="relative z-10 order-last col-span-6 space-y-6">
        {features.map((feature, featureIndex) => (
          <div
            key={feature.name}
            className="relative rounded-2xl transition-colors hover:bg-gray-800/30"
          >
            {featureIndex === selectedIndex && (
              <motion.div
                layoutId="activeBackground"
                className="absolute inset-0 bg-gray-800"
                initial={{ borderRadius: 16 }}
              />
            )}
            <div className="relative z-10 p-8">
              <feature.icon className="h-8 w-8" />
              <h3 className="mt-6 text-lg font-semibold text-white">
                <Tab className="text-left data-selected:not-data-focus:outline-hidden">
                  <span className="absolute inset-0 rounded-2xl" />
                  {feature.name}
                </Tab>
              </h3>
              <p className="mt-2 text-sm text-gray-400 leading-relaxed">
                {feature.description}
              </p>
            </div>
          </div>
        ))}
      </TabList>
      <div className="relative col-span-6">
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2">
          <CircleBackground color="#13B5C8" className="animate-spin-slower" />
        </div>
        <PhoneFrame className="z-10 mx-auto w-full max-w-[366px]">
          <TabPanels as={Fragment}>
            <AnimatePresence
              initial={false}
              custom={{ isForwards, changeCount }}
            >
              {features.map((feature, featureIndex) =>
                selectedIndex === featureIndex ? (
                  <TabPanel
                    static
                    key={feature.name + changeCount}
                    className="col-start-1 row-start-1 flex focus:outline-offset-32 data-selected:not-data-focus:outline-hidden"
                  >
                    <feature.screen
                      animated
                      custom={{ isForwards, changeCount }}
                    />
                  </TabPanel>
                ) : null,
              )}
            </AnimatePresence>
          </TabPanels>
        </PhoneFrame>
      </div>
    </TabGroup>
  )
}

function FeaturesMobile() {
  let [activeIndex, setActiveIndex] = useState(0)
  let slideContainerRef = useRef<React.ElementRef<'div'>>(null)
  let slideRefs = useRef<Array<React.ElementRef<'div'>>>([])

  useEffect(() => {
    let observer = new window.IntersectionObserver(
      (entries) => {
        for (let entry of entries) {
          if (entry.isIntersecting && entry.target instanceof HTMLDivElement) {
            setActiveIndex(slideRefs.current.indexOf(entry.target))
            break
          }
        }
      },
      {
        root: slideContainerRef.current,
        threshold: 0.6,
      },
    )

    for (let slide of slideRefs.current) {
      if (slide) {
        observer.observe(slide)
      }
    }

    return () => {
      observer.disconnect()
    }
  }, [slideContainerRef, slideRefs])

  return (
    <>
      <div
        ref={slideContainerRef}
        className="-mb-4 flex snap-x snap-mandatory -space-x-4 overflow-x-auto overscroll-x-contain scroll-smooth pb-4 [scrollbar-width:none] sm:-space-x-6 [&::-webkit-scrollbar]:hidden"
      >
        {features.map((feature, featureIndex) => (
          <div
            key={featureIndex}
            ref={(ref) => {
              if (ref) {
                slideRefs.current[featureIndex] = ref
              }
            }}
            className="w-full flex-none snap-center px-4 sm:px-6"
          >
            <div className="relative transform overflow-hidden rounded-2xl bg-gray-800 px-5 py-6">
              <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2">
                <CircleBackground
                  color="#13B5C8"
                  className={featureIndex % 2 === 1 ? 'rotate-180' : undefined}
                />
              </div>
              <PhoneFrame className="relative mx-auto w-full max-w-[366px]">
                <feature.screen />
              </PhoneFrame>
              <div className="absolute inset-x-0 bottom-0 bg-gray-800/95 p-6 backdrop-blur-sm sm:p-10">
                <feature.icon className="h-8 w-8" />
                <h3 className="mt-6 text-sm font-semibold text-white sm:text-lg">
                  {feature.name}
                </h3>
                <p className="mt-2 text-sm text-gray-400 leading-relaxed">
                  {feature.description}
                </p>
              </div>
            </div>
          </div>
        ))}
      </div>
      <div className="mt-6 flex justify-center gap-3">
        {features.map((_, featureIndex) => (
          <button
            type="button"
            key={featureIndex}
            className={clsx(
              'relative h-0.5 w-4 rounded-full',
              featureIndex === activeIndex ? 'bg-gray-300' : 'bg-gray-500',
            )}
            aria-label={`Go to slide ${featureIndex + 1}`}
            onClick={() => {
              slideRefs.current[featureIndex].scrollIntoView({
                block: 'nearest',
                inline: 'nearest',
              })
            }}
          >
            <span className="absolute -inset-x-1.5 -inset-y-3" />
          </button>
        ))}
      </div>
    </>
  )
}

export function PrimaryFeatures() {
  return (
    <section
      id="features"
      aria-label="Features for downloading books and papers"
      className="bg-gray-900 py-20 sm:py-32"
    >
      <Container>
        <div className="mx-auto max-w-2xl lg:mx-0 lg:max-w-3xl">
          <h2 className="text-3xl font-medium tracking-tight text-white sm:text-4xl">
            Engineered for reliability, speed, and clean code.
          </h2>
          <p className="mt-3 text-lg text-gray-400 leading-relaxed">
            Unlike web scrapers and heavyweight wrappers, LibGen GUI is compiled into a lightweight native binary with production-grade networking, resilient mirror rotation, and thread-safe queue management.
          </p>
        </div>
      </Container>
      <div className="mt-16 md:hidden">
        <FeaturesMobile />
      </div>
      <Container className="hidden md:mt-20 md:block">
        <FeaturesDesktop />
      </Container>
    </section>
  )
}
