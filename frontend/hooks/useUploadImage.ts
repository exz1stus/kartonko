"use client";
import { useRef, useState } from "react";
import {
    ImageBatchUploadRequest,
    ImageUploadRequest,
    uploadImage,
    uploadImageBatch,
} from "@/lib/image/upload";

export const useUploadImage = () => {
    const loadingRef = useRef(false);
    const [loading, setLoading] = useState(false);

    const runUpload = async <T>(upload: () => Promise<T>): Promise<T> => {
        if (loadingRef.current) throw new Error("Upload already in progress");
        loadingRef.current = true;
        setLoading(true);
        let res: T;
        try {
            res = await upload();
        } finally {
            loadingRef.current = false;
            setLoading(false);
        }
        return res;
    };

    const handleUploadImage = (
        metadata: ImageUploadRequest,
        file: File,
        captchaToken: string,
    ) => runUpload(() => uploadImage(metadata, file, captchaToken));

    const handleUploadBatch = (
        batchMetadata: ImageBatchUploadRequest,
        files: File[],
        captchaToken: string,
    ) => runUpload(() => uploadImageBatch(batchMetadata, files, captchaToken));

    return {
        uploadImage: handleUploadImage,
        uploadImageBatch: handleUploadBatch,
        loading,
    };
};
