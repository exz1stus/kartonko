"use client";

import { PointerEvent, ReactNode } from "react";
import { ImageMetadata } from "@/lib/api/generated/model";
import Masonry, { MasonryItem } from "@/components/template/Masonry";
import ImageCard from "./ImageCard";
import { setActiveGalleryScope, useGallerySelection } from "@/hooks/useGallerySelection";
import { getPointerSelectionBounds, trackGalleryMarquee } from "@/lib/image/rangeSelection";

interface Props {
    images: ImageMetadata[];
    getKey?: (image: ImageMetadata, index: number) => string;
    renderAction?: (image: ImageMetadata, index: number) => ReactNode;
    onImageClick?: (image: ImageMetadata, index: number) => void;
    actionClickThrough?: boolean;
    edgeToEdge?: boolean;
    dragSourceBoardId?: number;
    selectionScope?: string;
}

export default function GalleryImageGrid({ images, getKey, renderAction, onImageClick, actionClickThrough, edgeToEdge = false, dragSourceBoardId, selectionScope = "home" }: Props) {
    const selectionMode = useGallerySelection(selectionScope, (state) => state.selectionMode);
    const beginSpatialSelection = useGallerySelection(selectionScope, (state) => state.beginSpatialSelection);
    const extendSpatialSelection = useGallerySelection(selectionScope, (state) => state.extendSpatialSelection);
    const endSelection = useGallerySelection(selectionScope, (state) => state.endSelection);
    const selectionOrder = images.map((image) => image.id);
    const masonryItems: MasonryItem[] = images.map((image, index) => ({
        key: getKey?.(image, index) ?? (image.id > 0 ? String(image.id) : `${image.filename}-${index}`),
        ratio: image.height / image.width,
        // ImageCard's filename is one line with leading-6.
        extraHeightPx: 24,
        item: <ImageCard image={image} selectionOrder={selectionOrder} action={renderAction?.(image, index)} onClick={onImageClick ? () => onImageClick(image, index) : undefined} actionClickThrough={actionClickThrough} dragSourceBoardId={dragSourceBoardId} selectionScope={selectionScope} />,
    }));

    const startBlankSelection = (event: PointerEvent<HTMLDivElement>) => {
        if (!selectionMode || event.button !== 0 || (event.target as HTMLElement).closest("[data-gallery-image-id]")) return;
        event.preventDefault();
        setActiveGalleryScope(selectionScope);
        const grid = event.currentTarget;
        const startX = event.clientX;
        const startY = event.clientY;
        const additive = event.ctrlKey || event.metaKey;
        const bounds = getPointerSelectionBounds(startX, startY, startX, startY);
        beginSpatialSelection([], -1, -1, additive, bounds, "select");
        trackGalleryMarquee(grid, startX, startY, (ids, nextBounds) => extendSpatialSelection(ids, nextBounds), endSelection);
    };

    return <div data-gallery-grid data-gallery-selection-scope={selectionScope} onPointerDown={startBlankSelection} className={edgeToEdge ? "flex w-full p-3 sm:p-4 lg:p-6 2xl:p-8" : "flex justify-center p-4"}>
        <Masonry items={masonryItems} colWidthPx={180} maxColWidthPx={edgeToEdge ? 0 : 320} minCols={2} />
    </div>;
}
