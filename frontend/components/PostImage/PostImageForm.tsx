"use client";
import React, { useEffect, useState } from "react";
import { Tag, TagSelector } from "@/components/Gallery/TagSelector";
import { Metadata, useUploadImage } from "./useUploadImage";

interface PostImageFormProps {
    file: File;
    onSubmit: () => void;
}

const PostImageForm: React.FC<PostImageFormProps> = ({ file, onSubmit }) => {
    const [tags, setTags] = useState<string[]>([]);
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

        const tagsNames: string[] = tags.map((tag) => {
            return tag;
        });

        const data: Metadata = {
            name: name || "",
            tags: tagsNames,
        };

        const uploaded = await uploadImage(data, file);
        if (uploaded) {
            onSubmit();
        }
    };

    const onTagsUpdate = (tags: string[]) => {
        setTags(tags);
    };

    // const tagElements = tags.map((tag, index) => <TagElement key={index} tag={tag} removeTag={removeTag} />);

    return (
        <form className="space-y-6 w-full" onSubmit={submitImage}>
            <div className="flex items-center gap-4">
                <label className="w-32 font-medium text-gray-700 text-sm shrink-0">Name</label>
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
                <label className="w-32 font-medium text-gray-700 text-sm shrink-0">Add tags</label>
                <div className="flex flex-col flex-1 gap-2">
                    {/* {tagElements} */}
                    <TagSelector selected={true} tags={tags} onTagsUpdate={onTagsUpdate} />
                </div>
            </div>
            <div className="flex justify-center">
                <button
                    className="w-96 px-12 py-1 text-gray-700 bg-white rounded-2xl hover:bg-blue-500 hover:text-white transition }"
                    type="submit"
                >
                    Upload
                </button>
            </div>
        </form>
    );
};

export default PostImageForm;
