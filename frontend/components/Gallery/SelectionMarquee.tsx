"use client";

import { useActiveGalleryScope, useGallerySelection } from "@/hooks/useGallerySelection";

export default function SelectionMarquee() {
    const scope = useActiveGalleryScope((state) => state.scope);
    const bounds = useGallerySelection(scope, (state) => state.selectionBounds);
    if (!bounds) return null;
    return <div
        aria-hidden="true"
        className="pointer-events-none fixed z-[100] rounded-md border-2 border-primary-0 bg-primary-0/15 shadow-[0_0_0_1px_rgba(255,255,255,0.8)]"
        style={{ left: bounds.left, top: bounds.top, width: Math.max(bounds.width, 2), height: Math.max(bounds.height, 2) }}
    />;
}
