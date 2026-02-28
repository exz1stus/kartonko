import { serverFetch } from "./serverFetch";

export default interface ImageMetadata {
    filename: string;
    tags: string[];
    width: number;
    height: number;
    format: string;
    user_id: number;
    uploaded_at: string;
}

export async function getImageMetadataByNameServer(
    name: string,
): Promise<ImageMetadata | null> {
    const res = await serverFetch(`/image/${name}`);
    if (res.status !== 200) {
        return null;
    }
    const image: ImageMetadata = await res.json();
    return image;
}

export async function getImageMetadataByIdServer(
    id: number,
): Promise<ImageMetadata | null> {
    const res = await serverFetch(`/image/id/${id}`);
    if (res.status !== 200) {
        return null;
    }
    const image: ImageMetadata = await res.json();
    console.log(image);
    return image;
}
