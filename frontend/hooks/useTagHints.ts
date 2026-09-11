"use client";
import { useCallback } from "react";
import { noUse } from "@/app/AudioEffects";
import useTypingHints from "@/hooks/useTypingHints";
import { getTags } from "@/lib/api/generated/client";

interface Props {
    tags: string[];
    newTags?: string[];
    query: string;
    setQuery: (query: string) => void;
    onQueryMatchedHint?: () => void;
}

const useTagHints = ({ tags, query, setQuery, onQueryMatchedHint }: Props) => {
    const FETCH_HINTS_LIMIT = 10;

    const fetchTagHint = useCallback(
        async (tagQuery: string) => {
            const res = await getTags({
                prefix: tagQuery,
                limit: FETCH_HINTS_LIMIT,
            });
            let parsedTags = res.map((tag) => tag.name);
            if (parsedTags.length === 0) return [];
            const hints = parsedTags.filter(
                (tag) => !tags.some((t) => t === tag),
            );
            return hints;
        },
        [tags],
    );

    const autoComplete = () => {
        if (difference.length <= 0) {
            noUse();
            return;
        }

        setQuery(query.concat(difference));
    };

    const {
        hints,
        currentHint,
        difference,
        selectNext,
        selectPrevious,
        selectAtIndex,
        refresh,
    } = useTypingHints(query, fetchTagHint, onQueryMatchedHint);

    return {
        hints,
        currentHint,
        difference,
        autoComplete,
        selectAtIndex,
        selectNext,
        selectPrevious,
        refresh,
    };
};
export default useTagHints;
