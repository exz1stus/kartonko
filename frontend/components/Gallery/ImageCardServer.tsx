import Image from "next/image";
import ImageMetadata from "@/app/lib/image";

interface Props extends ImageMetadata {
    onLoad?: () => void;
}

const API_ORIGIN = process.env.NEXT_PUBLIC_API_LOCAL;

const ImageCardServer: React.FC<Props> = ({ filename, tags, format, onLoad, width, height }) => {
    return (
        <div className="flex flex-col items-center bg-surface-20 rounded-xl hover:cursor-pointer">
            <Image
                src={`${API_ORIGIN}/image/thumb/${filename + "." + format}`}
                alt={filename}
                className="rounded-t-xl w-full h-auto"
                width={width}
                height={height}
                onLoad={onLoad}
            />
            <span className="px-2 max-w-[20ch] truncate">{filename}</span>
        </div>
    );
};

export default ImageCardServer;
