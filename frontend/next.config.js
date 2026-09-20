/** @type {import('next').NextConfig} */
const nextConfig = {
  // stops Next from writing extra generated files into the repo
  agentRules: false,
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