"use client";

import { useEffect } from "react";
import { ListChecks } from "lucide-react";
import { setActiveGalleryScope, useActiveGalleryScope, useGallerySelection } from "@/hooks/useGallerySelection";

export default function GallerySelectionControls({ scope, images }: { scope: string; images: { id: number; user_id: number }[] }) {
    const imageIds = images.map((image) => image.id);
    const selectionMode = useGallerySelection(scope, (state) => state.selectionMode);
    const selectedCount = useGallerySelection(scope, (state) => state.selectedIds.length);
    const setSelectionMode = useGallerySelection(scope, (state) => state.setSelectionMode);
    const endSelection = useGallerySelection(scope, (state) => state.endSelection);
    const selectMany = useGallerySelection(scope, (state) => state.selectMany);
    const activeScope = useActiveGalleryScope((state) => state.scope);
    const resetSelection = useGallerySelection(scope, (state) => state.reset);
    const registerImages = useGallerySelection(scope, (state) => state.registerImages);

    useEffect(() => registerImages(images), [images, registerImages]);

    useEffect(() => {
        const onKeyDown = (event: KeyboardEvent) => {
            if (!selectionMode || activeScope !== scope || event.target instanceof HTMLInputElement || event.target instanceof HTMLTextAreaElement || (event.target instanceof HTMLElement && event.target.isContentEditable)) return;
            if (event.key === "Escape") setSelectionMode(false);
            else if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === "a") { event.preventDefault(); selectMany(imageIds); }
            else if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === "d") { event.preventDefault(); selectMany([]); }
        };
        window.addEventListener("pointerup", endSelection);
        window.addEventListener("keydown", onKeyDown);
        return () => { window.removeEventListener("pointerup", endSelection); window.removeEventListener("keydown", onKeyDown); };
    }, [activeScope, endSelection, imageIds, scope, selectMany, selectionMode, setSelectionMode]);
    useEffect(() => () => resetSelection(), [resetSelection]);

    return <button type="button" aria-pressed={selectionMode} onClick={() => { setActiveGalleryScope(scope); setSelectionMode(!selectionMode); }} className={`mr-4 inline-flex shrink-0 items-center gap-2 rounded-xl border px-3 py-2 text-sm transition ${selectionMode ? "border-primary-0 bg-primary-0 text-surface-0" : "border-surface-30 hover:bg-surface-20"}`}>
        <ListChecks size={18} />{selectionMode ? `Done${selectedCount ? ` · ${selectedCount}` : ""}` : "Select images"}
    </button>;
}
