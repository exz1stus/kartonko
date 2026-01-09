import useAuthFetch from "@/app/hooks/useAuthorizedFetch";

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

    const handleError = (error: string) => {
        if (error.includes("already exists")) {
            console.log("Image already exists");
            return;
        } else console.log(error);
    };

    const uploadImage = async (imageMetadata: ImageUploadRequest, file: File): Promise<boolean> => {
        const formData = new FormData();
        formData.append("metadata", JSON.stringify(imageMetadata));
        formData.append("file", file);

        const response = await authFetch(`${API_ORIGIN}/upload`, {
            method: "POST",
            body: formData,
        });

        const parsedResponse: ApiResponse = await response.json();
        if (parsedResponse?.error) {
            handleError(parsedResponse.error);
            throw new Error(parsedResponse.error);
        }

        return true;
    };

    return { uploadImage };
};

export { type ImageUploadRequest, useUploadImage };
