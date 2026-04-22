"use client";
import { useState } from "react";
import { apiFetch } from "@/lib/apiFetch";

export interface ApiResponse {
    error?: string;
    message?: string;
}

const API_ORIGIN = process.env.NEXT_PUBLIC_API_ORIGIN;

interface ImageUploadRequest {
    name: string;
    tags: string[];
}

const useUploadImage = () => {
    const [loading, setLoading] = useState(false);

    const uploadImage = async (
        imageMetadata: ImageUploadRequest,
        file: File,
    ): Promise<boolean> => {
        if (loading) return false;
        setLoading(true);
        const formData = new FormData();
        formData.append("metadata", JSON.stringify(imageMetadata));
        formData.append("file", file);

        const response = await apiFetch(`/upload`, {
            method: "POST",
            body: formData,
            credentials: "include",
        });

        setLoading(false);
        const parsedResponse: ApiResponse = await response.json();
        if (parsedResponse?.error) {
            throw new Error(parsedResponse.error);
        }

        return true;
    };

    return { uploadImage, loading };
};

export { type ImageUploadRequest, useUploadImage };
