"use client";
import AuthGuard from "@/components/AuthGuard";
import DragDropZone from "@/components/UploadImage/DragDropZone";
import UploadImageForm from "@/components/UploadImage/UploadImageForm";
import FancySpan from "@/components/template/FancySpan";
import useUploadStore, {
    selectGlobalNewTags,
    UploadItem,
} from "@/hooks/useUploadStore";
import { ArrowLeft, ArrowRight, UploadIcon } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";
import Image from "next/image";
import usePreUpload from "@/hooks/usePreUpload";
import { toast } from "sonner";
import CaptchaButton from "@/components/UploadImage/CaptchaButton";
import { useShallow } from "zustand/react/shallow";
import ImageCarousel from "@/components/UploadImage/ImageCarousel";
import { cn } from "@/lib/utils";
import { ImageBatchUploadRequest, createUploadSession, uploadImageBatch } from "@/lib/image/upload";
import ProgressBarToast from "@/components/ProgressBarToast";
import { Progress } from "@/components/ui/progress";
import { runUploadAction } from "@/lib/image/runUploadAction";
import { uploadActionKey } from "@/lib/image/uploadIntent";
import Loading from "@/components/Loading";
import UploadBoardActions from "@/components/UploadImage/UploadBoardActions";
import { useAuth } from "@/contexts/AuthContext";
import { useTranslations } from "@/contexts/LocaleContext";

const UploadPage = () => {
    const { user } = useAuth();
    const { t } = useTranslations();
    const {
        removeItems,
        updateMetadata,
        items: storeImages,
        commonActions,
        hydrated: queueHydrated,
    } = useUploadStore((state) => state);

    const objectUrlCache = useRef<Map<File, string>>(new Map());

    const [imageIndex, setImageIndex] = useState(0);
    const [imageUrl, setImageUrl] = useState("");
    const selectedFile = storeImages[imageIndex]?.file;

    const prevLengthRef = useRef(storeImages.length);

    const [loading, setLoading] = useState(false);
    const [uploadProgress, setUploadProgress] = useState<{ value: number; label: string } | null>(null);
    const [selectedIndices, setSelectedIndices] = useState<number[]>([]);

    const [captchaToken, setCaptchaToken] = useState<string | null>();
    const [captchaResetKey, setCaptchaResetKey] = useState(0);

    const globalNewTags = useUploadStore(useShallow(selectGlobalNewTags));

    const handleWheel = (e: React.WheelEvent<HTMLDivElement>) => {
        setImageIndex((prev) => {
            const next = prev + Math.sign(e.deltaY);
            return Math.min(Math.max(next, 0), storeImages.length - 1);
        });
    };

    const selectNextImage = () => {
        setImageIndex((prev) => {
            const next = prev + 1;
            return Math.min(Math.max(next, 0), storeImages.length - 1);
        });
    };

    const selectPreviousImage = () => {
        setImageIndex((prev) => {
            const next = prev - 1;
            return Math.min(Math.max(next, 0), storeImages.length - 1);
        });
    };

    const handleDroppedFiles = usePreUpload();

    const removeCurrentImage = useCallback(() => {
        const item = storeImages[imageIndex];
        if (!item) return;
        setImageIndex(Math.max(0, Math.min(imageIndex, storeImages.length - 2)));
        setSelectedIndices([]);
        removeItems([item.id]);
    }, [imageIndex, removeItems, storeImages]);

    const removeSelectedImages = useCallback(() => {
        const ids = selectedIndices
            .map((index) => storeImages[index]?.id)
            .filter((id): id is string => Boolean(id));
        if (ids.length === 0) return;

        const removed = new Set(ids);
        const remaining = storeImages.filter((item) => !removed.has(item.id));
        const currentId = storeImages[imageIndex]?.id;
        const retainedIndex = remaining.findIndex((item) => item.id === currentId);
        setImageIndex(retainedIndex >= 0 ? retainedIndex : Math.max(0, Math.min(imageIndex, remaining.length - 1)));
        setSelectedIndices([]);
        removeItems(ids);
    }, [imageIndex, removeItems, selectedIndices, storeImages]);

    useEffect(() => {
        const urls = objectUrlCache.current;
        return () => {
            urls.forEach((url) => URL.revokeObjectURL(url));
            urls.clear();
        };
    }, []);

    useEffect(() => {
        const prevLength = prevLengthRef.current;
        const nextLength = storeImages.length;
        if (nextLength === prevLength) return;
        prevLengthRef.current = nextLength;
        setImageIndex((prev) =>
            nextLength === 0 ? 0 : nextLength > prevLength ? nextLength - 1 : Math.min(prev, nextLength - 1),
        );
    }, [storeImages.length]);

    useEffect(() => {
        if (!selectedFile) {
            setImageUrl("");
            return;
        }
        let url = objectUrlCache.current.get(selectedFile);
        if (!url) {
            url = URL.createObjectURL(selectedFile);
            objectUrlCache.current.set(selectedFile, url);
        }
        setImageUrl(url);
    }, [selectedFile]);

    useEffect(() => {
        const onKeyDown = (e: KeyboardEvent) => {
            if (
                e.target instanceof HTMLInputElement ||
                e.target instanceof HTMLTextAreaElement
            ) {
                return;
            }
            if (e.key === "Delete" && !e.repeat && selectedIndices.length > 0) {
                e.preventDefault();
                removeSelectedImages();
            }
        };

        document.addEventListener("keydown", onKeyDown);
        return () => {
            document.removeEventListener("keydown", onKeyDown);
        };
    }, [selectedIndices.length, removeSelectedImages]);

    const browseImages = (
        <label className="inline-flex items-center gap-2 px-2 py-1 border border-surface-20 hover:border-surface-30 rounded-3xl cursor-pointer glass">
            <UploadIcon />
            Browse files
            <input
                type="file"
                className="hidden"
                multiple
                accept="image/*"
                onChange={(e) =>
                    handleDroppedFiles(Array.from(e.target.files || []))
                }
            />
        </label>
    );

    const onFormChange = useCallback(
        (updates: Partial<Omit<UploadItem, "file">>) => {
            updateMetadata(imageIndex, updates);
        },
        [updateMetadata, imageIndex],
    );

    const onUpload = async () => {
        if (loading || storeImages.length === 0) return;
        setLoading(true);
        const toastId = toast.loading(t("upload.prepare"));
        try {
            const allPending = useUploadStore
                .getState()
                .items.filter((item) => !item.uploadedImageId);
            const pending = user?.privilege === "Moderator" ? allPending : allPending.slice(0, 50);
            for (const item of allPending.slice(pending.length))
                useUploadStore.getState().setItemError(item.id, t("upload.standardLimit"));
            if (pending.length > 0) {
                const session = await createUploadSession(captchaToken ?? "");
                const totalBytes = pending.reduce((sum, item) => sum + item.file.size, 0);
                let sentBytes = 0;
                let processed = 0;
                let lastPercent = -1;
                let lastLabel = "";
                const showProgress = (value: number, label: string) => {
                    const percent = Math.min(100, Math.round(value));
                    if (percent !== lastPercent || label !== lastLabel) {
                        setUploadProgress({ value: percent, label });
                        toast.loading(label, {
                            id: toastId,
                            description: <ProgressBarToast value={percent} currentName={t("upload.processed", { done: processed, total: pending.length })} />,
                        });
                        lastPercent = percent;
                        lastLabel = label;
                    }
                };
                const maxChunkBytes = 32 * 1024 * 1024;
                const maxChunkFiles = 12;
                let quotaReached = false;
                for (let start = 0; start < pending.length && !quotaReached;) {
                    const chunk: UploadItem[] = [];
                    let chunkBytes = 0;
                    while (start < pending.length && chunk.length < maxChunkFiles &&
                        (chunk.length === 0 || chunkBytes + pending[start].file.size <= maxChunkBytes)) {
                        const item = pending[start++];
                        chunk.push(item);
                        chunkBytes += item.file.size;
                    }
                    const request: ImageBatchUploadRequest = {
                        data: chunk.map((item) => ({
                            client_id: item.id,
                            name: item.name,
                            tags: item.tags,
                            newTags: item.newTags,
                        })),
                        common_tags: [],
                    };
                    showProgress(totalBytes ? sentBytes / totalBytes * 100 : 0, t("upload.uploading", { done: processed, total: pending.length }));
                    try {
                        const response = await uploadImageBatch(
                            request,
                            chunk.map((item) => item.file),
                            session,
                            (loaded, total) => showProgress(
                                totalBytes ? (sentBytes + chunkBytes * loaded / total) / totalBytes * 100 : 100,
                                loaded === total ? t("upload.processing", { count: chunk.length }) : t("upload.uploading", { done: processed, total: pending.length }),
                            ),
                        );
                        if (!response.results || response.results.length !== chunk.length)
                            throw new Error("Upload server did not return per-image results");
                        response.results.forEach((result) => {
                            if (result.image)
                                useUploadStore.getState().markUploaded(result.client_id, result.image.id);
                            else
                                useUploadStore.getState().setItemError(result.client_id, result.error || t("upload.failed"));
                        });
                        quotaReached = response.results.some((result) => result.error?.includes("upload limit reached"));
                    } catch (error) {
                        const message = error instanceof Error ? error.message : t("upload.failed");
                        chunk.forEach((item) => useUploadStore.getState().setItemError(item.id, message));
                    }
                    sentBytes += chunkBytes;
                    processed += chunk.length;
                    showProgress(totalBytes ? sentBytes / totalBytes * 100 : 100, t("upload.processed", { done: processed, total: pending.length }));
                    if (quotaReached)
                        for (const item of pending.slice(start))
                            useUploadStore.getState().setItemError(item.id, t("upload.limitReached"));
                }
            }

            setUploadProgress({ value: 100, label: t("upload.finishing") });
            toast.loading(t("upload.finishing"), { id: toastId });

            const completedIds: string[] = [];
            for (const item of useUploadStore.getState().items) {
                if (!item.uploadedImageId) continue;
                for (const action of item.actions) {
                    const key = uploadActionKey(action);
                    if (
                        useUploadStore
                            .getState()
                            .items.find((current) => current.id === item.id)
                            ?.completedActions.includes(key)
                    )
                        continue;
                    try {
                        await runUploadAction(action, item.uploadedImageId);
                        useUploadStore
                            .getState()
                            .markActionComplete(item.id, key);
                    } catch (error) {
                        useUploadStore
                            .getState()
                            .setItemError(
                                item.id,
                                error instanceof Error
                                    ? error.message
                                    : "Board action failed",
                            );
                    }
                }
                const current = useUploadStore
                    .getState()
                    .items.find((entry) => entry.id === item.id);
                if (
                    current &&
                    current.actions.every((action) =>
                        current.completedActions.includes(
                            uploadActionKey(action),
                        ),
                    )
                ) {
                    completedIds.push(item.id);
                }
            }
            if (completedIds.length) {
                useUploadStore.getState().removeItems(completedIds);
                setSelectedIndices([]);
            }
            const remaining = useUploadStore.getState().items.length;
            if (remaining)
                toast.warning(t("upload.attention", { count: remaining }), {
                    id: toastId,
                });
            else
                toast.success(pending.length > 0 ? t("upload.success") : t("upload.boardActionsComplete"), {
                    id: toastId,
                });
        } catch (error) {
            toast.error(
                error instanceof Error ? error.message : t("upload.failed"),
                { id: toastId },
            );
        } finally {
            setLoading(false);
            setUploadProgress(null);
            setCaptchaToken(null);
            setCaptchaResetKey((x) => x + 1);
        }
    };

    const nextButton = (
        <ArrowRight
            onClick={selectNextImage}
            className="p-2 border border-surface-20 hover:border-surface-30 rounded-2xl min-w-10 min-h-10 transition cursor-pointer glass"
        />
    );
    const backButton = (
        <ArrowLeft
            onClick={selectPreviousImage}
            className="p-2 border border-surface-20 hover:border-surface-30 rounded-2xl min-w-10 min-h-10 transition cursor-pointer glass"
        />
    );

    const uploadBtnText = [
        storeImages.length > 1 ? t("upload.imagesCount", { count: storeImages.length }) : t("upload.image"),
        globalNewTags?.length ? t("upload.newTagsCount", { count: globalNewTags.length }) : "",
    ].filter(Boolean).join("");

    const formButtons =
        imageIndex === storeImages.length - 1 ? (
            <>
                {storeImages.length > 1 && backButton}
                {storeImages.some((item) => !item.uploadedImageId) ? (
                    <CaptchaButton
                        className={`max-w-[70%]
                        ${loading ? "animate-pulse bg-surface-20" : "bg-none"}
                    `}
                        onClick={onUpload}
                        disabled={loading}
                        onVerifySuccess={setCaptchaToken}
                        resetKey={captchaResetKey}
                    >
                        {loading ? t("upload.uploadingButton") : t("upload.button", { details: uploadBtnText })}
                    </CaptchaButton>
                ) : (
                    <button
                        onClick={onUpload}
                        disabled={loading}
                        className="disabled:opacity-50 px-4 py-2 border border-surface-30 rounded-xl"
                    >
                        Retry board actions
                    </button>
                )}
            </>
        ) : (
            <>
                {imageIndex > 0 && backButton}
                {nextButton}
            </>
        );

    const getFileUrl = useCallback((file: File) => {
        if (objectUrlCache.current.has(file)) {
            return objectUrlCache.current.get(file)!;
        }
        const url = URL.createObjectURL(file);
        objectUrlCache.current.set(file, url);
        return url;
    }, []);

    const content =
        storeImages.length > 0 ? (
            <div className="flex justify-center items-start p-4 lg:p-10 w-full h-full overflow-hidden">
                <div className="flex flex-col bg-surface-0/95 border border-surface-20 rounded-2xl w-full h-full max-h-full overflow-hidden glass">
                    <div className="flex lg:flex-row flex-col flex-1 min-h-0 overflow-hidden">
                        {imageUrl ? (
                            <>
                                <div
                                    className={cn(
                                        "relative flex flex-1 justify-center items-center bg-black/20 w-full min-w-0 h-[20vh] min-h-0 overflow-hidden shrink-0",
                                        "lg:w-[60%] lg:h-full",
                                    )}
                                >
                                    <Image
                                        src={imageUrl}
                                        alt={storeImages[imageIndex]?.name}
                                        className="loaded-content rounded-t-2xl lg:rounded-l-2xl lg:rounded-tr-none w-full h-full object-contain"
                                        draggable={false}
                                        width={1000}
                                        height={1000}
                                        onWheel={handleWheel}
                                        unoptimized
                                    />
                                </div>

                                <div className="flex flex-col gap-6 p-5 lg:p-8 border-l w-full lg:w-[40%] min-h-0 overflow-y-auto">
                                    <div className="flex flex-wrap justify-between items-center gap-2">
                                        <div className="inline-flex items-center gap-2">
                                            {browseImages}
                                            {storeImages.length > 1 && (
                                                <span className="text-gray-500 text-sm shrink-0">
                                                    {imageIndex + 1}/
                                                    {storeImages.length}
                                                </span>
                                            )}
                                        </div>

                                        <div className="flex items-center gap-2 min-w-0 max-w-[50%]">
                                            <FancySpan
                                                word={
                                                    storeImages[imageIndex]
                                                        ?.name
                                                }
                                            />
                                        </div>

                                        <button
                                            className="flex justify-center items-center hover:border hover:border-red-400 rounded-full w-8 h-8 text-gray-500 hover:text-red-500 transition-colors cursor-pointer shrink-0"
                                            onClick={removeCurrentImage}
                                            aria-label="Remove image"
                                        >
                                            ✕
                                        </button>
                                    </div>

                                    <section className="space-y-2 rounded-xl border border-primary-0/20 bg-primary-0/5 p-3">
                                        <div>
                                            <h2 className="text-sm font-medium">Add to board</h2>
                                            <p className="text-xs text-primary-0/60">Applied to every image in this upload queue.</p>
                                        </div>
                                        <UploadBoardActions
                                            actions={commonActions}
                                            onChange={(actions) => useUploadStore.getState().setCommonActions(actions)}
                                        />
                                    </section>

                                    <UploadImageForm
                                        key={storeImages[imageIndex]?.id}
                                        item={storeImages[imageIndex]}
                                        onChange={onFormChange}
                                    />
                                    {storeImages[imageIndex]
                                        ?.uploadedImageId && (
                                        <p
                                            role="status"
                                            className="text-green-600 text-sm"
                                        >
                                            Uploaded; finishing board actions
                                        </p>
                                    )}
                                    {storeImages[imageIndex]?.error && (
                                        <p
                                            role="alert"
                                            className="text-red-500 text-sm"
                                        >
                                            {storeImages[imageIndex].error}
                                        </p>
                                    )}
                                    <div className="flex justify-center items-center gap-4 mt-auto pt-6 border-surface-20 border-t w-full">
                                        {formButtons}
                                    </div>
                                    {user?.privilege !== "Moderator" && <p className="text-center text-xs text-muted-foreground">Standard accounts can upload 50 images per hour.</p>}
                                    {uploadProgress && <div role="status" className="space-y-2 text-sm"><span>{uploadProgress.label} {uploadProgress.value}%</span><Progress value={uploadProgress.value} /></div>}
                                </div>
                                {storeImages.length > 1 && (
                                    <ImageCarousel
                                        images={storeImages}
                                        currentIndex={imageIndex}
                                        onIndexChange={setImageIndex}
                                        selectedIndices={selectedIndices}
                                        setSelectedIndices={setSelectedIndices}
                                        onRemoveSelected={removeSelectedImages}
                                        getFileUrl={getFileUrl}
                                    />
                                )}
                            </>
                        ) : (
                            <div className="flex justify-center items-center w-full h-full text-gray-400 text-3xl text-center">
                                No image selected
                            </div>
                        )}
                    </div>
                </div>
            </div>
        ) : (
            <div className="flex flex-col justify-center items-center gap-4 h-full text-3xl">
                <span>drop images here</span>
                {browseImages}
            </div>
        );

    return (
        <AuthGuard>
            <title>upload</title>
            {!queueHydrated ? (
                <Loading />
            ) : (
                <DragDropZone
                    onFilesDropped={(files) => handleDroppedFiles(files)}
                >
                    {content}
                </DragDropZone>
            )}
        </AuthGuard>
    );
};

export default UploadPage;
