"use client";
import React, { useEffect, useState } from 'react'
import { ImageData, ImageCard } from './ImageCard'
import { SearchQuery, ImageSearch } from './ImageSearch';
import { useDebounce, useDebouncedValue } from '@/app/useDebounce';
import PostImageManager from '../PostImage/PostImageManager';
import Scrollbar from '../template/Scrollbar';

interface ImagesResponse {
    imageData: ImageData[]
    end: boolean
}

const EMPTY_QUERY: SearchQuery = {
    nameContains: "",
    withTags: []
}

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
        <Scrollbar onScroll={handleGalleryScroll}>
            <PostImageManager />
            <div
                className=" p-4"
                onMouseEnter={() => setIsHovered(true)}
                onMouseLeave={() => setIsHovered(false)}
            >
                <ImageSearch isHovered={isHovered} onQueryChange={handleSearch} />
                <div className="flex flex-wrap justify-evenly gap-5">
                    {ImageCards}
                </div>
            </div>
        </Scrollbar>
    )
}

export default Gallery

