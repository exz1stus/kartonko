"use client";
import React, { useEffect, useRef, useState } from "react";
import TagSelector, { TagSelectorRef } from "@/components/Tags/TagSelector";
import { useHover } from "@/contexts/HoverContex";
import { SearchQuery } from "@/lib/query";
import NameField from "../NameField";
import { useDebouncedCallback } from "use-debounce";

export function isQueryEmpty(query: SearchQuery) {
    return query.nameContains === "" && query.withTags?.length === 0;
}

export function isQueriesEqual(query1: SearchQuery, query2: SearchQuery) {
    return (
        query1.nameContains === query2.nameContains &&
        JSON.stringify(query1.withTags) === JSON.stringify(query2.withTags)
    );
}

interface Props {
    initialQuery?: SearchQuery;
    onQueryChange: (query: SearchQuery) => void;
    className?: string;
}

enum InsertingMode {
    NONE,
    NAME,
    TAG,
}

const ImageSearch: React.FC<Props> = ({
    initialQuery,
    onQueryChange,
    className,
}) => {
    const [insertingMode, setInsertingMode] = useState<InsertingMode>(
        InsertingMode.NONE,
    );
    const ref = useRef<HTMLDivElement>(null);
    const tagSelectorRef = React.useRef<TagSelectorRef>(null);

    const [tags, setTags] = useState<string[]>(initialQuery?.withTags || []);
    const [name, setName] = useState(initialQuery?.nameContains ?? "");

    const userID = initialQuery?.userID;

    const removeTag = (tag: string) => {
        setTags(tags.filter((t) => t !== tag));
    };

    const debouncedQuery = useDebouncedCallback(
        () =>
            onQueryChange({
                nameContains: name,
                withTags: tags,
                userID,
            }),
        200,
    );

    useEffect(() => {
        debouncedQuery();
    }, [debouncedQuery, name, tags, userID]);

    const onNameKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === "Escape") {
            e.preventDefault();
            setName("");
        }
    };

    return (
        <div ref={ref} className={className}>
            <div className="items-center gap-x-4 gap-y-1 grid grid-cols-[auto_1fr]">
                <span
                    className={`px-3 py-1 text-sm font-medium border rounded-full transition-colors select-none text-center
                ${
                    insertingMode === InsertingMode.NAME
                        ? "bg-surface-0/50 border-primary-0 text-primary-0"
                        : "border-surface-200 text-surface-600"
                }`}
                >
                    name
                </span>
                <div className="flex items-center gap-2">
                    <NameField
                        className="z-1 relative bg-transparent px-3 py-1 border-none outline-none w-full text-lg"
                        value={name}
                        onChangeSanitized={(value: string) => setName(value)}
                        onKeyDown={(e) => onNameKeyDown(e)}
                        onFocus={() => setInsertingMode(InsertingMode.NAME)}
                        onBlur={() => setInsertingMode(InsertingMode.NONE)}
                        placeholder={name.length === 0 ? "Search name" : ""}
                    />
                    <span className="inline-block h-10" />
                </div>
                <span
                    className={`px-3 py-1 text-sm font-medium border rounded-full transition-colors select-none text-center
                            ${
                                insertingMode === InsertingMode.TAG
                                    ? "bg-surface-0/50 border-primary-0"
                                    : "border-surface-200 text-surface-600"
                            }`}
                >
                    #
                </span>
                <div className="flex">
                    <TagSelector
                        ref={tagSelectorRef}
                        placeholder={"Start typing tags"}
                        tags={tags}
                        onTagsUpdate={(t: string[]) => setTags(t)}
                        removeTag={removeTag}
                        onFocus={() => setInsertingMode(InsertingMode.TAG)}
                        onBlur={() => setInsertingMode(InsertingMode.NONE)}
                        className="px-3 py-1"
                    />
                </div>
            </div>
        </div>
    );
};
export { type SearchQuery, ImageSearch };
