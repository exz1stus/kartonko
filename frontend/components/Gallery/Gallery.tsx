"use client";
import React, { useCallback, useState } from "react";
import { SearchQuery, ImageSearch } from "./ImageSearch";
import { useDebounce } from "use-debounce";
import Scrollbar from "@/components/template/Scrollbar";
import DragDropZone from "@/components/UploadImage/DragDropZone";
import InfiniteImageGrid from "@/components/Gallery/InfiniteImageGrid";
import { ImageMetadata } from "@/lib/api/generated/model";
import usePreUpload from "@/hooks/usePreUpload";
import useInfiniteScroll from "@/hooks/useInfiniteScroll";
import { ImageIcon } from "lucide-react";
import { getImagesByQuery } from "@/lib/api/generated/client";

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

    const handleDroppedFiles = usePreUpload();

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

    const { items, loading, reachedEnd, error, retry, sentinelRef } = useInfiniteScroll<
        SearchQuery,
        ImageMetadata
    >({
        fetchFn: fetchImages,
        query: debouncedQuery,
        initialItems: initialImages,
        initialReachedEnd: initReachedEnd,
    });

    const footer = (
        <div className="flex gap-3 px-5">
            <ImageIcon />
            Images: {items.length}
        </div>
    );

    return (
        <div className="flex flex-col h-full">
            <ImageSearch
                initialQuery={initialQuery}
                onQueryChange={(query: SearchQuery) => setSearchQuery(query)}
                className="shrink-0 border-b border-surface-20 p-4"
            />
            <div className="min-h-0 flex-1 overflow-hidden">
                <DragDropZone onFilesDropped={handleDroppedFiles}>
                    <div className="flex flex-col h-full">
                        <Scrollbar className="min-h-0 overflow-x-hidden overscroll-contain">
                            <InfiniteImageGrid images={items} loading={loading} reachedEnd={reachedEnd} error={error} retry={retry} sentinelRef={sentinelRef} />
                            {reachedEnd && footer}
                        </Scrollbar>
                    </div>
                </DragDropZone>
            </div>
        </div>
    );
};

export default Gallery;
