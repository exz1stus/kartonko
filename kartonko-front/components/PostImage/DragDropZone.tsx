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
                    className="fixed h-full w-full border-2 border-dashed z-50"
                >
                </div>
            )}
        </>
    );

};

export default DragDropZone;
