import { noUse } from '@/app/AudioEffects'
import { useEffect } from 'react'

const useNameSelector = (selected: boolean, name: string, onNameUpdated: (name: string) => void) => {
    const onNameKeyDown = (e: KeyboardEvent) => {
        const allowed = /^[a-z0-9\s]+$/i;
        if (!allowed.test(e.key)) {
            return;
        }

        let nameQuery = name;
        switch (e.key) {
            case "Backspace":
                nameQuery = nameQuery.slice(0, -1);
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

        onNameUpdated(nameQuery);
    }

    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (!selected) return;
            onNameKeyDown(e)
        };

        window.addEventListener("keydown", handleKeyDown);
        return () => window.removeEventListener("keydown", handleKeyDown);
    }, [onNameKeyDown, selected]);

    return onNameUpdated;
}

export default useNameSelector;
