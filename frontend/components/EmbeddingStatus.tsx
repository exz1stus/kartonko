"use client";

import { useEffect, useState } from "react";
import { getImage } from "@/lib/api/generated/client";
import { useTranslations } from "@/contexts/LocaleContext";

export default function EmbeddingStatus({ imageId, initialStatus }: {
    imageId: number;
    initialStatus: string;
}) {
    const [status, setStatus] = useState(initialStatus);
    const { t } = useTranslations();

    useEffect(() => {
        if (status === "ready") return;

        let active = true;
        let inFlight = false;
        const refresh = async () => {
            if (inFlight) return;
            inFlight = true;
            try {
                const image = await getImage(imageId);
                if (active) setStatus(image.embedding_status);
            } catch {
                // A temporary API failure should not stop status updates.
            } finally {
                inFlight = false;
            }
        };
        void refresh();
        const interval = window.setInterval(() => void refresh(), 10_000);
        return () => {
            active = false;
            window.clearInterval(interval);
        };
    }, [imageId, status]);

    if (status === "ready") return null;
    return (
        <p role="status" className="text-sm text-muted-foreground">
            {t(status === "failed" ? "image.searchRetrying" : "image.searchPreparing")}
        </p>
    );
}
