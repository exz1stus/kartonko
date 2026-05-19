import { sanitizeName } from "@/lib/sanitizeName";
import React, { useState } from "react";
import { useDebouncedCallback } from "use-debounce";
import { Ref } from "react";
import { cn } from "@/lib/utils";
import { TruckElectric } from "lucide-react";

interface Props extends React.InputHTMLAttributes<HTMLInputElement> {
    placeholder?: string;
    onChangeSanitized?: (value: string) => void;
    ref?: Ref<HTMLInputElement>;
    className?: string;
}

const NameField = ({
    onChangeSanitized,
    placeholder,
    ref,
    className,
    ...props
}: Props) => {
    const [name, setName] = useState("");
    const [sanitized, setSanitized] = useState<boolean>(true);

    const applySanitization = useDebouncedCallback((sanitizedName: string) => {
        setName(sanitizedName);
        setSanitized(true);
        onChangeSanitized?.(sanitizedName);
    }, 500);

    const onFieldChange = (value: string) => {
        setName(value);

        if (value.trim().length === 0) {
            applySanitization.cancel();
            setSanitized(true);
            onChangeSanitized?.("");
            return;
        }

        const sanitizedName = sanitizeName(value);

        if (sanitizedName !== value) {
            setSanitized(false);
            applySanitization(sanitizedName);
            return;
        }

        applySanitization.cancel();
        setSanitized(true);
        onChangeSanitized?.(sanitizedName);
    };

    return (
        <input
            className={cn(className, "transition-colors", {
                "text-red-500": !sanitized,
            })}
            {...props}
            ref={ref}
            type="text"
            placeholder={placeholder}
            value={name}
            autoComplete="off"
            spellCheck={false}
            onChange={(e) => onFieldChange(e.target.value)}
        />
    );
};

export default NameField;
