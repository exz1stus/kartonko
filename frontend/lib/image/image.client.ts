"use client";
import { getImageByHash } from "@/lib/api/generated/client";
import ApiError from "../api/error";

export async function existsOnServer(hash: string): Promise<boolean> {
    try {
        await getImageByHash(hash);
    } catch (error) {
        if (error instanceof ApiError && error.status === 404) return false;

        throw error;
    }

    return true;
}
