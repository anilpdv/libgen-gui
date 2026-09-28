const basePath = process.env.NEXT_PUBLIC_BASE_PATH || '/libgen-gui'

/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'export',
  distDir: 'out',
  images: {
    unoptimized: true,
  },
  basePath: basePath === '/' ? '' : basePath,
  assetPrefix: basePath === '/' ? '' : basePath,
  trailingSlash: true,
}

module.exports = nextConfig
