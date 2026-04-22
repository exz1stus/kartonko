"use client";
import { create } from "zustand";

interface UploadStore {
    files: File[];
    setFiles: (files: File[]) => void;
    addFiles: (files: File[]) => void;
    removeFileAt: (index: number) => void;
    clearFiles: () => void;
}

const useUploadStore = create<UploadStore>((set) => {
    return {
        files: [],
        setFiles: (files) => set({ files }),
        addFiles: (files) =>
            set((state) => ({
                files: [...state.files, ...files],
            })),
        removeFileAt: (index) =>
            set((state) => ({
                files: state.files.filter((_, i) => i !== index),
            })),
        clearFiles: () => set({ files: [] }),
    };
});

export default useUploadStore;
