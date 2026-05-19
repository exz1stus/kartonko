import { useState } from "react";
import { useDebouncedCallback } from "use-debounce";

const useSanitizedField = (
    sanitizeFunction: (value: string) => string,
    onValueChange?: (value: string) => void,
    onChangeSanitized?: (value: string) => void,
    debounceTime: number = 500,
) => {
    const [sanitized, setSanitized] = useState<boolean>(true);

    const applySanitization = useDebouncedCallback((sanitizedName: string) => {
        onValueChange?.(sanitizedName);
        setSanitized(true);
        onChangeSanitized?.(sanitizedName);
    }, debounceTime);

    const onHighlightForSanitization = (value: string) => {
        onValueChange?.(value);

        if (value.trim().length === 0) {
            setSanitized(true);
            return;
        }

        const sanitizedName = sanitizeFunction(value);

        if (sanitizedName !== value) {
            setSanitized(false);
            return;
        }

        setSanitized(true);
    };

    const onSanitize = (value: string) => {
        onValueChange?.(value);

        if (value.trim().length === 0) {
            applySanitization.cancel();
            setSanitized(true);
            onChangeSanitized?.("");
            return;
        }

        const sanitizedName = sanitizeFunction(value);

        if (sanitizedName !== value) {
            setSanitized(false);
            applySanitization(sanitizedName);
            return;
        }

        applySanitization.cancel();
        setSanitized(true);
        onChangeSanitized?.(sanitizedName);
    };

    return { sanitized, onSanitize, onHighlightForSanitization };
};

export default useSanitizedField;
