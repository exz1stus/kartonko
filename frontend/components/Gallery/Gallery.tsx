"use client";
import React, { useCallback, useEffect, useRef, useState } from 'react'
import { ImageData, ImageCard } from './ImageCard'
import { SearchQuery, ImageSearch } from './ImageSearch';
import { useDebounce, useDebouncedValue } from '@/app/hooks/useDebounce';
import Scrollbar from '../template/Scrollbar';
import DragDropZone from '@/components/PostImage/DragDropZone';
import useUploadModal from '../PostImage/useUploadModal';
import UploadModal from '../PostImage/UploadModal';
import { useHover } from '@/app/contexts/HoverContex';
import useFullscreen from '@/app/hooks/useFullscreen';

const REQUEST_SIZE = 30;
const EMPTY_QUERY: SearchQuery = {
    nameContains: "",
    withTags: []
}

interface ImagesResponse {
    imageData: ImageData[]
}

interface Props {
    initialImages: ImageData[]
}

//TODO : search to searchComponent + useSearch
const Gallery: React.FC<Props> = ({ initialImages }) => {
    const [images, setImages] = useState<ImageData[]>([]);
    const [cursor, setCursor] = useState<number>(0);
    const [reachedEnd, setReachedEnd] = useState<boolean>(false);
    const [searchQuery, setSearchQuery] = useState<SearchQuery>(EMPTY_QUERY);
    const loading = useRef<boolean>(false);
    const firstRender = useRef(true);

    const galleryHover = useHover<HTMLDivElement>();
    const searchHover = useHover<HTMLDivElement>();

    const API_ORIGIN = process.env.NEXT_PUBLIC_API_ORIGIN;
    const debouncedQuery = useDebouncedValue(searchQuery, 200);

    const handleSearch = (query: SearchQuery) => {
        setCursor(0);
        setImages([]);
        setReachedEnd(false);
        setSearchQuery(query);
    };

    const { recievedImages, handleClose, handleDroppedFiles, handleUploaded } = useUploadModal();

    const isQueryEmpty = (query: SearchQuery) =>
        query.nameContains === "" && query.withTags.length === 0;

    const fetchImages = async (searchQuery: SearchQuery, cursor: number) => {
        if (loading.current || reachedEnd) return;
        loading.current = true;

        try {
            const queryString = isQueryEmpty(debouncedQuery) ? "" : `query=${JSON.stringify(searchQuery)}&`;
            const response = await fetch(`${API_ORIGIN}/images?${queryString}&cursor=${cursor}&limit=${REQUEST_SIZE}`);
            if (!response.ok) throw new Error("Failed to fetch images");

            const parsedImages: ImagesResponse = await response.json();
            if (parsedImages.imageData.length < REQUEST_SIZE) setReachedEnd(true);

            setImages((prev) => [...prev, ...parsedImages.imageData]);
        }
        catch (error) {
            console.error("Error fetching images:", error);
        } finally {
            loading.current = false;
        }
    };

    useEffect(() => {
        setImages(initialImages);
        setCursor(prev => prev + initialImages.length);
    }, []);

    useEffect(() => {
        if (cursor === 0 && isQueryEmpty(debouncedQuery) && firstRender.current) return;
        firstRender.current = false;
        fetchImages(debouncedQuery, cursor);
    }, [cursor, debouncedQuery]);

    const handleGalleryScroll = useCallback(useDebounce(() => {
        if (loading.current || reachedEnd) return;
        const scrollPosition = window.innerHeight + window.scrollY;
        const threshold = document.body.offsetHeight * 0.8;
        if (scrollPosition >= threshold) {
            setCursor(prev => prev + REQUEST_SIZE);
        }
    }, 100), [reachedEnd]);

    const imageCards = images.map((image, index) => (
        <ImageCard key={index} filename={image.filename} tags={image.tags} />
    ));

    const isFullscreen = useFullscreen();

    const gallery = !isFullscreen ? (
        <div className="p-4">
            <div className="flex flex-wrap justify-evenly gap-5">
                {imageCards}
            </div>
        </div>
    ) : (
        <div className="grid grid-cols-12 w-full">
            {imageCards}
        </div>
    )

    return (
        <div className="flex flex-col h-full">
            <div ref={searchHover.ref} className="flex-shrink-0 p-4 border-b">
                <ImageSearch isGalleryHovered={galleryHover.isHovered()} isSearchHovered={searchHover.isHovered()} onQueryChange={handleSearch} />
            </div>
            <UploadModal images={recievedImages} onClose={handleClose} onUploaded={handleUploaded} />
            <div ref={galleryHover.ref} className="flex-1 overflow-hidden">
                <DragDropZone onFilesDropped={handleDroppedFiles}>
                    <Scrollbar onScroll={handleGalleryScroll}>
                        {gallery}
                        <div className={`${imageCards.length > 0 ? "border-t" : ""} flex justify-center p-4`}>
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

