"use client";

import { ReactNode, useCallback, useState } from "react";
import { Search } from "lucide-react";
import { useDebounce } from "use-debounce";
import { listBoards } from "@/lib/api/generated/client";
import { BoardResponse } from "@/lib/api/generated/model";
import { useAuth } from "@/contexts/AuthContext";
import useInfiniteScroll from "@/hooks/useInfiniteScroll";
import Scrollbar from "@/components/template/Scrollbar";

export default function BoardSearchList({ renderBoard, emptyMessage = "No editable boards found.", tiles = false }: {
    renderBoard: (board: BoardResponse) => ReactNode;
    emptyMessage?: string;
    tiles?: boolean;
}) {
    const { user } = useAuth();
    const [search, setSearch] = useState("");
    const [query] = useDebounce(search, 200);
    const ownerId = user?.privilege === "Moderator" ? undefined : user?.id;
    const fetchBoards = useCallback((name: string, cursor: number, limit: number) =>
        listBoards({ name, user_id: ownerId, cursor, limit }), [ownerId]);
    const boards = useInfiniteScroll({ query, fetchFn: fetchBoards, requestSize: 20, initialItems: [], initialReachedEnd: false });

    return <div className="flex min-h-0 flex-1 flex-col gap-2">
        <label className="flex items-center gap-2 rounded-lg border border-surface-30 px-2">
            <Search size={16} aria-hidden="true" />
            <input type="search" value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Find a board" aria-label="Find a board" className="min-w-0 flex-1 bg-transparent py-2 outline-none" />
        </label>
        <div className="min-h-0 flex-1"><Scrollbar className="min-h-0">
            <div className={tiles ? "space-y-1 lg:grid lg:grid-cols-3 lg:gap-2 lg:space-y-0 xl:grid-cols-5" : "space-y-1"}>
                {boards.items.map((board) => <div key={board.id}>{renderBoard(board)}</div>)}
                {!boards.reachedEnd && !boards.error && <div ref={boards.sentinelRef} className={tiles ? "h-1 lg:col-span-full" : "h-1"} />}
            </div>
            {boards.error && <p role="alert">{boards.error} <button onClick={boards.retry} className="underline">Retry</button></p>}
            {boards.loading && <p role="status">Loading boards…</p>}
            {!boards.loading && boards.items.length === 0 && <p className="p-3 text-sm text-primary-0/70">{emptyMessage}</p>}
        </Scrollbar></div>
    </div>;
}
