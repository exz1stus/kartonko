"use client";
import { create } from "zustand";

export interface UploadItem {
    file: File;
    name: string;
    tags: string[];
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
