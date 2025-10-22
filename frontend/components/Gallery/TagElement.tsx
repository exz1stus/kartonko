import React, { useEffect, useState } from "react";
import { Tag } from "../PostImage/TagSelector";

interface Props {
    tag: Tag;
    removeTag: (tag: Tag) => void;
}

const TagElement: React.FC<Props> = ({ tag, removeTag }) => {
    const [deleteKeyPressed, setDeleteKeyPressed] = useState(false);
    const [hovered, setHovered] = useState(false);

    const handleOnClick = () => {
        if (deleteKeyPressed) {
            removeTag(tag);
        }
    };

    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.ctrlKey || e.metaKey) setDeleteKeyPressed(true);
        };
        const handleKeyUp = (e: KeyboardEvent) => {
            if (!e.ctrlKey && !e.metaKey) setDeleteKeyPressed(false);
        };
        window.addEventListener("keydown", handleKeyDown);
        window.addEventListener("keyup", handleKeyUp);
        return () => {
            window.removeEventListener("keydown", handleKeyDown);
            window.removeEventListener("keyup", handleKeyUp);
        };
    }, []);

    return (
        <div
            className={`flex items-center px-3 rounded-full text-lg cursor-pointer transition
            ${deleteKeyPressed ? "shake" : "hover:"}
            ${deleteKeyPressed && hovered ? "line-through text-red-500" : ""}`}
            onMouseOver={() => { if (!hovered) setHovered(true) }}
            onMouseLeave={() => setHovered(false)}
            onClick={handleOnClick}
        >
            {tag.name}
        </div>
    );
};

export default TagElement;