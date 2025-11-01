"use client";
import { useEffect, useState } from "react";

const useFullscreen = () => {
    const [isFullscreen, setIsFullscreen] = useState(false);

    useEffect(() => {
        if (typeof document === "undefined") return;

        const checkFullscreen = () => {
            setIsFullscreen(!!document.fullscreenElement);
        };

        checkFullscreen();

        document.addEventListener("fullscreenchange", checkFullscreen);
        return () => {
            document.removeEventListener("fullscreenchange", checkFullscreen);
        };
    }, []);

    return isFullscreen;
};

export default useFullscreen;
