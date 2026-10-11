"use client";
import { clientMutator } from "../api/clientMutator";
import ApiError from "../api/error";
import { ImageMetadata, ImagePostBatchResponse } from "../api/generated/model";

interface ImageUploadRequest {
    client_id?: string;
    name: string;
    tags: string[];
    newTags?: string[];
}

interface ImageBatchUploadRequest {
    data: ImageUploadRequest[];
    common_tags: string[];
}

export async function uploadImage(
    metadata: ImageUploadRequest,
    file: File,
    captchaToken: string,
): Promise<ImageMetadata> {
    const formData = new FormData();

    formData.append("metadata", JSON.stringify(metadata));
    formData.append("cf-turnstile-response", captchaToken);
    formData.append("file", file);

    return await clientMutator(`/upload`, {
        method: "POST",
        body: formData,
        credentials: "include",
    });
}

export async function uploadImageBatch(
    batchMetadata: ImageBatchUploadRequest,
    files: File[],
    uploadSession: string,
    onProgress?: (loaded: number, total: number) => void,
): Promise<ImagePostBatchResponse> {
    const formData = new FormData();

    formData.append("metadata", JSON.stringify(batchMetadata));
    files.forEach((file) => {
        formData.append("files", file);
    });
    return new Promise((resolve, reject) => {
        const request = new XMLHttpRequest();
        request.open("POST", "/api/upload/batch");
        request.setRequestHeader("x-upload-session", uploadSession);
        request.upload.onprogress = (event) => {
            if (event.lengthComputable) onProgress?.(event.loaded, event.total);
        };
        request.onerror = () => reject(new Error("Upload connection failed"));
        request.onload = () => {
            let data: unknown;
            try {
                data = JSON.parse(request.responseText);
            } catch {
                reject(new Error(`Upload failed with status ${request.status}`));
                return;
            }
            if (request.status >= 200 && request.status < 300)
                resolve(data as ImagePostBatchResponse);
            else
                reject(new ApiError(request.status, data));
        };
        request.send(formData);
    });
}

export async function createUploadSession(captchaToken: string): Promise<string> {
    const formData = new FormData();
    formData.append("cf-turnstile-response", captchaToken);
    const response = await clientMutator<{ token: string }>("/upload/session", {
        method: "POST",
        body: formData,
    });
    return response.token;
}

export { type ImageUploadRequest, type ImageBatchUploadRequest };
