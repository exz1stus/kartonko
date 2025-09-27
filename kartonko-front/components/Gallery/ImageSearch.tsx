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
    }, [insertingMode]);

    let tagElements = tags.map((tag, index) => <span key={index}>{tag.name}</span>);

    return (
        <div className="flex flex-col items-center justify-center">
            <NameSelector selected={insertingMode === InsertingMode.NAME} name={name} onUpdateName={onNameUpdated} />
            <TagSelector selected={insertingMode === InsertingMode.TAG} tags={tags} onTagsUpdate={onTagsUpdate} />
            <div className="flex flex-col">{tagElements}</div>
        </div>
    )
}
export { type SearchQuery, ImageSearch };

