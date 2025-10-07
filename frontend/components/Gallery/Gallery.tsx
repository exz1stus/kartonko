"use client";
import React, { useEffect, useState } from 'react'
import { ImageData, ImageCard } from './ImageCard'
import { SearchQuery, ImageSearch } from './ImageSearch';
import { useDebounce, useDebouncedValue } from '@/app/hooks/useDebounce';
import Scrollbar from '../template/Scrollbar';
import DragDropZone from '@/components/PostImage/DragDropZone';
import useUploadModal from '../PostImage/useUploadModal';
import UploadModal from '../PostImage/UploadModal';

interface ImagesResponse {
    imageData: ImageData[]
    end: boolean
}

const EMPTY_QUERY: SearchQuery = {
    nameContains: "",
    withTags: []
}

//TODO : search to searchComponent + useSearch
const Gallery = () => {
    const [images, setImages] = useState<ImageData[]>([]);
    const [cursor, setCursor] = useState<number>(0);
    const [loading, setLoading] = useState<boolean>(false);
    const [reachedEnd, setReachedEnd] = useState<boolean>(false);
    const [searchQuery, setSearchQuery] = useState<SearchQuery>(EMPTY_QUERY);

    const [isHovered, setIsHovered] = useState<boolean>(false);

    const requestSize = 30;
    const API_ORIGIN = process.env.NEXT_PUBLIC_API_ORIGIN;

    const debouncedQuery = useDebouncedValue(searchQuery, 200);

    useEffect(() => {
        searchImages(debouncedQuery, 0);
    }, [debouncedQuery]);

    const handleSearch = (query: SearchQuery) => {
        setCursor(0);
        setImages([]);
        setSearchQuery(query);
    };

    const { recievedImages, handleClose, handleDroppedFiles, handleUploaded } = useUploadModal();

    const searchImages = async (searchQuery: SearchQuery, cursor: number) => {
        if (loading) return;
        setLoading(true);
        try {
            const response = await fetch(`${API_ORIGIN}/search-images?query=${JSON.stringify(searchQuery)}&cursor=${cursor}&limit=${requestSize}`);
            if (response.ok) {
                const parsedImages: ImagesResponse = await response.json();
                if (parsedImages.end) {
                    setReachedEnd(true);
                    return;
                }
                setImages((prev) => [...prev, ...parsedImages.imageData]);
                setCursor(cursor);
            }
        } catch (error) {
            console.error("Error fetching images:", error);
        } finally {
            setLoading(false);
            console.log(images.length);
        }
    };

    const handleGalleryScroll =
        useDebounce(() => {
            if (!reachedEnd && window.innerHeight + window.scrollY >= document.body.offsetHeight * 0.8) {
                searchImages(searchQuery, cursor + requestSize);
            }
        }, 100);

    const ImageCards = images?.length ? images.map((image, index) => (
        <ImageCard key={index} filename={image.filename} tags={image.tags} />
    )) : null;

    return (
        <div
            className="h-full"
            onMouseEnter={() => setIsHovered(true)}
            onMouseLeave={() => setIsHovered(false)}
        >
            <ImageSearch isHovered={isHovered} onQueryChange={handleSearch} />
            <DragDropZone onFilesDropped={handleDroppedFiles}>
                <Scrollbar onScroll={handleGalleryScroll}>
                    <UploadModal images={recievedImages} onClose={handleClose} onUploaded={handleUploaded} />
                    <div className=" p-4">
                        <div className="flex flex-wrap justify-evenly gap-5">
                            {ImageCards}
                        </div>
                    </div>
                </Scrollbar>
            </DragDropZone>
        </div>
    )
}

export default Gallery

