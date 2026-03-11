"use client";
import { useRouter, usePathname } from "next/navigation";
import useUploadStore from "./useUploadStore";
import isAllowed from "@/lib/allowedFormats";
import { toast } from "sonner";
import { hashFile } from "@/lib/image";
import { existsOnServer } from "@/lib/image.client";

const useUpload = () => {
    const addFiles = useUploadStore((state) => state.addFiles);

    const router = useRouter();
    const pathname = usePathname();

    const sanitizeFormats = (files: File[]) =>
        files.filter((file) => {
            const format = file.type.split("/")[1];
            if (!isAllowed(format))
                toast.error(`format ${format} is not allowed`);
            return isAllowed(format);
        });

    const checkForHashes = async (files: File[]): Promise<File[]> => {
        const uniqueFiles: File[] = [];

        for (const file of files) {
            const hash = await hashFile(file);
            const exists = await existsOnServer(hash);
            if (exists) {
                toast.error(`File with hash already uploaded!`);
                continue;
            }
            uniqueFiles.push(file);
        }

        return uniqueFiles;
    };

    const handleUploadFiles = async (files: File[]) => {
        files = sanitizeFormats(files);
        files = await checkForHashes(files);
        if (files.length === 0) return;
        addFiles(files);
        if (!pathname.startsWith("/upload")) {
            router.push("/upload");
        }
    };

    return { handleUploadFiles };
};

export default useUpload;
