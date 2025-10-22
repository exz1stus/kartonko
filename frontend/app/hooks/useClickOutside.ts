import { useEffect } from "react";

export const useClickOutside = (ref: React.RefObject<HTMLDivElement | null>, callback: () => void, except?: React.RefObject<HTMLDivElement | null>) => {
    useEffect(() => {
        const handleClickOutside = (event: any) => {
            if (ref.current && (!except || !except.current || !except.current.contains(event.target)) && !ref.current.contains(event.target)) {
                callback();
            }
        };
        document.addEventListener("mousedown", handleClickOutside);
        return () => {
            document.removeEventListener("mousedown", handleClickOutside);
        };
    }, [ref, except, callback]);
};

