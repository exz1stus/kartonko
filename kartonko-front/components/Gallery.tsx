import React from 'react'
import ImageCard from './ImageCard'

const Gallery = () => {
    const requestSize = 30;
    const API_ORIGIN = "localhost:8080";

    const fetchImages = async (): Promise<void> => {
        try {
            const response = await fetch(`${API_ORIGIN}/images/${requestSize}`);
            if (response.ok) {
                const data: ImageData[] = await response.json();
            }
        } catch (error) {
            console.error("Error fetching images:", error);
        }
    }

    return (
        <div>
            <ImageCard />
        </div>
    )
}

export default Gallery
