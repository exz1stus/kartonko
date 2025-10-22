import { ImageData } from '@/components/Gallery/ImageCard';
import Gallery from './Gallery';

interface ImagesResponse {
    imageData: ImageData[];
}

const fetchImages = async (intialFetchSize: number) => {
    const API_ORIGIN = process.env.NEXT_PUBLIC_API_LOCAL;
    const response = await fetch(`${API_ORIGIN}/images?cursor=0&limit=${intialFetchSize}`, {
        cache: 'no-store',
    });

    if (!response.ok) throw new Error("Failed to fetch images");
    const data: ImagesResponse = await response.json();
    return data.imageData;
};

interface Props {
    initialFetchSize: number;
}

const GalleryServer = async ({ initialFetchSize }: Props) => {
    const images = await fetchImages(initialFetchSize);
    return <Gallery initialImages={images} />;
}

export default GalleryServer;
