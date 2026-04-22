import type { NextConfig } from "next";

const nextConfig: NextConfig = {
    env: {
        API_ORIGIN: process.env.NEXT_PUBLIC_API_ORIGIN,
    },
    images: {
        remotePatterns: [
            new URL(`${process.env.NEXT_PUBLIC_API_ORIGIN}/**`),
            new URL(`${process.env.NEXT_PUBLIC_API_LOCAL}/**`),
            new URL(`https://lh3.googleusercontent.com/a/**`),
            new URL(`https://cdn-icons-png.flaticon.com/**`),
        ],
    },
    eslint: {
        ignoreDuringBuilds: true,
    },
};

export default nextConfig;
