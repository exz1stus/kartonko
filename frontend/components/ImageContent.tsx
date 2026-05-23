import ImageMetadata from "@/lib/image";
import Image from "next/image";
import React from "react";
import TagSpan from "./Tags/TagSpan";
import UserElement from "./UserElement";
import TimeField from "./TimeField";
import { getLoggedUserServer, getUserByIdServer } from "@/lib/user/user.server";
import { isModerator } from "@/lib/user/user";
import EditImage from "./EditImage";

interface Props {
    image: ImageMetadata;
}

const ImageContent: React.FC<Props> = async ({ image }) => {
    const API_ORIGIN = process.env.NEXT_PUBLIC_API_ORIGIN;
    const user = await getLoggedUserServer();
    const imageOwner = await getUserByIdServer(image.user_id);

    const { filename } = image;

    const tags = image.tags?.length > 0 && (
        <div className="flex flex-wrap gap-1">
            <span className="text-2xl">Tags:</span>
            <TagSpan
                tags={image.tags}
                tagStyle={"bg-surface-20 rounded-3xl px-2"}
            />
        </div>
    );

    return (
        <div className="bg-surface-0 w-full h-full">
            <div className="flex lg:flex-row flex-col items-stretch gap-6 mx-auto w-full h-full">
                <div className="relative flex flex-1 justify-center items-center bg-black/5 rounded-2xl min-w-0 min-h-0">
                    <Image
                        src={`${API_ORIGIN}/image/raw/${filename}`}
                        alt={filename}
                        width={image.width}
                        height={image.height}
                        unoptimized
                        className="w-full h-auto max-h-full object-contain"
                    />
                </div>
                <div className="flex justify-center">
                    <div className="flex flex-col gap-5 p-5 rounded-2xl min-w-0">
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
                        {user && (
                            <EditImage
                                image={image}
                                hasPermission={
                                    isModerator(user) ||
                                    user.id === image.user_id
                                }
                            />
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
};

export default ImageContent;
