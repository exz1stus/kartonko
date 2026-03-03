"use client";
import React, { useCallback, useEffect, useRef, useState } from "react";
import ImageCard from "@/components/Gallery/ImageCard";
import {
    SearchQuery,
    ImageSearch,
    isQueryEmpty,
    isQueriesEqual,
} from "./ImageSearch";
import { useDebounce } from "use-debounce";
import Scrollbar from "@/components/template/Scrollbar";
import DragDropZone from "@/components/PostImage/DragDropZone";
import { useHover } from "@/contexts/HoverContex";
import Masonry, { MasonryItem } from "@/components/template/Masonry";
import ImageMetadata from "@/lib/image";
import useUpload from "@/hooks/useUpload";
import { apiFetch } from "@/lib/apiFetch";
import useInfiniteScroll from "@/hooks/useInfiniteScroll";

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
    const galleryHover = useHover<HTMLDivElement>();
    const [searchQuery, setSearchQuery] = useState<SearchQuery>(initialQuery);
    const [debouncedQuery] = useDebounce(searchQuery, 200);

    const { handleUploadFiles } = useUpload();

    const fetchImages = useCallback(
        async (
            searchQuery: SearchQuery,
            cursor: number,
            requestSize: number,
        ): Promise<ImageMetadata[]> => {
            const name =
                searchQuery?.nameContains.length === 0
                    ? ""
                    : `name=${searchQuery.nameContains}&`;
            const tags =
                searchQuery?.withTags.length === 0
                    ? ""
                    : `tags=${JSON.stringify(searchQuery.withTags)}&`;
            const response = await apiFetch(
                `/images?${name}${tags}cursor=${cursor}&limit=${requestSize}`,
            );
            if (!response.ok) throw new Error("Failed to fetch images");

            const imageData: ImageMetadata[] = await response.json();
            console.log(`fetching cursor ${cursor} limit ${requestSize}`);
            return imageData;
        },
        [],
    );

    const {
        items,
        loading,
        reachedEnd,
        sentinelRef,
        containerRef,
        contentRef,
        debugEvents,
    } = useInfiniteScroll<SearchQuery, ImageMetadata>({
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

    const content = (
        <div ref={galleryHover.ref}>
            <div ref={contentRef} className="flex justify-center">
                <Masonry className="p-4" items={masonryItems} colWidth={240} />
            </div>
            <div ref={sentinelRef} style={{ height: 1 }} />
            <div
                className={`${
                    reachedEnd && !loading && masonryItems.length > 0
                        ? "border-t"
                        : ""
                } flex justify-center p-4`}
            >
                {loading && <div>Loading...</div>}
                {reachedEnd && <div>End of images</div>}
            </div>
        </div>
    );

    return (
        <div className="flex flex-col h-full">
            {false && (
                <div className="right-[50%] z-99 fixed bg-green-400 m-2 p-2 inset">
                    Images count : {items.length}
                    <div className="flex flex-col max-h-50 overflow-y-auto scroll-auto">
                        {debugEvents.current.map((event, i) => (
                            <span key={i}>{event}</span>
                        ))}
                    </div>
                </div>
            )}
            <ImageSearch
                initialQuery={initialQuery}
                selected={galleryHover.isHovered()}
                onQueryChange={(query: SearchQuery) => setSearchQuery(query)}
                className="p-4 border-b shrink-0"
            />
            <div ref={containerRef} className="flex-1 overflow-hidden">
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
