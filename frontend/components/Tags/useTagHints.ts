"use client";
import { useCallback } from "react";
import { noUse } from "@/app/AudioEffects";
import useTypingHints from "@/hooks/useTypingHints";
import { apiFetch } from "@/lib/apiFetch";

interface Props {
    tags: string[];
    newTags?: string[];
    query: string;
    setQuery: (query: string) => void;
    onQueryMatchedHint?: () => void;
}

interface TagHintResponse {
    tags: { name: string }[];
}

const useTagHints = ({ tags, query, setQuery, onQueryMatchedHint }: Props) => {
    const FETCH_HINTS_LIMIT = 10;

    const fetchTagHint = useCallback(
        async (tagQuery: string) => {
            try {
                const response = await apiFetch(
                    `/tags?query=${tagQuery}&limit=${FETCH_HINTS_LIMIT}`,
                );
                if (response.ok) {
                    const responseJson: TagHintResponse = await response.json();
                    let parsedTags = responseJson.tags.map((tag) => tag.name);
                    if (parsedTags.length === 0) return [];
                    const hints = parsedTags.filter(
                        (tag) => !tags.some((t) => t === tag),
                    );
                    return hints;
                }
                return [];
            } catch (error) {
                return [];
            }
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
