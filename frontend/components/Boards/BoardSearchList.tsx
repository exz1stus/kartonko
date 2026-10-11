"use client";

import { ReactNode, useCallback, useState } from "react";
import { Search } from "lucide-react";
import { useDebounce } from "use-debounce";
import { listBoards } from "@/lib/api/generated/client";
import { BoardResponse } from "@/lib/api/generated/model";
import { useAuth } from "@/contexts/AuthContext";
import useInfiniteScroll from "@/hooks/useInfiniteScroll";
import Scrollbar from "@/components/template/Scrollbar";
import { useTranslations } from "@/contexts/LocaleContext";

export default function BoardSearchList({ renderBoard, emptyMessage = "No editable boards found.", tiles = false }: {
    renderBoard: (board: BoardResponse) => ReactNode;
    emptyMessage?: string;
    tiles?: boolean;
}) {
    const { user } = useAuth();
    const { t } = useTranslations();
    const [search, setSearch] = useState("");
    const [query] = useDebounce(search, 200);
    const ownerId = user?.privilege === "Moderator" ? undefined : user?.id;
    const fetchBoards = useCallback((name: string, cursor: number, limit: number) =>
        listBoards({ name, user_id: ownerId, cursor, limit }), [ownerId]);
    const { items, loading, reachedEnd, error, retry, sentinelRef } = useInfiniteScroll({ query, fetchFn: fetchBoards, requestSize: 20, initialItems: [], initialReachedEnd: false });

    return <div className="flex min-h-0 flex-1 flex-col gap-2">
        <label className="control-shell flex items-center gap-2 px-3">
            <Search size={16} aria-hidden="true" />
            <input type="search" value={search} onChange={(event) => setSearch(event.target.value)} placeholder={t("search.findBoard")} aria-label={t("search.findBoard")} className="min-w-0 flex-1 py-3 outline-none" />
        </label>
        <div className="min-h-0 flex-1"><Scrollbar className="min-h-0">
            <div className={tiles ? "space-y-1 lg:grid lg:grid-cols-3 lg:gap-2 lg:space-y-0 xl:grid-cols-5" : "space-y-1"}>
                {items.map((board) => <div key={board.id}>{renderBoard(board)}</div>)}
                {!reachedEnd && !error && <div ref={sentinelRef} className={tiles ? "h-1 lg:col-span-full" : "h-1"} />}
            </div>
            {error && <p role="alert">{error} <button onClick={retry} className="underline">{t("common.retry")}</button></p>}
            {loading && <p role="status">{t("common.loading")}</p>}
            {!loading && items.length === 0 && <p className="p-3 text-sm text-muted-foreground">{emptyMessage}</p>}
        </Scrollbar></div>
    </div>;
}
