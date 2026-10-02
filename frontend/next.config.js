/** @type {import('next').NextConfig} */
const nextConfig = {
  // stops Next from writing extra generated files into the repo
  agentRules: false,
  // the dev-only "N" badge would sit on top of the log out button in the nav rail
  devIndicators: { position: 'bottom-right' },
  // no MIME sniffing, no framing (clickjacking), no referrer to other sites
  async headers() {
    return [
      {
        source: '/:path*',
        headers: [
          { key: 'X-Content-Type-Options', value: 'nosniff' },
          { key: 'X-Frame-Options', value: 'DENY' },
          { key: 'Referrer-Policy', value: 'same-origin' },
        ],
      },
    ]
  },
  async rewrites() {
    return [
      {
        source: '/api/v1/:path*',
        destination: 'http://localhost:8080/:path*',
      },
    ]
  },
}

module.exports = nextConfig