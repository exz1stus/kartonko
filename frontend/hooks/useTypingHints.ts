import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useDebounce } from "use-debounce";

export const useTypingHints = (
    inputQuery: string,
    fetchHints: (query: string) => Promise<string[]>,
    onQueryMatchedHint?: () => void,
) => {
    const query = inputQuery.toLowerCase().trim();
    const [hints, setHints] = useState<string[]>([]);
    const [selectedIndex, setSelectedIndex] = useState<number>(-1);

    const [debouncedQuery] = useDebounce(query, 200);

    const currentHint = selectedIndex >= 0 ? hints[selectedIndex] : "";
    const difference = currentHint ? currentHint.slice(query.length) : "";
    const requestIdRef = useRef<number>(0);

    const selectNext = () => {
        if (selectedIndex === -1) return;
        if (selectedIndex === hints.length - 1) {
            setSelectedIndex(0);
            return;
        }

        setSelectedIndex((prev) => prev + 1);
    };

    const selectPrevious = () => {
        if (selectedIndex === -1) return;
        if (selectedIndex === 0) {
            setSelectedIndex(hints.length - 1);
            return;
        }

        setSelectedIndex((prev) => prev - 1);
    };

    const selectAtIndex = (index: number) => {
        if (index >= 0 && index < hints.length) {
            setSelectedIndex(index);
        }
    };

    const hideHint = useCallback(() => {
        setHints([]);
        setSelectedIndex(-1);
    }, []);

    const executeFetch = useCallback(
        async (targetQuery: string) => {
            const currentRequestId = ++requestIdRef.current;

            try {
                let fetchedHints = await fetchHints(targetQuery);
                if (currentRequestId !== requestIdRef.current) return;

                if (fetchedHints.length === 0) {
                    hideHint();
                    return;
                }
                setSelectedIndex(0);
                setHints(fetchedHints);
            } catch (error) {
                if (currentRequestId === requestIdRef.current) {
                    hideHint();
                }
            }
        },
        [fetchHints, hideHint],
    );

    const refresh = useCallback(async () => {
        await executeFetch(query);
    }, [executeFetch, query]);

    useEffect(() => {
        executeFetch(debouncedQuery);

        return () => {
            requestIdRef.current++;
        };
    }, [debouncedQuery, executeFetch]);

    useEffect(() => {
        if (currentHint.startsWith(query) && query.length > 0) return;

        hideHint();
    }, [query]);

    useEffect(() => {
        if (hints.length == 1 && query == hints[0]) onQueryMatchedHint?.();
    }, [currentHint, query, onQueryMatchedHint, hints]);

    return {
        hints,
        currentHint,
        difference,
        selectNext,
        selectPrevious,
        selectAtIndex,
        refresh,
    };
};

export default useTypingHints;
