import { postBoardImage } from "@/lib/api/generated/client";
import ApiError from "@/lib/api/error";
import { UploadAction } from "./uploadIntent";

export async function runUploadAction(action: UploadAction, imageId: number) {
    switch (action.kind) {
        case "add_to_board":
            try {
                await postBoardImage(action.boardId, { image_id: imageId });
            } catch (error) {
                if (error instanceof ApiError && error.status === 409) return;
                throw error;
            }
    }
}
