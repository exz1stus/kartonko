"use client";

import { FormEvent, useCallback, useState } from "react";
import Link from "next/link";
import { Hash, Pencil, Plus, Search, Trash2 } from "lucide-react";
import { useDebounce } from "use-debounce";
import { deleteTag, getTags, patchTag, postTag, postTagsBatch } from "@/lib/api/generated/client";
import { useAuth } from "@/contexts/AuthContext";
import { useDialog } from "@/contexts/AlertDialogContext";
import usePagedResults from "@/hooks/usePagedResults";

export default function TagsPage() {
    const { user } = useAuth();
    const confirm = useDialog();
    const [search, setSearch] = useState("");
    const [prefix] = useDebounce(search.trim(), 250);
    const [creating, setCreating] = useState(false);
    const [editing, setEditing] = useState<number | null>(null);
    const [busy, setBusy] = useState(false);
    const [actionError, setActionError] = useState<string | null>(null);
    const [notice, setNotice] = useState<string | null>(null);
    const fetchPage = useCallback((query: string, cursor: number, limit: number) => getTags({ prefix: query, cursor, limit }), []);
    const tags = usePagedResults(prefix, fetchPage, 40);

    async function createTags(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();
        const form = event.currentTarget;
        const names = String(new FormData(form).get("names") ?? "").split(/[\n,]+/).map((name) => name.trim()).filter(Boolean);
        if (names.length === 0) return;
        setBusy(true);
        setActionError(null);
        setNotice(null);
        try {
            if (names.length === 1) {
                await postTag({ name: names[0] });
                setNotice("Tag created.");
            } else {
                const result = await postTagsBatch({ names });
                setNotice(`${result.successes.length} tags created${result.failures?.length ? `; ${result.failures.length} failed` : ""}.`);
                if (result.failures?.length) setActionError(result.failures.map((failure) => `${failure.name}: ${failure.error}`).join("; "));
            }
            form.reset();
            tags.refresh();
        } catch (cause) {
            setActionError(cause instanceof Error ? cause.message : "Could not create tags");
        } finally {
            setBusy(false);
        }
    }

    async function renameTag(event: FormEvent<HTMLFormElement>, id: number) {
        event.preventDefault();
        const name = String(new FormData(event.currentTarget).get("name") ?? "").trim();
        if (!name) return;
        setBusy(true);
        setActionError(null);
        try {
            await patchTag(id, { name });
            setEditing(null);
            tags.refresh();
        } catch (cause) {
            setActionError(cause instanceof Error ? cause.message : "Could not rename tag");
        } finally {
            setBusy(false);
        }
    }

    async function removeTag(id: number, name: string) {
        if (!await confirm(`Delete tag “${name}”?`, { title: "Delete tag", confirmText: "Delete" })) return;
        setBusy(true);
        setActionError(null);
        try {
            await deleteTag(id);
            tags.refresh();
        } catch (cause) {
            setActionError(cause instanceof Error ? cause.message : "Could not delete tag");
        } finally {
            setBusy(false);
        }
    }

    return <div className="h-full overflow-y-auto"><title>Tags | kartonko</title><div className="mx-auto max-w-4xl space-y-6 px-4 py-6 sm:px-8">
        <header className="flex flex-wrap items-start justify-between gap-4"><div><h1 className="text-3xl font-semibold">Tags</h1><p className="mt-1 text-sm text-primary-0/70">Browse images by tag or manage your tags. Sorted by image count.</p></div>{user && <button onClick={() => setCreating(!creating)} className="inline-flex items-center gap-2 rounded-xl bg-primary-0 px-4 py-2 text-surface-0 hover:opacity-90"><Plus size={18} /> New tags</button>}</header>
        {creating && <form onSubmit={createTags} className="grid gap-3 rounded-2xl border border-surface-20 bg-surface-0/70 p-4"><label className="grid gap-1 text-sm">Tag names<textarea name="names" required rows={3} placeholder="Enter one tag, or separate several with commas or new lines" className="rounded-lg border border-surface-30 bg-surface-10 px-3 py-2" /></label><div className="flex gap-2"><button disabled={busy} className="rounded-lg bg-primary-0 px-4 py-2 text-surface-0 disabled:opacity-50">Create</button><button type="button" onClick={() => setCreating(false)} className="rounded-lg border border-surface-30 px-4 py-2">Cancel</button></div></form>}
        <label className="flex items-center gap-2 rounded-xl border border-surface-30 bg-surface-0/70 px-3 focus-within:border-primary-0"><Search size={18} aria-hidden="true" /><input type="search" value={search} onChange={(event) => setSearch(event.target.value)} placeholder="Search tags by prefix" aria-label="Search tags by prefix" className="w-full bg-transparent py-3 outline-none" /></label>
        {notice && <p role="status" className="text-sm">{notice}</p>}
        {actionError && <p role="alert" className="text-red-500">{actionError}</p>}
        <div className="divide-y divide-surface-20 overflow-hidden rounded-2xl border border-surface-20 bg-surface-0/70">
            {tags.items.map((tag) => {
                const canEdit = !!user && (user.id === tag.user_id || user.privilege === "Moderator");
                const empty = tag.image_count === 0;
                return <div key={tag.id} className={`flex min-w-0 items-center gap-2 p-3 sm:px-5 ${empty ? "bg-surface-10/50 text-primary-0/50" : ""}`}>
                    <Hash size={18} className={`shrink-0 ${empty ? "opacity-40" : "text-primary-0/60"}`} aria-hidden="true" />
                    {editing === tag.id ? <form onSubmit={(event) => renameTag(event, tag.id)} className="flex min-w-0 flex-1 flex-wrap gap-2"><input name="name" defaultValue={tag.name} required autoFocus className="min-w-0 flex-1 rounded-lg border border-surface-30 bg-surface-10 px-3 py-1" /><button disabled={busy} className="rounded-lg bg-primary-0 px-3 py-1 text-surface-0 disabled:opacity-50">Save</button><button type="button" onClick={() => setEditing(null)} className="rounded-lg border border-surface-30 px-3 py-1">Cancel</button></form> : <><Link href={`/?tag=${encodeURIComponent(tag.name)}`} className="min-w-0 flex-1 truncate py-1 hover:underline">{tag.name}</Link><span className="shrink-0 tabular-nums text-sm">{tag.image_count} {tag.image_count === 1 ? "image" : "images"}</span>{canEdit && <div className="flex shrink-0 gap-1"><button onClick={() => setEditing(tag.id)} aria-label={`Rename ${tag.name}`} className="rounded-lg p-2 hover:bg-surface-20"><Pencil size={17} /></button><button disabled={busy} onClick={() => removeTag(tag.id, tag.name)} aria-label={`Delete ${tag.name}`} className="rounded-lg p-2 text-red-500 hover:bg-red-500/10 disabled:opacity-50"><Trash2 size={17} /></button></div>}</>}
                </div>;
            })}
        </div>
        {!tags.loading && !tags.error && tags.items.length === 0 && <p className="text-center text-primary-0/70">{prefix ? "No tags match this prefix." : "No tags yet."}</p>}
        {tags.error && <p role="alert" className="text-red-500">{tags.error} <button onClick={tags.retry} className="underline">Retry</button></p>}
        {tags.loading && <p role="status" className="text-center text-primary-0/70">Loading tags…</p>}
        {tags.hasMore && !tags.loading && <div className="text-center"><button onClick={tags.loadMore} className="rounded-xl border border-surface-30 px-5 py-2 hover:bg-surface-20">Load more</button></div>}
    </div></div>;
}
