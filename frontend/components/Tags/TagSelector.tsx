"use client";
import React, {
    forwardRef,
    useEffect,
    useImperativeHandle,
    useRef,
    useState,
} from "react";
import useTagHints from "./useTagHints";
import TagSpan from "./TagSpan";
import { cn } from "@/lib/utils";
import { sanitizeName } from "@/lib/sanitizeName";
import useSanitizedField from "@/hooks/useSanitizedField";
import { noUse } from "@/app/AudioEffects";
import useUploadStore, { selectGlobalNewTags } from "@/hooks/useUploadStore";
import { useShallow } from "zustand/react/shallow";
interface Props {
    tags: string[];
    removeTag: (tag: string) => void;
    onTagsUpdate: (tags: string[]) => void;

    newTags?: string[];
    removeNewTag?: (tag: string) => void;
    onNewTagsUpdate?: (tags: string[]) => void;

    className?: string;
    placeholder?: string;
    onFocus?: () => void;
    onBlur?: () => void;
}

export interface TagSelectorRef {
    focus: () => void;
}

const TagSelector = forwardRef<TagSelectorRef, Props>(
    (
        {
            tags,
            removeTag,
            onTagsUpdate,
            newTags,
            onNewTagsUpdate,
            removeNewTag,
            onFocus,
            onBlur,
            placeholder = "Add a tag...",
            className,
        },
        ref,
    ) => {
        const [query, setQuery] = useState<string>("");
        const [active, setActive] = useState<boolean>(false);
        const inputRef = React.useRef<HTMLInputElement>(null);

        const [queryNew, setQueryNew] = useState<boolean>(false);

        const { sanitized, onSanitize } = useSanitizedField(
            sanitizeName,
            setQuery,
        );

        const globalNewTags = useUploadStore(useShallow(selectGlobalNewTags));
        const newTagsSet = new Set(newTags);
        const remainingGlobalTags = globalNewTags
            .filter((tag) => !newTagsSet.has(tag))
            .filter((tag) => tag.startsWith(query));

        const disableAutocomplete = useRef<boolean>(false);

        const NO_HINT_DELAY_MS = 600;
        const noHintTimerRef = useRef<ReturnType<typeof setTimeout> | null>(
            null,
        );

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
            hints,
            currentHint,
            difference,
            autoComplete,
            selectNext,
            selectPrevious,
            refresh,
        } = useTagHints({
            tags,
            query,
            setQuery,
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

        const saveTag = async (tagToSend?: string) => {
            const finalTag = tagToSend || currentHint;

            if (finalTag.length <= 0 || tags.some((tag) => tag === finalTag)) {
                noUse();
                return;
            }

            setQuery("");
            await new Promise((resolve) => setTimeout(resolve, 0));
            let tagsQuery = tags.concat(finalTag);
            onTagsUpdate?.(tagsQuery);
        };

        useEffect(() => {
            if (!newTags) return;
            if (noHintTimerRef.current) clearTimeout(noHintTimerRef.current);
            const queryTrimmed = query.trim();
            if (
                queryTrimmed.length > 0 &&
                currentHint === "" &&
                !tags.includes(queryTrimmed) &&
                !newTags.includes(queryTrimmed)
            ) {
                noHintTimerRef.current = setTimeout(() => {
                    setQueryNew(true);
                }, NO_HINT_DELAY_MS);
            } else {
                setQueryNew(false);
            }

            return () => {
                if (noHintTimerRef.current)
                    clearTimeout(noHintTimerRef.current);
            };
        }, [query, currentHint, newTags, tags]);

        const addNewTag = (tag: string) => {
            if (!newTags || !onNewTagsUpdate) return;

            onNewTagsUpdate(newTags.concat(tag));
            setQuery("");
        };

        const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
            if (!active) return;
            // const allowed = /^[a-z0-9\s]+$/i;
            // if (!allowed.test(e.key)) return;
            const complete = "Tab";
            const comma = ",";
            const remove = "Backspace";
            const previous = "ArrowUp";
            const next = "ArrowDown";
            const cancel = "Escape";

            const keys = [complete, remove, previous, next, cancel, comma];

            if (keys.includes(e.key)) e.preventDefault();
            else disableAutocomplete.current = false;

            switch (e.key) {
                case comma:
                case complete:
                    if (currentHint !== "") {
                        saveTag();
                        setQuery("");
                    } else if (queryNew) {
                        addNewTag(query);
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

        const showList =
            active &&
            (hints.length > 0 || remainingGlobalTags.length > 0 || queryNew);

        return (
            <div className="relative w-full">
                <div
                    onClick={() => inputRef.current?.focus()}
                    className={cn(
                        "border-invisible inline-flex flex-wrap items-center gap-2 border w-full text-lg",
                        className,
                        active && "bg-neutral-900 ",
                        !showList
                            ? "rounded-md border-surface-10"
                            : " rounded-t-md",
                    )}
                >
                    <TagSpan
                        tags={tags}
                        removeTag={(t) => removeTag(t)}
                        tagStyle={"bg-surface-20 rounded-3xl px-2"}
                    />
                    <div className="relative flex-1 min-w-10">
                        <input
                            className={cn(
                                "z-1 relative bg-transparent border-none outline-none w-full text-lg transition-colors",
                                { "text-red-500": !sanitized },
                            )}
                            ref={inputRef}
                            value={query}
                            onChange={(e) => onSanitize(e.target.value)}
                            onFocus={() => {
                                setActive(true);
                                onFocus?.();
                            }}
                            onBlur={(e) => {
                                setActive(false);
                                onBlur?.();
                            }}
                            onKeyDown={handleKeyDown}
                            autoComplete="off"
                            spellCheck={false}
                            placeholder={
                                query.length === 0 &&
                                tags.length === 0 &&
                                (currentHint.length === 0 || !active)
                                    ? placeholder
                                    : ""
                            }
                        />
                        <div className="z-1 absolute inset-0 flex items-center text-lg">
                            <span className="invisible whitespace-pre pointer-events-none">
                                {query}
                            </span>
                            <span className="text-white/25 whitespace-pre pointer-events-none">
                                {active && query.length > 0 && difference}
                            </span>
                        </div>
                    </div>
                    {newTags && removeNewTag && (
                        <TagSpan
                            tags={newTags}
                            removeTag={(t) => removeNewTag(t)}
                            tagStyle={
                                "bg-surface-20 rounded-3xl px-2 border border-primary"
                            }
                        />
                    )}
                </div>
                {showList && (
                    <div className="top-full right-0 left-0 z-50 absolute bg-neutral-900 shadow-lg border-x border-b rounded-b-md max-h-60 overflow-y-auto">
                        <ul className="py-1 text-neutral-200 text-sm">
                            {hints.map((hintItem: string, index: number) => (
                                <li
                                    key={index}
                                    onMouseDown={(e) => {
                                        e.preventDefault();
                                        saveTag(hintItem);
                                    }}
                                    className={cn(
                                        "hover:bg-neutral-800 px-4 py-2 transition-colors cursor-pointer",
                                    )}
                                >
                                    {hintItem}
                                </li>
                            ))}

                            {queryNew && (
                                <li
                                    onMouseDown={(e) => {
                                        e.preventDefault();
                                        addNewTag(query);
                                    }}
                                    className="hover:bg-neutral-800 px-4 py-2 border-neutral-800 border-t font-medium text-primary-400 cursor-pointer"
                                >
                                    + Add new tag: &quot;{query}&quot;
                                </li>
                            )}
                            {remainingGlobalTags.map((tag, index) => (
                                <li
                                    key={`${tag}-${index}`}
                                    onMouseDown={(e) => {
                                        e.preventDefault();
                                        addNewTag(tag);
                                    }}
                                    className="hover:bg-neutral-800 px-4 py-2 italic transition-colors cursor-pointer"
                                >
                                    {tag}
                                </li>
                            ))}
                        </ul>
                    </div>
                )}
            </div>
        );
    },
);
TagSelector.displayName = "TagSelector";

export default TagSelector;
