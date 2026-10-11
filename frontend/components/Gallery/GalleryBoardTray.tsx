"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { Check, ListChecks, Trash2, X } from "lucide-react";
import { useAuth } from "@/contexts/AuthContext";
import { useActiveGalleryScope, useGallerySelection } from "@/hooks/useGallerySelection";
import { useDialog } from "@/contexts/AlertDialogContext";
import { deleteImage, getImage } from "@/lib/api/generated/client";
import ApiError from "@/lib/api/error";
import { notifyImagesDeleted } from "@/lib/image/deletedImages";
import ImageDropTarget from "./ImageDropTarget";
import { addImagesToBoard } from "@/lib/image/addImagesToBoard";
import { toast } from "sonner";
import BoardSearchList from "@/components/Boards/BoardSearchList";
import { useTranslations } from "@/contexts/LocaleContext";

export default function GalleryBoardTray() {
    const { user } = useAuth();
    const scope = useActiveGalleryScope((state) => state.scope);
    const selectedIds = useGallerySelection(scope, (state) => state.selectedIds);
    const draggingIds = useGallerySelection(scope, (state) => state.draggingIds);
    if (!user || (selectedIds.length === 0 && draggingIds.length === 0)) return null;
    return <GalleryBoardTrayContent scope={scope} selectedIds={selectedIds} draggingIds={draggingIds} />;
}

function GalleryBoardTrayContent({ scope, selectedIds, draggingIds }: { scope: string; selectedIds: number[]; draggingIds: number[] }) {
    const { t } = useTranslations();
    const { user } = useAuth();
    const clear = useGallerySelection(scope, (state) => state.clear);
    const selectMany = useGallerySelection(scope, (state) => state.selectMany);
    const imageOwners = useGallerySelection(scope, (state) => state.imageOwners);
    const confirm = useDialog();
    const [busyBoard, setBusyBoard] = useState<number | null>(null);
    const [deleting, setDeleting] = useState(false);
    const [progress, setProgress] = useState({ completed: 0, total: 0 });
    const deletePending = useRef(false);
    const canDelete = !!user && selectedIds.every((id) => user.privilege === "Moderator" || imageOwners[id] === user.id);

    const deleteSelected = useCallback(async () => {
        if (!selectedIds.length || !canDelete || deletePending.current || busyBoard !== null) return;
        deletePending.current = true;
        try {
            const confirmed = await confirm(
                t("gallery.deleteConfirm", { count: selectedIds.length }),
                { title: t("gallery.deleteSelected"), confirmText: t("common.delete") },
            );
            if (!confirmed) return;
            setDeleting(true);
            setProgress({ completed: 0, total: selectedIds.length });
            const toastId = toast.loading(t("gallery.deleting", { done: 0, total: selectedIds.length }));
            const deleted: number[] = [];
            const failed: number[] = [];
            for (const id of selectedIds) {
                try {
                    await deleteImage(id);
                    deleted.push(id);
                    notifyImagesDeleted([id]);
                } catch {
                    try {
                        await getImage(id);
                        failed.push(id);
                    } catch (lookupError) {
                        if (lookupError instanceof ApiError && lookupError.status === 404) {
                            deleted.push(id);
                            notifyImagesDeleted([id]);
                        } else {
                            failed.push(id);
                        }
                    }
                }
                const completed = deleted.length + failed.length;
                setProgress({ completed, total: selectedIds.length });
                toast.loading(t("gallery.deleting", { done: completed, total: selectedIds.length }), { id: toastId });
            }
            if (failed.length) selectMany(failed);
            else clear();
            if (failed.length)
                toast.error(t("gallery.deletedPartial", { done: deleted.length, total: selectedIds.length, failed: failed.length }), { id: toastId });
            else
                toast.success(t("gallery.deleted", { count: deleted.length }), { id: toastId });
        } finally {
            setDeleting(false);
            deletePending.current = false;
        }
    }, [busyBoard, canDelete, clear, confirm, selectMany, selectedIds, t]);

    useEffect(() => {
        const onKeyDown = (event: KeyboardEvent) => {
            const target = event.target;
            if (!selectedIds.length || !canDelete || event.key !== "Delete" || event.repeat || event.ctrlKey || event.metaKey || event.altKey ||
                target instanceof HTMLInputElement || target instanceof HTMLTextAreaElement ||
                (target instanceof HTMLElement && target.isContentEditable)) return;
            event.preventDefault();
            void deleteSelected();
        };
        window.addEventListener("keydown", onKeyDown);
        return () => window.removeEventListener("keydown", onKeyDown);
    }, [canDelete, deleteSelected, selectedIds.length]);

    const addToBoard = async (boardId: number, boardName: string, ids: number[]) => {
        if (busyBoard !== null || deletePending.current) return;
        setBusyBoard(boardId);
        const { added, already, failed } = await addImagesToBoard(boardId, ids);
        if (failed) toast.error(t("boards.addFailed", { count: failed, name: boardName }));
        if (added.length) toast.success(t("boards.added", { count: added.length, name: boardName }));
        if (already) toast.message(t("boards.alreadyAdded", { count: already, name: boardName }));
        if (!failed) clear();
        setBusyBoard(null);
    };

    return <aside aria-label={t("gallery.selectedActions")} className="interface-chrome fixed bottom-4 right-4 z-50 flex h-[min(28rem,70vh)] w-[min(22rem,calc(100vw-2rem))] flex-col rounded-2xl border border-surface-30 bg-surface-0 p-3 shadow-xl">
        <div className="mb-2 flex items-center gap-2"><ListChecks size={18} /><strong className="flex-1">{t("gallery.selected", { count: draggingIds.length || selectedIds.length })}</strong><button disabled={deleting} onClick={clear} aria-label={t("gallery.clearSelection")}><X size={18} /></button></div>
        {selectedIds.length > 0 && <button type="button" disabled={!canDelete || deleting || busyBoard !== null} onClick={() => void deleteSelected()} className="mb-2 flex items-center gap-2 rounded-lg px-3 py-2 text-left text-red-500 hover:bg-red-500/10 disabled:cursor-not-allowed disabled:text-muted-foreground disabled:opacity-50"><Trash2 size={17} />{deleting ? t("gallery.deleting", { done: progress.completed, total: progress.total }) : t("gallery.deleteSelected")}<kbd className="ml-auto text-xs font-normal">Del</kbd></button>}
        {selectedIds.length > 0 && !canDelete && <p className="mb-2 text-xs text-muted-foreground">{t("gallery.youCanOnlyDeleteOwn")}</p>}
        {deleting && <div className="mb-2" role="status"><div className="mb-1 text-xs">{t("gallery.deletingProgress", { done: progress.completed, total: progress.total })}</div><progress className="w-full accent-primary-0" value={progress.completed} max={progress.total} /></div>}
        <BoardSearchList renderBoard={(board) => <ImageDropTarget onDropImages={(ids) => void addToBoard(board.id, board.name, ids)} className="rounded-lg">
                <button disabled={busyBoard !== null || deleting} onClick={() => void addToBoard(board.id, board.name, selectedIds)} className="flex w-full items-center justify-between rounded-lg px-3 py-2 text-left hover:bg-surface-20 disabled:opacity-50"><span className="truncate">{board.name}</span>{busyBoard === board.id && <Check size={16} />}</button>
            </ImageDropTarget>} />
    </aside>;
}
