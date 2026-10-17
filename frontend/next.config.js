/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  // Only /embed/* may be framed by other sites; everything else is clickjacking-protected.
  async headers() {
    return [
      {
        source: "/((?!embed(?:/|$)).*)",
        headers: [
          { key: "Content-Security-Policy", value: "frame-ancestors 'none'" },
          { key: "X-Frame-Options", value: "DENY" },
        ],
      },
    ];
  },
};

module.exports = nextConfig;
