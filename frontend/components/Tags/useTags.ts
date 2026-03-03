"use client";
import { useState } from "react";

const useTags = (intialTags?: string[]) => {
    const [tags, setTags] = useState<string[]>(intialTags || []);

    const removeTag = (tag: string) => {
        setTags(tags.filter((t) => t !== tag));
    };

    return { tags, setTags, removeTag };
};

export default useTags;
