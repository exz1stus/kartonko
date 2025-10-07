"use client";
import React, { use, useEffect, useState } from "react";
import { TagSelector, Tag } from "./TagSelector";
import { Metadata, useUploadImage } from "./useUploadImage";

interface PostImageFormProps {
    file: File;
    onSubmit: () => void;
}

const PostImageForm: React.FC<PostImageFormProps> = ({ file, onSubmit }) => {
    const [tags, setTags] = useState<Tag[]>([]);
    const [name, setName] = useState<string>();

    const { uploadImage } = useUploadImage();

    useEffect(() => {
        const dotIndex = file.name.lastIndexOf(".");
        const name = dotIndex !== -1 ? file.name.substring(0, dotIndex) : file.name;
        setName(name);
    }, [file]);

    const extension = file.name.split(".").pop();

    const submitImage = async (e: React.FormEvent) => {
        e.preventDefault();

        if (file === null) {
            alert("image is null");
            return;
        }

        const tagsNames: string[] = tags.map((tag) => { return tag.name });

        const data: Metadata = {
            name: name || "",
            tags: tagsNames,
        }

        const uploaded = await uploadImage(data, file);
        if (uploaded) {
            onSubmit();
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
