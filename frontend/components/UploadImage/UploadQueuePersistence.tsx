"use client";

import { useEffect } from "react";
import useUploadStore, { UploadItem } from "@/hooks/useUploadStore";
import { loadUploadQueue, saveUploadQueue } from "@/lib/image/uploadQueuePersistence";
import { useAuth } from "@/contexts/AuthContext";

const sameFile = (first: UploadItem, second: UploadItem) =>
    first.file.name === second.file.name &&
    first.file.size === second.file.size &&
    first.file.lastModified === second.file.lastModified;

export default function UploadQueuePersistence() {
    const { user, loading } = useAuth();
    useEffect(() => {
        if (loading) return;
        const ownerId = user?.id;
        useUploadStore.setState({ items: [], hydrated: !ownerId });
        if (!ownerId) return;

        let active = true;
        let unsubscribe: (() => void) | undefined;
        let pendingSnapshot: UploadItem[] | undefined;
        let writing = false;
        const saveLatest = (items: UploadItem[]) => {
            pendingSnapshot = items;
            if (writing) return;
            writing = true;
            void (async () => {
                while (pendingSnapshot !== undefined) {
                    const snapshot = pendingSnapshot;
                    pendingSnapshot = undefined;
                    try {
                        await saveUploadQueue(ownerId, snapshot);
                    } catch (error) {
                        console.warn("Could not save upload queue", error);
                    }
                }
                writing = false;
            })();
        };

        void loadUploadQueue(ownerId)
            .catch((error) => {
                console.warn("Could not restore upload queue", error);
                return [];
            })
            .then((savedItems) => {
                if (!active) return;
                const currentItems = useUploadStore.getState().items;
                const mergedItems = [...savedItems];
                for (const item of currentItems) {
                    const duplicateIndex = mergedItems.findIndex((saved) =>
                        saved.id === item.id || sameFile(saved, item),
                    );
                    if (duplicateIndex === -1) mergedItems.push(item);
                    else mergedItems[duplicateIndex] = item;
                }
                const commonActions = [...new Map(mergedItems.flatMap((item) => item.actions).map((action) => [`${action.kind}:${action.boardId}`, action])).values()];
                const queueItems = mergedItems.map((item) => ({ ...item, actions: commonActions }));
                useUploadStore.setState({ items: queueItems, commonActions, hydrated: true });
                unsubscribe = useUploadStore.subscribe((state) => {
                    saveLatest(state.items);
                });
                if (currentItems.length > 0) saveLatest(queueItems);
            });

        return () => {
            active = false;
            unsubscribe?.();
        };
    }, [loading, user?.id]);

    return null;
}
