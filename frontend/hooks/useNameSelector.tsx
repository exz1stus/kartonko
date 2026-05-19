import { useState } from "react";
import { sanitizeName } from "@/lib/sanitizeName";

const useNameSelector = (initialName?: string) => {
    const [name, setName] = useState<string>(
        initialName ? sanitizeName(initialName) : "",
    );

    const onNameKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === "Escape") {
            e.preventDefault();
            setName("");
        }
    };

    const onNameChange = (value: string) => {
        const sanitized = sanitizeName(value);
        setName(sanitized);
    };

    return { name, onNameChange, onNameKeyDown };
};

export default useNameSelector;
