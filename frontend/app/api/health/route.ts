import { NextResponse } from "next/server";

export const dynamic = "force-dynamic";
export const runtime = "nodejs";

export async function GET() {
    const timestamp = new Date().toISOString();
    const uptime = process.uptime();

    try {
        return NextResponse.json(
            { status: "healthy", timestamp, uptime },
            {
                status: 200,
                headers: { "Cache-Control": "no-store, max-age=0" },
            },
        );
    } catch (error) {
        return NextResponse.json(
            { status: "unhealthy", timestamp, error: (error as Error).message },
            {
                status: 503,
                headers: { "Cache-Control": "no-store, max-age=0" },
            },
        );
    }
}
