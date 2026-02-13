import ImageMetadata from "@/app/lib/image";
import ImageContent from "@/components/ImageContent";
import { notFound } from "next/navigation";

const API_ORIGIN_SERVER = process.env.NEXT_PUBLIC_API_LOCAL;

const ImagePage = async ({ params }: { params: { filename: string } }) => {
    const { filename } = await params;

    let image: ImageMetadata;

    try {
        const res = await fetch(`${API_ORIGIN_SERVER}/image/${filename}`);
        if (!res.ok) return notFound();

        image = await res.json();
        if (!image) return notFound();
    } catch (err) {
        console.error(`Failed to fetch image metadata`, err);
        return notFound();
    }

    return (
        <div className="flex justify-center items-center w-full h-full">
            <ImageContent image={image} />
        </div>
    );
};

export default ImagePage;
