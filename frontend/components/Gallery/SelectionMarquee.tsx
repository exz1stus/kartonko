"use client";

import { useActiveGalleryScope, useGallerySelection } from "@/hooks/useGallerySelection";

export default function SelectionMarquee() {
    const scope = useActiveGalleryScope((state) => state.scope);
    const bounds = useGallerySelection(scope, (state) => state.selectionBounds);
    if (!bounds) return null;
    return <div
        aria-hidden="true"
        className="gallery-selection-marquee pointer-events-none fixed z-[100] rounded-md border-2"
        style={{ left: bounds.left, top: bounds.top, width: Math.max(bounds.width, 2), height: Math.max(bounds.height, 2) }}
    />;
}
