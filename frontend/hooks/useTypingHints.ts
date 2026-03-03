import { useEffect, useState } from "react";
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

    const hint = selectedIndex >= 0 ? hints[selectedIndex] : "";
    const difference = hint ? hint.slice(query.length) : "";

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

    const hideHint = () => {
        setHints([]);
        setSelectedIndex(-1);
    };

    useEffect(() => {
        if (debouncedQuery.length <= 0) return;
        if (
            hint.startsWith(debouncedQuery) &&
            debouncedQuery.length < hint.length
        )
            return;
        let cancelled = false;

        const fetchHintsAsync = async () => {
            const hints = await fetchHints(debouncedQuery);
            if (cancelled) return;
            if (hints.length === 0) {
                hideHint();
                return;
            }
            setSelectedIndex(0);
            setHints(hints);
        };

        fetchHintsAsync();
        return () => {
            cancelled = true;
        };
    }, [debouncedQuery, fetchHints]);

    useEffect(() => {
        if (hint.startsWith(query) && query.length > 0) return;

        hideHint();
    }, [query]);

    useEffect(() => {
        if (hints.length == 1 && query == hints[0]) onQueryMatchedHint?.();
    }, [hint, query, onQueryMatchedHint]);

    return { hint, difference, selectNext, selectPrevious };
};

export default useTypingHints;
