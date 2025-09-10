"use client";
import React, { useEffect, useState } from 'react'
import { ImageData, ImageCard } from './ImageCard'
import { SearchQuery, SearchOverlay } from './SearchOverlay';
import { useDebounce, useDebouncedValue } from '@/app/useDebounce';

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

    const handleScroll =
        useDebounce(() => {
            if (!reachedEnd && window.innerHeight + window.scrollY >= document.body.offsetHeight * 0.9) {
                searchImages(searchQuery, cursor + requestSize);
            }
        }, 100);

    useEffect(() => {
        window.addEventListener("scroll", handleScroll);
        return () => {
            window.removeEventListener("scroll", handleScroll);
        };
    }, [searchQuery, cursor, loading]);

    const ImageCards = images?.length ? images.map((image, index) => (
        <ImageCard key={index} filename={image.filename} tags={image.tags} />
    )) : null;

    return (
        <div className=''>
            <div
                onMouseEnter={() => setIsHovered(true)}
                onMouseLeave={() => setIsHovered(false)}
            >
                <SearchOverlay isHovered={isHovered} onQueryChange={handleSearch} />
                <div className="flex flex-wrap justify-evenly gap-5 p-4">
                    {ImageCards}
                </div>
            </div>
        </div>
    )
}

export default Gallery

