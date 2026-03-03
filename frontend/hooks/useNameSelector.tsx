import { noUse } from "@/app/AudioEffects";
import { useState } from "react";

const useNameSelector = (initialName?: string) => {
    const [name, setName] = useState<string>(initialName || "");
    const onNameKeyDown = (e: React.KeyboardEvent) => {
        const allowed = /^[a-z0-9\s]+$/i;
        if (!allowed.test(e.key)) {
            return;
        }

        let nameQuery = name;
        switch (e.key) {
            case "Backspace":
                if (e.ctrlKey || e.metaKey) nameQuery = "";
                else nameQuery = nameQuery.slice(0, -1);
                break;
            case "Escape":
                nameQuery = "";
                break;
            case " ":
                if (nameQuery.length == 0 || nameQuery.endsWith("_")) {
                    noUse();
                    return;
                }
                nameQuery = nameQuery.concat("_");
                break;
            default:
                if (e.key.length === 1)
                    nameQuery = nameQuery.concat(e.key.toLowerCase());
                break;
        }

        setName(nameQuery);
    };

    return { onNameKeyDown, name };
};

export default useNameSelector;
