"use client";

import { useState } from "react";
import { FolderPlus, X } from "lucide-react";
import { UploadAction, mergeUploadActions } from "@/lib/image/uploadIntent";
import BoardSearchList from "@/components/Boards/BoardSearchList";
import BoardPreview from "@/components/Boards/BoardPreview";

export default function UploadBoardActions({ actions, onChange }: {
    actions: UploadAction[];
    onChange: (actions: UploadAction[]) => void;
}) {
    const [open, setOpen] = useState(false);

    return <div className="min-w-0 space-y-2">
        <div className="flex flex-wrap gap-1">
            {actions.map((action) => <span key={`${action.kind}:${action.boardId}`} className="inline-flex items-center gap-1 rounded-full border border-surface-30 px-2 py-1 text-sm">
                {action.boardName}
                <button type="button" aria-label={`Remove ${action.boardName} after-upload action`} onClick={() => onChange(actions.filter((item) => item.boardId !== action.boardId))}><X size={14} /></button>
            </span>)}
        </div>
        <button type="button" onClick={() => setOpen(!open)} aria-expanded={open} className="inline-flex items-center gap-1 rounded-lg border border-surface-30 px-3 py-2 text-sm"><FolderPlus size={16} /> Add to board</button>
        {open && <div className="flex h-52 flex-col rounded-lg border border-surface-30 p-2">
            <BoardSearchList tiles renderBoard={(board) => <button type="button" onClick={() => {
                onChange(mergeUploadActions(actions, [{ kind: "add_to_board", boardId: board.id, boardName: board.name }]));
                setOpen(false);
            }} className="relative flex w-full items-center gap-3 overflow-hidden rounded-xl border border-surface-30 bg-surface-10 p-2 text-left transition hover:border-primary-0 hover:bg-surface-20 lg:aspect-square lg:flex-col lg:items-start lg:justify-end lg:gap-1 lg:p-3">
                <BoardPreview board={board} className="loaded-content relative size-12 shrink-0 grid grid-cols-2 grid-rows-2 gap-px overflow-hidden rounded-lg bg-surface-20 lg:absolute lg:inset-0 lg:h-full lg:w-full lg:rounded-none" />
                <span className="absolute inset-0 hidden bg-gradient-to-t from-black/80 via-black/15 to-transparent lg:block" />
                <span className="relative z-10 line-clamp-2 font-medium lg:w-full lg:text-white">{board.name}</span>
                <span className="relative z-10 ml-auto shrink-0 text-xs text-muted-foreground lg:ml-0 lg:w-full lg:text-white/80">{board.image_count} {board.image_count === 1 ? "image" : "images"}</span>
            </button>} />
        </div>}
    </div>;
}
