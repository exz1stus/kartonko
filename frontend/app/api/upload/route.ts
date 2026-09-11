import { assertCaptchaFromRequest } from "@/lib/captcha.server";
import { postImageUpload, postTagsBatch } from "@/lib/api/generated/server";
import { NextRequest, NextResponse } from "next/server";
import { ImageUploadRequest } from "@/lib/image/upload";
import ApiError from "@/lib/api/error";

function parseRequestData(formData: FormData): ImageUploadRequest {
    const metadataStr = formData.get("metadata")?.toString();
    if (!metadataStr) throw Error("no metadata attached");
    const metadata: ImageUploadRequest = JSON.parse(metadataStr);

    return metadata;
}

function toBackendUploadBody(
    metadata: ImageUploadRequest,
    file: File,
): FormData {
    const backendFormData = new FormData();
    backendFormData.append("name", metadata.name);
    backendFormData.append("file", file);

    const tags = metadata.tags.concat(metadata.newTags || []);

    tags.forEach((tag) => {
        backendFormData.append("tags", tag);
    });

    return backendFormData;
}

async function handleNewTags(newTags?: string[]): Promise<void> {
    if (Array.isArray(newTags) && newTags.length > 0) {
        await postTagsBatch({ names: newTags });
    }
}

export async function POST(req: NextRequest) {
    try {
        const formData = await req.formData();
        await assertCaptchaFromRequest(req.headers, formData);

        const file = formData.get("file") as File;

        if (!file) {
            return NextResponse.json(
                { error: "Failed not attached" },
                { status: 400 },
            );
        }

        const reqData = parseRequestData(formData);

        await handleNewTags(reqData.newTags);

        const backendBody = toBackendUploadBody(reqData, file);

        const metadata = await postImageUpload({
            body: backendBody,
            credentials: "include",
        });
        return NextResponse.json(metadata, { status: 201 });
    } catch (err) {
        if (err instanceof ApiError) {
            return NextResponse.json(
                { error: err.data },
                { status: err.status },
            );
        }

        console.error(err);
        return NextResponse.json(
            { error: "Internal server error" },
            { status: 500 },
        );
    }
}
