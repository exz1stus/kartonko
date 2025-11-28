import Image from 'next/image';

interface Tag {
    Name: string
}

interface Props {
    filename: string
    tags: Tag[]
    width: number
    height: number
    onLoad?: () => void
}

const API_ORIGIN = process.env.NEXT_PUBLIC_API_LOCAL;

const ImageCardServer: React.FC<Props> = ({ filename, tags, onLoad, width, height }) => {
    return (
        <div className="flex flex-col items-center bg-surface-20 rounded-xl hover:cursor-pointer">
            <Image
                src={`${API_ORIGIN}/raw-image/${filename}`}
                alt={filename}
                className="rounded-t-xl"
                width={width}
                height={height}
                onLoad={onLoad}
            />
            <span
                className="px-2 max-w-[20ch] truncate"
            >{filename}</span>
        </div>
    );
};

export { type Props as ImageData, ImageCardServer };

