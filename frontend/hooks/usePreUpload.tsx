"use client";
import { useRouter, usePathname } from "next/navigation";
import useUploadStore from "./useUploadStore";
import { toast } from "sonner";
import ProgressBarToast from "@/components/ProgressBarToast";
import {
    checkForHashes,
    sanitizeFormats,
    sanitizeNames,
} from "@/lib/image/uploadPreprocessing";
import { UploadIntent } from "@/lib/image/uploadIntent";
import { runUploadAction } from "@/lib/image/runUploadAction";
import { mergeUploadActions } from "@/lib/image/uploadIntent";

const usePreUpload = () => {
    const hasFile = useUploadStore((state) => state.hasFile);
    const addFiles = useUploadStore((state) => state.addFiles);
    const router = useRouter();
    const pathname = usePathname();

    const handleDroppedFiles = async (files: File[], intent: UploadIntent = {}) => {
        const effectiveIntent = {
            ...intent,
            actions: mergeUploadActions(useUploadStore.getState().commonActions, intent.actions ?? []),
        };
        const sanitized = sanitizeNames(sanitizeFormats(files));
        if (sanitized.length === 0) {
            toast.error("No supported image files found");
            return;
        }
        const seen = new Set<string>();
        const unique = sanitized.filter((file) => {
            const key = `${file.name}:${file.size}:${file.lastModified}`;
            if (hasFile(file) || seen.has(key)) return false;
            seen.add(key);
            return true;
        });
        const queued = sanitized.filter((file) => hasFile(file));
        if (queued.some((file) => useUploadStore.getState().items.some((item) =>
            item.uploadedImageId && item.file.name === file.name && item.file.size === file.size &&
            (effectiveIntent.tags ?? []).some((tag) => !item.tags.includes(tag))))) {
            toast.warning("Already uploaded images cannot receive new upload tags");
        }
        if (queued.length > 0) addFiles(queued, effectiveIntent);
        if (unique.length === 0) {
            if (queued.length > 0) {
                toast.success("Updated queued images");
                if (!pathname.startsWith("/upload")) router.push("/upload");
            }
            return;
        }
        const toastId = toast("Checking files...", {
            description: (
                <ProgressBarToast value={0} currentName={unique[0].name} />
            ),
            duration: Infinity,
        });

        const updateProgressBar = (completed: number) => {
            const percentage = Math.round((completed / unique.length) * 100);

            toast.message("Checking files...", {
                id: toastId,
                description: (
                    <ProgressBarToast
                        value={percentage}
                        currentName={unique[completed - 1].name}
                    />
                ),
            });
        };

        try {
            const { ready: readyFiles, existing } = await checkForHashes(unique, updateProgressBar);

            for (const image of existing) {
                for (const action of effectiveIntent.actions ?? []) {
                    try {
                        await runUploadAction(action, image.id);
                    } catch (error) {
                        toast.error(`Could not apply ${action.kind} to ${image.filename}: ${error instanceof Error ? error.message : "Unknown error"}`);
                    }
                }
            }

            if (readyFiles.length === 0) {
                toast.message(existing.length > 0 ? `${existing.length} existing images processed` : "All files already uploaded", {
                    id: toastId,
                    duration: 3000,
                });
                return;
            }

            toast.success(`Added ${readyFiles.length} files`, {
                id: toastId,
                duration: 3000,
            });

            addFiles(readyFiles, effectiveIntent);
            if (!pathname.startsWith("/upload")) {
                router.push("/upload");
            }
        } catch (error) {
            toast.error(error instanceof Error ? error.message : "Could not check files", { id: toastId, duration: 4000 });
        }
    };

    return handleDroppedFiles;
};

export default usePreUpload;
