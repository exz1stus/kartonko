"use client";
import React, { useCallback, useEffect, useRef, useState } from 'react'
import { ImageData, ImageCard } from './ImageCard'
import { SearchQuery, ImageSearch, isQueryEmpty, isQueriesEqual } from './ImageSearch';
import { useDebounce, useDebouncedCallback } from 'use-debounce';
import Scrollbar from '../template/Scrollbar';
import DragDropZone from '@/components/PostImage/DragDropZone';
import useUploadModal from '../PostImage/useUploadModal';
import UploadModal from '../PostImage/UploadModal';
import { useHover } from '@/app/contexts/HoverContex';
import Masonry from '../template/Masonry';

interface ImagesResponse {
    imageData: ImageData[]
}

interface Props {
    initialImages: ImageData[]
    initReachedEnd: boolean
    initialQuery: SearchQuery
}

const Gallery: React.FC<Props> = ({ initialImages, initReachedEnd, initialQuery }) => {
    const API_ORIGIN = process.env.NEXT_PUBLIC_API_ORIGIN;
    const REQUEST_SIZE = 30;
    const INIT_REQUEST_SIZE = 50;

    const [images, setImages] = useState<ImageData[]>(initialImages);
    const [cursor, setCursor] = useState<number>(initialImages.length);
    const [reachedEnd, setReachedEnd] = useState<boolean>(initReachedEnd);
    const [searchQuery, setSearchQuery] = useState<SearchQuery>(initialQuery);
    const loading = useRef<boolean>(false);
    const hasQueryChangedFromInit = useRef<boolean>(false);
    const hasScrolled = useRef<boolean>(false);
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
            const nameQueryString = searchQuery?.nameContains.length === 0 ? "" : `name=${searchQuery.nameContains}&`;
            const tagsQueryString = searchQuery?.withTags.length === 0 ? "" : `tags=${JSON.stringify(searchQuery.withTags)}&`;
            const response = await fetch(`${API_ORIGIN}/images?${nameQueryString}${tagsQueryString}cursor=${cursor}&limit=${requestSize}`);
            if (!response.ok) throw new Error("Failed to fetch images");

            const parsedImages: ImagesResponse = await response.json();
            if (parsedImages.imageData.length < requestSize) setReachedEnd(true);

            setImages((prev) => [...prev, ...parsedImages.imageData]);
            console.log(`fetching cursor ${cursor} limit ${requestSize}`);
        }
        catch (error) {
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
    }

    const canFetch = () => {
        if (!isQueryEmpty(debouncedQuery) && !isQueriesEqual(debouncedQuery, initialQuery))
            hasQueryChangedFromInit.current = true;

        if (!hasQueryChangedFromInit.current) return false;

        return true;
    }

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

    const handleGalleryScroll = useCallback(useDebouncedCallback(() => {
        if (loading.current || reachedEnd) return;
        const scrollPosition = window.innerHeight + window.scrollY;
        const threshold = document.body.offsetHeight * 0.8;
        if (scrollPosition >= threshold) {
            if (!hasScrolled.current) hasScrolled.current = true;
            setCursor(prev => prev + REQUEST_SIZE);
        }
    }, 100), [reachedEnd]);

    const imageCards = images.map((image) => (
        <ImageCard
            key={image.filename}
            filename={image.filename}
            tags={image.tags}
            className="break-inside-avoid"
        />
    ));

    const gallery = (
        <div className="flex justify-center">
            <Masonry className="p-4" >
                {imageCards}
            </Masonry>
        </div>
    );

    const isUploadModalShown = recievedImages.length > 0;

    return (
        <div className="flex flex-col h-full">
            {/* {<div>Client Fetches count: {FetchCount.current} Images count: {images.length}</div>} */}
            <ImageSearch
                initialQuery={initialQuery}
                selected={galleryHover.isHovered() && !isUploadModalShown}
                onQueryChange={handleSearch}
                className="flex-shrink-0 p-4 border-b"
            />
            <div ref={galleryHover.ref} className="flex-1 overflow-hidden">
                <DragDropZone onFilesDropped={handleDroppedFiles}>
                    <UploadModal images={recievedImages} onClose={handleClose} onUploaded={handleUploaded} />
                    <Scrollbar onScroll={handleGalleryScroll} className="overflow-x-hidden">
                        {gallery}
                        <div className={`${!loading.current && imageCards.length > 0 ? "border-t" : ""} flex justify-center p-4`}>
                            {loading.current && <div>Loading...</div>}
                            {reachedEnd && <div>End of images</div>}
                        </div>
                    </Scrollbar>
                </DragDropZone>
            </div>
        </div >
    )
}

export default Gallery

