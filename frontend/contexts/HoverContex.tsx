"use client";

import { createContext, useContext, useState, useEffect, useRef } from "react";

type HoverContextType = {
    hoveredElement: HTMLElement | null;
    onMouseOver: (element: HTMLElement) => void;
    onMouseOut: (element: HTMLElement) => void;
    isHovered: (element: HTMLElement) => boolean;
};

const HoverContext = createContext<HoverContextType | undefined>(undefined);

export default function HoverProvider({
    children,
}: {
    children: React.ReactNode;
}) {
    const [hoveredElement, setHoveredElement] = useState<HTMLElement | null>(
        null,
    );

    const onMouseOver = (element: HTMLElement) => {
        setHoveredElement(element);
    };

    const onMouseOut = (element: HTMLElement) => {
        if (isHovered(element)) {
            setHoveredElement(null);
        }
    };

    const isHovered = (element: HTMLElement) => hoveredElement === element;

    return (
        <HoverContext.Provider
            value={{ hoveredElement, onMouseOver, onMouseOut, isHovered }}
        >
            {children}
        </HoverContext.Provider>
    );
}

export function useHover<T extends HTMLElement>() {
    const ctx = useContext(HoverContext);
    if (!ctx) throw new Error("useHover must be used within HoverProvider");

    const { onMouseOver, onMouseOut, isHovered } = ctx;
    const ref = useRef<T | null>(null);

    useEffect(() => {
        const el = ref.current;
        if (!el) return;

        const handleMouseOver = () => onMouseOver(el);
        const handleMouseOut = () => onMouseOut(el);

        el.addEventListener("mouseover", handleMouseOver);
        el.addEventListener("mouseout", handleMouseOut);

        return () => {
            el.removeEventListener("mouseover", handleMouseOver);
            el.removeEventListener("mouseout", handleMouseOut);
        };
    }, [onMouseOver, onMouseOut]);

    return {
        ref,
        isHovered: () => (ref.current ? isHovered(ref.current) : false),
        onMouseOut,
        onMouseOver,
    };
}
