"use client";
import { useRouter, usePathname } from "next/navigation";
import useUploadStore from "./useUploadStore";
import isAllowed from "@/lib/allowedFormats";
import { toast } from "sonner";
import { hashFile } from "@/lib/image";
import { existsOnServer } from "@/lib/image.client";
import ProgressBarToast from "@/components/ProgressBarToast";
import pLimit from "p-limit";

const useUpload = () => {
    const { hasFile, addFiles } = useUploadStore((state) => state);

    const router = useRouter();
    const pathname = usePathname();

    const sanitizeFormats = (files: File[]) =>
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
        const sanitized = sanitizeFormats(files);
        const unique = filterDuplicates(sanitized);
        if (unique.length === 0) return;

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
