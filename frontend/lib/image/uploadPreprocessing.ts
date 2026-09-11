"use client";
import { toast } from "sonner";
import { sanitizeName } from "../sanitizeName";
import isAllowed from "./allowedFormats";
import { hashFile } from "./image";
import { existsOnServer } from "./image.client";
import pLimit from "p-limit";

export function sanitizeNames(files: File[]): File[] {
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
}
export function sanitizeFormats(files: File[]): File[] {
    return files.filter((file) => {
        const format = file.type.split("/")[1];
        if (!isAllowed(format)) toast.error(`format ${format} is not allowed`);
        return isAllowed(format);
    });
}

export async function checkForHashes(
    files: File[],
    onProgress: (completed: number) => void,
): Promise<File[]> {
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
}

const limit = pLimit(10);
