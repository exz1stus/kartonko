"use client";
import { useState, useEffect } from "react";

function useOrientation() {
    const [orientation, setOrientation] = useState<"portrait" | "landscape">(
        typeof window !== "undefined" && window.innerWidth > window.innerHeight
            ? "landscape"
            : "portrait",
    );

    useEffect(() => {
        const handleResize = () => {
            setOrientation(window.innerWidth > window.innerHeight ? "landscape" : "portrait");
        };

        window.addEventListener("resize", handleResize);
        return () => window.removeEventListener("resize", handleResize);
    }, []);

    return orientation;
}

export default useOrientation;
