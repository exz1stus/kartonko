import type { NextConfig } from "next";

function getRemotePattern(urlStr: string | undefined) {
    if (!urlStr) return [];
    try {
        const parsed = new URL(urlStr);
        return [
            {
                protocol: parsed.protocol.replace(":", "") as "http" | "https",
                hostname: parsed.hostname,
                port: parsed.port || "",
                pathname: "/**",
            },
        ];
    } catch {
        return [];
    }
}

const nextConfig: NextConfig = {
    env: {
        API_ORIGIN: process.env.NEXT_PUBLIC_API_ORIGIN,
        API_LOCAL: process.env.API_LOCAL,
    },
    async rewrites() {
        return [
            {
                source: "/apilocal/:path*",
                destination: `${process.env.API_LOCAL}/:path*`,
            },
        ];
    },
    images: {
        remotePatterns: [
            ...getRemotePattern(process.env.API_LOCAL),
            new URL(`https://lh3.googleusercontent.com/a/**`),
            new URL(`https://cdn-icons-png.flaticon.com/**`),
        ],
    },
    eslint: {
        ignoreDuringBuilds: true,
    },
};

export default nextConfig;
