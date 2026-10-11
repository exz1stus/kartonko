"use client";

import { useTranslations } from "@/contexts/LocaleContext";

const labelKeys = {
    tags: "image.tags",
    resolution: "image.resolution",
    uploadedBy: "image.uploadedBy",
    uploaded: "image.uploaded",
    searchPreparing: "image.searchPreparing",
    searchFailed: "image.searchFailed",
} as const;

export default function ImageMetadataLabels({ kind }: { kind: keyof typeof labelKeys }) {
    const { t } = useTranslations();
    return <>{t(labelKeys[kind])}</>;
}
