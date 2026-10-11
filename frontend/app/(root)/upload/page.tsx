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
import { ImageBatchUploadRequest, uploadImageBatch } from "@/lib/image/upload";
import { runUploadAction } from "@/lib/image/runUploadAction";
import { uploadActionKey } from "@/lib/image/uploadIntent";
import Loading from "@/components/Loading";
import UploadBoardActions from "@/components/UploadImage/UploadBoardActions";

const UploadPage = () => {
    const {
        removeFileAt,
        updateMetadata,
        items: storeImages,
        commonActions,
        hydrated: queueHydrated,
    } = useUploadStore((state) => state);

    const objectUrlCache = useRef<Map<File, string>>(new Map());

    const [imageIndex, setImageIndex] = useState(0);
    const [imageUrl, setImageUrl] = useState("");

    const prevLengthRef = useRef(storeImages.length);

    const [loading, setLoading] = useState(false);
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
        if (storeImages.length === 0) return;
        if (imageIndex > 0 && storeImages.length - 1 === imageIndex) {
            setImageIndex((prev) => prev - 1);
        }

        removeFileAt(imageIndex);
    }, [imageIndex, removeFileAt, storeImages.length]);

    const removeImageAtIndex = useCallback(
        (index: number) => {
            if (storeImages.length === 0) return;
            if (imageIndex === index) {
                removeCurrentImage();
                return;
            }
            removeFileAt(index);
        },
        [storeImages.length, imageIndex, removeFileAt, removeCurrentImage],
    );

    const selectImageAtIndex = useCallback(
        (index: number) => {
            const image = storeImages?.[index];

            if (!image) {
                setImageUrl("");
                return;
            }

            if (objectUrlCache.current.has(image.file)) {
                setImageUrl(objectUrlCache.current.get(image.file)!);
                return;
            }

            const url = URL.createObjectURL(image.file);
            objectUrlCache.current.set(image.file, url);
            setImageUrl(url);
        },
        [storeImages],
    );

    useEffect(() => {
        return () => {
            objectUrlCache.current.forEach((url) => URL.revokeObjectURL(url));
            objectUrlCache.current.clear();
        };
    }, []);

    useEffect(() => {
        const prevLength = prevLengthRef.current;
        prevLengthRef.current = storeImages.length;

        if (storeImages.length === 0) return;

        if (storeImages.length > prevLength) {
            setImageIndex(storeImages.length - 1);
        } else {
            setImageIndex((prev) => Math.min(prev, storeImages.length - 1));
        }
    }, [storeImages]);

    useEffect(() => {
        selectImageAtIndex(imageIndex);
    }, [storeImages, imageIndex, selectImageAtIndex]);

    useEffect(() => {
        const onKeyDown = (e: KeyboardEvent) => {
            if (
                e.target instanceof HTMLInputElement ||
                e.target instanceof HTMLTextAreaElement
            ) {
                return;
            }
            if (e.key === "Delete") {
                if (selectedIndices.length === 0) return;

                const sortedIndices = [...selectedIndices].sort(
                    (a, b) => b - a,
                );

                sortedIndices.forEach((index) => {
                    removeImageAtIndex(index);
                });

                setSelectedIndices([]);
            }
        };

        document.addEventListener("keydown", onKeyDown);
        return () => {
            document.removeEventListener("keydown", onKeyDown);
        };
    }, [selectedIndices, removeImageAtIndex]);

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
        const toastId = toast.loading("Processing images…");
        try {
            const pending = useUploadStore
                .getState()
                .items.filter((item) => !item.uploadedImageId);
            if (pending.length > 0) {
                const request: ImageBatchUploadRequest = {
                    data: pending.map((item) => ({
                        client_id: item.id,
                        name: item.name,
                        tags: item.tags,
                        newTags: item.newTags,
                    })),
                    common_tags: [],
                };
                const response = await uploadImageBatch(
                    request,
                    pending.map((item) => item.file),
                    captchaToken ?? "",
                );
                if (
                    !response.results ||
                    response.results.length !== pending.length
                ) {
                    throw new Error(
                        "Upload server did not return per-image results",
                    );
                }
                response.results.forEach((result) => {
                    if (result.image)
                        useUploadStore
                            .getState()
                            .markUploaded(result.client_id, result.image.id);
                    else
                        useUploadStore
                            .getState()
                            .setItemError(
                                result.client_id,
                                result.error || "Upload failed",
                            );
                });
            }

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
                    useUploadStore.getState().removeItem(item.id);
                }
            }
            const remaining = useUploadStore.getState().items.length;
            if (remaining)
                toast.warning(`${remaining} images need attention`, {
                    id: toastId,
                });
            else
                toast.success("Images and follow-up actions complete", {
                    id: toastId,
                });
        } catch (error) {
            toast.error(
                error instanceof Error ? error.message : "Upload failed",
                { id: toastId },
            );
        } finally {
            setLoading(false);
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

    const multipleImagesText =
        storeImages.length > 1 ? ` ${storeImages.length} images` : "";
    const addNewTagsText =
        globalNewTags && globalNewTags.length > 0
            ? ` and create ${globalNewTags.length} new tags`
            : "";
    const uploadBtnText =
        multipleImagesText || addNewTagsText
            ? `${multipleImagesText}${addNewTagsText}`
            : " Image";

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
                        {loading ? "Uploading..." : `Upload ${uploadBtnText}`}
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
                                        className="rounded-t-2xl lg:rounded-l-2xl lg:rounded-tr-none w-full h-full object-contain"
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
                                        key={imageIndex}
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
                                </div>
                                {storeImages.length > 1 && (
                                    <ImageCarousel
                                        images={storeImages}
                                        currentIndex={imageIndex}
                                        onIndexChange={setImageIndex}
                                        selectedIndices={selectedIndices}
                                        setSelectedIndices={setSelectedIndices}
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
