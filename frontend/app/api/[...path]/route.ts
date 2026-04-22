import { NextRequest } from "next/server";

async function proxyRequest(req: NextRequest, params: { path: string[] }) {
    const { path } = params;
    const backendUrl = `${process.env.API_LOCAL}/${path.join("/")}${req.nextUrl.search}`;

    // 1. Clone headers and remove problematic ones
    const headers = new Headers(req.headers);
    headers.delete("host"); // Let the client handle the new host
    headers.delete("connection"); // Do not forward connection state

    const options: RequestInit = {
        method: req.method,
        headers: headers,
    };

    if (req.method !== "GET" && req.method !== "HEAD") {
        options.body = req.body;
        options.duplex = "half";
    }

    const res = await fetch(backendUrl, options);

    const responseHeaders = new Headers(res.headers);
    responseHeaders.delete("content-encoding");

    return new Response(res.body, {
        status: res.status,
        headers: responseHeaders,
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
