import verifyTurnstile from "@/lib/captcha.server";
import {
    postImageUpload,
    postTagsBatch,
    PostImageUploadBody,
} from "@/lib/api/generated/server";
import { NextRequest, NextResponse } from "next/server";

async function verifyCaptcha(
    headers: Headers,
    formData: FormData,
): Promise<boolean> {
    const token = formData.get("cf-turnstile-response")?.toString();

    if (!token) {
        throw new Error("Missing captcha token");
    }

    const ip = headers.get("x-forwarded-for") ?? headers.get("x-real-ip") ?? "";

    return verifyTurnstile(token, ip);
}

export async function POST(req: NextRequest) {
    var postMetadata: PostImageUploadBody;
    try {
        const formData = await req.formData();
        let captchaValid = await verifyCaptcha(req.headers, formData);
        if (!captchaValid) {
            return NextResponse.json(
                { error: "Captcha failed" },
                { status: 403 },
            );
        }

        const metadataString = formData.get("metadata")?.toString();
        const file = formData.get("file") as Blob;
        if (!metadataString) {
            return NextResponse.json(
                { error: "Failed to parse metadata" },
                { status: 400 },
            );
        }
        if (!file) {
            return NextResponse.json(
                { error: "Failed not attached" },
                { status: 400 },
            );
        }

        const metadata = JSON.parse(metadataString);
        const newTags = metadata;

        if (Array.isArray(newTags) && newTags.length > 0) {
            await postTagsBatch({ names: newTags });
        }

        const currentTags = Array.isArray(metadata.tags) ? metadata.tags : [];
        metadata.tags = [...currentTags, ...newTags];
        delete metadata.newTags;
        postMetadata = {
            file: file,
            metadata: JSON.stringify(metadata),
        };

        return postImageUpload(postMetadata);
    } catch (err) {
        console.error(err);

        return NextResponse.json(
            {
                error:
                    err instanceof Error
                        ? err.message
                        : "Internal server error",
            },
            { status: 500 },
        );
    }
}
