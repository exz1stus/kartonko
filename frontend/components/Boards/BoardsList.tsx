"use client";

import { FormEvent, useCallback, useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { FolderPlus, Images, Search } from "lucide-react";
import { useDebounce } from "use-debounce";
import {
    listBoards,
    postBoard,
} from "@/lib/api/generated/client";
import { BoardResponse } from "@/lib/api/generated/model";
import { useAuth } from "@/contexts/AuthContext";
import useInfiniteScroll from "@/hooks/useInfiniteScroll";
import Scrollbar from "@/components/template/Scrollbar";
import ImageDropTarget from "@/components/Gallery/ImageDropTarget";
import { addImagesToBoard } from "@/lib/image/addImagesToBoard";
import { toast } from "sonner";
import BoardPreview from "@/components/Boards/BoardPreview";
import { useTranslations } from "@/contexts/LocaleContext";

export default function BoardsList({
    initialBoards,
}: {
    initialBoards: BoardResponse[];
}) {
    const router = useRouter();
    const { user } = useAuth();
    const { t } = useTranslations();
    const [search, setSearch] = useState("");
    const [query] = useDebounce(search.trim(), 250);
    const [creating, setCreating] = useState(false);
    const [mine, setMine] = useState(false);
    const [saving, setSaving] = useState(false);
    const [createError, setCreateError] = useState<string | null>(null);
    const fetchPage = useCallback(
        (
            filter: { name: string; userId?: number },
            cursor: number,
            limit: number,
        ) =>
            listBoards({
                name: filter.name,
                user_id: filter.userId,
                cursor,
                limit,
            }),
        [],
    );
    const { items, loading, reachedEnd, error, retry, sentinelRef } =
        useInfiniteScroll({
            query: { name: query, userId: mine ? user?.id : undefined },
            fetchFn: fetchPage,
            requestSize: 12,
            initRequestSize: 12,
            initialItems: initialBoards,
            initialReachedEnd: initialBoards.length < 12,
        });

    async function createBoard(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();
        const form = new FormData(event.currentTarget);
        const name = String(form.get("name") ?? "").trim();
        if (!name) return;
        setSaving(true);
        setCreateError(null);
        try {
            const board = await postBoard({
                name,
                description: String(form.get("description") ?? "").trim(),
            });
            router.push(`/board/${encodeURIComponent(board.slug)}`);
        } catch (cause) {
            setCreateError(
                cause instanceof Error
                    ? cause.message
                    : t("error.couldNotCreateBoard"),
            );
        } finally {
            setSaving(false);
        }
    }

    async function addDraggedImages(board: BoardResponse, ids: number[]) {
        const { added, already, failed } = await addImagesToBoard(board.id, ids);
        if (failed) toast.error(t("boards.addFailed", { count: failed, name: board.name }));
        if (added.length) toast.success(t("boards.added", { count: added.length, name: board.name }));
        if (already) toast.message(t("boards.alreadyAdded", { count: already, name: board.name }));
    }

    return (
        <div className="flex flex-col h-full min-h-0 overflow-hidden">
            <title>Boards | kartonko</title>
            <div className="border-surface-20 border-b shrink-0">
                <div className="space-y-4 mx-auto px-4 sm:px-8 py-4 max-w-6xl">
                    <header className="flex flex-wrap justify-between items-start gap-4">
                        <div>
                            <h1 className="font-semibold text-3xl">{t("nav.boards")}</h1>
                            <p className="mt-1 text-muted-foreground text-sm">
                                {t("boards.description")}
                            </p>
                        </div>
                        {user && (
                            <button
                                type="button"
                                onClick={() => setCreating(!creating)}
                                className="inline-flex items-center gap-2 bg-primary-0 hover:opacity-90 px-4 py-2 rounded-xl font-medium text-primary-foreground"
                            >
                                <FolderPlus size={18} /> {t("boards.new")}
                            </button>
                        )}
                    </header>

                    {creating && (
                        <form
                            onSubmit={createBoard}
                            className="glass gap-3 grid sm:grid-cols-2 p-4 border border-surface-20 rounded-2xl"
                        >
                            <label className="gap-1 grid text-sm">
                                {t("boards.name")}
                                <input
                                    name="name"
                                    required
                                    maxLength={120}
                                    autoFocus
                                    className="bg-surface-10 px-3 py-2 border border-surface-30 focus:border-primary-0 rounded-lg outline-none"
                                />
                            </label>
                            <label className="gap-1 grid text-sm">
                                {t("boards.descriptionField")}
                                <input
                                    name="description"
                                    maxLength={500}
                                    className="bg-surface-10 px-3 py-2 border border-surface-30 focus:border-primary-0 rounded-lg outline-none"
                                />
                            </label>
                            {createError && (
                                <p
                                    role="alert"
                                    className="sm:col-span-2 text-red-500"
                                >
                                    {createError}
                                </p>
                            )}
                            <div className="flex gap-2 sm:col-span-2">
                                <button
                                    disabled={saving}
                                    className="bg-primary-0 disabled:opacity-50 px-4 py-2 rounded-lg text-primary-foreground"
                                >
                                    {saving ? t("boards.creating") : t("boards.create")}
                                </button>
                                <button
                                    type="button"
                                    onClick={() => setCreating(false)}
                                    className="px-4 py-2 border border-surface-30 rounded-lg"
                                >
                                    {t("common.cancel")}
                                </button>
                            </div>
                        </form>
                    )}

                    <label className="control-shell flex items-center gap-2 px-3 rounded-xl">
                        <Search size={18} aria-hidden="true" />
                        <input
                            type="search"
                            value={search}
                            onChange={(event) => setSearch(event.target.value)}
                            placeholder={t("search.boards")}
                            aria-label={t("search.boards")}
                            className="py-3 outline-none w-full"
                        />
                    </label>
                    {user && (
                        <label className="flex items-center gap-2 w-fit text-sm">
                            <input
                                type="checkbox"
                                checked={mine}
                                onChange={(event) =>
                                    setMine(event.target.checked)
                                }
                            />{" "}
                            {t("boards.my")}
                        </label>
                    )}
                </div>
            </div>

            <Scrollbar className="flex-1 min-h-0 overflow-x-hidden overscroll-contain">
                <div className="space-y-6 mx-auto px-4 sm:px-8 py-6 max-w-6xl">
                    <div className="gap-4 grid">
                    {items.map((board) => (
                        <ImageDropTarget key={board.id} onDropImages={user && (user.id === board.user_id || user.privilege === "Moderator") ? (ids) => void addDraggedImages(board, ids) : undefined} className="rounded-2xl">
                            <Link
                                href={`/board/${encodeURIComponent(board.slug)}`}
                                className={`content-card group grid gap-4 rounded-2xl border border-surface-20 p-3 transition hover:border-primary-0/60 focus-visible:outline-2 focus-visible:outline-primary-0 ${board.image_count === 0 ? "content-card-muted" : "sm:grid-cols-[minmax(0,1fr)_minmax(320px,55%)] sm:items-center"}`}
                            >
                                {board.image_count > 0 && (
                                    <div className="sm:order-2">
                                        <BoardPreview board={board} />
                                    </div>
                                )}
                                <div className="sm:order-1 p-1 min-w-0">
                                    <h2 className="font-medium text-xl group-hover:underline truncate">
                                        {board.name}
                                    </h2>
                                    {board.description && (
                                        <p
                                            className="mt-2 line-clamp-2 text-sm text-muted-foreground"
                                        >
                                            {board.description}
                                        </p>
                                    )}
                                    <p
                                        className="mt-3 flex items-center gap-2 text-sm text-muted-foreground"
                                    >
                                        <Images size={16} /> {board.image_count}{" "}
                                        {board.image_count === 1 ? t("common.image") : t("common.images")}
                                    </p>
                                </div>
                            </Link>
                        </ImageDropTarget>
                        ))}
                    </div>
                    {!loading && !error && items.length === 0 && (
                        <p className="p-8 border border-surface-20 rounded-2xl text-center">
                            {query
                                ? t("boards.noResults")
                                : t("boards.empty")}
                        </p>
                    )}
                    {error && (
                        <div
                            role="alert"
                            className="flex items-center gap-3 text-red-500"
                        >
                            {error}
                            <button onClick={retry} className="underline">
                                {t("common.retry")}
                            </button>
                        </div>
                    )}
                    {loading && (
                        <p
                            role="status"
                            className="text-muted-foreground text-center"
                        >
                            Loading boards…
                        </p>
                    )}
                    {!reachedEnd && !error && (
                        <div ref={sentinelRef} className="h-1" />
                    )}
                </div>
            </Scrollbar>
        </div>
    );
}
