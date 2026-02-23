"use client";
import AuthGuard from "@/components/AuthGuard";
import DragDropZone from "@/components/PostImage/DragDropZone";
import PostImageForm from "@/components/PostImage/PostImageForm";
import FancySpan from "@/components/template/FancySpan";
import useUploadStore from "@/hooks/useUploadStore";
import { Upload } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import Image from "next/image";

const UploadPage = () => {
    const images = useUploadStore((state) => state.files);
    const addFiles = useUploadStore((state) => state.addFiles);
    const removeFileAt = useUploadStore((state) => state.removeFileAt);

    const [imageIndex, setImageIndex] = useState(0);
    const [imageUrl, setImageUrl] = useState("");

    const handleWheel = (e: React.WheelEvent<HTMLDivElement>) => {
        setImageIndex((prev) => {
            const next = prev + Math.sign(e.deltaY);
            return Math.min(Math.max(next, 0), images.length - 1);
        });
    };

    const onFilesDropped = (files: File[]) => {
        //check for file format

        addFiles(files);
    };

    const removeCurrentImage = () => {
        if(images.length === 0) return;
        if (imageIndex > 0 && images.length - 1 === imageIndex) {
            setImageIndex((prev) => prev - 1);
        }

        removeFileAt(imageIndex);
    }

    useEffect(() => {
        const image = images?.[imageIndex];

        if (!image) {
            setImageUrl("");
            return;
        }

        const url = URL.createObjectURL(image);

        setImageUrl(url);

        return () => {
            URL.revokeObjectURL(url);
        };
    }, [images, imageIndex]);

    useEffect(()=>{
        const onEscape = (e: KeyboardEvent) => {
            if (e.key === "Escape") {
                removeCurrentImage();
            }
        };

        document.addEventListener("keydown", onEscape);
        return () => {
            document.removeEventListener("keydown", onEscape);
        };
    },[removeCurrentImage]);

    const content =
        images.length > 0 ? (
            <div className="flex justify-center backdrop-blur-sm p-10 w-full h-full overflow-hidden">
                <div
                    className="flex lg:flex-row flex-col justify-center bg-surface-0/95 rounded-2xl h-full lg:h-[80%]"
                    onWheel={handleWheel}
                >
                    {imageUrl ? (
                        <>
                            <Image
                                src={imageUrl}
                                alt={images[imageIndex]?.name}
                                className="rounded-t-2xl lg:rounded-l-2xl lg:rounded-tr-none w-full lg:w-auto max-h-[50vh] lg:max-h-full object-contain"
                                draggable={false}
                                width={800}
                                height={800}
                                unoptimized
                            />
                            <div className="flex flex-col gap-10 p-8">
                                <div className="flex justify-between">
                                    <FancySpan word={images[imageIndex]?.name} />
                                    <div className="flex justify-center items-center w-10 h-10 text-gray-700">
                                        {imageIndex + 1}/{images.length}
                                    </div>
                                </div>
                                <PostImageForm
                                    file={images[imageIndex]}
                                    onSubmit={removeCurrentImage}
                                />
                            </div>
                            <div className="flex justify-center items-center w-10 h-10 cursor-pointer" onClick={removeCurrentImage}>X</div>
                        </>
                    ) : (
                        <div>No image selected</div>
                    )}
                </div>
            </div>
        ) : (
            <div className="flex flex-col justify-center items-center gap-2 h-full text-3xl">
                <span>ту жуцай світлини дядьку</span>
                <Upload className="w-10 h-10" />
            </div>
        );

    return (
        <AuthGuard>
            <DragDropZone onFilesDropped={(files) => onFilesDropped(files)}>{content}</DragDropZone>
        </AuthGuard>
    );
};

export default UploadPage;
