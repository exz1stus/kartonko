"use client";
import { useRef, useState } from "react";
import { apiFetch } from "@/lib/apiFetch";

export interface ApiResponse {
    error?: string;
    message?: string;
}

interface ImageUploadRequest {
    name: string;
    tags: string[];
}

interface ImageBatchUploadRequest {
    data: ImageUploadRequest[];
    common_tags: string[];
}

const useUploadImage = () => {
    const [loading, setLoading] = useState(false);

    const uploadImage = async (
        imageMetadata: ImageUploadRequest,
        file: File,
        captchaToken: string,
    ): Promise<boolean> => {
        if (loading) return false;
        setLoading(true);

        const formData = new FormData();

        formData.append("metadata", JSON.stringify(imageMetadata));
        formData.append("file", file);
        formData.append("cf-turnstile-response", captchaToken);

        const response = await apiFetch(`/upload`, {
            method: "POST",
            body: formData,
            credentials: "include",
        });

        try {
            setLoading(false);
            const parsedResponse: ApiResponse = await response.json();
            if (parsedResponse?.error) {
                throw new Error(parsedResponse.error);
            }

            return true;
        } finally {
            setLoading(false);
        }
    };

    const uploadImageBatch = async (
        batchMetadata: ImageBatchUploadRequest,
        files: File[],
        captchaToken: string,
    ): Promise<boolean> => {
        if (loading) return false;
        setLoading(true);

        const formData = new FormData();

        formData.append("metadata", JSON.stringify(batchMetadata));
        files.forEach((file) => {
            formData.append("files", file);
        });
        formData.append("cf-turnstile-response", captchaToken);

        try {
            const response = await apiFetch(`/upload/batch`, {
                method: "POST",
                body: formData,
                credentials: "include",
            });

            const parsedResponse: ApiResponse = await response.json();
            if (parsedResponse?.error) {
                throw new Error(parsedResponse.error);
            }
            return true;
        } finally {
            setLoading(false);
        }
    };

    return { uploadImageBatch, uploadImage, loading };
};

export {
    type ImageUploadRequest,
    type ImageBatchUploadRequest,
    useUploadImage,
};
