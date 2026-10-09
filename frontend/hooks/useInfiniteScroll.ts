"use client";
import { useCallback, useEffect, useRef, useState } from "react";
import { useInView } from "react-intersection-observer";

interface Props<TQuery, TItem> {
    fetchFn: (query: TQuery, cursor: number, limit: number) => Promise<TItem[]>;
    query: TQuery;
    requestSize?: number;
    initRequestSize?: number;
    initialItems?: TItem[];
    initialReachedEnd?: boolean;
}

export default function useInfiniteScroll<TQuery, TItem>({
    fetchFn,
    query,
    requestSize = 30,
    initRequestSize = 40,
    initialItems = [],
    initialReachedEnd = false,
}: Props<TQuery, TItem>) {
    const [items, setItems] = useState<TItem[]>(initialItems);
    const [loading, setLoading] = useState(false);
    const [reachedEnd, setReachedEnd] = useState(initialReachedEnd);
    const [error, setError] = useState<string | null>(null);

    const fetchingRef = useRef(false);
    const reachedEndRef = useRef(initialReachedEnd);
    const loadingRef = useRef(false);
    const errorRef = useRef(false);
    const cursorRef = useRef(initialItems.length);
    const fetchIdRef = useRef(0);
    const queryRef = useRef(query);
    const queryKeyRef = useRef(JSON.stringify(query));

    const fetchItems = useCallback(
        async (query: TQuery, limit: number) => {
            if (fetchingRef.current || reachedEndRef.current || errorRef.current) return;

            const cursor = cursorRef.current;
            cursorRef.current += limit;
            fetchingRef.current = true;
            loadingRef.current = true;
            const myFetchId = fetchIdRef.current;
            setLoading(true);

            try {
                const data = await fetchFn(query, cursor, limit);
                if (!data) return;
                if (myFetchId !== fetchIdRef.current) return;
                setError(null);
                setItems((prev) => [...prev, ...data]);

                if (data.length < limit) {
                    reachedEndRef.current = true;
                    setReachedEnd(true);
                }
            } catch (error) {
                console.error("useInfiniteScroll fetch error:", error);
                if (myFetchId === fetchIdRef.current) {
                    cursorRef.current = cursor;
                    errorRef.current = true;
                    setError(error instanceof Error ? error.message : "Could not load results");
                }
            } finally {
                if (myFetchId === fetchIdRef.current) {
                    fetchingRef.current = false;
                    loadingRef.current = false;
                    setLoading(false);
                }
            }
        },
        [fetchFn],
    );

    // Sentinel — place <div ref={sentinelRef} /> at the bottom of your list,
    // inside the scrollable container. That's all you need.
    const { ref: sentinelRef, inView } = useInView({
        threshold: 0,
        rootMargin: "400px",
    });

    // Reset and re-fetch when query changes
    useEffect(() => {
        const queryKey = JSON.stringify(query);
        if (queryKey === queryKeyRef.current) return;
        queryKeyRef.current = queryKey;

        queryRef.current = query;
        fetchIdRef.current++;
        fetchingRef.current = false;
        loadingRef.current = false;
        reachedEndRef.current = false;
        errorRef.current = false;
        cursorRef.current = 0;

        setItems([]);
        setReachedEnd(false);
        setError(null);
        fetchItems(query, requestSize);
    }, [fetchItems, query, requestSize]);

    // Initial fetch
    useEffect(() => {
        if (initialItems.length > 0) return;
        fetchItems(query, initRequestSize);
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    // Fires when sentinel enters view (user scrolled down, or content doesn't fill screen)
    useEffect(() => {
        if (!inView || loadingRef.current || reachedEndRef.current || errorRef.current) return;
        fetchItems(queryRef.current, requestSize);
    }, [inView, fetchItems, requestSize]);

    // inView won't re-fire if it's already true when loading ends (sentinel
    // stayed visible the whole time). Re-check manually after each fetch.
    useEffect(() => {
        if (loading || !inView || reachedEndRef.current || errorRef.current) return;
        fetchItems(queryRef.current, requestSize);
    }, [loading, inView, fetchItems, requestSize]);

    const retry = () => {
        errorRef.current = false;
        setError(null);
        fetchItems(queryRef.current, requestSize);
    };

    const refresh = () => {
        fetchIdRef.current++;
        fetchingRef.current = false;
        loadingRef.current = false;
        reachedEndRef.current = false;
        errorRef.current = false;
        cursorRef.current = 0;
        setItems([]);
        setReachedEnd(false);
        setError(null);
        fetchItems(queryRef.current, requestSize);
    };

    return { items, loading, reachedEnd, error, retry, refresh, sentinelRef };
}
