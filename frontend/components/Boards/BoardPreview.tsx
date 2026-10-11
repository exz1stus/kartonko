"use client";

import { useEffect, useState } from "react";
import Image from "next/image";
import { Images } from "lucide-react";
import { listBoardImages } from "@/lib/api/generated/client";
import { BoardItemResponse, BoardResponse } from "@/lib/api/generated/model";

export default function BoardPreview({ board, className }: {
    board: BoardResponse;
    className?: string;
}) {
    const [previews, setPreviews] = useState<BoardItemResponse[]>([]);
    const [previewError, setPreviewError] = useState(false);

    useEffect(() => {
        let active = true;
        if (board.image_count > 0) {
            listBoardImages(board.id, { limit: 4 })
                .then((images) => {
                    if (active) setPreviews(images);
                })
                .catch(() => {
                    if (active) setPreviewError(true);
                });
        }
        return () => {
            active = false;
        };
    }, [board.id, board.image_count]);

    return (
        <div className={className ?? "relative grid aspect-square grid-cols-2 grid-rows-2 gap-1 overflow-hidden rounded-xl bg-surface-20 sm:h-52 sm:aspect-auto sm:grid-cols-4 sm:grid-rows-1"} aria-label={previewError ? "Preview unavailable" : undefined}>
            {Array.from({ length: 4 }, (_, index) => {
                const image = previews[index]?.image_metadata;
                return (
                    <div key={index} className="relative flex min-h-0 min-w-0 items-center justify-center bg-surface-10">
                        {image ? (
                            <Image
                                src={`/apilocal/image/hash/${encodeURIComponent(image.hash)}/thumb`}
                                alt=""
                                fill
                                sizes="(max-width: 640px) 45vw, 160px"
                                className="object-cover"
                            />
                        ) : (
                            <Images size={24} aria-hidden="true" className="opacity-20" />
                        )}
                    </div>
                );
            })}
        </div>
    );
}
