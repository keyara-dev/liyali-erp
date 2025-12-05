import type { NextConfig } from "next";
import { config } from "dotenv";

config();

const isProduction = process.env.NODE_ENV === "production";

const nextConfig: NextConfig = {
  // assetPrefix: isProduction ? "https://dashboard.shadcnuikit.com" : undefined,
  // typescript: {
  //   ignoreBuildErrors: true,
  // },
  images: {
    remotePatterns: [
      {
        protocol: "http",
        hostname: "localhost",
      },
    ],
  },

  experimental: {
    optimizePackageImports: ["lucide-react"], // Optimize chunk splitting
    serverActions: {
      bodySizeLimit: (process.env.MAX_FILE_SIZE_LIMIT as any) || "60mb",
      // allowedForwardedHosts: ["liyali.io"] ,
      // allowedOrigins: ["liyali.io"],
    },
  },
};

export default nextConfig;
