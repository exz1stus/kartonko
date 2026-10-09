import ImageContent from "@/components/ImageContent";
import { getImageByName } from "@/lib/api/generated/server";

interface Props {
    filename: string;
}

const ImagePage = async ({ params }: { params: Promise<Props> }) => {
    const { filename } = await params;

    let image = await getImageByName(filename);

    return (
        <div className="flex justify-center items-center w-full h-full">
            <title>{image.filename}</title>
            <ImageContent image={image} />
        </div>
    );
};

export default ImagePage;
