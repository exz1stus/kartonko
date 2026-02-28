"use client"
import ImageMetadata from '@/lib/image'
import { Delete, Trash, Trash2 } from "lucide-react"
import { apiFetch } from "@/lib/apiFetch"
import { useCallback } from "react"
import { toast } from "sonner"
import { useRouter } from "next/navigation"

interface Props {
    image: ImageMetadata
    onDelete?: () => void
}

interface ApiResponse {
    error?: string;
    message?: string;
}

const EditImage = ({ image, onDelete }: Props) => {
    const router = useRouter();
    const fetchDelete = useCallback(async () => {
        const res = await apiFetch(`/image/${image.filename}`, { method: "DELETE", credentials:"include" });
        const data: ApiResponse = await res.json();
        if (data?.error) {
            throw new Error(data.error);
        }

        return true;
    },[]);

    const deleteImage = useCallback(async () => {
        toast.promise(fetchDelete, {
            loading: "Loading...",
            success: () => {
                onDelete?.();
                router.push("/");
                return `image has been deleted`;
            },
            error: (error) => error.message,
        });
    },[]);

    return (
      <div className="flex flex-row items-center gap-2">
          <span className="text-2xl">Edit:</span>
          <Trash2 onClick={deleteImage} className="m-2 hover:text-red-500 cursor-pointer"/>
      </div>
    )
}

export default EditImage
