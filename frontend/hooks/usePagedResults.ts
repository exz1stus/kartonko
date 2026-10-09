"use client";

import { useCallback, useEffect, useRef, useState } from "react";

export default function usePagedResults<T>(
    query: string,
    fetchPage: (query: string, cursor: number, limit: number) => Promise<T[]>,
    pageSize = 20,
    initialItems?: T[],
) {
    const [items, setItems] = useState<T[]>(initialItems ?? []);
    const [loading, setLoading] = useState(initialItems === undefined);
    const [hasMore, setHasMore] = useState(initialItems?.length === pageSize);
    const [error, setError] = useState<string | null>(null);
    const [revision, setRevision] = useState(0);
    const generation = useRef(0);
    const cursor = useRef(initialItems?.length ?? 0);
    const fetching = useRef(false);
    const initialPage = useRef(initialItems !== undefined);
    const initialQuery = useRef(query);

    const loadPage = useCallback(async (search: string, reset: boolean) => {
        if (fetching.current && !reset) return;
        const requestId = generation.current;
        fetching.current = true;
        setLoading(true);
        setError(null);
        try {
            const page = await fetchPage(search, cursor.current, pageSize);
            if (requestId !== generation.current) return;
            cursor.current += page.length;
            setItems((current) => reset ? page : [...current, ...page]);
            setHasMore(page.length === pageSize);
        } catch (cause) {
            if (requestId !== generation.current) return;
            setError(cause instanceof Error ? cause.message : "Could not load results");
        } finally {
            if (requestId === generation.current) {
                fetching.current = false;
                setLoading(false);
            }
        }
    }, [fetchPage, pageSize]);

    useEffect(() => {
        if (initialPage.current && revision === 0 && query === initialQuery.current) {
            return;
        }
        initialPage.current = false;
        generation.current += 1;
        cursor.current = 0;
        fetching.current = false;
        setItems([]);
        setHasMore(false);
        void loadPage(query, true);
        return () => { generation.current += 1; };
    }, [query, revision, loadPage]);

    const loadMore = () => {
        if (hasMore && !loading) void loadPage(query, false);
    };

    const retry = () => {
        if (!loading) void loadPage(query, cursor.current === 0);
    };

    const refresh = () => setRevision((current) => current + 1);

    return { items, loading, hasMore, error, loadMore, retry, refresh };
}
