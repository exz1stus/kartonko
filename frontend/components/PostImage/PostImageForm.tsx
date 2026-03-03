"use client";
import React, { useEffect, useState } from "react";
import { ImageUploadRequest, useUploadImage } from "@/hooks/useUploadImage";
import { toast } from "sonner";
import TagSelector from "@/components/Tags/TagSelector";
import useTags from "../Tags/useTags";
import { cn } from "@/lib/utils";

interface PostImageFormProps {
    file: File;
    onSubmit: () => void;
}

const PostImageForm: React.FC<PostImageFormProps> = ({ file, onSubmit }) => {
    const [name, setName] = useState<string>();

    const { uploadImage, loading } = useUploadImage();

    const { tags, removeTag, setTags } = useTags();

    useEffect(() => {
        const dotIndex = file.name.lastIndexOf(".");
        const name =
            dotIndex !== -1 ? file.name.substring(0, dotIndex) : file.name;
        setName(name);
    }, [file]);

    const extension = file.name.split(".").pop();

    const submitImage = async (e: React.FormEvent) => {
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
        <form className="space-y-6" onSubmit={submitImage}>
            <div className="flex sm:flex-row flex-col sm:items-center gap-2 sm:gap-4">
                <label className="w-32 font-medium text-gray-700 text-sm shrink-0">
                    Name
                </label>
                <div className="flex flex-1 items-center gap-2">
                    <input
                        type="text"
                        placeholder={name}
                        className="flex-1 px-3 py-2 border border-gray-300 rounded"
                        onChange={(e) => {
                            setName(e.target.value);
                        }}
                    />
                    <span className="text-gray-400 text-sm">.{extension}</span>
                </div>
            </div>
            <div className="flex items-center gap-4">
                <label className="w-32 font-medium text-gray-700 text-sm shrink-0">
                    Add tags
                </label>
                <div className="flex flex-wrap gap-1 w-full">
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
                        "hover:bg-blue-500 px-12 py-1 rounded-2xl w-96 text-gray-700 hover:text-white transition",
                        loading ? "animate-pulse bg-gray-400" : "bg-white",
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
