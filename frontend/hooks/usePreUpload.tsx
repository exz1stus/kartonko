"use client";
import { useRouter, usePathname } from "next/navigation";
import useUploadStore from "./useUploadStore";
import { getUploadSize } from "./useUploadStore";
import { toast } from "sonner";
import ProgressBarToast from "@/components/ProgressBarToast";
import { isUploadSizeValid } from "@/lib/image/upload";
import { useShallow } from "zustand/react/shallow";
import {
    checkForHashes,
    sanitizeFormats,
    sanitizeNames,
} from "@/lib/image/uploadPreprocessing";

const usePreUpload = () => {
    const { hasFile, addFiles } = useUploadStore((state) => state);
    const storeUploadSize = useUploadStore(useShallow(getUploadSize));
    const router = useRouter();
    const pathname = usePathname();

    const filterDuplicates = (files: File[]) => {
        return files.filter((incomingFile) => {
            const isAlreadyInStore = hasFile(incomingFile);
            return !isAlreadyInStore;
        });
    };

    const handleDroppedFiles = async (files: File[]) => {
        const sanitized = sanitizeNames(sanitizeFormats(files));
        const unique = filterDuplicates(sanitized);
        const batchSize = files.reduce((total, file) => total + file.size, 0);
        const toastId = toast("Checking files...", {
            description: (
                <ProgressBarToast value={0} currentName={unique[0].name} />
            ),
            duration: Infinity,
        });

        if (unique.length === 0) {
            toast.error("All files already added to upload", {
                id: toastId,
                duration: 3000,
            });
            return;
        }
        if (!isUploadSizeValid(batchSize + storeUploadSize)) {
            toast.error("The upload batch size maximum is 50MB", {
                id: toastId,
                duration: 3000,
            });
            return;
        }

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

        const readyFiles = await checkForHashes(unique, updateProgressBar);

        if (readyFiles.length === 0) {
            toast.error("All files already uploaded", {
                id: toastId,
                duration: 3000,
            });
            return;
        }

        toast.success(`Added ${readyFiles.length} files`, {
            id: toastId,
            duration: 3000,
        });

        addFiles(readyFiles);
        if (!pathname.startsWith("/upload")) {
            router.push("/upload");
        }
    };

    return handleDroppedFiles;
};

export default usePreUpload;
