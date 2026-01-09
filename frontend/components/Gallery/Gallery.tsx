"use client";
import React, { useEffect, useRef, useState } from "react";
import ImageCard from "./ImageCard";
import { SearchQuery, ImageSearch, isQueryEmpty, isQueriesEqual } from "./ImageSearch";
import { useDebounce } from "use-debounce";
import Scrollbar from "@/components/template/Scrollbar";
import DragDropZone from "@/components/PostImage/DragDropZone";
import useUploadModal from "@/hooks/useUploadModal";
import UploadModal from "@/components/PostImage/UploadModal";
import { useHover } from "@/app/contexts/HoverContex";
import Masonry, { MasonryItem } from "@/components/template/Masonry";
import { useInView } from "react-intersection-observer";
import ImageMetadata from "@/app/lib/image";

interface ImagesResponse {
    imageData: ImageMetadata[];
}

interface Props {
    initialImages: ImageMetadata[];
    initReachedEnd: boolean;
    initialQuery: SearchQuery;
}

const Gallery: React.FC<Props> = ({ initialImages, initReachedEnd, initialQuery }) => {
    const API_ORIGIN = process.env.NEXT_PUBLIC_API_ORIGIN;
    const REQUEST_SIZE = 30;
    const INIT_REQUEST_SIZE = 50;

    const [images, setImages] = useState<ImageMetadata[]>(initialImages);
    const [cursor, setCursor] = useState<number>(initialImages.length);
    const [reachedEnd, setReachedEnd] = useState<boolean>(initReachedEnd);
    const [searchQuery, setSearchQuery] = useState<SearchQuery>(initialQuery);
    const loading = useRef<boolean>(false);
    const hasQueryChangedFromInit = useRef<boolean>(false);
    const needInitalFetch = useRef<boolean>(initialImages.length === 0);

    const FetchCount = useRef<number>(0);

    const galleryHover = useHover<HTMLDivElement>();

    const [debouncedQuery] = useDebounce(searchQuery, 200);

    const handleSearch = (query: SearchQuery) => {
        setSearchQuery(query);
        setReachedEnd(false);
    };

    const { recievedImages, handleClose, handleDroppedFiles, handleUploaded } = useUploadModal();

    const fetchImages = async (searchQuery: SearchQuery, cursor: number, requestSize: number) => {
        if (loading.current || reachedEnd) return;
        loading.current = true;
        try {
            const nameQueryString =
                searchQuery?.nameContains.length === 0 ? "" : `name=${searchQuery.nameContains}&`;
            const tagsQueryString =
                searchQuery?.withTags.length === 0
                    ? ""
                    : `tags=${JSON.stringify(searchQuery.withTags)}&`;
            const response = await fetch(
                `${API_ORIGIN}/images?${nameQueryString}${tagsQueryString}cursor=${cursor}&limit=${requestSize}`
            );
            if (!response.ok) throw new Error("Failed to fetch images");

            const parsedImages: ImagesResponse = await response.json();
            if (parsedImages.imageData.length < requestSize) setReachedEnd(true);

            setImages((prev) => [...prev, ...parsedImages.imageData]);
            console.log(`fetching cursor ${cursor} limit ${requestSize}`);
        } catch (error) {
            console.error("Error fetching images:", error);
        } finally {
            loading.current = false;
        }
        FetchCount.current++;
    };

    useEffect(() => {
        if (cursor === 0 || reachedEnd) return;
        fetchImages(debouncedQuery, cursor, REQUEST_SIZE);
    }, [cursor]);

    const initialFetch = () => {
        fetchImages(initialQuery, cursor, INIT_REQUEST_SIZE);
    };

    const canFetch = () => {
        if (!isQueryEmpty(debouncedQuery) && !isQueriesEqual(debouncedQuery, initialQuery))
            hasQueryChangedFromInit.current = true;

        return hasQueryChangedFromInit.current;
    };

    useEffect(() => {
        if (needInitalFetch.current) {
            initialFetch();
            needInitalFetch.current = false;
        }
    }, []);

    useEffect(() => {
        if (!canFetch()) return;

        setImages([]);
        setCursor(0);
        fetchImages(debouncedQuery, 0, REQUEST_SIZE);
    }, [debouncedQuery]);

    const items: MasonryItem[] = images.map((image) => ({
        key: image.filename,
        ratio: image.height / image.width,
        item: (
            <ImageCard
                filename={image.filename}
                tags={image.tags}
                width={image.width}
                height={image.height}
            />
        ),
    }));

    const gallery = (
        <div className="flex justify-center">
            <Masonry className="p-4" items={items} colWidth={240} maxCols={0} />
        </div>
    );

    const isUploadModalShown = recievedImages.length > 0;

    const { ref: sentinelRef, inView } = useInView({
        threshold: 0,
        rootMargin: "400px",
    });

    useEffect(() => {
        if (inView && !loading.current && !reachedEnd) {
            setCursor((prev) => prev + REQUEST_SIZE);
        }
    }, [inView, reachedEnd]);

    return (
        <div className="flex flex-col h-full">
            {/* <div>Client Fetches count: {FetchCount.current} Images count: {images.length}</div> */}
            <ImageSearch
                initialQuery={initialQuery}
                selected={galleryHover.isHovered() && !isUploadModalShown}
                onQueryChange={handleSearch}
                className="flex-shrink-0 p-4 border-b"
            />
            <div ref={galleryHover.ref} className="flex-1 overflow-hidden">
                <DragDropZone onFilesDropped={handleDroppedFiles}>
                    <UploadModal
                        images={recievedImages}
                        onClose={handleClose}
                        onUploaded={handleUploaded}
                    />
                    <Scrollbar className="overflow-x-hidden">
                        {gallery}
                        <div ref={sentinelRef} style={{ height: 1 }} />
                        <div
                            className={`${
                                !loading.current && items.length > 0 ? "border-t" : ""
                            } flex justify-center p-4`}
                        >
                            {loading.current && <div>Loading...</div>}
                            {reachedEnd && <div>End of images</div>}
                        </div>
                    </Scrollbar>
                </DragDropZone>
            </div>
        </div>
    );
};

export default Gallery;
