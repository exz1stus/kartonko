"use client";
import React, { useCallback, useRef, useState } from "react";
import ImageCard from "@/components/Gallery/ImageCard";
import { SearchQuery, ImageSearch, isQueryEmpty } from "./ImageSearch";
import { useDebounce } from "use-debounce";
import Scrollbar from "@/components/template/Scrollbar";
import DragDropZone from "@/components/UploadImage/DragDropZone";
import { useHover } from "@/contexts/HoverContex";
import Masonry, { MasonryItem } from "@/components/template/Masonry";
import ImageMetadata from "@/lib/image";
import useUpload from "@/hooks/useUpload";
import { apiFetch } from "@/lib/apiFetch";
import useInfiniteScroll from "@/hooks/useInfiniteScroll";
import { constructQueryString } from "@/lib/query";
import Loading from "../Loading";
import { ImageIcon } from "lucide-react";

interface Props {
    initialImages: ImageMetadata[];
    initReachedEnd: boolean;
    initialQuery: SearchQuery;
}

const Gallery: React.FC<Props> = ({
    initialImages,
    initReachedEnd,
    initialQuery,
}) => {
    const galleryRef = useRef<HTMLDivElement>(null);
    const [searchQuery, setSearchQuery] = useState<SearchQuery>(initialQuery);
    const [debouncedQuery] = useDebounce(searchQuery, 200);

    const { handleUploadFiles } = useUpload();

    const fetchImages = useCallback(
        async (
            searchQuery: SearchQuery,
            cursor: number,
            requestSize: number,
        ): Promise<ImageMetadata[]> => {
            const queryString = constructQueryString(searchQuery);
            const response = await apiFetch(
                `/image?${queryString}cursor=${cursor}&limit=${requestSize}`,
            );
            if (!response.ok) throw new Error("Failed to fetch images");

            const imageData: ImageMetadata[] = await response.json();
            return imageData;
        },
        [],
    );

    const { items, loading, reachedEnd, sentinelRef } = useInfiniteScroll<
        SearchQuery,
        ImageMetadata
    >({
        fetchFn: fetchImages,
        query: debouncedQuery,
        initialItems: initialImages,
        isQueryEmpty: isQueryEmpty,
        initialReachedEnd: initReachedEnd,
    });

    const masonryItems: MasonryItem[] = items.map((image) => ({
        key: image.filename,
        ratio: image.height / image.width,
        item: <ImageCard image={image} />,
    }));

    const footer = (
        <div className="flex gap-3 w-full">
            <ImageIcon />
            Images: {items.length}
        </div>
    );

    const content = (
        <div ref={galleryRef}>
            <div className="flex justify-center grow">
                <Masonry
                    className="p-4"
                    items={masonryItems}
                    colWidthPx={200}
                />
            </div>
            {!reachedEnd && <div ref={sentinelRef} />}
            <div
                className={`${
                    reachedEnd && !loading && masonryItems.length > 0
                        ? "border-t"
                        : ""
                } flex  justify-center p-4`}
            >
                {loading && <Loading />}
                {reachedEnd && footer}
            </div>
        </div>
    );

    return (
        <div className="flex flex-col h-full">
            <ImageSearch
                initialQuery={initialQuery}
                onQueryChange={(query: SearchQuery) => setSearchQuery(query)}
                className="p-4 border-primary-0 border-b shrink-0"
            />
            <div className="flex-1 overflow-hidden">
                <DragDropZone onFilesDropped={handleUploadFiles}>
                    <Scrollbar className="overflow-x-hidden">
                        {content}
                    </Scrollbar>
                </DragDropZone>
            </div>
        </div>
    );
};

export default Gallery;
