"use client";
import Image from "next/image";
import PerspectiveCard from "./PerspectiveCard";
import React, { useRef } from "react";
import { useRouter } from "next/navigation";
import { ImageMetadata } from "@/lib/api/generated/model";
interface Props {
    image: ImageMetadata;
    className?: string;
    style?: React.CSSProperties;
}

const ImageCard: React.FC<Props> = ({ image, className, style }) => {
    const { filename, width, height } = image;
    const selfRef = useRef<HTMLDivElement>(null);
    const router = useRouter();
    const onClick = () => {
        router.push(`/image/${image.filename}`);
    };

    const onLoad = () => {
        selfRef.current?.classList.remove("opacity-0");
        selfRef.current?.clientHeight;
    };

    return (
        <div
            ref={selfRef}
            className={className}
            style={style}
            onClick={onClick}
        >
            <PerspectiveCard>
                <div className="flex flex-col items-center bg-surface-20 rounded-xl hover:cursor-pointer">
                    <Image
                        src={`/apilocal/image/thumb/${filename}`}
                        alt={filename}
                        className="rounded-t-xl w-full h-auto"
                        width={width}
                        height={height}
                        onLoad={onLoad}
                        draggable={false}
                    />
                    <span className="px-2 max-w-[20ch] truncate">
                        {filename}
                    </span>
                </div>
            </PerspectiveCard>
        </div>
    );
};

export default ImageCard;
