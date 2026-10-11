"use client";

import { useState, DragEvent } from "react";
import { GALLERY_IMAGES_MIME, readDraggedImageIds, readDraggedSourceBoardId, readDraggedSourceScope } from "@/lib/image/galleryDrag";

interface DragDropZoneProps {
    children: React.ReactNode;
    onFilesDropped: (files: File[]) => void;
    onImagesDropped?: (imageIds: number[], sourceBoardId?: number, sourceScope?: string) => void;
    className?: string;
    label?: string;
}

const DragDropZone = ({ children, onFilesDropped, onImagesDropped, className = "h-full", label = "Drop images" }: DragDropZoneProps) => {
    const [dragOver, setDragOver] = useState(false);

    const accepts = (event: DragEvent<HTMLDivElement>) =>
        event.dataTransfer.types.includes(GALLERY_IMAGES_MIME)
            ? onImagesDropped !== undefined
            : event.dataTransfer.types.includes("Files");

    const handleDragOver = (e: DragEvent<HTMLDivElement>) => {
        if (!accepts(e)) return;
        e.preventDefault();
        e.stopPropagation();
        e.dataTransfer.dropEffect = "copy";
        setDragOver(true);
    };

    const handleDragLeave = (e: DragEvent<HTMLDivElement>) => {
        if (!dragOver) return;
        if (e.currentTarget.contains(e.relatedTarget as Node)) return;
        e.preventDefault();
        e.stopPropagation();
        setDragOver(false);
    };

    const handleDrop = (e: DragEvent<HTMLDivElement>) => {
        if (!accepts(e)) return;
        e.preventDefault();
        e.stopPropagation();
        setDragOver(false);

        const files = Array.from(e.dataTransfer.files);

        const imageIds = readDraggedImageIds(e.dataTransfer);
        if (imageIds.length > 0 && onImagesDropped) {
            onImagesDropped(imageIds, readDraggedSourceBoardId(e.dataTransfer), readDraggedSourceScope(e.dataTransfer));
        } else if (files.length > 0) {
            onFilesDropped(files);
        }
    };

    return (
        <div
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            onDrop={handleDrop}
            className={`relative w-full border-2 transition ${className}
                ${dragOver ? "border-dashed border-primary-10 bg-primary-10/20" : "border-transparent"}
            `}
        >
            {children}

            {dragOver && (
                <div className="absolute inset-0 flex justify-center items-center pointer-events-none">
                    <span className="font-semibold text-lg">{label}</span>
                </div>
            )}
        </div>
    );
};

export default DragDropZone;
