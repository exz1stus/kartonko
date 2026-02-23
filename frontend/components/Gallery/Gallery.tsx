"use client";
import React, { useEffect, useRef, useState } from "react";
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
import { useInView } from "react-intersection-observer";
import ImageMetadata from "@/lib/image";
import useUpload from "@/hooks/useUpload";
import { apiFetch } from "@/lib/apiFetch";

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
    const REQUEST_SIZE = 30;
    const INIT_REQUEST_SIZE = 50;

    const [images, setImages] = useState<ImageMetadata[]>(initialImages);
    const [reachedEnd, setReachedEnd] = useState<boolean>(initReachedEnd);
    const [searchQuery, setSearchQuery] = useState<SearchQuery>(initialQuery);
    const [loading, setLoading] = useState(false);
    const hasQueryChangedFromInit = useRef<boolean>(false);
    const fetchingRef = useRef(false);

    const FetchCount = useRef<number>(0);

    const galleryHover = useHover<HTMLDivElement>();

    const [debouncedQuery] = useDebounce(searchQuery, 200);

    const handleSearch = (query: SearchQuery) => {
        setSearchQuery(query);
    };

    const { handleUploadFiles } = useUpload();

    const fetchImages = async (
        searchQuery: SearchQuery,
        cursor: number,
        requestSize: number,
    ) => {
        if (fetchingRef.current || reachedEnd) return;
        fetchingRef.current = true;
        setLoading(true);

        try {
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
            setImages((prev) => [...prev, ...imageData]);
            if (imageData.length < requestSize) setReachedEnd(true);

            console.log(`fetching cursor ${cursor} limit ${requestSize}`);
        } catch (error) {
            console.error("Error fetching images:", error);
        } finally {
            fetchingRef.current = false;
            setLoading(false);
        }
        FetchCount.current++;
    };

    const canFetch = () => {
        if (
            !isQueryEmpty(debouncedQuery) &&
            !isQueriesEqual(debouncedQuery, initialQuery)
        )
            hasQueryChangedFromInit.current = true;

        return hasQueryChangedFromInit.current;
    };

    useEffect(() => {
        if (initialImages.length > 0) return;
        fetchImages(initialQuery, images.length, INIT_REQUEST_SIZE);
    }, []);

    useEffect(() => {
        if (!canFetch()) return;

        setImages([]);
        setReachedEnd(false);
        fetchImages(debouncedQuery, 0, REQUEST_SIZE);
    }, [debouncedQuery]);

    const items: MasonryItem[] = images.map((image) => ({
        key: image.filename,
        ratio: image.height / image.width,
        item: <ImageCard image={image} />,
    }));

    const { ref: sentinelRef, inView } = useInView({
        threshold: 0,
        rootMargin: "400px",
    });

    const contentRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        const container = galleryHover.ref.current;
        const content = contentRef.current;

        if (!container || !content) return;

        const observer = new ResizeObserver(() => {
            if (fetchingRef.current || reachedEnd) return;

            if (content.scrollHeight <= container.clientHeight) {
                fetchImages(debouncedQuery, images.length, REQUEST_SIZE);
            }
        });

        observer.observe(content);

        return () => observer.disconnect();
    }, [debouncedQuery, loading, reachedEnd, images.length]);

    useEffect(() => {
        if (!inView || loading || reachedEnd) return;

        fetchImages(debouncedQuery, images.length, REQUEST_SIZE);
    }, [inView]);

    const content = (
        <>
            <div ref={contentRef}>
                <div className="flex justify-center">
                    <Masonry className="p-4" items={items} colWidth={240} />
                </div>
            </div>
            <div ref={sentinelRef} style={{ height: 1 }} />
            <div
                className={`${
                    reachedEnd && !loading && items.length > 0 ? "border-t" : ""
                } flex justify-center p-4`}
            >
                {loading && <div>Loading...</div>}
                {reachedEnd && <div>End of images</div>}
            </div>
        </>
    );

    return (
        <div className="flex flex-col h-full">
            {/* <div>Client Fetches count: {FetchCount.current} Images count: {images.length}</div> */}
            <ImageSearch
                initialQuery={initialQuery}
                selected={galleryHover.isHovered()}
                onQueryChange={handleSearch}
                className="flex-shrink-0 p-4 border-b"
            />
            <div ref={galleryHover.ref} className="flex-1 overflow-hidden">
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
