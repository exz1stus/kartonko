import Image from "next/image";
import ImageMetadata from "@/lib/image";

interface Props {
    image: ImageMetadata;
    onLoad?: () => void;
}

const ImageCardServer: React.FC<Props> = ({ image, onLoad }) => {
    const API_ORIGIN = process.env.NEXT_PUBLIC_API_LOCAL;
    const { filename, format, width, height } = image;
    return (
        <div className="flex flex-col items-center bg-surface-20 rounded-xl hover:cursor-pointer">
            <Image
                src={`${API_ORIGIN}/image/thumb/${filename + "." + format}`}
                alt={filename}
                className="rounded-t-xl w-full h-auto"
                width={width}
                height={height}
                onLoad={onLoad}
                draggable={false}
            />
            <span className="px-2 max-w-[20ch] truncate">{filename}</span>
        </div>
    );
};

export default ImageCardServer;
