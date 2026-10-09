"use client";

import { FormEvent, useCallback, useEffect, useState } from "react";
import Image from "next/image";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { FolderPlus, Images, Search } from "lucide-react";
import { useDebounce } from "use-debounce";
import {
    listBoardImages,
    listBoards,
    postBoard,
} from "@/lib/api/generated/client";
import { BoardResponse, BoardItemResponse } from "@/lib/api/generated/model";
import { useAuth } from "@/contexts/AuthContext";
import useInfiniteScroll from "@/hooks/useInfiniteScroll";
import Scrollbar from "@/components/template/Scrollbar";

function BoardPreview({ board }: { board: BoardResponse }) {
    const [previews, setPreviews] = useState<BoardItemResponse[]>([]);
    const [previewError, setPreviewError] = useState(false);

    useEffect(() => {
        let active = true;
        if (board.image_count > 0)
            listBoardImages(board.id, { limit: 4 })
                .then((images) => {
                    if (active) setPreviews(images);
                })
                .catch(() => {
                    if (active) setPreviewError(true);
                });
        return () => {
            active = false;
        };
    }, [board.id, board.image_count]);

    return (
        <div
            className="gap-1 grid grid-cols-2 sm:grid-cols-4 bg-surface-20 rounded-xl sm:h-52 aspect-square sm:aspect-auto overflow-hidden"
            aria-label={previewError ? "Preview unavailable" : undefined}
        >
            {Array.from({ length: 4 }, (_, index) => {
                const image = previews[index]?.image_metadata;
                return (
                    <div
                        key={index}
                        className="relative flex justify-center items-center bg-surface-10 min-w-0 min-h-0"
                    >
                        {image ? (
                            <Image
                                src={`/apilocal/image/hash/${encodeURIComponent(image.hash)}/thumb`}
                                alt=""
                                fill
                                sizes="(max-width: 640px) 45vw, 160px"
                                className="object-cover"
                            />
                        ) : (
                            <Images
                                size={24}
                                aria-hidden="true"
                                className="opacity-20"
                            />
                        )}
                    </div>
                );
            })}
        </div>
    );
}

export default function BoardsList({
    initialBoards,
}: {
    initialBoards: BoardResponse[];
}) {
    const router = useRouter();
    const { user } = useAuth();
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
                    : "Could not create board",
            );
        } finally {
            setSaving(false);
        }
    }

    return (
        <div className="flex flex-col h-full min-h-0 overflow-hidden">
            <title>Boards | kartonko</title>
            <div className="border-surface-20 border-b shrink-0">
                <div className="space-y-4 mx-auto px-4 sm:px-8 py-4 max-w-6xl">
                    <header className="flex flex-wrap justify-between items-start gap-4">
                        <div>
                            <h1 className="font-semibold text-3xl">Boards</h1>
                            <p className="mt-1 text-primary-0/70 text-sm">
                                Collections of images, sorted by image count.
                            </p>
                        </div>
                        {user && (
                            <button
                                type="button"
                                onClick={() => setCreating(!creating)}
                                className="inline-flex items-center gap-2 bg-primary-0 hover:opacity-90 px-4 py-2 rounded-xl font-medium text-surface-0"
                            >
                                <FolderPlus size={18} /> New board
                            </button>
                        )}
                    </header>

                    {creating && (
                        <form
                            onSubmit={createBoard}
                            className="gap-3 grid sm:grid-cols-2 bg-surface-0/80 p-4 border border-surface-20 rounded-2xl"
                        >
                            <label className="gap-1 grid text-sm">
                                Name
                                <input
                                    name="name"
                                    required
                                    maxLength={120}
                                    autoFocus
                                    className="bg-surface-10 px-3 py-2 border border-surface-30 focus:border-primary-0 rounded-lg outline-none"
                                />
                            </label>
                            <label className="gap-1 grid text-sm">
                                Description
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
                                    className="bg-primary-0 disabled:opacity-50 px-4 py-2 rounded-lg text-surface-0"
                                >
                                    {saving ? "Creating…" : "Create board"}
                                </button>
                                <button
                                    type="button"
                                    onClick={() => setCreating(false)}
                                    className="px-4 py-2 border border-surface-30 rounded-lg"
                                >
                                    Cancel
                                </button>
                            </div>
                        </form>
                    )}

                    <label className="flex items-center gap-2 bg-surface-0/70 px-3 border border-surface-30 focus-within:border-primary-0 rounded-xl">
                        <Search size={18} aria-hidden="true" />
                        <input
                            type="search"
                            value={search}
                            onChange={(event) => setSearch(event.target.value)}
                            placeholder="Search boards by name"
                            aria-label="Search boards by name"
                            className="bg-transparent py-3 outline-none w-full"
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
                            My boards
                        </label>
                    )}
                </div>
            </div>

            <Scrollbar className="flex-1 min-h-0 overflow-x-hidden overscroll-contain">
                <div className="space-y-6 mx-auto px-4 sm:px-8 py-6 max-w-6xl">
                    <div className="gap-4 grid">
                        {items.map((board) => (
                            <Link
                                key={board.id}
                                href={`/board/${encodeURIComponent(board.slug)}`}
                                className={`group grid gap-4 rounded-2xl border border-surface-20 p-3 transition hover:border-primary-0/60 focus-visible:outline-2 focus-visible:outline-primary-0 ${board.image_count === 0 ? "bg-surface-10/60 text-primary-0/55" : "bg-surface-0/70 sm:grid-cols-[minmax(0,1fr)_minmax(320px,55%)] sm:items-center"}`}
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
                                            className={`mt-2 line-clamp-2 text-sm ${board.image_count === 0 ? "text-primary-0/50" : "text-primary-0/70"}`}
                                        >
                                            {board.description}
                                        </p>
                                    )}
                                    <p
                                        className={`mt-3 flex items-center gap-2 text-sm ${board.image_count === 0 ? "text-primary-0/50" : "text-primary-0/70"}`}
                                    >
                                        <Images size={16} /> {board.image_count}{" "}
                                        {board.image_count === 1
                                            ? "image"
                                            : "images"}
                                    </p>
                                </div>
                            </Link>
                        ))}
                    </div>
                    {!loading && !error && items.length === 0 && (
                        <p className="p-8 border border-surface-20 rounded-2xl text-center">
                            {query
                                ? "No boards match your search."
                                : "No boards yet. Create the first one."}
                        </p>
                    )}
                    {error && (
                        <div
                            role="alert"
                            className="flex items-center gap-3 text-red-500"
                        >
                            {error}
                            <button onClick={retry} className="underline">
                                Retry
                            </button>
                        </div>
                    )}
                    {loading && (
                        <p
                            role="status"
                            className="text-primary-0/70 text-center"
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
