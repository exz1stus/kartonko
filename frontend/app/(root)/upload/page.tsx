"use client";
import AuthGuard from "@/components/AuthGuard";
import DragDropZone from "@/components/PostImage/DragDropZone";
import PostImageForm from "@/components/PostImage/PostImageForm";
import FancySpan from "@/components/template/FancySpan";
import useUploadStore from "@/hooks/useUploadStore";
import { UploadIcon } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";
import Image from "next/image";
import useUpload from "@/hooks/useUpload";

const UploadPage = () => {
    const images = useUploadStore((state) => state.files);
    const addFiles = useUploadStore((state) => state.addFiles);
    const removeFileAt = useUploadStore((state) => state.removeFileAt);

    const [imageIndex, setImageIndex] = useState(0);
    const [imageUrl, setImageUrl] = useState("");

    const prevLengthRef = useRef(images.length);

    const handleWheel = (e: React.WheelEvent<HTMLDivElement>) => {
        setImageIndex((prev) => {
            const next = prev + Math.sign(e.deltaY);
            return Math.min(Math.max(next, 0), images.length - 1);
        });
    };

    const { handleUploadFiles } = useUpload();

    const removeCurrentImage = () => {
        if (images.length === 0) return;
        if (imageIndex > 0 && images.length - 1 === imageIndex) {
            setImageIndex((prev) => prev - 1);
        }

        removeFileAt(imageIndex);
    };

    const selectImageAtIndex = useCallback(
        (index: number) => {
            const image = images?.[index];

            if (!image) {
                setImageUrl("");
                return;
            }

            const url = URL.createObjectURL(image);
            setImageUrl(url);
        },
        [images],
    );

    useEffect(() => {
        const prevLength = prevLengthRef.current;
        prevLengthRef.current = images.length;

        if (images.length === 0) return;

        if (images.length > prevLength) {
            setImageIndex(images.length - 1);
        } else {
            setImageIndex((prev) => Math.min(prev, images.length - 1));
        }
    }, [images]);

    useEffect(() => {
        selectImageAtIndex(imageIndex);
    }, [images, imageIndex]);

    useEffect(() => {
        const onEscape = (e: KeyboardEvent) => {
            if (e.key === "Escape") {
                removeCurrentImage();
            }
        };

        document.addEventListener("keydown", onEscape);
        return () => {
            document.removeEventListener("keydown", onEscape);
        };
    }, [removeCurrentImage]);

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

    const content =
        images.length > 0 ? (
            <div className="flex justify-center items-start p-4 lg:p-10 w-full h-full overflow-hidden">
                <div
                    className="flex lg:flex-row flex-col bg-surface-0/95 border border-surface-20 rounded-2xl w-full h-full max-h-full overflow-hidden glass"
                    onWheel={handleWheel}
                >
                    {imageUrl ? (
                        <>
                            <div className="relative flex justify-center items-center bg-black/20 w-full lg:w-[60%] h-[40vh] lg:h-full overflow-hidden shrink-0">
                                <Image
                                    src={imageUrl}
                                    alt={images[imageIndex]?.name}
                                    className="rounded-t-2xl lg:rounded-l-2xl lg:rounded-tr-none w-full h-full object-contain"
                                    draggable={false}
                                    width={800}
                                    height={800}
                                    unoptimized
                                />
                            </div>

                            <div className="flex flex-col gap-6 p-5 lg:p-8 w-full lg:w-[40%] min-h-0 overflow-y-auto">
                                <div className="flex flex-wrap justify-between items-center gap-2">
                                    <div className="inline-flex items-center gap-2">
                                        {browseImages}
                                        <span className="tabular-nums text-gray-500 text-sm shrink-0">
                                            {imageIndex + 1}/{images.length}
                                        </span>
                                    </div>

                                    <div className="flex items-center gap-2 min-w-0">
                                        <FancySpan
                                            word={images[imageIndex]?.name}
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

                                {/* Form */}
                                <PostImageForm
                                    key={images[imageIndex].name}
                                    file={images[imageIndex]}
                                    onSubmit={removeCurrentImage}
                                />
                            </div>
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
