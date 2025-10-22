"use client";
import { useState, DragEvent } from "react";

interface DragDropZoneProps {
    children: React.ReactNode;
    onFilesDropped: (files: FileList) => void;
}

const DragDropZone: React.FC<DragDropZoneProps> = ({ children, onFilesDropped }) => {
    const [dragOver, setDragOver] = useState<boolean>(false);

    const handleDragOver = (e: DragEvent<HTMLDivElement>): void => {
        e.preventDefault();
        setDragOver(true);
    };

    const handleDrop = (e: DragEvent<HTMLDivElement>): void => {
        e.preventDefault();
        setDragOver(false);

        if (e.dataTransfer) {
            onFilesDropped(e.dataTransfer.files);
        }
    };

    const handleDragEnd = () => {
        setDragOver(false);
    };

    const handleMouseLeave = () => {
        setDragOver(false);
    };

    return (
        <div
            onDrop={handleDrop}
            onDragOver={handleDragOver}
            onDragEnd={handleDragEnd}
            onMouseLeave={handleMouseLeave}
            className={`h-full w-full z-100 border-2 ${dragOver ? "border-dashed opacity-80" : "border-transparent"}`}
        >
            {children}
        </div >
    );
};

export default DragDropZone;
