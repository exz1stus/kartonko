"use client";
import Image from "next/image";
import PerspectiveCard from "./PerspectiveCard";
import React, { useRef } from "react";
import { useRouter } from "next/navigation";
import { ImageMetadata } from "@/lib/api/generated/model";
import { getGallerySelectionState, setActiveGalleryScope, useGallerySelection } from "@/hooks/useGallerySelection";
import { GALLERY_IMAGES_MIME, GALLERY_SOURCE_BOARD_MIME, GALLERY_SOURCE_SCOPE_MIME } from "@/lib/image/galleryDrag";
import { Check } from "lucide-react";
import { getPointerSelectionBounds, getPointerSelectionIds, getSpatialSelectionBounds, getSpatialSelectionIds, trackGalleryMarquee } from "@/lib/image/rangeSelection";
interface Props {
    image: ImageMetadata;
    selectionOrder: number[];
    className?: string;
    style?: React.CSSProperties;
    action?: React.ReactNode;
    onClick?: () => void;
    actionClickThrough?: boolean;
    dragSourceBoardId?: number;
    selectionScope?: string;
}

const ImageCard: React.FC<Props> = ({ image, selectionOrder, className, style, action, onClick, actionClickThrough = false, dragSourceBoardId, selectionScope = "home" }) => {
    const { filename, hash, width, height } = image;
    const selfRef = useRef<HTMLDivElement>(null);
    const dragPreviewRef = useRef<HTMLElement | null>(null);
    const router = useRouter();
    const selected = useGallerySelection(selectionScope, (state) => state.selectedIds.includes(image.id));
    const selectionMode = useGallerySelection(selectionScope, (state) => state.selectionMode);
    const openImage = () => {
        router.push(`/image/${image.filename}`);
    };
    const toggleSelection = useGallerySelection(selectionScope, (state) => state.toggle);
    const selectClick = useGallerySelection(selectionScope, (state) => state.selectClick);
    const beginSpatialSelection = useGallerySelection(selectionScope, (state) => state.beginSpatialSelection);
    const extendSpatialSelection = useGallerySelection(selectionScope, (state) => state.extendSpatialSelection);
    const endSelection = useGallerySelection(selectionScope, (state) => state.endSelection);
    const setDragging = useGallerySelection(selectionScope, (state) => state.setDragging);
    const activateImage = onClick ?? openImage;

    const selectSpatially = (event: { shiftKey: boolean; ctrlKey: boolean; metaKey: boolean }, finish = false) => {
        setActiveGalleryScope(selectionScope);
        const additive = event.ctrlKey || event.metaKey;
        if (additive && !event.shiftKey) {
            if (finish) {
                selectClick(selectionOrder, image.id, false, true);
                endSelection();
                return;
            }
        }
        const current = getGallerySelectionState(selectionScope);
        const anchor = event.shiftKey && current.anchorId !== null && selectionOrder.includes(current.anchorId) ? current.anchorId : image.id;
        const grid = selfRef.current?.closest("[data-gallery-grid]") ?? null;
        const bounds = getSpatialSelectionBounds(grid, anchor, image.id);
        const ids = getSpatialSelectionIds(grid, anchor, image.id);
        const operation = additive && !event.shiftKey && getGallerySelectionState(selectionScope).selectedIds.includes(image.id)
            ? "deselect"
            : "select";
        beginSpatialSelection(ids, anchor, image.id, additive, bounds, operation);
        if (finish) endSelection();
    };

    const startPointerSelection = (event: { shiftKey: boolean; ctrlKey: boolean; metaKey: boolean; clientX: number; clientY: number; preventDefault: () => void }) => {
        event.preventDefault();
        const additive = event.ctrlKey || event.metaKey;
        setActiveGalleryScope(selectionScope);
        const current = getGallerySelectionState(selectionScope);
        const anchor = event.shiftKey && current.anchorId !== null && selectionOrder.includes(current.anchorId) ? current.anchorId : image.id;
        const grid = selfRef.current?.closest("[data-gallery-grid]") ?? null;
        const operation = additive && !event.shiftKey && current.selectedIds.includes(image.id) ? "deselect" : "select";
        const pointerBounds = getPointerSelectionBounds(event.clientX, event.clientY, event.clientX, event.clientY);
        const initialIds = event.shiftKey
            ? getSpatialSelectionIds(grid, anchor, image.id)
            : getPointerSelectionIds(grid, pointerBounds);
        const initialBounds = event.shiftKey ? getSpatialSelectionBounds(grid, anchor, image.id) : pointerBounds;
        beginSpatialSelection(initialIds.length ? initialIds : [image.id], event.shiftKey ? anchor : image.id, image.id, additive, initialBounds, operation);

        if (grid instanceof HTMLElement) {
            trackGalleryMarquee(grid, event.clientX, event.clientY, (ids, bounds) => extendSpatialSelection(ids, bounds), endSelection);
        }
    };

    const onLoad = () => {
        selfRef.current?.classList.remove("opacity-0");
        selfRef.current?.clientHeight;
    };
    const cardContent = (
        <div className="flex flex-col items-center bg-surface-20 rounded-xl hover:cursor-pointer">
            <Image
                src={`/apilocal/image/hash/${encodeURIComponent(hash)}/thumb`}
                alt={filename}
                className="rounded-t-xl w-full h-auto"
                width={width}
                height={height}
                onLoad={onLoad}
                draggable={false}
            />
            <span className="px-2 max-w-[20ch] truncate leading-6">{filename}</span>
        </div>
    );

    return (
        <div
            ref={selfRef}
            data-gallery-image-id={image.id}
            data-gallery-selection-scope={selectionScope}
            className={`relative rounded-xl ${selected ? (selectionMode ? "z-10 ring-4 ring-primary-0 ring-offset-2 ring-offset-surface-0 shadow-xl" : "ring-2 ring-primary-0") : ""} ${className ?? ""}`}
            draggable={!selectionMode}
            onMouseDown={(event) => {
                if (!selectionMode || event.button !== 0) return;
                event.preventDefault();
                startPointerSelection(event);
            }}
            onDragStart={(event) => {
                const grid = selfRef.current?.closest("[data-gallery-grid]");
                const visibleIds = new Set(Array.from(grid?.querySelectorAll<HTMLElement>("[data-gallery-image-id]") ?? []).map((card) => Number(card.dataset.galleryImageId)));
                setActiveGalleryScope(selectionScope);
                const ids = selected ? getGallerySelectionState(selectionScope).selectedIds.filter((id) => visibleIds.has(id)) : [image.id];
                setDragging(ids);
                event.dataTransfer.setData(GALLERY_IMAGES_MIME, JSON.stringify(ids));
                event.dataTransfer.setData(GALLERY_SOURCE_SCOPE_MIME, selectionScope);
                if (dragSourceBoardId !== undefined) event.dataTransfer.setData(GALLERY_SOURCE_BOARD_MIME, String(dragSourceBoardId));
                event.dataTransfer.effectAllowed = "copy";
                if (ids.length > 1) {
                    const sources = ids.slice(0, 3).map((id) => {
                        const card = grid?.querySelector<HTMLElement>(`[data-gallery-image-id="${id}"]`);
                        return card?.querySelector("img")?.currentSrc || card?.querySelector("img")?.src;
                    }).filter((src): src is string => !!src);
                    if (sources.length) {
                        const preview = document.createElement("div");
                        preview.setAttribute("aria-hidden", "true");
                        Object.assign(preview.style, { position: "fixed", left: "-1000px", top: "0", width: "112px", height: "88px", pointerEvents: "none" });
                        sources.forEach((src, index) => {
                            const tile = document.createElement("div");
                            Object.assign(tile.style, { position: "absolute", left: `${index * 8}px`, top: `${index * 6}px`, width: "88px", height: "64px", overflow: "hidden", border: "2px solid white", borderRadius: "9px", background: "#e5e7eb", boxShadow: "0 3px 10px rgba(0,0,0,.28)" });
                            const thumb = document.createElement("img");
                            thumb.src = src;
                            Object.assign(thumb.style, { width: "100%", height: "100%", objectFit: "cover" });
                            tile.appendChild(thumb);
                            preview.appendChild(tile);
                        });
                        const count = document.createElement("span");
                        count.textContent = String(ids.length);
                        Object.assign(count.style, { position: "absolute", right: "0", bottom: "0", display: "grid", placeItems: "center", minWidth: "28px", height: "28px", padding: "0 6px", borderRadius: "999px", background: "#20242c", color: "white", font: "600 14px sans-serif", boxShadow: "0 2px 7px rgba(0,0,0,.35)" });
                        preview.appendChild(count);
                        document.body.appendChild(preview);
                        dragPreviewRef.current = preview;
                        event.dataTransfer.setDragImage(preview, 30, 30);
                    }
                }
            }}
            onDragEnd={() => {
                setDragging([]);
                dragPreviewRef.current?.remove();
                dragPreviewRef.current = null;
            }}
            style={style}
            onClick={(event) => {
                if (!selectionMode) activateImage();
            }}
            role="button"
            tabIndex={0}
            onKeyDown={(event) => {
                if (event.key === "Enter" || event.key === " ") {
                    event.preventDefault();
                    if (selectionMode) selectSpatially(event, true);
                    else activateImage();
                }
            }}
        >
            {selectionMode ? cardContent : <PerspectiveCard>{cardContent}</PerspectiveCard>}
            {selectionMode && <button
                type="button"
                aria-label={`${selected ? "Deselect" : "Select"} ${filename}`}
                aria-pressed={selected}
                className={`absolute left-2 top-2 z-20 flex size-8 items-center justify-center rounded-full border-2 shadow-md ${selected ? "border-primary-0 bg-primary-0 text-surface-0" : "border-surface-0 bg-surface-0/95 text-surface-600"}`}
                onPointerDown={(event) => {
                    event.preventDefault();
                    event.stopPropagation();
                    startPointerSelection(event);
                }}
                onClick={(event) => {
                    event.stopPropagation();
                    if (event.detail === 0) {
                        setActiveGalleryScope(selectionScope);
                        toggleSelection(image.id);
                    }
                }}
                onKeyDown={(event) => event.stopPropagation()}
            >
                {selected ? <Check size={16} /> : <span className="size-3 rounded-full border border-current" />}
            </button>}
            {action && <div className="absolute right-2 top-2 z-20" onClick={actionClickThrough ? undefined : (event) => event.stopPropagation()} onKeyDown={actionClickThrough ? undefined : (event) => event.stopPropagation()}>{action}</div>}
        </div>
    );
};

export default ImageCard;
