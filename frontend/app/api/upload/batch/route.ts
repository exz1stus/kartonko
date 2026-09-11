import { assertCaptchaFromRequest } from "@/lib/captcha.server";
import { NextRequest, NextResponse } from "next/server";
import { postTagsBatch } from "@/lib/api/generated/server";
import { serverFetch } from "@/lib/api/serverFetch";
import ApiError from "@/lib/api/error";

export async function POST(req: NextRequest) {
    const formData = await req.formData();
    await assertCaptchaFromRequest(req.headers, formData).catch(
        (err: ApiError) =>
            NextResponse.json({ error: err.data }, { status: err.status }),
    );

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

            const res = await serverFetch("/image/upload/batch", {
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
    } catch (err) {
        return NextResponse.json(
            { error: "Invalid metadata or failed to process batch tags" },
            { status: 400 },
        );
    }
}
