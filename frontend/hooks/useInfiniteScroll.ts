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
    isQueryEmpty: (query: TQuery) => boolean;
    isQueriesEqual?: (a: TQuery, b: TQuery) => boolean;
}

export default function useInfiniteScroll<TQuery, TItem>({
    fetchFn,
    query,
    requestSize = 30,
    initRequestSize = 70,
    initialItems = [],
    initialReachedEnd = false,
    isQueryEmpty,
}: Props<TQuery, TItem>) {
    const [items, setItems] = useState<TItem[]>(initialItems);
    const [reachedEnd, setReachedEnd] = useState<boolean>(initialReachedEnd);
    const [loading, setLoading] = useState(false);

    const hasQueryChangedFromInit = useRef<boolean>(false);
    const fetchingRef = useRef(false);
    const containerRef = useRef<HTMLDivElement>(null);
    const contentRef = useRef<HTMLDivElement>(null);

    const fetchCount = useRef<number>(0);
    const debugEvents = useRef<string[]>([]);

    const fetchItems = useCallback(
        async (query: TQuery, cursor: number, limit: number) => {
            if (fetchingRef.current || reachedEnd) return;
            fetchingRef.current = true;
            setLoading(true);

            try {
                const data = await fetchFn(query, cursor, limit);
                setItems((prev) => [...prev, ...data]);
                if (data.length < limit) setReachedEnd(true);
            } catch (error) {
                console.error("useInfiniteScroll fetch error:", error);
            } finally {
                fetchingRef.current = false;
                setLoading(false);
            }
            fetchCount.current++;
        },
        [fetchFn, reachedEnd],
    );

    const canFetch = (query: TQuery) => {
        if (!isQueryEmpty(query)) hasQueryChangedFromInit.current = true;

        return hasQueryChangedFromInit.current;
    };

    useEffect(() => {
        if (!canFetch(query)) return;

        setItems([]);
        setReachedEnd(false);
        fetchItems(query, 0, requestSize);
        debugEvents.current.push(
            `refetch on query cursor ${0} limit ${requestSize}`,
        );
    }, [query, requestSize]);

    useEffect(() => {
        if (initialItems.length > 0) return;
        fetchItems(query, items.length, initRequestSize);
        debugEvents.current.push(
            `initial fetch cursor ${items.length} limit ${requestSize}`,
        );
    }, []);

    const { ref: sentinelRef, inView } = useInView({
        threshold: 0,
        rootMargin: "400px",
    });

    useEffect(() => {
        const container = containerRef.current;
        const content = contentRef.current;

        if (!container || !content) return;

        const observer = new ResizeObserver(() => {
            if (fetchingRef.current || reachedEnd) return;

            if (content.scrollHeight <= container.clientHeight) {
                fetchItems(query, items.length, requestSize);
                debugEvents.current.push(
                    `fetch on resize cursor ${items.length} limit ${requestSize}`,
                );
            }
        });

        observer.observe(content);

        return () => observer.disconnect();
    }, [query, loading, reachedEnd, items.length]);

    // useEffect(() => {
    //     if (!inView || loading || reachedEnd) return;

    //     fetchItems(query, items.length, requestSize);
    //     debugEvents.current.push(
    //         `fetch on scroll cursor ${items.length} limit ${requestSize}`,
    //     );
    // }, [inView, query, loading, reachedEnd, items.length]);

    return {
        items,
        loading,
        reachedEnd,
        sentinelRef,
        containerRef,
        contentRef,
        debugEvents,
    };
}
