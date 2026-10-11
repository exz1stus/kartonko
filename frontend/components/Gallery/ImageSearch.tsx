"use client";
import React, { useEffect, useRef, useState } from "react";
import TagSelector, { TagSelectorRef } from "@/components/Tags/TagSelector";
import { SearchQuery } from "@/lib/query";
import { sanitizeName } from "@/lib/sanitizeName";
import { useDebouncedCallback } from "use-debounce";
import { useTranslations } from "@/contexts/LocaleContext";
import { Search, X } from "lucide-react";

export function isQueryEmpty(query: SearchQuery) {
    return query.prefix === "" && query.tags?.length === 0;
}

export function isQueriesEqual(query1: SearchQuery, query2: SearchQuery) {
    return (
        query1.prefix === query2.prefix &&
        JSON.stringify(query1.tags) === JSON.stringify(query2.tags)
    );
}

interface Props {
    initialQuery?: SearchQuery;
    onQueryChange: (query: SearchQuery) => void;
    className?: string;
}

enum InsertingMode {
    NONE,
    NAME,
    TAG,
}

const ImageSearch: React.FC<Props> = ({
    initialQuery,
    onQueryChange,
    className,
}) => {
    const [insertingMode, setInsertingMode] = useState<InsertingMode>(
        InsertingMode.NONE,
    );
    const ref = useRef<HTMLDivElement>(null);
    const nameInputRef = useRef<HTMLInputElement>(null);
    const tagSelectorRef = React.useRef<TagSelectorRef>(null);
    const { t } = useTranslations();

    const [tags, setTags] = useState<string[]>(initialQuery?.tags || []);
    const [name, setName] = useState(initialQuery?.prefix ?? "");

    const userID = initialQuery?.user_id;

    const removeTag = (tag: string) => {
        setTags(tags.filter((t) => t !== tag));
    };

    const debouncedQuery = useDebouncedCallback(
        () =>
            onQueryChange({
                prefix: name,
                tags: tags,
                user_id: userID,
            }),
        200,
    );

    useEffect(() => {
        debouncedQuery();
    }, [debouncedQuery, name, tags, userID]);

    const onNameKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === "Escape") {
            e.preventDefault();
            setName("");
        }
    };

    return (
        <div ref={ref} className={`interface-opacity-group ${className ?? ""}`}>
            <div className="grid grid-cols-[4.5rem_minmax(0,1fr)] items-center gap-x-3 gap-y-2">
                <span
                    className={`filter-chip px-3 py-1 text-sm font-medium transition-colors select-none text-center
                ${
                    insertingMode === InsertingMode.NAME
                        ? "filter-chip-active"
                        : ""
                }`}
                >
                    {t("gallery.name")}
                </span>
                <div className="control-shell flex min-h-11 min-w-0 items-center gap-2 px-3">
                    <Search aria-hidden="true" size={18} className="shrink-0 text-[var(--control-placeholder)]" />
                    <input
                        ref={nameInputRef}
                        type="search"
                        className="min-w-0 flex-1 py-2 text-base"
                        value={name}
                        onChange={(event) => setName(sanitizeName(event.target.value))}
                        onKeyDown={(e) => onNameKeyDown(e)}
                        onFocus={() => setInsertingMode(InsertingMode.NAME)}
                        onBlur={() => setInsertingMode(InsertingMode.NONE)}
                        placeholder={t("gallery.searchName")}
                        aria-label={t("gallery.searchName")}
                        autoComplete="off"
                        spellCheck={false}
                    />
                    <button
                        type="button"
                        disabled={!name}
                        onClick={() => {
                            setName("");
                            nameInputRef.current?.focus();
                        }}
                        aria-label={t("gallery.clearName")}
                        className="control-clear flex size-7 shrink-0 items-center justify-center rounded-md text-[var(--control-placeholder)] transition-colors disabled:pointer-events-none disabled:opacity-0"
                    >
                        <X aria-hidden="true" size={16} />
                    </button>
                </div>
                <span
                    className={`filter-chip px-3 py-1 text-sm font-medium transition-colors select-none text-center
                            ${
                                insertingMode === InsertingMode.TAG
                                    ? "filter-chip-active"
                                    : ""
                            }`}
                >
                    #
                </span>
                <div className="min-w-0">
                    <TagSelector
                        ref={tagSelectorRef}
                        inputStyle="min-h-11 px-3 py-2"
                        placeholder={t("gallery.searchTags")}
                        tags={tags}
                        onTagsUpdate={(t: string[]) => setTags(t)}
                        removeTag={removeTag}
                        onFocus={() => setInsertingMode(InsertingMode.TAG)}
                        onBlur={() => setInsertingMode(InsertingMode.NONE)}
                    />
                </div>
            </div>
        </div>
    );
};
export { type SearchQuery, ImageSearch };
