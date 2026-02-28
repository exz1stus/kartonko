import type { NextConfig } from "next";

const API_ORIGIN = process.env.NEXT_PUBLIC_API_ORIGIN;

const nextConfig: NextConfig = {
    images: {
        remotePatterns: [
            new URL(`${API_ORIGIN}/image/thumb/**`),
            new URL(`${API_ORIGIN}/image/thumb/**`),
            new URL(`https://lh3.googleusercontent.com/a/**`),
        ],
    },
};

export default nextConfig;
