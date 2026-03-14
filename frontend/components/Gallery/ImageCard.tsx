"use client";
import Image from "next/image";
import PerspectiveCard from "./PerspectiveCard";
import React, { useRef } from "react";
import { useRouter } from "next/navigation";
import ImageMetadata from "@/lib/image";

interface Props {
    image: ImageMetadata;
    className?: string;
    style?: React.CSSProperties;
}

const ImageCard: React.FC<Props> = ({ image, className, style }) => {
    const API_LOCAL = process.env.NEXT_PUBLIC_API_LOCAL;
    const { filename, format, width, height } = image;
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
            className={className + " opacity-0 "}
            style={style}
            onClick={onClick}
        >
            <PerspectiveCard>
                <div className="flex flex-col items-center bg-surface-20 rounded-xl hover:cursor-pointer">
                    <Image
                        src={`${API_LOCAL}/image/thumb/${filename + "." + format}`}
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
