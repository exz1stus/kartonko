"use client";
import { useCallback, useEffect, useState } from "react";
import { noUse } from "@/app/AudioEffects";
import useTypingHints from "@/hooks/useTypingHints";
import { apiFetch } from "@/lib/apiFetch";

interface Props {
    tags: string[];
    query: string;
    setQuery: (query: string) => void;
    onTagsUpdate: (tags: string[]) => void;
    onQueryMatchedHint?: () => void;
}

interface TagHintResponse {
    tags: { name: string }[];
}

const useTagHints = ({
    tags,
    query,
    setQuery,
    onTagsUpdate,
    onQueryMatchedHint,
}: Props) => {
    const FETCH_HINTS_LIMIT = 10;

    const fetchTagHint = useCallback(
        async (tagQuery: string) => {
            try {
                const response = await apiFetch(
                    `/tags?query=${tagQuery}&limit=${FETCH_HINTS_LIMIT}`,
                );
                if (response.ok) {
                    const responseJson: TagHintResponse = await response.json();
                    const parsedTags = responseJson.tags;
                    if (parsedTags.length === 0) return [];
                    const hints = parsedTags
                        .map((tag) => tag.name)
                        .filter((tag) => !tags.some((t) => t === tag));
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

    const saveTag = async () => {
        if (hint.length <= 0 || tags.some((tag) => tag === hint)) {
            noUse();
            return;
        }

        setQuery("");
        await new Promise((resolve) => setTimeout(resolve, 0));
        let tagsQuery = tags.concat(hint);
        onTagsUpdate?.(tagsQuery);
    };

    const { hint, difference, selectNext, selectPrevious } = useTypingHints(
        query,
        fetchTagHint,
        onQueryMatchedHint,
    );

    return {
        hint,
        difference,
        autoComplete,
        saveTag,
        selectNext,
        selectPrevious,
    };
};
export default useTagHints;
