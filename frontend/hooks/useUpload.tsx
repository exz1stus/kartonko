"use client";
import { useRouter, usePathname } from "next/navigation";
import useUploadStore from "./useUploadStore";
import { getUploadSize } from "./useUploadStore";
import isAllowed from "@/lib/image/allowedFormats";
import { toast } from "sonner";
import { hashFile } from "@/lib/image/image";
import { existsOnServer } from "@/lib/image/image.client";
import ProgressBarToast from "@/components/ProgressBarToast";
import pLimit from "p-limit";
import { sanitizeName } from "@/lib/sanitizeName";
import { isUploadSizeValid } from "@/lib/image/imageUpload";
import { useShallow } from "zustand/react/shallow";

const useUpload = () => {
    const { hasFile, addFiles } = useUploadStore((state) => state);
    const storeUploadSize = useUploadStore(useShallow(getUploadSize));
    const router = useRouter();
    const pathname = usePathname();

    const sanitizeNames = (files: File[]): File[] => {
        return files.map((file) => {
            const lastDotIndex = file.name.lastIndexOf(".");

            let baseName = file.name;
            let extension = "";

            if (lastDotIndex > 0) {
                baseName = file.name.substring(0, lastDotIndex);
                extension = file.name.substring(lastDotIndex);
            }
            const cleanBaseName = sanitizeName(baseName);
            const cleanName = `${cleanBaseName}${extension}`;
            if (cleanName === file.name) return file;
            return new File([file], cleanName, {
                type: file.type,
                lastModified: file.lastModified,
            });
        });
    };

    const sanitizeFormats = (files: File[]): File[] =>
        files.filter((file) => {
            const format = file.type.split("/")[1];
            if (!isAllowed(format))
                toast.error(`format ${format} is not allowed`);
            return isAllowed(format);
        });

    const limit = pLimit(10);

    const checkForHashes = async (
        files: File[],
        onProgress: (completed: number) => void,
    ): Promise<File[]> => {
        let completed = 0;

        const tasks = files.map((file) =>
            limit(async () => {
                try {
                    const hash = await hashFile(file);
                    const exists = await existsOnServer(hash);
                    return exists ? null : file;
                } finally {
                    completed++;
                    onProgress(completed);
                }
            }),
        );

        const results = await Promise.all(tasks);
        return results.filter((f): f is File => f !== null);
    };

    const filterDuplicates = (files: File[]) => {
        return files.filter((incomingFile) => {
            const isAlreadyInStore = hasFile(incomingFile);
            return !isAlreadyInStore;
        });
    };

    const handleUploadFiles = async (files: File[]) => {
        const sanitized = sanitizeNames(sanitizeFormats(files));
        const unique = filterDuplicates(sanitized);
        const batchSize = files.reduce((total, file) => total + file.size, 0);
        console.log("Size " + batchSize + storeUploadSize);
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

    return { handleUploadFiles };
};

export default useUpload;
