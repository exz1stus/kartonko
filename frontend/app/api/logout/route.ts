import { NextResponse } from "next/server";

export async function POST() {
    const res = await fetch(
        `${process.env.NEXT_PUBLIC_API_LOCAL}/auth/logout`,
        {
            method: "POST",
            credentials: "include",
        },
    );

    const logoutResponse = new NextResponse(res.body, { status: res.status });

    const setCookie = res.headers.get("set-cookie");
    if (setCookie) {
        logoutResponse.headers.set("set-cookie", setCookie);
    }

    return logoutResponse;
}
