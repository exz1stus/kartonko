import ImageMetadata from "@/app/lib/image";
import Image from "next/image";
import React from "react";
import TagSpan from "./Tags/TagSpan";
import UserElement from "./UserElement";
import TimeField from "./TimeField";

interface Props {
    image: ImageMetadata;
}

const API_ORIGIN = process.env.NEXT_PUBLIC_API_ORIGIN;

const ImageContent: React.FC<Props> = ({ image }) => {
    const { filename } = image;

    const tags = image.tags?.length > 0 && (
        <div className="flex flex-wrap gap-1">
            <span className="text-2xl">Tags:</span>
            <TagSpan tags={image.tags} tagStyle={"bg-surface-0 rounded-3xl px-2"} />
        </div>
    );

    return (
        <div className="flex justify-center gap-10 m-2">
            <div className="flex justify-center max-w-[50vw]">
                <Image
                    src={`${API_ORIGIN}/image/raw/${filename}`}
                    alt={filename}
                    width={image.width}
                    height={image.height}
                    unoptimized
                    className="rounded-2xl w-auto min-h-[30vh] max-h-[90vh]"
                />
            </div>
            <div className="flex flex-col gap-5 bg-surface-tonal-0 p-5 rounded-2xl min-w-[20vw] max-w-[50vw]">
                <span className="text-3xl">
                    {image.filename}.{image.format}
                </span>
                <span className="text-l">
                    <span>Resolution: </span>
                    <span>
                        {image.width}x{image.height}
                    </span>
                </span>
                <div className="flex flex-row items-center gap-2">
                    <span className="text-l">Uploaded by:</span>
                    <UserElement id={image.user_id} />
                </div>
                <div className="flex flex-row items-center gap-2">
                    <span className="text-l">Uploaded: </span>
                    <TimeField time={image.uploaded_at} />
                </div>
                {tags}
            </div>
        </div>
    );
};

export default ImageContent;
