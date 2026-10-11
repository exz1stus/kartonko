"use client";

import AuthGuard from "@/components/AuthGuard";
import Scrollbar from "@/components/template/Scrollbar";
import UserElement from "@/components/UserElement";
import { Button } from "@/components/ui/button";
import { useDialog } from "@/contexts/AlertDialogContext";
import useInfiniteScroll from "@/hooks/useInfiniteScroll";
import { getAuditLogEntries, revertAuditLogEntry } from "@/lib/api/generated/client";
import type { AuditActor } from "@/lib/api/generated/model";
import { ChevronDown, Loader, RotateCcw } from "lucide-react";
import Link from "next/link";
import { useMemo, useState } from "react";
import { toast } from "sonner";

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

function details(entry: Entry) {
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
        return `reverted a change to ${target}`;
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
        return `${entry.action === "add_image" ? "added" : "removed"} image ${imageName} ${entry.action === "add_image" ? "to" : "from"} board ${boardName}`;
    }
    if (entry.object_type === "image")
        return `${entry.action === "create" ? "uploaded" : "deleted"} image ${name}`;
    if (entry.object_type === "tag") {
        if (entry.action === "update")
            return `renamed tag ${before} to ${after}`;
        return `${entry.action === "create" ? "created" : "deleted"} tag ${name}`;
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

function Subject({ entry }: { entry: Entry }) {
    const name = subject(entry);
    if (entry.reverted_at && entry.action === "create" && (entry.object_type === "image" || entry.object_type === "tag"))
        return <s className="font-semibold text-muted-foreground">{name}</s>;
    if (entry.action === "create" && typeof entry.data?.name === "string") {
        if (entry.object_type === "image")
            return (
                <Link href={`/image/${encodeURIComponent(name)}`} className="font-semibold hover:underline">
                    {name}
                </Link>
            );
        if (entry.object_type === "tag")
            return (
                <Link href={`/?tag=${encodeURIComponent(name)}`} className="font-semibold hover:underline">
                    {name}
                </Link>
            );
    }
    return entry.action === "delete" || entry.action === "remove_image"
        ? <s className="text-muted-foreground">{name}</s>
        : <span>{name}</span>;
}

function actionDescription(entry: Entry) {
    const boardName = typeof entry.data?.board_name === "string"
        ? entry.data.board_name
        : `#${entry.affected_obj_id}`;
    if (entry.action === "revert")
        return <>reverted a change to <Subject entry={entry} /></>;
    if (entry.object_type === "image")
        return <>{entry.action === "create" ? "uploaded" : "deleted"} <Subject entry={entry} /></>;
    if (entry.object_type === "tag") {
        const verb = entry.action === "create" ? "created" : entry.action === "delete" ? "deleted" : "renamed";
        return <>{verb} tag <Subject entry={entry} /></>;
    }
    if (entry.object_type === "board" && (entry.action === "add_image" || entry.action === "remove_image"))
        return <>{entry.action === "add_image" ? "added" : "removed"} <Subject entry={entry} /> {entry.action === "add_image" ? "to" : "from"} board {boardName}</>;
    return details(entry);
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

type Group = { day: string; userID: number; entries: Entry[] };

function groupEntries(entries: Entry[]): Group[] {
    const groups: Group[] = [];
    for (const entry of entries) {
        const day = new Date(entry.created_at).toLocaleDateString(undefined, {
            dateStyle: "full",
        });
        const last = groups[groups.length - 1];
        const lastEntry = last?.entries[last.entries.length - 1];
        const closeInTime =
            lastEntry &&
            Math.abs(
                new Date(lastEntry.created_at).getTime() -
                    new Date(entry.created_at).getTime(),
            ) <
                5 * 60_000;
        if (last?.day === day && last.userID === entry.user_id && closeInTime)
            last.entries.push(entry);
        else groups.push({ day, userID: entry.user_id, entries: [entry] });
    }
    return groups;
}

function groupSummary(entries: Entry[]) {
    const counts = new Map<string, Entry[]>();
    for (const entry of entries) {
        const boardName = entry.object_type === "board" ? entry.data?.board_name : "";
        const key = `${entry.object_type}:${entry.action}:${boardName}`;
        const current = counts.get(key);
        if (current) current.push(entry);
        else counts.set(key, [entry]);
    }

    return [...counts.values()]
        .map((group, index) => {
            const entry = group[0];
            const count = group.length;
            const plural = (word: string) =>
                `${count} ${word}${count === 1 ? "" : "s"}`;
            let label = plural("action");
            if (entry.action === "revert")
                label = `Reverted ${plural("change")}`;
            else if (entry.object_type === "image") {
                if (entry.action === "delete")
                    label = `Deleted ${plural("image")}`;
                else if (entry.action === "create")
                    label = `Uploaded ${plural("image")}`;
            }
            else if (entry.object_type === "tag") {
                if (entry.action === "create")
                    label = `Created ${plural("tag")}`;
                else if (entry.action === "delete")
                    label = `Deleted ${plural("tag")}`;
                else if (entry.action === "update")
                    label = `Renamed ${plural("tag")}`;
            }
            else if (entry.object_type === "board") {
                const boardName = entry.data?.board_name;
                const board =
                    typeof boardName === "string"
                        ? ` board ${boardName}`
                        : " a board";
                if (entry.action === "add_image")
                    label = `Added ${plural("image")} to${board}`;
                else if (entry.action === "remove_image")
                    label = `Removed ${plural("image")} from${board}`;
            }
            return <span key={`${entry.object_type}:${entry.action}:${index}`}>
                {index > 0 && " · "}{label}: {group.slice(0, 2).map((sample, sampleIndex) =>
                    <span key={sample.id}>{sampleIndex > 0 && ", "}<Subject entry={sample} /></span>
                )}{count > 2 && ` +${count - 2} more`}
            </span>;
        })
}

export default function Log() {
    const openDialog = useDialog();
    const [reverting, setReverting] = useState<number | null>(null);
    const { items, loading, reachedEnd, error, retry, refresh, sentinelRef } =
        useInfiniteScroll<{}, Entry>({
            fetchFn: async (_query, cursor, limit) =>
                getAuditLogEntries({ cursor, limit }) as Promise<Entry[]>,
            query: {},
        });
    const groups = useMemo(() => groupEntries(items), [items]);

    const revert = async (entry: Entry) => {
        const confirmed = await openDialog(
            `Revert this change? ${details(entry)}`,
            { title: "Revert audit entry", confirmText: "Revert" },
        );
        if (!confirmed) return;
        setReverting(entry.id);
        try {
            await revertAuditLogEntry(entry.id);
            toast.success("Change reverted");
        } catch (cause) {
            toast.error(
                cause instanceof Error
                    ? cause.message
                    : "Could not revert change",
            );
        } finally {
            refresh();
            setReverting(null);
        }
    };

    const revertGroup = async (group: Group) => {
        const entries = group.entries.filter(canRevert);
        if (!entries.length) return;
        const includesUploads = entries.some((entry) => entry.object_type === "image");
        const confirmed = await openDialog(
            `Revert ${entries.length} change${entries.length === 1 ? "" : "s"} in this group?${includesUploads ? " Uploaded images will be deleted." : ""}`,
            { title: "Revert audit group", confirmText: "Revert group" },
        );
        if (!confirmed) return;
        setReverting(-group.entries[0].id);
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
            refresh();
            if (succeeded === entries.length)
                toast.success(`Reverted ${succeeded} change${succeeded === 1 ? "" : "s"}`);
            else
                toast.error(`Reverted ${succeeded} of ${entries.length} changes. Check the refreshed log for entries that failed.`);
        } finally {
            setReverting(null);
        }
    };

    return (
        <AuthGuard moderator>
            <Scrollbar>
                <div className="mx-auto px-4 py-6 max-w-3xl">
                    <div className="flex justify-between items-center gap-3 mb-6">
                        <div>
                            <h1 className="font-semibold text-2xl">
                                Audit log
                            </h1>
                        </div>
                        <Button variant="outline" size="sm" onClick={refresh}>
                            Refresh
                        </Button>
                    </div>
                    {groups.map((group, index) => (
                        <div key={group.entries[0].id}>
                            {(index === 0 ||
                                groups[index - 1].day !== group.day) && (
                                <h2 className="mt-6 mb-3 pb-2 border-b font-semibold text-muted-foreground text-sm">
                                    {group.day}
                                </h2>
                            )}
                            <div className="bg-surface-0 mb-3 p-4 border rounded-lg">
                                <div className="flex flex-wrap items-center justify-between gap-2 mb-2 font-medium text-sm">
                                    <div className="flex items-center gap-2">
                                        <UserElement
                                            user={group.entries[0].actor ?? null}
                                        />
                                        <span className="font-normal text-muted-foreground text-xs">
                                            {new Date(
                                                group.entries[0].created_at,
                                            ).toLocaleTimeString([], {
                                                hour: "2-digit",
                                                minute: "2-digit",
                                                hour12: false,
                                            })}
                                        </span>
                                    </div>
                                    {group.entries.length > 1 && group.entries.some(canRevert) && (
                                        <Button
                                            variant="ghost"
                                            size="xs"
                                            disabled={reverting !== null}
                                            onClick={() => revertGroup(group)}
                                        >
                                            <RotateCcw /> Revert group ({group.entries.filter(canRevert).length})
                                        </Button>
                                    )}
                                </div>
                                <details className="group">
                                    <summary className="flex items-center justify-between gap-2 py-1 cursor-pointer list-none rounded-sm focus-visible:outline-2 focus-visible:outline-primary [&::-webkit-details-marker]:hidden">
                                        <span className="font-medium text-sm">{groupSummary(group.entries)}</span>
                                        <ChevronDown className="size-4 shrink-0 text-muted-foreground transition-transform group-open:rotate-180" aria-hidden="true" />
                                    </summary>
                                    <ul className="space-y-2 mt-3 pl-2">
                                    {group.entries.map((entry) => (
                                        <li
                                            key={entry.id}
                                            className="flex flex-wrap justify-between items-center gap-2 pl-3 border-surface-50 border-l-2 text-sm"
                                        >
                                            <div>
                                                <span>{actionDescription(entry)}</span>
                                                {entry.reverted_at && (
                                                    <span className="ml-2 text-muted-foreground text-xs">
                                                        Reverted
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
                                                    aria-label={`Revert ${details(entry)}`}
                                                >
                                                    <RotateCcw /> Revert
                                                </Button>
                                            )}
                                        </li>
                                    ))}
                                    </ul>
                                </details>
                            </div>
                        </div>
                    ))}
                    {!items.length && !loading && !error && (
                        <p className="text-muted-foreground text-sm">
                            No audit entries yet.
                        </p>
                    )}
                    {error && (
                        <div className="text-destructive text-sm">
                            {error}{" "}
                            <Button variant="outline" size="sm" onClick={retry}>
                                Retry
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
