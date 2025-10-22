import Image from 'next/image';

interface Tag {
    Name: string
}

interface Props {
    filename: string
    tags: Tag[]
}

const API_ORIGIN = process.env.NEXT_PUBLIC_API_LOCAL;

const ImageCardServer: React.FC<Props> = ({ filename, tags }) => {
    return (
        <div className="flex flex-col items-center bg-surface-20 rounded-xl hover:cursor-pointer">
            <Image
                src={`${API_ORIGIN}/raw-image/${filename}`}
                alt={filename}
                className="rounded-t-xl w-auto h-[25vh] object-cover"
                width={250}
                height={250}
            />
            <span className="px-2 max-w-[20ch] text-2xl truncate">{filename}</span>
        </div>
    );
};

export { type Props as ImageData, ImageCardServer };

