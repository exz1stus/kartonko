"use client";
import { create } from "zustand";
import { UploadAction, UploadIntent, mergeUploadActions } from "@/lib/image/uploadIntent";

export interface UploadItem {
    id: string;
    file: File;
    name: string;
    tags: string[];
    newTags: string[];
    actions: UploadAction[];
    completedActions: string[];
    uploadedImageId?: number;
    error?: string;
}

interface UploadStore {
    items: UploadItem[];
    commonActions: UploadAction[];
    hydrated: boolean;
    setHydrated: (hydrated: boolean) => void;
    setCommonActions: (actions: UploadAction[]) => void;
    hasFile: (file: File) => boolean;
    addFiles: (files: File[], intent?: UploadIntent) => void;

    setItems: (items: UploadItem[]) => void;
    updateMetadata: (
        index: number,
        updates: Partial<Omit<UploadItem, "file">>,
    ) => void;
    removeItems: (ids: string[]) => void;
    markUploaded: (id: string, imageId: number) => void;
    markActionComplete: (id: string, actionKey: string) => void;
    setItemError: (id: string, error?: string) => void;
    clearStore: () => void;
}

export const selectGlobalNewTags = (state: {
    items: UploadItem[];
}): string[] => {
    const allTags = state.items.flatMap((item) => item.newTags);
    return Array.from(new Set(allTags));
};

const isSameFile = (first: File, second: File) =>
    first.name === second.name && first.size === second.size && first.lastModified === second.lastModified;

const useUploadStore = create<UploadStore>((set, get) => ({
    items: [],
    commonActions: [],
    hydrated: false,
    setHydrated: (hydrated) => set({ hydrated }),
    setCommonActions: (actions) => set((state) => ({
        commonActions: actions,
        items: state.items.map((item) => ({ ...item, actions })),
    })),
    hasFile: (file) => {
        return get().items.some(
            (storedItem) =>
                isSameFile(storedItem.file, file),
        );
    },
    setItems: (items) => set({ items }),
    addFiles: (files, intent = {}) =>
        set((state) => {
            const commonActions = mergeUploadActions(state.commonActions, intent.actions ?? []);
            const items = [...state.items];
            files.forEach((file) => {
                const existingIndex = items.findIndex((item) => isSameFile(item.file, file));
                if (existingIndex >= 0) {
                    const existing = items[existingIndex];
                    items[existingIndex] = {
                        ...existing,
                        tags: existing.uploadedImageId ? existing.tags : [...new Set([...existing.tags, ...(intent.tags ?? [])])],
                        actions: commonActions,
                    };
                    return;
                }
                const dotIndex = file.name.lastIndexOf(".");
                const baseName =
                    dotIndex !== -1
                        ? file.name.substring(0, dotIndex)
                        : file.name;
                items.push({
                    id: crypto.randomUUID(),
                    file,
                    name: baseName,
                    tags: [...new Set(intent.tags ?? [])],
                    newTags: [],
                    actions: commonActions,
                    completedActions: [],
                });
            });
            return { items: items.map((item) => ({ ...item, actions: commonActions })), commonActions };
        }),
    updateMetadata: (index, updates) =>
        set((state) => ({
            items: state.items.map((item, i) =>
                i === index ? { ...item, ...updates } : item,
            ),
        })),
    removeItems: (ids) => set((state) => {
        const removed = new Set(ids);
        return { items: state.items.filter((item) => !removed.has(item.id)) };
    }),
    markUploaded: (id, imageId) => set((state) => ({ items: state.items.map((item) => item.id === id ? { ...item, uploadedImageId: imageId, error: undefined } : item) })),
    markActionComplete: (id, actionKey) => set((state) => ({ items: state.items.map((item) => item.id === id ? { ...item, completedActions: [...new Set([...item.completedActions, actionKey])], error: undefined } : item) })),
    setItemError: (id, error) => set((state) => ({ items: state.items.map((item) => item.id === id ? { ...item, error } : item) })),
    clearStore: () => set({ items: [] }),
}));

export default useUploadStore;
