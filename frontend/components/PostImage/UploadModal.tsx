"use client";
import React, { useEffect, useState } from "react";
import PostImageForm from "./PostImageForm";
import FancySpan from "@/components/FancySpan";
import Image from "next/image";

interface PostImageModalProps {
    images?: File[];
    onClose: () => void;
    onUploaded: () => void;
}

const UploadModal: React.FC<PostImageModalProps> = ({ images = [], onUploaded, onClose }) => {
    const [imageIndex, setImageIndex] = useState(0);

    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.key === "Escape") {
                onClose();
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
        }
    }

    const handleUploaded = async () => {
        onUploaded();
        if (images.length <= 1) {
            onClose();
            return;
        }

        if (images.length - 1 === imageIndex) {
            setImageIndex(prev => prev - 1);
        }
    }

    return (
        <div className="z-50 fixed inset-0 flex justify-center items-center backdrop-blur-sm overflow-hidden">
            <div className="flex flex-row bg-surface-0/95 rounded-2xl h-[50vh]" onWheel={handleWheel}>
                <Image
                    src={URL.createObjectURL(images[imageIndex])}
                    alt={images[imageIndex].name}
                    className="rounded-l-2xl max-w-full object-contain"
                    width={256}
                    height={256}
                />
                <div className="flex flex-col gap-10 p-8">
                    <div className="flex justify-between">
                        <FancySpan word={images[imageIndex].name} />
                        <div
                            className="flex justify-center items-center w-10 h-10 text-gray-700"
                        >
                            {imageIndex + 1}/{images.length}
                        </div>
                    </div>
                    <PostImageForm
                        file={images[imageIndex]}
                        onSubmit={handleUploaded}
                    />
                </div>
            </div>
        </div>
    );
};

export default UploadModal;

