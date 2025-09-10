"use client";
import React, { useEffect, useState } from "react";
import PostImageForm, { ApiResponse } from "./PostImageForm";
import FancySpan from "@/components/FancySpan";
import PostImageErrorSpan from "./PostImageErrorSpan";

interface PostImageModalProps {
    images?: File[];
    onClose: () => void;
    onUploaded: () => void;
}

const PostImageModal: React.FC<PostImageModalProps> = ({ images = [], onUploaded, onClose }) => {
    const [imageIndex, setImageIndex] = useState(0);
    const [uploadingError, setError] = useState<string | null>(null);

    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.key === "Escape") {
                onClose();
                setError(null);
            }
        };

        window.addEventListener("keydown", handleKeyDown);

        return () => {
            window.removeEventListener("keydown", handleKeyDown);
        };
    }, [onClose]);

    if (images?.length === 0) return null;

    const handleWheel = (e: React.WheelEvent<HTMLDivElement>) => {
        const newIndex = imageIndex + Math.sign(e.deltaY);
        if (newIndex >= 0 && newIndex < images.length) {
            setImageIndex(newIndex);
            setError(null);
        }
    }

    const handleError = (error: string) => {
        if (error.includes("already exists")) {
            setError("Image already exists");
            return;
        }
        else setError(error);
    }

    const handleServerResponse = (response: ApiResponse) => {
        if (response?.error) {
            handleError(response.error);
            return;
        }

        setError(null);
        handleUploaded();
    }

    const handleUploaded = async () => {
        onUploaded();
        if (images.length <= 1) {
            onClose();
            setError(null);
            return;
        }

        if (images.length - 1 === imageIndex) {
            setImageIndex(prev => prev - 1);
        }
    }

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center backdrop-blur-sm overflow-hidden">
            <div className="flex flex-row rounded-2xl bg-surface-0/95 h-[50vh]" onWheel={handleWheel}>
                <img
                    src={URL.createObjectURL(images[imageIndex])}
                    alt={images[imageIndex].name}
                    className="max-w-full object-contain rounded-l-2xl"
                />
                <div className="p-8 flex flex-col gap-10">
                    <div className="flex justify-between">
                        <FancySpan word={images[imageIndex].name} />
                        <div
                            className="flex justify-center items-center w-10 h-10 text-gray-700"
                        >
                            {imageIndex + 1}/{images.length}
                        </div>
                    </div>
                    <PostImageForm
                        image={images[imageIndex]}
                        onSubmit={handleServerResponse}
                    />
                </div>
            </div>
            <PostImageErrorSpan error={uploadingError} />
        </div>
    );
};

export default PostImageModal;

