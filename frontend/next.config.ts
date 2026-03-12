import type { NextConfig } from "next";

const nextConfig: NextConfig = {
    env: {
        API_ORIGIN: process.env.NEXT_PUBLIC_API_ORIGIN
    },
    images: {
        remotePatterns: [
            new URL(`${process.env.NEXT_PUBLIC_API_ORIGIN}/image/thumb/**`),
            new URL(`https://lh3.googleusercontent.com/a/**`),
        ],
    },
    eslint: {
        ignoreDuringBuilds: true,
    },
    
};

export default nextConfig;
