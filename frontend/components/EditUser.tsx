"use client";
import { Trash2 } from "lucide-react";
import { useCallback } from "react";
import { toast } from "sonner";
import { useRouter } from "next/navigation";
import { deleteImage } from "@/lib/api/generated/client";
import { UserDataResponse } from "@/lib/api/generated/model/userDataResponse";

interface Props {
    user: UserDataResponse;
    hasPermission: boolean;
}

const EditUser = ({ user, hasPermission }: Props) => {
    const router = useRouter();

    const deleteUsersImages = useCallback(async () => {
        toast.promise(
            deleteImage({ user_id: user.id }, { credentials: "include" }),
            {
                loading: "Loading...",
                success: () => {
                    return `user images have been deleted`;
                },
                error: (error) => error.message,
            },
        );
    }, [user.id]);

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
