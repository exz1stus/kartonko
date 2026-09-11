"use client";
import LogEntry from "@/components/Log/LogEntry";
import AuthGuard from "@/components/AuthGuard";
import Scrollbar from "@/components/template/Scrollbar";
import useInfiniteScroll from "@/hooks/useInfiniteScroll";
import { Loader } from "lucide-react";
import { getLog } from "@/lib/api/generated/client";
import { EntryResponse } from "@/lib/api/generated/model/entryResponse";
import { notFound } from "next/navigation";

interface SearchLogEntriesQuery {}

const Log = () => {
    const fetchEntries = async (
        query: SearchLogEntriesQuery,
        cursor: number,
        limit: number,
    ): Promise<EntryResponse[]> => {
        return getLog({ cursor, limit });
    };

    const { items, loading, reachedEnd, sentinelRef } = useInfiniteScroll<
        SearchLogEntriesQuery,
        EntryResponse
    >({
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

    return (
        <AuthGuard moderator={true}>
            <div className="h-full">
                <Scrollbar>
                    <div className="flex flex-col items-center gap-2 m-2 overflow-auto">
                        {entries}
                    </div>
                    {loading && (
                        <div className="flex justify-center items-center w-full">
                            <Loader />
                        </div>
                    )}
                    {!reachedEnd && <div ref={sentinelRef} />}
                </Scrollbar>
            </div>
        </AuthGuard>
    );
};

export default Log;
