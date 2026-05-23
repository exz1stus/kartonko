import verifyTurnstile from "@/lib/captcha.server";
import { serverFetch } from "@/lib/serverFetch";
import { addNewTagsBatchServer } from "@/lib/tag/tag.server";
import { NextRequest, NextResponse } from "next/server";

export async function POST(req: NextRequest) {
    const formData = await req.formData();
    const token = formData.get("cf-turnstile-response")?.toString();
    formData.delete("cf-turnstile-response");

    if (!token) {
        return NextResponse.json(
            { error: "Missing captcha token" },
            { status: 400 },
        );
    }

    const ip =
        req.headers.get("x-forwarded-for") ??
        req.headers.get("x-real-ip") ??
        "";
    const valid = await verifyTurnstile(token, ip);

    if (!valid) {
        return NextResponse.json({ error: "Captcha failed" }, { status: 403 });
    }

    const metadataString = formData.get("metadata")?.toString();
    if (metadataString) {
        try {
            const metadata = JSON.parse(metadataString);
            const newTags = metadata.newTags;

            if (Array.isArray(newTags) && newTags.length > 0) {
                await addNewTagsBatchServer(newTags);
            }

            const currentTags = Array.isArray(metadata.tags)
                ? metadata.tags
                : [];
            metadata.tags = [...currentTags, ...newTags];

            delete metadata.newTags;

            formData.set("metadata", JSON.stringify(metadata));
        } catch (err) {
            return NextResponse.json(
                { error: "Invalid metadata or failed to process new tags" },
                { status: 400 },
            );
        }
    }

    const res = await serverFetch("/upload", {
        method: "POST",
        body: formData,
    });

    const responseHeaders = new Headers(res.headers);
    responseHeaders.delete("content-encoding");

    return new Response(res.body, {
        status: res.status,
        headers: responseHeaders,
    });
}
