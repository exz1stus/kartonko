"use client";
import React, { useCallback, useEffect, useState } from "react";
import { SearchQuery, ImageSearch } from "./ImageSearch";
import { useDebounce } from "use-debounce";
import Scrollbar from "@/components/template/Scrollbar";
import InfiniteImageGrid from "@/components/Gallery/InfiniteImageGrid";
import { ImageMetadata } from "@/lib/api/generated/model";
import useInfiniteScroll from "@/hooks/useInfiniteScroll";
import { ImageIcon } from "lucide-react";
import { getImagesByQuery } from "@/lib/api/generated/client";
import GallerySelectionControls from "@/components/Gallery/GallerySelectionControls";
import { imagesDeletedEvent } from "@/lib/image/deletedImages";

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
    const [searchQuery, setSearchQuery] = useState<SearchQuery>(initialQuery);
    const [debouncedQuery] = useDebounce(searchQuery, 200);

    const fetchImages = useCallback(
        async (
            searchQuery: SearchQuery,
            cursor: number,
            requestSize: number,
        ): Promise<ImageMetadata[]> => {
            return await getImagesByQuery({
                ...searchQuery,
                semantic: true,
                cursor: cursor,
                limit: requestSize,
            });
        },
        [],
    );

    const { items, loading, reachedEnd, error, retry, removeItems, sentinelRef } = useInfiniteScroll<
        SearchQuery,
        ImageMetadata
    >({
        fetchFn: fetchImages,
        query: debouncedQuery,
        initialItems: initialImages,
        initialReachedEnd: initReachedEnd,
    });

    useEffect(() => {
        const onImagesDeleted = (event: Event) => {
            const deleted = new Set((event as CustomEvent<number[]>).detail);
            if (deleted.size) removeItems((image) => deleted.has(image.id));
        };
        window.addEventListener(imagesDeletedEvent, onImagesDeleted);
        return () => window.removeEventListener(imagesDeletedEvent, onImagesDeleted);
    }, [removeItems]);

    const footer = (
        <div className="flex gap-3 px-5">
            <ImageIcon />
            Images: {items.length}
        </div>
    );

    return (
        <div className="flex flex-col h-full">
            <div className="flex shrink-0 items-center border-b border-surface-20">
                <ImageSearch
                    initialQuery={initialQuery}
                    onQueryChange={(query: SearchQuery) => setSearchQuery(query)}
                    className="min-w-0 flex-1 p-4"
                />
                <GallerySelectionControls scope="home" images={items} />
            </div>
            <div className="min-h-0 flex-1 overflow-hidden">
                    <div className="flex flex-col h-full">
                        <Scrollbar className="min-h-0 overflow-x-hidden overscroll-contain">
                            <InfiniteImageGrid images={items} loading={loading} reachedEnd={reachedEnd} error={error} retry={retry} sentinelRef={sentinelRef} dropIntent={{ tags: searchQuery.tags }} />
                            {reachedEnd && footer}
                        </Scrollbar>
                    </div>
            </div>
        </div>
    );
};

export default Gallery;
