"use client";
import React from "react";
import TagSelector from "@/components/Tags/TagSelector";
import NameField from "@/components/NameField";
import { UploadItem } from "@/hooks/useUploadStore";

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
    const format = item.file.name.split(".").pop();
    return (
        <form className="items-center gap-y-6 grid grid-cols-[1fr_2fr] grid-rows-[1fr_minmax(auto, 1fr)]">
            <label className="font-medium text-gray-700 text-sm">Name</label>
            <div className="flex items-center bg-neutral-900 px-3 py-2 border rounded w-full">
                <NameField
                    onChangeSanitized={(value: string) =>
                        onChange?.({ name: value })
                    }
                    placeholder={intialFilename}
                    className="flex-1 bg-transparent border-none outline-none min-w-0 h-full"
                />
                <span className="pl-2 text-primary-0/50 text-sm">
                    .{format}
                </span>
            </div>
            <label className="font-medium text-gray-700 text-sm">
                Add tags
            </label>
            <TagSelector
                inputStyle={
                    "flex flex-wrap gap-1 border rounded-t-md w-full h-full bg-neutral-900 px-3 py-2"
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
        </form>
    );
};

export default UploadImageForm;
