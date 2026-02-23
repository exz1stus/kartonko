import ImageMetadata from "@/lib/image";
import ImageContent from "@/components/ImageContent";
import { notFound } from "next/navigation";
import { serverFetch } from "@/lib/serverFetch";

interface Props {
    filename: string;
}

const ImagePage = async ({ params }: { params: Promise<Props> }) => {
    const { filename } = await params;

    let image: ImageMetadata;

    try {
        const res = await serverFetch(`/image/${filename}`);
        if (!res.ok) return notFound();

        image = await res.json();
        if (!image) return notFound();
    } catch (err) {
        console.error(`Failed to fetch image metadata`, err);
        return notFound();
    }

    return (
        <div className="flex justify-center items-center w-full h-full">
            <title>{image.filename}</title>
            <ImageContent image={image} />
        </div>
    );
};

export default ImagePage;
