import { assertCaptchaFromRequest } from "@/lib/captcha.server";
import { createUploadSession } from "@/lib/uploadSession.server";
import ApiError from "@/lib/api/error";
import { NextRequest, NextResponse } from "next/server";

export async function POST(req: NextRequest) {
    try {
        const formData = await req.formData();
        await assertCaptchaFromRequest(req.headers, formData);
        return NextResponse.json({ token: await createUploadSession() });
    } catch (error) {
        if (error instanceof ApiError)
            return NextResponse.json({ error: error.message }, { status: error.status });
        return NextResponse.json({ error: "Could not start upload" }, { status: 500 });
    }
}
