import { useEffect, useState } from "react";
import { noUse } from "@/app/AudioEffects";
import useTypingHints from "@/hooks/useTypingHints";
import { apiFetch } from "@/lib/apiFetch";

interface Props {
    active: boolean;
    tags: string[];
    onTagsUpdate: (tags: string[]) => void;
}

interface TagHintResponse {
    tags: { name: string }[];
}

const useTagSelector = ({ active, tags, onTagsUpdate }: Props) => {
    const [tagField, setTagField] = useState<string>("");

    const onQueryMatchedHint = () => {
        write();
    };

    const FETCH_HINTS_LIMIT = 10;

    const fetchTagFieldHint = async (query: string) => {
        try {
            if (query.length === 0) return [];
            const response = await apiFetch(`/tags?query=${query}&limit=${FETCH_HINTS_LIMIT}`);
            if (response.ok) {
                const responseJson: TagHintResponse = await response.json();
                const parsedTags = responseJson.tags;
                if (parsedTags.length === 0) {
                    return [];
                }
                const hints = parsedTags
                    .map((tag) => tag.name)
                    .filter((tag) => !tags.some((t) => t === tag));
                return hints;
            }
            return [];
        } catch (error) {
            console.log("Error fetching tag hint", error);
            return [];
        }
    };

    const write = async () => {
        if (hint.length <= 0 || tags.some((tag) => tag === hint)) {
            noUse();
            return;
        }

        setTagField("");
        await new Promise((resolve) => setTimeout(resolve, 0));
        let tagsQuery = tags.concat(hint);
        onTagsUpdate?.(tagsQuery);
    };

    const { hint, difference, selectNext, selectPrevious } = useTypingHints(
        tagField,
        fetchTagFieldHint,
        onQueryMatchedHint,
    );

    const onTagKeyDown = (e: KeyboardEvent) => {
        const allowed = /^[a-z0-9\s]+$/i;
        if (!allowed.test(e.key)) return;

        const insert = () => {
            setTagField((tagField) => tagField.concat(e.key));
        };

        const remove = () => {
            setTagField((tagField) => tagField.slice(0, -1));
        };

        const clear = () => {
            setTagField("");
        };

        const autoComplete = () => {
            if (difference.length <= 0) {
                noUse();
                return;
            }

            setTagField((tagField) => tagField.concat(difference));
        };

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
                if (e.key.length === 1) insert();
                break;
        }
    };

    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (!active) return;
            onTagKeyDown(e);
        };

        window.addEventListener("keydown", handleKeyDown);
        return () => window.removeEventListener("keydown", handleKeyDown);
    }, [active, difference]);

    return { hint, tagField, difference };
};
export default useTagSelector;
