"use client";
import React, { useEffect, useState } from "react";
import useNameSelector from "@/hooks/useNameSelector";
import TagSelector, { TagSelectorRef } from "@/components/Tags/TagSelector";
import { useHover } from "@/contexts/HoverContex";
import useTags from "@/components/Tags/useTags";

interface SearchQuery {
    nameContains: string;
    withTags: string[];
}

export function isQueryEmpty(query: SearchQuery) {
    return query.nameContains === "" && query.withTags.length === 0;
}

export function isQueriesEqual(query1: SearchQuery, query2: SearchQuery) {
    return (
        query1.nameContains === query2.nameContains &&
        JSON.stringify(query1.withTags) === JSON.stringify(query2.withTags)
    );
}

interface Props {
    selected: boolean;
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
    const { ref, isHovered } = useHover<HTMLDivElement>();

    const hovered = isHovered();

    const tagSelectorRef = React.useRef<TagSelectorRef>(null);
    const nameSelectorRef = React.useRef<HTMLInputElement>(null);

    const { tags, setTags, removeTag } = useTags(initialQuery?.withTags);
    const { onNameKeyDown, name } = useNameSelector(initialQuery?.nameContains);

    useEffect(() => {
        onQueryChange({
            nameContains: name,
            withTags: tags,
        });
    }, [name, tags]);

    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (!hovered) return;

            switch (e.key) {
                case "ArrowDown":
                    setInsertingMode(InsertingMode.TAG);
                    tagSelectorRef.current?.focus();
                    break;
                case "ArrowUp":
                    setInsertingMode(InsertingMode.NAME);
                    nameSelectorRef.current?.focus();
                    break;
            }
        };

        window.addEventListener("keydown", handleKeyDown);
        return () => window.removeEventListener("keydown", handleKeyDown);
    }, [hovered]);

    return (
        <div ref={ref} className={className}>
            <div className="gap-2 grid grid-rows-2">
                <div className="flex items-center gap-2">
                    <span
                        className={`px-2 border rounded-2xl hover:cursor-pointer 
                        ${
                            insertingMode === InsertingMode.NAME
                                ? "bg-surface-0/50 border-primary-0"
                                : "border-transparent"
                        }`}
                    >
                        name
                    </span>
                    <input
                        ref={nameSelectorRef}
                        className="z-10 relative bg-transparent border-none outline-none w-full text-3xl"
                        value={name}
                        onChange={() => {}}
                        onKeyDown={onNameKeyDown}
                        onFocus={() => setInsertingMode(InsertingMode.NAME)}
                        onBlur={() => setInsertingMode(InsertingMode.NONE)}
                        autoComplete="off"
                        spellCheck={false}
                        placeholder={name.length === 0 ? "Search name" : ""}
                    />
                    <span className="inline-block h-10" />
                </div>
                <div>
                    <div className="flex">
                        <span
                            className={`px-2 text-2xl border rounded-2xl hover:cursor-pointer 
                            ${
                                insertingMode === InsertingMode.TAG
                                    ? "bg-surface-0/50 border-primary-0"
                                    : "border-transparent"
                            }`}
                        >
                            #
                        </span>
                        <TagSelector
                            ref={tagSelectorRef}
                            tags={tags}
                            onTagsUpdate={(t: string[]) => setTags(t)}
                            removeTag={removeTag}
                            onFocus={() => setInsertingMode(InsertingMode.TAG)}
                        />
                    </div>
                </div>
            </div>
        </div>
    );
};
export { type SearchQuery, ImageSearch };
