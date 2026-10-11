export const imagesDeletedEvent = "kartonko:images-deleted";

export function notifyImagesDeleted(ids: number[]) {
    window.dispatchEvent(new CustomEvent<number[]>(imagesDeletedEvent, { detail: ids }));
}
