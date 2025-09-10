import { noUse } from '@/app/ AudioEffects'
import React, { useEffect } from 'react'

interface NameSelectorProps {
    selected: boolean
    name: string
    onUpdateName: (name: string) => void
}

const NameSelector: React.FC<NameSelectorProps> = ({ selected, name, onUpdateName }) => {
    const onNameKeyDown = (e: KeyboardEvent) => {
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

        onUpdateName(nameQuery);
    }

    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (!selected) return;
            onNameKeyDown(e)
        };

        window.addEventListener("keydown", handleKeyDown);
        return () => window.removeEventListener("keydown", handleKeyDown);
    }, [onNameKeyDown, selected]);

    return (
        <span className="text-2xl">{name.toLowerCase()}</span>
    )
}

export default NameSelector
