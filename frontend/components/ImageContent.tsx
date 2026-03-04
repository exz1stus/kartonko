import ImageMetadata from "@/lib/image";
import Image from "next/image";
import React from "react";
import TagSpan from "./Tags/TagSpan";
import UserElement from "./UserElement";
import TimeField from "./TimeField";
import { getLoggedUserServer, getUserByIdServer } from "@/lib/user.server";
import { isModerator } from "@/lib/user";
import EditImage from "./EditImage";

interface Props {
    image: ImageMetadata;
}

const API_ORIGIN = process.env.NEXT_PUBLIC_API_ORIGIN;

const ImageContent: React.FC<Props> = async ({ image }) => {
    const user = await getLoggedUserServer();
    const imageOwner = await getUserByIdServer(image.user_id);
    const editImage = user &&
        (isModerator(user) || user.id === image.user_id) && (
            <EditImage image={image} />
        );

    const { filename } = image;

    const tags = image.tags?.length > 0 && (
        <div className="flex flex-wrap gap-1">
            <span className="text-2xl">Tags:</span>
            <TagSpan
                tags={image.tags}
                tagStyle={"bg-surface-0 rounded-3xl px-2"}
            />
        </div>
    );

    return (
        <div className="flex justify-center gap-10 m-2">
            <Image
                src={`${API_ORIGIN}/image/raw/${filename}`}
                alt={filename}
                width={image.width}
                height={image.height}
                unoptimized
                className="rounded-2xl w-auto max-w-[50vw] min-h-[70vh] max-h-[80vh]"
            />
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
                    {imageOwner && <UserElement user={imageOwner} />}
                </div>
                <div className="flex flex-row items-center gap-2">
                    <span className="text-l">Uploaded: </span>
                    <TimeField time={image.uploaded_at} />
                </div>
                {tags}
                {editImage}
            </div>
        </div>
    );
};

export default ImageContent;
