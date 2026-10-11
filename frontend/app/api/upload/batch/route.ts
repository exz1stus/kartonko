import { assertUploadSession } from "@/lib/uploadSession.server";
import { NextRequest, NextResponse } from "next/server";
import { postTagsBatch } from "@/lib/api/generated/server";
import { serverFetch } from "@/lib/api/serverFetch";
import ApiError from "@/lib/api/error";

export async function POST(req: NextRequest) {
    try {
        await assertUploadSession(req.headers);
    } catch (err) {
        if (err instanceof ApiError) return NextResponse.json({ error: err.message }, { status: err.status });
        return NextResponse.json({ error: "Upload session verification failed" }, { status: 500 });
    }

    let formData: FormData;
    try {
        formData = await req.formData();
    } catch {
        return NextResponse.json({ error: "Could not read upload data" }, { status: 400 });
    }

    const metadataString = formData.get("metadata")?.toString();
    if (!metadataString)
        return NextResponse.json(
            { error: "Metadata is empty" },
            { status: 400 },
        );
    try {
        const metadata = JSON.parse(metadataString);
        const items = metadata.data;

        if (Array.isArray(items)) {
            // Extract all unique new tags from all items combined
            const allNewTags: string[] = [
                ...new Set(items.flatMap((item: any) => item.newTags || [])),
            ].filter(Boolean);

            if (allNewTags?.length > 0) {
                await postTagsBatch({ names: allNewTags });
            }

            //Loop through each image entry, merge tags, and remove newTags
            items.forEach((item: any) => {
                const currentTags = Array.isArray(item.tags) ? item.tags : [];
                const itemNewTags = Array.isArray(item.newTags)
                    ? item.newTags
                    : [];

                item.tags = [...currentTags, ...itemNewTags];
                delete item.newTags;
            });

            formData.set("metadata", JSON.stringify(metadata));

            const res = await serverFetch("/image/batch", {
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
        return NextResponse.json({ error: "Invalid batch metadata" }, { status: 400 });
    } catch (err) {
        if (err instanceof ApiError)
            return NextResponse.json({ error: err.message }, { status: err.status });
        return NextResponse.json(
            { error: "Could not process upload batch" },
            { status: 500 },
        );
    }
}
