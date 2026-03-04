"use client";
import React, { useEffect, useState } from "react";
import { ImageUploadRequest, useUploadImage } from "@/hooks/useUploadImage";
import { toast } from "sonner";
import TagSelector from "@/components/Tags/TagSelector";
import useTags from "../Tags/useTags";
import { cn } from "@/lib/utils";
import useNameSelector from "@/hooks/useNameSelector";

interface PostImageFormProps {
    file: File;
    onSubmit: () => void;
}

const PostImageForm: React.FC<PostImageFormProps> = ({ file, onSubmit }) => {
    const { uploadImage, loading } = useUploadImage();

    const { tags, removeTag, setTags } = useTags();

    const dotIndex = file.name.lastIndexOf(".");
    const intialFilename =
        dotIndex !== -1 ? file.name.substring(0, dotIndex) : file.name;

    const extension = file.name.split(".").pop();

    const { name, onNameChange, onNameKeyDown } =
        useNameSelector(intialFilename);

    const submitImage = async (e: React.SubmitEvent) => {
        if (loading) return;
        e.preventDefault();

        if (file === null) {
            alert("image is null");
            return;
        }

        const data: ImageUploadRequest = {
            name: name || "",
            tags: tags,
        };

        const uploaded = toast.promise<boolean>(uploadImage(data, file), {
            loading: "Loading...",
            success: () => {
                onSubmit();
                return `image has been uploaded`;
            },
            error: (error) => error.message,
        });
    };

    return (
        <form className="space-y-6 w-full" onSubmit={submitImage}>
            <div className="flex sm:flex-row flex-col sm:items-center gap-2 sm:gap-4">
                <label className="w-[20%] font-medium text-gray-700 text-sm shrink-0">
                    Name
                </label>
                <div className="flex flex-1 items-center gap-2 border rounded-l">
                    <input
                        type="text"
                        placeholder={name}
                        value={name || ""}
                        onKeyDown={onNameKeyDown}
                        onChange={(e) => onNameChange(e.target.value)}
                        className="flex-1 px-3 py-2 border-none outline-none"
                    />
                </div>
                <span className="text-primary-0/50 text-sm">.{extension}</span>
            </div>
            <div className="flex items-center gap-4">
                <label className="w-[20%] font-medium text-gray-700 text-sm shrink-0">
                    Add tags
                </label>
                <div className="flex flex-wrap gap-1 border rounded-l w-full">
                    <TagSelector
                        tags={tags}
                        removeTag={removeTag}
                        onTagsUpdate={setTags}
                    />
                </div>
            </div>
            <div className="flex justify-center">
                <button
                    className={cn(
                        "px-12 py-1 border border-surface-20 hover:border-surface-30 rounded-2xl w-96 transition cursor-pointer glass",
                        loading ? "animate-pulse bg-surface-20" : "bg-none",
                    )}
                    type="submit"
                    disabled={loading}
                >
                    Upload
                </button>
            </div>
        </form>
    );
};

export default PostImageForm;
