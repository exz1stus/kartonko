"use client";

import english from "@/messages/en.json";
import macedonian from "@/messages/mk.json";
import carpathoRusyn from "@/messages/rue.json";
import { useCallback, useSyncExternalStore } from "react";

export type Language = "en" | "mk" | "rue";
export type MessageKey = keyof typeof english;
export type MessageValues = Record<string, string | number>;
const LANGUAGE_STORAGE_KEY = "kartonko-language";
const LANGUAGE_EVENT = "kartonko-language-change";

function subscribe(onChange: () => void) {
    window.addEventListener("storage", onChange);
    window.addEventListener(LANGUAGE_EVENT, onChange);
    return () => {
        window.removeEventListener("storage", onChange);
        window.removeEventListener(LANGUAGE_EVENT, onChange);
    };
}

function getLanguage(): Language {
    const saved = localStorage.getItem(LANGUAGE_STORAGE_KEY);
    return saved === "mk" || saved === "rue" ? saved : "en";
}

export function setLanguage(language: Language) {
    localStorage.setItem(LANGUAGE_STORAGE_KEY, language);
    window.dispatchEvent(new Event(LANGUAGE_EVENT));
}

export function useTranslations() {
    const language = useSyncExternalStore(subscribe, getLanguage, () => "en");
    const messages: Record<MessageKey, string> = language === "mk"
        ? macedonian as Record<MessageKey, string>
        : language === "rue"
          ? carpathoRusyn as Record<MessageKey, string>
          : english;
    const t = useCallback((key: MessageKey, values?: MessageValues) => {
        const message = messages[key];
        return values ? message.replace(/\{(\w+)\}/g, (match, name: string) => String(values[name] ?? match)) : message;
    }, [messages]);
    return {
        language,
        setLanguage,
        t,
    };
}
