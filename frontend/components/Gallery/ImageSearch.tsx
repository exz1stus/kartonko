"use client";
import React, { useEffect, useState } from 'react'
import { noUse } from '@/app/AudioEffects';
import NameSelector from './NameSelector';
import { Tag } from '../PostImage/TagSelector';
import TagSelector from './TagSelector';
import TagElement from './TagElement';

interface SearchQuery {
    nameContains: string
    withTags: string[]
}

interface SearchOverlayProps {
    isGalleryHovered: boolean
    isSearchHovered: boolean
    onQueryChange: (query: SearchQuery) => void
}

enum InsertingMode {
    NONE,
    NAME,
    TAG
}

const ImageSearch: React.FC<SearchOverlayProps> = ({ isGalleryHovered, isSearchHovered, onQueryChange }) => {
    const [tags, setTags] = useState<Tag[]>([]);
    const [name, setName] = useState<string>("");
    const [insertingMode, setInsertingMode] = useState<InsertingMode>(InsertingMode.NAME);

    useEffect(() => {
        onQueryChange({
            nameContains: name,
            withTags: tags.map((t) => t.name)
        });
    }, [name, tags]);

    const onNameUpdated = (name: string) => {
        setName(name);
    }

    const onTagsUpdate = (tags: Tag[]) => {
        setTags(tags);
    }

    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (!isSearchHovered) return;
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
    }, [isSearchHovered]);

    const removeTag = (tag: Tag) => {
        setTags(tags.filter((t) => t !== tag));
    }

    const tagElements = tags.map((tag, index) => <TagElement key={index} tag={tag} removeTag={removeTag} />);

    const selected = isGalleryHovered || isSearchHovered;

    return (
        <div className="flex flex-col justify-center gap-2">
            <div className="flex items-center gap-2">
                <span
                    className={
                        `px-2 border-1 rounded-2xl hover:cursor-pointer 
                        ${insertingMode === InsertingMode.NAME ? "bg-surface-0/50 border-primary-0" : "border-transparent"}`
                    }
                    onClick={() => setInsertingMode(InsertingMode.NAME)}
                >
                    name
                </span>
                <NameSelector selected={selected && insertingMode === InsertingMode.NAME} name={name} onUpdateName={onNameUpdated} />
            </div>
            <div>
                <div className="flex">
                    <span
                        className={
                            `px-2 text-2xl border-1 rounded-2xl hover:cursor-pointer 
                            ${insertingMode === InsertingMode.TAG ? "bg-surface-0/50 border-primary-0" : "border-transparent"}`
                        }
                        onClick={() => setInsertingMode(InsertingMode.TAG)}
                    >
                        #
                    </span>
                    {tagElements}
                    <TagSelector selected={selected && insertingMode === InsertingMode.TAG} tags={tags} onTagsUpdate={onTagsUpdate} />
                </div>
            </div>
        </div >
    )
}
export { type SearchQuery, ImageSearch };

