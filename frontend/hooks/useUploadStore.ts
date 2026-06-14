"use client";
import { create } from "zustand";

export interface UploadItem {
    file: File;
    name: string;
    tags: string[];
    newTags: string[];
}

interface UploadStore {
    items: UploadItem[];
    hasFile: (file: File) => boolean;
    addFiles: (files: File[]) => void;

    setItems: (items: UploadItem[]) => void;
    updateMetadata: (
        index: number,
        updates: Partial<Omit<UploadItem, "file">>,
    ) => void;
    removeFileAt: (index: number) => void;
    clearStore: () => void;
}

export const selectGlobalNewTags = (state: {
    items: UploadItem[];
}): string[] => {
    const allTags = state.items.flatMap((item) => item.newTags);
    return Array.from(new Set(allTags));
};

export const getUploadSize = (state: { items: UploadItem[] }): number =>
    state.items.reduce((total, item) => total + item.file.size, 0);

const useUploadStore = create<UploadStore>((set, get) => ({
    items: [],
    hasFile: (file) => {
        return get().items.some(
            (storedItem) =>
                storedItem.file.name === file.name &&
                storedItem.file.size === file.size,
        );
    },
    setItems: (items) => set({ items }),
    addFiles: (files) =>
        set((state) => {
            const newItems: UploadItem[] = files.map((file) => {
                const dotIndex = file.name.lastIndexOf(".");
                const baseName =
                    dotIndex !== -1
                        ? file.name.substring(0, dotIndex)
                        : file.name;
                return {
                    file,
                    name: baseName,
                    tags: [],
                    newTags: [],
                };
            });
            return { items: [...state.items, ...newItems] };
        }),
    updateMetadata: (index, updates) =>
        set((state) => ({
            items: state.items.map((item, i) =>
                i === index ? { ...item, ...updates } : item,
            ),
        })),
    removeFileAt: (index) =>
        set((state) => ({
            items: state.items.filter((_, i) => i !== index),
        })),
    clearStore: () => set({ items: [] }),
}));

export default useUploadStore;
