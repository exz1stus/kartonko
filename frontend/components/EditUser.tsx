"use client";
import { Pencil, Save, Trash2, X } from "lucide-react";
import { ChangeEvent, useCallback, useEffect, useState } from "react";
import { toast } from "sonner";
import { useRouter } from "next/navigation";
import {
    deleteImagesByQuery,
    patchUser,
} from "@/lib/api/generated/client";
import { UserDataResponse } from "@/lib/api/generated/model/userDataResponse";
import { sanitizeName } from "@/lib/sanitizeName";

interface Props {
    user: UserDataResponse;
    hasPermission: boolean;
}

const EditUser = ({ user, hasPermission }: Props) => {
    const router = useRouter();
    const [editing, setEditing] = useState(false);
    const [username, setUsername] = useState(user.username);
    const [pending, setPending] = useState(false);

    useEffect(() => {
        if (editing) return;
        setUsername(user.username);
    }, [editing, user.username]);

    const save = useCallback(async () => {
        if (pending) return;
        setPending(true);
        try {
            const nextUsername = sanitizeName(username.trim());
            if (!nextUsername) throw new Error("Username cannot be empty");

            const updated = await patchUser(
                user.id,
                { username: nextUsername },
                { credentials: "include" },
            );
            setEditing(false);
            router.replace(`/user/${encodeURIComponent(updated.username)}`);
            router.refresh();
        } finally {
            setPending(false);
        }
    }, [pending, router, user.id, username]);

    const uploadPicture = useCallback(async (file: File) => {
        if (pending) return;
        setPending(true);
        try {
            const form = new FormData();
            form.append("picture", file);
            const response = await fetch(`/apilocal/user/${user.id}/picture`, {
                method: "POST",
                credentials: "include",
                body: form,
            });
            const result = await response.json().catch(() => null);
            if (!response.ok) throw new Error(result?.error ?? "Could not upload profile picture");
            router.refresh();
        } finally {
            setPending(false);
        }
    }, [pending, router, user.id]);

    const onPictureSelected = (event: ChangeEvent<HTMLInputElement>) => {
        const input = event.currentTarget;
        const file = input.files?.[0];
        if (!file) return;
        void toast.promise(uploadPicture(file), {
            loading: "Uploading profile picture…",
            success: "Profile picture updated",
            error: (error) => error.message,
        }).finally(() => { input.value = ""; });
    };

    const deleteUsersImages = useCallback(async () => {
        if (pending) return;
        if (!window.confirm(`Delete every image uploaded by ${user.username}?`)) return;
        setPending(true);
        toast.promise(
            deleteImagesByQuery({ user_id: user.id }, { credentials: "include" }).finally(() => setPending(false)),
            {
                loading: "Loading...",
                success: () => {
                    router.refresh();
                    return `user images have been deleted`;
                },
                error: (error) => error.message,
            },
        );
    }, [pending, router, user.id, user.username]);

    if (!hasPermission) return null;

    return (
        <div className="flex flex-col items-center gap-2">
            <span className="text-2xl">Edit:</span>
            {hasPermission && (
                <>
                    {!editing && (
                        <button
                            type="button"
                            onClick={() => setEditing(true)}
                            className="flex flex-row items-center gap-2 border rounded-2xl px-2 py-1 hover:border-primary"
                        >
                            <Pencil size={16} /> Edit profile
                        </button>
                    )}
                    {editing && (
                        <div className="flex flex-col gap-2 w-full max-w-sm">
                            <label className="inline-flex w-fit cursor-pointer items-center gap-2 border rounded-md px-2 py-1 hover:border-primary">
                                <Pencil size={16} /> Browse profile picture
                                <input type="file" accept="image/png,image/jpeg,image/gif" disabled={pending} onChange={onPictureSelected} className="sr-only" />
                            </label>
                            <input
                                className="border rounded-md px-2 py-1 bg-transparent"
                                value={username}
                                onChange={(event) => setUsername(sanitizeName(event.target.value))}
                                aria-label="Username"
                            />
                            <div className="flex gap-2">
                                <button
                                    type="button"
                                    disabled={pending}
                                    onClick={() => void toast.promise(save(), {
                                        loading: "Saving...",
                                        success: "Profile updated",
                                        error: (error) => error.message,
                                    })}
                                    className="inline-flex items-center gap-1 border rounded-md px-2 py-1 hover:border-primary"
                                >
                                    <Save size={16} /> Save
                                </button>
                                <button
                                    type="button"
                                    disabled={pending}
                                    onClick={() => {
                                        setUsername(user.username);
                                        setEditing(false);
                                    }}
                                    className="inline-flex items-center gap-1 border rounded-md px-2 py-1"
                                >
                                    <X size={16} /> Cancel
                                </button>
                            </div>
                        </div>
                    )}
                    <button
                        type="button"
                        disabled={pending}
                        onClick={() => void deleteUsersImages()}
                        className="flex flex-row items-center gap-2 border hover:border-red-500 rounded-2xl hover:text-red-500 px-2 py-1"
                    >
                        <Trash2 size={16} />
                        <span>Delete user images</span>
                    </button>
                </>
            )}
        </div>
    );
};

export default EditUser;
