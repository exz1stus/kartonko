"use client";
import React, {
    forwardRef,
    useImperativeHandle,
    useRef,
    useState,
} from "react";
import useTagHints from "./useTagHints";
import TagSpan from "./TagSpan";
import { cn } from "@/lib/utils";

interface Props {
    tags: string[];
    removeTag: (tag: string) => void;
    onTagsUpdate: (tags: string[]) => void;
    onFocus?: () => void;
    onBlur?: () => void;
}

export interface TagSelectorRef {
    focus: () => void;
}

const TagSelector = forwardRef<TagSelectorRef, Props>(
    ({ tags, removeTag, onTagsUpdate, onFocus, onBlur }, ref) => {
        const [query, setQuery] = useState<string>("");
        const [active, setActive] = useState<boolean>(false);
        const inputRef = React.useRef<HTMLInputElement>(null);
        const placeholder = "Add a tag...";

        const disableAutocomplete = useRef<boolean>(false);

        useImperativeHandle(ref, () => ({
            focus: () => {
                inputRef.current?.focus();
            },
        }));

        const onQueryMatchedHint = () => {
            if (disableAutocomplete.current) return;
            saveTag();
        };

        const {
            hint,
            difference,
            autoComplete,
            saveTag,
            selectNext,
            selectPrevious,
        } = useTagHints({
            tags,
            query,
            setQuery,
            onTagsUpdate,
            onQueryMatchedHint,
        });

        const tryRemove = () => {
            if (query.length !== 0) {
                setQuery((prev) => prev.slice(0, prev.length - 1));
                return;
            }

            if (tags.length === 0) return;
            disableAutocomplete.current = true;
            const last = tags[tags.length - 1];
            removeTag(last);
            setQuery(last);
        };

        const removeLast = () => {
            if (tags.length === 0) return;
            const last = tags[tags.length - 1];
            removeTag(last);
        };

        const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
            if (!active) return;
            // const allowed = /^[a-z0-9\s]+$/i;
            // if (!allowed.test(e.key)) return;
            const complete = "Tab";
            const remove = "Backspace";
            const previous = "ArrowUp";
            const next = "ArrowDown";
            const cancel = "Escape";

            const keys = [complete, remove, previous, next, cancel];

            if (keys.includes(e.key)) e.preventDefault();
            else disableAutocomplete.current = false;

            switch (e.key) {
                case complete:
                    if (hint !== "") {
                        autoComplete();
                        saveTag();
                        removeTag(hint);
                        setQuery("");
                    }
                    break;
                case remove:
                    if (e.ctrlKey || e.metaKey) {
                        if (query.length === 0) removeLast();
                        else setQuery("");
                    } else tryRemove();
                    break;
                case previous:
                    selectPrevious();
                    break;
                case next:
                    selectNext();
                    break;
                case cancel:
                    setQuery("");
                    break;
                default:
                    break;
            }
        };

        return (
            <div
                onClick={() => inputRef.current?.focus()}
                className={cn(
                    "inline-flex relative flex-wrap items-center gap-2 px-3 py-2 rounded-md w-full text-lg",
                    active,
                )}
            >
                <TagSpan
                    tags={tags}
                    removeTag={(t) => removeTag(t)}
                    tagStyle={"bg-surface-20 rounded-3xl px-2"}
                />
                <div className="relative flex-1 min-w-10">
                    <input
                        className="z-10 relative bg-transparent border-none outline-none w-full text-lg"
                        ref={inputRef}
                        value={query}
                        onChange={(e) => setQuery(e.target.value)}
                        onFocus={() => {
                            setActive(true);
                            onFocus?.();
                        }}
                        onBlur={() => {
                            setActive(false);
                            onBlur?.();
                        }}
                        onKeyDown={handleKeyDown}
                        autoComplete="off"
                        spellCheck={false}
                        placeholder={
                            query.length === 0 && tags.length === 0
                                ? placeholder
                                : ""
                        }
                    />

                    {/* Hint overlay */}
                    <div className="absolute inset-0 flex items-center text-lg pointer-events-none">
                        <span className="invisible whitespace-pre">
                            {query}
                        </span>
                        <span className="text-white/25 whitespace-pre">
                            {query.length !== 0 && difference}
                        </span>
                    </div>
                </div>
            </div>
        );
    },
);

export default TagSelector;
