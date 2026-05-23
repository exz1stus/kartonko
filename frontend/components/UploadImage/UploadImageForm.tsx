"use client";
import React, { useState } from "react";
import TagSelector from "@/components/Tags/TagSelector";
import NameField from "@/components/NameField";
import useUploadStore, {
    selectGlobalNewTags,
    UploadItem,
} from "@/hooks/useUploadStore";

interface UploadImageFormProps {
    item?: UploadItem;
    onChange?: (updates: Partial<Omit<UploadItem, "file">>) => void;
}

const UploadImageForm: React.FC<UploadImageFormProps> = ({
    item,
    onChange,
}) => {
    if (!item) return null;
    const intialFilename = item.name;
    const extension = item.file.name.split(".").pop();
    return (
        <form className="space-y-6">
            <div className="flex sm:flex-row flex-col sm:items-center gap-2 sm:gap-4">
                <label className="w-[20%] font-medium text-gray-700 text-sm shrink-0">
                    Name
                </label>
                <div className="flex flex-1 items-center gap-2 border rounded">
                    <NameField
                        onChangeSanitized={(value: string) => {
                            onChange?.({ name: value });
                        }}
                        placeholder={intialFilename}
                        className="flex-1 bg-neutral-900 px-3 py-2 border-none outline-none"
                    />
                    <span className="px-2 text-primary-0/50 text-sm">
                        .{extension}
                    </span>
                </div>
            </div>
            <div className="flex items-center gap-4">
                <label className="w-[20%] font-medium text-gray-700 text-sm shrink-0">
                    Add tags
                </label>
                <TagSelector
                    className={
                        "flex flex-wrap gap-1 border rounded-t-md w-full bg-neutral-900"
                    }
                    tags={item.tags}
                    removeTag={(tag: string) => {
                        const newTags = item.tags.filter((t) => t !== tag);
                        onChange?.({ tags: newTags });
                    }}
                    onTagsUpdate={(tags: string[]) => {
                        onChange?.({ tags: tags });
                    }}
                    newTags={item.newTags}
                    removeNewTag={(tag: string) => {
                        const newTags = item.newTags.filter((t) => t !== tag);
                        onChange?.({ newTags: newTags });
                    }}
                    onNewTagsUpdate={(tags: string[]) => {
                        onChange?.({ newTags: tags });
                    }}
                />
            </div>
        </form>
    );
};

export default UploadImageForm;
