import { postBoardImage } from "@/lib/api/generated/client";
import ApiError from "@/lib/api/error";
import { BoardItemResponse } from "@/lib/api/generated/model";

export async function addImagesToBoard(boardId: number, imageIds: number[]): Promise<{
    added: BoardItemResponse[];
    already: number;
    failed: number;
}> {
    const ids = [...new Set(imageIds)];
    const added: BoardItemResponse[] = [];
    let already = 0;
    let failed = 0;
    for (let offset = 0; offset < ids.length; offset += 4) {
        const results = await Promise.allSettled(ids.slice(offset, offset + 4).map((id) =>
            postBoardImage(boardId, { image_id: id })));
        for (const result of results) {
            if (result.status === "fulfilled") added.push(result.value);
            else if (result.reason instanceof ApiError && result.reason.status === 409) already++;
            else failed++;
        }
    }
    return { added, already, failed };
}
