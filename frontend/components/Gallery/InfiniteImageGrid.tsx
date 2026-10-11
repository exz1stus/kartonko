"use client";

import { ReactNode, Ref } from "react";
import { ImageMetadata } from "@/lib/api/generated/model";
import GalleryImageGrid from "./GalleryImageGrid";
import DragDropZone from "@/components/UploadImage/DragDropZone";
import usePreUpload from "@/hooks/usePreUpload";
import { UploadIntent } from "@/lib/image/uploadIntent";
import { useTranslations } from "@/contexts/LocaleContext";

interface Props {
    images: ImageMetadata[];
    loading: boolean;
    reachedEnd: boolean;
    error: string | null;
    retry: () => void;
    sentinelRef: Ref<HTMLDivElement>;
    emptyMessage?: string;
    renderAction?: (image: ImageMetadata, index: number) => ReactNode;
    onImageClick?: (image: ImageMetadata, index: number) => void;
    actionClickThrough?: boolean;
    edgeToEdge?: boolean;
    dropIntent?: UploadIntent;
    onImagesDropped?: (imageIds: number[], sourceBoardId?: number, sourceScope?: string) => void;
    dragSourceBoardId?: number;
    selectionScope?: string;
}

export default function InfiniteImageGrid({
    images,
    loading,
    reachedEnd,
    error,
    retry,
    sentinelRef,
    emptyMessage = "No images match this search.",
    renderAction,
    onImageClick,
    actionClickThrough,
    edgeToEdge,
    dropIntent,
    onImagesDropped,
    dragSourceBoardId,
    selectionScope,
}: Props) {
    const { t } = useTranslations();
    const queueFiles = usePreUpload();
    return (
        <DragDropZone
            onFilesDropped={(files) => void queueFiles(files, dropIntent)}
            onImagesDropped={onImagesDropped}
            className="min-h-full"
            label={
                dropIntent?.actions?.length
                    ? t("gallery.uploadAddBoard")
                    : t("gallery.dropFiles")
            }
        >
            <GalleryImageGrid
                images={images}
                renderAction={renderAction}
                onImageClick={onImageClick}
                actionClickThrough={actionClickThrough}
                edgeToEdge={edgeToEdge}
                dragSourceBoardId={dragSourceBoardId}
                selectionScope={selectionScope}
            />
            {!loading && !error && images.length === 0 && (
                <p className="p-6 text-muted-foreground text-center">
                    {emptyMessage}
                </p>
            )}
            {error && (
                <p role="alert" className="p-3 text-red-500 text-center">
                    {error}{" "}
                    <button onClick={retry} className="underline">
                        {t("common.retry")}
                    </button>
                </p>
            )}
            {loading && (
                <p role="status" className="p-3 text-muted-foreground text-center">
                    {t("common.loading")}
                </p>
            )}
            {!reachedEnd && !error && <div ref={sentinelRef} className="h-1" />}
        </DragDropZone>
    );
}
