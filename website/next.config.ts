import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  async rewrites() {
    const bff = process.env.BFF_BASE_URL ?? "http://localhost:8080"
    return [
      {
        source: "/api/:path*",
        destination: `${bff}/api/:path*`,
      },
    ]
  },
};

export default nextConfig;
