import { NextRequest } from "next/server";

export async function handler(req: NextRequest, { params }: any) {
    const { path } = await params;
    const backendUrl =
        process.env.NEXT_PUBLIC_API_LOCAL + "/" + path.join("/") + req.nextUrl.search;

    const res = await fetch(backendUrl, {
        method: req.method,
        headers: {
            ...Object.fromEntries(req.headers),
        },
        body: req.method !== "GET" && req.method !== "HEAD" ? await req.text() : undefined,
        credentials: "include",
    });

    return new Response(res.body, {
        status: res.status,
        headers: res.headers,
    });
}

export { handler as GET };
export { handler as POST };
export { handler as PUT };
export { handler as PATCH };
export { handler as DELETE };
