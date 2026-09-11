import { ImageMetadata } from "@/lib/api/generated/model/imageMetadata";
import ImageContent from "@/components/ImageContent";
import { notFound } from "next/navigation";
import { getImageName } from "@/lib/api/generated/server";

interface Props {
    filename: string;
}

const ImagePage = async ({ params }: { params: Promise<Props> }) => {
    const { filename } = await params;

    let image = await getImageName(filename);

    return (
        <div className="flex justify-center items-center w-full h-full">
            <title>{image.filename}</title>
            <ImageContent image={image} />
        </div>
    );
};

export default ImagePage;
