"use client";
import { useRouter } from "next/navigation";
import useUploadStore from "./useUploadStore";

const useUpload = () => {
    const setFiles = useUploadStore((s) => s.setFiles);

    const router = useRouter();
    const handleUploadFiles = async (files: File[]) => {
        setFiles(files);
        router.push("/upload");
    };

    return { handleUploadFiles };
};

export default useUpload;
