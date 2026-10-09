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
    action?: React.ReactNode;
    onClick?: () => void;
    actionClickThrough?: boolean;
}

const ImageCard: React.FC<Props> = ({ image, className, style, action, onClick, actionClickThrough = false }) => {
    const { filename, hash, width, height } = image;
    const selfRef = useRef<HTMLDivElement>(null);
    const router = useRouter();
    const openImage = () => {
        router.push(`/image/${image.filename}`);
    };

    const onLoad = () => {
        selfRef.current?.classList.remove("opacity-0");
        selfRef.current?.clientHeight;
    };

    return (
        <div
            ref={selfRef}
            className={`relative ${className ?? ""}`}
            style={style}
            onClick={onClick ?? openImage}
            role="button"
            tabIndex={0}
            onKeyDown={(event) => {
                if (event.key === "Enter" || event.key === " ") {
                    event.preventDefault();
                    (onClick ?? openImage)();
                }
            }}
        >
            <PerspectiveCard>
                <div className="flex flex-col items-center bg-surface-20 rounded-xl hover:cursor-pointer">
                    <Image
                        src={`/apilocal/image/hash/${encodeURIComponent(hash)}/thumb`}
                        alt={filename}
                        className="rounded-t-xl w-full h-auto"
                        width={width}
                        height={height}
                        onLoad={onLoad}
                        draggable={false}
                    />
                    <span className="px-2 max-w-[20ch] truncate leading-6">
                        {filename}
                    </span>
                </div>
            </PerspectiveCard>
            {action && <div className="absolute right-2 top-2 z-20" onClick={actionClickThrough ? undefined : (event) => event.stopPropagation()} onKeyDown={actionClickThrough ? undefined : (event) => event.stopPropagation()}>{action}</div>}
        </div>
    );
};

export default ImageCard;
