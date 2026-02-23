"use client";
import TagElement from "./TagElement";

interface Props {
    tags: string[];
    removeTag?: (tag: string) => void;
    tagStyle?: string;
}

const TagSpan = ({ tags, removeTag, tagStyle }: Props) => {
    const tagElements = tags.map((tag, index) => (
        <TagElement key={index} tag={tag} removeTag={removeTag} className={tagStyle} />
    ));

    return <>{tagElements}</>;
};

export default TagSpan;
