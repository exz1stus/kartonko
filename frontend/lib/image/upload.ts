"use client";
import { clientMutator } from "../api/clientMutator";
import { ImageMetadata, ImagePostBatchResponse } from "../api/generated/model";

interface ImageUploadRequest {
    name: string;
    tags: string[];
    newTags?: string[];
}

interface ImageBatchUploadRequest {
    data: ImageUploadRequest[];
    common_tags: string[];
}

const MAXIMUM_UPLOAD_SIZE = 50 * 1024 * 1024; //50MB

export function isUploadSizeValid(sizeBytes: number): boolean {
    return sizeBytes <= MAXIMUM_UPLOAD_SIZE;
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
    captchaToken: string,
): Promise<ImagePostBatchResponse> {
    const formData = new FormData();

    formData.append("metadata", JSON.stringify(batchMetadata));
    files.forEach((file) => {
        formData.append("files", file);
    });
    formData.append("cf-turnstile-response", captchaToken);

    return await clientMutator(`/upload/batch`, {
        method: "POST",
        body: formData,
        credentials: "include",
    });
}

export { type ImageUploadRequest, type ImageBatchUploadRequest };
