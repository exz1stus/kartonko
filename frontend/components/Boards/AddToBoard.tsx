"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import Link from "next/link";
import { Check, FolderPlus, Search } from "lucide-react";
import { useDebounce } from "use-debounce";
import { listBoardIDsForImage, listBoards, postBoardImage } from "@/lib/api/generated/client";
import { useAuth } from "@/contexts/AuthContext";
import usePagedResults from "@/hooks/usePagedResults";

export default function AddToBoard({ imageId }: { imageId: number }) {
    const { user } = useAuth();
    const [open, setOpen] = useState(false);
    if (!user) return null;
    return <div className="rounded-xl border border-surface-20 p-3">
        <button onClick={() => setOpen(!open)} aria-expanded={open} className="inline-flex items-center gap-2"><FolderPlus size={18} /> Add to board</button>
        {open && <BoardPicker imageId={imageId} />}
    </div>;
}

function BoardPicker({ imageId }: { imageId: number }) {
    const { user } = useAuth();
    const [search, setSearch] = useState("");
    const [query] = useDebounce(search.trim(), 250);
    const [pendingIds, setPendingIds] = useState<Set<number>>(new Set());
    const pendingRef = useRef(new Set<number>());
    const [addedIds, setAddedIds] = useState<Set<number>>(new Set());
    const [membershipLoaded, setMembershipLoaded] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [notice, setNotice] = useState<string | null>(null);
    const ownerId = user?.privilege === "Moderator" ? undefined : user?.id;
    const fetchPage = useCallback((name: string, cursor: number, limit: number) => listBoards({ name, user_id: ownerId, cursor, limit }), [ownerId]);
    const boards = usePagedResults(query, fetchPage, 20);

    const loadMembership = useCallback(async () => {
        setMembershipLoaded(false);
        try {
            setAddedIds(new Set(await listBoardIDsForImage(imageId)));
            setMembershipLoaded(true);
        } catch (cause) {
            setError(cause instanceof Error ? cause.message : "Could not check boards");
        }
    }, [imageId]);

    useEffect(() => { void loadMembership(); }, [loadMembership]);

    async function add(boardId: number, boardName: string) {
        if (!membershipLoaded || addedIds.has(boardId) || pendingRef.current.has(boardId)) return;
        pendingRef.current.add(boardId);
        setPendingIds(new Set(pendingRef.current));
        setError(null);
        setNotice(null);
        try {
            await postBoardImage(boardId, { image_id: imageId });
            setAddedIds((current) => new Set(current).add(boardId));
            setNotice(`Added to ${boardName}.`);
        } catch (cause) {
            setError(cause instanceof Error ? cause.message : "Could not add image");
            if (cause instanceof Error && cause.message.includes("already on this board")) {
                setAddedIds((current) => new Set(current).add(boardId));
            }
        } finally {
            pendingRef.current.delete(boardId);
            setPendingIds(new Set(pendingRef.current));
        }
    }

    if (!user) return null;
    return <div className="mt-3 space-y-3">
            <label className="flex items-center gap-2 rounded-lg border border-surface-30 px-2"><Search size={16} /><input type="search" value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Find a board" aria-label="Find a board" className="min-w-0 flex-1 bg-transparent py-2 outline-none" /></label>
            {notice && <p role="status" className="text-sm">{notice}</p>}
            {error && <p role="alert" className="text-sm text-red-500">{error}</p>}
            <div className="max-h-56 space-y-1 overflow-y-auto">
                {boards.items.filter((board) => board.user_id === user.id || user.privilege === "Moderator").map((board) => <button key={board.id} disabled={!membershipLoaded || addedIds.has(board.id) || pendingIds.has(board.id)} onClick={() => add(board.id, board.name)} className="flex w-full items-center justify-between gap-2 truncate rounded-lg px-2 py-2 text-left hover:bg-surface-20 disabled:opacity-60"><span className="truncate">{board.name}</span>{addedIds.has(board.id) ? <span className="inline-flex items-center gap-1 text-green-600 text-xs"><Check size={14} /> Added</span> : pendingIds.has(board.id) ? <span className="text-xs">Adding…</span> : null}</button>)}
            </div>
            {boards.error && <p role="alert" className="text-sm text-red-500">{boards.error} <button onClick={boards.retry} className="underline">Retry</button></p>}
            {boards.loading && <p role="status" className="text-sm">Loading boards…</p>}
            {!membershipLoaded && !error && <p role="status" className="text-sm">Checking boards…</p>}
            {!membershipLoaded && error && <button onClick={() => void loadMembership()} className="text-sm underline">Retry board check</button>}
            {boards.hasMore && !boards.loading && <button onClick={boards.loadMore} className="text-sm underline">Load more boards</button>}
            <Link href="/boards" className="block text-sm underline">Create a board</Link>
    </div>;
}
