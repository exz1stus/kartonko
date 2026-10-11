"use client";

import { useState } from "react";
import { Check, FolderPlus, X } from "lucide-react";
import { useAuth } from "@/contexts/AuthContext";
import { useActiveGalleryScope, useGallerySelection } from "@/hooks/useGallerySelection";
import ImageDropTarget from "./ImageDropTarget";
import { addImagesToBoard } from "@/lib/image/addImagesToBoard";
import { toast } from "sonner";
import BoardSearchList from "@/components/Boards/BoardSearchList";

export default function GalleryBoardTray() {
    const { user } = useAuth();
    const scope = useActiveGalleryScope((state) => state.scope);
    const selectedIds = useGallerySelection(scope, (state) => state.selectedIds);
    const draggingIds = useGallerySelection(scope, (state) => state.draggingIds);
    if (!user || (selectedIds.length === 0 && draggingIds.length === 0)) return null;
    return <GalleryBoardTrayContent scope={scope} selectedIds={selectedIds} draggingIds={draggingIds} />;
}

function GalleryBoardTrayContent({ scope, selectedIds, draggingIds }: { scope: string; selectedIds: number[]; draggingIds: number[] }) {
    const clear = useGallerySelection(scope, (state) => state.clear);
    const [busyBoard, setBusyBoard] = useState<number | null>(null);

    const addToBoard = async (boardId: number, boardName: string, ids: number[]) => {
        if (busyBoard !== null) return;
        setBusyBoard(boardId);
        const { added, already, failed } = await addImagesToBoard(boardId, ids);
        if (failed) toast.error(`Could not add ${failed} images to ${boardName}`);
        if (added.length) toast.success(`Added ${added.length} images to ${boardName}`);
        if (already) toast.message(`${already} images were already on ${boardName}`);
        if (!failed) clear();
        setBusyBoard(null);
    };

    return <aside aria-label="Add selected images to a board" className="fixed bottom-4 right-4 z-50 flex h-[min(28rem,70vh)] w-[min(22rem,calc(100vw-2rem))] flex-col rounded-2xl border border-surface-30 bg-surface-0 p-3 shadow-xl">
        <div className="mb-2 flex items-center gap-2"><FolderPlus size={18} /><strong className="flex-1">{(draggingIds.length || selectedIds.length)} selected</strong><button onClick={clear} aria-label="Clear image selection"><X size={18} /></button></div>
        <BoardSearchList renderBoard={(board) => <ImageDropTarget onDropImages={(ids) => void addToBoard(board.id, board.name, ids)} className="rounded-lg">
                <button disabled={busyBoard !== null} onClick={() => void addToBoard(board.id, board.name, selectedIds)} className="flex w-full items-center justify-between rounded-lg px-3 py-2 text-left hover:bg-surface-20 disabled:opacity-50"><span className="truncate">{board.name}</span>{busyBoard === board.id && <Check size={16} />}</button>
            </ImageDropTarget>} />
    </aside>;
}
