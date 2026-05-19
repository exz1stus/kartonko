"use client";

import { useState, DragEvent } from "react";

interface DragDropZoneProps {
    children: React.ReactNode;
    onFilesDropped: (files: File[]) => void;
}

const DragDropZone = ({ children, onFilesDropped }: DragDropZoneProps) => {
    const [dragOver, setDragOver] = useState(false);

    const handleDragOver = (e: DragEvent<HTMLDivElement>) => {
        e.preventDefault();
        e.stopPropagation();
        setDragOver(true);
    };

    const handleDragLeave = (e: DragEvent<HTMLDivElement>) => {
        e.preventDefault();
        e.stopPropagation();
        setDragOver(false);
    };

    const handleDrop = (e: DragEvent<HTMLDivElement>) => {
        e.preventDefault();
        e.stopPropagation();
        setDragOver(false);

        const files = Array.from(e.dataTransfer.files);

        if (files.length > 0) {
            onFilesDropped(files);
        }
    };

    return (
        <div
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            onDrop={handleDrop}
            className={`relative h-full w-full border-2 transition 
                ${dragOver ? "border-dashed border-primary-10 bg-primary-10/20" : "border-transparent"}
            `}
        >
            {children}

            {dragOver && (
                <div className="absolute inset-0 flex justify-center items-center pointer-events-none">
                    <span className="font-semibold text-lg">Drop images</span>
                </div>
            )}
        </div>
    );
};

export default DragDropZone;
