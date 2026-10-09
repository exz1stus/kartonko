"use client";

import { ReactNode, Ref } from "react";
import { ImageMetadata } from "@/lib/api/generated/model";
import GalleryImageGrid from "./GalleryImageGrid";

interface Props {
    images: ImageMetadata[];
    loading: boolean;
    reachedEnd: boolean;
    error: string | null;
    retry: () => void;
    sentinelRef: Ref<HTMLDivElement>;
    emptyMessage?: string;
    renderAction?: (image: ImageMetadata, index: number) => ReactNode;
    onImageClick?: (image: ImageMetadata, index: number) => void;
    actionClickThrough?: boolean;
    edgeToEdge?: boolean;
}

export default function InfiniteImageGrid({
    images, loading, reachedEnd, error, retry, sentinelRef, emptyMessage = "No images match this search.",
    renderAction, onImageClick, actionClickThrough, edgeToEdge,
}: Props) {
    return <>
        <GalleryImageGrid images={images} renderAction={renderAction} onImageClick={onImageClick} actionClickThrough={actionClickThrough} edgeToEdge={edgeToEdge} />
        {!loading && !error && images.length === 0 && <p className="p-6 text-center text-primary-0/70">{emptyMessage}</p>}
        {error && <p role="alert" className="p-3 text-center text-red-500">{error} <button onClick={retry} className="underline">Retry</button></p>}
        {loading && <p role="status" className="p-3 text-center text-primary-0/70">Loading images…</p>}
        {!reachedEnd && !error && <div ref={sentinelRef} className="h-1" />}
    </>;
}
