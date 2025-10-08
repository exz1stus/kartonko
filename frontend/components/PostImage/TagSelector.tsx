"use client";

import { useState, useEffect } from "react";

export interface Tag {
    name: string;
}

interface TagSelectorProps {
    selectedTags: Tag[];
    setSelectedTags: (tags: Tag[]) => void;
}
const API_ORIGIN = process.env.NEXT_PUBLIC_API_ORIGIN;

export function TagSelector({
    selectedTags,
    setSelectedTags,
}: TagSelectorProps) {
    const debounceTime = 100;

    const [search, setSearch] = useState("");
    const [tags, setTags] = useState<Tag[]>([]);

    useEffect(() => {
        const handler = setTimeout(() => {
            const fetchTags = async () => {
                if (search.trim() === "") {
                    setTags([]);
                    return;
                }
                try {
                    const res = await fetch(
                        `http://${API_ORIGIN}/search-tags/${encodeURIComponent(
                            search
                        )}`
                    );

                    const data = await res.json();
                    const recievedTags = data.tags as Tag[];

                    const filteredTags = recievedTags?.filter(
                        (tag) =>
                            !selectedTags.some(
                                (selected) => selected.name === tag.name
                            )
                    );

                    if (filteredTags) setTags(filteredTags);
                } catch (error) {
                    console.error("Failed to fetch tags", error);
                }
            };

            fetchTags();
        }, debounceTime);

        return () => clearTimeout(handler);
    }, [search, selectedTags]);

    const addTag = (tag: Tag) => {
        setSelectedTags([...selectedTags, tag]);
        setSearch("");
    };

    const removeTag = (tag: Tag) => {
        setSelectedTags(selectedTags.filter((t) => t !== tag));
    };

    return (
        <>
            <div className="flex flex-wrap gap-2 mb-4">
                {selectedTags?.map((tag) => (
                    <span
                        key={tag.name}
                        onClick={() => removeTag(tag)}
                        className="bg-blue-200 text-blue-800 px-3 py-1 rounded-full text-sm cursor-pointer"
                    >
                        {tag.name} ✕
                    </span>
                ))}
            </div>

            <input
                type="text"
                placeholder="Search tags..."
                className="flex-1 border border-gray-300 rounded px-3 py-2"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
            />

            <div className="flex flex-wrap gap-2 max-w-[50vh]">
                {tags.length > 0 ? (
                    tags.slice(0, 100000).map((tag) => (
                        <button
                            key={tag.name}
                            onClick={() => addTag(tag)}
                            className="px-3 py-0 bg-gray-500 rounded-full text-sm hover:bg-gray-300 transition"
                            type="button"
                            style={{ flex: "none" }}
                        >
                            {tag.name}
                        </button>
                    ))
                ) : search.length !== 0 ? (
                    <span>No tags found</span>
                ) : null}
            </div>
        </>
    );
}

//TODO : rewrite
