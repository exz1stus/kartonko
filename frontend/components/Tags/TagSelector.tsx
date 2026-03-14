"use client";
import React, {
    forwardRef,
    useEffect,
    useImperativeHandle,
    useRef,
    useState,
} from "react";
import useTagHints from "./useTagHints";
import { useDialog } from "@/contexts/AlertDialogContext";
import TagSpan from "./TagSpan";
import { cn } from "@/lib/utils";
import { addNewTag } from "@/lib/tag.client";
import { toast } from "sonner";
import { sanitizeName } from "@/lib/sanitizeName";
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
        const [showNewTagChip, setShowNewTagChip] = useState<boolean>(false);
        const inputRef = React.useRef<HTMLInputElement>(null);
        const placeholder = "Add a tag...";

        const disableAutocomplete = useRef<boolean>(false);

        const dialog = useDialog();

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

        useEffect(() => {
            if (noHintTimerRef.current) clearTimeout(noHintTimerRef.current);
            if (query.trim().length > 0 && hint === "") {
                noHintTimerRef.current = setTimeout(() => {
                    setShowNewTagChip(true);
                }, NO_HINT_DELAY_MS);
            } else {
                setShowNewTagChip(false);
            }

            return () => {
                if (noHintTimerRef.current)
                    clearTimeout(noHintTimerRef.current);
            };
        }, [query, hint]);

        const showNewTagDialog = async () => {
            const result = await dialog(`Add tag ${query}?`, {
                confirmText: "Yes",
                cancelText: "No",
            });

            if (!result) return;
            toast.promise(addNewTag(query), {
                loading: "Loading...",
                success: () => {
                    onTagsUpdate(tags.concat(query));
                    setQuery("");
                    return `tag has been added`;
                },
                error: (error) => error.message,
            });
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
                    if (hint !== "") {
                        saveTag();
                        setQuery("");
                    } else if (showNewTagChip) {
                        showNewTagDialog();
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

        const onFieldChange = (value: string) => {
            const sanitized = sanitizeName(value);
            setQuery(sanitized);
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
                        className="z-1 relative bg-transparent border-none outline-none w-full text-lg"
                        ref={inputRef}
                        value={query}
                        onChange={(e) => onFieldChange(e.target.value)}
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

                    <div className="z-1 absolute inset-0 flex items-center text-lg">
                        <span className="invisible whitespace-pre pointer-events-none">
                            {query}
                        </span>
                        <span className="text-white/25 whitespace-pre pointer-events-none">
                            {difference}
                        </span>
                        <span
                            className={"px-2 whitespace-pre cursor-pointer"}
                            onClick={() => {
                                if (showNewTagChip) showNewTagDialog();
                            }}
                        >
                            {showNewTagChip && "+"}
                        </span>
                    </div>
                </div>
            </div>
        );
    },
);
TagSelector.displayName = "TagSelector";

export default TagSelector;
