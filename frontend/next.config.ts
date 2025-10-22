import type { NextConfig } from "next";

const API_ORIGIN = process.env.NEXT_PUBLIC_API_LOCAL;

const nextConfig: NextConfig = {
    images: {
        remotePatterns: [
            new URL(`${API_ORIGIN}/raw-image/**`),
            new URL(`https://lh3.googleusercontent.com/a/**`),
        ],
    },
};

export default nextConfig;
