"use client";
import { useState, useEffect, DragEvent } from "react";

interface DragDropZoneProps {
    onFilesDropped: (files: FileList) => void;
}

const DragDropZone: React.FC<DragDropZoneProps> = ({ onFilesDropped }) => {
    const [dragOver, setDragOver] = useState<boolean>(false);

    useEffect(() => {
        const handleDragOver = (e: Event): void => {
            e.preventDefault();
            setDragOver(true);
        };

        const handleDrop = (e: Event): void => {
            e.preventDefault();
            setDragOver(false);

            const dragEvent = e as unknown as DragEvent;
            if (dragEvent.dataTransfer) {
                onFilesDropped(dragEvent.dataTransfer.files);
            }
        };

        const handleDragLeave = (e: Event): void => {
            e.preventDefault();
            // setDragOver(false);
        };

        window.addEventListener("dragover", handleDragOver);
        window.addEventListener("drop", handleDrop);
        window.addEventListener("dragleave", handleDragLeave);

        return () => {
            window.removeEventListener("dragover", handleDragOver);
            window.removeEventListener("drop", handleDrop);
            window.removeEventListener("dragleave", handleDragLeave);
        };

    }, [onFilesDropped]);

    return (
        <>
            {dragOver && (
                <div
                    className="text-center transition-colors fixed inset-0 border-2 border-dashed animated-gradient-border z-50"
                ></div>
            )}
        </>
    );

};

export default DragDropZone;
