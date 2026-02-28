import { useEffect, useState } from "react";

export default function useObjectUrl(file?: File) {
    const [url, setUrl] = useState<string>();

    useEffect(() => {
        if (!file) return;

        const objectUrl = URL.createObjectURL(file);
        setUrl(objectUrl);

        return () => URL.revokeObjectURL(objectUrl);
    }, [file]);

    return url;
}
