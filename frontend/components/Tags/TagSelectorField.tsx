"use client";
import React from "react";
import useTagSelector from "./useTagSelector";

interface Props {
    active: boolean;
    tags: string[];
    onTagsUpdate: (tags: string[]) => void;
}

const TagSelectorField: React.FC<Props> = ({ active, tags, onTagsUpdate }) => {
    const { hint, tagField, difference } = useTagSelector({
        active,
        tags,
        onTagsUpdate,
    });

    return (
        <div className="flex flex-row items-center px-3">
            <span className={`text-lg ${hint.length > 0 ? "text-amber-200" : "text-primary-10"}`}>
                {tagField}
            </span>
            <span className="opacity-50 text-primary-10 text-lg">{difference}</span>
        </div>
    );
};

export default TagSelectorField;
