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
import useUpload from "@/hooks/useUpload";
import {
    ImageBatchUploadRequest,
    ImageUploadRequest,
    useUploadImage,
} from "@/hooks/useUploadImage";
import { toast } from "sonner";
import CaptchaButton from "@/components/UploadImage/CaptchaButton";
import { useShallow } from "zustand/react/shallow";
import ImageCarousel from "@/components/UploadImage/ImageCarousel";

const UploadPage = () => {
    const {
        removeFileAt,
        clearStore,
        updateMetadata,
        items: storeImages,
    } = useUploadStore((state) => state);

    const objectUrlCache = useRef<Map<File, string>>(new Map());

    const [imageIndex, setImageIndex] = useState(0);
    const [imageUrl, setImageUrl] = useState("");
    // const [previousName, setPreviousName] = useState("");

    const prevLengthRef = useRef(storeImages.length);

    const { uploadImage, uploadImageBatch, loading } = useUploadImage();
    const [captchaToken, setCaptchaToken] = useState<string | null>();
    const [selectedIndices, setSelectedIndices] = useState<number[]>([]);

    const globalNewTags = useUploadStore(useShallow(selectGlobalNewTags));

    const handleWheel = (e: React.WheelEvent<HTMLDivElement>) => {
        e.preventDefault();
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

    const { handleUploadFiles } = useUpload();
    const onSubmitCurrentImage = () => {
        removeCurrentImage();
    };

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
                    handleUploadFiles(Array.from(e.target.files || []))
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

    const submitBatch = async (
        data: ImageBatchUploadRequest,
        files: File[],
    ) => {
        if (loading) return;

        toast.promise(uploadImageBatch(data, files, captchaToken ?? ""), {
            loading: `Uploading batch of ${files.length} images...`,
            success: () => {
                clearStore();
                setCaptchaToken(null);
                return "All files uploaded successfully!";
            },
            error: (error) => error.message || "Batch upload failed",
        });
    };

    const submitImage = async (data: ImageUploadRequest, file: File) => {
        if (loading) return;

        toast.promise<boolean>(uploadImage(data, file, captchaToken ?? ""), {
            loading: "Loading...",
            success: () => {
                onSubmitCurrentImage();
                return `image has been uploaded`;
            },
            error: (error) => error.message,
        });
    };

    const onUpload = async () => {
        if (storeImages.length === 0) return;

        if (storeImages.length === 1) {
            const currentImage = storeImages[imageIndex];
            const data = {
                name: currentImage.name,
                tags: currentImage.tags,
                newTags: currentImage.newTags,
            };

            await submitImage(data, currentImage.file);
            return;
        }

        const batchMetadata: ImageUploadRequest[] = storeImages.map((item) => ({
            name: item.name,
            tags: item.tags,
            newTags: item.newTags,
        }));

        const request: ImageBatchUploadRequest = {
            data: batchMetadata,
            common_tags: [],
        };
        const files = storeImages.map((item) => item.file);

        await submitBatch(request, files);
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
                <CaptchaButton
                    className={`max-w-[70%]
                        ${loading ? "animate-pulse bg-surface-20" : "bg-none"}
                    `}
                    onClick={onUpload}
                    disabled={loading}
                    onVerifySuccess={setCaptchaToken}
                >
                    {loading ? "Uploading..." : `Upload ${uploadBtnText}`}
                </CaptchaButton>
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
                <div className="flex lg:flex-row flex-col bg-surface-0/95 border border-surface-20 rounded-2xl w-full h-full max-h-full overflow-hidden glass">
                    {imageUrl ? (
                        <>
                            <div className="relative flex flex-1 justify-center items-center bg-black/20 w-full lg:w-[60%] min-w-0 h-[20vh] lg:h-full min-h-0 overflow-hidden shrink-0">
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
                                            word={storeImages[imageIndex]?.name}
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

                                <UploadImageForm
                                    key={imageIndex}
                                    item={storeImages[imageIndex]}
                                    onChange={onFormChange}
                                />
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
        ) : (
            <div className="flex flex-col justify-center items-center gap-4 h-full text-3xl">
                <span>drop images here</span>
                {browseImages}
            </div>
        );

    return (
        <AuthGuard>
            <title>upload</title>
            <DragDropZone onFilesDropped={(files) => handleUploadFiles(files)}>
                {content}
            </DragDropZone>
        </AuthGuard>
    );
};

export default UploadPage;
