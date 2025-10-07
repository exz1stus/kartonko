"use client";
import React, { useEffect, useState } from 'react'
import { Tag } from '../PostImage/TagSelector';
import TagSelector from '../Navbar/TagSelector';
import { noUse } from '@/app/ AudioEffects';
import NameSelector from './NameSelector';

interface SearchQuery {
    nameContains: string
    withTags: Tag[]
}

interface SearchOverlayProps {
    isHovered: boolean
    onQueryChange: (query: SearchQuery) => void
}

enum InsertingMode {
    NAME,
    TAG
}

const ImageSearch: React.FC<SearchOverlayProps> = ({ isHovered, onQueryChange }) => {
    const [tags, setTags] = useState<Tag[]>([]);
    const [name, setName] = useState<string>("");
    const [insertingMode, setInsertingMode] = useState<InsertingMode>(InsertingMode.NAME);


    useEffect(() => {
        onQueryChange({
            nameContains: name,
            withTags: tags
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
            const allowed = /^[a-z0-9\s]+$/i;
            if (!allowed.test(e.key)) {
                noUse();
                return;
            }

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
    }, [insertingMode]);

    let tagElements = tags.map((tag, index) => <span key={index} className="bg-blue-200 text-blue-800 px-3 py-1 rounded-full text-sm cursor-pointer">{tag.name}</span>);

    return (
        <div className="p-4 flex flex-col items-center justify-center">
            <NameSelector selected={isHovered && insertingMode === InsertingMode.NAME} name={name} onUpdateName={onNameUpdated} />
            <TagSelector selected={isHovered && insertingMode === InsertingMode.TAG} tags={tags} onTagsUpdate={onTagsUpdate} />
            <div className="flex flex-row gap-2">{tagElements}</div>
        </div>
    )
}
export { type SearchQuery, ImageSearch };

