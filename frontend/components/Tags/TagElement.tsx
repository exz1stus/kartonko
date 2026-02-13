"use client";
import React, { useEffect, useState } from "react";

interface Props {
    tag: string;
    removeTag?: (tag: string) => void;
    className?: string;
}

const TagElement: React.FC<Props> = ({ tag, removeTag, className }) => {
    const [deleteKeyPressed, setDeleteKeyPressed] = useState(false);
    const [hovered, setHovered] = useState(false);

    const interactive = removeTag !== undefined;

    const handleOnClick = () => {
        if (deleteKeyPressed) {
            removeTag?.(tag);
        }
    };

    useEffect(() => {
        if (!interactive) return;

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
            className={`flex items-center px-3 rounded-full text-lg transition
            ${deleteKeyPressed ? "shake" : "hover:"}
            ${deleteKeyPressed && hovered ? "line-through text-red-500" : ""}
            ${interactive ? "cursor-pointer" : ""}
            ${className}`}
            onMouseOver={() => {
                if (!hovered) setHovered(true);
            }}
            onMouseLeave={() => setHovered(false)}
            onClick={handleOnClick}
        >
            {tag}
        </div>
    );
};

export default TagElement;
