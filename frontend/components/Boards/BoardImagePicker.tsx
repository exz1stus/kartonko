"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { Check, EyeOff, Plus } from "lucide-react";
import { getImagesByQuery, listBoardImageIDs, postBoardImage } from "@/lib/api/generated/client";
import { BoardItemResponse, ImageMetadata } from "@/lib/api/generated/model";
import { ImageSearch, SearchQuery } from "@/components/Gallery/ImageSearch";
import InfiniteImageGrid from "@/components/Gallery/InfiniteImageGrid";
import useInfiniteScroll from "@/hooks/useInfiniteScroll";
import Scrollbar from "@/components/template/Scrollbar";
import GallerySelectionControls from "@/components/Gallery/GallerySelectionControls";
import { imagesDeletedEvent } from "@/lib/image/deletedImages";

const emptySearch: SearchQuery = { prefix: "", tags: [] };

export default function BoardImagePicker({ boardId, boardName, initialImageIds, initialImages, onAdded, onImagesDropped }: {
    boardId: number;
    boardName: string;
    initialImageIds?: number[];
    initialImages?: ImageMetadata[];
    onAdded: (item: BoardItemResponse) => void;
    onImagesDropped: (imageIds: number[], sourceBoardId?: number, sourceScope?: string) => void;
}) {
    const [query, setQuery] = useState(JSON.stringify(emptySearch));
    const [hideAdded, setHideAdded] = useState(false);
    const [pendingIds, setPendingIds] = useState<Set<number>>(new Set());
    const pendingRef = useRef(new Set<number>());
    const [error, setError] = useState<string | null>(null);
    const [membershipError, setMembershipError] = useState<string | null>(null);
    const [membershipLoaded, setMembershipLoaded] = useState(initialImageIds !== undefined);
    const [addedIds, setAddedIds] = useState<Set<number>>(new Set(initialImageIds ?? []));
    const fetchPage = useCallback((search: string, cursor: number, limit: number) =>
        getImagesByQuery({ ...JSON.parse(search) as SearchQuery, semantic: true, cursor, limit }), []);
    const results = useInfiniteScroll({
        query,
        fetchFn: fetchPage,
        requestSize: 30,
        initRequestSize: 30,
        initialItems: initialImages,
        initialReachedEnd: initialImages !== undefined && initialImages.length < 30,
    });
    const removeImageItems = results.removeItems;
    useEffect(() => {
        const onImagesDeleted = (event: Event) => {
            const deleted = new Set((event as CustomEvent<number[]>).detail);
            if (!deleted.size) return;
            removeImageItems((image) => deleted.has(image.id));
            setAddedIds((current) => new Set([...current].filter((id) => !deleted.has(id))));
        };
        window.addEventListener(imagesDeletedEvent, onImagesDeleted);
        return () => window.removeEventListener(imagesDeletedEvent, onImagesDeleted);
    }, [removeImageItems]);
    const visibleItems = hideAdded ? results.items.filter((image) => !addedIds.has(image.id)) : results.items;
    const selectionScope = `board:${boardId}:picker`;
    const handleImagesDropped = (imageIds: number[], sourceBoardId?: number, sourceScope?: string) => {
        if (sourceScope === selectionScope) return;
        onImagesDropped(imageIds, sourceBoardId, sourceScope);
    };

    const loadMembership = useCallback(async () => {
        setMembershipLoaded(false);
        setMembershipError(null);
        try {
            const ids = await listBoardImageIDs(boardId);
            setAddedIds((current) => new Set([...ids, ...current]));
            setMembershipLoaded(true);
        } catch (cause) {
            setMembershipError(cause instanceof Error ? cause.message : "Could not check board images");
        }
    }, [boardId]);

    useEffect(() => { if (initialImageIds === undefined) void loadMembership(); }, [initialImageIds, loadMembership]);
    useEffect(() => { if (initialImageIds !== undefined) setAddedIds(new Set(initialImageIds)); }, [initialImageIds]);

    async function add(image: ImageMetadata) {
        if (!membershipLoaded || addedIds.has(image.id) || pendingRef.current.has(image.id)) return;
        pendingRef.current.add(image.id);
        setPendingIds(new Set(pendingRef.current));
        setError(null);
        try {
            const item = await postBoardImage(boardId, { image_id: image.id });
            setAddedIds((current) => new Set(current).add(image.id));
            onAdded(item);
        } catch (cause) {
            setError(cause instanceof Error ? cause.message : "Could not add image");
            if (cause instanceof Error && cause.message.includes("already on this board")) {
                setAddedIds((current) => new Set(current).add(image.id));
            }
        } finally {
            pendingRef.current.delete(image.id);
            setPendingIds(new Set(pendingRef.current));
        }
    }

    return <div className="flex h-full min-h-0 flex-col rounded-2xl border border-surface-20 bg-surface-0/70 p-3 sm:p-4">
        <div className="shrink-0 border-b border-surface-20 pb-3">
            <div className="mb-3 flex items-center gap-2">
                <h2 className="flex-1 text-lg font-medium">Find images for this board</h2>
                <button type="button" aria-pressed={hideAdded} onClick={() => setHideAdded((value) => !value)} className={`inline-flex items-center gap-2 rounded-lg border px-3 py-2 text-sm ${hideAdded ? "border-primary-0 bg-primary-0 text-surface-0" : "border-surface-30 hover:bg-surface-20"}`}>
                    <EyeOff size={16} /> Hide added
                </button>
            </div>
            <div className="flex items-center">
                <ImageSearch initialQuery={emptySearch} onQueryChange={(next) => setQuery(JSON.stringify(next))} className="min-w-0 flex-1" />
                <GallerySelectionControls scope={selectionScope} images={visibleItems} />
            </div>
            {error && <p role="alert" className="text-red-500">{error}</p>}
            {membershipError && <p role="alert" className="text-red-500">{membershipError} <button onClick={() => void loadMembership()} className="underline">Retry</button></p>}
        </div>
        <div className="min-h-0 flex-1">
          <Scrollbar className="min-h-0 overflow-x-hidden lg:overscroll-contain">
            <InfiniteImageGrid selectionScope={selectionScope} images={visibleItems} loading={results.loading} reachedEnd={results.reachedEnd} error={results.error} retry={results.retry} sentinelRef={results.sentinelRef} dropIntent={{ tags: (JSON.parse(query) as SearchQuery).tags, actions: [{ kind: "add_to_board", boardId, boardName }] }} onImagesDropped={handleImagesDropped} onImageClick={(image) => void add(image)} actionClickThrough renderAction={(image) => {
                const alreadyAdded = addedIds.has(image.id);
                const pending = pendingIds.has(image.id);
                return <span className={`inline-flex items-center gap-1 rounded-full bg-surface-0/90 px-2 py-1 text-sm ${alreadyAdded ? "text-green-600" : ""}`} aria-label={alreadyAdded ? `${image.filename} is on this board` : undefined}>
                    {alreadyAdded ? <><Check size={16} /> Added</> : pending ? "Adding…" : <><Plus size={16} /> Add</>}
                </span>;
            }} />
            {!membershipLoaded && !membershipError && <p role="status" className="p-3 text-center">Checking board images…</p>}
          </Scrollbar>
        </div>
    </div>;
}
