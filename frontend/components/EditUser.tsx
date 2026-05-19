"use client";
import { Trash2 } from "lucide-react";
import { apiFetch } from "@/lib/apiFetch";
import { useCallback } from "react";
import { toast } from "sonner";
import { useRouter } from "next/navigation";
import { UserData } from "@/lib/user";

interface Props {
    user: UserData;
    hasPermission: boolean;
}

interface ApiResponse {
    error?: string;
    message?: string;
}

const EditUser = ({ user, hasPermission }: Props) => {
    const router = useRouter();
    const fetchDelete = useCallback(async () => {
        const res = await apiFetch(`/image?user_id=${user.id}`, {
            method: "DELETE",
            credentials: "include",
        });
        const data: ApiResponse = await res.json();
        if (data?.error) {
            throw new Error(data.error);
        }

        return true;
    }, [user]);

    const deleteUsersImages = useCallback(async () => {
        toast.promise(fetchDelete, {
            loading: "Loading...",
            success: () => {
                return `user images have been deleted`;
            },
            error: (error) => error.message,
        });
    }, [fetchDelete, router]);

    return (
        <div className="flex flex-col items-center gap-2">
            <span className="text-2xl">Edit:</span>
            {hasPermission && (
                <div
                    onClick={deleteUsersImages}
                    className="flex flex-row items-center gap-2 border hover:border-red-500 rounded-2xl hover:text-red-500 cursor-pointer"
                >
                    <Trash2 className="m-2" />
                    <span className="px-2 py-1">Delete user images</span>
                </div>
            )}
        </div>
    );
};

export default EditUser;
