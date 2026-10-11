export const GALLERY_IMAGES_MIME = "application/x-kartonko-image-ids";
export const GALLERY_SOURCE_BOARD_MIME = "application/x-kartonko-source-board";
export const GALLERY_SOURCE_SCOPE_MIME = "application/x-kartonko-source-scope";

export function readDraggedSourceBoardId(data: DataTransfer): number | undefined {
    const id = Number(data.getData(GALLERY_SOURCE_BOARD_MIME));
    return Number.isSafeInteger(id) && id > 0 ? id : undefined;
}

export function readDraggedSourceScope(data: DataTransfer): string | undefined {
    return data.getData(GALLERY_SOURCE_SCOPE_MIME) || undefined;
}

export function readDraggedImageIds(data: DataTransfer): number[] {
    try {
        const value = JSON.parse(data.getData(GALLERY_IMAGES_MIME));
        return Array.isArray(value) ? value.filter((id): id is number => Number.isSafeInteger(id) && id > 0) : [];
    } catch {
        return [];
    }
}
