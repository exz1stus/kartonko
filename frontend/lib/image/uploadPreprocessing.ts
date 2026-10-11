"use client";
import { toast } from "sonner";
import { sanitizeName } from "../sanitizeName";
import isAllowed from "./allowedFormats";
import { hashFile } from "./image";
import { getImageByHash } from "@/lib/api/generated/client";
import ApiError from "@/lib/api/error";
import { ImageMetadata } from "@/lib/api/generated/model";
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
): Promise<{ ready: File[]; existing: ImageMetadata[] }> {
    let completed = 0;

    const tasks = files.map((file) =>
        limit(async () => {
            try {
                const hash = await hashFile(file);
                try {
                    return { file, hash, existing: await getImageByHash(hash) };
                } catch (error) {
                    if (error instanceof ApiError && error.status === 404) return { file, hash, existing: null };
                    throw error;
                }
            } finally {
                completed++;
                onProgress(completed);
            }
        }),
    );

    const results = await Promise.all(tasks);
    const seenHashes = new Set<string>();
    const ready = results.filter((result) => {
        if (result.existing || seenHashes.has(result.hash)) return false;
        seenHashes.add(result.hash);
        return true;
    }).map((result) => result.file);
    return {
        ready,
        existing: results.map((result) => result.existing).filter((image): image is ImageMetadata => image !== null),
    };
}

const limit = pLimit(10);
