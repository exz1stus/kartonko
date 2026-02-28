"use client";
import useAuthFetch from "@/hooks/useAuthorizedFetch";
import { useState } from "react";

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
    const authFetch = useAuthFetch();
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

        const response = await authFetch(`${API_ORIGIN}/upload`, {
            method: "POST",
            body: formData,
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
