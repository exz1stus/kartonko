import ImageMetadata from "@/app/lib/image";
import FancySpan from "@/components/template/FancySpan";
import Scrollbar from "@/components/template/Scrollbar";
import Image from "next/image";
import { notFound } from "next/navigation";

const API_ORIGIN_SERVER = process.env.NEXT_PUBLIC_API_LOCAL;
const API_ORIGIN = process.env.NEXT_PUBLIC_API_ORIGIN;

const ImagePage = async ({ params }: { params: { filename: string } }) => {
    const { filename } = params;

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
        <>
            <FancySpan word={filename} className="text-2xl" />
            <Scrollbar>
                <div className="flex justify-center">
                    <Image
                        src={`${API_ORIGIN}/raw-image/${filename}`}
                        alt={filename}
                        width={image.width}
                        height={image.height}
                        unoptimized
                    />
                </div>
            </Scrollbar>
        </>
    );
};

export default ImagePage;
