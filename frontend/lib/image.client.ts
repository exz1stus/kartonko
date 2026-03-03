import { apiFetch } from "./apiFetch";
import ImageMetadata from "./image";

export async function getImageMetadataByName(
    name: string,
): Promise<ImageMetadata | null> {
    const res = await apiFetch(`/image/${name}`);
    if (res.status !== 200) {
        return null;
    }
    const image: ImageMetadata = await res.json();
    return image;
}

export async function getImageMetadataById(
    id: number,
): Promise<ImageMetadata | null> {
    const res = await apiFetch(`/image/id/${id}`);
    console.log(res);
    if (res.status !== 200) {
        return null;
    }
    const image: ImageMetadata = await res.json();
    console.log(image);
    return image;
}
