import { NextRequest } from "next/server";

async function proxyRequest(req: NextRequest, params: { path: string[] }) {
    const { path } = params;
    const backendUrl =
        process.env.NEXT_PUBLIC_API_LOCAL +
        "/" +
        path.join("/") +
        req.nextUrl.search;

    const res = await fetch(backendUrl, {
        method: req.method,
        headers: {
            ...Object.fromEntries(req.headers),
        },
        body:
            req.method !== "GET" && req.method !== "HEAD"
                ? await req.text()
                : undefined,
        credentials: "include",
    });

    return new Response(res.body, {
        status: res.status,
        headers: res.headers,
    });
}

type RouteContext = { params: Promise<{ path: string[] }> };

export async function GET(req: NextRequest, { params }: RouteContext) {
    return proxyRequest(req, await params);
}

export async function POST(req: NextRequest, { params }: RouteContext) {
    return proxyRequest(req, await params);
}

export async function PUT(req: NextRequest, { params }: RouteContext) {
    return proxyRequest(req, await params);
}

export async function PATCH(req: NextRequest, { params }: RouteContext) {
    return proxyRequest(req, await params);
}

export async function DELETE(req: NextRequest, { params }: RouteContext) {
    return proxyRequest(req, await params);
}
