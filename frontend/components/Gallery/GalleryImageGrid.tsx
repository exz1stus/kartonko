"use client";

import { ReactNode } from "react";
import { ImageMetadata } from "@/lib/api/generated/model";
import Masonry, { MasonryItem } from "@/components/template/Masonry";
import ImageCard from "./ImageCard";

interface Props {
    images: ImageMetadata[];
    getKey?: (image: ImageMetadata, index: number) => string;
    renderAction?: (image: ImageMetadata, index: number) => ReactNode;
    onImageClick?: (image: ImageMetadata, index: number) => void;
    actionClickThrough?: boolean;
    edgeToEdge?: boolean;
}

export default function GalleryImageGrid({ images, getKey, renderAction, onImageClick, actionClickThrough, edgeToEdge = false }: Props) {
    const masonryItems: MasonryItem[] = images.map((image, index) => ({
        key: getKey?.(image, index) ?? (image.id > 0 ? String(image.id) : `${image.filename}-${index}`),
        ratio: image.height / image.width,
        // ImageCard's filename is one line with leading-6.
        extraHeightPx: 24,
        item: <ImageCard image={image} action={renderAction?.(image, index)} onClick={onImageClick ? () => onImageClick(image, index) : undefined} actionClickThrough={actionClickThrough} />,
    }));

    return <div className={edgeToEdge ? "flex w-full p-3 sm:p-4 lg:p-6 2xl:p-8" : "flex justify-center p-4"}>
        <Masonry items={masonryItems} colWidthPx={180} maxColWidthPx={edgeToEdge ? 0 : 320} minCols={2} />
    </div>;
}
