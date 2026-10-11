"use client";

import { FormEvent, useCallback, useState } from "react";
import Link from "next/link";
import { Hash, Pencil, Plus, Search, Trash2 } from "lucide-react";
import { useDebounce } from "use-debounce";
import { deleteTag, getTags, patchTag, postTag } from "@/lib/api/generated/client";
import { useAuth } from "@/contexts/AuthContext";
import { useDialog } from "@/contexts/AlertDialogContext";
import usePagedResults from "@/hooks/usePagedResults";
import NameField from "@/components/NameField";
import { sanitizeName } from "@/lib/sanitizeName";
import { useTranslations } from "@/contexts/LocaleContext";

export default function TagsPage() {
    const { user } = useAuth();
    const { t } = useTranslations();
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

    async function createTag(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();
        const form = event.currentTarget;
        const name = sanitizeName(String(new FormData(form).get("name") ?? "").trim());
        if (!name) {
            setActionError(t("tags.invalid"));
            return;
        }
        setBusy(true);
        setActionError(null);
        setNotice(null);
        try {
            await postTag({ name });
            setNotice(t("tags.created"));
            form.reset();
            tags.refresh();
        } catch (cause) {
            setActionError(cause instanceof Error ? cause.message : t("error.couldNotCreateTags"));
        } finally {
            setBusy(false);
        }
    }

    async function renameTag(event: FormEvent<HTMLFormElement>, id: number) {
        event.preventDefault();
        const name = sanitizeName(String(new FormData(event.currentTarget).get("name") ?? "").trim());
        if (!name) {
            setActionError(t("tags.invalid"));
            return;
        }
        setBusy(true);
        setActionError(null);
        try {
            await patchTag(id, { name });
            setEditing(null);
            tags.refresh();
        } catch (cause) {
            setActionError(cause instanceof Error ? cause.message : t("error.couldNotRenameTag"));
        } finally {
            setBusy(false);
        }
    }

    async function removeTag(id: number, name: string) {
        if (!await confirm(`${t("common.delete")} “${name}”?`, { title: t("common.delete"), confirmText: t("common.delete") })) return;
        setBusy(true);
        setActionError(null);
        try {
            await deleteTag(id);
            tags.refresh();
        } catch (cause) {
            setActionError(cause instanceof Error ? cause.message : t("error.couldNotDeleteTag"));
        } finally {
            setBusy(false);
        }
    }

    return <div className="h-full overflow-y-auto"><title>Tags | kartonko</title><div className="mx-auto max-w-4xl space-y-6 px-4 py-6 sm:px-8">
        <header className="flex flex-wrap items-start justify-between gap-4"><div><h1 className="text-3xl font-semibold">{t("nav.tags")}</h1><p className="mt-1 text-sm text-muted-foreground">{t("tags.description")}</p></div>{user && <button onClick={() => setCreating(!creating)} className="inline-flex items-center gap-2 rounded-xl bg-primary-0 px-4 py-2 text-primary-foreground hover:opacity-90"><Plus size={18} /> {t("tags.new")}</button>}</header>
        {creating && <form onSubmit={createTag} className="glass grid gap-3 rounded-2xl border border-surface-20 p-4"><label className="grid gap-1 text-sm">{t("tags.name")}<NameField name="name" required maxLength={120} placeholder={t("tags.enterName")} className="rounded-lg border border-surface-30 bg-surface-10 px-3 py-2" /></label><div className="flex gap-2"><button disabled={busy} className="rounded-lg bg-primary-0 px-4 py-2 text-primary-foreground disabled:opacity-50">{t("tags.create")}</button><button type="button" onClick={() => setCreating(false)} className="rounded-lg border border-surface-30 px-4 py-2">{t("common.cancel")}</button></div></form>}
        <label className="control-shell flex items-center gap-2 px-3 rounded-xl"><Search size={18} aria-hidden="true" /><input type="search" value={search} onChange={(event) => setSearch(event.target.value)} placeholder={t("tags.search")} aria-label={t("tags.search")} className="w-full py-3 outline-none" /></label>
        {notice && <p role="status" className="text-sm">{notice}</p>}
        {actionError && <p role="alert" className="text-red-500">{actionError}</p>}
        <div className="content-card divide-y divide-surface-20 overflow-hidden rounded-2xl border border-surface-20">
            {tags.items.map((tag) => {
                const canEdit = !!user && (user.id === tag.user_id || user.privilege === "Moderator");
                const empty = tag.image_count === 0;
                return <div key={tag.id} className={`flex min-w-0 items-center gap-2 p-3 sm:px-5 ${empty ? "bg-surface-10/50" : ""}`}>
                    <Hash size={18} className="shrink-0 text-muted-foreground" aria-hidden="true" />
                    {editing === tag.id ? <form onSubmit={(event) => renameTag(event, tag.id)} className="flex min-w-0 flex-1 flex-wrap gap-2"><NameField name="name" defaultValue={tag.name} required autoFocus className="min-w-0 flex-1 rounded-lg border border-surface-30 bg-surface-10 px-3 py-1" /><button disabled={busy} className="rounded-lg bg-primary-0 px-3 py-1 text-primary-foreground disabled:opacity-50">{t("common.save")}</button><button type="button" onClick={() => setEditing(null)} className="rounded-lg border border-surface-30 px-3 py-1">{t("common.cancel")}</button></form> : <><Link href={`/?tag=${encodeURIComponent(tag.name)}`} className="min-w-0 flex-1 truncate py-1 hover:underline">{tag.name}</Link><span className="shrink-0 tabular-nums text-sm">{t("tags.count", { count: tag.image_count })}</span>{canEdit && <div className="flex shrink-0 gap-1"><button onClick={() => setEditing(tag.id)} aria-label={t("tags.rename", { name: tag.name })} className="rounded-lg p-2 hover:bg-surface-20"><Pencil size={17} /></button><button disabled={busy} onClick={() => removeTag(tag.id, tag.name)} aria-label={`${t("common.delete")} ${tag.name}`} className="rounded-lg p-2 text-red-500 hover:bg-red-500/10 disabled:opacity-50"><Trash2 size={17} /></button></div>}</>}
                </div>;
            })}
        </div>
        {!tags.loading && !tags.error && tags.items.length === 0 && <p className="text-center text-muted-foreground">{prefix ? t("tags.noResults") : t("tags.empty")}</p>}
        {tags.error && <p role="alert" className="text-red-500">{tags.error} <button onClick={tags.retry} className="underline">{t("common.retry")}</button></p>}
        {tags.loading && <p role="status" className="text-center text-muted-foreground">{t("tags.loading")}</p>}
        {tags.hasMore && !tags.loading && <div className="text-center"><button onClick={tags.loadMore} className="rounded-xl border border-surface-30 px-5 py-2 hover:bg-surface-20">{t("tags.loadMore")}</button></div>}
    </div></div>;
}
