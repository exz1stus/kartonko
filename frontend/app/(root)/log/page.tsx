"use client";

import AuthGuard from "@/components/AuthGuard";
import Scrollbar from "@/components/template/Scrollbar";
import UserElement from "@/components/UserElement";
import { Button } from "@/components/ui/button";
import { useDialog } from "@/contexts/AlertDialogContext";
import useInfiniteScroll from "@/hooks/useInfiniteScroll";
import { getAuditLogGroupEntries, getAuditLogGroups, revertAuditLogEntry } from "@/lib/api/generated/client";
import type { AuditActor, AuditActionSummary, AuditGroup } from "@/lib/api/generated/model";
import { ChevronDown, Loader, RotateCcw } from "lucide-react";
import Link from "next/link";
import { useCallback, useRef, useState } from "react";
import { toast } from "sonner";
import { useTranslations } from "@/contexts/LocaleContext";

type Entry = {
    id: number;
    created_at: string;
    user_id: number;
    actor?: AuditActor | null;
    action: string;
    object_type: string;
    affected_obj_id: number;
    data?: Record<string, unknown> | null;
    reverted_at?: string | null;
};

function details(entry: Entry, t: ReturnType<typeof useTranslations>["t"]) {
    const data = entry.data || {};
    const name =
        typeof data.name === "string" ? data.name : `#${entry.affected_obj_id}`;
    const before =
        typeof data.before === "string" ? data.before : "previous name";
    const after = typeof data.after === "string" ? data.after : name;
    if (entry.action === "revert") {
        const target =
            entry.object_type === "board" && typeof data.board_name === "string"
                ? `board ${data.board_name}`
                : entry.object_type === "tag" && typeof data.after === "string"
                  ? `tag ${data.after}`
                  : `${entry.object_type} ${name}`;
        return t("log.action.revert", { target });
    }
    if (
        entry.object_type === "board" &&
        (entry.action === "add_image" || entry.action === "remove_image")
    ) {
        const boardName =
            typeof data.board_name === "string"
                ? data.board_name
                : `#${entry.affected_obj_id}`;
        const imageName =
            typeof data.image_name === "string"
                ? data.image_name
                : `#${data.image_id ?? "?"}`;
        return t(entry.action === "add_image" ? "log.action.boardAdd" : "log.action.boardRemove", { image: imageName, board: boardName });
    }
    if (entry.object_type === "image")
        return t(entry.action === "create" ? "log.action.create" : "log.action.delete", { name });
    if (entry.object_type === "tag") {
        if (entry.action === "update")
            return t("log.action.tagRename", { before, after });
        return t(entry.action === "create" ? "log.action.tagCreate" : "log.action.tagDelete", { name });
    }
    return `${entry.action} ${entry.object_type} ${name}`;
}

function subject(entry: Entry) {
    const data = entry.data || {};
    if (entry.object_type === "board" && entry.action !== "revert") {
        return typeof data.image_name === "string"
            ? data.image_name
            : `#${data.image_id ?? "?"}`;
    }
    if (entry.object_type === "tag" && entry.action === "update") {
        return `${typeof data.before === "string" ? data.before : "?"} → ${typeof data.after === "string" ? data.after : "?"}`;
    }
    if (entry.object_type === "board" && typeof data.board_name === "string")
        return data.board_name;
    if (entry.action === "revert" && typeof data.after === "string")
        return data.after;
    return typeof data.name === "string"
        ? data.name
        : `#${entry.affected_obj_id}`;
}

function boardSlug(data?: Record<string, unknown> | null, name?: string) {
    if (typeof data?.board_slug === "string" && data.board_slug) return data.board_slug;
    const boardName = name ?? (typeof data?.board_name === "string" ? data.board_name : "");
    return boardName.toLowerCase().trim().replace(/[^\p{L}\p{N}]+/gu, "-").replace(/^-|-$/g, "") || "board";
}

function BoardLink({ name, slug }: { name: string; slug?: string }) {
    return <Link href={`/board/${encodeURIComponent(slug || boardSlug(null, name))}`} className="font-semibold hover:underline">{name}</Link>;
}

function TagLink({ name }: { name: string }) {
    return <Link href={`/?tag=${encodeURIComponent(name)}`} className="font-semibold hover:underline">{name}</Link>;
}

function Subject({ entry }: { entry: Entry }) {
    const name = subject(entry);
    const revertedCreation = entry.reverted_at && entry.action === "create" && (entry.object_type === "image" || entry.object_type === "tag");
    const removedReference = entry.action === "delete" || entry.action === "remove_image";
    if (entry.action === "revert" && entry.object_type === "image")
        return <s className="font-semibold text-muted-foreground">{name}</s>;
    if (revertedCreation)
        return <s className="font-semibold text-muted-foreground">{name}</s>;
    if (removedReference) return <s className="text-muted-foreground">{name}</s>;
    if (entry.object_type === "image" && entry.action !== "revert")
        return <Link href={`/image/${encodeURIComponent(name)}`} className="font-semibold hover:underline">{name}</Link>;
    if (entry.object_type === "tag") {
        return <TagLink name={name} />;
    }
    if (entry.object_type === "board") {
        const boardName = typeof entry.data?.board_name === "string" ? entry.data.board_name : name;
        return <BoardLink name={boardName} slug={typeof entry.data?.board_slug === "string" ? entry.data.board_slug : undefined} />;
    }
    return <span>{name}</span>;
}

function ActionDescription({ entry }: { entry: Entry }) {
    const { t } = useTranslations();
    const boardName = typeof entry.data?.board_name === "string"
        ? entry.data.board_name
        : `#${entry.affected_obj_id}`;
    if (entry.action === "revert") {
        const parts = t("log.action.revert", { target: "{target}" }).split("{target}");
        return <>{parts[0]}<Subject entry={entry} />{parts[1]}</>;
    }
    if (entry.object_type === "image") {
        const parts = t(entry.action === "create" ? "log.action.create" : "log.action.delete", { name: "{name}" }).split("{name}");
        return <>{parts[0]}<Subject entry={entry} />{parts[1]}</>;
    }
    if (entry.object_type === "tag") {
        if (entry.action === "update") {
            const before = typeof entry.data?.before === "string" ? entry.data.before : "?";
            const after = typeof entry.data?.after === "string" ? entry.data.after : "?";
            const template = t("log.action.tagRename", { before: "{before}", after: "{after}" });
            const [first, remainder] = template.split("{before}");
            const [middle, last] = remainder.split("{after}");
            return <>{first}<TagLink name={before} />{middle}<TagLink name={after} />{last}</>;
        }
        const key = entry.action === "create" ? "log.action.tagCreate" : "log.action.tagDelete";
        const parts = t(key, { name: "{name}" }).split("{name}");
        return <>{parts[0]}<Subject entry={entry} />{parts[1]}</>;
    }
    if (entry.object_type === "board" && (entry.action === "add_image" || entry.action === "remove_image")) {
        const imageName = typeof entry.data?.image_name === "string" ? entry.data.image_name : `#${entry.data?.image_id ?? "?"}`;
        const image = <Link href={`/image/${encodeURIComponent(imageName)}`} className="font-semibold hover:underline">{imageName}</Link>;
        const template = t(entry.action === "add_image" ? "log.action.boardAdd" : "log.action.boardRemove", { image: "{image}", board: "{board}" });
        const [first, remainder] = template.split("{image}");
        const [middle, last] = remainder.split("{board}");
        return <>{first}{image}{middle}<BoardLink name={boardName} slug={typeof entry.data?.board_slug === "string" ? entry.data.board_slug : undefined} />{last}</>;
    }
    return details(entry, t);
}

function canRevert(entry: Entry) {
    if (entry.reverted_at || !entry.data) return false;
    if (entry.object_type === "image")
        return entry.action === "create" && typeof entry.data.name === "string";
    if (entry.object_type === "board")
        return (
            (entry.action === "add_image" || entry.action === "remove_image") &&
            typeof entry.data.image_id === "number"
        );
    if (entry.object_type !== "tag") return false;
    return entry.action === "update"
        ? typeof entry.data.before === "string" &&
              typeof entry.data.after === "string"
        : (entry.action === "create" || entry.action === "delete") &&
              typeof entry.data.name === "string";
}

type Group = AuditGroup;

function summaryLabel(summary: AuditActionSummary, t: ReturnType<typeof useTranslations>["t"]) {
    const count = summary.count;
    if (summary.action === "revert") return t("log.reverted", { count });
    if (summary.object_type === "image") {
        if (summary.action === "create") return t("log.uploaded", { count });
        if (summary.action === "delete") return t("log.deleted", { count });
    }
    if (summary.object_type === "tag") {
        if (summary.action === "create") return t("log.createdTag", { count });
        if (summary.action === "delete") return t("log.deletedTag", { count });
        if (summary.action === "update") return t("log.renamedTag", { count });
    }
    if (summary.object_type === "board") {
        const board = summary.board_name ? ` board ${summary.board_name}` : " a board";
        if (summary.action === "add_image") return `${t("log.addedToBoard", { count })}${board}`;
        if (summary.action === "remove_image") return `${t("log.removedFromBoard", { count })}${board}`;
    }
    return `${count} ${summary.action} ${summary.object_type}`;
}

function SummaryLabel({ summary }: { summary: AuditActionSummary }) {
    const { t } = useTranslations();
    const count = summary.count;
    if (summary.object_type === "board" && (summary.action === "add_image" || summary.action === "remove_image")) {
        const boardName = summary.board_name;
        const slug = (summary as AuditActionSummary & { board_slug?: string }).board_slug;
        return <>{summary.action === "add_image" ? t("log.addedToBoard", { count }) : t("log.removedFromBoard", { count })} {boardName ? <BoardLink name={boardName} slug={slug} /> : "a board"}</>;
    }
    return <>{summaryLabel(summary, t)}</>;
}

export default function Log() {
    const { t, language } = useTranslations();
    const openDialog = useDialog();
    const [reverting, setReverting] = useState<number | null>(null);
    const [groupEntries, setGroupEntries] = useState<Record<number, Entry[]>>({});
    const [loadingGroups, setLoadingGroups] = useState<Record<number, boolean>>({});
    const [groupErrors, setGroupErrors] = useState<Record<number, string>>({});
    const groupEntriesCache = useRef(new Map<number, Entry[]>());
    const groupRequests = useRef(new Map<number, Promise<Entry[]>>());
    const { items, loading, reachedEnd, error, retry, refresh, sentinelRef } =
        useInfiniteScroll<{}, Group>({
            fetchFn: async (_query, cursor, limit) =>
                getAuditLogGroups({ cursor, limit }) as Promise<Group[]>,
            query: {},
        });

    const loadGroupEntries = useCallback(async (groupID: number) => {
        const cached = groupEntriesCache.current.get(groupID);
        if (cached) return cached;
        const pending = groupRequests.current.get(groupID);
        if (pending) return pending;

        setLoadingGroups((current) => ({ ...current, [groupID]: true }));
        setGroupErrors((current) => ({ ...current, [groupID]: "" }));
        const request = getAuditLogGroupEntries(groupID)
            .then((entries) => {
                const loaded = entries as Entry[];
                groupEntriesCache.current.set(groupID, loaded);
                setGroupEntries((current) => ({ ...current, [groupID]: loaded }));
                return loaded;
            })
            .catch((cause) => {
                const message = cause instanceof Error ? cause.message : t("log.groupLoadFailed");
                setGroupErrors((current) => ({ ...current, [groupID]: message }));
                throw cause;
            })
            .finally(() => {
                groupRequests.current.delete(groupID);
                setLoadingGroups((current) => ({ ...current, [groupID]: false }));
            });
        groupRequests.current.set(groupID, request);
        return request;
    }, [t]);

    const refreshGroups = () => {
        groupEntriesCache.current.clear();
        groupRequests.current.clear();
        setGroupEntries({});
        setLoadingGroups({});
        setGroupErrors({});
        refresh();
    };

    const revert = async (entry: Entry) => {
        const confirmed = await openDialog(
            t("log.revertConfirm", { details: details(entry, t) }),
            { title: t("log.title"), confirmText: t("log.undo") },
        );
        if (!confirmed) return;
        setReverting(entry.id);
        try {
            await revertAuditLogEntry(entry.id);
            toast.success(t("log.changeReverted"));
        } catch (cause) {
            toast.error(
                cause instanceof Error
                    ? cause.message
                    : t("log.revertFailed"),
            );
        } finally {
            refreshGroups();
            setReverting(null);
        }
    };

    const revertGroup = async (group: Group) => {
        let loadedEntries: Entry[];
        try {
            loadedEntries = await loadGroupEntries(group.id);
        } catch {
            toast.error(t("log.groupLoadFailed"));
            return;
        }
        const entries = loadedEntries.filter(canRevert);
        if (!entries.length) return;
        const includesUploads = entries.some((entry) => entry.object_type === "image");
        const confirmed = await openDialog(
            t("log.revertGroupConfirm", { count: entries.length, uploadNote: includesUploads ? t("log.uploadNote") : "" }),
            { title: t("log.title"), confirmText: t("log.revertGroup") },
        );
        if (!confirmed) return;
        setReverting(-group.id);
        let succeeded = 0;
        try {
            for (const entry of entries) {
                try {
                    await revertAuditLogEntry(entry.id);
                    succeeded++;
                } catch {
                    // Continue with the other entries when one can no longer be reverted.
                }
            }
            refreshGroups();
            if (succeeded === entries.length)
                toast.success(t("log.reverted", { count: succeeded }));
            else
                toast.error(t("log.revertedPartial", { done: succeeded, total: entries.length }));
        } finally {
            setReverting(null);
        }
    };

    const groups = items;

    return (
        <AuthGuard moderator>
            <Scrollbar>
                <div className="mx-auto px-4 py-6 max-w-3xl">
                    <div className="flex justify-between items-center gap-3 mb-6">
                        <div>
                            <h1 className="font-semibold text-2xl">
                                {t("log.title")}
                            </h1>
                        </div>
                        <Button variant="outline" size="sm" onClick={refreshGroups}>
                            {t("log.refresh")}
                        </Button>
                    </div>
                    {groups.map((group, index) => {
                        const locale = language === "mk" ? "mk-MK" : language === "rue" ? "rue" : "en-GB";
                        const day = new Date(group.last_at).toLocaleDateString(locale, { dateStyle: "full" });
                        const previousDay = index > 0
                            ? new Date(groups[index - 1].last_at).toLocaleDateString(locale, { dateStyle: "full" })
                            : null;
                        return <div key={group.id}>
                            {(index === 0 ||
                                previousDay !== day) && (
                                <h2 className="mt-6 mb-3 pb-2 border-b font-semibold text-muted-foreground text-sm">
                                    {day}
                                </h2>
                            )}
                            <div className="content-card mb-3 p-4 border rounded-lg">
                                <div className="flex flex-wrap items-center justify-between gap-2 mb-2 font-medium text-sm">
                                    <div className="flex items-center gap-2">
                                        <UserElement user={group.actor ?? null} />
                                        <span className="font-normal text-muted-foreground text-xs">
                                            {new Date(group.last_at).toLocaleTimeString(locale, {
                                                hour: "2-digit",
                                                minute: "2-digit",
                                                hour12: false,
                                            })}
                                        </span>
                                    </div>
                                    {group.entry_count > 1 && group.revertable_count > 0 && (
                                        <Button
                                            variant="ghost"
                                            size="xs"
                                            disabled={reverting !== null}
                                            onClick={() => revertGroup(group)}
                                        >
                                            <RotateCcw /> {t("log.revertGroup")} ({group.revertable_count})
                                        </Button>
                                    )}
                                </div>
                                <details className="group" onToggle={(event) => {
                                    if (event.currentTarget.open) void loadGroupEntries(group.id).catch(() => undefined);
                                }}>
                                    <summary className="flex items-center justify-between gap-2 py-1 cursor-pointer list-none rounded-sm focus-visible:outline-2 focus-visible:outline-primary [&::-webkit-details-marker]:hidden">
                                        <span className="font-medium text-sm">
                                            {group.summaries.map((summary, summaryIndex) => (
                                                <span key={`${summary.object_type}:${summary.action}:${summary.board_name ?? ""}`}>
                                                    {summaryIndex > 0 && " · "}<SummaryLabel summary={summary} />
                                                </span>
                                            ))}
                                        </span>
                                        <ChevronDown className="size-4 shrink-0 text-muted-foreground transition-transform group-open:rotate-180" aria-hidden="true" />
                                    </summary>
                                    {loadingGroups[group.id] && <div className="flex justify-center py-3"><Loader className="size-4 animate-spin" /></div>}
                                    {groupErrors[group.id] && <div className="mt-2 text-destructive text-sm">{groupErrors[group.id]} <Button variant="outline" size="xs" onClick={() => void loadGroupEntries(group.id).catch(() => undefined)}>{t("common.retry")}</Button></div>}
                                    <ul className="space-y-2 mt-3 pl-2">
                                    {(groupEntries[group.id] ?? []).map((entry) => (
                                        <li
                                            key={entry.id}
                                            className="flex flex-wrap justify-between items-center gap-2 pl-3 border-surface-50 border-l-2 text-sm"
                                        >
                                            <div>
                                                <span><ActionDescription entry={entry} /></span>
                                                {entry.reverted_at && (
                                                    <span className="ml-2 text-muted-foreground text-xs">
                                                        {t("log.revertedStatus")}
                                                    </span>
                                                )}
                                            </div>
                                            {canRevert(entry) && (
                                                <Button
                                                    variant="ghost"
                                                    size="xs"
                                                    disabled={
                                                        reverting !== null
                                                    }
                                                    onClick={() =>
                                                        revert(entry)
                                                    }
                                                    aria-label={`${t("log.undo")} ${details(entry, t)}`}
                                                >
                                                    <RotateCcw /> {t("log.undo")}
                                                </Button>
                                            )}
                                        </li>
                                    ))}
                                    </ul>
                                </details>
                            </div>
                        </div>;
                    })}
                    {!items.length && !loading && !error && (
                        <p className="text-muted-foreground text-sm">
                            {t("log.empty")}
                        </p>
                    )}
                    {error && (
                        <div className="text-destructive text-sm">
                            {error}{" "}
                            <Button variant="outline" size="sm" onClick={retry}>
                                {t("common.retry")}
                            </Button>
                        </div>
                    )}
                    {loading && (
                        <div className="flex justify-center py-4">
                            <Loader className="animate-spin" />
                        </div>
                    )}
                    {!reachedEnd && <div ref={sentinelRef} />}
                </div>
            </Scrollbar>
        </AuthGuard>
    );
}
