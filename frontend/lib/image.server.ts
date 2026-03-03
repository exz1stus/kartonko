import ImageMetadata from "./image";
import { serverFetch } from "./serverFetch";

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
    console.log(res);
    if (res.status !== 200) {
        return null;
    }
    const image: ImageMetadata = await res.json();
    console.log(image);
    return image;
}
