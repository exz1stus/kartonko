import type { NextConfig } from "next";

const API_ORIGIN =
    process.env.NEXT_PUBLIC_API_ORIGIN || "http://localhost:8080";

const nextConfig: NextConfig = {
    images: {
        remotePatterns: [
            new URL(`${API_ORIGIN}/image/thumb/**`),
            new URL(`https://lh3.googleusercontent.com/a/**`),
        ],
    },
    eslint: {
        ignoreDuringBuilds: true,
    },
};

export default nextConfig;
