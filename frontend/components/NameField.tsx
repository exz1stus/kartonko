import { sanitizeName } from "@/lib/sanitizeName";
import React, { useState } from "react";
import { Ref } from "react";
import { cn } from "@/lib/utils";
import useSanitizedField from "@/hooks/useSanitizedField";

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
    const { sanitized, onSanitize } = useSanitizedField(
        sanitizeName,
        setName,
        onChangeSanitized,
    );

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
            onChange={(e) => onSanitize(e.target.value)}
        />
    );
};

export default NameField;
