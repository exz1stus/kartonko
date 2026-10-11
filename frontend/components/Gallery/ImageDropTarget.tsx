"use client";

import { DragEvent, ReactNode, useState } from "react";
import { GALLERY_IMAGES_MIME, readDraggedImageIds } from "@/lib/image/galleryDrag";

export default function ImageDropTarget({ children, onDropImages, className = "" }: {
    children: ReactNode;
    onDropImages?: (ids: number[]) => void;
    className?: string;
}) {
    const [over, setOver] = useState(false);
    const accepts = (event: DragEvent) => onDropImages !== undefined && event.dataTransfer.types.includes(GALLERY_IMAGES_MIME);
    return <div className={`${className} ${over ? "ring-2 ring-primary-0" : ""}`}
        onDragOver={(event) => {
            if (!accepts(event)) return;
            event.preventDefault();
            event.stopPropagation();
            event.dataTransfer.dropEffect = "copy";
            setOver(true);
        }}
        onDragLeave={(event) => {
            if (event.currentTarget.contains(event.relatedTarget as Node)) return;
            setOver(false);
        }}
        onDrop={(event) => {
            if (!accepts(event)) return;
            event.preventDefault();
            event.stopPropagation();
            setOver(false);
            const ids = readDraggedImageIds(event.dataTransfer);
            if (ids.length > 0) onDropImages?.(ids);
        }}
    >{children}</div>;
}
