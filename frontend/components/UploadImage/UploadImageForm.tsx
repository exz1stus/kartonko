"use client";
import React from "react";
import TagSelector from "@/components/Tags/TagSelector";
import NameField from "@/components/NameField";
import { UploadItem } from "@/hooks/useUploadStore";
import { useTranslations } from "@/contexts/LocaleContext";

interface UploadImageFormProps {
    item?: UploadItem;
    onChange?: (updates: Partial<Omit<UploadItem, "file">>) => void;
}

const UploadImageForm: React.FC<UploadImageFormProps> = ({
    item,
    onChange,
}) => {
    const { t } = useTranslations();
    if (!item) return null;
    const intialFilename = item.name;
    const format = item.file.name.split(".").pop();
    return (
        <div className="items-center gap-y-6 grid grid-cols-[1fr_2fr] grid-rows-[1fr_minmax(auto, 1fr)]">
            <label className="font-medium text-foreground text-sm">{t("image.name")}</label>
            <div className="control-shell flex items-center px-3 py-2 w-full">
                {item.uploadedImageId ? <span className="min-w-0 flex-1 truncate">{item.name}</span> : <NameField
                    onChangeSanitized={(value: string) =>
                        onChange?.({ name: value })
                    }
                    placeholder={intialFilename}
                    className="flex-1 bg-transparent border-none outline-none min-w-0 h-full"
                />}
                <span className="pl-2 text-[var(--control-placeholder)] text-sm">
                    .{format}
                </span>
            </div>
            <label className="font-medium text-foreground text-sm">
                {t("image.tags")}
            </label>
            {item.uploadedImageId ? <div className="text-sm">{[...item.tags, ...item.newTags].join(", ") || t("upload.noTags")}</div> : <TagSelector
                inputStyle={
                    "flex flex-wrap gap-1 w-full h-full px-3 py-2"
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
            />}
        </div>
    );
};

export default UploadImageForm;
