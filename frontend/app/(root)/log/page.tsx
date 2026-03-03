"use client";
import { LogEntryData } from "@/lib/log";
import LogEntry from "@/components/Log/LogEntry";
import AuthGuard from "@/components/AuthGuard";
import Scrollbar from "@/components/template/Scrollbar";
import useInfiniteScroll from "@/hooks/useInfiniteScroll";
import { apiFetch } from "@/lib/apiFetch";

interface SearchLogEntriesQuery {}

const Log = () => {
    const fetchEntries = async (
        query: SearchLogEntriesQuery,
        cursor: number,
        limit: number,
    ) => {
        const res = await apiFetch(`/log?cursor=${cursor}&limit=${limit}`);
        const data = await res.json();

        return data;
    };

    const { items, loading, sentinelRef, containerRef, contentRef } =
        useInfiniteScroll<SearchLogEntriesQuery, LogEntryData>({
            fetchFn: fetchEntries,
            query: {},
            isQueryEmpty: () => true,
        });

    const entries = items.map((entry, index) => {
        return (
            <div key={index} className="w-full lg:max-w-[50%] h-full">
                <LogEntry data={entry} />
            </div>
        );
    });

    if (loading) return <div>Loading...</div>;

    return (
        <AuthGuard moderator={true}>
            <div ref={containerRef} className="h-full">
                <Scrollbar>
                    <div
                        ref={contentRef}
                        className="flex flex-col items-center gap-2 m-2 overflow-auto"
                    >
                        {entries}
                    </div>
                    <div ref={sentinelRef} style={{ height: 1 }} />
                </Scrollbar>
            </div>
        </AuthGuard>
    );
};

export default Log;
