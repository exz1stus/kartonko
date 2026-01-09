"use client";
import React, { useEffect, useState } from "react";
import useNameSelector from "./NameSelector";
import TagSelector from "./TagSelector";
import TagElement from "./TagElement";
import { useHover } from "@/app/contexts/HoverContex";

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
    initialQuery?: SearchQuery;
    selected: boolean;
    onQueryChange: (query: SearchQuery) => void;
    className?: string;
}

enum InsertingMode {
    NONE,
    NAME,
    TAG,
}

const ImageSearch: React.FC<Props> = ({ initialQuery, selected, onQueryChange, className }) => {
    const [tags, setTags] = useState<string[]>([]);
    const [name, setName] = useState<string>("");
    const [insertingMode, setInsertingMode] = useState<InsertingMode>(InsertingMode.NAME);

    const { ref, isHovered } = useHover<HTMLDivElement>();

    const hovered = isHovered();

    useEffect(() => {
        onQueryChange({
            nameContains: name,
            withTags: tags,
        });
    }, [name, tags]);

    const onNameUpdated = (name: string) => {
        setName(name);
    };

    const onTagsUpdate = (tags: string[]) => {
        setTags(tags);
    };

    useEffect(() => {
        if (initialQuery) {
            setTags(initialQuery.withTags);
            setName(initialQuery.nameContains);
        }
    }, [initialQuery]);

    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (!hovered) return;
            e.preventDefault();
            switch (e.key) {
                case "ArrowDown":
                    setInsertingMode(InsertingMode.TAG);
                    break;
                case "ArrowUp":
                    setInsertingMode(InsertingMode.NAME);
                    break;
            }
        };

        window.addEventListener("keydown", handleKeyDown);
        return () => window.removeEventListener("keydown", handleKeyDown);
    }, [hovered]);

    const removeTag = (tag: string) => {
        setTags(tags.filter((t) => t !== tag));
    };

    const tagElements = tags.map((tag, index) => (
        <TagElement key={index} tag={tag} removeTag={removeTag} />
    ));

    const active = selected || hovered;

    useNameSelector(active && insertingMode === InsertingMode.NAME, name, onNameUpdated);

    return (
        <div ref={ref} className={className}>
            <div className="gap-2 grid grid-rows-2">
                <div className="flex items-center gap-2">
                    <span
                        className={`px-2 border-1 rounded-2xl hover:cursor-pointer 
                        ${
                            insertingMode === InsertingMode.NAME
                                ? "bg-surface-0/50 border-primary-0"
                                : "border-transparent"
                        }`}
                        onClick={() => setInsertingMode(InsertingMode.NAME)}
                    >
                        name
                    </span>
                    <span className="text-3xl">{name}</span>
                    <span className="inline-block h-[2.5rem]" />
                </div>
                <div>
                    <div className="flex">
                        <span
                            className={`px-2 text-2xl border-1 rounded-2xl hover:cursor-pointer 
                            ${
                                insertingMode === InsertingMode.TAG
                                    ? "bg-surface-0/50 border-primary-0"
                                    : "border-transparent"
                            }`}
                            onClick={() => setInsertingMode(InsertingMode.TAG)}
                        >
                            #
                        </span>
                        {tagElements}
                        <TagSelector
                            selected={active && insertingMode === InsertingMode.TAG}
                            tags={tags}
                            onTagsUpdate={onTagsUpdate}
                        />
                    </div>
                </div>
            </div>
        </div>
    );
};
export { type SearchQuery, ImageSearch };
