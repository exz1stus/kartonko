"use client";

import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ArrowLeft, Images, Pencil, Plus, Trash2 } from "lucide-react";
import {
    BoardItemResponse,
    BoardResponse,
    ImageMetadata,
} from "@/lib/api/generated/model";
import {
    deleteBoard,
    deleteBoardImage,
    listBoardImages,
    patchBoard,
} from "@/lib/api/generated/client";
import { useAuth } from "@/contexts/AuthContext";
import { useDialog } from "@/contexts/AlertDialogContext";
import useInfiniteScroll from "@/hooks/useInfiniteScroll";
import InfiniteImageGrid from "@/components/Gallery/InfiniteImageGrid";
import { ImageSearch, SearchQuery } from "@/components/Gallery/ImageSearch";
import Scrollbar from "@/components/template/Scrollbar";
import BoardImagePicker from "./BoardImagePicker";
import { toast } from "sonner";
import { addImagesToBoard } from "@/lib/image/addImagesToBoard";
import GallerySelectionControls from "@/components/Gallery/GallerySelectionControls";
import { imagesDeletedEvent } from "@/lib/image/deletedImages";
import { useTranslations } from "@/contexts/LocaleContext";
import {
    ResizableHandle,
    ResizablePanel,
    ResizablePanelGroup,
} from "@/components/ui/resizable";

const emptySearch: SearchQuery = { prefix: "", tags: [] };

export default function BoardDetail({
    initialBoard,
    initialItems,
    initialImageIds,
    initialPickerImages,
}: {
    initialBoard: BoardResponse;
    initialItems: BoardItemResponse[];
    initialImageIds: number[];
    initialPickerImages: ImageMetadata[];
}) {
    const router = useRouter();
    const { user } = useAuth();
    const { t } = useTranslations();
    const confirm = useDialog();
    const [board, setBoard] = useState(initialBoard);
    const [editing, setEditing] = useState(false);
    const [ctrlHeld, setCtrlHeld] = useState(false);
    const [adding, setAdding] = useState(false);
    const [busy, setBusy] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [addedItems, setAddedItems] = useState<BoardItemResponse[]>([]);
    const [search, setSearch] = useState(JSON.stringify(emptySearch));
    const [membershipIds, setMembershipIds] = useState<Set<number>>(
        new Set(initialImageIds),
    );
    const membershipList = useMemo(() => [...membershipIds], [membershipIds]);
    const [wide, setWide] = useState(false);
    useEffect(() => {
        const media = window.matchMedia("(min-width: 1024px)");
        const update = () => setWide(media.matches);
        update();
        media.addEventListener("change", update);
        return () => media.removeEventListener("change", update);
    }, []);
    const canEdit =
        !!user && (user.id === board.user_id || user.privilege === "Moderator");
    const showDeleteActions = canEdit && (editing || ctrlHeld);
    useEffect(() => {
        const onKeyDown = (event: KeyboardEvent) => {
            if (event.ctrlKey || event.key === "Control") setCtrlHeld(true);
        };
        const onKeyUp = (event: KeyboardEvent) => setCtrlHeld(event.ctrlKey);
        const reset = () => setCtrlHeld(false);
        const onVisibilityChange = () => {
            if (document.hidden) reset();
        };
        window.addEventListener("keydown", onKeyDown);
        window.addEventListener("keyup", onKeyUp);
        window.addEventListener("blur", reset);
        document.addEventListener("visibilitychange", onVisibilityChange);
        return () => {
            window.removeEventListener("keydown", onKeyDown);
            window.removeEventListener("keyup", onKeyUp);
            window.removeEventListener("blur", reset);
            document.removeEventListener("visibilitychange", onVisibilityChange);
        };
    }, []);
    const fetchPage = useCallback(
        (query: string, cursor: number, limit: number) =>
            listBoardImages(initialBoard.id, {
                ...(JSON.parse(query) as SearchQuery),
                cursor,
                limit,
            }),
        [initialBoard.id],
    );
    const images = useInfiniteScroll({
        query: search,
        fetchFn: fetchPage,
        requestSize: 24,
        initRequestSize: 24,
        initialItems,
        initialReachedEnd: initialItems.length < 24,
    });
    const boardItems = images.items;
    const removeImageItems = images.removeItems;
    useEffect(() => {
        const onImagesDeleted = (event: Event) => {
            const deleted = new Set((event as CustomEvent<number[]>).detail);
            if (!deleted.size) return;
            removeImageItems((item) => deleted.has(item.image_metadata.id));
            const removedFromBoard = [...deleted].filter((id) =>
                membershipIds.has(id) ||
                boardItems.some((item) => item.image_metadata.id === id) ||
                addedItems.some((item) => item.image_metadata.id === id),
            ).length;
            if (removedFromBoard)
                setBoard((current) => ({
                    ...current,
                    image_count: Math.max(0, current.image_count - removedFromBoard),
                }));
            setAddedItems((current) => current.filter((item) => !deleted.has(item.image_metadata.id)));
            setMembershipIds((current) => new Set([...current].filter((id) => !deleted.has(id))));
        };
        window.addEventListener(imagesDeletedEvent, onImagesDeleted);
        return () => window.removeEventListener(imagesDeletedEvent, onImagesDeleted);
    }, [addedItems, boardItems, membershipIds, removeImageItems]);
    const displayedItems = useMemo(() => {
        const seen = new Set<number>();
        const query = JSON.parse(search) as SearchQuery;
        const matchingAdded = addedItems.filter(
            (item) =>
                item.image_metadata.filename
                    .toLowerCase()
                    .startsWith((query.prefix ?? "").toLowerCase()) &&
                (query.tags ?? []).every((tag) =>
                    item.image_metadata.tags.includes(tag),
                ),
        );
        return [...images.items, ...matchingAdded].filter((item) => {
            if (seen.has(item.image_metadata.id)) return false;
            seen.add(item.image_metadata.id);
            return true;
        });
    }, [addedItems, images.items, search]);

    async function save(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();
        const data = new FormData(event.currentTarget);
        const name = String(data.get("name") ?? "").trim();
        if (!name) return;
        setBusy(true);
        setError(null);
        try {
            const updated = await patchBoard(board.id, {
                name,
                description: String(data.get("description") ?? "").trim(),
            });
            setBoard(updated);
            setEditing(false);
            if (updated.slug !== board.slug) {
                router.replace(`/board/${encodeURIComponent(updated.slug)}`);
            } else {
                router.refresh();
            }
        } catch (cause) {
            setError(
                cause instanceof Error ? cause.message : t("error.couldNotSaveBoard"),
            );
        } finally {
            setBusy(false);
        }
    }

    async function removeBoard() {
        if (
            !(await confirm(
                t("boards.deleteConfirm", { name: board.name }),
                { title: t("common.delete"), confirmText: t("common.delete") },
            ))
        )
            return;
        setBusy(true);
        setError(null);
        try {
            await deleteBoard(board.id);
            router.push("/boards");
            router.refresh();
        } catch (cause) {
            setError(
                cause instanceof Error
                    ? cause.message
                    : t("error.couldNotDeleteBoard"),
            );
            setBusy(false);
        }
    }

    async function removeImage(imageId: number) {
        if (
            !(await confirm(
                t("boards.removeImageConfirm"),
                { title: t("common.remove"), confirmText: t("common.remove") },
            ))
        )
            return;
        setBusy(true);
        setError(null);
        try {
            await deleteBoardImage(board.id, imageId);
            setBoard((current) => ({
                ...current,
                image_count: Math.max(0, current.image_count - 1),
            }));
            setMembershipIds((current) => {
                const next = new Set(current);
                next.delete(imageId);
                return next;
            });
            setAddedItems((current) => current.filter((item) => item.image_metadata.id !== imageId));
            images.removeItems((item) => item.image_metadata.id === imageId);
        } catch (cause) {
            setError(
                cause instanceof Error
                    ? cause.message
                    : t("error.couldNotRemoveImage"),
            );
        } finally {
            setBusy(false);
        }
    }

    const recordAdded = (item: BoardItemResponse) => {
        setBoard((current) => ({ ...current, image_count: current.image_count + 1 }));
        setAddedItems((current) => [item, ...current]);
        setMembershipIds((current) => new Set(current).add(item.image_metadata.id));
    };

    const addDraggedImages = async (imageIds: number[]) => {
        const missing = [...new Set(imageIds)].filter((id) => !membershipIds.has(id));
        const { added, failed } = await addImagesToBoard(board.id, missing);
        added.forEach(recordAdded);
        if (added.length > 0) toast.success(t("boards.added", { count: added.length, name: board.name }));
        if (failed > 0) toast.error(t("boards.addFailed", { count: failed, name: board.name }));
        if (missing.length === 0) toast.message(t("boards.allAlready"));
    };

    const handlePickerImageDrop = async (imageIds: number[], sourceBoardId?: number) => {
        if (sourceBoardId !== board.id) {
            await addDraggedImages(imageIds);
            return;
        }
        const existing = [...new Set(imageIds)].filter((id) => membershipIds.has(id));
        if (existing.length === 0) return;
        const results = await Promise.allSettled(existing.map((id) => deleteBoardImage(board.id, id)));
        const removed = existing.filter((_, index) => results[index].status === "fulfilled");
        const failed = existing.length - removed.length;
        if (removed.length) {
            setBoard((current) => ({ ...current, image_count: Math.max(0, current.image_count - removed.length) }));
            setMembershipIds((current) => {
                const next = new Set(current);
                removed.forEach((id) => next.delete(id));
                return next;
            });
            setAddedItems((current) => current.filter((item) => !removed.includes(item.image_metadata.id)));
            const removedIds = new Set(removed);
            images.removeItems((item) => removedIds.has(item.image_metadata.id));
            toast.success(t("boards.removed", { count: removed.length, name: board.name }));
        }
        if (failed) toast.error(t("boards.removeFailed", { count: failed }));
    };

    const handleBoardGalleryImageDrop = (imageIds: number[], sourceBoardId?: number) => {
        if (sourceBoardId === board.id) return;
        void addDraggedImages(imageIds);
    };

    const picker =
        canEdit && adding ? (
            <BoardImagePicker
                boardId={board.id}
                initialImageIds={membershipList}
                initialImages={initialPickerImages}
                onAdded={recordAdded}
                onImagesDropped={handlePickerImageDrop}
                boardName={board.name}
            />
        ) : null;

    const header = (
        <>
            <title>{board.name} | kartonko</title>
            <Link
                href="/boards"
                className="inline-flex items-center gap-2 text-muted-foreground text-sm hover:underline"
            >
                <ArrowLeft size={16} /> All boards
            </Link>
            <header className="flex flex-wrap justify-between items-start gap-4">
                <div className="loaded-content min-w-0">
                    <h1 className="font-semibold text-3xl break-words">
                        {board.name}
                    </h1>
                    {board.description && (
                        <p className="mt-2 max-w-2xl text-muted-foreground whitespace-pre-wrap">
                            {board.description}
                        </p>
                    )}
                    <p className="flex items-center gap-2 mt-3 text-muted-foreground text-sm">
                        <Images size={16} /> {board.image_count}{" "}
                        {board.image_count === 1 ? "image" : "images"}
                    </p>
                </div>
                {canEdit && (
                    <div className="flex gap-2">
                        <button
                            onClick={() => setEditing(!editing)}
                            aria-pressed={editing}
                            className="inline-flex items-center gap-2 hover:bg-surface-20 px-3 py-2 border border-surface-30 rounded-lg"
                        >
                            <Pencil size={16} /> Edit
                        </button>
                        {showDeleteActions && (
                            <button
                                disabled={busy}
                                onClick={removeBoard}
                                className="inline-flex items-center gap-2 hover:bg-red-500/10 disabled:opacity-50 px-3 py-2 border border-red-500/50 rounded-lg text-red-500"
                            >
                                <Trash2 size={16} /> Delete
                            </button>
                        )}
                    </div>
                )}
            </header>
            {editing && canEdit && (
                <form
                    onSubmit={save}
                    className="glass gap-3 grid p-4 border border-surface-20 rounded-2xl"
                >
                    <label className="gap-1 grid">
                        Name
                        <input
                            name="name"
                            required
                            maxLength={120}
                            defaultValue={board.name}
                            className="bg-surface-10 px-3 py-2 border border-surface-30 rounded-lg"
                        />
                    </label>
                    <label className="gap-1 grid">
                        Description
                        <textarea
                            name="description"
                            maxLength={500}
                            rows={3}
                            defaultValue={board.description}
                            className="bg-surface-10 px-3 py-2 border border-surface-30 rounded-lg"
                        />
                    </label>
                    <div className="flex gap-2">
                        <button
                            disabled={busy}
                            className="bg-primary-0 disabled:opacity-50 px-4 py-2 rounded-lg text-primary-foreground"
                        >
                            Save changes
                        </button>
                        <button
                            type="button"
                            onClick={() => setEditing(false)}
                            className="px-4 py-2 border border-surface-30 rounded-lg"
                        >
                            Cancel
                        </button>
                    </div>
                </form>
            )}
            {canEdit && (
                <div className="space-y-3">
                    <button
                        onClick={() => setAdding(!adding)}
                        aria-expanded={adding}
                        className="inline-flex items-center gap-2 bg-primary-0 hover:opacity-90 px-4 py-2 rounded-lg text-primary-foreground"
                    >
                        <Plus size={16} /> Add images
                    </button>
                </div>
            )}
            {error && (
                <p role="alert" className="text-red-500">
                    {error}
                </p>
            )}
            <div className="flex items-center">
                <ImageSearch
                    initialQuery={emptySearch}
                    onQueryChange={(next) => setSearch(JSON.stringify(next))}
                    className="min-w-0 flex-1"
                />
                <GallerySelectionControls scope={`board:${board.id}:gallery`} images={displayedItems.map((item) => item.image_metadata)} />
            </div>
        </>
    );

    const gallery = (
        <InfiniteImageGrid
            edgeToEdge
            onImagesDropped={canEdit ? handleBoardGalleryImageDrop : undefined}
            dropIntent={{ tags: (JSON.parse(search) as SearchQuery).tags, actions: canEdit ? [{ kind: "add_to_board", boardId: board.id, boardName: board.name }] : [] }}
            images={displayedItems.map((item) => item.image_metadata)}
            dragSourceBoardId={board.id}
            selectionScope={`board:${board.id}:gallery`}
            loading={images.loading}
            reachedEnd={images.reachedEnd}
            error={images.error}
            retry={images.retry}
            sentinelRef={images.sentinelRef}
            emptyMessage={
                search === JSON.stringify(emptySearch)
                    ? t("boards.emptyDetail")
                    : t("gallery.noResults")
            }
            renderAction={
                showDeleteActions
                    ? (image) => (
                          <button
                              disabled={busy}
                              onClick={() => removeImage(image.id)}
                              aria-label={`Remove ${image.filename} from board`}
                              className="bg-surface-0/90 hover:bg-surface-0 disabled:opacity-50 p-2 rounded-full text-red-500"
                          >
                              <Trash2 size={16} />
                          </button>
                      )
                    : undefined
            }
        />
    );

    const boardHeader = (
        <div className="border-surface-20 border-b shrink-0">
            <div className="space-y-4 mx-auto px-4 sm:px-8 py-4 max-w-6xl">
                {header}
            </div>
        </div>
    );

    const boardPane = wide ? (
        <div className="flex flex-col h-full min-h-0 overflow-hidden">
            {boardHeader}
            <Scrollbar className="flex-1 min-h-0 overflow-x-hidden overscroll-contain">
                {gallery}
            </Scrollbar>
        </div>
    ) : (
        <Scrollbar className="overflow-x-hidden">
            {boardHeader}
            {picker && <div className="p-3 h-[65vh] min-h-[300px]">{picker}</div>}
            {gallery}
        </Scrollbar>
    );

    return (
        <div className="h-full min-h-0 overflow-hidden">
            {picker && wide ? (
                <ResizablePanelGroup direction="horizontal">
                    <ResizablePanel
                        defaultSize={62}
                        minSize={35}
                        className="min-w-0 overflow-hidden"
                    >
                        {boardPane}
                    </ResizablePanel>
                    <ResizableHandle
                        withHandle
                        className="border-surface-20 border-x"
                    />
                    <ResizablePanel
                        defaultSize={38}
                        minSize={25}
                        className="p-3 min-w-0 overflow-hidden"
                    >
                        {picker}
                    </ResizablePanel>
                </ResizablePanelGroup>
            ) : (
                boardPane
            )}
        </div>
    );
}
