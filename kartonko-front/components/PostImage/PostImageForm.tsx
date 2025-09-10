"use client";
import React, { useEffect, useState } from "react";
import { TagSelector, Tag } from "./TagSelector";

interface PostImageFormProps {
    image: File;
    onSubmit: (response: ApiResponse) => void;
}

export interface ApiResponse {
    error?: string;
    message?: string;
}
const API_ORIGIN = process.env.NEXT_PUBLIC_API_ORIGIN;

const PostImageForm: React.FC<PostImageFormProps> = ({ image, onSubmit }) => {
    const [tags, setTags] = useState<Tag[]>([]);
    const [name, setName] = useState<string>();


    useEffect(() => {
        const dotIndex = image.name.lastIndexOf(".");
        const name = dotIndex !== -1 ? image.name.substring(0, dotIndex) : image.name;
        setName(name);
    }, [image]);

    const extension = image.name.split(".").pop();

    const submitImage = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!image) {
            alert("image is null");
        }

        const tagsNames: string[] = tags.map((tag) => { return tag.name });

        const metadata = {
            name: name,
            tags: tagsNames,
        };

        const formData = new FormData();

        formData.append("metadata", JSON.stringify(metadata));
        formData.append("file", image);

        try {
            const res = await fetch(`${API_ORIGIN}/upload`, {
                method: "POST",
                body: formData,
                credentials: 'include'
            });

            const response: ApiResponse = await res.json();
            onSubmit(response);
        } catch (err) {
            console.error(err);
            onSubmit({ error: "Failed to upload, server doesn't respond" });
        }
    };

    return (
        <form className="space-y-6 w-full" onSubmit={submitImage}>
            <div className="flex items-center gap-4">
                <label className="w-32 shrink-0 text-sm font-medium text-gray-700">
                    Name
                </label>
                <div className="flex flex-1 items-center gap-2">
                    <input
                        type="text"
                        placeholder={name}
                        className="flex-1 border border-gray-300 rounded px-3 py-2"
                        onChange={(e) => { setName(e.target.value) }}
                    />
                    <span className="text-gray-400 text-sm">.{extension}</span>
                </div>
            </div>
            <div className="flex items-center gap-4">
                <label className="w-32 shrink-0 text-sm font-medium text-gray-700">
                    Add tags
                </label>
                <div className="flex-1 flex flex-col gap-2">
                    <TagSelector selectedTags={tags} setSelectedTags={setTags} />
                </div>
            </div>
            <div className="flex justify-center">
                <button className="w-96 px-12 py-1 text-gray-700 bg-white rounded-2xl hover:bg-blue-500 hover:text-white transition }" type="submit">
                    Upload
                </button>
            </div>
        </form>
    );
};

export default PostImageForm;
