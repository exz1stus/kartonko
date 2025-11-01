"use client";
import React, { useEffect, useState } from 'react'
import { noUse } from '@/app/AudioEffects';
import useTypingHints from '@/app/hooks/useTypingHints';

interface Tag {
    name: string;
}

interface TagSelectorProps {
    selected: boolean
    tags: string[]
    onTagsUpdate: (tags: string[]) => void
}

interface TagHintResponse {
    tags: Tag[]
}

const API_ORIGIN = process.env.NEXT_PUBLIC_API_ORIGIN;

const TagSelector: React.FC<TagSelectorProps> = ({ selected, tags, onTagsUpdate }) => {
    const [tagField, setTagField] = useState<string>("");

    const onQueryMatchedHint = () => {
        write();
    }

    const FETCH_HINTS_LIMIT = 10;

    const fetchTagFieldHint = async (query: string) => {
        try {
            if (query.length === 0) return [];
            const response = await fetch(`${API_ORIGIN}/tags?query=${query}&limit=${FETCH_HINTS_LIMIT}`);
            if (response.ok) {
                const responseJson: TagHintResponse = await response.json();
                const parsedTags = responseJson.tags;
                if (parsedTags.length === 0) {
                    return [];
                }
                const hints = parsedTags
                    .filter((tag) => !tags.some(t => t === tag.name))
                    .map((tag) => tag.name);
                return hints;
            }
            return [];
        }
        catch (error) {
            console.log("Error fetching tag hint", error);
            return [];
        }
    }

    const write = async () => {
        if (hint.length <= 0 || tags.some(tag => tag === hint)) {
            noUse();
            return;
        }

        setTagField("");
        await new Promise(resolve => setTimeout(resolve, 0));
        let tagsQuery = tags.concat(hint);
        onTagsUpdate(tagsQuery);
    }

    const { hint, difference, selectNext, selectPrevious } = useTypingHints(tagField, fetchTagFieldHint, onQueryMatchedHint);

    const onTagKeyDown = (e: KeyboardEvent) => {
        const allowed = /^[a-z0-9\s]+$/i;
        if (!allowed.test(e.key))
            return;

        const insert = () => {
            setTagField(tagField => tagField.concat(e.key));
        }

        const remove = () => {
            setTagField(tagField => tagField.slice(0, -1));
        }

        const clear = () => {
            setTagField("");
        }

        const autoComplete = () => {
            if (difference.length <= 0) {
                noUse();
                return;
            }

            setTagField(tagField => tagField.concat(difference));
        }


        switch (e.key) {
            case "Tab":
                autoComplete();
                write();
                break;
            case "Backspace":
                remove();
                break;
            case "ArrowLeft":
                selectPrevious();
                break;
            case "ArrowRight":
                selectNext();
                break;
            case "Escape":
                clear();
                break;
            default:
                if (e.key.length === 1)
                    insert();
                break;
        }
    }

    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (!selected) return;
            onTagKeyDown(e)
        };

        window.addEventListener("keydown", handleKeyDown);
        return () => window.removeEventListener("keydown", handleKeyDown);
    }, [selected, difference]);

    return (
        <div className="flex flex-row items-center px-3">
            <span className={`text-lg ${hint.length > 0 ? "text-amber-200" : "text-primary-10"}`}>{tagField}</span>
            <span className="opacity-50 text-primary-10 text-lg">{difference}</span>
        </div>
    )
}
export default TagSelector;
export { type Tag, TagSelector }