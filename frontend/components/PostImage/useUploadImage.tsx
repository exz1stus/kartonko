import useAuthFetch from "@/app/hooks/useAuthorizedFetch";
import useErrors from "@/components/Errors/useErrors";

interface Metadata {
    name: string,
    tags: string[],
}

export interface ApiResponse {
    error?: string;
    message?: string;
}

const API_ORIGIN = process.env.NEXT_PUBLIC_API_ORIGIN;

const useUploadImage = () => {
    const { showError } = useErrors();

    const authFetch = useAuthFetch();

    const handleError = (error: string) => {
        if (error.includes("already exists")) {
            showError("Image already exists");
            return;
        }
        else showError(error);
    }

    const uploadImage = async (imageMetadata: Metadata, file: File): Promise<boolean> => {
        const formData = new FormData();
        formData.append("metadata", JSON.stringify(imageMetadata));
        formData.append("file", file);

        try {
            const response = await authFetch(`${API_ORIGIN}/upload`, {
                method: "POST",
                body: formData,
            });

            const parsedResponse: ApiResponse = await response.json();
            if (parsedResponse?.error) {
                handleError(parsedResponse.error);
                return false;
            }

            if (parsedResponse.message) {
                return true;
            }
        } catch (err: any) {
            showError(err);
        }

        return false;
    }

    return { uploadImage }
}

export { useUploadImage, type Metadata };
