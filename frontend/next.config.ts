import type { NextConfig } from "next";
import { config } from "dotenv";

config();

const isProduction = process.env.NODE_ENV === "production";

const nextConfig: NextConfig = {
  // Enable standalone output for Docker deployment
  output: "standalone",

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
      {
        protocol: "https",
        hostname: "ik.imagekit.io",
        pathname: "/**",
      },
      {
        protocol: "https",
        hostname: "imagekit.io",
        pathname: "/**",
      },
      {
        protocol: "https",
        hostname: "*.run.app", // Allow Cloud Run images
        pathname: "/**",
      },
      {
        protocol: "https",
        hostname: "bundui-images.netlify.app",
        pathname: "/**",
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
