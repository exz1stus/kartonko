"use client";
import { useState } from "react";

const useTags = () => {
    const [tags, setTags] = useState<string[]>([]);

    const removeTag = (tag: string) => {
        setTags(tags.filter((t) => t !== tag));
    };

    return { tags, setTags, removeTag };
};

export default useTags;
